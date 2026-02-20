package state

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusFailed     = "failed"
	stateFilePath        = "DB/multi_agent_state.json"
)

type TaskPayload struct {
	Text          string   `json:"text"`
	Videos        []string `json:"videos"`
	Audios        []string `json:"audios"`
	Images        []string `json:"images"`
	VideoInsights []string `json:"video_insights,omitempty"`
	AudioInsights []string `json:"audio_insights,omitempty"`
	ImageInsights []string `json:"image_insights,omitempty"`
}

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

type UserStateView struct {
	UserID  string                `json:"user_id"`
	Pending map[string]TaskRecord `json:"pending"`
	History []CaseHistoryRecord   `json:"history"`
}

type userState struct {
	pending map[string]*TaskRecord
	history []CaseHistoryRecord
}

type diskUserState struct {
	Pending map[string]TaskRecord `json:"pending"`
	History []CaseHistoryRecord   `json:"history"`
}

type diskState struct {
	UpdatedAt time.Time                `json:"updated_at"`
	Users     map[string]diskUserState `json:"users"`
}

var (
	storeMu  sync.Mutex
	users    = map[string]*userState{}
	loadOnce sync.Once
)

func StateFilePath() string {
	return stateFilePath
}

func CreateTask(userID string, payload TaskPayload) TaskRecord {
	ensureLoaded()
	uid := normalizeUserID(userID)
	now := time.Now()
	task := TaskRecord{
		TaskID:    newID("TASK"),
		UserID:    uid,
		Title:     normalizeTaskTitle(payload),
		Status:    TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Payload:   payload,
	}

	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	state.pending[task.TaskID] = cloneTask(task)
	persistLocked()
	return task
}

func MarkTaskProcessing(userID, taskID string) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	task, exists := state.pending[taskID]
	if !exists {
		return
	}
	task.Status = TaskStatusProcessing
	task.UpdatedAt = time.Now()
	persistLocked()
}

func MarkTaskCompleted(userID, taskID, report string) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	task, exists := state.pending[taskID]
	if !exists {
		return
	}
	task.Status = TaskStatusCompleted
	task.Report = strings.TrimSpace(report)
	task.UpdatedAt = time.Now()
	delete(state.pending, taskID)
	persistLocked()
}

func UpdateTaskInsights(userID, taskID string, videoInsights []string, audioInsights []string, imageInsights []string) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	tid := strings.TrimSpace(taskID)
	if tid == "" {
		return
	}

	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	task, exists := state.pending[tid]
	if !exists {
		return
	}
	task.Payload.VideoInsights = append([]string{}, videoInsights...)
	task.Payload.AudioInsights = append([]string{}, audioInsights...)
	task.Payload.ImageInsights = append([]string{}, imageInsights...)
	task.UpdatedAt = time.Now()
	persistLocked()
}

func MarkTaskFailed(userID, taskID, errMsg string) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	task, exists := state.pending[taskID]
	if !exists {
		return
	}
	task.Status = TaskStatusFailed
	task.Error = strings.TrimSpace(errMsg)
	task.UpdatedAt = time.Now()
	state.history = append([]CaseHistoryRecord{newFailedHistoryRecord(uid, *task)}, state.history...)
	delete(state.pending, taskID)
	persistLocked()
}

func GetTask(userID, taskID string) (TaskRecord, bool) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	if task, exists := state.pending[taskID]; exists {
		return *cloneTaskPtr(task), true
	}
	return TaskRecord{}, false
}

func GetTaskDetailByID(userID, id string) (TaskRecord, bool) {
	ensureLoaded()
	uid := normalizeUserID(userID)
	targetID := strings.TrimSpace(id)
	if targetID == "" {
		return TaskRecord{}, false
	}

	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)

	if pendingTask, exists := state.pending[targetID]; exists {
		return *cloneTaskPtr(pendingTask), true
	}

	for _, history := range state.history {
		if strings.TrimSpace(history.RecordID) != targetID {
			continue
		}

		report := strings.TrimSpace(history.Report)
		if report == "" {
			report = strings.TrimSpace(history.CaseSummary)
		}

		return TaskRecord{
			TaskID:    history.RecordID,
			UserID:    history.UserID,
			Title:     history.Title,
			Status:    TaskStatusCompleted,
			CreatedAt: history.CreatedAt,
			UpdatedAt: history.CreatedAt,
			Payload: TaskPayload{
				Text:          history.Payload.Text,
				Videos:        append([]string{}, history.Payload.Videos...),
				Audios:        append([]string{}, history.Payload.Audios...),
				Images:        append([]string{}, history.Payload.Images...),
				VideoInsights: append([]string{}, history.Payload.VideoInsights...),
				AudioInsights: append([]string{}, history.Payload.AudioInsights...),
				ImageInsights: append([]string{}, history.Payload.ImageInsights...),
			},
			Report: report,
		}, true
	}

	return TaskRecord{}, false
}

func GetUserStateView(userID string) UserStateView {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)

	pending := map[string]TaskRecord{}
	for key, task := range state.pending {
		pending[key] = *cloneTaskPtr(task)
	}

	history := make([]CaseHistoryRecord, len(state.history))
	copy(history, state.history)

	return UserStateView{
		UserID:  uid,
		Pending: pending,
		History: history,
	}
}

func AddCaseHistory(userID, taskID, title, summary, riskLevel string, payload TaskPayload, report string) CaseHistoryRecord {
	ensureLoaded()
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

	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	state.history = append([]CaseHistoryRecord{record}, state.history...)
	persistLocked()
	return record
}

