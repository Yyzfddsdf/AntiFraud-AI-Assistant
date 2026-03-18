package case_library_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"antifraud/database"
	case_library "antifraud/multi_agent/case_library"
)

func TestCreatePendingReview_DuplicateHistoricalCaseRejected(t *testing.T) {
	originalGenerateCaseEmbedding := generateCaseEmbedding
	originalSearchHistoricalCases := searchHistoricalCasesByVector
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
	})

	generateCaseEmbedding = func(ctx context.Context, input string) ([]float64, string, error) {
		return []float64{0.1, 0.2, 0.3}, "mock-embedding", nil
	}
	searchHistoricalCasesByVector = func(queryVector []float64, topK int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{
			{
				CaseID:      "HCASE-EXISTING",
				Title:       "已存在诈骗案件",
				TargetGroup: "老人",
				RiskLevel:   "高",
				ScamType:    "冒充客服类",
				Similarity:  0.95,
			},
		}, 1, nil
	}

	_, err := case_library.CreatePendingReview(context.Background(), "u1", case_library.CreateHistoricalCaseInput{
		Title:           "冒充客服诈骗",
		TargetGroup:     "老人",
		RiskLevel:       "高",
		ScamType:        "冒充客服类",
		CaseDescription: "受害人收到自称客服电话，被诱导下载远程控制软件并转账。",
	})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
	if !case_library.IsDuplicateHistoricalCaseError(err) {
		t.Fatalf("expected duplicate error type, got: %v", err)
	}
}

func TestCreatePendingReview_StoresEmbeddingFields(t *testing.T) {
	resetHistoricalCaseDB()
	dbPath, err := prepareHistoricalCaseDBPath()
	if err != nil {
		t.Fatalf("prepare historical case db path failed: %v", err)
	}
	t.Setenv("HISTORICAL_CASE_DB_PATH", dbPath)

	originalGenerateCaseEmbedding := generateCaseEmbedding
	originalSearchHistoricalCases := searchHistoricalCasesByVector
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
		resetHistoricalCaseDB()
		_ = os.Remove(dbPath)
	})

	generateCaseEmbedding = func(ctx context.Context, input string) ([]float64, string, error) {
		return []float64{0.4, 0.5, 0.6}, "mock-embedding", nil
	}
	searchHistoricalCasesByVector = func(queryVector []float64, topK int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{}, 1, nil
	}

	record, err := case_library.CreatePendingReview(context.Background(), "u1", case_library.CreateHistoricalCaseInput{
		Title:           "冒充客服诈骗",
		TargetGroup:     "老人",
		RiskLevel:       "高",
		ScamType:        "冒充客服类",
		CaseDescription: "受害人收到自称客服电话，被诱导下载远程控制软件并转账。",
	})
	if err != nil {
		t.Fatalf("create pending review failed: %v", err)
	}

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		t.Fatalf("get historical case db failed: %v", err)
	}

	var row struct {
		EmbeddingVector    string
		EmbeddingModel     string
		EmbeddingDimension int
	}
	if err := db.Table("pending_review_cases").
		Select("embedding_vector", "embedding_model", "embedding_dimension").
		Where("record_id = ?", record.RecordID).
		Take(&row).Error; err != nil {
		t.Fatalf("query pending review row failed: %v", err)
	}

	if row.EmbeddingModel != "mock-embedding" {
		t.Fatalf("unexpected embedding model: %q", row.EmbeddingModel)
	}
	if row.EmbeddingDimension != 3 {
		t.Fatalf("unexpected embedding dimension: %d", row.EmbeddingDimension)
	}
	if !strings.Contains(row.EmbeddingVector, "0.4") {
		t.Fatalf("unexpected embedding vector payload: %q", row.EmbeddingVector)
	}
}

func TestRejectPendingReview_DeletesRecord(t *testing.T) {
	resetHistoricalCaseDB()
	dbPath, err := prepareHistoricalCaseDBPath()
	if err != nil {
		t.Fatalf("prepare historical case db path failed: %v", err)
	}
	t.Setenv("HISTORICAL_CASE_DB_PATH", dbPath)

	originalGenerateCaseEmbedding := generateCaseEmbedding
	originalSearchHistoricalCases := searchHistoricalCasesByVector
	t.Cleanup(func() {
		generateCaseEmbedding = originalGenerateCaseEmbedding
		searchHistoricalCasesByVector = originalSearchHistoricalCases
		resetHistoricalCaseDB()
		_ = os.Remove(dbPath)
	})

	generateCaseEmbedding = func(ctx context.Context, input string) ([]float64, string, error) {
		return []float64{0.7, 0.8, 0.9}, "mock-embedding", nil
	}
	searchHistoricalCasesByVector = func(queryVector []float64, topK int) ([]case_library.SimilarCaseResult, int, error) {
		return []case_library.SimilarCaseResult{}, 1, nil
	}

	record, err := case_library.CreatePendingReview(context.Background(), "u2", case_library.CreateHistoricalCaseInput{
		Title:           "虚假投资诈骗",
		TargetGroup:     "青年",
		RiskLevel:       "中",
		ScamType:        "虚假投资理财类",
		CaseDescription: "以高收益理财为诱饵引导受害人多次充值。",
	})
	if err != nil {
		t.Fatalf("create pending review failed: %v", err)
	}

	if err := case_library.RejectPendingReview(context.Background(), record.RecordID); err != nil {
		t.Fatalf("reject pending review failed: %v", err)
	}

	_, exists, err := case_library.GetPendingReviewByID(record.RecordID)
	if err != nil {
		t.Fatalf("query pending review failed: %v", err)
	}
	if exists {
		t.Fatalf("pending review should be deleted after reject, record_id=%s", record.RecordID)
	}
}
