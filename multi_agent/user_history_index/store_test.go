package user_history_index

import (
	"strings"
	"sync"
	"testing"
	"time"

	"antifraud/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBuildEmbeddingInput_UsesTextualFields(t *testing.T) {
	got := BuildEmbeddingInput(ArchiveInput{
		Title:       "冒充客服退款",
		CaseSummary: "对方要求共享屏幕并转账",
		ScamType:    "冒充客服类",
	})

	checks := []string{
		"标题: 冒充客服退款",
		"案件摘要: 对方要求共享屏幕并转账",
		"诈骗类型: 冒充客服类",
	}
	for _, want := range checks {
		if !strings.Contains(got, want) {
			t.Fatalf("embedding input should contain %q, got: %s", want, got)
		}
	}

	unwanted := []string{"风险等级:", "原始文本:", "视频洞察:", "音频洞察:", "图片洞察:", "最终报告:"}
	for _, item := range unwanted {
		if strings.Contains(got, item) {
			t.Fatalf("embedding input should not contain %q, got: %s", item, got)
		}
	}
}

func TestSearchTopKSimilarHistoryByVector_ReturnsRankedResults(t *testing.T) {
	oldDB := database.DB
	defer func() {
		database.DB = oldDB
	}()

	userHistorySchemaOnce = sync.Once{}
	userHistorySchemaErr = nil

	db, err := gorm.Open(sqlite.Open(t.TempDir()+"\\user_history_test.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db failed: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	database.DB = db

	if err := ensureUserHistoryVectorSchema(db); err != nil {
		t.Fatalf("ensure schema failed: %v", err)
	}

	now := time.Now()
	rows := []userHistoryVectorEntity{
		{
			RecordID:           "TASK-1",
			UserID:             "u-1",
			EmbeddingVector:    encodeFloatList([]float64{1, 0}),
			EmbeddingModel:     "test-model",
			EmbeddingDimension: 2,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
		{
			RecordID:           "TASK-2",
			UserID:             "u-1",
			EmbeddingVector:    encodeFloatList([]float64{0, 1}),
			EmbeddingModel:     "test-model",
			EmbeddingDimension: 2,
			CreatedAt:          now.Add(-time.Minute),
			UpdatedAt:          now.Add(-time.Minute),
		},
		{
			RecordID:           "TASK-3",
			UserID:             "u-2",
			EmbeddingVector:    encodeFloatList([]float64{1, 0}),
			EmbeddingModel:     "test-model",
			EmbeddingDimension: 2,
			CreatedAt:          now,
			UpdatedAt:          now,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("insert rows failed: %v", err)
	}

	results, appliedTopK, err := SearchTopKSimilarHistoryByVector("u-1", []float64{1, 0}, 5)
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}
	if appliedTopK != 5 {
		t.Fatalf("expected appliedTopK=5, got %d", appliedTopK)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for same user, got %d", len(results))
	}
	if results[0].RecordID != "TASK-1" {
		t.Fatalf("expected TASK-1 ranked first, got %+v", results[0])
	}
	if results[1].RecordID != "TASK-2" {
		t.Fatalf("expected TASK-2 ranked second, got %+v", results[1])
	}
	if results[0].Similarity <= results[1].Similarity {
		t.Fatalf("expected first similarity > second similarity, got %.4f <= %.4f", results[0].Similarity, results[1].Similarity)
	}
}
