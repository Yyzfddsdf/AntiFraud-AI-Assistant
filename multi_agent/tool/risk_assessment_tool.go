package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "antifraud/llm"
)

const RiskAssessmentToolName = "submit_current_risk_assessment"

type RiskAssessmentInput struct {
	Impersonation            bool     `json:"impersonation"`
	Urgency                  bool     `json:"urgency"`
	ThreatPressure           bool     `json:"threat_pressure"`
	BenefitInducement        bool     `json:"benefit_inducement"`
	MoneyTransferRequest     bool     `json:"money_transfer_request"`
	VerificationCodeRequest  bool     `json:"verification_code_request"`
	RemoteControlRequest     bool     `json:"remote_control_request"`
	LinkOrAppInstallRequest  bool     `json:"link_or_app_install_request"`
	SensitiveInfoRequest     bool     `json:"sensitive_info_request"`
	PrivateAccountCollection bool     `json:"private_account_collection"`
	FakeOfficialVisuals      bool     `json:"fake_official_visuals"`
	SimilarCaseStrength      string   `json:"similar_case_strength,omitempty"`
	MultimodalEvidence       string   `json:"multimodal_evidence,omitempty"`
	VictimActionStage        string   `json:"victim_action_stage,omitempty"`
	KeyEvidence              []string `json:"key_evidence,omitempty"`
}

type RiskAssessmentResult struct {
	Score              int            `json:"score"`
	StructuredSummary  string         `json:"structured_summary"`
	DimensionBreakdown map[string]int `json:"dimension_breakdown"`
	HitRules           []string       `json:"hit_rules"`
}

var RiskAssessmentTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        RiskAssessmentToolName,
		Description: "提交当前案件的结构化风险因子，由系统计算本次案件风险分数与结构化摘要。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"impersonation":               map[string]interface{}{"type": "boolean", "description": "是否存在冒充官方/客服/领导/熟人等身份。"},
				"urgency":                     map[string]interface{}{"type": "boolean", "description": "是否存在强烈紧迫感催促。"},
				"threat_pressure":             map[string]interface{}{"type": "boolean", "description": "是否存在恐吓、威胁、处罚压力。"},
				"benefit_inducement":          map[string]interface{}{"type": "boolean", "description": "是否存在返利、赔偿、高收益等诱导。"},
				"money_transfer_request":      map[string]interface{}{"type": "boolean", "description": "是否明确要求转账、充值、刷流水。"},
				"verification_code_request":   map[string]interface{}{"type": "boolean", "description": "是否要求验证码。"},
				"remote_control_request":      map[string]interface{}{"type": "boolean", "description": "是否要求屏幕共享或远程控制。"},
				"link_or_app_install_request": map[string]interface{}{"type": "boolean", "description": "是否要求点击链接或安装应用。"},
				"sensitive_info_request":      map[string]interface{}{"type": "boolean", "description": "是否要求提供身份证、银行卡、人脸等敏感信息。"},
				"private_account_collection":  map[string]interface{}{"type": "boolean", "description": "是否要求向私人账号或非官方账号收款。"},
				"fake_official_visuals":       map[string]interface{}{"type": "boolean", "description": "是否存在仿冒官方页面、伪通知、伪客服界面等。"},
				"similar_case_strength":       buildStrengthSchema("相似案例支持强度。"),
				"multimodal_evidence":         buildStrengthSchema("多模态证据强度。"),
				"victim_action_stage":         buildVictimActionStageSchema(),
				"key_evidence": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "关键证据片段，最多 6 条。",
				},
			},
			"required": []string{
				"impersonation",
				"urgency",
				"threat_pressure",
				"benefit_inducement",
				"money_transfer_request",
				"verification_code_request",
				"remote_control_request",
				"link_or_app_install_request",
				"sensitive_info_request",
				"private_account_collection",
				"fake_official_visuals",
			},
		},
	},
}

func ParseRiskAssessmentInput(arguments string) (RiskAssessmentInput, error) {
	return ParseArgs[RiskAssessmentInput](arguments)
}

type RiskAssessmentHandler struct{}

func (h *RiskAssessmentHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseRiskAssessmentInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"status": "failed", "error": fmt.Sprintf("invalid risk assessment args: %v", err)}}, nil
	}

	result, err := CalculateRiskAssessment(input)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"status": "failed", "error": err.Error()}}, nil
	}

	return ToolResponse{
		Payload: map[string]interface{}{
			"score":  result.Score,
			"report": result.StructuredSummary,
		},
		ContextMutator: func(base context.Context) context.Context {
			return BindRiskAssessment(base, result.Score, result.StructuredSummary)
		},
	}, nil
}

