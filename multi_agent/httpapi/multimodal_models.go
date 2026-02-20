package httpapi

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
	CreatedAt  string                `json:"created_at"`
	UpdatedAt  string                `json:"updated_at"`
	Payload    MultimodalTaskPayload `json:"payload"`
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
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// MultimodalTaskStateResponse 多模态任务状态列表。
type MultimodalTaskStateResponse struct {
	UserID string                   `json:"user_id"`
	Tasks  []MultimodalTaskListItem `json:"tasks"`
}

// MultimodalHistoryItem 历史案件条目（详细，包含payload/base64）。
type MultimodalHistoryItem struct {
	RecordID    string                `json:"record_id"`
	Title       string                `json:"title"`
	CaseSummary string                `json:"case_summary"`
	RiskLevel   string                `json:"risk_level"`
	CreatedAt   string                `json:"created_at"`
	Payload     MultimodalTaskPayload `json:"payload"`
	Report      string                `json:"report,omitempty"`
}

// MultimodalHistoryResponse 历史案件明细列表。
type MultimodalHistoryResponse struct {
	UserID  string                  `json:"user_id"`
	History []MultimodalHistoryItem `json:"history"`
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
