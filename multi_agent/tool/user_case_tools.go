package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"image_recognition/login_system/database"
	"image_recognition/login_system/models"
	"image_recognition/multi_agent/state"

	"github.com/sashabaranov/go-openai"
)

const QueryUserHistoryCasesToolName = "query_user_history_cases"
const QueryUserInfoToolName = "query_user_info"
const WriteUserHistoryCaseToolName = "write_user_history_case"

type QueryUserHistoryCasesInput struct {
}

type QueryUserInfoInput struct {
}

type WriteUserHistoryCaseInput struct {
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	RiskLevel   string `json:"risk_level"`
}

var QueryUserHistoryCasesTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        QueryUserHistoryCasesToolName,
		Description: "查询当前用户历史案件记录（用户ID由服务端HTTP上下文自动获取）",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
}

var QueryUserInfoTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        QueryUserInfoToolName,
		Description: "查询当前用户基础信息与风险画像（用户ID由服务端HTTP上下文自动获取）",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
}

var WriteUserHistoryCaseTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        WriteUserHistoryCaseToolName,
		Description: "写入当前用户历史案件记录（用户ID由服务端HTTP上下文自动获取）",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "案件标题",
				},
				"case_summary": map[string]interface{}{
					"type":        "string",
					"description": "案件摘要",
				},
				"risk_level": map[string]interface{}{
					"type":        "string",
					"description": "风险等级",
				},
			},
			"required": []string{"title", "case_summary", "risk_level"},
		},
	},
}

