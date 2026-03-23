package application

import (
	"context"
	"fmt"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/modules/multi_agent/core"
)

// Analyzer 定义任务处理所需的分析器端口。
type Analyzer interface {
	Analyze(ctx context.Context, userID string, taskID string, text string, videos []string, audios []string, images []string) (string, error)
}

// TaskStore 定义任务状态持久化端口。
type TaskStore interface {
	CreateTask(userID string, payload state.TaskPayload) state.TaskRecord
	GetUserTaskState(userID string) state.UserStateView
	MarkTaskProcessing(userID string, taskID string)
	GetTask(userID string, taskID string) (state.TaskRecord, bool)
	MarkTaskFailed(userID string, taskID string, errMsg string)
	MarkTaskCompleted(userID string, taskID string, report string)
}

// TaskService 编排多模态任务入队和处理。
type TaskService struct {
	store    TaskStore
	analyzer Analyzer
}

func NewTaskService(store TaskStore, analyzer Analyzer) *TaskService {
	if store == nil {
		store = defaultTaskStore{}
	}
	if analyzer == nil {
		analyzer = defaultAnalyzer{}
	}
	return &TaskService{store: store, analyzer: analyzer}
}

func DefaultTaskService() *TaskService {
	return NewTaskService(nil, nil)
}

func (s *TaskService) EnqueueTask(userID string, payload state.TaskPayload) (state.TaskRecord, error) {
	if s == nil || s.store == nil {
		return state.TaskRecord{}, fmt.Errorf("task service is unavailable")
	}
	task := s.store.CreateTask(userID, payload)
	go s.processTask(userID, task.TaskID)
	return task, nil
}

func (s *TaskService) GetUserTaskState(userID string) state.UserStateView {
	if s == nil || s.store == nil {
		return state.UserStateView{}
	}
	return s.store.GetUserTaskState(userID)
}

func (s *TaskService) processTask(userID, taskID string) {
	s.store.MarkTaskProcessing(userID, taskID)

	task, exists := s.store.GetTask(userID, taskID)
	if !exists {
		return
	}

	report, err := s.analyzer.Analyze(context.Background(), userID, taskID, task.Payload.Text, task.Payload.Videos, task.Payload.Audios, task.Payload.Images)
	if err != nil {
		s.store.MarkTaskFailed(userID, taskID, err.Error())
		return
	}
	s.store.MarkTaskCompleted(userID, taskID, report)
}

type defaultTaskStore struct{}

func (defaultTaskStore) CreateTask(userID string, payload state.TaskPayload) state.TaskRecord {
	return state.CreateTask(userID, payload)
}

func (defaultTaskStore) GetUserTaskState(userID string) state.UserStateView {
	return state.GetUserStateView(userID)
}

func (defaultTaskStore) MarkTaskProcessing(userID string, taskID string) {
	state.MarkTaskProcessing(userID, taskID)
}

func (defaultTaskStore) GetTask(userID string, taskID string) (state.TaskRecord, bool) {
	return state.GetTask(userID, taskID)
}

func (defaultTaskStore) MarkTaskFailed(userID string, taskID string, errMsg string) {
	state.MarkTaskFailed(userID, taskID, errMsg)
}

func (defaultTaskStore) MarkTaskCompleted(userID string, taskID string, report string) {
	state.MarkTaskCompleted(userID, taskID, report)
}

type defaultAnalyzer struct{}

func (defaultAnalyzer) Analyze(ctx context.Context, userID string, taskID string, text string, videos []string, audios []string, images []string) (string, error) {
	_ = ctx
	return multi_agent.AnalyzeMainReportForUser(userID, taskID, text, videos, audios, images)
}
