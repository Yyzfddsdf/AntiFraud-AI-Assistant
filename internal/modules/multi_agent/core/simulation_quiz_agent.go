package multi_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"antifraud/internal/modules/multi_agent/adapters/outbound/tool"
	"antifraud/internal/platform/config"
	openai "antifraud/internal/platform/llm"
)

var (
	simulationQuizToolsProvider       = tool.SimulationQuizTools
	simulationQuizToolHandlerResolver = tool.GetSimulationQuizToolHandler
)

type SimulationQuizGenerationInput struct {
	CaseType      string `json:"case_type"`
	TargetPersona string `json:"target_persona"`
	Difficulty    string `json:"difficulty"`
	Locale        string `json:"locale"`
	QuestionCount int    `json:"question_count"`
}

// SimulationQuizAgent 负责按固定结构生成反诈模拟题包。
type SimulationQuizAgent struct {
	CommonAgent
	client       *openai.Client
	modelID      string
	systemPrompt string
}

func NewSimulationQuizAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *SimulationQuizAgent {
	common := NewCommonAgent("SimulationQuizAgent", modelCfg, retryCfg)
	return &SimulationQuizAgent{
		CommonAgent: common,
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  common.APIKey,
			BaseURL: common.BaseURL,
		}),
		modelID:      strings.TrimSpace(modelCfg.Model),
		systemPrompt: strings.TrimSpace(systemPrompt),
	}
}

func GenerateSimulationQuizPack(input SimulationQuizGenerationInput) (tool.SimulationQuizPackPayload, error) {
	cfg, err := config.LoadConfig("internal/platform/config/config.json")
	if err != nil {
		return tool.SimulationQuizPackPayload{}, fmt.Errorf("load simulation quiz config failed: %w", err)
	}

	agent := NewSimulationQuizAgent(cfg.Agents.SimulationQuiz, cfg.Retry, cfg.Prompts.SimulationQuiz)
	return agent.Generate(context.Background(), input)
}

func (a *SimulationQuizAgent) Generate(ctx context.Context, input SimulationQuizGenerationInput) (tool.SimulationQuizPackPayload, error) {
	if a == nil || a.client == nil {
		return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz agent client is not initialized")
	}

	normalized := normalizeSimulationQuizInput(input)
	repairHint := ""
	maxAttempts := 2

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		payload, err := a.generateOnce(ctx, normalized, repairHint)
		if err == nil {
			return payload, nil
		}
		repairHint = err.Error()
	}

	return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz generation failed after %d attempts: %s", maxAttempts, repairHint)
}

func (a *SimulationQuizAgent) generateOnce(ctx context.Context, input SimulationQuizGenerationInput, repairHint string) (tool.SimulationQuizPackPayload, error) {
	userPrompt := buildSimulationQuizUserPrompt(input, repairHint)
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: firstNonEmptyForLog(a.systemPrompt, defaultSimulationQuizSystemPrompt())},
		{Role: openai.ChatMessageRoleUser, Content: userPrompt},
	}

	maxRounds := 8
	var finalJSON string

	for round := 0; round < maxRounds; round++ {
		action := fmt.Sprintf("create simulation quiz chat completion round %d", round+1)
		var resp openai.ChatCompletionResponse
		if err := a.Retry(action, func() error {
			var callErr error
			resp, callErr = a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       a.modelID,
				Messages:    messages,
				Tools:       simulationQuizToolsProvider(),
				ToolChoice:  "required",
				Stream:      false,
				MaxTokens:   a.MaxTokens,
				Temperature: float32(a.Temperature),
				TopP:        float32(a.TopP),
			})
			return callErr
		}); err != nil {
			return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz api error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz returned empty choices")
		}

		msg := resp.Choices[0].Message
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: msg.Content, ToolCalls: msg.ToolCalls})
		if len(msg.ToolCalls) == 0 {
			if finalJSON != "" {
				break
			}
			continue
		}

		toolResponseAdded := false
		for _, call := range msg.ToolCalls {
			handler := simulationQuizToolHandlerResolver(call.Function.Name)
			if handler == nil {
				content := `{"status":"failed","error":"unsupported tool"}`
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: call.ID, Content: content})
				toolResponseAdded = true
				continue
			}

			response, err := handler.Handle(ctx, call.Function.Arguments)
			if err != nil {
				content := fmt.Sprintf(`{"status":"failed","error":%q}`, err.Error())
				messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: call.ID, Content: content})
				toolResponseAdded = true
				continue
			}

			encoded, _ := json.Marshal(response.Payload)
			messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleTool, ToolCallID: call.ID, Content: string(encoded)})
			toolResponseAdded = true
			if strings.TrimSpace(response.FinalResultStr) != "" {
				finalJSON = strings.TrimSpace(response.FinalResultStr)
			}
		}

		if !toolResponseAdded {
			return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz tool calls returned but no tool response was added")
		}
		if finalJSON != "" {
			break
		}
	}

	if finalJSON == "" {
		return tool.SimulationQuizPackPayload{}, fmt.Errorf("simulation quiz did not return final pack")
	}

	var pack tool.SimulationQuizPackPayload
	if err := json.Unmarshal([]byte(finalJSON), &pack); err != nil {
		return tool.SimulationQuizPackPayload{}, fmt.Errorf("unmarshal simulation quiz pack failed: %w", err)
	}
	return pack, nil
}

