package model

import "time"

// ValidationError 表示业务参数校验错误。
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// CreateHistoricalCaseInput 表示创建历史案件的输入载荷。
type CreateHistoricalCaseInput struct {
	Title           string
	TargetGroup     string
	RiskLevel       string
	ScamType        string
	CaseDescription string
	TypicalScripts  []string
	Keywords        []string
	ViolatedLaw     string
	Suggestion      string
}

// HistoricalCaseRecord 表示历史案件完整记录模型。
type HistoricalCaseRecord struct {
	CaseID             string
	CreatedBy          string
	Title              string
	TargetGroup        string
	RiskLevel          string
	ScamType           string
	CaseDescription    string
	TypicalScripts     []string
	Keywords           []string
	ViolatedLaw        string
	Suggestion         string
	EmbeddingVector    []float64
	EmbeddingModel     string
	EmbeddingDimension int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// HistoricalCasePreview 表示历史案件预览模型。
type HistoricalCasePreview struct {
	CaseID      string
	Title       string
	TargetGroup string
	RiskLevel   string
	ScamType    string
}

// HistoricalCaseEntity 是 historical_case_library 表 ORM 映射实体。
type HistoricalCaseEntity struct {
	ID                 uint      `gorm:"primaryKey"`
	CaseID             string    `gorm:"size:32;uniqueIndex;not null"`
	CreatedBy          string    `gorm:"size:64;index;not null"`
	Title              string    `gorm:"type:text;not null"`
	TargetGroup        string    `gorm:"size:32;index;not null"`
	RiskLevel          string    `gorm:"size:16;index;not null;default:'中'"`
	ScamType           string    `gorm:"size:64;index;not null;default:'其他诈骗类'"`
	CaseDescription    string    `gorm:"type:text;not null"`
	TypicalScripts     string    `gorm:"type:text;not null"`
	Keywords           string    `gorm:"type:text;not null"`
	ViolatedLaw        string    `gorm:"type:text;not null"`
	Suggestion         string    `gorm:"type:text;not null"`
	EmbeddingVector    string    `gorm:"type:text;not null"`
	EmbeddingModel     string    `gorm:"size:128;not null"`
	EmbeddingDimension int       `gorm:"not null"`
	CreatedAt          time.Time `gorm:"index"`
	UpdatedAt          time.Time
}

func (HistoricalCaseEntity) TableName() string {
	return "historical_case_library"
}
