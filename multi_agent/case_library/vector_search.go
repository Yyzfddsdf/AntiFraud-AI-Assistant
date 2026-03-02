package case_library

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	defaultSimilarCaseTopK = 5
	maxSimilarCaseTopK     = 20
)

// SimilarCaseResult represents one ranked case from vector search.
type SimilarCaseResult struct {
	CaseID          string
	Title           string
	TargetGroup     string
	RiskLevel       string
	CaseDescription string
	Keywords        []string
	ViolatedLaw     string
	Similarity      float64
	CreatedAt       time.Time
}

type historicalCaseVectorCache struct {
	mu             sync.RWMutex
	loaded         bool
	byID           map[string]HistoricalCaseRecord
	list           []HistoricalCaseRecord
	pendingUpserts map[string]HistoricalCaseRecord
	pendingDeletes map[string]struct{}
}

var caseVectorCache = historicalCaseVectorCache{
	byID:           map[string]HistoricalCaseRecord{},
	list:           []HistoricalCaseRecord{},
	pendingUpserts: map[string]HistoricalCaseRecord{},
	pendingDeletes: map[string]struct{}{},
}

// QueryAllHistoricalCases keeps the full database query behavior for non-search callers.
func QueryAllHistoricalCases() ([]HistoricalCaseRecord, error) {
	return queryAllHistoricalCasesFromDB()
}

func queryAllHistoricalCasesFromDB() ([]HistoricalCaseRecord, error) {
	db, err := getHistoricalCaseDB()
	if err != nil {
		return nil, err
	}

	rows := make([]historicalCaseEntity, 0)
	if err := db.Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query all historical cases failed: %w", err)
	}

	records := make([]HistoricalCaseRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, recordFromEntity(row))
	}
	return records, nil
}

