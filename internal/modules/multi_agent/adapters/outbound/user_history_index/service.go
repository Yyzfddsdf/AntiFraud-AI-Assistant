package user_history_index

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"antifraud/internal/platform/database"
	"antifraud/internal/platform/embedding"

	"gorm.io/gorm"
)

// VectorGenerator 定义向量生成端口。
type VectorGenerator interface {
	Generate(ctx context.Context, input string) ([]float64, string, error)
}

// Repository 定义用户历史向量索引仓储端口。
type Repository interface {
	EnsureSchema() error
	Save(entity userHistoryVectorEntity) error
	ListByUser(userID string) ([]userHistoryVectorEntity, error)
}

// Service 编排用户历史向量索引。
type Service struct {
	repo      Repository
	vectorGen VectorGenerator
}

func NewService(repo Repository, vectorGen VectorGenerator) *Service {
	if repo == nil {
		repo = &gormRepository{db: database.DB}
	}
	if vectorGen == nil {
		vectorGen = embeddingVectorGenerator{}
	}
	return &Service{repo: repo, vectorGen: vectorGen}
}

func DefaultService() *Service {
	return NewService(nil, nil)
}

func (s *Service) UpsertHistoryVector(ctx context.Context, input ArchiveInput) (IndexRecord, error) {
	normalized := normalizeArchiveInput(input)
	if err := validateArchiveInput(normalized); err != nil {
		return IndexRecord{}, err
	}

	embeddingText := BuildEmbeddingInput(normalized)
	if strings.TrimSpace(embeddingText) == "" {
		return IndexRecord{}, fmt.Errorf("user history embedding input is empty")
	}

	vector, modelName, err := s.vectorGen.Generate(ctx, embeddingText)
	if err != nil {
		return IndexRecord{}, fmt.Errorf("generate user history embedding failed: %w", err)
	}

	entity := entityFromArchiveInput(normalized, vector, modelName)
	if err := s.repo.EnsureSchema(); err != nil {
		return IndexRecord{}, err
	}
	if err := s.repo.Save(entity); err != nil {
		return IndexRecord{}, err
	}
	return indexRecordFromEntity(entity), nil
}

func (s *Service) SearchTopKSimilarHistoryByQuery(ctx context.Context, userID, query string, topK int) ([]SimilarHistoryResult, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}
	queryVector, _, err := s.vectorGen.Generate(ctx, trimmedQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("generate query embedding failed: %w", err)
	}
	return s.SearchTopKSimilarHistoryByVector(userID, queryVector, topK)
}

func (s *Service) SearchTopKSimilarHistoryByVector(userID string, queryVector []float64, topK int) ([]SimilarHistoryResult, int, error) {
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		return nil, 0, fmt.Errorf("user_id is empty")
	}
	normalizedQuery, ok := normalizeL2Vector(queryVector)
	if !ok {
		return nil, 0, fmt.Errorf("query embedding vector is empty or invalid")
	}

	appliedTopK := normalizeTopK(topK)
	if err := s.repo.EnsureSchema(); err != nil {
		return nil, appliedTopK, err
	}
	rows, err := s.repo.ListByUser(normalizedUserID)
	if err != nil {
		return nil, appliedTopK, err
	}
	results := scoreSimilarities(normalizedUserID, normalizedQuery, rows)
	if len(results) > appliedTopK {
		results = results[:appliedTopK]
	}
	return results, appliedTopK, nil
}

type embeddingVectorGenerator struct{}

func (embeddingVectorGenerator) Generate(ctx context.Context, input string) ([]float64, string, error) {
	return embedding.GenerateVector(ctx, input)
}

type gormRepository struct {
	db *gorm.DB
}

func (r *gormRepository) EnsureSchema() error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return ensureUserHistoryVectorSchema(r.db)
}

func (r *gormRepository) Save(entity userHistoryVectorEntity) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("database not initialized")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("record_id = ? AND user_id = ?", entity.RecordID, entity.UserID).Delete(&userHistoryVectorEntity{}).Error; err != nil {
			return err
		}
		return tx.Create(&entity).Error
	})
}

func (r *gormRepository) ListByUser(userID string) ([]userHistoryVectorEntity, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	rows := make([]userHistoryVectorEntity, 0)
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query user history vectors failed: %w", err)
	}
	return rows, nil
}

func scoreSimilarities(normalizedUserID string, normalizedQuery []float64, rows []userHistoryVectorEntity) []SimilarHistoryResult {
	results := make([]SimilarHistoryResult, 0, len(rows))
	for _, row := range rows {
		storedVector, err := decodeFloatList(row.EmbeddingVector)
		if err != nil {
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
	sortSimilarResults(results)
	return results
}

func sortSimilarResults(results []SimilarHistoryResult) {
	sort.Slice(results, func(i, j int) bool {
		if math.Abs(results[i].Similarity-results[j].Similarity) > 1e-12 {
			return results[i].Similarity > results[j].Similarity
		}
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})
}
