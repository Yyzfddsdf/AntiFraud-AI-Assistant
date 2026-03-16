package tool

import (
	"context"
	"fmt"
	"strings"

	openai "antifraud/llm"
)

const DynamicRiskLevelToolName = "resolve_dynamic_risk_level"

type DynamicRiskLevelInput struct {
	KnowledgeBaseHit string `json:"knowledge_base_hit"`
	UserHistoryHit   string `json:"user_history_hit"`
}

type DynamicRiskLevelResult struct {
	CurrentScore     int    `json:"current_score"`
	HistoricalScore  int    `json:"historical_score"`
	DynamicThreshold int    `json:"dynamic_threshold"`
	RiskLevel        string `json:"risk_level"`
}

var DynamicRiskLevelTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        DynamicRiskLevelToolName,
		Description: "根据 historical_score 推导动态阈值，并结合当前案件分数与相似命中（也就是索引到的相似案件情况）计算最终风险等级。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"knowledge_base_hit": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"high", "low", "none"},
					"description": "知识库相似命中结果。high 表示命中高风险相似案件，low 表示命中低风险相似案件，none 表示未命中。",
				},
				"user_history_hit": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"high", "low", "none"},
					"description": "用户历史相似命中结果。high 表示命中高风险相似案件，low 表示命中低风险相似案件，none 表示未命中。",
				},
			},
			"required": []string{"knowledge_base_hit", "user_history_hit"},
		},
	},
}

type DynamicRiskLevelHandler struct{}

func (h *DynamicRiskLevelHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseArgs[DynamicRiskLevelInput](args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid dynamic risk level args: %v", err)}}, nil
	}

	assessment := CurrentRiskAssessment(ctx)
	if assessment.Score <= 0 && assessment.StructuredSummary == "" {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("risk assessment is missing, please call %s first", RiskAssessmentToolName)}}, nil
	}
	historicalScore, ok := CurrentHistoricalScore(ctx)
	if !ok {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("historical score is missing, please call %s first", QueryUserInfoToolName)}}, nil
	}

	result, err := ResolveDynamicRiskLevel(assessment.Score, historicalScore, input.KnowledgeBaseHit, input.UserHistoryHit)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": err.Error()}}, nil
	}
	return ToolResponse{Payload: map[string]interface{}{
		"current_score":     result.CurrentScore,
		"historical_score":  result.HistoricalScore,
		"dynamic_threshold": result.DynamicThreshold,
		"risk_level":        result.RiskLevel,
	}}, nil
}

func ResolveDynamicRiskLevel(currentScore int, historicalScore int, knowledgeBaseHit string, userHistoryHit string) (DynamicRiskLevelResult, error) {
	threshold := DynamicThresholdFromHistoricalScore(historicalScore)
	kbAdjustment, kbLabel, err := normalizeRiskMatch(knowledgeBaseHit)
	if err != nil {
		return DynamicRiskLevelResult{}, err
	}
	userAdjustment, userLabel, err := normalizeRiskMatch(userHistoryHit)
	if err != nil {
		return DynamicRiskLevelResult{}, err
	}

	adjustedScore := normalizeScore(currentScore + kbAdjustment + userAdjustment)
	riskLevel := "低"
	if adjustedScore > threshold {
		if adjustedScore >= threshold+10 && (kbLabel == "high" || userLabel == "high") {
			riskLevel = "高"
		} else {
			riskLevel = "中"
		}
	} else if kbLabel == "high" || userLabel == "high" {
		riskLevel = "中"
	}

	return DynamicRiskLevelResult{
		CurrentScore:     adjustedScore,
		HistoricalScore:  normalizeScore(historicalScore),
		DynamicThreshold: threshold,
		RiskLevel:        riskLevel,
	}, nil
}

func DynamicThresholdFromHistoricalScore(historicalScore int) int {
	score := normalizeScore(historicalScore)
	switch {
	case score <= 20:
		return 60
	case score <= 40:
		return 55
	case score <= 60:
		return 50
	case score <= 80:
		return 45
	default:
		return 40
	}
}

func normalizeScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func normalizeRiskMatch(raw string) (int, string, error) {
	switch strings.TrimSpace(raw) {
	case "", "none":
		return 0, "none", nil
	case "high":
		return 8, "high", nil
	case "low":
		return -8, "low", nil
	default:
		return 0, "", fmt.Errorf("risk match must be one of high/low/none")
	}
}