func CalculateRiskAssessment(input RiskAssessmentInput) (RiskAssessmentResult, error) {
	similarCaseStrength, err := normalizeStrength(input.SimilarCaseStrength)
	if err != nil {
		return RiskAssessmentResult{}, err
	}
	multimodalEvidence, err := normalizeStrength(input.MultimodalEvidence)
	if err != nil {
		return RiskAssessmentResult{}, err
	}
	victimActionStage, victimActionLabel, victimActionScore, err := normalizeVictimActionStage(input.VictimActionStage)
	if err != nil {
		return RiskAssessmentResult{}, err
	}

	keyEvidence := sanitizeRiskEvidence(input.KeyEvidence)
	dimensions := map[string]int{
		"social_engineering": 0,
		"requested_actions":  0,
		"evidence_strength":  0,
		"loss_exposure":      victimActionScore,
	}
	hitRules := make([]string, 0, 16)

	addHit := func(condition bool, dimension string, score int, label string) {
		if !condition {
			return
		}
		dimensions[dimension] += score
		hitRules = append(hitRules, label)
	}

	addHit(input.Impersonation, "social_engineering", 10, "冒充身份")
	addHit(input.Urgency, "social_engineering", 8, "紧迫催促")
	addHit(input.ThreatPressure, "social_engineering", 10, "恐吓施压")
	addHit(input.BenefitInducement, "social_engineering", 8, "利益诱导")

	addHit(input.MoneyTransferRequest, "requested_actions", 25, "要求转账/充值")
	addHit(input.VerificationCodeRequest, "requested_actions", 25, "要求验证码")
	addHit(input.RemoteControlRequest, "requested_actions", 30, "要求远程控制")
	addHit(input.LinkOrAppInstallRequest, "requested_actions", 18, "要求点击链接/安装应用")
	addHit(input.SensitiveInfoRequest, "requested_actions", 20, "索要敏感信息")
	addHit(input.PrivateAccountCollection, "requested_actions", 18, "要求向私人账户收款")
	addHit(input.FakeOfficialVisuals, "evidence_strength", 12, "出现仿冒官方视觉证据")

	if similarCaseStrength.score > 0 {
		dimensions["evidence_strength"] += similarCaseStrength.score
		hitRules = append(hitRules, "相似案例支持："+similarCaseStrength.label)
	}
	if multimodalEvidence.score > 0 {
		dimensions["evidence_strength"] += multimodalEvidence.score
		hitRules = append(hitRules, "多模态证据："+multimodalEvidence.label)
	}
	if victimActionLabel != "" {
		hitRules = append(hitRules, "受害动作阶段："+victimActionLabel)
	}

	score := dimensions["social_engineering"] + dimensions["requested_actions"] + dimensions["evidence_strength"] + dimensions["loss_exposure"]
	if input.RemoteControlRequest || input.VerificationCodeRequest {
		score = maxInt(score, 65)
	}
	if input.MoneyTransferRequest && input.PrivateAccountCollection {
		score = maxInt(score, 70)
	}
	switch victimActionStage {
	case "transferred_money":
		score = maxInt(score, 78)
	case "multiple_transfers":
		score = maxInt(score, 88)
	}
	if score > 100 {
		score = 100
	}
	structuredSummary := buildRiskStructuredSummary(score, dimensions, hitRules, victimActionLabel, similarCaseStrength.label, multimodalEvidence.label, keyEvidence)

	return RiskAssessmentResult{
		Score:              score,
		StructuredSummary:  structuredSummary,
		DimensionBreakdown: dimensions,
		HitRules:           append([]string{}, hitRules...),
	}, nil
}

type normalizedStrength struct {
	value string
	label string
	score int
}

func buildStrengthSchema(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"enum":        []string{"none", "weak", "medium", "strong"},
		"description": description + " 可选值：none/weak/medium/strong。",
	}
}

func buildVictimActionStageSchema() map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"enum":        []string{"none", "clicked_link", "downloaded_app", "shared_sensitive_info", "transferred_money", "multiple_transfers"},
		"description": "用户已发生的动作阶段。可选：none/clicked_link/downloaded_app/shared_sensitive_info/transferred_money/multiple_transfers。",
	}
}

func normalizeStrength(raw string) (normalizedStrength, error) {
	switch strings.TrimSpace(raw) {
	case "", "none":
		return normalizedStrength{value: "none", label: "无", score: 0}, nil
	case "weak":
		return normalizedStrength{value: "weak", label: "弱", score: 4}, nil
	case "medium":
		return normalizedStrength{value: "medium", label: "中", score: 8}, nil
	case "strong":
		return normalizedStrength{value: "strong", label: "强", score: 12}, nil
	default:
		return normalizedStrength{}, fmt.Errorf("strength must be one of none/weak/medium/strong")
	}
}

func normalizeVictimActionStage(raw string) (string, string, int, error) {
	switch strings.TrimSpace(raw) {
	case "", "none":
		return "none", "未操作", 0, nil
	case "clicked_link":
		return "clicked_link", "已点击链接", 8, nil
	case "downloaded_app":
		return "downloaded_app", "已下载应用", 12, nil
	case "shared_sensitive_info":
		return "shared_sensitive_info", "已提供敏感信息", 22, nil
	case "transferred_money":
		return "transferred_money", "已转账", 30, nil
	case "multiple_transfers":
		return "multiple_transfers", "已多次转账", 38, nil
	default:
		return "", "", 0, fmt.Errorf("victim_action_stage is invalid")
	}
}

func sanitizeRiskEvidence(items []string) []string {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
		if len(cleaned) == 6 {
			break
		}
	}
	return cleaned
}

func buildRiskStructuredSummary(score int, dimensions map[string]int, hitRules []string, victimActionLabel string, similarCaseLabel string, multimodalEvidenceLabel string, keyEvidence []string) string {
	payload := map[string]interface{}{
		"score": score,
		"dimensions": map[string]int{
			"social_engineering": dimensions["social_engineering"],
			"requested_actions":  dimensions["requested_actions"],
			"evidence_strength":  dimensions["evidence_strength"],
			"loss_exposure":      dimensions["loss_exposure"],
		},
		"victim_action_stage":   victimActionLabel,
		"similar_case_strength": similarCaseLabel,
		"multimodal_evidence":   multimodalEvidenceLabel,
		"hit_rules":             append([]string{}, hitRules...),
		"key_evidence":          append([]string{}, keyEvidence...),
	}
	bytes, _ := json.Marshal(payload)
	return string(bytes)
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
