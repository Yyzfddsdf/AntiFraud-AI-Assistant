package user_history_index

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"antifraud/internal/platform/database"

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
	return DefaultService().UpsertHistoryVector(ctx, input)
}

// SearchTopKSimilarHistoryByQuery 对当前用户历史案件执行向量召回。
func SearchTopKSimilarHistoryByQuery(ctx context.Context, userID, query string, topK int) ([]SimilarHistoryResult, int, error) {
	return DefaultService().SearchTopKSimilarHistoryByQuery(ctx, userID, query, topK)
}

// SearchTopKSimilarHistoryByVector 使用调用方提供的 query 向量执行召回。
func SearchTopKSimilarHistoryByVector(userID string, queryVector []float64, topK int) ([]SimilarHistoryResult, int, error) {
	return DefaultService().SearchTopKSimilarHistoryByVector(userID, queryVector, topK)
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