func ParseQueryUserHistoryCasesInput(arguments string) (QueryUserHistoryCasesInput, error) {
	var input QueryUserHistoryCasesInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func ParseQueryUserInfoInput(arguments string) (QueryUserInfoInput, error) {
	var input QueryUserInfoInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func ParseWriteUserHistoryCaseInput(arguments string) (WriteUserHistoryCaseInput, error) {
	var input WriteUserHistoryCaseInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func QueryUserHistoryCases(ctx context.Context) ([]string, error) {
	history := state.GetCaseHistory(CurrentUserID(ctx))
	if len(history) == 0 {
		return []string{"暂无历史案件记录"}, nil
	}

	results := make([]string, 0, len(history))
	for _, record := range history {
		videoInsights := "无"
		if len(record.Payload.VideoInsights) > 0 {
			videoInsights = strings.Join(record.Payload.VideoInsights, "；")
		}
		audioInsights := "无"
		if len(record.Payload.AudioInsights) > 0 {
			audioInsights = strings.Join(record.Payload.AudioInsights, "；")
		}
		imageInsights := "无"
		if len(record.Payload.ImageInsights) > 0 {
			imageInsights = strings.Join(record.Payload.ImageInsights, "；")
		}

		results = append(results, fmt.Sprintf("%s | 标题: %s | 摘要: %s | 风险等级: %s | 视频解读: %s | 音频解读: %s | 图像解读: %s", record.CreatedAt.Format("2006-01-02 15:04:05"), record.Title, record.CaseSummary, record.RiskLevel, videoInsights, audioInsights, imageInsights))
	}
	return results, nil
}

func QueryUserInfo(ctx context.Context) (map[string]interface{}, error) {
	uid := CurrentUserID(ctx)
	view := state.GetUserStateView(uid)
	var age *int
	if userID, err := strconv.ParseUint(strings.TrimSpace(uid), 10, 64); err == nil {
		var user models.User
		if queryErr := database.DB.Where("id = ?", uint(userID)).First(&user).Error; queryErr == nil {
			age = user.Age
		}
	}

	risk := "低"
	riskCaseCount := map[string]int{
		"低": 0,
		"中": 0,
		"高": 0,
	}

	for _, item := range view.History {
		itemRisk := strings.TrimSpace(item.RiskLevel)
		if itemRisk == "" {
			switch {
			case strings.Contains(item.Report, "风险等级：高"):
				itemRisk = "高"
			case strings.Contains(item.Report, "风险等级：中"):
				itemRisk = "中"
			default:
				itemRisk = "低"
			}
		}

		if _, ok := riskCaseCount[itemRisk]; ok {
			riskCaseCount[itemRisk]++
		}

		if itemRisk == "高" || strings.Contains(item.Report, "风险等级：高") {
			risk = "高"
			break
		}
		if risk != "高" && (itemRisk == "中" || strings.Contains(item.Report, "风险等级：中")) {
			risk = "中"
		}
	}

	return map[string]interface{}{
		"user_id":              view.UserID,
		"user_name":            fmt.Sprintf("用户%s", view.UserID),
		"age":                  age,
		"account_status":       "active",
		"pending_task_count":   len(view.Pending),
		"completed_task_count": len(view.History),
		"recent_case_count":    len(view.History),
		"historical_risk":      risk,
		"risk_case_count":      riskCaseCount,
		"high_risk_case_count": riskCaseCount["高"],
		"mid_risk_case_count":  riskCaseCount["中"],
		"low_risk_case_count":  riskCaseCount["低"],
	}, nil
}

func WriteUserHistoryCase(ctx context.Context, input WriteUserHistoryCaseInput) (map[string]interface{}, error) {
	payload := CurrentTaskPayload(ctx)
	insights := CurrentTaskInsights(ctx)
	state.AddCaseHistory(CurrentUserID(ctx), CurrentTaskID(ctx), input.Title, input.CaseSummary, input.RiskLevel, state.TaskPayload{
		Text:          payload.Text,
		Videos:        append([]string{}, payload.Videos...),
		Audios:        append([]string{}, payload.Audios...),
		Images:        append([]string{}, payload.Images...),
		VideoInsights: append([]string{}, insights.VideoInsights...),
		AudioInsights: append([]string{}, insights.AudioInsights...),
		ImageInsights: append([]string{}, insights.ImageInsights...),
	}, CurrentFinalReport(ctx))
	return map[string]interface{}{
		"status":       "success",
		"record_id":    "CASE-WRITE-" + CurrentUserID(ctx),
		"user_id":      CurrentUserID(ctx),
		"message":      "历史案件写入成功(内存JSON)",
		"title":        input.Title,
		"created_at":   time.Now().Format(time.RFC3339),
		"case_summary": input.CaseSummary,
		"report":       CurrentFinalReport(ctx),
		"stored_level": strings.TrimSpace(input.RiskLevel),
	}, nil
}

type QueryUserHistoryCasesHandler struct{}

func (h *QueryUserHistoryCasesHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	_, err := ParseQueryUserHistoryCasesInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user history args: %v", err), "cases": []string{"无"}}}, nil
	}
	cases, queryErr := QueryUserHistoryCases(ctx)
	if queryErr != nil {
		boundUserID := CurrentUserID(ctx)
		return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "error": queryErr.Error(), "cases": []string{"查询失败(模拟)"}}}, nil
	}
	boundUserID := CurrentUserID(ctx)
	return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "cases": cases}}, nil
}

type QueryUserInfoHandler struct{}

func (h *QueryUserInfoHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	_, err := ParseQueryUserInfoInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user info args: %v", err), "user": map[string]interface{}{"user_id": "demo-user", "user_name": "张三"}}}, nil
	}
	userInfo, queryErr := QueryUserInfo(ctx)
	if queryErr != nil {
		boundUserID := CurrentUserID(ctx)
		return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "error": queryErr.Error(), "user": map[string]interface{}{"user_id": boundUserID, "user_name": "用户" + boundUserID}}}, nil
	}
	boundUserID := CurrentUserID(ctx)
	return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "user": userInfo}}, nil
}

type WriteUserHistoryCaseHandler struct{}

func (h *WriteUserHistoryCaseHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseWriteUserHistoryCaseInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid write user history case args: %v", err), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "参数错误，已模拟写入"}}}, nil
	}
	writeResult, writeErr := WriteUserHistoryCase(ctx, input)
	if writeErr != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": writeErr.Error(), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "写入失败，返回模拟结果"}}}, nil
	}
	return ToolResponse{Payload: map[string]interface{}{"status": "success", "record": writeResult}, SetHistoryWriteAfterFinal: true}, nil
}
