package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"image_recognition/multi_agent/state"

	"github.com/sashabaranov/go-openai"
)

const ChatQueryUserCaseHistoryToolName = "chat_query_user_case_history"

type ChatQueryUserCaseHistoryInput struct{}

var ChatQueryUserCaseHistoryTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatQueryUserCaseHistoryToolName,
		Description: "查询当前登录用户历史案件摘要列表",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
	},
}

func ParseChatQueryUserCaseHistoryInput(arguments string) (ChatQueryUserCaseHistoryInput, error) {
	if strings.TrimSpace(arguments) == "" {
		return ChatQueryUserCaseHistoryInput{}, nil
	}
	var input ChatQueryUserCaseHistoryInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func QueryUserCaseHistory(userID string) ([]map[string]interface{}, error) {
	uid := strings.TrimSpace(userID)
	if uid == "" {
		uid = "demo-user"
	}

	history := state.GetCaseHistory(uid)
	if len(history) == 0 {
		return []map[string]interface{}{}, nil
	}

	result := make([]map[string]interface{}, 0, len(history))
	for _, item := range history {
		result = append(result, map[string]interface{}{
			"record_id":    item.RecordID,
			"title":        item.Title,
			"case_summary": item.CaseSummary,
			"risk_level":   item.RiskLevel,
			"created_at":   item.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return result, nil
}

type ChatQueryUserCaseHistoryHandler struct{}

func (h *ChatQueryUserCaseHistoryHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	_, err := ParseChatQueryUserCaseHistoryInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user case history args: %v", err)}}, nil
	}

	cases, queryErr := QueryUserCaseHistory(userID)
	if queryErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": queryErr.Error()}}, nil
	}

	return ChatToolResponse{Payload: map[string]interface{}{"cases": cases, "count": len(cases)}}, nil
}
