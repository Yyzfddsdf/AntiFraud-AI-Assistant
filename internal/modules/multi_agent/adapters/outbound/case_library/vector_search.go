package case_library

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"antifraud/internal/platform/cache"
	"antifraud/internal/platform/database"
)

const (
	defaultSimilarCaseTopK = 5
	maxSimilarCaseTopK     = 20
)

const (
	historicalCaseVectorCacheHashKey  = "cache:case_library:vector_records"
	historicalCaseVectorCacheReadyKey = "cache:case_library:vector_records_ready"
)

var (
	historicalCaseVectorCacheReadyGet    = cache.GetJSON
	historicalCaseVectorCacheReadySet    = cache.SetJSON
	historicalCaseVectorCacheDelete      = cache.Delete
	historicalCaseVectorCacheHashGet     = cache.HashGetJSON
	historicalCaseVectorCacheHashGetAll  = cache.HashGetAll
	historicalCaseVectorCacheHashSetJSON = cache.HashSetJSON
	historicalCaseVectorCacheHashDelete  = cache.HashDelete
)

// SimilarCaseResult represents one ranked case from vector search.
type SimilarCaseResult struct {
	CaseID          string
	Title           string
	TargetGroup     string
	RiskLevel       string
	ScamType        string
	CaseDescription string
	Keywords        []string
	ViolatedLaw     string
	Similarity      float64
	CreatedAt       time.Time
}

// SimilarCaseRecallFilter defines optional exact-match filters for vector recall.
type SimilarCaseRecallFilter struct {
	TargetGroup string
	ScamType    string
}

// QueryAllHistoricalCases keeps the full database query behavior for non-search callers.
func QueryAllHistoricalCases() ([]HistoricalCaseRecord, error) {
	return queryAllHistoricalCasesFromDB()
}

// StreamAllHistoricalCases executes a streaming query to avoid loading all records into memory at once.
func StreamAllHistoricalCases(callback func(HistoricalCaseRecord) error) error {
	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return err
	}

	rows, err := db.Model(&historicalCaseEntity{}).Rows()
	if err != nil {
		return fmt.Errorf("stream historical cases failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entity historicalCaseEntity
		if err := db.ScanRows(rows, &entity); err != nil {
			return fmt.Errorf("scan historical case row failed: %w", err)
		}
		if err := callback(recordFromEntity(entity)); err != nil {
			return err
		}
	}
	return nil
}

func queryAllHistoricalCasesFromDB() ([]HistoricalCaseRecord, error) {
	db, err := database.GetHistoricalCaseDB()
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

// SearchTopKSimilarCasesByVector executes cosine similarity search from distributed Redis cache.
// Cache behavior:
// 1) service startup attempts to warm the Redis snapshot;
// 2) if cache is still not ready, the first search lazily loads all records from DB once;
// 3) after cache is loaded, writes are incrementally synced by create/delete paths.
func SearchTopKSimilarCasesByVector(queryVector []float64, topK int) ([]SimilarCaseResult, int, error) {
	return SearchTopKSimilarCasesByVectorWithFilter(queryVector, topK, SimilarCaseRecallFilter{})
}

// SearchTopKSimilarCasesByVectorWithConditions executes vector recall with optional
// target group and scam type restrictions.
func SearchTopKSimilarCasesByVectorWithConditions(queryVector []float64, topK int, targetGroup, scamType string) ([]SimilarCaseResult, int, error) {
	return SearchTopKSimilarCasesByVectorWithFilter(queryVector, topK, SimilarCaseRecallFilter{
		TargetGroup: targetGroup,
		ScamType:    scamType,
	})
}

// SearchTopKSimilarCasesByVectorWithFilter executes cosine similarity search from
// distributed Redis cache with optional exact-match filters.
func SearchTopKSimilarCasesByVectorWithFilter(queryVector []float64, topK int, filter SimilarCaseRecallFilter) ([]SimilarCaseResult, int, error) {
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

	results := collectSimilarCaseResults(normalizedQuery, cases, normalizeSimilarCaseRecallFilter(filter))
	sortSimilarCaseResults(results)
	if len(results) > appliedTopK {
		results = results[:appliedTopK]
	}
	return results, appliedTopK, nil
}

func collectSimilarCaseResults(normalizedQuery []float64, cases []HistoricalCaseRecord, filter SimilarCaseRecallFilter) []SimilarCaseResult {
	results := make([]SimilarCaseResult, 0, len(cases))
	for _, item := range cases {
		if !matchSimilarCaseRecallFilter(item, filter) {
			continue
		}

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
			ScamType:        item.ScamType,
			CaseDescription: item.CaseDescription,
			Keywords:        append([]string{}, item.Keywords...),
			ViolatedLaw:     item.ViolatedLaw,
			Similarity:      sim,
			CreatedAt:       item.CreatedAt,
		})
	}
	return results
}

