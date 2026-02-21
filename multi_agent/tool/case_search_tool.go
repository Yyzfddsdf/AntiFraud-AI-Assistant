package tool

import (
	"context"
	"encoding/json"

	"github.com/sashabaranov/go-openai"
)

const CaseSearchToolName = "search_similar_cases"

type CaseSearchInput struct {
	Query string `json:"query"`
}

var CaseSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        CaseSearchToolName,
		Description: "根据输入的案件描述查询数据库中的相似案件，query 由模型自行生成",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "用于检索相似案件的查询描述，应包含案件核心特征、话术、关键实体和风险线索",
				},
			},
			"required": []string{"query"},
		},
	},
}

// SearchSimilarCases 预留数据库检索逻辑。
// TODO: 接入案件数据库检索。
func SearchSimilarCases(query string) ([]string, error) {
	_ = query
	return nil, nil
}

func ParseCaseSearchInput(arguments string) (CaseSearchInput, error) {
	var input CaseSearchInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
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
	return ToolResponse{
		Payload:       map[string]interface{}{"query": input.Query, "cases": cases},
		SetCaseSearch: true,
	}, nil
}
