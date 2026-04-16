package overview

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
)

const (
	IntervalDay   = "day"
	IntervalWeek  = "week"
	IntervalMonth = "month"
)

// RiskStats 表示风险等级统计总览。
type RiskStats struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

// RiskTrendPoint 表示某个时间桶下的风险分布。
type RiskTrendPoint struct {
	TimeBucket string `json:"time_bucket"`
	High       int    `json:"high"`
	Medium     int    `json:"medium"`
	Low        int    `json:"low"`
	Total      int    `json:"total"`
}

// RiskTrendAnalysis 表示对最近两个活跃统计窗口的轻量趋势判断。
type RiskTrendAnalysis struct {
	CurrentWindow  string `json:"current_window"`
	PreviousWindow string `json:"previous_window,omitempty"`
	OverallTrend   string `json:"overall_trend"`
	HighRiskTrend  string `json:"high_risk_trend"`
	Summary        string `json:"summary"`
}

// UserRiskOverview 是用户维度的风险总览结果。
type UserRiskOverview struct {
	UserID   string            `json:"user_id"`
	Interval string            `json:"interval"`
	Stats    RiskStats         `json:"stats"`
	Trend    []RiskTrendPoint  `json:"trend"`
	Analysis RiskTrendAnalysis `json:"analysis"`
}

// NormalizeInterval 归一化时间聚合粒度。
func NormalizeInterval(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", IntervalDay:
		return IntervalDay, true
	case IntervalWeek:
		return IntervalWeek, true
	case IntervalMonth:
		return IntervalMonth, true
	default:
		return "", false
	}
}

// BuildUserRiskOverview 使用 state.GetCaseHistory 获取数据并构建风险总览。
func BuildUserRiskOverview(userID string, interval string) UserRiskOverview {
	normalized, ok := NormalizeInterval(interval)
	if !ok {
		normalized = IntervalDay
	}
	history := state.GetCaseHistory(userID)
	return BuildRiskOverviewFromHistory(userID, history, normalized)
}

// BuildRiskOverviewFromHistory 根据历史记录构建风险总览，便于复用与测试。
func BuildRiskOverviewFromHistory(userID string, history []state.CaseHistoryRecord, interval string) UserRiskOverview {
	normalized, ok := NormalizeInterval(interval)
	if !ok {
		normalized = IntervalDay
	}

	stats := RiskStats{}
	buckets := make(map[string]*RiskTrendPoint, len(history))
	for _, item := range history {
		bucket := buildTimeBucket(item.CreatedAt, normalized)
		point, exists := buckets[bucket]
		if !exists {
			point = &RiskTrendPoint{TimeBucket: bucket}
			buckets[bucket] = point
		}

		switch normalizeRiskLevel(item.RiskLevel) {
		case "高":
			stats.High++
			point.High++
		case "低":
			stats.Low++
			point.Low++
		default:
			stats.Medium++
			point.Medium++
		}
	}

	stats.Total = stats.High + stats.Medium + stats.Low

	trend := make([]RiskTrendPoint, 0, len(buckets))
	for _, point := range buckets {
		point.Total = point.High + point.Medium + point.Low
		trend = append(trend, *point)
	}
	sort.Slice(trend, func(i, j int) bool {
		return trend[i].TimeBucket < trend[j].TimeBucket
	})

	return UserRiskOverview{
		UserID:   strings.TrimSpace(userID),
		Interval: normalized,
		Stats:    stats,
		Trend:    trend,
		Analysis: buildTrendAnalysis(history, normalized),
	}
}

func buildTrendAnalysis(history []state.CaseHistoryRecord, interval string) RiskTrendAnalysis {
	if len(history) == 0 {
		return RiskTrendAnalysis{
			OverallTrend:  "暂无数据",
			HighRiskTrend: "暂无数据",
			Summary:       "暂无历史案件数据，暂时无法分析风险趋势。",
		}
	}

	currentWindow, previousWindow := buildAnalysisWindows(time.Now().UTC(), interval)
	currentStats := aggregateHistoryWindow(history, currentWindow.Start, currentWindow.End)
	if currentStats.Total == 0 {
		return RiskTrendAnalysis{
			CurrentWindow:  currentWindow.Label,
			PreviousWindow: previousWindow.Label,
			OverallTrend:   "近期无案件",
			HighRiskTrend:  "近期无案件",
			Summary:        fmt.Sprintf("最近%s内暂无新增案件，暂不进行风险趋势判断。", currentWindow.HumanLabel),
		}
	}

	previousStats := aggregateHistoryWindow(history, previousWindow.Start, previousWindow.End)
	currentPressure := riskPressureScore(currentStats)
	previousPressure := riskPressureScore(previousStats)
	overallTrend := classifyOverallTrend(currentStats, previousStats)
	highRiskTrend := classifyHighRiskTrend(currentStats.High, previousStats.High)

	return RiskTrendAnalysis{
		CurrentWindow:  currentWindow.Label,
		PreviousWindow: previousWindow.Label,
		OverallTrend:   overallTrend,
		HighRiskTrend:  highRiskTrend,
		Summary: fmt.Sprintf(
			"基于最近%s与上一窗口的对比，高风险案件%s（%d→%d），整体风险%s（风险压力 %d→%d）。",
			currentWindow.HumanLabel,
			highRiskTrend,
			previousStats.High,
			currentStats.High,
			overallTrend,
			previousPressure,
			currentPressure,
		),
	}
}

