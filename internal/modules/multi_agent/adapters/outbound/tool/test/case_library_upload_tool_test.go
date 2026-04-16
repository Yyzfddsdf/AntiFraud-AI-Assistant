package tool_test

import (
	"context"
	"testing"

	case_library "antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	agenttool "antifraud/internal/modules/multi_agent/adapters/outbound/tool"
)

func TestParseUploadHistoricalCaseToVectorDBInput(t *testing.T) {
	raw := `{
		"title":"冒充客服诈骗",
		"target_group":"老人",
		"risk_level":"高",
		"scam_type":"冒充电商物流客服类",
		"case_description":"受害人收到自称客服电话，被诱导下载远程控制软件并转账。"
	}`

	input, err := agenttool.ParseUploadHistoricalCaseToVectorDBInput(raw)
	if err != nil {
		t.Fatalf("parse input failed: %v", err)
	}

	if input.Title != "冒充客服诈骗" {
		t.Fatalf("unexpected title: %q", input.Title)
	}
	if input.TargetGroup != "老人" {
		t.Fatalf("unexpected target_group: %q", input.TargetGroup)
	}
	if input.RiskLevel != "高" {
		t.Fatalf("unexpected risk_level: %q", input.RiskLevel)
	}
}

func TestUploadHistoricalCaseToVectorDBHandler_InvalidJSON(t *testing.T) {
	handler := &agenttool.UploadHistoricalCaseToVectorDBHandler{}

	resp, err := handler.Handle(context.Background(), "{")
	if err != nil {
		t.Fatalf("handle should not return error, got: %v", err)
	}

	if status, _ := resp.Payload["status"].(string); status != "failed" {
		t.Fatalf("expected status=failed, got: %v", resp.Payload["status"])
	}
	if _, ok := resp.Payload["error"]; !ok {
		t.Fatalf("expected error in payload")
	}
}

func TestNormalizeViolatedLaw_DefaultsToFraudCrime(t *testing.T) {
	if got := normalizeViolatedLaw("   "); got != "涉嫌违反《中华人民共和国刑法》第二百六十六条（诈骗罪）" {
		t.Fatalf("unexpected default violated law: %q", got)
	}
	if got := normalizeViolatedLaw("  涉嫌违反其他条款 "); got != "涉嫌违反其他条款" {
		t.Fatalf("unexpected normalized violated law: %q", got)
	}
}

func TestUploadHistoricalCaseToVectorDBHandler_DuplicateHistoricalCase(t *testing.T) {
	originalCreatePendingReview := createPendingReview
	t.Cleanup(func() {
		createPendingReview = originalCreatePendingReview
	})

	createPendingReview = func(ctx context.Context, userID string, input case_library.CreateHistoricalCaseInput) (case_library.PendingReviewRecord, error) {
		return case_library.PendingReviewRecord{}, &case_library.DuplicateHistoricalCaseError{
			TopMatch: case_library.SimilarCaseResult{
				CaseID:      "HCASE-001",
				Title:       "冒充客服退款诈骗",
				TargetGroup: "老人",
				RiskLevel:   "高",
				ScamType:    "冒充电商物流客服类",
				Similarity:  0.95,
			},
		}
	}

	handler := &agenttool.UploadHistoricalCaseToVectorDBHandler{}
	resp, err := handler.Handle(context.Background(), `{
		"title":"冒充客服诈骗",
		"target_group":"老人",
		"risk_level":"高",
		"scam_type":"冒充电商物流客服类",
		"case_description":"受害人收到自称客服电话，被诱导下载远程控制软件并转账。"
	}`)
	if err != nil {
		t.Fatalf("handle should not return error, got: %v", err)
	}

	if status, _ := resp.Payload["status"].(string); status != "failed" {
		t.Fatalf("expected status=failed, got: %#v", resp.Payload["status"])
	}
	duplicateCase, ok := resp.Payload["duplicate_case"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected duplicate_case payload, got: %#v", resp.Payload["duplicate_case"])
	}
	if duplicateCase["case_id"] != "HCASE-001" {
		t.Fatalf("unexpected duplicate case: %+v", duplicateCase)
	}
}
