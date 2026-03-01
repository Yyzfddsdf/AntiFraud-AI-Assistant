package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"image_recognition/login_system/database"
	"image_recognition/login_system/models"
	"image_recognition/multi_agent/state"

	openai "image_recognition/llm"
)

const ChatQueryUserInfoToolName = "chat_query_user_info"

type ChatQueryUserInfoInput struct{}

var ChatQueryUserInfoTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatQueryUserInfoToolName,
		Description: "查询当前登录用户的画像信息与风险摘要。",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
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

func QueryUserInfo(userID string) (map[string]interface{}, error) {
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
		"completed_case_count": len(view.History),
		"historical_risk":      risk,
		"risk_case_count":      riskCaseCount,
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
	_, err := ParseChatQueryUserInfoInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user info args: %v", err)}}, nil
	}

	userInfo, queryErr := QueryUserInfo(userID)
	if queryErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": queryErr.Error()}}, nil
	}

	return ChatToolResponse{Payload: map[string]interface{}{"user": userInfo}}, nil
}
