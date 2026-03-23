package tool

import (
	"context"
	"fmt"
	"strings"

	"antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	"antifraud/internal/platform/embedding"

	openai "antifraud/internal/platform/llm"
)

const CaseSearchToolName = "search_similar_cases"

type CaseSearchInput struct {
	Query       string `json:"query"`
	TopK        int    `json:"top_k,omitempty"`
	TargetGroup string `json:"target_group,omitempty"`
	ScamType    string `json:"scam_type,omitempty"`
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
				"target_group": buildTargetGroupSchema("可选，按目标人群精确过滤后再做向量召回。必须来自 config/target_groups.json 配置。"),
				"scam_type":    buildScamTypeSchema("可选，按诈骗类型精确过滤后再做向量召回。必须来自 config/scam_types.json 配置。"),
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
	return SearchSimilarCasesWithFilters(query, topK, "", "")
}

func SearchSimilarCasesWithFilters(query string, topK int, targetGroup, scamType string) ([]string, int, error) {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	queryVector, _, err := embedding.GenerateVector(context.Background(), trimmedQuery)
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

func ParseCaseSearchInput(arguments string) (CaseSearchInput, error) {
	return ParseArgs[CaseSearchInput](arguments)
}

type CaseSearchHandler struct{}

func (h *CaseSearchHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseCaseSearchInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": err.Error()}}, nil
	}

	cases, appliedTopK, err := SearchSimilarCasesWithFilters(input.Query, input.TopK, input.TargetGroup, input.ScamType)
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
