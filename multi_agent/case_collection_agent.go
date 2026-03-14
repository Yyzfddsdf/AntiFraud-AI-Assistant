package multi_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"antifraud/config"
	openai "antifraud/llm"
	"antifraud/multi_agent/tool"
)

const (
	maxCaseCollectionRounds = 24
	maxCaseCollectionCount  = 20
)

var (
	caseCollectionToolsProvider       = tool.CaseCollectionTools
	caseCollectionToolHandlerResolver = tool.GetCaseCollectionToolHandler
)

// CaseCollectionAgent 负责“搜索公开案例 -> 生成结构化案件 -> 写入待审核库”。
type CaseCollectionAgent struct {
	CommonAgent
	client       *openai.Client
	modelID      string
	systemPrompt string
}

// NewCaseCollectionAgent 按配置创建案件采集智能体。
func NewCaseCollectionAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *CaseCollectionAgent {
	common := NewCommonAgent("CaseCollectionAgent", modelCfg, retryCfg)

	return &CaseCollectionAgent{
		CommonAgent: common,
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  common.APIKey,
			BaseURL: common.BaseURL,
		}),
		modelID:      strings.TrimSpace(modelCfg.Model),
		systemPrompt: strings.TrimSpace(systemPrompt),
	}
}

// NewCaseCollectionAgentWithClient 允许测试注入自定义客户端。
func NewCaseCollectionAgentWithClient(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string, client *openai.Client) *CaseCollectionAgent {
	common := NewCommonAgent("CaseCollectionAgent", modelCfg, retryCfg)

	return &CaseCollectionAgent{
		CommonAgent:  common,
		client:       client,
		modelID:      strings.TrimSpace(modelCfg.Model),
		systemPrompt: strings.TrimSpace(systemPrompt),
	}
}

// CollectCases 使用默认用户标识执行案件采集。
func CollectCases(query string, caseCount int) error {
	return CollectCasesForUser("demo-user", query, caseCount)
}

// CollectCasesForUser 是案件采集的主入口。
func CollectCasesForUser(userID string, query string, caseCount int) error {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return fmt.Errorf("load case collection config failed: %w", err)
	}

	agent := NewCaseCollectionAgent(cfg.Agents.CaseCollection, cfg.Retry, cfg.Prompts.CaseCollection)
	return agent.CollectCases(context.Background(), userID, query, caseCount)
}

