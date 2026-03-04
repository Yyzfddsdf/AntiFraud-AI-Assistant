package overview

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"antifraud/multi_agent/state"
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

// UserRiskOverview 是用户维度的风险总览结果。
type UserRiskOverview struct {
	UserID   string           `json:"user_id"`
	Interval string           `json:"interval"`
	Stats    RiskStats        `json:"stats"`
	Trend    []RiskTrendPoint `json:"trend"`
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
	}
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
