package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"antifraud/embedding"
	openai "antifraud/llm"
	"antifraud/multi_agent/case_library"
)

const ChatSearchSimilarCasesToolName = "search_similar_cases"

type ChatSearchSimilarCasesInput struct {
	Query       string `json:"query"`
	TopK        int    `json:"top_k,omitempty"`
	TargetGroup string `json:"target_group,omitempty"`
	ScamType    string `json:"scam_type,omitempty"`
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
				"target_group": buildChatTargetGroupSchema("可选，按目标人群精确过滤后再做向量召回。必须来自 config/target_groups.json 配置。"),
				"scam_type":    buildChatScamTypeSchema("可选，按诈骗类型精确过滤后再做向量召回。必须来自 config/scam_types.json 配置。"),
			},
			"required": []string{"query"},
		},
	},
}

func buildChatTargetGroupSchema(description string) map[string]interface{} {
	allowed := append([]string{}, case_library.ListTargetGroups()...)
	trimmedDesc := strings.TrimSpace(description)
	if trimmedDesc == "" {
		trimmedDesc = "目标人群。"
	}

	if len(allowed) > 0 {
		trimmedDesc = fmt.Sprintf("%s 可选值：%s。", trimmedDesc, strings.Join(allowed, "、"))
	} else {
		trimmedDesc = fmt.Sprintf("%s 可选值读取失败，请检查 config/target_groups.json。", trimmedDesc)
	}

	schema := map[string]interface{}{
		"type":        "string",
		"description": trimmedDesc,
	}
	if len(allowed) > 0 {
		schema["enum"] = allowed
	}
	return schema
}

func buildChatScamTypeSchema(description string) map[string]interface{} {
	allowed := append([]string{}, case_library.ListScamTypes()...)
	trimmedDesc := strings.TrimSpace(description)
	if trimmedDesc == "" {
		trimmedDesc = "诈骗类型。"
	}

	if len(allowed) > 0 {
		trimmedDesc = fmt.Sprintf("%s 可选值：%s。", trimmedDesc, strings.Join(allowed, "、"))
	} else {
		trimmedDesc = fmt.Sprintf("%s 可选值读取失败，请检查 config/scam_types.json。", trimmedDesc)
	}

	schema := map[string]interface{}{
		"type":        "string",
		"description": trimmedDesc,
	}
	if len(allowed) > 0 {
		schema["enum"] = allowed
	}
	return schema
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

	cases, appliedTopK, searchErr := SearchSimilarCasesForChatWithFilters(ctx, input.Query, input.TopK, input.TargetGroup, input.ScamType)
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
	return SearchSimilarCasesForChatWithFilters(ctx, query, topK, "", "")
}

func SearchSimilarCasesForChatWithFilters(ctx context.Context, query string, topK int, targetGroup, scamType string) ([]string, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	queryVector, _, err := embedding.GenerateVector(ctx, trimmedQuery)
	if err != nil {
		return nil, 0, err
	}

	results, appliedTopK, err := case_library.SearchTopKSimilarCasesByVectorWithConditions(
		queryVector,
		topK,
		strings.TrimSpace(targetGroup),
		strings.TrimSpace(scamType),
	)
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
func noneFallback(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}