func sortSimilarCaseResults(results []SimilarCaseResult) {
	sort.Slice(results, func(i, j int) bool {
		if math.Abs(results[i].Similarity-results[j].Similarity) > 1e-12 {
			return results[i].Similarity > results[j].Similarity
		}
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})
}

func normalizeSimilarCaseRecallFilter(filter SimilarCaseRecallFilter) SimilarCaseRecallFilter {
	return SimilarCaseRecallFilter{
		TargetGroup: strings.TrimSpace(filter.TargetGroup),
		ScamType:    strings.TrimSpace(filter.ScamType),
	}
}

func matchSimilarCaseRecallFilter(record HistoricalCaseRecord, filter SimilarCaseRecallFilter) bool {
	if filter.TargetGroup != "" && strings.TrimSpace(record.TargetGroup) != filter.TargetGroup {
		return false
	}
	if filter.ScamType != "" && strings.TrimSpace(record.ScamType) != filter.ScamType {
		return false
	}
	return true
}

func snapshotHistoricalCaseVectorCache() ([]HistoricalCaseRecord, error) {
	records, ready, err := loadHistoricalCaseVectorCacheFromRedis()
	if err != nil {
		log.Printf("[case_library] load vector cache from redis failed: %v", err)
	} else if ready {
		count, countErr := countHistoricalCasesFromDB()
		if countErr != nil {
			log.Printf("[case_library] count historical cases failed during cache validation: %v", countErr)
		} else if int(count) != len(records) {
			log.Printf("[case_library] detected vector cache drift: redis_records=%d db_records=%d, rebuilding cache", len(records), count)
			recordsFromDB, dbErr := queryAllHistoricalCasesFromDB()
			if dbErr != nil {
				return nil, dbErr
			}
			if cacheErr := replaceHistoricalCaseVectorCache(recordsFromDB); cacheErr != nil {
				log.Printf("[case_library] rebuild vector cache after drift detection failed: %v", cacheErr)
			}
			return cloneHistoricalCaseRecords(recordsFromDB), nil
		}
		return cloneHistoricalCaseRecords(records), nil
	}

	recordsFromDB, dbErr := queryAllHistoricalCasesFromDB()
	if dbErr != nil {
		// Redis 异常与 DB 异常不叠加传播，优先返回明确的 DB 错误。
		return nil, dbErr
	}

	if cacheErr := replaceHistoricalCaseVectorCache(recordsFromDB); cacheErr != nil {
		log.Printf("[case_library] refresh vector cache to redis failed: %v", cacheErr)
	}
	return cloneHistoricalCaseRecords(recordsFromDB), nil
}

func loadHistoricalCaseVectorRecordFromRedis(caseID string) (HistoricalCaseRecord, bool, error) {
	trimmedCaseID := strings.TrimSpace(caseID)
	if trimmedCaseID == "" {
		return HistoricalCaseRecord{}, false, nil
	}

	var ready bool
	found, err := historicalCaseVectorCacheReadyGet(historicalCaseVectorCacheReadyKey, &ready)
	if err != nil {
		return HistoricalCaseRecord{}, false, err
	}
	if !found || !ready {
		return HistoricalCaseRecord{}, false, nil
	}

	var record HistoricalCaseRecord
	found, err = historicalCaseVectorCacheHashGet(historicalCaseVectorCacheHashKey, trimmedCaseID, &record)
	if err != nil {
		return HistoricalCaseRecord{}, false, err
	}
	if !found {
		return HistoricalCaseRecord{}, false, nil
	}
	if strings.TrimSpace(record.CaseID) == "" {
		record.CaseID = trimmedCaseID
	}
	return cloneHistoricalCaseRecord(record), true, nil
}

func countHistoricalCasesFromDB() (int64, error) {
	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		return 0, err
	}

	var count int64
	if err := db.Model(&historicalCaseEntity{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count historical cases failed: %w", err)
	}
	return count, nil
}

func loadHistoricalCaseVectorCacheFromRedis() ([]HistoricalCaseRecord, bool, error) {
	var ready bool
	readyFound, err := historicalCaseVectorCacheReadyGet(historicalCaseVectorCacheReadyKey, &ready)
	if err != nil {
		return nil, false, err
	}

	values, err := historicalCaseVectorCacheHashGetAll(historicalCaseVectorCacheHashKey)
	if err != nil {
		return nil, false, err
	}
	if len(values) == 0 {
		return []HistoricalCaseRecord{}, readyFound && ready, nil
	}

	records := make([]HistoricalCaseRecord, 0, len(values))
	for caseID, raw := range values {
		var item HistoricalCaseRecord
		if err := json.Unmarshal([]byte(raw), &item); err != nil {
			return nil, false, fmt.Errorf("decode redis vector cache failed: case_id=%s err=%w", strings.TrimSpace(caseID), err)
		}
		records = append(records, cloneHistoricalCaseRecord(item))
	}
	return records, true, nil
}

