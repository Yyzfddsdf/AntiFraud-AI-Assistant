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
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"image_recognition/config"
	openai "image_recognition/llm"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const historicalCaseDBPathEnv = "HISTORICAL_CASE_DB_PATH"
const (
	scamTypesConfigPath    = "config/scam_types.json"
	targetGroupsConfigPath = "config/target_groups.json"
)

const (
	minCaseDescriptionRunes   = 12
	maxCaseDescriptionRunes   = 400
	randomLikeAlnumChunkLimit = 16
)

var (
	caseDBOnce sync.Once
	caseDB     *gorm.DB
	caseDBErr  error
)

// FixedTargetGroups 为上传历史案件时允许的人群枚举。
// 如果后续需要扩展，新增值应同步更新 API 文档。
var FixedTargetGroups = []string{
	"老人",
	"青年",
	"中年",
	"未成年",
	"学生",
	"其他",
}

// FixedRiskLevels 为上传历史案件时允许的风险等级枚举。
var FixedRiskLevels = []string{
	"高",
	"中",
	"低",
}

type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

func newValidationError(format string, args ...interface{}) error {
	return &ValidationError{message: fmt.Sprintf(format, args...)}
}

func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

type CreateHistoricalCaseInput struct {
	Title           string
	TargetGroup     string
	RiskLevel       string
	ScamType        string
	CaseDescription string
	TypicalScripts  []string
	Keywords        []string
	ViolatedLaw     string
	Suggestion      string
}

