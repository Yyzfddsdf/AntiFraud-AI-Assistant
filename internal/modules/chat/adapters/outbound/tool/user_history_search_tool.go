package tool

import (
	"context"
	"fmt"
	"strings"

	agenttool "antifraud/internal/modules/multi_agent/adapters/outbound/tool"

	openai "antifraud/internal/platform/llm"
)

const ChatSearchUserHistoryToolName = agenttool.SearchUserHistoryToolName

type ChatSearchUserHistoryInput = agenttool.SearchUserHistoryInput

var ChatSearchUserHistoryTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatSearchUserHistoryToolName,
		Description: "基于语义搜索当前登录用户的历史案件（向量化召回）。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词或案件描述（语义搜索）。",
				},
				"top_k": map[string]interface{}{
					"type":        "integer",
					"description": "返回结果数量，默认 5，最大 20。",
				},
			},
			"required": []string{"query"},
		},
	},
}

type ChatSearchUserHistoryHandler struct{}

func (h *ChatSearchUserHistoryHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	input, err := agenttool.ParseSearchUserHistoryInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid search user history args: %v", err), "cases": []string{}}}, nil
	}

	boundCtx := agenttool.BindUserID(ctx, userID)
	cases, appliedTopK, searchErr := agenttool.SearchUserHistory(boundCtx, input)
	if searchErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"query":           strings.TrimSpace(input.Query),
			"requested_top_k": input.TopK,
			"applied_top_k":   appliedTopK,
			"error":           searchErr.Error(),
			"cases":           []string{},
		}}, nil
	}

	return ChatToolResponse{Payload: map[string]interface{}{
		"query":           strings.TrimSpace(input.Query),
		"requested_top_k": input.TopK,
		"applied_top_k":   appliedTopK,
		"cases":           cases,
	}}, nil
}