func normalizeSimulationQuizInput(input SimulationQuizGenerationInput) SimulationQuizGenerationInput {
	result := SimulationQuizGenerationInput{
		CaseType:      strings.TrimSpace(input.CaseType),
		TargetPersona: strings.TrimSpace(input.TargetPersona),
		Difficulty:    strings.TrimSpace(strings.ToLower(input.Difficulty)),
		Locale:        strings.TrimSpace(input.Locale),
		QuestionCount: input.QuestionCount,
	}
	if result.CaseType == "" {
		result.CaseType = "综合诈骗"
	}
	if result.TargetPersona == "" {
		result.TargetPersona = "普通居民"
	}
	switch result.Difficulty {
	case "easy", "medium", "hard":
	default:
		result.Difficulty = "medium"
	}
	if result.Locale == "" {
		result.Locale = "zh-CN"
	}
	if result.QuestionCount <= 0 {
		result.QuestionCount = 10
	}
	return result
}

func buildSimulationQuizUserPrompt(input SimulationQuizGenerationInput, repairHint string) string {
	var builder strings.Builder
	builder.WriteString("请生成一套反诈模拟题包。\n")
	builder.WriteString("输出语言：中文。\n")
	builder.WriteString("结构要求：固定 10 步，step_type 必须按如下顺序：scenario_intro, signal_identification, low_risk_decision, escalation_signal, info_protection, transfer_or_link, official_verification, loss_control, scam_recap, transfer_learning。\n")
	builder.WriteString("每步必须包含 2-4 个选项，选项 risk_tag 仅允许 safe/caution/danger，score_delta 必须在 [-30, 20]。\n")
	builder.WriteString("严禁生成可直接执行诈骗的具体操作步骤。\n")
	builder.WriteString(fmt.Sprintf("案件类型：%s\n", input.CaseType))
	builder.WriteString(fmt.Sprintf("目标人群：%s\n", input.TargetPersona))
	builder.WriteString(fmt.Sprintf("难度：%s\n", input.Difficulty))
	builder.WriteString(fmt.Sprintf("本地化：%s\n", input.Locale))
	builder.WriteString(fmt.Sprintf("题目步数：%d\n", input.QuestionCount))
	if strings.TrimSpace(repairHint) != "" {
		builder.WriteString("上一次结构校验失败，请按以下错误修复并重提完整题包：\n")
		builder.WriteString(strings.TrimSpace(repairHint))
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func defaultSimulationQuizSystemPrompt() string {
	return "你是反诈模拟题目生成智能体。你必须通过 submit_simulation_quiz_pack 工具提交完整题包，不允许输出散文或解释。每一道题的正确选项分布必须有变化，不允许所有题目都使用同一个答案字母作为正确答案。"
}