func GetCaseHistory(userID string) []CaseHistoryRecord {
	ensureLoaded()
	uid := normalizeUserID(userID)
	storeMu.Lock()
	defer storeMu.Unlock()
	state := ensureUserState(uid)
	history := make([]CaseHistoryRecord, len(state.history))
	copy(history, state.history)
	return history
}

func ensureLoaded() {
	loadOnce.Do(func() {
		storeMu.Lock()
		defer storeMu.Unlock()
		if err := loadFromDiskLocked(); err != nil {
			log.Printf("[state] load state file failed: %v", err)
		}
	})
}

func loadFromDiskLocked() error {
	users = map[string]*userState{}
	if _, err := os.Stat(stateFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	data, err := os.ReadFile(stateFilePath)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}

	var snapshot diskState
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}

	for userID, userSnapshot := range snapshot.Users {
		state := ensureUserState(userID)
		for taskID, task := range userSnapshot.Pending {
			state.pending[taskID] = cloneTask(task)
		}
		state.history = append([]CaseHistoryRecord{}, userSnapshot.History...)
	}

	return nil
}

func persistLocked() {
	snapshot := diskState{
		UpdatedAt: time.Now(),
		Users:     map[string]diskUserState{},
	}

	for userID, user := range users {
		pending := map[string]TaskRecord{}
		for taskID, task := range user.pending {
			pending[taskID] = *cloneTaskPtr(task)
		}

		history := append([]CaseHistoryRecord{}, user.history...)
		snapshot.Users[userID] = diskUserState{
			Pending: pending,
			History: history,
		}
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		log.Printf("[state] marshal snapshot failed: %v", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(stateFilePath), 0755); err != nil {
		log.Printf("[state] create dir failed: %v", err)
		return
	}

	tmp := stateFilePath + ".tmp"

	if err := os.WriteFile(tmp, data, 0644); err != nil {
		log.Printf("[state] write tmp file failed: %v", err)
		return
	}
	if err := os.Rename(tmp, stateFilePath); err != nil {
		if removeErr := os.Remove(stateFilePath); removeErr == nil || os.IsNotExist(removeErr) {
			if retryErr := os.Rename(tmp, stateFilePath); retryErr == nil {
				return
			} else {
				log.Printf("[state] rename state file retry failed: %v", retryErr)
			}
		} else {
			log.Printf("[state] remove old state file failed: %v", removeErr)
		}
		log.Printf("[state] rename state file failed: %v", err)
	}
}

func newFailedHistoryRecord(userID string, task TaskRecord) CaseHistoryRecord {
	reason := strings.TrimSpace(task.Error)
	if reason == "" {
		reason = "任务执行失败"
	}
	now := task.UpdatedAt
	if now.IsZero() {
		now = time.Now()
	}
	return CaseHistoryRecord{
		RecordID:    task.TaskID,
		UserID:      normalizeUserID(userID),
		Title:       normalizeCaseTitle(task.Title, reason),
		CaseSummary: reason,
		RiskLevel:   "中",
		CreatedAt:   now,
		Payload: TaskPayload{
			Text:          task.Payload.Text,
			Videos:        append([]string{}, task.Payload.Videos...),
			Audios:        append([]string{}, task.Payload.Audios...),
			Images:        append([]string{}, task.Payload.Images...),
			VideoInsights: append([]string{}, task.Payload.VideoInsights...),
			AudioInsights: append([]string{}, task.Payload.AudioInsights...),
			ImageInsights: append([]string{}, task.Payload.ImageInsights...),
		},
		Report: reason,
	}
}

func ensureUserState(userID string) *userState {
	state, exists := users[userID]
	if !exists {
		state = &userState{
			pending: map[string]*TaskRecord{},
			history: []CaseHistoryRecord{},
		}
		users[userID] = state
	}
	return state
}

func normalizeUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return "demo-user"
	}
	return trimmed
}

func normalizeRiskLevel(level string) string {
	trimmed := strings.TrimSpace(level)
	if trimmed == "" {
		return "中"
	}
	return trimmed
}

func normalizeCaseTitle(title, summary string) string {
	trimmed := strings.TrimSpace(title)
	if trimmed != "" {
		return trimmed
	}
	s := strings.TrimSpace(summary)
	if s == "" {
		return "未命名案件"
	}
	runes := []rune(s)
	if len(runes) <= 20 {
		return s
	}
	return string(runes[:20]) + "..."
}

func normalizeTaskTitle(payload TaskPayload) string {
	text := strings.TrimSpace(payload.Text)
	if text != "" {
		runes := []rune(text)
		if len(runes) <= 24 {
			return text
		}
		return string(runes[:24]) + "..."
	}
	return fmt.Sprintf("多模态任务(V%d/A%d/I%d)", len(payload.Videos), len(payload.Audios), len(payload.Images))
}

func newID(prefix string) string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		now := time.Now().UnixNano()
		return fmt.Sprintf("%s-%d", prefix, now)
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(hex.EncodeToString(bytes)))
}

func cloneTask(task TaskRecord) *TaskRecord {
	copyTask := task
	copyTask.Payload.Videos = append([]string{}, task.Payload.Videos...)
	copyTask.Payload.Audios = append([]string{}, task.Payload.Audios...)
	copyTask.Payload.Images = append([]string{}, task.Payload.Images...)
	copyTask.Payload.VideoInsights = append([]string{}, task.Payload.VideoInsights...)
	copyTask.Payload.AudioInsights = append([]string{}, task.Payload.AudioInsights...)
	copyTask.Payload.ImageInsights = append([]string{}, task.Payload.ImageInsights...)
	return &copyTask
}

func cloneTaskPtr(task *TaskRecord) *TaskRecord {
	if task == nil {
		return nil
	}
	return cloneTask(*task)
}
