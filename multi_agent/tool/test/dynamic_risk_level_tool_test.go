package tool_test

import (
	"testing"

	agenttool "antifraud/multi_agent/tool"
)

func TestDynamicThresholdFromHistoricalScore(t *testing.T) {
	cases := []struct {
		score int
		want  int
	}{
		{0, 60},
		{20, 60},
		{21, 55},
		{40, 55},
		{41, 50},
		{60, 50},
		{61, 45},
		{80, 45},
		{81, 40},
		{100, 40},
	}

	for _, tc := range cases {
		got := agenttool.DynamicThresholdFromHistoricalScore(tc.score)
		if got != tc.want {
			t.Fatalf("score=%d: want threshold=%d got=%d", tc.score, tc.want, got)
		}
	}
}

func TestResolveDynamicRiskLevel(t *testing.T) {
	high, err := agenttool.ResolveDynamicRiskLevel(72, 65, "high", "high")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if high.RiskLevel != "高" || high.DynamicThreshold != 45 {
		t.Fatalf("unexpected high decision: %+v", high)
	}

	mediumByAbove, err := agenttool.ResolveDynamicRiskLevel(52, 65, "none", "none")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if mediumByAbove.RiskLevel != "中" {
		t.Fatalf("unexpected medium-by-above decision: %+v", mediumByAbove)
	}

	mediumByHit, err := agenttool.ResolveDynamicRiskLevel(32, 65, "high", "none")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if mediumByHit.RiskLevel != "中" {
		t.Fatalf("unexpected medium-by-hit decision: %+v", mediumByHit)
	}

	lowBelowBufferedFloor, err := agenttool.ResolveDynamicRiskLevel(20, 65, "high", "none")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if lowBelowBufferedFloor.RiskLevel != "低" {
		t.Fatalf("expected far-below-threshold high hit to remain low, got %+v", lowBelowBufferedFloor)
	}

	mediumByDoubleHit, err := agenttool.ResolveDynamicRiskLevel(22, 65, "high", "high")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if mediumByDoubleHit.RiskLevel != "中" {
		t.Fatalf("expected double high hit near buffered floor to become medium, got %+v", mediumByDoubleHit)
	}

	low, err := agenttool.ResolveDynamicRiskLevel(40, 10, "none", "none")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if low.RiskLevel != "低" || low.DynamicThreshold != 60 {
		t.Fatalf("unexpected low decision: %+v", low)
	}

	lowByLowHit, err := agenttool.ResolveDynamicRiskLevel(48, 45, "low", "none")
	if err != nil {
		t.Fatalf("resolve dynamic risk level failed: %v", err)
	}
	if lowByLowHit.CurrentScore >= 48 || lowByLowHit.RiskLevel != "低" {
		t.Fatalf("expected low-hit adjustment to reduce score and keep low risk, got %+v", lowByLowHit)
	}
}
