package case_library_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	case_library "antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
)

func TestSelfCheckHistoricalCaseVectorCacheDetectsDrift(t *testing.T) {
	resetHistoricalCaseDB()
	dbPath, err := prepareHistoricalCaseDBPath()
	if err != nil {
		t.Fatalf("prepare historical case db path failed: %v", err)
	}
	t.Setenv("HISTORICAL_CASE_DB_PATH", dbPath)
	t.Cleanup(func() {
		resetHistoricalCaseDB()
		_ = os.Remove(dbPath)
	})

	originalGenerateCaseEmbedding := generateCaseEmbedding
	originalSearchHistoricalCases := searchHistoricalCasesByVector
	originalReadyGet := historicalCaseVectorCacheReadyGet
	originalReadySet := historicalCaseVectorCacheReadySet
	originalHashSet := historicalCaseVectorCacheHashSetJSON
	originalHashGetAll := historicalCaseVectorCacheHashGetAll
	originalDelete := historicalCaseVectorCacheDelete
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
		historicalCaseVectorCacheReadyGet = originalReadyGet
		historicalCaseVectorCacheReadySet = originalReadySet
		historicalCaseVectorCacheHashSetJSON = originalHashSet
		historicalCaseVectorCacheHashGetAll = originalHashGetAll
		historicalCaseVectorCacheDelete = originalDelete
	})

	generateCaseEmbedding = func(context.Context, string) ([]float64, string, error) {
		return []float64{0.1, 0.2, 0.3}, "mock-self-check", nil
	}
	searchHistoricalCasesByVector = func([]float64, int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{}, 1, nil
	}
	historicalCaseVectorCacheReadyGet = func(_ string, out interface{}) (bool, error) {
		receiver, ok := out.(*bool)
		if !ok {
			t.Fatalf("unexpected ready receiver type: %T", out)
		}
		*receiver = true
		return true, nil
	}
	historicalCaseVectorCacheReadySet = func(string, interface{}, time.Duration) error {
		return nil
	}
	historicalCaseVectorCacheHashSetJSON = func(string, string, interface{}) error {
		return nil
	}
	historicalCaseVectorCacheDelete = func(string) error {
		return nil
	}

	record, err := case_library.CreateHistoricalCase(context.Background(), "admin-self-check", case_library.CreateHistoricalCaseInput{
		Title:           "缓存自检检测缺失案件",
		TargetGroup:     "老人",
		RiskLevel:       "高",
		ScamType:        "冒充客服类",
		CaseDescription: "案件库自检需要发现 Redis 中遗漏的历史案件和多余的脏数据。",
	})
	if err != nil {
		t.Fatalf("create historical case failed: %v", err)
	}

	staleRecord := case_library.HistoricalCaseRecord{
		CaseID:             "HCASE-STALE",
		CreatedBy:          "redis-only",
		Title:              "Redis 脏缓存",
		TargetGroup:        "老人",
		RiskLevel:          "中",
		ScamType:           "冒充客服类",
		CaseDescription:    "这条记录只存在于缓存中。",
		EmbeddingVector:    []float64{0.7, 0.8, 0.9},
		EmbeddingModel:     "stale-model",
		EmbeddingDimension: 3,
	}
	stalePayload, err := json.Marshal(staleRecord)
	if err != nil {
		t.Fatalf("marshal stale record failed: %v", err)
	}

	historicalCaseVectorCacheHashGetAll = func(string) (map[string]string, error) {
		return map[string]string{
			staleRecord.CaseID: string(stalePayload),
		}, nil
	}

	report, err := case_library.SelfCheckHistoricalCaseVectorCache(false)
	if err != nil {
		t.Fatalf("self check historical case vector cache failed: %v", err)
	}
	if report.Healthy {
		t.Fatal("expected unhealthy vector cache report")
	}
	if len(report.MissingCaseIDs) != 1 || report.MissingCaseIDs[0] != record.CaseID {
		t.Fatalf("unexpected missing case ids: %+v", report.MissingCaseIDs)
	}
	if len(report.StaleCaseIDs) != 1 || report.StaleCaseIDs[0] != staleRecord.CaseID {
		t.Fatalf("unexpected stale case ids: %+v", report.StaleCaseIDs)
	}
	if report.Repaired {
		t.Fatal("self check without repair should not report repaired")
	}
}