type analysisWindow struct {
	Start      time.Time
	End        time.Time
	Label      string
	HumanLabel string
}

func buildAnalysisWindows(now time.Time, interval string) (analysisWindow, analysisWindow) {
	currentStart, currentEnd, humanLabel := currentAnalysisWindowBounds(now, interval)
	previousDuration := currentEnd.Sub(currentStart)
	previousStart := currentStart.Add(-previousDuration)
	previousEnd := currentStart
	return analysisWindow{
			Start:      currentStart,
			End:        currentEnd,
			Label:      buildWindowRangeLabel(currentStart, currentEnd, interval),
			HumanLabel: humanLabel,
		}, analysisWindow{
			Start:      previousStart,
			End:        previousEnd,
			Label:      buildWindowRangeLabel(previousStart, previousEnd, interval),
			HumanLabel: humanLabel,
		}
}

func currentAnalysisWindowBounds(now time.Time, interval string) (time.Time, time.Time, string) {
	utcNow := now.UTC()
	switch interval {
	case IntervalWeek:
		end := startOfUTCDay(utcNow).AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -14)
		return start, end, "14天"
	case IntervalMonth:
		end := startOfUTCDay(utcNow).AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -30)
		return start, end, "30天"
	default:
		end := startOfUTCDay(utcNow).AddDate(0, 0, 1)
		start := end.AddDate(0, 0, -7)
		return start, end, "7天"
	}
}

func aggregateHistoryWindow(history []state.CaseHistoryRecord, start time.Time, end time.Time) RiskStats {
	stats := RiskStats{}
	for _, item := range history {
		createdAt := item.CreatedAt.UTC()
		if createdAt.Before(start) || !createdAt.Before(end) {
			continue
		}
		switch normalizeRiskLevel(item.RiskLevel) {
		case "高":
			stats.High++
		case "低":
			stats.Low++
		default:
			stats.Medium++
		}
	}
	stats.Total = stats.High + stats.Medium + stats.Low
	return stats
}

func buildWindowRangeLabel(start time.Time, end time.Time, interval string) string {
	endInclusive := end.Add(-time.Nanosecond)
	startLabel := start.UTC().Format("2006-01-02")
	endLabel := endInclusive.UTC().Format("2006-01-02")
	if startLabel == endLabel {
		return startLabel
	}
	return fmt.Sprintf("%s ~ %s", startLabel, endLabel)
}

func startOfUTCDay(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func classifyOverallTrend(current RiskStats, previous RiskStats) string {
	currentPressure := riskPressureScore(current)
	previousPressure := riskPressureScore(previous)

	if current.High > previous.High {
		return "上升"
	}

	if current.High < previous.High && currentPressure < previousPressure {
		return "下降"
	}

	if previous.Total == 0 {
		if current.High >= 2 || currentPressure >= 10 {
			return "上升"
		}
		return "平稳"
	}

	return classifyTrendWithThreshold(currentPressure, previousPressure, 3, 0.25)
}

func classifyHighRiskTrend(currentHigh, previousHigh int) string {
	if previousHigh == 0 {
		if currentHigh > 0 {
			return "上升"
		}
		return "平稳"
	}

	return classifyTrendWithThreshold(currentHigh, previousHigh, 1, 0.2)
}

func classifyTrendWithThreshold(current, previous, minAbsDelta int, minRelativeDelta float64) string {
	delta := current - previous
	if delta == 0 {
		return "平稳"
	}

	absDelta := absInt(delta)
	if absDelta < minAbsDelta {
		return "平稳"
	}

	baseline := maxInt(previous, 1)
	relativeDelta := float64(absDelta) / float64(baseline)
	if relativeDelta < minRelativeDelta {
		return "平稳"
	}

	if delta > 0 {
		return "上升"
	}
	return "下降"
}

func riskPressureScore(stats RiskStats) int {
	return stats.High*3 + stats.Medium*2 + stats.Low
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func buildTimeBucket(t time.Time, interval string) string {
	utc := t.UTC()
	switch interval {
	case IntervalWeek:
		year, week := utc.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case IntervalMonth:
		return utc.Format("2006-01")
	default:
		return utc.Format("2006-01-02")
	}
}

func normalizeRiskLevel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "高":
		return "高"
	case "低":
		return "低"
	default:
		return "中"
	}
}
