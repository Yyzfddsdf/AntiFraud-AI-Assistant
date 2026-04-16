package case_library_test

import (
	"context"
	"os"
	"testing"
	"time"

	case_library "antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
)

func TestGetHistoricalCaseByID_UsesVectorCacheWhenAvailable(t *testing.T) {
	originalReadyGet := historicalCaseVectorCacheReadyGet
	originalHashGet := historicalCaseVectorCacheHashGet
	t.Cleanup(func() {
		historicalCaseVectorCacheReadyGet = originalReadyGet
		historicalCaseVectorCacheHashGet = originalHashGet
	})

	cached := case_library.HistoricalCaseRecord{
		CaseID:             "HCASE-CACHED",
		CreatedBy:          "admin-cache",
		Title:              "缓存中的历史案件",
		TargetGroup:        "老年人",
		RiskLevel:          "高",
		ScamType:           "冒充电商物流客服类",
		CaseDescription:    "命中详情缓存后不应再回源数据库。",
		TypicalScripts:     []string{"我是客服", "现在给您退款"},
		Keywords:           []string{"退款", "客服"},
		ViolatedLaw:        "诈骗罪",
		Suggestion:         "立即止付并报警",
		EmbeddingVector:    []float64{0.1, 0.2, 0.3},
		EmbeddingModel:     "cached-model",
		EmbeddingDimension: 3,
	}

	historicalCaseVectorCacheReadyGet = func(key string, out interface{}) (bool, error) {
		receiver, ok := out.(*bool)
		if !ok {
			t.Fatalf("unexpected ready receiver type: %T", out)
		}
		*receiver = true
		return true, nil
	}
	historicalCaseVectorCacheHashGet = func(hashKey string, field string, out interface{}) (bool, error) {
		receiver, ok := out.(*case_library.HistoricalCaseRecord)
		if !ok {
			t.Fatalf("unexpected hash receiver type: %T", out)
		}
		if field != cached.CaseID {
			t.Fatalf("unexpected case id lookup: %s", field)
		}
		*receiver = cached
		return true, nil
	}

	record, found, err := case_library.GetHistoricalCaseByID(cached.CaseID)
	if err != nil {
		t.Fatalf("get historical case by id failed: %v", err)
	}
	if !found {
		t.Fatal("expected cached historical case to be found")
	}
	if record.CaseID != cached.CaseID || record.Title != cached.Title || record.EmbeddingModel != cached.EmbeddingModel {
		t.Fatalf("unexpected cached record: %+v", record)
	}
}

func TestDeleteHistoricalCaseByID_RemovesVectorCacheEntry(t *testing.T) {
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
	originalHashGet := historicalCaseVectorCacheHashGet
	originalHashSet := historicalCaseVectorCacheHashSetJSON
	originalHashDelete := historicalCaseVectorCacheHashDelete
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
		historicalCaseVectorCacheReadyGet = originalReadyGet
		historicalCaseVectorCacheReadySet = originalReadySet
		historicalCaseVectorCacheHashGet = originalHashGet
		historicalCaseVectorCacheHashSetJSON = originalHashSet
		historicalCaseVectorCacheHashDelete = originalHashDelete
	})

	generateCaseEmbedding = func(context.Context, string) ([]float64, string, error) {
		return []float64{0.4, 0.5, 0.6}, "mock-detail-cache", nil
	}
	searchHistoricalCasesByVector = func([]float64, int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{}, 1, nil
	}

	cacheStore := map[string]case_library.HistoricalCaseRecord{}
	setFields := make([]string, 0, 1)
	deleteFields := make([]string, 0, 1)

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
	historicalCaseVectorCacheHashGet = func(hashKey string, field string, out interface{}) (bool, error) {
		record, exists := cacheStore[field]
		if !exists {
			return false, nil
		}
		receiver, ok := out.(*case_library.HistoricalCaseRecord)
		if !ok {
			t.Fatalf("unexpected hash receiver type: %T", out)
		}
		*receiver = record
		return true, nil
	}
	historicalCaseVectorCacheHashSetJSON = func(hashKey string, field string, value interface{}) error {
		record, ok := value.(case_library.HistoricalCaseRecord)
		if !ok {
			t.Fatalf("unexpected hash value type: %T", value)
		}
		cacheStore[field] = record
		setFields = append(setFields, field)
		return nil
	}
	historicalCaseVectorCacheHashDelete = func(hashKey string, field string) error {
		delete(cacheStore, field)
		deleteFields = append(deleteFields, field)
		return nil
	}

	record, err := case_library.CreateHistoricalCase(context.Background(), "admin-delete", case_library.CreateHistoricalCaseInput{
		Title:           "删除时同步清理缓存",
		TargetGroup:     "老人",
		RiskLevel:       "高",
		ScamType:        "冒充客服类",
		CaseDescription: "案件详情缓存会在删除历史案件时同步失效，避免残留旧详情。",
		TypicalScripts:  []string{"我是平台客服"},
		Keywords:        []string{"客服", "退款"},
		ViolatedLaw:     "诈骗罪",
		Suggestion:      "及时止付并报警",
	})
	if err != nil {
		t.Fatalf("create historical case failed: %v", err)
	}
	if len(setFields) == 0 {
		t.Fatal("expected vector cache to be populated during historical case creation")
	}

	cachedField := setFields[len(setFields)-1]
	if _, exists := cacheStore[cachedField]; !exists {
		t.Fatalf("expected cached vector record for case_id %s", cachedField)
	}

	fetched, found, err := case_library.GetHistoricalCaseByID(record.CaseID)
	if err != nil {
		t.Fatalf("get historical case by id failed: %v", err)
	}
	if !found {
		t.Fatal("expected historical case to be found before delete")
	}
	if fetched.CaseID != record.CaseID || fetched.Title != record.Title {
		t.Fatalf("unexpected fetched historical case: %+v", fetched)
	}

	deleted, err := case_library.DeleteHistoricalCaseByID(record.CaseID)
	if err != nil {
		t.Fatalf("delete historical case failed: %v", err)
	}
	if !deleted {
		t.Fatal("expected historical case to be deleted")
	}
	if len(deleteFields) == 0 {
		t.Fatal("expected vector cache delete to run during historical case deletion")
	}
	if deleteFields[len(deleteFields)-1] != cachedField {
		t.Fatalf("unexpected deleted case_id: got %s want %s", deleteFields[len(deleteFields)-1], cachedField)
	}
	if _, exists := cacheStore[cachedField]; exists {
		t.Fatalf("expected vector cache entry to be removed for case_id %s", cachedField)
	}

	_, found, err = case_library.GetHistoricalCaseByID(record.CaseID)
	if err != nil {
		t.Fatalf("get historical case after delete failed: %v", err)
	}
	if found {
		t.Fatalf("expected deleted historical case to stay unavailable, case_id=%s", record.CaseID)
	}
}
