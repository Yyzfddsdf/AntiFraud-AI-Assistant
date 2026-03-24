package models

// CreateHistoricalCaseRequest 上传历史案件请求体。
type CreateHistoricalCaseRequest struct {
	Title           string   `json:"title"`
	TargetGroup     string   `json:"target_group"`
	RiskLevel       string   `json:"risk_level"`
	ScamType        string   `json:"scam_type"`
	CaseDescription string   `json:"case_description"`
	TypicalScripts  []string `json:"typical_scripts"`
	Keywords        []string `json:"keywords"`
	ViolatedLaw     string   `json:"violated_law"`
	Suggestion      string   `json:"suggestion"`
}

// HistoricalCaseItem 历史案件入库成功后的回显信息。
type HistoricalCaseItem struct {
	CaseID             string   `json:"case_id"`
	CreatedBy          string   `json:"created_by"`
	Title              string   `json:"title"`
	TargetGroup        string   `json:"target_group"`
	RiskLevel          string   `json:"risk_level"`
	ScamType           string   `json:"scam_type"`
	CaseDescription    string   `json:"case_description"`
	TypicalScripts     []string `json:"typical_scripts"`
	Keywords           []string `json:"keywords"`
	ViolatedLaw        string   `json:"violated_law"`
	Suggestion         string   `json:"suggestion"`
	EmbeddingModel     string   `json:"embedding_model"`
	EmbeddingDimension int      `json:"embedding_dimension"`
	CreatedAt          string   `json:"created_at"`
}

// CreateHistoricalCaseResponse 上传历史案件成功响应体。
type CreateHistoricalCaseResponse struct {
	Message string             `json:"message"`
	Case    HistoricalCaseItem `json:"case"`
}

// HistoricalCasePreviewItem 历史案件预览条目。
type HistoricalCasePreviewItem struct {
	CaseID      string `json:"case_id"`
	Title       string `json:"title"`
	TargetGroup string `json:"target_group"`
	RiskLevel   string `json:"risk_level"`
	ScamType    string `json:"scam_type"`
}

// HistoricalCasePreviewResponse 历史案件预览列表响应体。
type HistoricalCasePreviewResponse struct {
	Total      int                         `json:"total"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	TotalPages int                         `json:"total_pages"`
	HasNext    bool                        `json:"has_next"`
	HasPrev    bool                        `json:"has_prev"`
	Cases      []HistoricalCasePreviewItem `json:"cases"`
}

// HistoricalCaseDetailItem 历史案件详情条目（包含向量）。
type HistoricalCaseDetailItem struct {
	CaseID             string    `json:"case_id"`
	CreatedBy          string    `json:"created_by"`
	Title              string    `json:"title"`
	TargetGroup        string    `json:"target_group"`
	RiskLevel          string    `json:"risk_level"`
	ScamType           string    `json:"scam_type"`
	CaseDescription    string    `json:"case_description"`
	TypicalScripts     []string  `json:"typical_scripts"`
	Keywords           []string  `json:"keywords"`
	ViolatedLaw        string    `json:"violated_law"`
	Suggestion         string    `json:"suggestion"`
	EmbeddingVector    []float64 `json:"embedding_vector"`
	EmbeddingModel     string    `json:"embedding_model"`
	EmbeddingDimension int       `json:"embedding_dimension"`
	CreatedAt          string    `json:"created_at"`
	UpdatedAt          string    `json:"updated_at"`
}

// HistoricalCaseDetailResponse 历史案件详情响应体。
type HistoricalCaseDetailResponse struct {
	Case HistoricalCaseDetailItem `json:"case"`
}

// DeleteHistoricalCaseResponse 删除历史案件响应体。
type DeleteHistoricalCaseResponse struct {
	CaseID  string `json:"case_id"`
	Message string `json:"message"`
}
