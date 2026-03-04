package tool_test

import (
	"strings"
	"testing"

	agenttool "antifraud/multi_agent/tool"
)

func TestParseAnalysisResult_ValidJSON(t *testing.T) {
	raw := `{"visual_impression":"v","key_content":"k","suspicious_points":["a","b"]}`
	got, err := agenttool.ParseAnalysisResult(raw)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.VisualImpression != "v" || got.KeyContent != "k" || len(got.SuspiciousPoints) != 2 {
		t.Fatalf("unexpected parse result: %+v", got)
	}
}

func TestParseAnalysisResult_InvalidJSON(t *testing.T) {
	_, err := agenttool.ParseAnalysisResult("{invalid")
	if err == nil {
		t.Fatalf("expected parse error for invalid json")
	}
}

func TestFormatAnalysisResult_NoSuspiciousPoints(t *testing.T) {
	out := agenttool.FormatAnalysisResult(agenttool.AnalysisResult{
		VisualImpression: "视觉描述",
		KeyContent:       "关键信息",
		SuspiciousPoints: nil,
	})
	if !strings.Contains(out, "未发现明显可疑信号") {
		t.Fatalf("expected fallback line, got %q", out)
	}
}

func TestFormatAnalysisResult_WithSuspiciousPoints(t *testing.T) {
	out := agenttool.FormatAnalysisResult(agenttool.AnalysisResult{
		VisualImpression: "视觉描述",
		KeyContent:       "关键信息",
		SuspiciousPoints: []string{"可疑点A", "可疑点B"},
	})
	if !strings.Contains(out, "1. 可疑点A") || !strings.Contains(out, "2. 可疑点B") {
		t.Fatalf("expected numbered suspicious points, got %q", out)
	}
}
