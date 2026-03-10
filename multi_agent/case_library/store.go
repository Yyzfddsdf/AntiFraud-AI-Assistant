package case_library

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"antifraud/cache"
	"antifraud/database"
	"antifraud/embedding"
	model "antifraud/multi_agent/case_library/model"
)

const (
	scamTypesConfigPath    = "config/scam_types.json"
	targetGroupsConfigPath = "config/target_groups.json"
)

const (
	minCaseDescriptionRunes   = 12
	maxCaseDescriptionRunes   = 400
	randomLikeAlnumChunkLimit = 16
)

const historicalCaseGraphCacheVersionKey = "cache:case_library:graph:v1:version"

// FixedRiskLevels 为上传历史案件时允许的风险等级枚举。
var FixedRiskLevels = []string{
	"高",
	"中",
	"低",
}

type ValidationError = model.ValidationError

func newValidationError(format string, args ...interface{}) error {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

type CreateHistoricalCaseInput = model.CreateHistoricalCaseInput
type HistoricalCaseRecord = model.HistoricalCaseRecord
type HistoricalCasePreview = model.HistoricalCasePreview
type historicalCaseEntity = model.HistoricalCaseEntity

// CreateHistoricalCase 将历史案件写入独立数据库，并保存 embedding 向量。
func CreateHistoricalCase(ctx context.Context, userID string, input CreateHistoricalCaseInput) (HistoricalCaseRecord, error) {
	normalizedInput, err := normalizeAndValidateInput(input)
	if err != nil {
		return HistoricalCaseRecord{}, err
	}

	embeddingText := BuildEmbeddingInput(normalizedInput)
	vector, modelName, err := embedding.GenerateVector(ctx, embeddingText)
	if err != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("generate embedding failed: %w", err)
	}

	entity := historicalCaseEntity{
		CaseID:             newHistoricalCaseID(),
		CreatedBy:          normalizeUserID(userID),
		Title:              normalizedInput.Title,
		TargetGroup:        normalizedInput.TargetGroup,
		RiskLevel:          normalizedInput.RiskLevel,
		ScamType:           normalizedInput.ScamType,
		CaseDescription:    normalizedInput.CaseDescription,
		TypicalScripts:     encodeStringList(normalizedInput.TypicalScripts),
		Keywords:           encodeStringList(normalizedInput.Keywords),
		ViolatedLaw:        normalizedInput.ViolatedLaw,
		Suggestion:         normalizedInput.Suggestion,
		EmbeddingVector:    encodeFloatList(vector),
		EmbeddingModel:     strings.TrimSpace(modelName),
		EmbeddingDimension: len(vector),
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return HistoricalCaseRecord{}, err
	}
	if err := db.Create(&entity).Error; err != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("insert historical case failed: %w", err)
	}
	record := recordFromEntity(entity)
	upsertHistoricalCaseVectorCache(record)
	touchHistoricalCaseGraphCacheVersion()
	return record, nil
}

// ListHistoricalCasePreviews 返回历史案件预览数据，用于列表页展示。
func ListHistoricalCasePreviews() ([]HistoricalCasePreview, error) {
	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return nil, err
	}

	rows := make([]historicalCaseEntity, 0)
	if err := db.Select("case_id", "title", "target_group", "risk_level", "scam_type", "created_at").
		Order("created_at desc").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query historical case previews failed: %w", err)
	}

	previews := make([]HistoricalCasePreview, 0, len(rows))
	for _, row := range rows {
		normalizedRiskLevel := normalizeRiskLevel(row.RiskLevel)
		if normalizedRiskLevel == "" {
			normalizedRiskLevel = strings.TrimSpace(row.RiskLevel)
		}

		previews = append(previews, HistoricalCasePreview{
			CaseID:      strings.TrimSpace(row.CaseID),
			Title:       strings.TrimSpace(row.Title),
			TargetGroup: strings.TrimSpace(row.TargetGroup),
			RiskLevel:   normalizedRiskLevel,
			ScamType:    strings.TrimSpace(row.ScamType),
			CreatedAt:   row.CreatedAt,
		})
	}
	return previews, nil
}

