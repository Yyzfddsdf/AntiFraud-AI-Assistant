package scam_simulation

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/tool"
	multi_agent "antifraud/internal/modules/multi_agent/core"
	"antifraud/internal/platform/database"

	"gorm.io/gorm"
)

const (
	sessionStatusInProgress = "in_progress"
	sessionStatusCompleted  = "completed"
	taskStatusPending       = "pending"
	taskStatusProcessing    = "processing"
	taskStatusCompleted     = "completed"
	taskStatusFailed        = "failed"
	taskGenerationTimeout   = 10 * time.Minute
)

var (
	ErrSimulationServiceUnavailable = errors.New("模拟题服务不可用")
	ErrPackAlreadyAttempted         = errors.New("该题包已作答，不允许重复作答")
	ErrUnfinishedSessionExists      = errors.New("存在未完成题目，请先继续完成后再创建新题目")
	ErrGeneratingTaskExists         = errors.New("题目正在生成中，请稍后查看")
)

type PackEntity struct {
	PackID        string    `gorm:"column:pack_id;primaryKey;size:64"`
	UserID        string    `gorm:"column:user_id;index;size:64;not null"`
	Title         string    `gorm:"column:title;size:255;not null"`
	CaseType      string    `gorm:"column:case_type;size:128;not null"`
	TargetPersona string    `gorm:"column:target_persona;size:128;not null"`
	Difficulty    string    `gorm:"column:difficulty;size:32;not null"`
	Intro         string    `gorm:"column:intro;type:text;not null"`
	StepsJSON     string    `gorm:"column:steps_json;type:text;not null"`
	GeneratedByAI bool      `gorm:"column:generated_by_ai;not null"`
	SourceModel   string    `gorm:"column:source_model;size:128;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (PackEntity) TableName() string { return "simulation_quiz_packs" }

type SessionEntity struct {
	PackID           string     `gorm:"column:pack_id;index;size:64;not null"`
	UserID           string     `gorm:"column:user_id;index;size:64;not null"`
	CurrentStep      int        `gorm:"column:current_step;not null"`
	Score            int        `gorm:"column:score;not null"`
	Status           string     `gorm:"column:status;size:32;not null"`
	PackSnapshotJSON string     `gorm:"column:pack_snapshot_json;type:text;not null"`
	AnswersJSON      string     `gorm:"column:answers_json;type:text;not null"`
	ResultJSON       string     `gorm:"column:result_json;type:text;not null"`
	CompletedAt      *time.Time `gorm:"column:completed_at"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (SessionEntity) TableName() string { return "simulation_quiz_sessions" }

func (s *SessionEntity) BeforeCreate(tx *gorm.DB) error {
	if s == nil {
		return nil
	}
	if strings.TrimSpace(s.PackID) == "" {
		return fmt.Errorf("pack_id 不能为空")
	}
	return nil
}

type SessionAnswer struct {
	StepID     string `json:"step_id"`
	StepType   string `json:"step_type"`
	OptionKey  string `json:"option_key"`
	OptionText string `json:"option_text"`
	RiskTag    string `json:"risk_tag"`
	ScoreDelta int    `json:"score_delta"`
	Rationale  string `json:"rationale"`
	AnsweredAt string `json:"answered_at"`
}

type SessionResult struct {
	TotalScore int      `json:"total_score"`
	Level      string   `json:"level"`
	Weaknesses []string `json:"weaknesses"`
	Strengths  []string `json:"strengths"`
	Advice     []string `json:"advice"`
}

type GeneratePackInput struct {
	CaseType      string `json:"case_type"`
	TargetPersona string `json:"target_persona"`
	Difficulty    string `json:"difficulty"`
	Locale        string `json:"locale"`
}

type GenerationTaskEntity struct {
	TaskID       string     `gorm:"column:task_id;primaryKey;size:64"`
	UserID       string     `gorm:"column:user_id;index;size:64;not null"`
	Status       string     `gorm:"column:status;size:32;not null"`
	PackID       string     `gorm:"column:pack_id;index;size:64"`
	ErrorMessage string     `gorm:"column:error_message;type:text"`
	InputJSON    string     `gorm:"column:input_json;type:text;not null"`
	CompletedAt  *time.Time `gorm:"column:completed_at"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (GenerationTaskEntity) TableName() string { return "simulation_quiz_generation_tasks" }

type PackSummary struct {
	PackID       string                        `json:"pack_id"`
	Title        string                        `json:"title"`
	CaseType     string                        `json:"case_type"`
	Difficulty   string                        `json:"difficulty"`
	Intro        string                        `json:"intro"`
	Steps        []tool.SimulationQuizPackStep `json:"steps,omitempty"`
	CreatedAt    time.Time                     `json:"created_at"`
	HasSession   bool                          `json:"has_session"`
	SessionScore int                           `json:"session_score,omitempty"`
	SessionLevel string                        `json:"session_level,omitempty"`
}

type SessionSummary struct {
	PackID      string                         `json:"pack_id"`
	Title       string                         `json:"title"`
	CaseType    string                         `json:"case_type"`
	Difficulty  string                         `json:"difficulty"`
	CurrentStep int                            `json:"current_step"`
	Score       int                            `json:"score"`
	Level       string                         `json:"level"`
	Status      string                         `json:"status"`
	Pack        tool.SimulationQuizPackPayload `json:"pack"`
	Answers     []SessionAnswer                `json:"answers,omitempty"`
	Result      SessionResult                  `json:"result"`
	CompletedAt *time.Time                     `json:"completed_at"`
	CreatedAt   time.Time                      `json:"created_at"`
}

type Service struct {
	db *gorm.DB
}

func init() {
	database.RegisterMainDBSchemaInitializer("simulation_quiz", initSimulationSchema)
}

func initSimulationSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("simulation schema db is nil")
	}
	return db.AutoMigrate(&PackEntity{}, &SessionEntity{}, &GenerationTaskEntity{})
}

func NewService() *Service {
	return &Service{db: database.DB}
}

func (s *Service) GeneratePack(userID string, input GeneratePackInput) (PackEntity, tool.SimulationQuizPackPayload, error) {
	return s.generateAndStorePack(userID, input)
}

func (s *Service) generateAndStorePack(userID string, input GeneratePackInput) (PackEntity, tool.SimulationQuizPackPayload, error) {
	if s == nil || s.db == nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}

	quizPack, err := multi_agent.GenerateSimulationQuizPack(multi_agent.SimulationQuizGenerationInput{
		CaseType:      strings.TrimSpace(input.CaseType),
		TargetPersona: strings.TrimSpace(input.TargetPersona),
		Difficulty:    strings.TrimSpace(input.Difficulty),
		Locale:        strings.TrimSpace(input.Locale),
		QuestionCount: 10,
	})
	if err != nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, err
	}

	stepsJSON, _ := json.Marshal(quizPack.Steps)
	entity := PackEntity{
		PackID:        newSimulationID("PACK"),
		UserID:        normalizedUserID,
		Title:         strings.TrimSpace(quizPack.Title),
		CaseType:      strings.TrimSpace(quizPack.CaseType),
		TargetPersona: strings.TrimSpace(quizPack.TargetPersona),
		Difficulty:    strings.TrimSpace(quizPack.Difficulty),
		Intro:         strings.TrimSpace(quizPack.Intro),
		StepsJSON:     string(stepsJSON),
		GeneratedByAI: true,
		SourceModel:   "deepseek",
	}
	if err := s.db.Create(&entity).Error; err != nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, err
	}
	return entity, quizPack, nil
}

func (s *Service) StartGeneratePackTask(userID string, input GeneratePackInput) (GenerationTaskEntity, error) {
	if s == nil || s.db == nil {
		return GenerationTaskEntity{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}

	// 每次发起新建前先清理当前用户历史失败任务，避免失败记录持续堆积。
	if err := s.db.Where("user_id = ? AND status = ?", normalizedUserID, taskStatusFailed).Delete(&GenerationTaskEntity{}).Error; err != nil {
		return GenerationTaskEntity{}, err
	}

	var unfinishedCount int64
	if err := s.db.Model(&SessionEntity{}).Where("user_id = ? AND status = ?", normalizedUserID, sessionStatusInProgress).Count(&unfinishedCount).Error; err != nil {
		return GenerationTaskEntity{}, err
	}
	if unfinishedCount > 0 {
		return GenerationTaskEntity{}, ErrUnfinishedSessionExists
	}

	var generatingTaskCount int64
	if err := s.db.Model(&GenerationTaskEntity{}).
		Where("user_id = ?", normalizedUserID).
		Where("status IN ?", []string{taskStatusPending, taskStatusProcessing}).
		Count(&generatingTaskCount).Error; err != nil {
		return GenerationTaskEntity{}, err
	}
	if generatingTaskCount > 0 {
		staleBefore := time.Now().Add(-taskGenerationTimeout)
		_ = s.db.Model(&GenerationTaskEntity{}).
			Where("user_id = ?", normalizedUserID).
			Where("status IN ?", []string{taskStatusPending, taskStatusProcessing}).
			Where("created_at < ?", staleBefore).
			Updates(map[string]interface{}{
				"status":        taskStatusFailed,
				"error_message": "题目生成超时，任务已自动终止",
				"completed_at":  time.Now(),
				"updated_at":    time.Now(),
			})

		generatingTaskCount = 0
		if err := s.db.Model(&GenerationTaskEntity{}).
			Where("user_id = ?", normalizedUserID).
			Where("status IN ?", []string{taskStatusPending, taskStatusProcessing}).
			Count(&generatingTaskCount).Error; err != nil {
			return GenerationTaskEntity{}, err
		}
	}
	if generatingTaskCount > 0 {
		return GenerationTaskEntity{}, ErrGeneratingTaskExists
	}

	var unresolvedPackCount int64
	completedSessionSubquery := s.db.
		Table("simulation_quiz_sessions AS s2").
		Select("1").
		Where("s2.pack_id = p.pack_id").
		Where("s2.user_id = p.user_id").
		Where("s2.status = ?", sessionStatusCompleted)
	if err := s.db.
		Table("simulation_quiz_packs AS p").
		Where("p.user_id = ?", normalizedUserID).
		Where("NOT EXISTS (?)", completedSessionSubquery).
		Count(&unresolvedPackCount).Error; err != nil {
		return GenerationTaskEntity{}, err
	}
	if unresolvedPackCount > 0 {
		return GenerationTaskEntity{}, ErrUnfinishedSessionExists
	}

	payload, _ := json.Marshal(input)
	task := GenerationTaskEntity{
		TaskID:    newSimulationID("SIMGEN"),
		UserID:    normalizedUserID,
		Status:    taskStatusPending,
		InputJSON: string(payload),
	}
	if err := s.db.Create(&task).Error; err != nil {
		return GenerationTaskEntity{}, err
	}
	return task, nil
}

func (s *Service) ProcessGeneratePackTask(taskID string) error {
	if s == nil || s.db == nil {
		return ErrSimulationServiceUnavailable
	}
	var task GenerationTaskEntity
	if err := s.db.Where("task_id = ?", strings.TrimSpace(taskID)).Take(&task).Error; err != nil {
		return err
	}
	if task.Status == taskStatusCompleted || task.Status == taskStatusFailed {
		return nil
	}

	if err := s.db.Model(&GenerationTaskEntity{}).Where("task_id = ?", task.TaskID).Updates(map[string]interface{}{
		"status":     taskStatusProcessing,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return err
	}

	input := GeneratePackInput{}
	_ = json.Unmarshal([]byte(task.InputJSON), &input)
	packEntity, _, err := s.generateAndStorePack(task.UserID, input)
	now := time.Now()
	if err != nil {
		return s.db.Model(&GenerationTaskEntity{}).Where("task_id = ?", task.TaskID).Updates(map[string]interface{}{
			"status":        taskStatusFailed,
			"error_message": err.Error(),
			"completed_at":  &now,
			"updated_at":    now,
		}).Error
	}

	if err := s.db.Model(&GenerationTaskEntity{}).Where("task_id = ?", task.TaskID).Updates(map[string]interface{}{
		"status":        taskStatusCompleted,
		"pack_id":       packEntity.PackID,
		"error_message": "",
		"completed_at":  &now,
		"updated_at":    now,
	}).Error; err != nil {
		return err
	}

	_ = s.db.Where("user_id = ? AND task_id <> ?", task.UserID, task.TaskID).Delete(&GenerationTaskEntity{}).Error
	return nil
}

func (s *Service) GetGeneratePackTask(userID string, taskID string) (GenerationTaskEntity, error) {
	if s == nil || s.db == nil {
		return GenerationTaskEntity{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	var task GenerationTaskEntity
	err := s.db.Where("task_id = ? AND user_id = ?", strings.TrimSpace(taskID), normalizedUserID).Take(&task).Error
	if err != nil {
		return GenerationTaskEntity{}, err
	}
	return task, nil
}

func (s *Service) GetLatestPendingTask(userID string) (GenerationTaskEntity, error) {
	if s == nil || s.db == nil {
		return GenerationTaskEntity{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	var task GenerationTaskEntity
	err := s.db.
		Where("user_id = ?", normalizedUserID).
		Where("status IN ?", []string{taskStatusPending, taskStatusProcessing}).
		Order("created_at ASC").
		Take(&task).Error
	if err != nil {
		return GenerationTaskEntity{}, err
	}
	return task, nil
}

func (s *Service) GetPack(userID string, packID string) (PackEntity, tool.SimulationQuizPackPayload, error) {
	if s == nil || s.db == nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, ErrSimulationServiceUnavailable
	}
	var entity PackEntity
	if err := s.db.Where("pack_id = ? AND user_id = ?", strings.TrimSpace(packID), strings.TrimSpace(userID)).First(&entity).Error; err != nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, err
	}
	steps := make([]tool.SimulationQuizPackStep, 0)
	if err := json.Unmarshal([]byte(entity.StepsJSON), &steps); err != nil {
		return PackEntity{}, tool.SimulationQuizPackPayload{}, err
	}
	return entity, tool.SimulationQuizPackPayload{
		Title:         entity.Title,
		CaseType:      entity.CaseType,
		TargetPersona: entity.TargetPersona,
		Difficulty:    entity.Difficulty,
		Intro:         entity.Intro,
		Steps:         steps,
	}, nil
}

func (s *Service) StartSession(userID string, packID string) (SessionEntity, tool.SimulationQuizPackPayload, error) {
	if s == nil || s.db == nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, ErrSimulationServiceUnavailable
	}
	packEntity, pack, err := s.GetPack(userID, packID)
	if err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, err
	}

	var unfinished SessionEntity
	err = s.db.Where("user_id = ? AND status = ?", strings.TrimSpace(userID), sessionStatusInProgress).Order("created_at DESC").Take(&unfinished).Error
	if err == nil {
		if unfinished.PackID == packEntity.PackID {
			return unfinished, pack, nil
		}
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, ErrUnfinishedSessionExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, err
	}

	var existing SessionEntity
	if err := s.db.Where("pack_id = ? AND user_id = ?", packEntity.PackID, strings.TrimSpace(userID)).Order("created_at DESC").Take(&existing).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return SessionEntity{}, tool.SimulationQuizPackPayload{}, err
		}
	} else {
		if existing.Status == sessionStatusInProgress {
			return existing, pack, nil
		}
		if existing.Status == sessionStatusCompleted {
			return SessionEntity{}, tool.SimulationQuizPackPayload{}, ErrPackAlreadyAttempted
		}
	}
	answersJSON, _ := json.Marshal([]SessionAnswer{})
	resultJSON, _ := json.Marshal(SessionResult{})
	packSnapshotJSON, _ := json.Marshal(pack)
	session := SessionEntity{
		PackID:           packEntity.PackID,
		UserID:           strings.TrimSpace(userID),
		CurrentStep:      0,
		Score:            60,
		Status:           sessionStatusInProgress,
		PackSnapshotJSON: string(packSnapshotJSON),
		AnswersJSON:      string(answersJSON),
		ResultJSON:       string(resultJSON),
	}
	if err := s.db.Create(&session).Error; err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, err
	}
	return session, pack, nil
}

func (s *Service) GetOngoingSession(userID string) (SessionEntity, tool.SimulationQuizPackPayload, SessionResult, error) {
	if s == nil || s.db == nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	var session SessionEntity
	if err := s.db.Where("user_id = ? AND status = ?", normalizedUserID, sessionStatusInProgress).Order("created_at DESC").Take(&session).Error; err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}
	pack, err := parsePackSnapshotFromSession(session)
	if err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}
	result := SessionResult{}
	_ = json.Unmarshal([]byte(session.ResultJSON), &result)
	return session, pack, result, nil
}

func (s *Service) GetPackStatus(userID string, packID string) (SessionEntity, tool.SimulationQuizPackPayload, SessionResult, error) {
	if s == nil || s.db == nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	packEntity, pack, err := s.GetPack(normalizedUserID, packID)
	if err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}

	var session SessionEntity
	err = s.db.Where("user_id = ? AND pack_id = ?", normalizedUserID, packEntity.PackID).Order("created_at DESC").Take(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result := evaluateSessionResult(60, []SessionAnswer{})
			return SessionEntity{
				PackID:      packEntity.PackID,
				UserID:      normalizedUserID,
				CurrentStep: 0,
				Score:       60,
				Status:      "not_started",
			}, pack, result, nil
		}
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}

	result := SessionResult{}
	_ = json.Unmarshal([]byte(session.ResultJSON), &result)
	return session, pack, result, nil
}

func (s *Service) SubmitAnswerByPack(userID, packID, stepID, optionKey string) (SessionEntity, tool.SimulationQuizPackPayload, SessionResult, error) {
	if s == nil || s.db == nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, ErrSimulationServiceUnavailable
	}
	var session SessionEntity
	if err := s.db.Where("pack_id = ? AND user_id = ?", strings.TrimSpace(packID), strings.TrimSpace(userID)).First(&session).Error; err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}
	if session.Status != sessionStatusInProgress {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, fmt.Errorf("会话已结束")
	}

	_, pack, err := s.GetPack(userID, session.PackID)
	if err != nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
	}
	if session.CurrentStep < 0 || session.CurrentStep >= len(pack.Steps) {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, fmt.Errorf("当前步骤无效")
	}
	step := pack.Steps[session.CurrentStep]
	if strings.TrimSpace(stepID) != "" && strings.TrimSpace(stepID) != step.StepID {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, fmt.Errorf("step_id 不匹配")
	}

	selectedKey := strings.TrimSpace(strings.ToUpper(optionKey))
	if selectedKey == "" {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, fmt.Errorf("option_key 不能为空")
	}
	var selected *tool.SimulationQuizPackOption
	for i := range step.Options {
		if strings.TrimSpace(strings.ToUpper(step.Options[i].Key)) == selectedKey {
			selected = &step.Options[i]
			break
		}
	}
	if selected == nil {
		return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, fmt.Errorf("option_key 不合法")
	}

	answers := make([]SessionAnswer, 0)
	_ = json.Unmarshal([]byte(session.AnswersJSON), &answers)
	answers = append(answers, SessionAnswer{
		StepID:     step.StepID,
		StepType:   step.StepType,
		OptionKey:  selected.Key,
		OptionText: selected.Text,
		RiskTag:    selected.RiskTag,
		ScoreDelta: selected.ScoreDelta,
		Rationale:  selected.Rationale,
		AnsweredAt: time.Now().Format(time.RFC3339),
	})

	session.Score = clampScore(session.Score + selected.ScoreDelta)
	session.CurrentStep++
	if session.CurrentStep >= len(pack.Steps) {
		now := time.Now()
		session.Status = sessionStatusCompleted
		session.CompletedAt = &now
	}
	updatedAnswers, _ := json.Marshal(answers)
	session.AnswersJSON = string(updatedAnswers)

	result := evaluateSessionResult(session.Score, answers)
	resultBytes, _ := json.Marshal(result)
	session.ResultJSON = string(resultBytes)

	updatePayload := map[string]interface{}{
		"current_step": session.CurrentStep,
		"score":        session.Score,
		"status":       session.Status,
		"answers_json": session.AnswersJSON,
		"result_json":  session.ResultJSON,
		"completed_at": session.CompletedAt,
		"updated_at":   time.Now(),
	}
	if session.Status == sessionStatusCompleted {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&SessionEntity{}).Where("pack_id = ? AND user_id = ?", session.PackID, session.UserID).Updates(updatePayload).Error; err != nil {
				return err
			}
			return tx.Where("pack_id = ? AND user_id = ?", session.PackID, session.UserID).Delete(&PackEntity{}).Error
		}); err != nil {
			return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
		}
	} else {
		if err := s.db.Model(&SessionEntity{}).Where("pack_id = ? AND user_id = ?", session.PackID, session.UserID).Updates(updatePayload).Error; err != nil {
			return SessionEntity{}, tool.SimulationQuizPackPayload{}, SessionResult{}, err
		}
	}

	return session, pack, result, nil
}

func (s *Service) ListPacks(userID string, limit int, offset int) ([]PackSummary, error) {
	if s == nil || s.db == nil {
		return nil, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	packs := make([]PackEntity, 0)
	if err := s.db.Where("user_id = ?", normalizedUserID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&packs).Error; err != nil {
		return nil, err
	}

	summaries := make([]PackSummary, 0, len(packs))
	for _, item := range packs {
		steps := make([]tool.SimulationQuizPackStep, 0)
		_ = json.Unmarshal([]byte(item.StepsJSON), &steps)
		summary := PackSummary{
			PackID:     item.PackID,
			Title:      item.Title,
			CaseType:   item.CaseType,
			Difficulty: item.Difficulty,
			Intro:      item.Intro,
			Steps:      steps,
			CreatedAt:  item.CreatedAt,
		}
		var session SessionEntity
		err := s.db.Where("pack_id = ? AND user_id = ?", item.PackID, normalizedUserID).Order("created_at DESC").Take(&session).Error
		if err == nil {
			summary.HasSession = true
			summary.SessionScore = session.Score
			result := SessionResult{}
			_ = json.Unmarshal([]byte(session.ResultJSON), &result)
			summary.SessionLevel = strings.TrimSpace(result.Level)
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
}

func (s *Service) ListSessions(userID string, limit int, offset int) ([]SessionSummary, error) {
	if s == nil || s.db == nil {
		return nil, ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	sessions := make([]SessionEntity, 0)
	if err := s.db.Where("user_id = ?", normalizedUserID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, err
	}

	summaries := make([]SessionSummary, 0, len(sessions))
	for _, item := range sessions {
		packPayload, _ := parsePackSnapshotFromSession(item)
		answers := make([]SessionAnswer, 0)
		_ = json.Unmarshal([]byte(item.AnswersJSON), &answers)
		result := SessionResult{}
		_ = json.Unmarshal([]byte(item.ResultJSON), &result)
		summaries = append(summaries, SessionSummary{
			PackID:      item.PackID,
			Title:       packPayload.Title,
			CaseType:    packPayload.CaseType,
			Difficulty:  packPayload.Difficulty,
			CurrentStep: item.CurrentStep,
			Score:       item.Score,
			Level:       strings.TrimSpace(result.Level),
			Status:      item.Status,
			Pack:        packPayload,
			Answers:     answers,
			Result:      result,
			CompletedAt: item.CompletedAt,
			CreatedAt:   item.CreatedAt,
		})
	}

	return summaries, nil
}

func parsePackSnapshotFromSession(session SessionEntity) (tool.SimulationQuizPackPayload, error) {
	pack := tool.SimulationQuizPackPayload{}
	if strings.TrimSpace(session.PackSnapshotJSON) == "" {
		return pack, fmt.Errorf("会话缺少题包快照")
	}
	if err := json.Unmarshal([]byte(session.PackSnapshotJSON), &pack); err != nil {
		return tool.SimulationQuizPackPayload{}, err
	}
	return pack, nil
}

func (s *Service) DeleteSessionByPack(userID string, packID string) error {
	if s == nil || s.db == nil {
		return ErrSimulationServiceUnavailable
	}
	normalizedUserID := strings.TrimSpace(userID)
	if normalizedUserID == "" {
		normalizedUserID = "demo-user"
	}
	return s.db.Where("pack_id = ? AND user_id = ?", strings.TrimSpace(packID), normalizedUserID).Delete(&SessionEntity{}).Error
}

// delete pack API removed by design

func evaluateSessionResult(score int, answers []SessionAnswer) SessionResult {
	level := "需提升"
	switch {
	case score >= 90:
		level = "防护优秀"
	case score >= 70:
		level = "基础良好"
	case score >= 50:
		level = "风险较高"
	default:
		level = "高危易受骗"
	}

	weakSet := map[string]struct{}{}
	strongSet := map[string]struct{}{}
	for _, item := range answers {
		switch item.RiskTag {
		case "danger":
			weakSet[item.StepType] = struct{}{}
		case "safe":
			strongSet[item.StepType] = struct{}{}
		}
	}

	weaknesses := make([]string, 0, len(weakSet))
	for key := range weakSet {
		weaknesses = append(weaknesses, key)
	}
	strengths := make([]string, 0, len(strongSet))
	for key := range strongSet {
		strengths = append(strengths, key)
	}

	advice := []string{
		"遇到催促转账、索要验证码、引导点击不明链接时，先中止操作再核验。",
		"涉及客服退款、公检法办案、投资拉群等场景，必须通过官方渠道二次确认。",
		"若已发生损失，立即联系银行与支付平台止付，并报警与联系 96110。",
	}

	return SessionResult{
		TotalScore: score,
		Level:      level,
		Weaknesses: weaknesses,
		Strengths:  strengths,
		Advice:     advice,
	}
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func newSimulationID(prefix string) string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(hex.EncodeToString(buf)))
}
