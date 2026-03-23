package queue

import (
	"strings"

	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/modules/multi_agent/application"
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
	return application.DefaultTaskService().EnqueueTask(userID, payload)
}

// GetUserTaskState 返回用户任务视图（进行中 + 历史）。
func GetUserTaskState(userID string) state.UserStateView {
	return application.DefaultTaskService().GetUserTaskState(userID)
}