// GetHistoricalCaseByID 根据 case_id 返回完整历史案件详情。
func GetHistoricalCaseByID(caseID string) (HistoricalCaseRecord, bool, error) {
	trimmedCaseID := strings.TrimSpace(caseID)
	if trimmedCaseID == "" {
		return HistoricalCaseRecord{}, false, nil
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return HistoricalCaseRecord{}, false, err
	}

	var entity historicalCaseEntity
	query := db.Where("case_id = ?", trimmedCaseID).Limit(1).Find(&entity)
	if query.Error != nil {
		return HistoricalCaseRecord{}, false, fmt.Errorf("query historical case detail failed: %w", query.Error)
	}
	if query.RowsAffected == 0 {
		return HistoricalCaseRecord{}, false, nil
	}
	return recordFromEntity(entity), true, nil
}

// DeleteHistoricalCaseByID 根据 case_id 删除历史案件。
func DeleteHistoricalCaseByID(caseID string) (bool, error) {
	trimmedCaseID := strings.TrimSpace(caseID)
	if trimmedCaseID == "" {
		return false, nil
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return false, err
	}

	result := db.Where("case_id = ?", trimmedCaseID).Delete(&historicalCaseEntity{})
	if result.Error != nil {
		return false, fmt.Errorf("delete historical case failed: %w", result.Error)
	}
	deleted := result.RowsAffected > 0
	if deleted {
		removeHistoricalCaseVectorCache(trimmedCaseID)
		touchHistoricalCaseGraphCacheVersion()
	}
	return deleted, nil
}

// HistoricalCaseGraphCacheVersion 返回当前图谱缓存版本；读取失败时回退为 "0"。
func HistoricalCaseGraphCacheVersion() string {
	var version string
	found, err := cache.GetJSON(historicalCaseGraphCacheVersionKey, &version)
	if err != nil || !found {
		return "0"
	}
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return "0"
	}
	return trimmed
}

func touchHistoricalCaseGraphCacheVersion() {
	version := fmt.Sprintf("%d", time.Now().UnixNano())
	if err := cache.SetJSON(historicalCaseGraphCacheVersionKey, version, 0); err != nil {
		log.Printf("[case_library] touch graph cache version failed: %v", err)
	}
}

func normalizeAndValidateInput(input CreateHistoricalCaseInput) (CreateHistoricalCaseInput, error) {
	normalized := CreateHistoricalCaseInput{
		Title:           strings.TrimSpace(input.Title),
		TargetGroup:     normalizeTargetGroup(input.TargetGroup),
		RiskLevel:       normalizeRiskLevel(input.RiskLevel),
		ScamType:        normalizeScamType(input.ScamType),
		CaseDescription: strings.TrimSpace(input.CaseDescription),
		TypicalScripts:  normalizeStringList(input.TypicalScripts),
		Keywords:        normalizeStringList(input.Keywords),
		ViolatedLaw:     strings.TrimSpace(input.ViolatedLaw),
		Suggestion:      strings.TrimSpace(input.Suggestion),
	}

	if normalized.Title == "" {
		return CreateHistoricalCaseInput{}, newValidationError("title is required")
	}
	if normalized.TargetGroup == "" {
		return CreateHistoricalCaseInput{}, newValidationError("target_group is invalid, allowed values: %s", strings.Join(ListTargetGroups(), ", "))
	}
	if normalized.RiskLevel == "" {
		return CreateHistoricalCaseInput{}, newValidationError("risk_level is invalid, allowed values: %s", strings.Join(FixedRiskLevels, ", "))
	}
	if normalized.ScamType == "" {
		return CreateHistoricalCaseInput{}, newValidationError("scam_type is invalid, allowed values: %s", strings.Join(ListScamTypes(), ", "))
	}
	if normalized.CaseDescription == "" {
		return CreateHistoricalCaseInput{}, newValidationError("case_description is required")
	}
	if err := validateCaseDescriptionQuality(normalized.CaseDescription); err != nil {
		return CreateHistoricalCaseInput{}, err
	}

	return normalized, nil
}

