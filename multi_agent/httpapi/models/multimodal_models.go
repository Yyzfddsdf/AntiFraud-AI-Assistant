package models

// MultimodalScamAnalyzeRequest 多模态诈骗智能助手请求体。
type MultimodalScamAnalyzeRequest struct {
	Text   string   `json:"text"`
	Videos []string `json:"videos"`
	Audios []string `json:"audios"`
	Images []string `json:"images"`
}

// MultimodalScamEnqueueResponse 多模态分析任务入队响应。
type MultimodalScamEnqueueResponse struct {
	TaskID  string `json:"task_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MultimodalTaskPayload 多模态任务输入。
type MultimodalTaskPayload struct {
	Text          string   `json:"text"`
	Videos        []string `json:"videos"`
	Audios        []string `json:"audios"`
	Images        []string `json:"images"`
	VideoInsights []string `json:"video_insights,omitempty"`
	AudioInsights []string `json:"audio_insights,omitempty"`
	ImageInsights []string `json:"image_insights,omitempty"`
}

// MultimodalTaskItem 多模态任务详情。
type MultimodalTaskItem struct {
	TaskID     string                `json:"task_id"`
	UserID     string                `json:"user_id"`
	Title      string                `json:"title"`
	Status     string                `json:"status"`
	ScamType   string                `json:"scam_type,omitempty"`
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
	Payload    MultimodalTaskPayload `json:"payload"`
	Summary    string                `json:"summary"`
	Report     string                `json:"report,omitempty"`
	Error      string                `json:"error,omitempty"`
	HistoryRef string                `json:"history_ref,omitempty"`
}

// MultimodalTaskListItem 多模态任务状态列表条目（轻量，不返回原始payload）。
type MultimodalTaskListItem struct {
	TaskID    string `json:"task_id"`
	UserID    string `json:"user_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MultimodalTaskStateResponse 多模态任务状态列表。
type MultimodalTaskStateResponse struct {
	UserID string                   `json:"user_id"`
	Tasks  []MultimodalTaskListItem `json:"tasks"`
}

// MultimodalHistoryItem 历史案件条目（仅元数据，不含 payload/report）。
type MultimodalHistoryItem struct {
	RecordID    string `json:"record_id"`
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	ScamType    string `json:"scam_type,omitempty"`
	RiskLevel   string `json:"risk_level"`
	CreatedAt   string `json:"created_at"`
}

// MultimodalHistoryResponse 历史案件明细列表。
type MultimodalHistoryResponse struct {
	UserID  string                  `json:"user_id"`
	History []MultimodalHistoryItem `json:"history"`
}

// DeleteMultimodalHistoryResponse 删除历史案件响应体。
type DeleteMultimodalHistoryResponse struct {
	UserID   string `json:"user_id"`
	RecordID string `json:"record_id"`
	Message  string `json:"message"`
}

// MultimodalTaskDetailResponse 单任务查询响应。
type MultimodalTaskDetailResponse struct {
	Task MultimodalTaskItem `json:"task"`
}

// UpdateUserAgeRequest 更新用户年龄请求。
type UpdateUserAgeRequest struct {
	Age int `json:"age"`
}

// UpdateUserAgeResponse 更新用户年龄响应。
type UpdateUserAgeResponse struct {
	UserID  string `json:"user_id"`
	Age     int    `json:"age"`
	Message string `json:"message"`
}

// MultimodalRiskLevelStats 风险等级统计（高/中/低）。
type MultimodalRiskLevelStats struct {
	High   int `json:"high"`
	Medium int `json:"medium"`
	Low    int `json:"low"`
	Total  int `json:"total"`
}

// MultimodalRiskTrendItem 风险变化趋势条目（按时间桶聚合）。
type MultimodalRiskTrendItem struct {
	TimeBucket string `json:"time_bucket"`
	High       int    `json:"high"`
	Medium     int    `json:"medium"`
	Low        int    `json:"low"`
	Total      int    `json:"total"`
}

// MultimodalRiskTrendAnalysis 风险趋势中文分析结果。
type MultimodalRiskTrendAnalysis struct {
	CurrentBucket  string `json:"current_bucket"`
	PreviousBucket string `json:"previous_bucket,omitempty"`
	OverallTrend   string `json:"overall_trend"`
	HighRiskTrend  string `json:"high_risk_trend"`
	Summary        string `json:"summary"`
}

// MultimodalRiskOverviewResponse 用户风险总览响应。
type MultimodalRiskOverviewResponse struct {
	Stats    MultimodalRiskLevelStats    `json:"stats"`
	Trend    []MultimodalRiskTrendItem   `json:"trend"`
	Analysis MultimodalRiskTrendAnalysis `json:"analysis"`
}
