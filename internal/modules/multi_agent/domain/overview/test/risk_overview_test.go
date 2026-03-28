package overview_test

import (
	"testing"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	overview "antifraud/internal/modules/multi_agent/domain/overview"
)

func TestBuildRiskOverviewFromHistory_Day(t *testing.T) {
	now := time.Now().UTC()
	day2 := startOfUTCDayForTest(now).Add(-6 * 24 * time.Hour)
	day1 := day2.Add(-24 * time.Hour)
	history := []state.CaseHistoryRecord{
		{RiskLevel: "高", CreatedAt: day2.Add(9*time.Hour + 30*time.Minute)},
		{RiskLevel: "中", CreatedAt: day2.Add(11*time.Hour + 45*time.Minute)},
		{RiskLevel: "低", CreatedAt: day1.Add(13*time.Hour + 15*time.Minute)},
		{RiskLevel: "unknown", CreatedAt: day1.Add(16*time.Hour + 5*time.Minute)},
	}

	result := overview.BuildRiskOverviewFromHistory("u-1", history, overview.IntervalDay)

	if result.UserID != "u-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if result.Interval != overview.IntervalDay {
		t.Fatalf("unexpected interval: %s", result.Interval)
	}

	if result.Stats.High != 1 || result.Stats.Medium != 2 || result.Stats.Low != 1 || result.Stats.Total != 4 {
		t.Fatalf("unexpected stats: %+v", result.Stats)
	}

	if len(result.Trend) != 2 {
		t.Fatalf("unexpected trend size: %d", len(result.Trend))
	}

	first := result.Trend[0]
	if first.TimeBucket != day1.Format("2006-01-02") || first.High != 0 || first.Medium != 1 || first.Low != 1 || first.Total != 2 {
		t.Fatalf("unexpected first trend point: %+v", first)
	}

	second := result.Trend[1]
	if second.TimeBucket != day2.Format("2006-01-02") || second.High != 1 || second.Medium != 1 || second.Low != 0 || second.Total != 2 {
		t.Fatalf("unexpected second trend point: %+v", second)
	}

	currentWindowStart := startOfUTCDayForTest(now).AddDate(0, 0, -6)
	currentWindowEnd := startOfUTCDayForTest(now)
	previousWindowStart := currentWindowStart.AddDate(0, 0, -7)
	previousWindowEnd := currentWindowStart.AddDate(0, 0, -1)
	if result.Analysis.CurrentWindow != currentWindowStart.Format("2006-01-02")+" ~ "+currentWindowEnd.Format("2006-01-02") || result.Analysis.PreviousWindow != previousWindowStart.Format("2006-01-02")+" ~ "+previousWindowEnd.Format("2006-01-02") {
		t.Fatalf("unexpected analysis buckets: %+v", result.Analysis)
	}
	if result.Analysis.OverallTrend != "上升" {
		t.Fatalf("unexpected overall trend: %+v", result.Analysis)
	}
	if result.Analysis.HighRiskTrend != "上升" {
		t.Fatalf("unexpected high risk trend: %+v", result.Analysis)
	}
	if result.Analysis.Summary == "" {
		t.Fatalf("analysis summary should not be empty")
	}
}

func TestBuildRiskOverviewFromHistory_InsufficientData(t *testing.T) {
	now := time.Now().UTC()
	history := []state.CaseHistoryRecord{
		{RiskLevel: "中", CreatedAt: startOfUTCDayForTest(now).Add(-24*time.Hour + 10*time.Hour)},
	}

	result := overview.BuildRiskOverviewFromHistory("u-2", history, overview.IntervalDay)
	if result.Analysis.OverallTrend != "平稳" {
		t.Fatalf("unexpected overall trend: %+v", result.Analysis)
	}
	if result.Analysis.HighRiskTrend != "平稳" {
		t.Fatalf("unexpected high risk trend: %+v", result.Analysis)
	}
}

func TestBuildRiskOverviewFromHistory_WindowComparison(t *testing.T) {
	now := time.Now().UTC()
	base := startOfUTCDayForTest(now).Add(-4 * 24 * time.Hour)
	history := []state.CaseHistoryRecord{
		{RiskLevel: "高", CreatedAt: base.Add(9 * time.Hour)},
		{RiskLevel: "低", CreatedAt: base.Add(24*time.Hour + 9*time.Hour)},
		{RiskLevel: "高", CreatedAt: base.Add(48*time.Hour + 9*time.Hour)},
		{RiskLevel: "高", CreatedAt: base.Add(72*time.Hour + 9*time.Hour)},
	}

	result := overview.BuildRiskOverviewFromHistory("u-3", history, overview.IntervalDay)
	currentWindowStart := startOfUTCDayForTest(now).AddDate(0, 0, -6)
	currentWindowEnd := startOfUTCDayForTest(now)
	previousWindowStart := currentWindowStart.AddDate(0, 0, -7)
	previousWindowEnd := currentWindowStart.AddDate(0, 0, -1)
	if result.Analysis.PreviousWindow != previousWindowStart.Format("2006-01-02")+" ~ "+previousWindowEnd.Format("2006-01-02") {
		t.Fatalf("unexpected previous window: %+v", result.Analysis)
	}
	if result.Analysis.CurrentWindow != currentWindowStart.Format("2006-01-02")+" ~ "+currentWindowEnd.Format("2006-01-02") {
		t.Fatalf("unexpected current window: %+v", result.Analysis)
	}
	if result.Analysis.HighRiskTrend != "上升" {
		t.Fatalf("unexpected high risk trend: %+v", result.Analysis)
	}
	if result.Analysis.OverallTrend != "上升" {
		t.Fatalf("overall trend should rise when high risk count rises: %+v", result.Analysis)
	}
}

func TestBuildRiskOverviewFromHistory_NoRecentCases(t *testing.T) {
	now := time.Now().UTC()
	history := []state.CaseHistoryRecord{
		{RiskLevel: "高", CreatedAt: startOfUTCDayForTest(now).AddDate(0, 0, -20).Add(9 * time.Hour)},
	}

	result := overview.BuildRiskOverviewFromHistory("u-4", history, overview.IntervalDay)
	if result.Analysis.OverallTrend != "近期无案件" {
		t.Fatalf("unexpected overall trend: %+v", result.Analysis)
	}
	if result.Analysis.HighRiskTrend != "近期无案件" {
		t.Fatalf("unexpected high risk trend: %+v", result.Analysis)
	}
}

func startOfUTCDayForTest(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func TestNormalizeInterval(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "", want: overview.IntervalDay, ok: true},
		{input: "day", want: overview.IntervalDay, ok: true},
		{input: "week", want: overview.IntervalWeek, ok: true},
		{input: "month", want: overview.IntervalMonth, ok: true},
		{input: "invalid", want: "", ok: false},
	}

	for _, tc := range tests {
		got, ok := overview.NormalizeInterval(tc.input)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("NormalizeInterval(%q) = (%q, %v), want (%q, %v)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}