func validateCaseDescriptionQuality(description string) error {
	trimmed := strings.TrimSpace(description)
	normalized := strings.Join(strings.Fields(trimmed), " ")
	runes := []rune(normalized)

	if len(runes) < minCaseDescriptionRunes {
		return newValidationError("case_description is too short, please provide at least %d characters", minCaseDescriptionRunes)
	}
	if len(runes) > maxCaseDescriptionRunes {
		return newValidationError("case_description is too long, max %d characters allowed", maxCaseDescriptionRunes)
	}

	uniqueChars := map[rune]struct{}{}
	var (
		hasHan            bool
		hasSeparator      bool
		maxAlnumChunk     int
		currentAlnumChunk int
		alnumCount        int
		digitCount        int
	)

	for _, r := range runes {
		uniqueChars[r] = struct{}{}

		switch {
		case unicode.In(r, unicode.Han):
			hasHan = true
			currentAlnumChunk = 0
		case unicode.IsLetter(r):
			alnumCount++
			currentAlnumChunk++
		case unicode.IsDigit(r):
			alnumCount++
			digitCount++
			currentAlnumChunk++
		default:
			hasSeparator = true
			currentAlnumChunk = 0
		}

		if currentAlnumChunk > maxAlnumChunk {
			maxAlnumChunk = currentAlnumChunk
		}
	}

	if len(uniqueChars) <= 2 {
		return newValidationError("case_description appears invalid, please provide meaningful content")
	}

	if !hasHan && !hasSeparator && maxAlnumChunk >= randomLikeAlnumChunkLimit {
		return newValidationError("case_description appears random or lacks semantics")
	}

	if !hasHan && alnumCount >= minCaseDescriptionRunes {
		digitRatio := float64(digitCount) / float64(alnumCount)
		if maxAlnumChunk >= minCaseDescriptionRunes && digitRatio > 0.35 {
			return newValidationError("case_description appears random or lacks semantics")
		}
	}

	return nil
}

func normalizeTargetGroup(raw string) string {
	group := strings.TrimSpace(raw)
	if group == "" {
		return ""
	}
	for _, allowed := range ListTargetGroups() {
		if strings.TrimSpace(allowed) == group {
			return group
		}
	}
	return ""
}

func normalizeScamType(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	for _, allowed := range ListScamTypes() {
		if strings.TrimSpace(allowed) == value {
			return value
		}
	}
	return ""
}

// ListScamTypes 返回诈骗类型配置（动态增改依赖配置文件）。
func ListScamTypes() []string {
	configPath := resolveScamTypesConfigPath()
	raw, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("[case_library] read scam types config failed: path=%s err=%v", configPath, err)
		return []string{}
	}

	var wrapper struct {
		ScamTypes []string `json:"scam_types"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		var plain []string
		if err2 := json.Unmarshal(raw, &plain); err2 != nil {
			log.Printf("[case_library] parse scam types config failed: path=%s err=%v", configPath, err2)
			return []string{}
		}
		return normalizeStringList(plain)
	}
	return normalizeStringList(wrapper.ScamTypes)
}

// ListTargetGroups 返回目标人群配置（动态增改依赖配置文件）。
func ListTargetGroups() []string {
	configPath := resolveTargetGroupsConfigPath()
	raw, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("[case_library] read target groups config failed: path=%s err=%v", configPath, err)
		return []string{}
	}

	var wrapper struct {
		TargetGroups []string `json:"target_groups"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		var plain []string
		if err2 := json.Unmarshal(raw, &plain); err2 != nil {
			log.Printf("[case_library] parse target groups config failed: path=%s err=%v", configPath, err2)
			return []string{}
		}
		return normalizeStringList(plain)
	}
	return normalizeStringList(wrapper.TargetGroups)
}

func resolveTargetGroupsConfigPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		return filepath.Join(projectRoot, targetGroupsConfigPath)
	}
	abs, err := filepath.Abs(targetGroupsConfigPath)
	if err == nil {
		return abs
	}
	return targetGroupsConfigPath
}

func resolveScamTypesConfigPath() string {
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		return filepath.Join(projectRoot, scamTypesConfigPath)
	}
	abs, err := filepath.Abs(scamTypesConfigPath)
	if err == nil {
		return abs
	}
	return scamTypesConfigPath
}

func normalizeRiskLevel(raw string) string {
	level := strings.TrimSpace(raw)
	if level == "" {
		return ""
	}

	normalizedEnglish := strings.ToLower(level)
	alias := map[string]string{
		"高":      "高",
		"高风险":    "高",
		"high":   "高",
		"HIGH":   "高",
		"中":      "中",
		"中风险":    "中",
		"medium": "中",
		"MEDIUM": "中",
		"低":      "低",
		"低风险":    "低",
		"low":    "低",
		"LOW":    "低",
	}
	if result := alias[level]; result != "" {
		return result
	}
	return alias[normalizedEnglish]
}

