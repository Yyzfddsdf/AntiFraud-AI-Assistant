package queue

import (
	"strings"

	"image_recognition/multi_agent"
	"image_recognition/multi_agent/state"
)

type EnqueueRequest struct {
	Text   string
	Videos []string
	Audios []string
	Images []string
}

func EnqueueMultimodalTask(userID string, request EnqueueRequest) (state.TaskRecord, error) {
	payload := state.TaskPayload{
		Text:   strings.TrimSpace(request.Text),
		Videos: append([]string{}, request.Videos...),
		Audios: append([]string{}, request.Audios...),
		Images: append([]string{}, request.Images...),
	}
	task := state.CreateTask(userID, payload)

	go processTask(userID, task.TaskID)
	return task, nil
}

func GetUserTaskState(userID string) state.UserStateView {
	return state.GetUserStateView(userID)
}

func processTask(userID, taskID string) {
	state.MarkTaskProcessing(userID, taskID)

	task, exists := state.GetTask(userID, taskID)
	if !exists {
		return
	}

	report, err := multi_agent.AnalyzeMainReportForUser(
		userID,
		taskID,
		task.Payload.Text,
		task.Payload.Videos,
		task.Payload.Audios,
		task.Payload.Images,
	)
	if err != nil {
		state.MarkTaskFailed(userID, taskID, err.Error())
		return
	}

	state.MarkTaskCompleted(userID, taskID, report)
}
