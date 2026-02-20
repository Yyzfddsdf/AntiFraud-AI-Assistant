package tool

import (
	"encoding/json"
	"fmt"
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

var AnalysisToolDefinition = map[string]interface{}{
	"type": "function",
	"function": map[string]interface{}{
		"name":        AnalysisToolName,
		"description": "提交分析结果，包含整体视觉感受、关键内容提取和可疑点清单",
		"parameters": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"visual_impression": map[string]interface{}{
					"type":        "string",
					"description": "整体视觉感受（主观特征）：描述整体风格、高风险视觉特征（颜色、广告、诱导按钮等）",
				},
				"key_content": map[string]interface{}{
					"type":        "string",
					"description": "关键内容提取（客观信息）：提取文字信息（APP名称、网址、金额、电话、机构名）和核心场景描述",
				},
				"suspicious_points": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "可疑点清单（仅列出，不判断）：逐条列出的异常之处",
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
	output := "【整体视觉感受（主观特征）】\n" + result.VisualImpression + "\n\n"
	output += "【关键内容提取（客观信息）】\n" + result.KeyContent + "\n\n"
	output += "【可疑点清单（仅列出，不判断）】\n"
	if len(result.SuspiciousPoints) == 0 {
		output += "- 未发现明显视觉异常\n"
	} else {
		for i, point := range result.SuspiciousPoints {
			output += fmt.Sprintf("%d. %s\n", i+1, point)
		}
	}
	return output
}