// CollectCases 执行案件采集工具闭环，直到达到目标数量或达到最大轮次。
func (a *CaseCollectionAgent) CollectCases(ctx context.Context, userID string, query string, caseCount int) error {
	trimmedQuery := strings.TrimSpace(query)
	if trimmedQuery == "" {
		return fmt.Errorf("query 不能为空")
	}
	if caseCount <= 0 || caseCount > maxCaseCollectionCount {
		return fmt.Errorf("case_count 取值范围应为 1-%d", maxCaseCollectionCount)
	}
	if a == nil || a.client == nil {
		return fmt.Errorf("case collection agent client is not initialized")
	}

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}
	ctx = tool.BindUserID(ctx, trimmedUserID)

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: buildCaseCollectionSystemPrompt(a.systemPrompt),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: buildCaseCollectionUserPrompt(trimmedQuery, caseCount),
		},
	}

	createdCount := 0

	for round := 0; round < maxCaseCollectionRounds; round++ {
		action := fmt.Sprintf("create case collection chat completion round %d", round+1)
		fmt.Printf("[CaseCollectionAgent][Round %d] request model=%s messages=%d target=%d query=%s\n",
			round+1, strings.TrimSpace(a.modelID), len(messages), caseCount, truncateForLog(trimmedQuery, 120))

		var resp openai.ChatCompletionResponse
		if err := a.Retry(action, func() error {
			var callErr error
			resp, callErr = a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       a.modelID,
				Messages:    messages,
				Tools:       caseCollectionToolsProvider(),
				ToolChoice:  "required",
				Stream:      false,
				MaxTokens:   a.MaxTokens,
				Temperature: float32(a.Temperature),
				TopP:        float32(a.TopP),
			})
			return callErr
		}); err != nil {
			return fmt.Errorf("case collection api error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return fmt.Errorf("case collection returned empty choices")
		}

		msg := resp.Choices[0].Message
		fmt.Printf("[CaseCollectionAgent][Round %d] ai_reply=%s\n", round+1, truncateForLog(msg.Content, 240))
		if len(msg.ToolCalls) > 0 {
			fmt.Printf("[CaseCollectionAgent][Round %d] tool_calls=%d\n", round+1, len(msg.ToolCalls))
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		if len(msg.ToolCalls) == 0 {
			if createdCount == 0 {
				return fmt.Errorf("model returned no tool calls and no pending review case")
			}
			return fmt.Errorf("case collection stopped early: created %d/%d pending review cases", createdCount, caseCount)
		}

		toolResponseAdded := false
		appendToolResponse := func(callID string, payload map[string]interface{}) {
			toolPayload, _ := json.Marshal(payload)
			fmt.Printf("[CaseCollectionAgent][Round %d] tool_result call_id=%s payload=%s\n",
				round+1, strings.TrimSpace(callID), truncateForLog(string(toolPayload), 320))
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: strings.TrimSpace(callID),
				Content:    string(toolPayload),
			})
			toolResponseAdded = true
		}

		for _, call := range msg.ToolCalls {
			fmt.Printf("[CaseCollectionAgent][Round %d] tool_call name=%s args=%s\n",
				round+1, call.Function.Name, truncateForLog(call.Function.Arguments, 220))

			handler := caseCollectionToolHandlerResolver(call.Function.Name)
			if handler == nil {
				appendToolResponse(call.ID, map[string]interface{}{"status": "failed", "error": "unsupported tool"})
				continue
			}

			if call.Function.Name == tool.UploadHistoricalCaseToVectorDBToolName && createdCount >= caseCount {
				appendToolResponse(call.ID, map[string]interface{}{
					"status": "failed",
					"error":  fmt.Sprintf("case quota reached: already created %d pending review cases", caseCount),
				})
				continue
			}

			response, err := handler.Handle(ctx, call.Function.Arguments)
			if err != nil {
				appendToolResponse(call.ID, map[string]interface{}{"status": "failed", "error": err.Error()})
				continue
			}

			appendToolResponse(call.ID, response.Payload)
			if call.Function.Name == tool.UploadHistoricalCaseToVectorDBToolName {
				if recordID, ok := extractPendingReviewRecordID(response.Payload); ok {
					createdCount++
					fmt.Printf("[CaseCollectionAgent][Round %d] created pending review count=%d/%d record=%s\n",
						round+1, createdCount, caseCount, strings.TrimSpace(recordID))
				}
			}
		}

		if !toolResponseAdded {
			return fmt.Errorf("tool calls returned but no tool response was added")
		}
		if createdCount >= caseCount {
			fmt.Printf("[CaseCollectionAgent] completed query=%s created=%d/%d\n",
				truncateForLog(trimmedQuery, 120), createdCount, caseCount)
			return nil
		}
	}

	if createdCount > 0 {
		return fmt.Errorf("case collection exceeded max tool rounds (%d), created %d/%d pending review cases",
			maxCaseCollectionRounds, createdCount, caseCount)
	}
	return fmt.Errorf("case collection exceeded max tool rounds (%d) without pending review case", maxCaseCollectionRounds)
}

func extractPendingReviewRecordID(payload map[string]interface{}) (string, bool) {
	status, _ := payload["status"].(string)
	if strings.TrimSpace(status) != "success" {
		return "", false
	}

	review, ok := payload["review"].(map[string]interface{})
	if !ok {
		return "", false
	}

	recordID := stringifyPayload(review["record_id"])
	if recordID == "" {
		return "", false
	}
	return recordID, true
}

func stringifyPayload(value interface{}) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func buildCaseCollectionSystemPrompt(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed != "" {
		return trimmed
	}
	return "你是一名案件库扩容助手，负责基于联网检索结果整理待审核案件草稿。所有输出必须使用中文。"
}

func buildCaseCollectionUserPrompt(query string, caseCount int) string {
	return fmt.Sprintf("请围绕以下主题搜集诈骗案件，并提交 %d 个待审核案件。\n\n搜索主题：%s",
		caseCount, strings.TrimSpace(query))
}
