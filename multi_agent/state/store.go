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

	"antifraud/database"
	model "antifraud/multi_agent/state/model"

	"gorm.io/gorm"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"

	pendingTaskTTL = 20 * time.Minute
)

// 结构体模型已迁移至 multi_agent/state/model 目录。
// 这里保留类型别名，保持 state 包对外 API 与现有调用兼容。
type TaskPayload = model.TaskPayload
type TaskRecord = model.TaskRecord
type CaseHistoryRecord = model.CaseHistoryRecord
type UserStateView = model.UserStateView
type pendingTaskEntity = model.PendingTaskEntity
type historyCaseEntity = model.HistoryCaseEntity

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
				ScamType:             "",
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
			ScamType:             "",
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

	return taskFromHistoryEntity(history), true
}

// GetUserStateView 返回用户任务总览（进行中 + 历史）。
// 设计说明：
// 1) 这是“总览读模型”，优先服务列表与统计场景，而非单任务详情；
// 2) 函数内部对 pending/history 采用并行查询，降低总览接口等待时间；
// 3) 查询使用部分字段（Select），只读取总览必需列，避免加载 payload/report 大字段；
//   - pending_tasks 读取字段：task_id、user_id、title、status、created_at、updated_at
//   - history_cases 读取字段：record_id、user_id、title、case_summary、risk_level、created_at、updated_at
//
// 4) 单任务详情请使用 GetTaskDetailByID（该函数会读取完整字段）。
func GetUserStateView(userID string) UserStateView {
	db := database.DB
	uid := normalizeUserID(userID)
	if db == nil {
		return UserStateView{UserID: uid, Pending: map[string]TaskRecord{}, History: []CaseHistoryRecord{}}
	}
	ensureStateSchema(db)
	expireStalePendingTasks(uid, "")

	pendingRows := make([]pendingTaskEntity, 0)
	historyRows := make([]historyCaseEntity, 0)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		// 总览场景仅需任务列表元信息，不加载 payload/report 等大字段。
		if err := db.Model(&pendingTaskEntity{}).
			Select("task_id", "user_id", "title", "status", "created_at", "updated_at").
			Where("user_id = ?", uid).
			Find(&pendingRows).Error; err != nil {
			log.Printf("[state] query pending failed: user=%s err=%v", uid, err)
		}
	}()

	go func() {
		defer wg.Done()
		// 总览与风险统计仅需历史元数据，不加载 payload/report 等大字段。
		if err := db.Model(&historyCaseEntity{}).
			Select("record_id", "user_id", "title", "case_summary", "scam_type", "risk_level", "created_at", "updated_at").
			Where("user_id = ?", uid).
			Order("created_at desc").
			Find(&historyRows).Error; err != nil {
			log.Printf("[state] query history failed: user=%s err=%v", uid, err)
		}
	}()

	wg.Wait()

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

