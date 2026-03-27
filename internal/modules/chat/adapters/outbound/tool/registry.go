package tool

import (
	"context"

	openai "antifraud/internal/platform/llm"
)

type ChatToolHandler interface {
	Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error)
}

type ChatToolResponse struct {
	Payload map[string]interface{}
}

var userChatToolRegistry = []openai.Tool{
	ChatQueryUserInfoTool,
	ChatUpdateUserRecentTagsTool,
	ChatQueryUserCaseHistoryTool,
	ChatSearchUserHistoryTool,
	ChatSearchSimilarCasesTool,
}

var adminChatToolRegistry = []openai.Tool{
	AdminQueryRegionTopScamTypesTool,
	AdminQueryRegionCaseRankingTool,
	ChatSearchSimilarCasesTool,
}

var chatToolBlacklist = map[string]struct{}{}

var userChatToolHandlers = map[string]ChatToolHandler{
	ChatQueryUserInfoToolName:        &ChatQueryUserInfoHandler{},
	ChatUpdateUserRecentTagsToolName: &ChatUpdateUserRecentTagsHandler{},
	ChatQueryUserCaseHistoryToolName: &ChatQueryUserCaseHistoryHandler{},
	ChatSearchUserHistoryToolName:    &ChatSearchUserHistoryHandler{},
	ChatSearchSimilarCasesToolName:   &ChatSearchSimilarCasesHandler{},
}

var adminChatToolHandlers = map[string]ChatToolHandler{
	AdminQueryRegionTopScamTypesToolName: &AdminQueryRegionTopScamTypesHandler{},
	AdminQueryRegionCaseRankingToolName:  &AdminQueryRegionCaseRankingHandler{},
	ChatSearchSimilarCasesToolName:       &ChatSearchSimilarCasesHandler{},
}

func ChatTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(userChatToolRegistry))
	for _, registeredTool := range userChatToolRegistry {
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
	return userChatToolHandlers[name]
}

func AdminChatTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(adminChatToolRegistry))
	for _, registeredTool := range adminChatToolRegistry {
		if registeredTool.Function != nil {
			if _, blocked := chatToolBlacklist[registeredTool.Function.Name]; blocked {
				continue
			}
		}
		tools = append(tools, registeredTool)
	}
	return tools
}

func GetAdminChatToolHandler(name string) ChatToolHandler {
	return adminChatToolHandlers[name]
}
