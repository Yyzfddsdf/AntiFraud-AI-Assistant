package httpapi

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"antifraud/database"
	loginmodel "antifraud/login_system/models"
	apimodel "antifraud/multi_agent/httpapi/models"
	"antifraud/multi_agent/queue"
	"antifraud/multi_agent/state"

	"github.com/gin-gonic/gin"
)

// AnalyzeMultimodalScamHandle 处理多模态诈骗智能助手分析请求。
func AnalyzeMultimodalScamHandle(c *gin.Context) {
	var payload apimodel.MultimodalScamAnalyzeRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	hasText := strings.TrimSpace(payload.Text) != ""
	hasVideos := len(payload.Videos) > 0
	hasAudios := len(payload.Audios) > 0
	hasImages := len(payload.Images) > 0
	if !hasText && !hasVideos && !hasAudios && !hasImages {
		c.JSON(http.StatusBadRequest, gin.H{"error": "至少提供 text/videos/audios/images 其中一种输入"})
		return
	}

	userID := getCurrentUserID(c)
	task, err := queue.EnqueueMultimodalTask(userID, queue.EnqueueRequest{
		Text:   payload.Text,
		Videos: payload.Videos,
		Audios: payload.Audios,
		Images: payload.Images,
	})
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "任务入队失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, apimodel.MultimodalScamEnqueueResponse{
		TaskID:  task.TaskID,
		Status:  task.Status,
		Message: "任务已入队，后台处理中，请通过查询接口获取状态与结果",
	})
}

// GetMultimodalTaskStateHandle 查询当前用户任务简要列表（仅标题与状态）。
func GetMultimodalTaskStateHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	view := queue.GetUserTaskState(userID)
	tasks := make([]apimodel.MultimodalTaskListItem, 0, len(view.Pending))

	for _, task := range view.Pending {
		tasks = append(tasks, toTaskListItem(task))
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt > tasks[j].UpdatedAt
	})

	c.JSON(http.StatusOK, apimodel.MultimodalTaskStateResponse{
		UserID: view.UserID,
		Tasks:  tasks,
	})
}

// GetMultimodalHistoryHandle 查询当前用户历史案件明细（仅元数据）。
func GetMultimodalHistoryHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	view := queue.GetUserTaskState(userID)

	history := make([]apimodel.MultimodalHistoryItem, 0, len(view.History))
	for _, item := range view.History {
		history = append(history, apimodel.MultimodalHistoryItem{
			RecordID:    item.RecordID,
			Title:       item.Title,
			CaseSummary: item.CaseSummary,
			ScamType:    item.ScamType,
			RiskLevel:   item.RiskLevel,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, apimodel.MultimodalHistoryResponse{
		UserID:  view.UserID,
		History: history,
	})
}

// DeleteMultimodalHistoryHandle 删除当前用户指定历史案件。
func DeleteMultimodalHistoryHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	recordID := strings.TrimSpace(c.Param("recordId"))
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recordId 不能为空"})
		return
	}

	deleted, err := state.DeleteCaseHistory(userID, recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除历史案件失败: " + err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "历史案件不存在"})
		return
	}

	c.JSON(http.StatusOK, apimodel.DeleteMultimodalHistoryResponse{
		UserID:   userID,
		RecordID: recordID,
		Message:  "历史案件删除成功",
	})
}

// GetMultimodalTaskDetailHandle 查询当前用户指定任务详情。
func GetMultimodalTaskDetailHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	taskID := strings.TrimSpace(c.Param("taskId"))
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "taskId 不能为空"})
		return
	}

	task, exists := state.GetTaskDetailByID(userID, taskID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
		return
	}

	c.JSON(http.StatusOK, apimodel.MultimodalTaskDetailResponse{Task: toTaskItem(task)})
}

// UpdateUserAgeHandle 更新当前登录用户年龄（写入 user 基础数据 DB）。
func UpdateUserAgeHandle(c *gin.Context) {
	var payload apimodel.UpdateUserAgeRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if payload.Age < 1 || payload.Age > 150 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "age 取值范围应为 1-150"})
		return
	}

	userID := getCurrentUserID(c)
	if err := database.DB.Model(&loginmodel.User{}).Where("id = ?", userID).Update("age", payload.Age).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "年龄写入失败"})
		return
	}

	c.JSON(http.StatusOK, apimodel.UpdateUserAgeResponse{
		UserID:  userID,
		Age:     payload.Age,
		Message: "年龄更新成功",
	})
}

// toTaskItem 将内部任务结构转换为 API 任务详情结构。
func toTaskItem(task state.TaskRecord) apimodel.MultimodalTaskItem {
	return apimodel.MultimodalTaskItem{
		TaskID:    task.TaskID,
		UserID:    task.UserID,
		Title:     task.Title,
		Status:    task.Status,
		ScamType:  task.ScamType,
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
		Payload: apimodel.MultimodalTaskPayload{
			Text:          task.Payload.Text,
			Videos:        append([]string{}, task.Payload.Videos...),
			Audios:        append([]string{}, task.Payload.Audios...),
			Images:        append([]string{}, task.Payload.Images...),
			VideoInsights: append([]string{}, task.Payload.VideoInsights...),
			AudioInsights: append([]string{}, task.Payload.AudioInsights...),
			ImageInsights: append([]string{}, task.Payload.ImageInsights...),
		},
		Summary:    strings.TrimSpace(task.Summary),
		Report:     task.Report,
		Error:      task.Error,
		HistoryRef: task.HistoryRef,
	}
}

// toTaskListItem 将内部任务结构转换为任务列表项结构。
func toTaskListItem(task state.TaskRecord) apimodel.MultimodalTaskListItem {
	return apimodel.MultimodalTaskListItem{
		TaskID:    task.TaskID,
		UserID:    task.UserID,
		Title:     task.Title,
		Status:    task.Status,
		Summary:   strings.TrimSpace(task.Summary),
		CreatedAt: task.CreatedAt.Format(time.RFC3339),
		UpdatedAt: task.UpdatedAt.Format(time.RFC3339),
	}
}

// getCurrentUserID 从鉴权上下文读取用户 ID，缺省回退到 demo-user。
func getCurrentUserID(c *gin.Context) string {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return "demo-user"
	}
	if value, ok := userIDValue.(uint); ok {
		return fmt.Sprintf("%d", value)
	}
	return fmt.Sprintf("%v", userIDValue)
}
