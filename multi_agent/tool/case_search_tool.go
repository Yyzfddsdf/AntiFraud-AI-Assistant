package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"image_recognition/config"

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

// GenerateQueryEmbedding 将 query 文本发送到 embeddings 接口并返回向量。
func GenerateQueryEmbedding(query string) ([]float64, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("query is empty")
	}

	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return nil, fmt.Errorf("load config failed: %w", err)
	}

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  cfg.Embedding.APIKey,
		BaseURL: cfg.Embedding.BaseURL,
	})

	req := openai.EmbeddingRequest{
		Model:          cfg.Embedding.Model,
		Input:          []string{trimmedQuery},
		EncodingFormat: "float",
	}
	req.SetField("truncate", "NONE")

	resp, err := client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("create query embedding failed: %w", err)
	}
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("embedding response is empty")
	}

	sort.Slice(resp.Data, func(i, j int) bool {
		return resp.Data[i].Index < resp.Data[j].Index
	})

	vector := resp.Data[0].Embedding
	if len(vector) == 0 {
		return nil, fmt.Errorf("embedding vector is empty")
	}
	return vector, nil
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
