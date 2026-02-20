package queue

import (
	"fmt"
	"strings"
	"sync"

	"image_recognition/multi_agent"
	"image_recognition/multi_agent/state"
)

type EnqueueRequest struct {
	Text   string
	Videos []string
	Audios []string
	Images []string
}

type queuedTask struct {
	UserID string
	TaskID string
}

const perUserQueueSize = 16

var (
	userQueuesMu sync.Mutex
	userQueues   = map[string]chan queuedTask{}
)

func EnqueueMultimodalTask(userID string, request EnqueueRequest) (state.TaskRecord, error) {
	payload := state.TaskPayload{
		Text:   strings.TrimSpace(request.Text),
		Videos: append([]string{}, request.Videos...),
		Audios: append([]string{}, request.Audios...),
		Images: append([]string{}, request.Images...),
	}
	task := state.CreateTask(userID, payload)

	ch := getOrCreateUserQueue(userID)
	select {
	case ch <- queuedTask{UserID: userID, TaskID: task.TaskID}:
		return task, nil
	default:
		state.MarkTaskFailed(userID, task.TaskID, "任务队列已满，请稍后重试")
		return state.TaskRecord{}, fmt.Errorf("task queue is full")
	}
}

func GetUserTaskState(userID string) state.UserStateView {
	return state.GetUserStateView(userID)
}

func getOrCreateUserQueue(userID string) chan queuedTask {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "demo-user"
	}
	userQueuesMu.Lock()
	defer userQueuesMu.Unlock()
	if ch, exists := userQueues[uid]; exists {
		return ch
	}
	ch := make(chan queuedTask, perUserQueueSize)
	userQueues[uid] = ch
	go userWorker(ch)
	return ch
}

func userWorker(ch chan queuedTask) {
	for job := range ch {
		state.MarkTaskProcessing(job.UserID, job.TaskID)

		task, exists := state.GetTask(job.UserID, job.TaskID)
		if !exists {
			continue
		}

		report, err := multi_agent.AnalyzeMainReportForUser(
			job.UserID,
			job.TaskID,
			task.Payload.Text,
			task.Payload.Videos,
			task.Payload.Audios,
			task.Payload.Images,
		)
		if err != nil {
			state.MarkTaskFailed(job.UserID, job.TaskID, err.Error())
			continue
		}

		state.MarkTaskCompleted(job.UserID, job.TaskID, report)
	}
}
