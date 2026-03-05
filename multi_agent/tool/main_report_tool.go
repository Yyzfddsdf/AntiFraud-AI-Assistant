package tool

import (
	"context"
	"fmt"
	"strings"

	openai "antifraud/llm"
)

const FinalReportToolName = "submit_final_report"

type FinalReportPayload struct {
	Summary              string   `json:"summary"`
	TextFinding          string   `json:"text_finding"`
	ImageFinding         string   `json:"image_finding"`
	VideoFinding         string   `json:"video_finding"`
	AudioFinding         string   `json:"audio_finding"`
	ScamType             string   `json:"scam_type"`
	RiskSignals          []string `json:"risk_signals"`
	RiskLevel            string   `json:"risk_level"`
	RiskReason           string   `json:"risk_reason"`
	NextActions          []string `json:"next_actions"`
	AttackSteps          []string `json:"attack_steps"`
	ScamKeywordSentences []string `json:"scam_keyword_sentences"`
}

var FinalReportTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        FinalReportToolName,
		Description: "提交最终结构化反诈骗分析报告。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "综合摘要。",
				},
				"text_finding": map[string]interface{}{
					"type":        "string",
					"description": "文本模态关键发现。",
				},
				"image_finding": map[string]interface{}{
					"type":        "string",
					"description": "图像模态关键发现。",
				},
				"video_finding": map[string]interface{}{
					"type":        "string",
					"description": "视频模态关键发现。",
				},
				"audio_finding": map[string]interface{}{
					"type":        "string",
					"description": "音频模态关键发现。",
				},
				"scam_type": buildScamTypeSchema("诈骗类型（必填）。必须来自 config/scam_types.json 配置。"),
				"risk_signals": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "风险信号列表。",
				},
				"risk_level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"低", "中", "高"},
					"description": "整体风险等级，仅允许：低/中/高。",
				},
				"risk_reason": map[string]interface{}{
					"type":        "string",
					"description": "风险等级判定理由。",
				},
				"next_actions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "建议的下一步核查动作。",
				},
				"attack_steps": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "诈骗链路步骤（可选，严格数组）。如有充分证据可提供；如无可不传。每个元素只能写一个步骤，按时间顺序排列；禁止把多个步骤写成单个元素（如“步骤A→步骤B→步骤C”）。",
				},
				"scam_keyword_sentences": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "诈骗关键词句（可选，严格数组）。如有明确关键词句可提供；如无可不传。每个元素只包含一个关键词或关键句；禁止把多个关键词句写在同一个元素里。",
				},
			},
			"required": []string{
				"summary",
				"text_finding",
				"image_finding",
				"video_finding",
				"audio_finding",
				"scam_type",
				"risk_signals",
				"risk_level",
				"risk_reason",
				"next_actions",
			},
		},
	},
}

func ParseFinalReportPayload(arguments string) (FinalReportPayload, error) {
	return ParseArgs[FinalReportPayload](arguments)
}

func FormatFinalReport(payload FinalReportPayload) string {
	riskLevel := normalizeFinalReportRiskLevel(payload.RiskLevel)
	riskSignals := sanitizeNonEmptyList(payload.RiskSignals)
	nextActions := sanitizeNonEmptyList(payload.NextActions)
	attackSteps := sanitizeNonEmptyList(payload.AttackSteps)
	scamKeywordSentences := sanitizeNonEmptyList(payload.ScamKeywordSentences)

	var report strings.Builder
	report.WriteString("1. 综合摘要\n")
	report.WriteString(strings.TrimSpace(payload.Summary))
	report.WriteString("\n\n2. 多模态关键发现\n")
	report.WriteString("- 文本: ")
	report.WriteString(strings.TrimSpace(payload.TextFinding))
	report.WriteString("\n- 图像: ")
	report.WriteString(strings.TrimSpace(payload.ImageFinding))
	report.WriteString("\n- 视频: ")
	report.WriteString(strings.TrimSpace(payload.VideoFinding))
	report.WriteString("\n- 音频: ")
	report.WriteString(strings.TrimSpace(payload.AudioFinding))

	report.WriteString("\n\n3. 风险信号\n")
	if len(riskSignals) == 0 {
		report.WriteString("- 未发现明确风险信号\n")
	} else {
		for _, signal := range riskSignals {
			report.WriteString("- ")
			report.WriteString(signal)
			report.WriteString("\n")
		}
	}

	report.WriteString("\n4. 风险等级与理由\n")
	report.WriteString("- 风险等级: ")
	report.WriteString(riskLevel)
	report.WriteString("\n- 诈骗类型: ")
	report.WriteString(strings.TrimSpace(payload.ScamType))
	report.WriteString("\n- 理由: ")
	report.WriteString(strings.TrimSpace(payload.RiskReason))

	report.WriteString("\n\n5. 建议的下一步动作\n")
	if len(nextActions) == 0 {
		report.WriteString("- 补充上下文信息后再次核验\n")
	} else {
		for _, action := range nextActions {
			report.WriteString("- ")
			report.WriteString(action)
			report.WriteString("\n")
		}
	}

	// 两个字段是可选项：为空就不输出对应章节，不补固定兜底文案。
	nextSectionID := 6
	if len(attackSteps) > 0 {
		report.WriteString("\n\n")
		report.WriteString(fmt.Sprintf("%d. 诈骗链路还原\n", nextSectionID))
		for _, step := range attackSteps {
			report.WriteString("- ")
			report.WriteString(step)
			report.WriteString("\n")
		}
		nextSectionID++
	}

	if len(scamKeywordSentences) > 0 {
		report.WriteString("\n\n")
		report.WriteString(fmt.Sprintf("%d. 诈骗关键词句\n", nextSectionID))
		for _, sentence := range scamKeywordSentences {
			report.WriteString("- ")
			report.WriteString(sentence)
			report.WriteString("\n")
		}
	}

	return strings.TrimSpace(report.String())
}

func sanitizeNonEmptyList(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

func normalizeFinalReportRiskLevel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "高":
		return "高"
	case "低":
		return "低"
	default:
		return "中"
	}
}

type FinalReportHandler struct{}

func (h *FinalReportHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	payload, err := ParseFinalReportPayload(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("parse final report tool payload failed: %v", err)}}, nil
	}

	normalizedScamType, scamTypeErr := normalizeAndValidateScamType(payload.ScamType)
	if scamTypeErr != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid scam_type: %v", scamTypeErr)}}, nil
	}
	payload.ScamType = normalizedScamType

	return ToolResponse{
		Payload:        map[string]interface{}{"status": "success", "message": "最终报告已提交"},
		FinalResultStr: FormatFinalReport(payload),
	}, nil
}
