package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	openai "antifraud/llm"
	"antifraud/multi_agent/case_library"
)

const UploadHistoricalCaseToVectorDBToolName = "upload_historical_case_to_vector_db"

// UploadHistoricalCaseToVectorDBInput 表示“上传向量数据库”工具输入。
// 该工具会自动完成 embedding 生成并写入 historical_case_library。
type UploadHistoricalCaseToVectorDBInput struct {
	Title           string   `json:"title"`
	TargetGroup     string   `json:"target_group"`
	RiskLevel       string   `json:"risk_level"`
	ScamType        string   `json:"scam_type"`
	CaseDescription string   `json:"case_description"`
	TypicalScripts  []string `json:"typical_scripts,omitempty"`
	Keywords        []string `json:"keywords,omitempty"`
	ViolatedLaw     string   `json:"violated_law,omitempty"`
	Suggestion      string   `json:"suggestion,omitempty"`
}

var UploadHistoricalCaseToVectorDBTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        UploadHistoricalCaseToVectorDBToolName,
		Description: "上传历史案件到向量数据库（自动向量化并入库）。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "案件标题。",
				},
				"target_group": buildTargetGroupSchema("目标人群（必填）。"),
				"risk_level":   buildRiskLevelSchema("风险等级（必填）。"),
				"scam_type":    buildScamTypeSchema("诈骗类型（必填）。必须来自 config/scam_types.json 配置。"),
				"case_description": map[string]interface{}{
					"type":        "string",
					"description": "案件描述（必填）。必须基于当前已掌握的事实客观整理，建议包含受害对象、诈骗手法、关键诱导步骤和风险线索，不要编造未被事实支持的细节。",
				},
				"typical_scripts": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "典型话术列表（可选）。仅在当前信息中能明确提炼出原话术、诱导语或高频表达时填写；没有明确依据就不要传该字段。",
				},
				"keywords": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "关键词列表（可选）。仅填写可明确抽取的诈骗关键词、平台名、话术标签或关键实体；不要为了凑字段随意概括。",
				},
				"violated_law": map[string]interface{}{
					"type":        "string",
					"description": "涉及法律条款（可选）。只有在当前信息中明确提到法律条款、罪名或监管定性时才填写；没有明确依据就不要传该字段。",
				},
				"suggestion": map[string]interface{}{
					"type":        "string",
					"description": "防范建议（可选）。可根据当前已确认的风险点提炼简洁防范建议；若信息不足以支撑建议，可不传。",
				},
			},
			"required": []string{"title", "target_group", "risk_level", "scam_type", "case_description"},
		},
	},
}

func ParseUploadHistoricalCaseToVectorDBInput(arguments string) (UploadHistoricalCaseToVectorDBInput, error) {
	return ParseArgs[UploadHistoricalCaseToVectorDBInput](arguments)
}

type UploadHistoricalCaseToVectorDBHandler struct{}

func (h *UploadHistoricalCaseToVectorDBHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseUploadHistoricalCaseToVectorDBInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid upload vector case args: %v", err),
		}}, nil
	}

	record, createErr := case_library.CreatePendingReview(CurrentUserID(ctx), case_library.CreateHistoricalCaseInput{
		Title:           input.Title,
		TargetGroup:     input.TargetGroup,
		RiskLevel:       input.RiskLevel,
		ScamType:        input.ScamType,
		CaseDescription: input.CaseDescription,
		TypicalScripts:  append([]string{}, input.TypicalScripts...),
		Keywords:        append([]string{}, input.Keywords...),
		ViolatedLaw:     input.ViolatedLaw,
		Suggestion:      input.Suggestion,
	})
	if createErr != nil {
		payload := map[string]interface{}{
			"status":                "failed",
			"error":                 createErr.Error(),
			"allowed_target_groups": append([]string{}, case_library.ListTargetGroups()...),
			"allowed_risk_levels":   append([]string{}, case_library.FixedRiskLevels...),
			"allowed_scam_types":    append([]string{}, case_library.ListScamTypes()...),
		}

		if !case_library.IsValidationError(createErr) {
			payload["message"] = "pending review case storage failed"
		}
		return ToolResponse{Payload: payload}, nil
	}

	return ToolResponse{Payload: map[string]interface{}{
		"status":  "success",
		"message": "案件已提交，等待管理员审核后入库",
		"review": map[string]interface{}{
			"record_id":  strings.TrimSpace(record.RecordID),
			"user_id":    strings.TrimSpace(record.UserID),
			"title":      record.Title,
			"risk_level": record.RiskLevel,
			"scam_type":  record.ScamType,
			"status":     record.Status,
			"created_at": record.CreatedAt.Format(time.RFC3339),
		},
	}}, nil
}

func buildTargetGroupSchema(description string) map[string]interface{} {
	allowed := append([]string{}, case_library.ListTargetGroups()...)
	trimmedDesc := strings.TrimSpace(description)
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

func buildRiskLevelSchema(description string) map[string]interface{} {
	allowed := append([]string{}, case_library.FixedRiskLevels...)
	trimmedDesc := strings.TrimSpace(description)
	if len(allowed) > 0 {
		trimmedDesc = fmt.Sprintf("%s 可选值：%s。", trimmedDesc, strings.Join(allowed, "、"))
	}

	return map[string]interface{}{
		"type":        "string",
		"description": trimmedDesc,
		"enum":        allowed,
	}
}
