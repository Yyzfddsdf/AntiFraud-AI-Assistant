package tool_test

import (
	"strings"
	"testing"

	agenttool "antifraud/multi_agent/tool"
)

func TestParseImageQuickRiskResult_Valid(t *testing.T) {
	got, err := agenttool.ParseImageQuickRiskResult(`{"risk_level":"高","reason":"页面出现伪造客服与转账引导"}`)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.RiskLevel != "高" || got.Reason == "" {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestParseImageQuickRiskResult_InvalidRiskLevel(t *testing.T) {
	_, err := agenttool.ParseImageQuickRiskResult(`{"risk_level":"紧急","reason":"x"}`)
	if err == nil {
		t.Fatalf("expected risk level validation error")
	}
}

func TestFormatImageQuickRiskResult(t *testing.T) {
	out := agenttool.FormatImageQuickRiskResult(agenttool.ImageQuickRiskResult{
		RiskLevel: "中",
		Reason:    "出现可疑下载引导",
	})
	if !strings.Contains(out, `"risk_level":"中"`) || !strings.Contains(out, `"reason":"出现可疑下载引导"`) {
		t.Fatalf("unexpected output: %q", out)
	}
}
