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

var (
	queueOnce sync.Once
	taskQueue chan queuedTask
)

func EnqueueMultimodalTask(userID string, request EnqueueRequest) (state.TaskRecord, error) {
	initQueue()

	payload := state.TaskPayload{
		Text:   strings.TrimSpace(request.Text),
		Videos: append([]string{}, request.Videos...),
		Audios: append([]string{}, request.Audios...),
		Images: append([]string{}, request.Images...),
	}
	task := state.CreateTask(userID, payload)

	select {
	case taskQueue <- queuedTask{UserID: userID, TaskID: task.TaskID}:
		return task, nil
	default:
		state.MarkTaskFailed(userID, task.TaskID, "任务队列已满，请稍后重试")
		return state.TaskRecord{}, fmt.Errorf("task queue is full")
	}
}

func GetUserTaskState(userID string) state.UserStateView {
	return state.GetUserStateView(userID)
}

func initQueue() {
	queueOnce.Do(func() {
		taskQueue = make(chan queuedTask, 256)
		go worker()
	})
}

func worker() {
	for job := range taskQueue {
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
