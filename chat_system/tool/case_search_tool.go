package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	appcfg "antifraud/config"
	openai "antifraud/llm"
	"antifraud/multi_agent/case_library"
)

const ChatSearchSimilarCasesToolName = "search_similar_cases"

type ChatSearchSimilarCasesInput struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

var ChatSearchSimilarCasesTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ChatSearchSimilarCasesToolName,
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

func ParseChatSearchSimilarCasesInput(arguments string) (ChatSearchSimilarCasesInput, error) {
	if strings.TrimSpace(arguments) == "" {
		return ChatSearchSimilarCasesInput{}, nil
	}

	var input ChatSearchSimilarCasesInput
	if err := json.Unmarshal([]byte(arguments), &input); err != nil {
		return ChatSearchSimilarCasesInput{}, err
	}
	return input, nil
}

type ChatSearchSimilarCasesHandler struct{}

func (h *ChatSearchSimilarCasesHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	_ = userID

	input, err := ParseChatSearchSimilarCasesInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid search similar cases args: %v", err)}}, nil
	}

	cases, appliedTopK, searchErr := SearchSimilarCasesForChat(ctx, input.Query, input.TopK)
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

func SearchSimilarCasesForChat(ctx context.Context, query string, topK int) ([]string, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	queryVector, err := generateQueryEmbeddingForChat(ctx, trimmedQuery)
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
			"TOP%d | case_id:%s | score:%.4f | title:%s | target_group:%s | risk:%s | scam_type:%s | keywords:%s | description:%s | violated_law:%s",
			index+1,
			item.CaseID,
			item.Similarity,
			item.Title,
			item.TargetGroup,
			item.RiskLevel,
			noneFallback(item.ScamType),
			keywordText,
			description,
			violatedLawText,
		))
	}

	return cases, appliedTopK, nil
}

func generateQueryEmbeddingForChat(ctx context.Context, query string) ([]float64, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, fmt.Errorf("query is empty")
	}

	cfg, err := appcfg.LoadConfig("config/config.json")
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

	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}
	resp, err := client.CreateEmbeddings(callCtx, req)
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

func noneFallback(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}
