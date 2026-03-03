package tool

import (
	"encoding/json"
	"fmt"

	openai "antifraud/llm"
)

type AnalysisResult struct {
	VisualImpression string   `json:"visual_impression"`
	KeyContent       string   `json:"key_content"`
	SuspiciousPoints []string `json:"suspicious_points"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

const AnalysisToolName = "submit_analysis_result"

var AnalysisTool = openai.Tool{
	Type: "function",
	Function: &openai.FunctionDefinition{
		Name:        AnalysisToolName,
		Description: "提交结构化分析结果，包含视觉感受、关键信息和可疑点清单。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"visual_impression": map[string]interface{}{
					"type":        "string",
					"description": "整体视觉感受与显著风险特征。",
				},
				"key_content": map[string]interface{}{
					"type":        "string",
					"description": "提取出的客观关键信息。",
				},
				"suspicious_points": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "可疑点列表。",
				},
			},
			"required": []string{"visual_impression", "key_content", "suspicious_points"},
		},
	},
}

func ParseAnalysisResult(arguments string) (AnalysisResult, error) {
	var result AnalysisResult
	err := json.Unmarshal([]byte(arguments), &result)
	return result, err
}

func FormatAnalysisResult(result AnalysisResult) string {
	output := "【整体视觉感受】\n" + result.VisualImpression + "\n\n"
	output += "【关键信息提取】\n" + result.KeyContent + "\n\n"
	output += "【可疑点清单】\n"
	if len(result.SuspiciousPoints) == 0 {
		output += "- 未发现明显可疑信号\n"
	} else {
		for i, point := range result.SuspiciousPoints {
			output += fmt.Sprintf("%d. %s\n", i+1, point)
		}
	}
	return output
}
