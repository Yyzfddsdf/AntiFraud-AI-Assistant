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

	"github.com/sashabaranov/go-openai"
)

const ChatQueryUserInfoToolName = "chat_query_user_info"

type ChatQueryUserInfoInput struct{}

var ChatQueryUserInfoTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatQueryUserInfoToolName,
		Description: "查询当前登录用户基础信息与风险画像",
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

	risk := "低"
	riskCaseCount := map[string]int{"低": 0, "中": 0, "高": 0}
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

		if itemRisk == "高" {
			risk = "高"
			break
		}
		if risk != "高" && itemRisk == "中" {
			risk = "中"
		}
	}

	return map[string]interface{}{
		"user_id":              view.UserID,
		"user_name":            fmt.Sprintf("用户%s", view.UserID),
		"age":                  age,
		"account_status":       "active",
		"pending_task_count":   len(view.Pending),
		"completed_case_count": len(view.History),
		"historical_risk":      risk,
		"risk_case_count":      riskCaseCount,
	}, nil
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
