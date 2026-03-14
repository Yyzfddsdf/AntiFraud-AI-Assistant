package tool

import (
	"context"
	"fmt"
	"strings"

	appcfg "antifraud/config"
	openai "antifraud/llm"
	"antifraud/web_search_system"
)

const WebSearchToolName = "search_web"

const (
	defaultWebSearchResults = 3
	minWebSearchResults     = 1
	maxWebSearchResults     = 5
)

// WebSearchInput 定义联网搜索工具的输入参数。
type WebSearchInput struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
}

var WebSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        WebSearchToolName,
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

var loadWebSearchConfig = appcfg.LoadConfig
var newWebSearcher = func(cfg appcfg.TavilyConfig) web_search_system.Searcher {
	return web_search_system.NewTavilyClient(cfg)
}

func ParseWebSearchInput(arguments string) (WebSearchInput, error) {
	return ParseArgs[WebSearchInput](arguments)
}

// ExecuteWebSearch 执行联网搜索并构造统一输出。
func ExecuteWebSearch(ctx context.Context, searcher web_search_system.Searcher, input WebSearchInput) (map[string]interface{}, error) {
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

type WebSearchHandler struct{}

func (h *WebSearchHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseWebSearchInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid web search args: %v", err),
		}}, nil
	}

	cfg, err := loadWebSearchConfig("config/config.json")
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"query":  strings.TrimSpace(input.Query),
			"status": "failed",
			"error":  fmt.Sprintf("load config failed: %v", err),
		}}, nil
	}

	payload, err := ExecuteWebSearch(ctx, newWebSearcher(cfg.Tavily), input)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"query":  strings.TrimSpace(input.Query),
			"status": "failed",
			"error":  err.Error(),
		}}, nil
	}

	return ToolResponse{Payload: payload}, nil
}
