package model

import "time"

// TaskPayload 保存任务原始输入和各子模态解读结果。
type TaskPayload struct {
	Text          string   `json:"text"`
	Videos        []string `json:"videos"`
	Audios        []string `json:"audios"`
	Images        []string `json:"images"`
	VideoInsights []string `json:"video_insights,omitempty"`
	AudioInsights []string `json:"audio_insights,omitempty"`
	ImageInsights []string `json:"image_insights,omitempty"`
}

// TaskRecord 表示“任务视角”的统一记录模型。
type TaskRecord struct {
	TaskID     string      `json:"task_id"`
	UserID     string      `json:"user_id"`
	Title      string      `json:"title"`
	Status     string      `json:"status"`
	ScamType   string      `json:"scam_type,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Payload    TaskPayload `json:"payload"`
	Summary    string      `json:"summary"`
	Report     string      `json:"report,omitempty"`
	Error      string      `json:"error,omitempty"`
	HistoryRef string      `json:"history_ref,omitempty"`
}

// CaseHistoryRecord 表示“历史案件视角”的归档记录模型。
type CaseHistoryRecord struct {
	RecordID    string      `json:"record_id"`
	UserID      string      `json:"user_id"`
	Title       string      `json:"title"`
	Status      string      `json:"status"`
	CaseSummary string      `json:"case_summary"`
	ScamType    string      `json:"scam_type,omitempty"`
	RiskLevel   string      `json:"risk_level"`
	CreatedAt   time.Time   `json:"created_at"`
	Payload     TaskPayload `json:"payload"`
	Report      string      `json:"report,omitempty"`
}

// UserStateView 是用户维度的聚合视图模型。
type UserStateView struct {
	UserID  string                `json:"user_id"`
	Pending map[string]TaskRecord `json:"pending"`
	History []CaseHistoryRecord   `json:"history"`
}

// PendingTaskEntity 是 pending_tasks 表的 ORM 映射实体。
type PendingTaskEntity struct {
	TaskID string `gorm:"primaryKey;size:64"`
	UserID string `gorm:"index;not null"`
	Title  string `gorm:"size:255;not null"`
	Status string `gorm:"size:32;index;not null"`

	PayloadText          string `gorm:"type:text"`
	PayloadVideos        string `gorm:"type:text"`
	PayloadAudios        string `gorm:"type:text"`
	PayloadImages        string `gorm:"type:text"`
	PayloadVideoInsights string `gorm:"type:text"`
	PayloadAudioInsights string `gorm:"type:text"`
	PayloadImageInsights string `gorm:"type:text"`

	Report     string `gorm:"type:text"`
	Error      string `gorm:"type:text"`
	HistoryRef string `gorm:"size:64"`

	CreatedAt time.Time `gorm:"index;not null"`
	UpdatedAt time.Time `gorm:"index;not null"`
}

func (PendingTaskEntity) TableName() string {
	return "pending_tasks"
}

// HistoryCaseEntity 是 history_cases 表的 ORM 映射实体。
type HistoryCaseEntity struct {
	RecordID    string `gorm:"primaryKey;size:64"`
	UserID      string `gorm:"index;not null"`
	Title       string `gorm:"size:255;not null"`
	CaseSummary string `gorm:"type:text"`
	ScamType    string `gorm:"size:64;index"`
	Status      string `gorm:"size:32;index;not null"`
	RiskLevel   string `gorm:"size:32;index"`

	PayloadText          string `gorm:"type:text"`
	PayloadVideos        string `gorm:"type:text"`
	PayloadAudios        string `gorm:"type:text"`
	PayloadImages        string `gorm:"type:text"`
	PayloadVideoInsights string `gorm:"type:text"`
	PayloadAudioInsights string `gorm:"type:text"`
	PayloadImageInsights string `gorm:"type:text"`

	Report string `gorm:"type:text"`

	CreatedAt time.Time `gorm:"index;not null"`
	UpdatedAt time.Time `gorm:"index;not null"`
}

func (HistoryCaseEntity) TableName() string {
	return "history_cases"
}