// expireStalePendingTasks 清理超时未完成的 pending 任务。
// 规则：
// 1) 仅处理 created_at 早于 pendingTaskTTL 的记录；
// 2) 仅清理 pending/processing 状态；
// 3) 当传入 taskID 时只检查该任务，否则检查当前用户全部任务；
// 4) 过期任务直接删除，不迁移到历史表。
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
		Where("created_at <= ?", cutoff).
		Where("status IN ?", []string{TaskStatusPending, TaskStatusProcessing})
	if tid != "" {
		query = query.Where("user_id = ? AND task_id = ?", uid, tid)
	} else {
		query = query.Where("user_id = ?", uid)
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
func AddCaseHistory(userID, taskID, title, summary, scamType, riskLevel string, payload TaskPayload, report string) CaseHistoryRecord {
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
		Status:      TaskStatusCompleted,
		CaseSummary: strings.TrimSpace(summary),
		ScamType:    strings.TrimSpace(scamType),
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

// pendingEntityFromTask 将业务层 TaskRecord 转换为 pending_tasks 表实体。
// 说明：
// 1) 用于 CreateTask 写入数据库时的字段映射；
// 2) 对切片字段做编码存储，避免直接使用复杂类型落库；
// 3) 对字符串字段统一 trim，减少脏数据。
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

// taskFromPendingEntity 将 pending_tasks 表实体转换为业务层 TaskRecord。
// 说明：
// 1) 用于进行中任务查询返回；
// 2) 将编码后的列表字段解码回 []string；
// 3) 保持对外返回结构稳定，不暴露底层存储细节。
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

// historyEntityFromRecord 将业务层 CaseHistoryRecord 转换为 history_cases 表实体。
// 说明：
// 1) 用于历史记录写入时的统一映射；
// 2) status 由调用方显式传入，便于区分 completed/failed；
// 3) UpdatedAt 在写入时刷新为当前时间。
func historyEntityFromRecord(record CaseHistoryRecord, status string) historyCaseEntity {
	return historyCaseEntity{
		RecordID:             strings.TrimSpace(record.RecordID),
		UserID:               normalizeUserID(record.UserID),
		Title:                strings.TrimSpace(record.Title),
		CaseSummary:          strings.TrimSpace(record.CaseSummary),
		ScamType:             strings.TrimSpace(record.ScamType),
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

// historyFromEntity 将 history_cases 表实体转换为业务层 CaseHistoryRecord。
// 说明：
// 1) 用于历史列表/详情返回；
// 2) 自动解码 payload 列表字段；
// 3) 统一风险等级取值，保证前端展示一致。
func historyFromEntity(entity historyCaseEntity) CaseHistoryRecord {
	status := strings.TrimSpace(entity.Status)
	if status == "" {
		status = TaskStatusCompleted
	}

	return CaseHistoryRecord{
		RecordID:    strings.TrimSpace(entity.RecordID),
		UserID:      normalizeUserID(entity.UserID),
		Title:       strings.TrimSpace(entity.Title),
		Status:      status,
		CaseSummary: strings.TrimSpace(entity.CaseSummary),
		ScamType:    strings.TrimSpace(entity.ScamType),
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

// taskFromHistoryEntity 将 history_cases 表实体转换为任务详情模型 TaskRecord。
// 说明：
// 1) 先复用 historyFromEntity 做基础字段标准化，避免重复维护字段映射逻辑；
// 2) summary 直接取历史案件的 case_summary；
// 3) report 为空时回退到 case_summary，保证详情页至少有可读文本；
// 4) pending/processing 不会走此函数，本函数仅用于历史分支。
func taskFromHistoryEntity(entity historyCaseEntity) TaskRecord {
	record := historyFromEntity(entity)
	report := strings.TrimSpace(record.Report)
	if report == "" {
		report = strings.TrimSpace(record.CaseSummary)
	}

	return TaskRecord{
		TaskID:    record.RecordID,
		UserID:    record.UserID,
		Title:     record.Title,
		Status:    record.Status,
		ScamType:  strings.TrimSpace(record.ScamType),
		CreatedAt: record.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
		Payload: TaskPayload{
			Text:          strings.TrimSpace(record.Payload.Text),
			Videos:        append([]string{}, record.Payload.Videos...),
			Audios:        append([]string{}, record.Payload.Audios...),
			Images:        append([]string{}, record.Payload.Images...),
			VideoInsights: append([]string{}, record.Payload.VideoInsights...),
			AudioInsights: append([]string{}, record.Payload.AudioInsights...),
			ImageInsights: append([]string{}, record.Payload.ImageInsights...),
		},
		Summary: strings.TrimSpace(record.CaseSummary),
		Report:  report,
	}
}

// encodeStringList 将字符串数组编码为逗号分隔的 base64 串。
// 说明：
// 1) 用于数据库单字段存储列表值；
// 2) 自动忽略空字符串项；
// 3) 返回空字符串表示空列表。
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

// decodeStringList 将逗号分隔的 base64 串解码为字符串数组。
// 说明：
// 1) 与 encodeStringList 成对使用；
// 2) 对历史遗留明文数据做兼容（解码失败时保留原值）；
// 3) 空输入返回空切片，避免上层空指针分支。
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

// firstNonEmpty 返回参数列表中第一个非空（trim 后）字符串。
// 用途：在多候选值场景下做兜底选择。
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// normalizeUserID 标准化用户 ID。
// 说明：当输入为空时回退为 demo-user，避免空 user_id 写入数据库。
func normalizeUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return "demo-user"
	}
	return trimmed
}

// normalizeRiskLevel 标准化风险等级为固定三值：高/中/低。
// 说明：非法或空值默认回退为“中”。
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

// normalizeCaseTitle 生成合法案件标题。
// 说明：
// 1) 优先使用显式 title；
// 2) 否则使用 summary 截断生成；
// 3) 两者都为空时返回默认标题。
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

// normalizeTaskTitle 生成任务标题。
// 说明：
// 1) 优先使用文本输入前 48 个字符；
// 2) 无文本时根据多模态输入数量生成占位标题。
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

// newID 生成业务 ID（格式：PREFIX-XXXXXXHEX）。
// 说明：
// 1) 正常路径使用加密随机数；
// 2) 随机失败时回退为时间戳，保证可用性。
func newID(prefix string) string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		now := time.Now().UnixNano()
		return fmt.Sprintf("%s-%d", prefix, now)
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(hex.EncodeToString(bytes)))
}
