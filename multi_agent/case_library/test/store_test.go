package case_library_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"antifraud/database"
	case_library "antifraud/multi_agent/case_library"
)

func TestBuildEmbeddingInput_UsesOnlyFocusedFields(t *testing.T) {
	input := case_library.CreateHistoricalCaseInput{
		Title:           "冒充客服退款",
		TargetGroup:     "老年人",
		RiskLevel:       "高",
		ScamType:        "冒充电商物流客服类",
		CaseDescription: "嫌疑人冒充平台客服，以退款为由诱导受害人下载远程控制软件并转账。",
		TypicalScripts:  []string{"我是平台客服", "现在帮你退款"},
		Keywords:        []string{"退款", "远程控制", "客服"},
		ViolatedLaw:     "诈骗罪",
		Suggestion:      "立即报警并联系银行止付",
	}

	got := case_library.BuildEmbeddingInput(input)

	for _, want := range []string{input.Title, input.ScamType, input.CaseDescription, "退款", "远程控制", "客服"} {
		if !strings.Contains(got, want) {
			t.Fatalf("embedding text should contain %q, got: %s", want, got)
		}
	}

	for _, unwanted := range []string{input.TargetGroup, input.RiskLevel, input.TypicalScripts[0], input.TypicalScripts[1], input.ViolatedLaw, input.Suggestion} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("embedding text should not contain %q, got: %s", unwanted, got)
		}
	}
}

func TestBuildEmbeddingInput_DropsLowQualityKeywords(t *testing.T) {
	input := case_library.CreateHistoricalCaseInput{
		Title:           "投资返利诈骗",
		ScamType:        "虚假投资理财类",
		CaseDescription: "嫌疑人通过社交平台引导受害人进入虚假投资平台，承诺高收益并诱导持续充值。",
		Keywords: []string{
			"a",
			"这是一个明显过长且更像完整句子的关键词描述内容",
			"返利，稳赚不赔",
		},
	}

	got := case_library.BuildEmbeddingInput(input)
	if strings.Contains(got, "关键词:") {
		t.Fatalf("low-quality keywords should be excluded from embedding text, got: %s", got)
	}
	if !strings.Contains(got, input.Title) || !strings.Contains(got, input.ScamType) || !strings.Contains(got, input.CaseDescription) {
		t.Fatalf("core fields should still remain in embedding text, got: %s", got)
	}
	if strings.Contains(got, "目标人群:") || strings.Contains(got, "风险等级:") || strings.Contains(got, "建议:") {
		t.Fatalf("non-focused fields should not appear in embedding text, got: %s", got)
	}
}

func TestCreateHistoricalCase_DuplicateHistoricalCaseRejected(t *testing.T) {
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

	_, err = case_library.CreateHistoricalCase(context.Background(), "admin-u1", case_library.CreateHistoricalCaseInput{
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

	db, err := database.GetHistoricalCaseDB()
	if err != nil {
		t.Fatalf("get historical case db failed: %v", err)
	}

	var count int64
	if err := db.Table("historical_case_library").Count(&count).Error; err != nil {
		t.Fatalf("count historical case rows failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("duplicate create should not insert rows, got count=%d", count)
	}
}
