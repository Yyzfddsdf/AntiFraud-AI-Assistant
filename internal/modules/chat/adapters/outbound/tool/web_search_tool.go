package tool

import (
	"context"
	"fmt"
	"strings"

	appcfg "antifraud/internal/platform/config"
	openai "antifraud/internal/platform/llm"
	websearch_system "antifraud/internal/platform/websearch"
)

const ChatWebSearchToolName = "search_web"

const (
	defaultWebSearchResults = 3
	minWebSearchResults     = 1
	maxWebSearchResults     = 5
)

type ChatWebSearchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

var ChatWebSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatWebSearchToolName,
		Description: "使用 Tavily 执行联网搜索，返回摘要答案和相关网页结果。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "需要联网检索的问题、事件或关键词。",
				},
				"max_results": map[string]interface{}{
					"type":        "integer",
					"description": "可选，返回结果数量，取值范围 1-5。未传时默认 3。",
				},
			},
			"required": []string{"query"},
		},
	},
}

var loadChatWebSearchConfig = appcfg.LoadConfig
var newChatWebSearcher = func(cfg appcfg.TavilyConfig) websearch_system.Searcher {
	return websearch_system.NewTavilyClient(cfg)
}

func ParseChatWebSearchInput(arguments string) (ChatWebSearchInput, error) {
	return ParseArgs[ChatWebSearchInput](arguments)
}

func ExecuteChatWebSearch(ctx context.Context, searcher websearch_system.Searcher, input ChatWebSearchInput) (map[string]interface{}, error) {
	if searcher == nil {
		return nil, fmt.Errorf("web searcher is nil")
	}

	trimmedQuery := strings.TrimSpace(input.Query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("query is empty")
	}

	resolvedMaxResults := input.MaxResults
	if resolvedMaxResults == 0 {
		resolvedMaxResults = defaultWebSearchResults
	}
	if resolvedMaxResults < minWebSearchResults || resolvedMaxResults > maxWebSearchResults {
		return nil, fmt.Errorf("max_results must be between %d and %d", minWebSearchResults, maxWebSearchResults)
	}

	result, err := searcher.Search(ctx, trimmedQuery, resolvedMaxResults)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"query":       result.Query,
		"answer":      result.Answer,
		"max_results": resolvedMaxResults,
		"results":     result.Results,
		"status":      "success",
	}, nil
}

type ChatWebSearchHandler struct{}

func (h *ChatWebSearchHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	_ = userID

	input, err := ParseChatWebSearchInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid web search args: %v", err),
		}}, nil
	}

	cfg, err := loadChatWebSearchConfig("internal/platform/config/config.json")
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"query":  strings.TrimSpace(input.Query),
			"status": "failed",
			"error":  fmt.Sprintf("load config failed: %v", err),
		}}, nil
	}

	payload, err := ExecuteChatWebSearch(ctx, newChatWebSearcher(cfg.Tavily), input)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{
			"query":  strings.TrimSpace(input.Query),
			"status": "failed",
			"error":  err.Error(),
		}}, nil
	}

	return ChatToolResponse{Payload: payload}, nil
}
