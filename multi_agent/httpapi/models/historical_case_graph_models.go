package models

// HistoricalCaseGraphNamedCountItem 表示图谱画像中的名称计数项。
type HistoricalCaseGraphNamedCountItem struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// HistoricalCaseGraphRiskStats 表示某诈骗类型的风险分布。
type HistoricalCaseGraphRiskStats struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

// HistoricalCaseGraphSimilarityItem 表示诈骗类型之间的相似关系。
type HistoricalCaseGraphSimilarityItem struct {
	ScamType string  `json:"scam_type"`
	Score    float64 `json:"score"`
}

// HistoricalCaseGraphProfile 表示某个诈骗类型的画像摘要。
type HistoricalCaseGraphProfile struct {
	ScamType         string                              `json:"scam_type"`
	CaseCount        int                                 `json:"case_count"`
	RiskDistribution HistoricalCaseGraphRiskStats        `json:"risk_distribution"`
	TopTargetGroups  []HistoricalCaseGraphNamedCountItem `json:"top_target_groups"`
	TopKeywords      []HistoricalCaseGraphNamedCountItem `json:"top_keywords"`
	SimilarTypes     []HistoricalCaseGraphSimilarityItem `json:"similar_types"`
}

// HistoricalCaseGraphNode 表示图谱节点。
type HistoricalCaseGraphNode struct {
	ID       string `json:"id"`
	NodeType string `json:"node_type"`
	Label    string `json:"label"`
	Weight   int    `json:"weight"`
}

// HistoricalCaseGraphEdge 表示图谱边。
type HistoricalCaseGraphEdge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Relation string  `json:"relation"`
	Score    float64 `json:"score"`
}

// HistoricalCaseGraphData 表示节点与边集合。
type HistoricalCaseGraphData struct {
	Nodes []HistoricalCaseGraphNode `json:"nodes"`
	Edges []HistoricalCaseGraphEdge `json:"edges"`
}

// HistoricalCaseGraphSummary 表示图谱总览信息。
type HistoricalCaseGraphSummary struct {
	FocusType        string `json:"focus_type,omitempty"`
	TopK             int    `json:"top_k"`
	TotalCases       int    `json:"total_cases"`
	ScamTypeCount    int    `json:"scam_type_count"`
	TargetGroupCount int    `json:"target_group_count"`
	KeywordCount     int    `json:"keyword_count"`
}

// HistoricalCaseGraphResponse 表示案件知识库图谱分析响应。
type HistoricalCaseGraphResponse struct {
	Summary  HistoricalCaseGraphSummary   `json:"summary"`
	Profiles []HistoricalCaseGraphProfile `json:"profiles"`
	Graph    HistoricalCaseGraphData      `json:"graph"`
}
