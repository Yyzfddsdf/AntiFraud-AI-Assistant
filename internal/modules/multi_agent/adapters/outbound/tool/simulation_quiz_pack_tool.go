package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "antifraud/internal/platform/llm"
)

const SubmitSimulationQuizPackToolName = "submit_simulation_quiz_pack"

var simulationStepTypeOrder = []string{
	"scenario_intro",
	"signal_identification",
	"low_risk_decision",
	"escalation_signal",
	"info_protection",
	"transfer_or_link",
	"official_verification",
	"loss_control",
	"scam_recap",
	"transfer_learning",
}

type SimulationQuizPackPayload struct {
	Title         string                   `json:"title"`
	CaseType      string                   `json:"case_type"`
	TargetPersona string                   `json:"target_persona"`
	Difficulty    string                   `json:"difficulty"`
	Intro         string                   `json:"intro"`
	Steps         []SimulationQuizPackStep `json:"steps"`
}

type SimulationQuizPackStep struct {
	StepID         string                     `json:"step_id"`
	StepType       string                     `json:"step_type"`
	Narrative      string                     `json:"narrative"`
	Question       string                     `json:"question"`
	Options        []SimulationQuizPackOption `json:"options"`
	KnowledgePoint string                     `json:"knowledge_point"`
	Difficulty     int                        `json:"difficulty"`
	TimeLimitSec   int                        `json:"time_limit_sec"`
}

type SimulationQuizPackOption struct {
	Key        string `json:"key"`
	Text       string `json:"text"`
	RiskTag    string `json:"risk_tag"`
	ScoreDelta int    `json:"score_delta"`
	Rationale  string `json:"rationale"`
}

var SubmitSimulationQuizPackTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        SubmitSimulationQuizPackToolName,
		Description: "提交一套固定10步结构的反诈模拟题包，用于后续用户答题评分。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title":          map[string]interface{}{"type": "string", "description": "题包标题"},
				"case_type":      map[string]interface{}{"type": "string", "description": "诈骗案件类型"},
				"target_persona": map[string]interface{}{"type": "string", "description": "目标人群画像"},
				"difficulty":     map[string]interface{}{"type": "string", "description": "难度（easy/medium/hard）"},
				"intro":          map[string]interface{}{"type": "string", "description": "开场导语"},
				"steps": map[string]interface{}{
					"type":     "array",
					"minItems": 10,
					"maxItems": 10,
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"step_id": map[string]interface{}{"type": "string"},
							"step_type": map[string]interface{}{
								"type": "string",
								"enum": simulationStepTypeOrder,
							},
							"narrative": map[string]interface{}{"type": "string"},
							"question":  map[string]interface{}{"type": "string"},
							"options": map[string]interface{}{
								"type":     "array",
								"minItems": 2,
								"maxItems": 4,
								"items": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"key":         map[string]interface{}{"type": "string"},
										"text":        map[string]interface{}{"type": "string"},
										"risk_tag":    map[string]interface{}{"type": "string", "enum": []string{"safe", "caution", "danger"}},
										"score_delta": map[string]interface{}{"type": "integer"},
										"rationale":   map[string]interface{}{"type": "string"},
									},
									"required": []string{"key", "text", "risk_tag", "score_delta", "rationale"},
								},
							},
							"knowledge_point": map[string]interface{}{"type": "string"},
							"difficulty":      map[string]interface{}{"type": "integer"},
							"time_limit_sec":  map[string]interface{}{"type": "integer"},
						},
						"required": []string{"step_id", "step_type", "narrative", "question", "options", "knowledge_point", "difficulty", "time_limit_sec"},
					},
				},
			},
			"required": []string{"title", "case_type", "target_persona", "difficulty", "intro", "steps"},
		},
	},
}

func ParseSimulationQuizPackPayload(arguments string) (SimulationQuizPackPayload, error) {
	return ParseArgs[SimulationQuizPackPayload](arguments)
}

