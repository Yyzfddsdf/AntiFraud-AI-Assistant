package models

// PendingReviewPreviewItem 待审核案件预览条目。
type PendingReviewPreviewItem struct {
	RecordID    string `json:"record_id"`
	Title       string `json:"title"`
	TargetGroup string `json:"target_group"`
	RiskLevel   string `json:"risk_level"`
	ScamType    string `json:"scam_type"`
	ViolatedLaw string `json:"violated_law"`
	CreatedAt   string `json:"created_at"`
}

// PendingReviewPreviewResponse 待审核案件预览列表响应体。
type PendingReviewPreviewResponse struct {
	Total int                        `json:"total"`
	Cases []PendingReviewPreviewItem `json:"cases"`
}

// PendingReviewDetailItem 待审核案件详情条目。
type PendingReviewDetailItem struct {
	RecordID        string   `json:"record_id"`
	UserID          string   `json:"user_id"`
	Title           string   `json:"title"`
	TargetGroup     string   `json:"target_group"`
	RiskLevel       string   `json:"risk_level"`
	ScamType        string   `json:"scam_type"`
	CaseDescription string   `json:"case_description"`
	TypicalScripts  []string `json:"typical_scripts"`
	Keywords        []string `json:"keywords"`
	ViolatedLaw     string   `json:"violated_law"`
	Suggestion      string   `json:"suggestion"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

// PendingReviewDetailResponse 待审核案件详情响应体。
type PendingReviewDetailResponse struct {
	Case PendingReviewDetailItem `json:"case"`
}

// ApproveReviewResponse 审核通过响应体。
type ApproveReviewResponse struct {
	Message string `json:"message"`
	CaseID  string `json:"case_id"`
}
