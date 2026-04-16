package models

// HistoricalCaseStatisticsDimensionItem 维度统计项（诈骗类型/目标人群）。
type HistoricalCaseStatisticsDimensionItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// HistoricalCaseStatisticsTrendItem 时间趋势统计项。
type HistoricalCaseStatisticsTrendItem struct {
	TimeBucket string `json:"time_bucket"`
	Count      int    `json:"count"`
}

// HistoricalCaseStatisticsOverviewResponse 历史案件库总览统计响应体。
type HistoricalCaseStatisticsOverviewResponse struct {
	Interval      string                                  `json:"interval"`
	Total         int                                     `json:"total"`
	ByScamType    []HistoricalCaseStatisticsDimensionItem `json:"by_scam_type"`
	ByTargetGroup []HistoricalCaseStatisticsDimensionItem `json:"by_target_group"`
	Trend         []HistoricalCaseStatisticsTrendItem     `json:"trend"`
}
