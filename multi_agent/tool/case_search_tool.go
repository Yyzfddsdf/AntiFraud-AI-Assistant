package tool

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"image_recognition/config"
	"image_recognition/multi_agent/case_library"

	openai "image_recognition/llm"
)

const CaseSearchToolName = "search_similar_cases"

type CaseSearchInput struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

var CaseSearchTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        CaseSearchToolName,
		Description: "根据查询语句检索历史案件库中的相似案件，返回按相似度排序的结果。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "用于检索的案件查询语句，建议包含关键实体、场景和风险线索。",
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

// SearchSimilarCases 执行完整检索链路：
// 1) 文本 query -> embedding 向量；
// 2) 向量与历史案件库全量向量做余弦相似度排序；
// 3) 按 topK 返回格式化后的案件摘要。
func SearchSimilarCases(query string, topK int) ([]string, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	queryVector, err := GenerateQueryEmbedding(trimmedQuery)
	if err != nil {
		return nil, 0, err
	}

	results, appliedTopK, err := case_library.SearchTopKSimilarCasesByVector(queryVector, topK)
	if err != nil {
		return nil, appliedTopK, err
	}
	if len(results) == 0 {
		return []string{}, appliedTopK, nil
	}

	cases := make([]string, 0, len(results))
	for index, item := range results {
		keywordText := noneFallback("")
		if len(item.Keywords) > 0 {
			keywordText = strings.Join(item.Keywords, "、")
		}
		violatedLawText := noneFallback(item.ViolatedLaw)

		description := noneFallback(item.CaseDescription)
		cases = append(cases, fmt.Sprintf(
			"TOP%d | case_id:%s | score:%.4f | title:%s | target_group:%s | risk:%s | keywords:%s | description:%s | violated_law:%s",
			index+1,
			item.CaseID,
			item.Similarity,
			item.Title,
			item.TargetGroup,
			item.RiskLevel,
			keywordText,
			description,
			violatedLawText,
		))
	}
	return cases, appliedTopK, nil
}

func noneFallback(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "none"
	}
	return trimmed
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

	cases, appliedTopK, err := SearchSimilarCases(input.Query, input.TopK)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"query":           input.Query,
			"requested_top_k": input.TopK,
			"applied_top_k":   appliedTopK,
			"error":           err.Error(),
			"cases":           []string{},
		}}, nil
	}

	return ToolResponse{Payload: map[string]interface{}{
		"query":           strings.TrimSpace(input.Query),
		"requested_top_k": input.TopK,
		"applied_top_k":   appliedTopK,
		"cases":           cases,
	}}, nil
}
