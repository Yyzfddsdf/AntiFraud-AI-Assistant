package httpapi

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"antifraud/multi_agent/case_library"
	apimodel "antifraud/multi_agent/httpapi/models"
)

const (
	historicalCaseStatsIntervalDay   = "day"
	historicalCaseStatsIntervalWeek  = "week"
	historicalCaseStatsIntervalMonth = "month"
)

const historicalCaseStatsUnknownBucket = "unknown"
const historicalCaseStatsUnknownCategory = "未知"

func normalizeHistoricalCaseStatisticsInterval(raw string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", historicalCaseStatsIntervalDay:
		return historicalCaseStatsIntervalDay, true
	case historicalCaseStatsIntervalWeek:
		return historicalCaseStatsIntervalWeek, true
	case historicalCaseStatsIntervalMonth:
		return historicalCaseStatsIntervalMonth, true
	default:
		return "", false
	}
}

func buildHistoricalCaseStatistics(interval string) (apimodel.HistoricalCaseStatisticsOverviewResponse, error) {
	normalized, ok := normalizeHistoricalCaseStatisticsInterval(interval)
	if !ok {
		normalized = historicalCaseStatsIntervalDay
	}

	byScamTypeCounter := map[string]int{}
	byTargetGroupCounter := map[string]int{}
	trendCounter := map[string]int{}
	total := 0

	err := case_library.StreamHistoricalCasePreviews(func(p case_library.HistoricalCasePreview) error {
		byScamTypeCounter[normalizeHistoricalCaseStatisticsDimension(p.ScamType)]++
		byTargetGroupCounter[normalizeHistoricalCaseStatisticsDimension(p.TargetGroup)]++
		trendCounter[buildHistoricalCaseStatisticsTimeBucket(p.CreatedAt, normalized)]++
		total++
		return nil
	})
	if err != nil {
		return apimodel.HistoricalCaseStatisticsOverviewResponse{}, err
	}

	return apimodel.HistoricalCaseStatisticsOverviewResponse{
		Interval:      normalized,
		Total:         total,
		ByScamType:    sortedHistoricalCaseStatisticsDimensions(byScamTypeCounter),
		ByTargetGroup: sortedHistoricalCaseStatisticsDimensions(byTargetGroupCounter),
		Trend:         sortedHistoricalCaseStatisticsTrend(trendCounter),
	}, nil
}

// buildHistoricalCaseStatisticsFromPreviews 基于预览列表构建统计（仅用于测试或向后兼容）。
func buildHistoricalCaseStatisticsFromPreviews(previews []case_library.HistoricalCasePreview, interval string) apimodel.HistoricalCaseStatisticsOverviewResponse {
	normalized, _ := normalizeHistoricalCaseStatisticsInterval(interval)
	if normalized == "" {
		normalized = historicalCaseStatsIntervalDay
	}

	byScamTypeCounter := map[string]int{}
	byTargetGroupCounter := map[string]int{}
	trendCounter := map[string]int{}

	for _, item := range previews {
		byScamTypeCounter[normalizeHistoricalCaseStatisticsDimension(item.ScamType)]++
		byTargetGroupCounter[normalizeHistoricalCaseStatisticsDimension(item.TargetGroup)]++
		trendCounter[buildHistoricalCaseStatisticsTimeBucket(item.CreatedAt, normalized)]++
	}

	return apimodel.HistoricalCaseStatisticsOverviewResponse{
		Interval:      normalized,
		Total:         len(previews),
		ByScamType:    sortedHistoricalCaseStatisticsDimensions(byScamTypeCounter),
		ByTargetGroup: sortedHistoricalCaseStatisticsDimensions(byTargetGroupCounter),
		Trend:         sortedHistoricalCaseStatisticsTrend(trendCounter),
	}
}

func normalizeHistoricalCaseStatisticsDimension(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return historicalCaseStatsUnknownCategory
	}
	return trimmed
}

func buildHistoricalCaseStatisticsTimeBucket(t time.Time, interval string) string {
	if t.IsZero() {
		return historicalCaseStatsUnknownBucket
	}

	utc := t.UTC()
	switch interval {
	case historicalCaseStatsIntervalWeek:
		year, week := utc.ISOWeek()
		return fmt.Sprintf("%04d-W%02d", year, week)
	case historicalCaseStatsIntervalMonth:
		return utc.Format("2006-01")
	default:
		return utc.Format("2006-01-02")
	}
}

func sortedHistoricalCaseStatisticsDimensions(counter map[string]int) []apimodel.HistoricalCaseStatisticsDimensionItem {
	items := make([]apimodel.HistoricalCaseStatisticsDimensionItem, 0, len(counter))
	for name, count := range counter {
		items = append(items, apimodel.HistoricalCaseStatisticsDimensionItem{
			Name:  name,
			Count: count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	return items
}

func sortedHistoricalCaseStatisticsTrend(counter map[string]int) []apimodel.HistoricalCaseStatisticsTrendItem {
	items := make([]apimodel.HistoricalCaseStatisticsTrendItem, 0, len(counter))
	for bucket, count := range counter {
		items = append(items, apimodel.HistoricalCaseStatisticsTrendItem{
			TimeBucket: bucket,
			Count:      count,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].TimeBucket == items[j].TimeBucket {
			return items[i].Count > items[j].Count
		}
		return items[i].TimeBucket < items[j].TimeBucket
	})
	return items
}
