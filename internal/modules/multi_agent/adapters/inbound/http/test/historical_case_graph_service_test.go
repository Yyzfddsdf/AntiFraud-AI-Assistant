package httpapi_test

import (
	"testing"
	"time"

	httpapi "antifraud/internal/modules/multi_agent/adapters/inbound/http"
	"antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
)

func TestBuildHistoricalCaseGraphFromRecords(t *testing.T) {
	records := []case_library.HistoricalCaseRecord{
		{
			CaseID:          "HCASE-1",
			Title:           "客服退款诈骗",
			TargetGroup:     "老人",
			RiskLevel:       "高",
			ScamType:        "冒充客服类",
			CaseDescription: "以退款为由诱导转账",
			Keywords:        []string{"退款", "客服", "征信"},
			EmbeddingVector: []float64{1, 0},
			CreatedAt:       time.Now(),
		},
		{
			CaseID:          "HCASE-2",
			Title:           "征信修复诈骗",
			TargetGroup:     "老人",
			RiskLevel:       "高",
			ScamType:        "虚假征信类",
			CaseDescription: "声称征信受损需转账处理",
			Keywords:        []string{"征信", "客服", "修复"},
			EmbeddingVector: []float64{0.9, 0.1},
			CreatedAt:       time.Now(),
		},
		{
			CaseID:          "HCASE-3",
			Title:           "投资理财诈骗",
			TargetGroup:     "青年",
			RiskLevel:       "中",
			ScamType:        "虚假投资理财类",
			CaseDescription: "诱导充值投资平台",
			Keywords:        []string{"理财", "投资", "收益"},
			EmbeddingVector: []float64{0, 1},
			CreatedAt:       time.Now(),
		},
	}

	result := httpapi.BuildHistoricalCaseGraphFromRecords(records, "", "", 3)
	if result.Summary.TotalCases != 3 {
		t.Fatalf("unexpected total cases: %+v", result.Summary)
	}
	if result.Summary.ScamTypeCount != 3 {
		t.Fatalf("unexpected scam type count: %+v", result.Summary)
	}
	if len(result.Profiles) != 3 {
		t.Fatalf("unexpected profiles size: %d", len(result.Profiles))
	}
	if len(result.Graph.Nodes) == 0 || len(result.Graph.Edges) == 0 {
		t.Fatalf("graph should not be empty: %+v", result.Graph)
	}
	if len(result.TargetGroupTopScamTypes) != 2 {
		t.Fatalf("unexpected target group top scam types size: %d", len(result.TargetGroupTopScamTypes))
	}

	foundSimilarity := false
	foundElderTopK := false
	for _, profile := range result.Profiles {
		if profile.ScamType == "冒充客服类" && len(profile.SimilarTypes) > 0 {
			if profile.SimilarTypes[0].ScamType != "虚假征信类" {
				t.Fatalf("unexpected top similar type: %+v", profile.SimilarTypes)
			}
			foundSimilarity = true
		}
	}
	if !foundSimilarity {
		t.Fatalf("expected similar types for 冒充客服类")
	}

	for _, item := range result.TargetGroupTopScamTypes {
		if item.TargetGroup != "老人" {
			continue
		}
		foundElderTopK = true
		if item.TotalCases != 2 {
			t.Fatalf("unexpected 老人 total cases: %+v", item)
		}
		if len(item.TopScamTypes) != 2 {
			t.Fatalf("unexpected 老人 top scam types: %+v", item.TopScamTypes)
		}
		if item.TopScamTypes[0].Score != 0.5 || item.TopScamTypes[1].Score != 0.5 {
			t.Fatalf("unexpected 老人 top scam type scores: %+v", item.TopScamTypes)
		}
	}
	if !foundElderTopK {
		t.Fatalf("expected 老人 target group top scam types")
	}

	focusedByGroup := httpapi.BuildHistoricalCaseGraphFromRecords(records, "", "老人", 3)
	if focusedByGroup.Summary.FocusGroup != "老人" {
		t.Fatalf("unexpected focus group in summary: %+v", focusedByGroup.Summary)
	}
	if len(focusedByGroup.TargetGroupTopScamTypes) != 1 {
		t.Fatalf("unexpected focused target group top scam types size: %d", len(focusedByGroup.TargetGroupTopScamTypes))
	}
	if focusedByGroup.TargetGroupTopScamTypes[0].TargetGroup != "老人" {
		t.Fatalf("unexpected focused target group item: %+v", focusedByGroup.TargetGroupTopScamTypes[0])
	}
}
