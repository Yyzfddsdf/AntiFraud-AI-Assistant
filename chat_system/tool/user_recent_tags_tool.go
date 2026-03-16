package tool

import (
	"context"
	"fmt"

	agenttool "antifraud/multi_agent/tool"
	"antifraud/user_profile_system"

	openai "antifraud/llm"
)

const ChatUpdateUserRecentTagsToolName = agenttool.UpdateUserRecentTagsToolName

type ChatUpdateUserRecentTagsInput = agenttool.UpdateUserRecentTagsInput

var ChatUpdateUserRecentTagsTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatUpdateUserRecentTagsToolName,
		Description: "更新当前登录用户的近期标签。标签可为词语或句子，用于描述近期状态。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"recent_tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "近期标签数组，会整体覆盖当前记录。",
				},
			},
			"required": []string{"recent_tags"},
		},
	},
}

type ChatUpdateUserRecentTagsHandler struct{}

func (h *ChatUpdateUserRecentTagsHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	input, err := agenttool.ParseUpdateUserRecentTagsInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid update user recent tags args: %v", err),
		}}, nil
	}

	recentTags, updateErr := user_profile_system.UpdateRecentTagsByStringUserID(userID, input.RecentTags)
	if updateErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  updateErr.Error(),
		}}, nil
	}

	return ChatToolResponse{Payload: map[string]interface{}{
		"status":      "success",
		"user_id":     userID,
		"recent_tags": recentTags,
		"message":     "user recent tags updated",
	}}, nil
}
