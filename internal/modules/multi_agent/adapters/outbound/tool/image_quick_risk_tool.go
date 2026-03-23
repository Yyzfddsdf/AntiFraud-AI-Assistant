package tool

import (
	"encoding/json"
	"fmt"
	"strings"

	openai "antifraud/internal/platform/llm"
)

const ImageQuickRiskToolName = "submit_image_quick_risk_result"

type ImageQuickRiskResult struct {
	RiskLevel string `json:"risk_level"`
	Reason    string `json:"reason"`
}

var ImageQuickRiskTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ImageQuickRiskToolName,
		Description: "提交图片快速风险识别结果，仅包含风险等级和理由。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"risk_level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"高", "中", "低"},
					"description": "图片风险等级，只能是高/中/低。",
				},
				"reason": map[string]interface{}{
					"type":        "string",
					"description": "给出简洁、客观、可追踪的风险判断理由。",
				},
			},
			"required": []string{"risk_level", "reason"},
		},
	},
}

func ParseImageQuickRiskResult(arguments string) (ImageQuickRiskResult, error) {
	result, err := ParseArgs[ImageQuickRiskResult](arguments)
	if err != nil {
		return ImageQuickRiskResult{}, err
	}

	result.RiskLevel = strings.TrimSpace(result.RiskLevel)
	result.Reason = strings.TrimSpace(result.Reason)
	switch result.RiskLevel {
	case "高", "中", "低":
	default:
		return ImageQuickRiskResult{}, fmt.Errorf("risk_level must be one of 高/中/低")
	}
	if result.Reason == "" {
		return ImageQuickRiskResult{}, fmt.Errorf("reason is required")
	}
	return result, nil
}

func FormatImageQuickRiskResult(result ImageQuickRiskResult) string {
	payload := map[string]string{
		"risk_level": result.RiskLevel,
		"reason":     result.Reason,
	}
	bytes, _ := json.Marshal(payload)
	return string(bytes)
}