func normalizeSimulationQuizPackPayload(payload SimulationQuizPackPayload) (SimulationQuizPackPayload, error) {
	payload.Title = strings.TrimSpace(payload.Title)
	payload.CaseType = strings.TrimSpace(payload.CaseType)
	payload.TargetPersona = strings.TrimSpace(payload.TargetPersona)
	payload.Difficulty = strings.TrimSpace(strings.ToLower(payload.Difficulty))
	payload.Intro = strings.TrimSpace(payload.Intro)

	if payload.Title == "" || payload.CaseType == "" || payload.TargetPersona == "" || payload.Intro == "" {
		return SimulationQuizPackPayload{}, fmt.Errorf("title/case_type/target_persona/intro 不能为空")
	}
	if payload.Difficulty == "" {
		payload.Difficulty = "medium"
	}
	if len(payload.Steps) != len(simulationStepTypeOrder) {
		return SimulationQuizPackPayload{}, fmt.Errorf("steps 必须固定为 %d 步", len(simulationStepTypeOrder))
	}

	for i := range payload.Steps {
		step := payload.Steps[i]
		step.StepID = strings.TrimSpace(step.StepID)
		if step.StepID == "" {
			step.StepID = fmt.Sprintf("step_%02d", i+1)
		}
		step.StepType = strings.TrimSpace(step.StepType)
		if step.StepType != simulationStepTypeOrder[i] {
			return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步 step_type 必须为 %s", i+1, simulationStepTypeOrder[i])
		}
		step.Narrative = strings.TrimSpace(step.Narrative)
		step.Question = strings.TrimSpace(step.Question)
		step.KnowledgePoint = strings.TrimSpace(step.KnowledgePoint)
		if step.Narrative == "" || step.Question == "" || step.KnowledgePoint == "" {
			return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步 narrative/question/knowledge_point 不能为空", i+1)
		}
		if step.Difficulty < 1 || step.Difficulty > 5 {
			step.Difficulty = 3
		}
		if step.TimeLimitSec <= 0 {
			step.TimeLimitSec = 30
		}
		if len(step.Options) < 2 || len(step.Options) > 4 {
			return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步 options 数量必须在 2-4", i+1)
		}

		seenKeys := map[string]struct{}{}
		for j := range step.Options {
			option := step.Options[j]
			option.Key = strings.TrimSpace(strings.ToUpper(option.Key))
			option.Text = strings.TrimSpace(option.Text)
			option.RiskTag = strings.TrimSpace(strings.ToLower(option.RiskTag))
			option.Rationale = strings.TrimSpace(option.Rationale)
			if option.Key == "" || option.Text == "" || option.Rationale == "" {
				return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步第 %d 个选项 key/text/rationale 不能为空", i+1, j+1)
			}
			if _, exists := seenKeys[option.Key]; exists {
				return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步选项 key 重复: %s", i+1, option.Key)
			}
			seenKeys[option.Key] = struct{}{}
			switch option.RiskTag {
			case "safe", "caution", "danger":
			default:
				return SimulationQuizPackPayload{}, fmt.Errorf("第 %d 步选项 risk_tag 非法: %s", i+1, option.RiskTag)
			}
			if option.ScoreDelta < -30 {
				option.ScoreDelta = -30
			}
			if option.ScoreDelta > 20 {
				option.ScoreDelta = 20
			}
			step.Options[j] = option
		}
		payload.Steps[i] = step
	}

	return payload, nil
}

type SubmitSimulationQuizPackHandler struct{}

func (h *SubmitSimulationQuizPackHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	_ = ctx
	payload, err := ParseSimulationQuizPackPayload(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"status": "failed", "error": fmt.Sprintf("parse simulation quiz pack payload failed: %v", err)}}, nil
	}

	normalized, err := normalizeSimulationQuizPackPayload(payload)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"status": "failed", "error": err.Error()}}, nil
	}

	return ToolResponse{
		Payload: map[string]interface{}{
			"status":  "success",
			"message": "simulation quiz pack accepted",
		},
		FinalResultStr: formatSimulationQuizPackPayload(normalized),
	}, nil
}

func formatSimulationQuizPackPayload(payload SimulationQuizPackPayload) string {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}
