package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const FinalReportToolName = "submit_final_report"

type FinalReportPayload struct {
	Summary      string   `json:"summary"`
	TextFinding  string   `json:"text_finding"`
	ImageFinding string   `json:"image_finding"`
	VideoFinding string   `json:"video_finding"`
	AudioFinding string   `json:"audio_finding"`
	RiskSignals  []string `json:"risk_signals"`
	RiskLevel    string   `json:"risk_level"`
	RiskReason   string   `json:"risk_reason"`
	NextActions  []string `json:"next_actions"`
}

var FinalReportTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        FinalReportToolName,
		Description: "提交最终完整报告的结构化字段",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"summary": map[string]interface{}{
					"type":        "string",
					"description": "综合摘要",
				},
				"text_finding": map[string]interface{}{
					"type":        "string",
					"description": "文本维度关键发现",
				},
				"image_finding": map[string]interface{}{
					"type":        "string",
					"description": "图像维度关键发现",
				},
				"video_finding": map[string]interface{}{
					"type":        "string",
					"description": "视频维度关键发现",
				},
				"audio_finding": map[string]interface{}{
					"type":        "string",
					"description": "音频维度关键发现",
				},
				"risk_signals": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "风险信号清单",
				},
				"risk_level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"低", "中", "高"},
					"description": "初步风险等级",
				},
				"risk_reason": map[string]interface{}{
					"type":        "string",
					"description": "风险等级理由",
				},
				"next_actions": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "建议的下一步核查动作",
				},
			},
			"required": []string{
				"summary",
				"text_finding",
				"image_finding",
				"video_finding",
				"audio_finding",
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
	riskLevel := strings.TrimSpace(payload.RiskLevel)
	if riskLevel != "低" && riskLevel != "中" && riskLevel != "高" {
		riskLevel = "中"
	}

	var report strings.Builder
	report.WriteString("1. 综合摘要\n")
	report.WriteString(strings.TrimSpace(payload.Summary))
	report.WriteString("\n\n2. 多模态关键发现\n")
	report.WriteString("- 文本：")
	report.WriteString(strings.TrimSpace(payload.TextFinding))
	report.WriteString("\n- 图像：")
	report.WriteString(strings.TrimSpace(payload.ImageFinding))
	report.WriteString("\n- 视频：")
	report.WriteString(strings.TrimSpace(payload.VideoFinding))
	report.WriteString("\n- 音频：")
	report.WriteString(strings.TrimSpace(payload.AudioFinding))

	report.WriteString("\n\n3. 风险信号清单\n")
	if len(payload.RiskSignals) == 0 {
		report.WriteString("- 未识别到明确风险信号\n")
	} else {
		for _, signal := range payload.RiskSignals {
			signal = strings.TrimSpace(signal)
			if signal == "" {
				continue
			}
			report.WriteString("- ")
			report.WriteString(signal)
			report.WriteString("\n")
		}
	}

	report.WriteString("\n4. 初步风险等级与理由\n")
	report.WriteString("- 风险等级：")
	report.WriteString(riskLevel)
	report.WriteString("\n- 理由：")
	report.WriteString(strings.TrimSpace(payload.RiskReason))

	report.WriteString("\n\n5. 建议的下一步核查动作\n")
	if len(payload.NextActions) == 0 {
		report.WriteString("- 建议补充更多上下文后复核\n")
	} else {
		for _, action := range payload.NextActions {
			action = strings.TrimSpace(action)
			if action == "" {
				continue
			}
			report.WriteString("- ")
			report.WriteString(action)
			report.WriteString("\n")
		}
	}

	return strings.TrimSpace(report.String())
}

type FinalReportHandler struct{}

func (h *FinalReportHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	payload, err := ParseFinalReportPayload(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("parse final report tool payload failed: %v", err)}}, nil
	}
	return ToolResponse{
		Payload:        map[string]interface{}{"status": "success", "message": "final report submitted successfully"},
		FinalResultStr: FormatFinalReport(payload),
	}, nil
}