// SearchTopKSimilarCasesByVector executes cosine similarity search from in-memory cache.
// Cache behavior:
// 1) first search lazily loads all records from DB once;
// 2) after cache is loaded, writes are incrementally synced by create/delete paths.
func SearchTopKSimilarCasesByVector(queryVector []float64, topK int) ([]SimilarCaseResult, int, error) {
	normalizedQuery, ok := normalizeL2Vector(queryVector)
	if !ok {
		return nil, 0, fmt.Errorf("query embedding vector is empty or invalid")
	}

	cases, err := snapshotHistoricalCaseVectorCache()
	if err != nil {
		return nil, 0, err
	}

	appliedTopK := normalizeTopK(topK)
	if len(cases) == 0 {
		return []SimilarCaseResult{}, appliedTopK, nil
	}

	results := make([]SimilarCaseResult, 0, len(cases))
	for _, item := range cases {
		normalizedCaseVector, ok := normalizeL2Vector(item.EmbeddingVector)
		if !ok {
			continue
		}

		sim := cosineSimilarityNormalized(normalizedQuery, normalizedCaseVector)
		results = append(results, SimilarCaseResult{
			CaseID:          item.CaseID,
			Title:           item.Title,
			TargetGroup:     item.TargetGroup,
			RiskLevel:       item.RiskLevel,
			CaseDescription: item.CaseDescription,
			Keywords:        append([]string{}, item.Keywords...),
			ViolatedLaw:     item.ViolatedLaw,
			Similarity:      sim,
			CreatedAt:       item.CreatedAt,
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

func snapshotHistoricalCaseVectorCache() ([]HistoricalCaseRecord, error) {
	if err := ensureHistoricalCaseVectorCacheLoaded(); err != nil {
		return nil, err
	}

	caseVectorCache.mu.RLock()
	defer caseVectorCache.mu.RUnlock()

	snapshot := make([]HistoricalCaseRecord, 0, len(caseVectorCache.list))
	for _, item := range caseVectorCache.list {
		snapshot = append(snapshot, cloneHistoricalCaseRecord(item))
	}
	return snapshot, nil
}

func ensureHistoricalCaseVectorCacheLoaded() error {
	caseVectorCache.mu.RLock()
	loaded := caseVectorCache.loaded
	caseVectorCache.mu.RUnlock()
	if loaded {
		return nil
	}

	records, err := queryAllHistoricalCasesFromDB()
	if err != nil {
		return err
	}

	caseVectorCache.mu.Lock()
	defer caseVectorCache.mu.Unlock()

	if caseVectorCache.loaded {
		return nil
	}
	caseVectorCache.byID = make(map[string]HistoricalCaseRecord, len(records))
	caseVectorCache.list = make([]HistoricalCaseRecord, 0, len(records))
	for _, item := range records {
		upsertHistoricalCaseVectorCacheUnsafe(item)
	}

	for caseID := range caseVectorCache.pendingDeletes {
		removeHistoricalCaseVectorCacheUnsafe(caseID)
	}
	for _, item := range caseVectorCache.pendingUpserts {
		upsertHistoricalCaseVectorCacheUnsafe(item)
	}
	caseVectorCache.pendingUpserts = map[string]HistoricalCaseRecord{}
	caseVectorCache.pendingDeletes = map[string]struct{}{}
	caseVectorCache.loaded = true
	return nil
}

// upsertHistoricalCaseVectorCache incrementally updates one record in memory.
// It only runs after cache is already loaded; before that, first search will full-load from DB.
func upsertHistoricalCaseVectorCache(record HistoricalCaseRecord) {
	trimmedCaseID := strings.TrimSpace(record.CaseID)
	if trimmedCaseID == "" {
		return
	}

	caseVectorCache.mu.Lock()
	defer caseVectorCache.mu.Unlock()
	if !caseVectorCache.loaded {
		delete(caseVectorCache.pendingDeletes, trimmedCaseID)
		caseVectorCache.pendingUpserts[trimmedCaseID] = cloneHistoricalCaseRecord(record)
		return
	}
	upsertHistoricalCaseVectorCacheUnsafe(record)
}

func upsertHistoricalCaseVectorCacheUnsafe(record HistoricalCaseRecord) {
	trimmedCaseID := strings.TrimSpace(record.CaseID)
	if trimmedCaseID == "" {
		return
	}

	normalizedRecord := cloneHistoricalCaseRecord(record)
	normalizedRecord.CaseID = trimmedCaseID
	caseVectorCache.byID[trimmedCaseID] = normalizedRecord

	for index := range caseVectorCache.list {
		if strings.TrimSpace(caseVectorCache.list[index].CaseID) == trimmedCaseID {
			caseVectorCache.list[index] = normalizedRecord
			return
		}
	}
	caseVectorCache.list = append(caseVectorCache.list, normalizedRecord)
}

// removeHistoricalCaseVectorCache incrementally removes one record in memory.
func removeHistoricalCaseVectorCache(caseID string) {
	trimmedCaseID := strings.TrimSpace(caseID)
	if trimmedCaseID == "" {
		return
	}

	caseVectorCache.mu.Lock()
	defer caseVectorCache.mu.Unlock()
	if !caseVectorCache.loaded {
		delete(caseVectorCache.pendingUpserts, trimmedCaseID)
		caseVectorCache.pendingDeletes[trimmedCaseID] = struct{}{}
		return
	}

	removeHistoricalCaseVectorCacheUnsafe(trimmedCaseID)
}

func removeHistoricalCaseVectorCacheUnsafe(trimmedCaseID string) {
	if _, exists := caseVectorCache.byID[trimmedCaseID]; !exists {
		return
	}
	delete(caseVectorCache.byID, trimmedCaseID)

	for index := range caseVectorCache.list {
		if strings.TrimSpace(caseVectorCache.list[index].CaseID) == trimmedCaseID {
			caseVectorCache.list = append(caseVectorCache.list[:index], caseVectorCache.list[index+1:]...)
			break
		}
	}
}

func cloneHistoricalCaseRecord(record HistoricalCaseRecord) HistoricalCaseRecord {
	return HistoricalCaseRecord{
		CaseID:             strings.TrimSpace(record.CaseID),
		CreatedBy:          strings.TrimSpace(record.CreatedBy),
		Title:              strings.TrimSpace(record.Title),
		TargetGroup:        strings.TrimSpace(record.TargetGroup),
		RiskLevel:          strings.TrimSpace(record.RiskLevel),
		CaseDescription:    strings.TrimSpace(record.CaseDescription),
		TypicalScripts:     append([]string{}, record.TypicalScripts...),
		Keywords:           append([]string{}, record.Keywords...),
		ViolatedLaw:        strings.TrimSpace(record.ViolatedLaw),
		Suggestion:         strings.TrimSpace(record.Suggestion),
		EmbeddingVector:    append([]float64{}, record.EmbeddingVector...),
		EmbeddingModel:     strings.TrimSpace(record.EmbeddingModel),
		EmbeddingDimension: record.EmbeddingDimension,
		CreatedAt:          record.CreatedAt,
		UpdatedAt:          record.UpdatedAt,
	}
}

func normalizeTopK(topK int) int {
	if topK <= 0 {
		return defaultSimilarCaseTopK
	}
	if topK > maxSimilarCaseTopK {
		return maxSimilarCaseTopK
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
	for i := range normalized {
		normalized[i] = normalized[i] / norm
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
	for i := 0; i < minLen; i++ {
		dot += left[i] * right[i]
	}

	if dot > 1 {
		return 1
	}
	if dot < -1 {
		return -1
	}
	return dot
}
