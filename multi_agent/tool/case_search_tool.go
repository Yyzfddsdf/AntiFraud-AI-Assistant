package tool

import (
	"context"

	openai "image_recognition/llm"
)

const CaseSearchToolName = "search_similar_cases"

type CaseSearchInput struct {
	Query string `json:"query"`
}

var CaseSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        CaseSearchToolName,
		Description: "根据模型生成的查询语句检索相似历史案件。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "用于检索的案件查询语句，需包含关键词、实体和风险线索。",
				},
			},
			"required": []string{"query"},
		},
	},
}

// SearchSimilarCases 当前为占位实现，后续可接入数据库检索。
func SearchSimilarCases(query string) ([]string, error) {
	_ = query
	return nil, nil
}

func ParseCaseSearchInput(arguments string) (CaseSearchInput, error) {
	return ParseArgs[CaseSearchInput](arguments)
}

type CaseSearchHandler struct{}

func (h *CaseSearchHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseCaseSearchInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": err.Error()}}, nil
	}
	cases, err := SearchSimilarCases(input.Query)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"query": input.Query, "error": err.Error(), "cases": []string{}}}, nil
	}
	return ToolResponse{Payload: map[string]interface{}{"query": input.Query, "cases": cases}}, nil
}
