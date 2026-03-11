package user_history_index

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"antifraud/database"
	"antifraud/embedding"

	"gorm.io/gorm"
)

const (
	defaultUserHistoryTopK = 5
	maxUserHistoryTopK     = 20
)

// ArchiveInput 描述用户历史向量索引所需的最小归档信息。
type ArchiveInput struct {
	RecordID    string
	UserID      string
	Title       string
	CaseSummary string
	ScamType    string
	CreatedAt   time.Time
}

// IndexRecord 表示向量索引表中的一条记录。
type IndexRecord struct {
	RecordID           string
	UserID             string
	EmbeddingModel     string
	EmbeddingDimension int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// SimilarHistoryResult 表示一条按相似度召回的用户历史案件。
type SimilarHistoryResult struct {
	RecordID   string
	UserID     string
	Similarity float64
	CreatedAt  time.Time
}

type userHistoryVectorEntity struct {
	RecordID           string    `gorm:"primaryKey;size:64"`
	UserID             string    `gorm:"primaryKey;size:64;index;not null"`
	EmbeddingVector    string    `gorm:"type:text;not null"`
	EmbeddingModel     string    `gorm:"size:128;not null"`
	EmbeddingDimension int       `gorm:"not null"`
	CreatedAt          time.Time `gorm:"index;not null"`
	UpdatedAt          time.Time `gorm:"index;not null"`
}

func (userHistoryVectorEntity) TableName() string {
	return "user_history_vectors"
}

var (
	userHistorySchemaOnce sync.Once
	userHistorySchemaErr  error
)

func init() {
	database.RegisterMainDBSchemaInitializer("user_history_index", initUserHistoryVectorSchema)
}

// BuildEmbeddingInput 将用户历史归档字段拼接为 embedding 输入文本。
func BuildEmbeddingInput(input ArchiveInput) string {
	normalized := normalizeArchiveInput(input)
	segments := make([]string, 0, 3)
	segments = appendEmbeddingSegment(segments, "标题", normalized.Title)
	segments = appendEmbeddingSegment(segments, "案件摘要", normalized.CaseSummary)
	segments = appendEmbeddingSegment(segments, "诈骗类型", normalized.ScamType)
	return strings.Join(segments, "\n")
}

// UpsertHistoryVector 为一条已归档的用户历史写入或更新向量索引。
func UpsertHistoryVector(ctx context.Context, input ArchiveInput) (IndexRecord, error) {
	normalized := normalizeArchiveInput(input)
	if err := validateArchiveInput(normalized); err != nil {
		return IndexRecord{}, err
	}

	embeddingText := BuildEmbeddingInput(normalized)
	if strings.TrimSpace(embeddingText) == "" {
		return IndexRecord{}, fmt.Errorf("user history embedding input is empty")
	}

	vector, modelName, err := embedding.GenerateVector(ctx, embeddingText)
	if err != nil {
		return IndexRecord{}, fmt.Errorf("generate user history embedding failed: %w", err)
	}

	entity := entityFromArchiveInput(normalized, vector, modelName)
	db := database.DB
	if err := ensureUserHistoryVectorSchema(db); err != nil {
		return IndexRecord{}, err
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("record_id = ? AND user_id = ?", entity.RecordID, entity.UserID).Delete(&userHistoryVectorEntity{}).Error; err != nil {
			return err
		}
		return tx.Create(&entity).Error
	}); err != nil {
		return IndexRecord{}, fmt.Errorf("upsert user history vector failed: %w", err)
	}

	return indexRecordFromEntity(entity), nil
}

// SearchTopKSimilarHistoryByQuery 对当前用户历史案件执行向量召回。
func SearchTopKSimilarHistoryByQuery(ctx context.Context, userID, query string, topK int) ([]SimilarHistoryResult, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	queryVector, _, err := embedding.GenerateVector(ctx, trimmedQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("generate query embedding failed: %w", err)
	}
	return SearchTopKSimilarHistoryByVector(userID, queryVector, topK)
}

// SearchTopKSimilarHistoryByVector 使用调用方提供的 query 向量执行召回。
func SearchTopKSimilarHistoryByVector(userID string, queryVector []float64, topK int) ([]SimilarHistoryResult, int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, 0, fmt.Errorf("user_id is empty")
	}

	normalizedQuery, ok := normalizeL2Vector(queryVector)
	if !ok {
		return nil, 0, fmt.Errorf("query embedding vector is empty or invalid")
	}

	appliedTopK := normalizeTopK(topK)
	db := database.DB
	if err := ensureUserHistoryVectorSchema(db); err != nil {
		return nil, appliedTopK, err
	}

	rows := make([]userHistoryVectorEntity, 0)
	if err := db.Where("user_id = ?", normalizedUserID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, appliedTopK, fmt.Errorf("query user history vectors failed: %w", err)
	}
	if len(rows) == 0 {
		return []SimilarHistoryResult{}, appliedTopK, nil
	}

	results := make([]SimilarHistoryResult, 0, len(rows))
	for _, row := range rows {
		storedVector, err := decodeFloatList(row.EmbeddingVector)
		if err != nil {
			log.Printf("[user_history_index] decode vector failed: user=%s record=%s err=%v", normalizedUserID, strings.TrimSpace(row.RecordID), err)
			continue
		}

		normalizedStored, ok := normalizeL2Vector(storedVector)
		if !ok {
			continue
		}

		results = append(results, SimilarHistoryResult{
			RecordID:   strings.TrimSpace(row.RecordID),
			UserID:     strings.TrimSpace(row.UserID),
			Similarity: cosineSimilarityNormalized(normalizedQuery, normalizedStored),
			CreatedAt:  row.CreatedAt,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if math.Abs(results[i].Similarity-results[j].Similarity) > 1e-12 {
			return results[i].Similarity > results[j].Similarity
		}
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	if len(results) > appliedTopK {
		results = results[:appliedTopK]
	}
	return results, appliedTopK, nil
}

func ensureUserHistoryVectorSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	userHistorySchemaOnce.Do(func() {
		userHistorySchemaErr = initUserHistoryVectorSchema(db)
	})
	if userHistorySchemaErr != nil {
		return fmt.Errorf("auto migrate user history vector table failed: %w", userHistorySchemaErr)
	}
	return nil
}

func initUserHistoryVectorSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}
	return db.AutoMigrate(&userHistoryVectorEntity{})
}

