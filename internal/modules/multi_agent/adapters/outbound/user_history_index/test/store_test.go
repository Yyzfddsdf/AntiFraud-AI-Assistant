package user_history_index_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	indexpkg "antifraud/internal/modules/multi_agent/adapters/outbound/user_history_index"
	"antifraud/internal/platform/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBuildEmbeddingInput_UsesTextualFields(t *testing.T) {
	got := indexpkg.BuildEmbeddingInput(indexpkg.ArchiveInput{
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

	if err := createUserHistoryVectorsTable(db); err != nil {
		t.Fatalf("create table failed: %v", err)
	}

	now := time.Now()
	rows := []struct {
		recordID string
		userID   string
		vector   []float64
		created  time.Time
	}{
		{recordID: "TASK-1", userID: "u-1", vector: []float64{1, 0}, created: now},
		{recordID: "TASK-2", userID: "u-1", vector: []float64{0, 1}, created: now.Add(-time.Minute)},
		{recordID: "TASK-3", userID: "u-2", vector: []float64{1, 0}, created: now},
	}
	for _, row := range rows {
		if err := insertUserHistoryVector(db, row.recordID, row.userID, row.vector, row.created); err != nil {
			t.Fatalf("insert row failed: %v", err)
		}
	}

	results, appliedTopK, err := indexpkg.SearchTopKSimilarHistoryByVector("u-1", []float64{1, 0}, 5)
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

func createUserHistoryVectorsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS user_history_vectors (
			record_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			embedding_vector TEXT NOT NULL,
			embedding_model TEXT NOT NULL,
			embedding_dimension INTEGER NOT NULL,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL,
			PRIMARY KEY (record_id, user_id)
		)
	`).Error
}

func insertUserHistoryVector(db *gorm.DB, recordID string, userID string, vector []float64, createdAt time.Time) error {
	rawVector, err := json.Marshal(vector)
	if err != nil {
		return err
	}
	return db.Exec(
		`INSERT INTO user_history_vectors (record_id, user_id, embedding_vector, embedding_model, embedding_dimension, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		recordID,
		userID,
		string(rawVector),
		"test-model",
		len(vector),
		createdAt,
		createdAt,
	).Error
}
