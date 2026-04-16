package case_library

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type HistoricalCaseVectorCacheSelfCheckResult struct {
	CheckedAt         time.Time
	Ready             bool
	Healthy           bool
	RepairRequested   bool
	Repaired          bool
	DBCount           int
	RedisCount        int
	MissingCaseIDs    []string
	StaleCaseIDs      []string
	MismatchedCaseIDs []string
}

func SelfCheckHistoricalCaseVectorCache(repair bool) (HistoricalCaseVectorCacheSelfCheckResult, error) {
	dbRecords, err := queryAllHistoricalCasesFromDB()
	if err != nil {
		return HistoricalCaseVectorCacheSelfCheckResult{}, err
	}

	redisRecords, ready, err := loadHistoricalCaseVectorCacheFromRedis()
	if err != nil {
		return HistoricalCaseVectorCacheSelfCheckResult{}, fmt.Errorf("load historical case vector cache failed: %w", err)
	}

	dbIndex := buildHistoricalCaseRecordIndex(dbRecords)
	redisIndex := buildHistoricalCaseRecordIndex(redisRecords)
	missingCaseIDs, staleCaseIDs, mismatchedCaseIDs := diffHistoricalCaseRecordIndex(dbIndex, redisIndex)
	healthy := ready && len(missingCaseIDs) == 0 && len(staleCaseIDs) == 0 && len(mismatchedCaseIDs) == 0

	result := HistoricalCaseVectorCacheSelfCheckResult{
		CheckedAt:         time.Now(),
		Ready:             ready,
		Healthy:           healthy,
		RepairRequested:   repair,
		DBCount:           len(dbRecords),
		RedisCount:        len(redisRecords),
		MissingCaseIDs:    missingCaseIDs,
		StaleCaseIDs:      staleCaseIDs,
		MismatchedCaseIDs: mismatchedCaseIDs,
	}

	if healthy || !repair {
		return result, nil
	}

	if err := replaceHistoricalCaseVectorCache(dbRecords); err != nil {
		return result, fmt.Errorf("repair historical case vector cache failed: %w", err)
	}

	result.Repaired = true
	return result, nil
}

func buildHistoricalCaseRecordIndex(records []HistoricalCaseRecord) map[string]HistoricalCaseRecord {
	index := make(map[string]HistoricalCaseRecord, len(records))
	for _, record := range records {
		normalized := cloneHistoricalCaseRecord(record)
		if normalized.CaseID == "" {
			continue
		}
		index[normalized.CaseID] = normalized
	}
	return index
}

func diffHistoricalCaseRecordIndex(dbIndex map[string]HistoricalCaseRecord, redisIndex map[string]HistoricalCaseRecord) ([]string, []string, []string) {
	missingCaseIDs := make([]string, 0)
	staleCaseIDs := make([]string, 0)
	mismatchedCaseIDs := make([]string, 0)

	for caseID, dbRecord := range dbIndex {
		redisRecord, exists := redisIndex[caseID]
		if !exists {
			missingCaseIDs = append(missingCaseIDs, caseID)
			continue
		}
		if !historicalCaseRecordsEqual(dbRecord, redisRecord) {
			mismatchedCaseIDs = append(mismatchedCaseIDs, caseID)
		}
	}

	for caseID := range redisIndex {
		if _, exists := dbIndex[caseID]; !exists {
			staleCaseIDs = append(staleCaseIDs, caseID)
		}
	}

	sort.Strings(missingCaseIDs)
	sort.Strings(staleCaseIDs)
	sort.Strings(mismatchedCaseIDs)
	return missingCaseIDs, staleCaseIDs, mismatchedCaseIDs
}

func historicalCaseRecordsEqual(left HistoricalCaseRecord, right HistoricalCaseRecord) bool {
	leftNormalized := cloneHistoricalCaseRecord(left)
	rightNormalized := cloneHistoricalCaseRecord(right)

	if leftNormalized.CaseID != rightNormalized.CaseID ||
		leftNormalized.CreatedBy != rightNormalized.CreatedBy ||
		leftNormalized.Title != rightNormalized.Title ||
		leftNormalized.TargetGroup != rightNormalized.TargetGroup ||
		leftNormalized.RiskLevel != rightNormalized.RiskLevel ||
		leftNormalized.ScamType != rightNormalized.ScamType ||
		leftNormalized.CaseDescription != rightNormalized.CaseDescription ||
		leftNormalized.ViolatedLaw != rightNormalized.ViolatedLaw ||
		leftNormalized.Suggestion != rightNormalized.Suggestion ||
		leftNormalized.EmbeddingModel != rightNormalized.EmbeddingModel ||
		leftNormalized.EmbeddingDimension != rightNormalized.EmbeddingDimension {
		return false
	}

	if !equalStringSlices(leftNormalized.TypicalScripts, rightNormalized.TypicalScripts) ||
		!equalStringSlices(leftNormalized.Keywords, rightNormalized.Keywords) ||
		!equalFloatSlices(leftNormalized.EmbeddingVector, rightNormalized.EmbeddingVector) {
		return false
	}

	return equalTimeValue(leftNormalized.CreatedAt, rightNormalized.CreatedAt) &&
		equalTimeValue(leftNormalized.UpdatedAt, rightNormalized.UpdatedAt)
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if strings.TrimSpace(left[index]) != strings.TrimSpace(right[index]) {
			return false
		}
	}
	return true
}

func equalFloatSlices(left []float64, right []float64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalTimeValue(left time.Time, right time.Time) bool {
	if left.IsZero() && right.IsZero() {
		return true
	}
	return left.UTC().UnixNano() == right.UTC().UnixNano()
}