func TestSelfCheckHistoricalCaseVectorCacheRepairsRedisSnapshot(t *testing.T) {
	resetHistoricalCaseDB()
	dbPath, err := prepareHistoricalCaseDBPath()
	if err != nil {
		t.Fatalf("prepare historical case db path failed: %v", err)
	}
	t.Setenv("HISTORICAL_CASE_DB_PATH", dbPath)
	t.Cleanup(func() {
		resetHistoricalCaseDB()
		_ = os.Remove(dbPath)
	})

	originalGenerateCaseEmbedding := generateCaseEmbedding
	originalSearchHistoricalCases := searchHistoricalCasesByVector
	originalReadyGet := historicalCaseVectorCacheReadyGet
	originalReadySet := historicalCaseVectorCacheReadySet
	originalHashSet := historicalCaseVectorCacheHashSetJSON
	originalHashGetAll := historicalCaseVectorCacheHashGetAll
	originalDelete := historicalCaseVectorCacheDelete
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
		historicalCaseVectorCacheReadyGet = originalReadyGet
		historicalCaseVectorCacheReadySet = originalReadySet
		historicalCaseVectorCacheHashSetJSON = originalHashSet
		historicalCaseVectorCacheHashGetAll = originalHashGetAll
		historicalCaseVectorCacheDelete = originalDelete
	})

	generateCaseEmbedding = func(context.Context, string) ([]float64, string, error) {
		return []float64{0.2, 0.3, 0.4}, "mock-self-check-repair", nil
	}
	searchHistoricalCasesByVector = func([]float64, int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{}, 1, nil
	}
	historicalCaseVectorCacheReadyGet = func(_ string, out interface{}) (bool, error) {
		receiver, ok := out.(*bool)
		if !ok {
			t.Fatalf("unexpected ready receiver type: %T", out)
		}
		*receiver = true
		return true, nil
	}
	historicalCaseVectorCacheReadySet = func(string, interface{}, time.Duration) error {
		return nil
	}
	historicalCaseVectorCacheHashSetJSON = func(string, string, interface{}) error {
		return nil
	}
	historicalCaseVectorCacheDelete = func(string) error {
		return nil
	}

	record, err := case_library.CreateHistoricalCase(context.Background(), "admin-repair", case_library.CreateHistoricalCaseInput{
		Title:           "缓存自检修复",
		TargetGroup:     "老人",
		RiskLevel:       "高",
		ScamType:        "冒充客服类",
		CaseDescription: "缓存重建需要用数据库中的真实案件覆盖 Redis 脏数据。",
	})
	if err != nil {
		t.Fatalf("create historical case failed: %v", err)
	}

	historicalCaseVectorCacheHashGetAll = func(string) (map[string]string, error) {
		return map[string]string{}, nil
	}

	deleteCalls := 0
	setFields := make([]string, 0, 1)
	historicalCaseVectorCacheDelete = func(string) error {
		deleteCalls++
		return nil
	}
	historicalCaseVectorCacheHashSetJSON = func(_ string, field string, value interface{}) error {
		_, ok := value.(case_library.HistoricalCaseRecord)
		if !ok {
			t.Fatalf("unexpected cache value type: %T", value)
		}
		setFields = append(setFields, field)
		return nil
	}

	report, err := case_library.SelfCheckHistoricalCaseVectorCache(true)
	if err != nil {
		t.Fatalf("repair historical case vector cache failed: %v", err)
	}
	if !report.Repaired {
		t.Fatal("expected repair to be executed")
	}
	if deleteCalls == 0 {
		t.Fatal("expected cache delete during repair")
	}
	if len(setFields) != 1 || setFields[0] != record.CaseID {
		t.Fatalf("unexpected repaired cache fields: %+v", setFields)
	}
}
