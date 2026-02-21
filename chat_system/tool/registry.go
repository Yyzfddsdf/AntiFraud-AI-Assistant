package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type ChatToolHandler interface {
	Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error)
}

type ChatToolResponse struct {
	Payload map[string]interface{}
}

var chatToolRegistry = []openai.Tool{
	ChatQueryUserInfoTool,
	ChatQueryUserCaseHistoryTool,
}

var chatToolBlacklist = map[string]struct{}{}

var chatToolHandlers = map[string]ChatToolHandler{
	ChatQueryUserInfoToolName:        &ChatQueryUserInfoHandler{},
	ChatQueryUserCaseHistoryToolName: &ChatQueryUserCaseHistoryHandler{},
}

func ChatTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(chatToolRegistry))
	for _, registeredTool := range chatToolRegistry {
		if registeredTool.Function != nil {
			if _, blocked := chatToolBlacklist[registeredTool.Function.Name]; blocked {
				continue
			}
		}
		tools = append(tools, registeredTool)
	}
	return tools
}

func GetChatToolHandler(name string) ChatToolHandler {
	return chatToolHandlers[name]
}