func normalizeStringList(items []string) []string {
	normalized := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

// BuildEmbeddingInput 将结构化字段转换为 embedding 输入文本。
func BuildEmbeddingInput(input CreateHistoricalCaseInput) string {
	segments := []string{}
	segments = appendEmbeddingSegment(segments, "标题", input.Title)
	segments = appendEmbeddingSegment(segments, "诈骗类型", input.ScamType)
	segments = appendEmbeddingSegment(segments, "案件描述", input.CaseDescription)
	if keywords := selectEmbeddingKeywords(input.Keywords); len(keywords) > 0 {
		segments = appendEmbeddingSegment(segments, "关键词", strings.Join(keywords, "、"))
	}

	return strings.Join(segments, "\n")
}

func selectEmbeddingKeywords(keywords []string) []string {
	normalized := normalizeStringList(keywords)
	if len(normalized) == 0 || len(normalized) > 8 {
		return nil
	}

	filtered := make([]string, 0, len(normalized))
	for _, keyword := range normalized {
		if isHighQualityEmbeddingKeyword(keyword) {
			filtered = append(filtered, keyword)
		}
	}

	if len(filtered) == 0 {
		return nil
	}
	if len(filtered)*2 < len(normalized) {
		return nil
	}
	return filtered
}

func isHighQualityEmbeddingKeyword(keyword string) bool {
	trimmed := strings.Join(strings.Fields(strings.TrimSpace(keyword)), " ")
	if trimmed == "" {
		return false
	}

	runes := []rune(trimmed)
	if len(runes) < 2 || len(runes) > 16 {
		return false
	}
	if strings.ContainsAny(trimmed, "，。；！？、;!?\r\n\t") {
		return false
	}

	hasAlphaNumeric := false
	for _, r := range runes {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r):
			hasAlphaNumeric = true
		case unicode.IsSpace(r), r == '-', r == '_', r == '/', r == '&', r == '+':
			continue
		default:
			return false
		}
	}
	return hasAlphaNumeric
}

func appendEmbeddingSegment(segments []string, key string, value string) []string {
	trimmedKey := strings.TrimSpace(key)
	trimmedValue := strings.TrimSpace(value)
	if trimmedKey == "" || trimmedValue == "" {
		return segments
	}
	return append(segments, trimmedKey+": "+trimmedValue)
}

func newHistoricalCaseID() string {
	buffer := make([]byte, 6)
	if _, err := rand.Read(buffer); err != nil {
		return fmt.Sprintf("HCASE-%d", time.Now().UnixNano())
	}
	return "HCASE-" + strings.ToUpper(hex.EncodeToString(buffer))
}

func normalizeUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return "unknown-user"
	}
	return trimmed
}

func encodeStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}

	payload, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func decodeStringList(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}
	var list []string
	if err := json.Unmarshal([]byte(trimmed), &list); err != nil {
		return []string{}
	}
	return normalizeStringList(list)
}

func encodeFloatList(items []float64) string {
	payload, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(payload)
}

func decodeFloatList(raw string) []float64 {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []float64{}
	}

	var list []float64
	if err := json.Unmarshal([]byte(trimmed), &list); err != nil {
		return []float64{}
	}
	return append([]float64{}, list...)
}

func recordFromEntity(entity historicalCaseEntity) HistoricalCaseRecord {
	normalizedRiskLevel := normalizeRiskLevel(entity.RiskLevel)
	if normalizedRiskLevel == "" {
		normalizedRiskLevel = strings.TrimSpace(entity.RiskLevel)
	}

	return HistoricalCaseRecord{
		CaseID:             strings.TrimSpace(entity.CaseID),
		CreatedBy:          normalizeUserID(entity.CreatedBy),
		Title:              strings.TrimSpace(entity.Title),
		TargetGroup:        strings.TrimSpace(entity.TargetGroup),
		RiskLevel:          normalizedRiskLevel,
		ScamType:           strings.TrimSpace(entity.ScamType),
		CaseDescription:    strings.TrimSpace(entity.CaseDescription),
		TypicalScripts:     decodeStringList(entity.TypicalScripts),
		Keywords:           decodeStringList(entity.Keywords),
		ViolatedLaw:        strings.TrimSpace(entity.ViolatedLaw),
		Suggestion:         strings.TrimSpace(entity.Suggestion),
		EmbeddingVector:    decodeFloatList(entity.EmbeddingVector),
		EmbeddingModel:     strings.TrimSpace(entity.EmbeddingModel),
		EmbeddingDimension: entity.EmbeddingDimension,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}