func replaceHistoricalCaseVectorCache(records []HistoricalCaseRecord) error {
	if err := historicalCaseVectorCacheDelete(historicalCaseVectorCacheHashKey); err != nil {
		return err
	}

	for _, item := range records {
		trimmedCaseID := strings.TrimSpace(item.CaseID)
		if trimmedCaseID == "" {
			continue
		}
		normalized := cloneHistoricalCaseRecord(item)
		normalized.CaseID = trimmedCaseID
		if err := historicalCaseVectorCacheHashSetJSON(historicalCaseVectorCacheHashKey, trimmedCaseID, normalized); err != nil {
			return err
		}
	}

	if err := historicalCaseVectorCacheReadySet(historicalCaseVectorCacheReadyKey, true, 0); err != nil {
		return err
	}
	return nil
}

// upsertHistoricalCaseVectorCache incrementally updates one record in Redis.
func upsertHistoricalCaseVectorCache(record HistoricalCaseRecord) {
	trimmedCaseID := strings.TrimSpace(record.CaseID)
	if trimmedCaseID == "" {
		return
	}

	if _, err := ensureHistoricalCaseVectorCacheReady(); err != nil {
		log.Printf("[case_library] ensure vector cache ready before upsert failed: case_id=%s err=%v", trimmedCaseID, err)
		return
	}

	normalized := cloneHistoricalCaseRecord(record)
	normalized.CaseID = trimmedCaseID
	if err := historicalCaseVectorCacheHashSetJSON(historicalCaseVectorCacheHashKey, trimmedCaseID, normalized); err != nil {
		log.Printf("[case_library] upsert vector cache failed: case_id=%s err=%v", trimmedCaseID, err)
		return
	}
	if err := historicalCaseVectorCacheReadySet(historicalCaseVectorCacheReadyKey, true, 0); err != nil {
		log.Printf("[case_library] mark vector cache ready failed: case_id=%s err=%v", trimmedCaseID, err)
	}
}

// removeHistoricalCaseVectorCache incrementally removes one record from Redis.
func removeHistoricalCaseVectorCache(caseID string) {
	trimmedCaseID := strings.TrimSpace(caseID)
	if trimmedCaseID == "" {
		return
	}

	if _, err := ensureHistoricalCaseVectorCacheReady(); err != nil {
		log.Printf("[case_library] ensure vector cache ready before remove failed: case_id=%s err=%v", trimmedCaseID, err)
		return
	}

	if err := historicalCaseVectorCacheHashDelete(historicalCaseVectorCacheHashKey, trimmedCaseID); err != nil {
		log.Printf("[case_library] remove vector cache failed: case_id=%s err=%v", trimmedCaseID, err)
		return
	}
	if err := historicalCaseVectorCacheReadySet(historicalCaseVectorCacheReadyKey, true, 0); err != nil {
		log.Printf("[case_library] mark vector cache ready failed: case_id=%s err=%v", trimmedCaseID, err)
	}
}

func ensureHistoricalCaseVectorCacheReady() (bool, error) {
	var ready bool
	found, err := historicalCaseVectorCacheReadyGet(historicalCaseVectorCacheReadyKey, &ready)
	if err != nil {
		return false, fmt.Errorf("read vector cache ready flag failed: %w", err)
	}
	if found && ready {
		return true, nil
	}

	records, err := queryAllHistoricalCasesFromDB()
	if err != nil {
		return false, fmt.Errorf("query all cases for cache warmup failed: %w", err)
	}
	if err := replaceHistoricalCaseVectorCache(records); err != nil {
		return false, fmt.Errorf("warmup vector cache failed: %w", err)
	}
	return true, nil
}

func WarmupHistoricalCaseVectorCache() error {
	report, err := SelfCheckHistoricalCaseVectorCache(true)
	if err != nil {
		return err
	}
	if report.Repaired {
		log.Printf("[case_library] repaired vector cache during warmup: redis=%d db=%d missing=%d stale=%d mismatched=%d",
			report.RedisCount,
			report.DBCount,
			len(report.MissingCaseIDs),
			len(report.StaleCaseIDs),
			len(report.MismatchedCaseIDs),
		)
	}
	return nil
}

func cloneHistoricalCaseRecords(records []HistoricalCaseRecord) []HistoricalCaseRecord {
	cloned := make([]HistoricalCaseRecord, 0, len(records))
	for _, item := range records {
		cloned = append(cloned, cloneHistoricalCaseRecord(item))
	}
	return cloned
}

func cloneHistoricalCaseRecord(record HistoricalCaseRecord) HistoricalCaseRecord {
	return HistoricalCaseRecord{
		CaseID:             strings.TrimSpace(record.CaseID),
		CreatedBy:          strings.TrimSpace(record.CreatedBy),
		Title:              strings.TrimSpace(record.Title),
		TargetGroup:        strings.TrimSpace(record.TargetGroup),
		RiskLevel:          strings.TrimSpace(record.RiskLevel),
		ScamType:           strings.TrimSpace(record.ScamType),
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
