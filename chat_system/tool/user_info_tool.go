package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"antifraud/database"
	"antifraud/login_system/models"
	"antifraud/multi_agent/overview"
	"antifraud/multi_agent/state"

	openai "antifraud/llm"
)

const ChatQueryUserInfoToolName = "chat_query_user_info"

type ChatQueryUserInfoInput struct {
	Interval string `json:"interval,omitempty"`
}

var ChatQueryUserInfoTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatQueryUserInfoToolName,
		Description: "查询当前登录用户的画像信息与风险摘要。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"interval": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"day", "week", "month"},
					"description": "可选，风险趋势分析的时间粒度。允许：day/week/month，默认 day。",
				},
			},
		},
	},
}

func ParseChatQueryUserInfoInput(arguments string) (ChatQueryUserInfoInput, error) {
	if strings.TrimSpace(arguments) == "" {
		return ChatQueryUserInfoInput{}, nil
	}
	var input ChatQueryUserInfoInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func QueryUserInfo(userID string, interval string) (map[string]interface{}, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "demo-user"
	}

	view := state.GetUserStateView(uid)
	var age *int
	if numericID, err := strconv.ParseUint(uid, 10, 64); err == nil {
		var user models.User
		if queryErr := database.DB.Where("id = ?", uint(numericID)).First(&user).Error; queryErr == nil {
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
		itemRisk := normalizeRiskLevel(strings.TrimSpace(item.RiskLevel))

		if _, ok := riskCaseCount[itemRisk]; ok {
			riskCaseCount[itemRisk]++
		}

		if itemRisk == "\u9ad8" {
			risk = "\u9ad8"
		}
		if risk != "\u9ad8" && itemRisk == "\u4e2d" {
			risk = "\u4e2d"
		}
	}

	riskOverview := overview.BuildUserRiskOverview(uid, strings.TrimSpace(interval))

	return map[string]interface{}{
		"user_id":              view.UserID,
		"user_name":            fmt.Sprintf("user-%s", view.UserID),
		"age":                  age,
		"account_status":       "active",
		"pending_task_count":   len(view.Pending),
		"completed_case_count": len(view.History),
		"historical_risk":      risk,
		"risk_case_count":      riskCaseCount,
		"risk_trend_analysis": map[string]interface{}{
			"interval":        riskOverview.Interval,
			"current_bucket":  riskOverview.Analysis.CurrentBucket,
			"previous_bucket": riskOverview.Analysis.PreviousBucket,
			"overall_trend":   riskOverview.Analysis.OverallTrend,
			"high_risk_trend": riskOverview.Analysis.HighRiskTrend,
			"summary":         riskOverview.Analysis.Summary,
		},
	}, nil
}

func normalizeRiskLevel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "\u9ad8":
		return "\u9ad8"
	case "\u4f4e":
		return "\u4f4e"
	default:
		return "\u4e2d"
	}
}

type ChatQueryUserInfoHandler struct{}

func (h *ChatQueryUserInfoHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	input, err := ParseChatQueryUserInfoInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user info args: %v", err)}}, nil
	}

	userInfo, queryErr := QueryUserInfo(userID, input.Interval)
	if queryErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": queryErr.Error()}}, nil
	}

	return ChatToolResponse{Payload: map[string]interface{}{"user": userInfo}}, nil
}
