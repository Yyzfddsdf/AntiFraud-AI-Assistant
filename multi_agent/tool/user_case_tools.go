package tool

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"image_recognition/login_system/database"
	"image_recognition/login_system/models"
	"image_recognition/multi_agent/state"

	openai "image_recognition/llm"
)

const QueryUserHistoryCasesToolName = "query_user_history_cases"
const QueryUserInfoToolName = "query_user_info"
const WriteUserHistoryCaseToolName = "write_user_history_case"

type QueryUserHistoryCasesInput struct{}

type QueryUserInfoInput struct{}

type WriteUserHistoryCaseInput struct {
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	RiskLevel   string `json:"risk_level"`
}

var QueryUserHistoryCasesTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        QueryUserHistoryCasesToolName,
		Description: "查询当前绑定用户的历史案件记录。",
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
		Description: "查询当前绑定用户的画像信息与风险摘要。",
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
		Description: "将当前分析案件写入用户历史记录。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "案件标题。",
				},
				"case_summary": map[string]interface{}{
					"type":        "string",
					"description": "案件摘要。",
				},
				"risk_level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"低", "中", "高"},
					"description": "风险等级，仅允许：低/中/高。",
				},
			},
			"required": []string{"title", "case_summary", "risk_level"},
		},
	},
}

func ParseQueryUserHistoryCasesInput(arguments string) (QueryUserHistoryCasesInput, error) {
	return ParseArgs[QueryUserHistoryCasesInput](arguments)
}

func ParseQueryUserInfoInput(arguments string) (QueryUserInfoInput, error) {
	return ParseArgs[QueryUserInfoInput](arguments)
}

func ParseWriteUserHistoryCaseInput(arguments string) (WriteUserHistoryCaseInput, error) {
	return ParseArgs[WriteUserHistoryCaseInput](arguments)
}

func QueryUserHistoryCases(ctx context.Context) ([]string, error) {
	history := state.GetCaseHistory(CurrentUserID(ctx))
	if len(history) == 0 {
		return []string{"No historical case record"}, nil
	}

	results := make([]string, 0, len(history))
	for _, record := range history {
		videoInsights := "none"
		if len(record.Payload.VideoInsights) > 0 {
			videoInsights = strings.Join(record.Payload.VideoInsights, "; ")
		}
		audioInsights := "none"
		if len(record.Payload.AudioInsights) > 0 {
			audioInsights = strings.Join(record.Payload.AudioInsights, "; ")
		}
		imageInsights := "none"
		if len(record.Payload.ImageInsights) > 0 {
			imageInsights = strings.Join(record.Payload.ImageInsights, "; ")
		}

		results = append(results, fmt.Sprintf(
			"%s | title: %s | summary: %s | risk: %s | video: %s | audio: %s | image: %s",
			record.CreatedAt.Format("2006-01-02 15:04:05"),
			record.Title,
			record.CaseSummary,
			record.RiskLevel,
			videoInsights,
			audioInsights,
			imageInsights,
		))
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

	risk := "\u4f4e"
	riskCaseCount := map[string]int{
		"\u4f4e": 0,
		"\u4e2d": 0,
		"\u9ad8": 0,
	}

	for _, item := range view.History {
		itemRisk := normalizeRiskLevelFromHistory(item.RiskLevel)

		if _, ok := riskCaseCount[itemRisk]; ok {
			riskCaseCount[itemRisk]++
		}

		if itemRisk == "\u9ad8" {
			risk = "\u9ad8"
			break
		}
		if risk != "\u9ad8" && itemRisk == "\u4e2d" {
			risk = "\u4e2d"
		}
	}

	return map[string]interface{}{
		"user_id":              view.UserID,
		"user_name":            fmt.Sprintf("user-%s", view.UserID),
		"age":                  age,
		"account_status":       "active",
		"pending_task_count":   len(view.Pending),
		"completed_task_count": len(view.History),
		"recent_case_count":    len(view.History),
		"historical_risk":      risk,
		"risk_case_count":      riskCaseCount,
		"high_risk_case_count": riskCaseCount["\u9ad8"],
		"mid_risk_case_count":  riskCaseCount["\u4e2d"],
		"low_risk_case_count":  riskCaseCount["\u4f4e"],
	}, nil
}

func normalizeRiskLevelFromHistory(raw string) string {
	switch strings.TrimSpace(raw) {
	case "\u9ad8":
		return "\u9ad8"
	case "\u4f4e":
		return "\u4f4e"
	default:
		return "\u4e2d"
	}
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
		"message":      "history case persisted",
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
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user history args: %v", err), "cases": []string{"none"}}}, nil
	}
	cases, queryErr := QueryUserHistoryCases(ctx)
	if queryErr != nil {
		boundUserID := CurrentUserID(ctx)
		return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "error": queryErr.Error(), "cases": []string{"query failed"}}}, nil
	}
	boundUserID := CurrentUserID(ctx)
	return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "cases": cases}}, nil
}

type QueryUserInfoHandler struct{}

func (h *QueryUserInfoHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	_, err := ParseQueryUserInfoInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user info args: %v", err), "user": map[string]interface{}{"user_id": "demo-user", "user_name": "demo-user"}}}, nil
	}
	userInfo, queryErr := QueryUserInfo(ctx)
	if queryErr != nil {
		boundUserID := CurrentUserID(ctx)
		return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "error": queryErr.Error(), "user": map[string]interface{}{"user_id": boundUserID, "user_name": "user-" + boundUserID}}}, nil
	}
	boundUserID := CurrentUserID(ctx)
	return ToolResponse{Payload: map[string]interface{}{"user_id": boundUserID, "user": userInfo}}, nil
}

type WriteUserHistoryCaseHandler struct{}

func (h *WriteUserHistoryCaseHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseWriteUserHistoryCaseInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid write user history case args: %v", err), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "invalid input"}}}, nil
	}
	_, writeErr := WriteUserHistoryCase(ctx, input)
	if writeErr != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": writeErr.Error(), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "persist failed"}}}, nil
	}
	boundUserID := CurrentUserID(ctx)
	return ToolResponse{Payload: map[string]interface{}{
		"status":             "success",
		"user_id":            boundUserID,
		"message":            "user history case persisted",
		"system_instruction": "CRITICAL: Case archiving is the FINAL step. All tasks are completed. You MUST STOP calling any tools now and end the conversation immediately.",
	}}, nil
}
