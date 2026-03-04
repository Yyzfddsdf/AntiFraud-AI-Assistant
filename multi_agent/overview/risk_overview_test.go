package overview

import (
	"testing"
	"time"

	"antifraud/multi_agent/state"
)

func TestBuildRiskOverviewFromHistory_Day(t *testing.T) {
	history := []state.CaseHistoryRecord{
		{RiskLevel: "高", CreatedAt: time.Date(2026, 3, 2, 9, 30, 0, 0, time.UTC)},
		{RiskLevel: "中", CreatedAt: time.Date(2026, 3, 2, 11, 45, 0, 0, time.UTC)},
		{RiskLevel: "低", CreatedAt: time.Date(2026, 3, 1, 13, 15, 0, 0, time.UTC)},
		{RiskLevel: "unknown", CreatedAt: time.Date(2026, 3, 1, 16, 5, 0, 0, time.UTC)},
	}

	result := BuildRiskOverviewFromHistory("u-1", history, IntervalDay)

	if result.UserID != "u-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if result.Interval != IntervalDay {
		t.Fatalf("unexpected interval: %s", result.Interval)
	}

	if result.Stats.High != 1 || result.Stats.Medium != 2 || result.Stats.Low != 1 || result.Stats.Total != 4 {
		t.Fatalf("unexpected stats: %+v", result.Stats)
	}

	if len(result.Trend) != 2 {
		t.Fatalf("unexpected trend size: %d", len(result.Trend))
	}

	first := result.Trend[0]
	if first.TimeBucket != "2026-03-01" || first.High != 0 || first.Medium != 1 || first.Low != 1 || first.Total != 2 {
		t.Fatalf("unexpected first trend point: %+v", first)
	}

	second := result.Trend[1]
	if second.TimeBucket != "2026-03-02" || second.High != 1 || second.Medium != 1 || second.Low != 0 || second.Total != 2 {
		t.Fatalf("unexpected second trend point: %+v", second)
	}
}

func TestNormalizeInterval(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "", want: IntervalDay, ok: true},
		{input: "day", want: IntervalDay, ok: true},
		{input: "week", want: IntervalWeek, ok: true},
		{input: "month", want: IntervalMonth, ok: true},
		{input: "invalid", want: "", ok: false},
	}

	for _, tc := range tests {
		got, ok := NormalizeInterval(tc.input)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("NormalizeInterval(%q) = (%q, %v), want (%q, %v)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}
