package tool_test

import (
	"context"
	"testing"

	agenttool "antifraud/multi_agent/tool"
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