func validateArchiveInput(input ArchiveInput) error {
	if input.RecordID == "" {
		return fmt.Errorf("record_id is empty")
	}
	if input.UserID == "" {
		return fmt.Errorf("user_id is empty")
	}
	if input.Title == "" && input.CaseSummary == "" && input.ScamType == "" {
		return fmt.Errorf("user history content is empty")
	}
	return nil
}

func normalizeArchiveInput(input ArchiveInput) ArchiveInput {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	title := strings.TrimSpace(input.Title)
	caseSummary := strings.TrimSpace(input.CaseSummary)
	if title == "" {
		title = caseSummary
	}
	return ArchiveInput{
		RecordID:    strings.TrimSpace(input.RecordID),
		UserID:      strings.TrimSpace(input.UserID),
		Title:       title,
		CaseSummary: caseSummary,
		ScamType:    strings.TrimSpace(input.ScamType),
		CreatedAt:   createdAt,
	}
}

func entityFromArchiveInput(input ArchiveInput, vector []float64, modelName string) userHistoryVectorEntity {
	cleanVector := sanitizeFloatList(vector)
	return userHistoryVectorEntity{
		RecordID:           input.RecordID,
		UserID:             input.UserID,
		EmbeddingVector:    encodeFloatList(cleanVector),
		EmbeddingModel:     strings.TrimSpace(modelName),
		EmbeddingDimension: len(cleanVector),
		CreatedAt:          input.CreatedAt,
		UpdatedAt:          time.Now(),
	}
}

func indexRecordFromEntity(entity userHistoryVectorEntity) IndexRecord {
	return IndexRecord{
		RecordID:           strings.TrimSpace(entity.RecordID),
		UserID:             strings.TrimSpace(entity.UserID),
		EmbeddingModel:     strings.TrimSpace(entity.EmbeddingModel),
		EmbeddingDimension: entity.EmbeddingDimension,
		CreatedAt:          entity.CreatedAt,
		UpdatedAt:          entity.UpdatedAt,
	}
}

func appendEmbeddingSegment(segments []string, key string, value string) []string {
	trimmedKey := strings.TrimSpace(key)
	trimmedValue := strings.TrimSpace(value)
	if trimmedKey == "" || trimmedValue == "" {
		return segments
	}
	return append(segments, trimmedKey+": "+trimmedValue)
}

func sanitizeFloatList(values []float64) []float64 {
	cleaned := make([]float64, 0, len(values))
	for _, value := range values {
		cleanValue := value
		if math.IsNaN(cleanValue) || math.IsInf(cleanValue, 0) {
			cleanValue = 0
		}
		cleaned = append(cleaned, cleanValue)
	}
	return cleaned
}

func encodeFloatList(values []float64) string {
	if len(values) == 0 {
		return "[]"
	}
	raw, err := json.Marshal(sanitizeFloatList(values))
	if err != nil {
		return "[]"
	}
	return string(raw)
}

func decodeFloatList(raw string) ([]float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []float64{}, nil
	}
	var values []float64
	if err := json.Unmarshal([]byte(trimmed), &values); err != nil {
		return nil, err
	}
	return sanitizeFloatList(values), nil
}

func normalizeTopK(topK int) int {
	if topK <= 0 {
		return defaultUserHistoryTopK
	}
	if topK > maxUserHistoryTopK {
		return maxUserHistoryTopK
	}
	return topK
}

func normalizeL2Vector(vec []float64) ([]float64, bool) {
	if len(vec) == 0 {
		return nil, false
	}

	normalized := make([]float64, 0, len(vec))
	var norm2 float64
	for _, value := range vec {
		cleanValue := value
		if math.IsNaN(cleanValue) || math.IsInf(cleanValue, 0) {
			cleanValue = 0
		}
		normalized = append(normalized, cleanValue)
		norm2 += cleanValue * cleanValue
	}
	if norm2 <= 0 {
		return nil, false
	}

	norm := math.Sqrt(norm2)
	for index := range normalized {
		normalized[index] = normalized[index] / norm
	}
	return normalized, true
}

func cosineSimilarityNormalized(left, right []float64) float64 {
	minLen := len(left)
	if len(right) < minLen {
		minLen = len(right)
	}
	if minLen == 0 {
		return 0
	}

	var dot float64
	for index := 0; index < minLen; index++ {
		dot += left[index] * right[index]
	}
	if dot > 1 {
		return 1
	}
	if dot < -1 {
		return -1
	}
	return dot
}