type HistoricalCaseRecord struct {
	CaseID             string
	CreatedBy          string
	Title              string
	TargetGroup        string
	RiskLevel          string
	ScamType           string
	CaseDescription    string
	TypicalScripts     []string
	Keywords           []string
	ViolatedLaw        string
	Suggestion         string
	EmbeddingVector    []float64
	EmbeddingModel     string
	EmbeddingDimension int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type HistoricalCasePreview struct {
	CaseID      string
	Title       string
	TargetGroup string
	RiskLevel   string
	ScamType    string
}

type historicalCaseEntity struct {
	ID                 uint      `gorm:"primaryKey"`
	CaseID             string    `gorm:"size:32;uniqueIndex;not null"`
	CreatedBy          string    `gorm:"size:64;index;not null"`
	Title              string    `gorm:"type:text;not null"`
	TargetGroup        string    `gorm:"size:32;index;not null"`
	RiskLevel          string    `gorm:"size:16;index;not null;default:'中'"`
	ScamType           string    `gorm:"size:64;index;not null;default:'其他诈骗类'"`
	CaseDescription    string    `gorm:"type:text;not null"`
	TypicalScripts     string    `gorm:"type:text;not null"`
	Keywords           string    `gorm:"type:text;not null"`
	ViolatedLaw        string    `gorm:"type:text;not null"`
	Suggestion         string    `gorm:"type:text;not null"`
	EmbeddingVector    string    `gorm:"type:text;not null"`
	EmbeddingModel     string    `gorm:"size:128;not null"`
	EmbeddingDimension int       `gorm:"not null"`
	CreatedAt          time.Time `gorm:"index"`
	UpdatedAt          time.Time
}

func (historicalCaseEntity) TableName() string {
	return "historical_case_library"
}

// CreateHistoricalCase 将历史案件写入独立数据库，并保存 embedding 向量。
func CreateHistoricalCase(ctx context.Context, userID string, input CreateHistoricalCaseInput) (HistoricalCaseRecord, error) {
	normalizedInput, err := normalizeAndValidateInput(input)
	if err != nil {
		return HistoricalCaseRecord{}, err
	}

	embeddingText := buildEmbeddingInput(normalizedInput)
	vector, modelName, err := generateEmbeddingVector(ctx, embeddingText)
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

	db, err := getHistoricalCaseDB()
	if err != nil {
		return HistoricalCaseRecord{}, err
	}
	if err := db.Create(&entity).Error; err != nil {
		return HistoricalCaseRecord{}, fmt.Errorf("insert historical case failed: %w", err)
	}
	record := recordFromEntity(entity)
	upsertHistoricalCaseVectorCache(record)
	return record, nil
}

// ListHistoricalCasePreviews 返回历史案件预览数据，用于列表页展示。
func ListHistoricalCasePreviews() ([]HistoricalCasePreview, error) {
	db, err := getHistoricalCaseDB()
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

	db, err := getHistoricalCaseDB()
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

	db, err := getHistoricalCaseDB()
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
	}
	return deleted, nil
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

// buildEmbeddingInput 将结构化字段转换为 embedding 输入文本。
func buildEmbeddingInput(input CreateHistoricalCaseInput) string {
	segments := []string{}
	segments = appendEmbeddingSegment(segments, "标题", input.Title)
	segments = appendEmbeddingSegment(segments, "目标人群", input.TargetGroup)
	segments = appendEmbeddingSegment(segments, "风险等级", input.RiskLevel)
	segments = appendEmbeddingSegment(segments, "诈骗类型", input.ScamType)
	segments = appendEmbeddingSegment(segments, "案件描述", input.CaseDescription)

	if len(input.TypicalScripts) > 0 {
		segments = appendEmbeddingSegment(segments, "典型话术", strings.Join(input.TypicalScripts, "；"))
	}
	if len(input.Keywords) > 0 {
		segments = appendEmbeddingSegment(segments, "关键词", strings.Join(input.Keywords, "、"))
	}
	segments = appendEmbeddingSegment(segments, "违反法律", input.ViolatedLaw)
	segments = appendEmbeddingSegment(segments, "建议", input.Suggestion)

	return strings.Join(segments, "\n")
}

func appendEmbeddingSegment(segments []string, key string, value string) []string {
	trimmedKey := strings.TrimSpace(key)
	trimmedValue := strings.TrimSpace(value)
	if trimmedKey == "" || trimmedValue == "" {
		return segments
	}
	return append(segments, trimmedKey+": "+trimmedValue)
}

func generateEmbeddingVector(ctx context.Context, inputText string) ([]float64, string, error) {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return nil, "", fmt.Errorf("load config failed: %w", err)
	}

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  cfg.Embedding.APIKey,
		BaseURL: cfg.Embedding.BaseURL,
	})

	req := openai.EmbeddingRequest{
		Model:          cfg.Embedding.Model,
		Input:          []string{inputText},
		EncodingFormat: "float",
	}
	req.SetField("truncate", "NONE")

	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}
	resp, err := client.CreateEmbeddings(callCtx, req)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Data) == 0 {
		return nil, "", fmt.Errorf("embedding response is empty")
	}
	sort.Slice(resp.Data, func(i, j int) bool {
		return resp.Data[i].Index < resp.Data[j].Index
	})

	vector := append([]float64{}, resp.Data[0].Embedding...)
	if len(vector) == 0 {
		return nil, "", fmt.Errorf("embedding vector is empty")
	}

	modelName := strings.TrimSpace(resp.Model)
	if modelName == "" {
		modelName = strings.TrimSpace(cfg.Embedding.Model)
	}
	return vector, modelName, nil
}

func getHistoricalCaseDB() (*gorm.DB, error) {
	caseDBOnce.Do(func() {
		dbPath := resolveHistoricalCaseDBPath()
		dbDir := filepath.Dir(dbPath)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				caseDBErr = fmt.Errorf("create historical case db directory failed: %w", err)
				return
			}
		}

		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			caseDBErr = fmt.Errorf("open historical case db failed: %w", err)
			return
		}
		if err := db.AutoMigrate(&historicalCaseEntity{}); err != nil {
			caseDBErr = fmt.Errorf("auto migrate historical case db failed: %w", err)
			return
		}

		log.Printf("[case_library] historical case db path: %s", dbPath)
		caseDB = db
	})
	if caseDBErr != nil {
		return nil, caseDBErr
	}
	return caseDB, nil
}

func resolveHistoricalCaseDBPath() string {
	if configuredPath := strings.TrimSpace(os.Getenv(historicalCaseDBPathEnv)); configuredPath != "" {
		return configuredPath
	}

	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
		return filepath.Join(projectRoot, "DB", "historical_case_library.db")
	}

	workingDir, err := os.Getwd()
	if err == nil {
		return filepath.Join(workingDir, "DB", "historical_case_library.db")
	}
	return filepath.Join("DB", "historical_case_library.db")
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
