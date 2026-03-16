package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"antifraud/user_profile_system"

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
	info, err := user_profile_system.BuildUserRiskInfo(userID, interval)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"user_name":            info.UserName,
		"age":                  info.Age,
		"occupation":           info.Occupation,
		"recent_tags":          info.RecentTags,
		"total_case_count":     info.TotalCaseCount,
		"historical_risk":      info.HistoricalRisk,
		"high_risk_case_ratio": info.HighRiskCaseRatio,
		"mid_risk_case_ratio":  info.MidRiskCaseRatio,
		"low_risk_case_ratio":  info.LowRiskCaseRatio,
		"risk_trend_analysis":  info.RiskTrendAnalysis,
	}, nil
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
