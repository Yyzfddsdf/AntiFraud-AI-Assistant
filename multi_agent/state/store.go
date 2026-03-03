package state

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"image_recognition/login_system/database"

	"gorm.io/gorm"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"

	pendingTaskTTL = 20 * time.Minute
)

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

// TaskRecord 表示进行中任务的完整状态。
type TaskRecord struct {
	TaskID     string      `json:"task_id"`
	UserID     string      `json:"user_id"`
	Title      string      `json:"title"`
	Status     string      `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	Payload    TaskPayload `json:"payload"`
	Report     string      `json:"report,omitempty"`
	Error      string      `json:"error,omitempty"`
	HistoryRef string      `json:"history_ref,omitempty"`
}

// CaseHistoryRecord 表示归档后的历史案件记录。
type CaseHistoryRecord struct {
	RecordID    string      `json:"record_id"`
	UserID      string      `json:"user_id"`
	Title       string      `json:"title"`
	CaseSummary string      `json:"case_summary"`
	RiskLevel   string      `json:"risk_level"`
	CreatedAt   time.Time   `json:"created_at"`
	Payload     TaskPayload `json:"payload"`
	Report      string      `json:"report,omitempty"`
}

// UserStateView 聚合用户进行中任务与历史记录，供接口直接返回。
type UserStateView struct {
	UserID  string                `json:"user_id"`
	Pending map[string]TaskRecord `json:"pending"`
	History []CaseHistoryRecord   `json:"history"`
}

type pendingTaskEntity struct {
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

func (pendingTaskEntity) TableName() string {
	return "pending_tasks"
}

type historyCaseEntity struct {
	RecordID    string `gorm:"primaryKey;size:64"`
	UserID      string `gorm:"index;not null"`
	Title       string `gorm:"size:255;not null"`
	CaseSummary string `gorm:"type:text"`
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

func (historyCaseEntity) TableName() string {
	return "history_cases"
}

var stateSchemaOnce sync.Once

// ensureStateSchema 确保状态相关表结构存在。
func ensureStateSchema(db *gorm.DB) {
	if db == nil {
		return
	}
	stateSchemaOnce.Do(func() {
		if err := db.AutoMigrate(&pendingTaskEntity{}, &historyCaseEntity{}); err != nil {
			log.Printf("[state] auto migrate state tables failed: %v", err)
		}
	})
}

// CreateTask 创建任务并落库到 pending_tasks。
func CreateTask(userID string, payload TaskPayload) TaskRecord {
	uid := normalizeUserID(userID)
	now := time.Now()
	task := TaskRecord{
		TaskID:    newID("TASK"),
		UserID:    uid,
		Title:     normalizeTaskTitle(payload),
		Status:    TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Payload: TaskPayload{
			Text:          strings.TrimSpace(payload.Text),
			Videos:        append([]string{}, payload.Videos...),
			Audios:        append([]string{}, payload.Audios...),
			Images:        append([]string{}, payload.Images...),
			VideoInsights: append([]string{}, payload.VideoInsights...),
			AudioInsights: append([]string{}, payload.AudioInsights...),
			ImageInsights: append([]string{}, payload.ImageInsights...),
		},
	}

	db := database.DB
	if db == nil {
		log.Printf("[state] create task skipped: db not initialized")
		return task
	}
	ensureStateSchema(db)

	entity := pendingEntityFromTask(task)
	if err := db.Create(&entity).Error; err != nil {
		log.Printf("[state] create pending task failed: user=%s task=%s err=%v", uid, task.TaskID, err)
	}

	return task
}

// MarkTaskProcessing 将任务状态更新为 processing。
func MarkTaskProcessing(userID, taskID string) {
	db := database.DB
	if db == nil {
		return
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return
	}

	if err := db.Model(&pendingTaskEntity{}).
		Where("task_id = ? AND user_id = ?", tid, uid).
		Updates(map[string]interface{}{
			"status":     TaskStatusProcessing,
			"updated_at": time.Now(),
		}).Error; err != nil {
		log.Printf("[state] mark processing failed: user=%s task=%s err=%v", uid, tid, err)
	}
}

// MarkTaskCompleted 将任务从 pending 迁移到 history，并写入最终报告。
func MarkTaskCompleted(userID, taskID, report string) {
	db := database.DB
	if db == nil {
		return
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return
	}
	trimmedReport := strings.TrimSpace(report)

	err := db.Transaction(func(tx *gorm.DB) error {
		var pending pendingTaskEntity
		if err := tx.Where("task_id = ? AND user_id = ?", tid, uid).First(&pending).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		var existing historyCaseEntity
		historyErr := tx.Where("record_id = ? AND user_id = ?", tid, uid).First(&existing).Error
		if historyErr != nil {
			if !errors.Is(historyErr, gorm.ErrRecordNotFound) {
				return historyErr
			}

			summary := trimmedReport
			if summary == "" {
				summary = strings.TrimSpace(pending.Report)
			}
			if summary == "" {
				summary = "task completed"
			}

			history := historyCaseEntity{
				RecordID:             tid,
				UserID:               uid,
				Title:                normalizeCaseTitle(pending.Title, summary),
				CaseSummary:          summary,
				Status:               TaskStatusCompleted,
				RiskLevel:            normalizeRiskLevel(""),
				PayloadText:          pending.PayloadText,
				PayloadVideos:        pending.PayloadVideos,
				PayloadAudios:        pending.PayloadAudios,
				PayloadImages:        pending.PayloadImages,
				PayloadVideoInsights: pending.PayloadVideoInsights,
				PayloadAudioInsights: pending.PayloadAudioInsights,
				PayloadImageInsights: pending.PayloadImageInsights,
				Report:               firstNonEmpty(trimmedReport, pending.Report),
				CreatedAt:            pending.CreatedAt,
				UpdatedAt:            time.Now(),
			}

			if err := tx.Create(&history).Error; err != nil {
				return err
			}
		} else if trimmedReport != "" {
			if err := tx.Model(&historyCaseEntity{}).
				Where("record_id = ? AND user_id = ?", tid, uid).
				Updates(map[string]interface{}{"report": trimmedReport, "updated_at": time.Now()}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("task_id = ? AND user_id = ?", tid, uid).Delete(&pendingTaskEntity{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("[state] mark completed failed: user=%s task=%s err=%v", uid, tid, err)
	}
}

// UpdateTaskInsights 更新任务的子模态解读摘要。
func UpdateTaskInsights(userID, taskID string, videoInsights []string, audioInsights []string, imageInsights []string) {
	db := database.DB
	if db == nil {
		return
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return
	}

	if err := db.Model(&pendingTaskEntity{}).
		Where("task_id = ? AND user_id = ?", tid, uid).
		Updates(map[string]interface{}{
			"payload_video_insights": encodeStringList(videoInsights),
			"payload_audio_insights": encodeStringList(audioInsights),
			"payload_image_insights": encodeStringList(imageInsights),
			"updated_at":             time.Now(),
		}).Error; err != nil {
		log.Printf("[state] update insights failed: user=%s task=%s err=%v", uid, tid, err)
	}
}

// MarkTaskFailed 将失败任务写入历史并从 pending 删除。
func MarkTaskFailed(userID, taskID, errMsg string) {
	db := database.DB
	if db == nil {
		return
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var pending pendingTaskEntity
		if err := tx.Where("task_id = ? AND user_id = ?", tid, uid).First(&pending).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}

		reason := strings.TrimSpace(errMsg)
		if reason == "" {
			reason = "task execution failed"
		}

		history := historyCaseEntity{
			RecordID:             tid,
			UserID:               uid,
			Title:                normalizeCaseTitle(pending.Title, reason),
			CaseSummary:          reason,
			Status:               TaskStatusFailed,
			RiskLevel:            normalizeRiskLevel("中"),
			PayloadText:          pending.PayloadText,
			PayloadVideos:        pending.PayloadVideos,
			PayloadAudios:        pending.PayloadAudios,
			PayloadImages:        pending.PayloadImages,
			PayloadVideoInsights: pending.PayloadVideoInsights,
			PayloadAudioInsights: pending.PayloadAudioInsights,
			PayloadImageInsights: pending.PayloadImageInsights,
			Report:               reason,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}

		if err := tx.Where("record_id = ? AND user_id = ?", tid, uid).Delete(&historyCaseEntity{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&history).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ? AND user_id = ?", tid, uid).Delete(&pendingTaskEntity{}).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Printf("[state] mark failed failed: user=%s task=%s err=%v", uid, tid, err)
	}
}

// GetTask 查询进行中任务。
func GetTask(userID, taskID string) (TaskRecord, bool) {
	db := database.DB
	if db == nil {
		return TaskRecord{}, false
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return TaskRecord{}, false
	}
	expireStalePendingTasks(uid, tid)

	var entity pendingTaskEntity
	query := db.Where("task_id = ? AND user_id = ?", tid, uid).Limit(1).Find(&entity)
	if query.Error != nil {
		return TaskRecord{}, false
	}
	if query.RowsAffected == 0 {
		return TaskRecord{}, false
	}
	return taskFromPendingEntity(entity), true
}

// GetTaskDetailByID 优先查 pending，再查 history，统一返回任务详情。
func GetTaskDetailByID(userID, id string) (TaskRecord, bool) {
	db := database.DB
	if db == nil {
		return TaskRecord{}, false
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	targetID := strings.TrimSpace(id)
	if targetID == "" {
		return TaskRecord{}, false
	}
	expireStalePendingTasks(uid, targetID)

	var pending pendingTaskEntity
	pendingQuery := db.Where("task_id = ? AND user_id = ?", targetID, uid).Limit(1).Find(&pending)
	if pendingQuery.Error != nil {
		return TaskRecord{}, false
	}
	if pendingQuery.RowsAffected > 0 {
		return taskFromPendingEntity(pending), true
	}

	var history historyCaseEntity
	historyQuery := db.Where("record_id = ? AND user_id = ?", targetID, uid).Limit(1).Find(&history)
	if historyQuery.Error != nil {
		return TaskRecord{}, false
	}
	if historyQuery.RowsAffected == 0 {
		return TaskRecord{}, false
	}

	report := strings.TrimSpace(history.Report)
	if report == "" {
		report = strings.TrimSpace(history.CaseSummary)
	}
	status := strings.TrimSpace(history.Status)
	if status == "" {
		status = TaskStatusCompleted
	}

	return TaskRecord{
		TaskID:    history.RecordID,
		UserID:    history.UserID,
		Title:     history.Title,
		Status:    status,
		CreatedAt: history.CreatedAt,
		UpdatedAt: history.UpdatedAt,
		Payload: TaskPayload{
			Text:          history.PayloadText,
			Videos:        decodeStringList(history.PayloadVideos),
			Audios:        decodeStringList(history.PayloadAudios),
			Images:        decodeStringList(history.PayloadImages),
			VideoInsights: decodeStringList(history.PayloadVideoInsights),
			AudioInsights: decodeStringList(history.PayloadAudioInsights),
			ImageInsights: decodeStringList(history.PayloadImageInsights),
		},
		Report: report,
	}, true
}

// GetUserStateView 返回用户任务总览（进行中 + 历史）。
func GetUserStateView(userID string) UserStateView {
	db := database.DB
	uid := normalizeUserID(userID)
	if db == nil {
		return UserStateView{UserID: uid, Pending: map[string]TaskRecord{}, History: []CaseHistoryRecord{}}
	}
	ensureStateSchema(db)
	expireStalePendingTasks(uid, "")

	pendingRows := make([]pendingTaskEntity, 0)
	if err := db.Where("user_id = ?", uid).Find(&pendingRows).Error; err != nil {
		log.Printf("[state] query pending failed: user=%s err=%v", uid, err)
	}

	historyRows := make([]historyCaseEntity, 0)
	if err := db.Where("user_id = ?", uid).Order("created_at desc").Find(&historyRows).Error; err != nil {
		log.Printf("[state] query history failed: user=%s err=%v", uid, err)
	}

	pending := make(map[string]TaskRecord, len(pendingRows))
	for _, row := range pendingRows {
		record := taskFromPendingEntity(row)
		pending[record.TaskID] = record
	}

	history := make([]CaseHistoryRecord, 0, len(historyRows))
	for _, row := range historyRows {
		history = append(history, historyFromEntity(row))
	}

	return UserStateView{
		UserID:  uid,
		Pending: pending,
		History: history,
	}
}

func expireStalePendingTasks(userID, taskID string) {
	db := database.DB
	if db == nil {
		return
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	cutoff := time.Now().Add(-pendingTaskTTL)

	query := db.Model(&pendingTaskEntity{}).
		Where("user_id = ?", uid).
		Where("created_at <= ?", cutoff).
		Where("status IN ?", []string{TaskStatusPending, TaskStatusProcessing})
	if tid != "" {
		query = query.Where("task_id = ?", tid)
	}

	staleRows := make([]pendingTaskEntity, 0)
	if err := query.Find(&staleRows).Error; err != nil {
		log.Printf("[state] query stale pending failed: user=%s task=%s err=%v", uid, firstNonEmpty(tid, "<all>"), err)
		return
	}
	if len(staleRows) == 0 {
		return
	}

	for _, row := range staleRows {
		staleTaskID := strings.TrimSpace(row.TaskID)
		if staleTaskID == "" {
			continue
		}
		if err := db.Where("task_id = ? AND user_id = ?", staleTaskID, uid).Delete(&pendingTaskEntity{}).Error; err != nil {
			log.Printf("[state] delete stale pending failed: user=%s task=%s err=%v", uid, staleTaskID, err)
		}
	}
}

// AddCaseHistory 直接写入历史记录（用于工具显式归档场景）。
func AddCaseHistory(userID, taskID, title, summary, riskLevel string, payload TaskPayload, report string) CaseHistoryRecord {
	uid := normalizeUserID(userID)
	now := time.Now()
	recordID := strings.TrimSpace(taskID)
	if recordID == "" {
		recordID = newID("TASK")
	}

	record := CaseHistoryRecord{
		RecordID:    recordID,
		UserID:      uid,
		Title:       normalizeCaseTitle(title, summary),
		CaseSummary: strings.TrimSpace(summary),
		RiskLevel:   normalizeRiskLevel(riskLevel),
		CreatedAt:   now,
		Payload: TaskPayload{
			Text:          strings.TrimSpace(payload.Text),
			Videos:        append([]string{}, payload.Videos...),
			Audios:        append([]string{}, payload.Audios...),
			Images:        append([]string{}, payload.Images...),
			VideoInsights: append([]string{}, payload.VideoInsights...),
			AudioInsights: append([]string{}, payload.AudioInsights...),
			ImageInsights: append([]string{}, payload.ImageInsights...),
		},
		Report: strings.TrimSpace(report),
	}

	db := database.DB
	if db == nil {
		return record
	}
	ensureStateSchema(db)

	entity := historyEntityFromRecord(record, TaskStatusCompleted)
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("record_id = ? AND user_id = ?", entity.RecordID, entity.UserID).Delete(&historyCaseEntity{}).Error; err != nil {
			return err
		}
		return tx.Create(&entity).Error
	})
	if err != nil {
		log.Printf("[state] add case history failed: user=%s record=%s err=%v", uid, recordID, err)
	}

	return historyFromEntity(entity)
}

// GetCaseHistory 查询用户历史案件列表。
func GetCaseHistory(userID string) []CaseHistoryRecord {
	db := database.DB
	if db == nil {
		return []CaseHistoryRecord{}
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	rows := make([]historyCaseEntity, 0)
	if err := db.Where("user_id = ?", uid).Order("created_at desc").Find(&rows).Error; err != nil {
		log.Printf("[state] get case history failed: user=%s err=%v", uid, err)
		return []CaseHistoryRecord{}
	}

	result := make([]CaseHistoryRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, historyFromEntity(row))
	}
	return result
}

// DeleteCaseHistory 按记录 ID 删除当前用户的一条历史案件。
func DeleteCaseHistory(userID, recordID string) (bool, error) {
	db := database.DB
	if db == nil {
		return false, nil
	}
	ensureStateSchema(db)

	uid := normalizeUserID(userID)
	rid := strings.TrimSpace(recordID)
	if rid == "" {
		return false, nil
	}

	result := db.Where("record_id = ? AND user_id = ?", rid, uid).Delete(&historyCaseEntity{})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func pendingEntityFromTask(task TaskRecord) pendingTaskEntity {
	return pendingTaskEntity{
		TaskID:               strings.TrimSpace(task.TaskID),
		UserID:               normalizeUserID(task.UserID),
		Title:                strings.TrimSpace(task.Title),
		Status:               strings.TrimSpace(task.Status),
		PayloadText:          strings.TrimSpace(task.Payload.Text),
		PayloadVideos:        encodeStringList(task.Payload.Videos),
		PayloadAudios:        encodeStringList(task.Payload.Audios),
		PayloadImages:        encodeStringList(task.Payload.Images),
		PayloadVideoInsights: encodeStringList(task.Payload.VideoInsights),
		PayloadAudioInsights: encodeStringList(task.Payload.AudioInsights),
		PayloadImageInsights: encodeStringList(task.Payload.ImageInsights),
		Report:               strings.TrimSpace(task.Report),
		Error:                strings.TrimSpace(task.Error),
		HistoryRef:           strings.TrimSpace(task.HistoryRef),
		CreatedAt:            task.CreatedAt,
		UpdatedAt:            task.UpdatedAt,
	}
}

func taskFromPendingEntity(entity pendingTaskEntity) TaskRecord {
	return TaskRecord{
		TaskID:     strings.TrimSpace(entity.TaskID),
		UserID:     normalizeUserID(entity.UserID),
		Title:      strings.TrimSpace(entity.Title),
		Status:     strings.TrimSpace(entity.Status),
		CreatedAt:  entity.CreatedAt,
		UpdatedAt:  entity.UpdatedAt,
		Report:     strings.TrimSpace(entity.Report),
		Error:      strings.TrimSpace(entity.Error),
		HistoryRef: strings.TrimSpace(entity.HistoryRef),
		Payload: TaskPayload{
			Text:          strings.TrimSpace(entity.PayloadText),
			Videos:        decodeStringList(entity.PayloadVideos),
			Audios:        decodeStringList(entity.PayloadAudios),
			Images:        decodeStringList(entity.PayloadImages),
			VideoInsights: decodeStringList(entity.PayloadVideoInsights),
			AudioInsights: decodeStringList(entity.PayloadAudioInsights),
			ImageInsights: decodeStringList(entity.PayloadImageInsights),
		},
	}
}

func historyEntityFromRecord(record CaseHistoryRecord, status string) historyCaseEntity {
	return historyCaseEntity{
		RecordID:             strings.TrimSpace(record.RecordID),
		UserID:               normalizeUserID(record.UserID),
		Title:                strings.TrimSpace(record.Title),
		CaseSummary:          strings.TrimSpace(record.CaseSummary),
		Status:               strings.TrimSpace(status),
		RiskLevel:            normalizeRiskLevel(record.RiskLevel),
		PayloadText:          strings.TrimSpace(record.Payload.Text),
		PayloadVideos:        encodeStringList(record.Payload.Videos),
		PayloadAudios:        encodeStringList(record.Payload.Audios),
		PayloadImages:        encodeStringList(record.Payload.Images),
		PayloadVideoInsights: encodeStringList(record.Payload.VideoInsights),
		PayloadAudioInsights: encodeStringList(record.Payload.AudioInsights),
		PayloadImageInsights: encodeStringList(record.Payload.ImageInsights),
		Report:               strings.TrimSpace(record.Report),
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            time.Now(),
	}
}

func historyFromEntity(entity historyCaseEntity) CaseHistoryRecord {
	return CaseHistoryRecord{
		RecordID:    strings.TrimSpace(entity.RecordID),
		UserID:      normalizeUserID(entity.UserID),
		Title:       strings.TrimSpace(entity.Title),
		CaseSummary: strings.TrimSpace(entity.CaseSummary),
		RiskLevel:   normalizeRiskLevel(entity.RiskLevel),
		CreatedAt:   entity.CreatedAt,
		Report:      strings.TrimSpace(entity.Report),
		Payload: TaskPayload{
			Text:          strings.TrimSpace(entity.PayloadText),
			Videos:        decodeStringList(entity.PayloadVideos),
			Audios:        decodeStringList(entity.PayloadAudios),
			Images:        decodeStringList(entity.PayloadImages),
			VideoInsights: decodeStringList(entity.PayloadVideoInsights),
			AudioInsights: decodeStringList(entity.PayloadAudioInsights),
			ImageInsights: decodeStringList(entity.PayloadImageInsights),
		},
	}
}

func encodeStringList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	encoded := make([]string, 0, len(items))
	for _, item := range items {
		raw := []byte(strings.TrimSpace(item))
		if len(raw) == 0 {
			continue
		}
		encoded = append(encoded, base64.StdEncoding.EncodeToString(raw))
	}
	return strings.Join(encoded, ",")
}

func decodeStringList(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return []string{}
	}

	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(item)
		if err != nil {
			// Forward compatible fallback for any legacy plain-text rows.
			result = append(result, item)
			continue
		}
		result = append(result, string(decoded))
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func normalizeUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return "demo-user"
	}
	return trimmed
}

func normalizeRiskLevel(level string) string {
	switch strings.TrimSpace(level) {
	case "\u9ad8":
		return "\u9ad8"
	case "\u4f4e":
		return "\u4f4e"
	default:
		return "\u4e2d"
	}
}

func normalizeCaseTitle(title, summary string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed != "" {
		return trimmed
	}
	s := strings.TrimSpace(summary)
	if s == "" {
		return "untitled case"
	}
	runes := []rune(s)
	if len(runes) <= 48 {
		return s
	}
	return string(runes[:48]) + "..."
}

func normalizeTaskTitle(payload TaskPayload) string {
	text := strings.TrimSpace(payload.Text)
	if text != "" {
		runes := []rune(text)
		if len(runes) <= 48 {
			return text
		}
		return string(runes[:48]) + "..."
	}
	return fmt.Sprintf("multimodal task (V%d/A%d/I%d)", len(payload.Videos), len(payload.Audios), len(payload.Images))
}

func newID(prefix string) string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		now := time.Now().UnixNano()
		return fmt.Sprintf("%s-%d", prefix, now)
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(hex.EncodeToString(bytes)))
}
