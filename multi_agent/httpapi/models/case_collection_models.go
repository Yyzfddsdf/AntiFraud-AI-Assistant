package models

// CaseCollectionRequest 是案件采集接口请求体。
type CaseCollectionRequest struct {
	Query     string `json:"query"`
	CaseCount int    `json:"case_count"`
}
