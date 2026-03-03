package queue

import (
	"strings"

	"antifraud/multi_agent"
	"antifraud/multi_agent/state"
)

type EnqueueRequest struct {
	Text   string
	Videos []string
	Audios []string
	Images []string
}

// EnqueueMultimodalTask 创建任务并异步触发处理。
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

// GetUserTaskState 返回用户任务视图（进行中 + 历史）。
func GetUserTaskState(userID string) state.UserStateView {
	return state.GetUserStateView(userID)
}

// processTask 在后台执行主智能体分析流程并写回任务状态。
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
