package multi_agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"antifraud/internal/modules/multi_agent/adapters/outbound/tool"
	"antifraud/internal/platform/config"
	openai "antifraud/internal/platform/llm"
)

const maxCaseCollectionCount = 20

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
	cfg, err := config.LoadConfig("internal/platform/config/config.json")
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

	for _, requiredToolName := range []string{tool.WebSearchToolName, tool.UploadHistoricalCaseToVectorDBToolName} {
		if !caseCollectionHasTool(requiredToolName) {
			return fmt.Errorf("case collection required tool not registered: %s", requiredToolName)
		}
	}

	systemPrompt := buildCaseCollectionSystemPrompt(a.systemPrompt)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: buildCaseCollectionUserPrompt(trimmedQuery, caseCount),
		},
	}

	createdCount := 0
	maxRounds := caseCollectionMaxRounds(caseCount)
	searchSummaries := make([]string, 0, caseCount)
	uploadPhasePrepared := false

	for round := 0; round < maxRounds; round++ {
		if round == caseCount && !uploadPhasePrepared {
			messages = buildCaseCollectionUploadPhaseMessages(systemPrompt, trimmedQuery, caseCount, searchSummaries)
			uploadPhasePrepared = true
		}

		forcedToolName := caseCollectionRoundToolName(round, caseCount)
		roundTools, err := caseCollectionToolsForTool(forcedToolName)
		if err != nil {
			return err
		}
		action := fmt.Sprintf("create case collection chat completion round %d", round+1)
		fmt.Printf("[CaseCollectionAgent][Round %d] request model=%s messages=%d target=%d exposed_tool=%s query=%s\n",
			round+1, strings.TrimSpace(a.modelID), len(messages), caseCount, forcedToolName, truncateForLog(trimmedQuery, 120))

		var resp openai.ChatCompletionResponse
		if err := a.Retry(action, func() error {
			var callErr error
			resp, callErr = a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       a.modelID,
				Messages:    messages,
				Tools:       roundTools,
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
				return fmt.Errorf("model returned no tool calls while %s was required", forcedToolName)
			}
			return fmt.Errorf("case collection stopped early: created %d/%d pending review cases while %s was required",
				createdCount, caseCount, forcedToolName)
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

			if strings.TrimSpace(call.Function.Name) != forcedToolName {
				appendToolResponse(call.ID, map[string]interface{}{
					"status": "failed",
					"error":  fmt.Sprintf("unexpected tool for round %d: required %s, got %s", round+1, forcedToolName, strings.TrimSpace(call.Function.Name)),
				})
				continue
			}

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
			if call.Function.Name == tool.WebSearchToolName {
				if summary := buildCaseCollectionSearchSummary(len(searchSummaries)+1, response.Payload); summary != "" {
					searchSummaries = append(searchSummaries, summary)
				}
			}
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
			maxRounds, createdCount, caseCount)
	}
	return fmt.Errorf("case collection exceeded max tool rounds (%d) without pending review case", maxRounds)
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
	currentTime := time.Now().Format("2006-01-02 15:04:05 -07:00 MST")
	return fmt.Sprintf(
		"请围绕以下主题搜集诈骗案件，并提交 %d 个待审核案件。\n\n当前时间：%s\n要求：优先选择案发时间或发布时间明确、且时间尽量接近当前时间的最新公开案件。\n流程约束：你最多可进行 %d 轮 search_web 搜索；完成搜索后，请逐条调用 upload_historical_case_to_vector_db，直到提交满目标数量。\n搜索主题：%s",
		caseCount,
		currentTime,
		caseCount,
		strings.TrimSpace(query),
	)
}

func buildCaseCollectionUploadPhaseMessages(systemPrompt string, query string, caseCount int, searchSummaries []string) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: strings.TrimSpace(systemPrompt),
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: buildCaseCollectionUploadPhaseUserPrompt(query, caseCount, searchSummaries),
		},
	}
}

func buildCaseCollectionUploadPhaseUserPrompt(query string, caseCount int, searchSummaries []string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05 -07:00 MST")
	compressedSummary := "未提取到有效的 search_web 摘要，请谨慎整理案件。"
	if len(searchSummaries) > 0 {
		compressedSummary = strings.Join(searchSummaries, "\n\n")
	}

	return fmt.Sprintf(
		"请围绕以下主题继续整理诈骗案件，并提交 %d 个待审核案件。\n\n当前时间：%s\n搜索主题：%s\n阶段说明：前置 search_web 阶段已经完成，共 %d 轮。不要继续搜索，只能基于下面的压缩搜索结果整理案件，并逐条调用 upload_historical_case_to_vector_db。\n压缩搜索结果：\n%s",
		caseCount,
		currentTime,
		strings.TrimSpace(query),
		caseCount,
		compressedSummary,
	)
}

func caseCollectionMaxRounds(caseCount int) int {
	return caseCount * 2
}

func caseCollectionRoundToolName(round int, caseCount int) string {
	if round < caseCount {
		return tool.WebSearchToolName
	}
	return tool.UploadHistoricalCaseToVectorDBToolName
}

func caseCollectionToolsForTool(name string) ([]openai.Tool, error) {
	trimmedName := strings.TrimSpace(name)
	for _, registeredTool := range caseCollectionToolsProvider() {
		if registeredTool.Function == nil {
			continue
		}
		if strings.TrimSpace(registeredTool.Function.Name) == trimmedName {
			return []openai.Tool{registeredTool}, nil
		}
	}
	return nil, fmt.Errorf("case collection required tool not registered: %s", trimmedName)
}

type caseCollectionSearchSummaryPayload struct {
	Status  string                              `json:"status"`
	Query   string                              `json:"query"`
	Answer  string                              `json:"answer"`
	Results []caseCollectionSearchSummaryResult `json:"results"`
}

type caseCollectionSearchSummaryResult struct {
	Title         string `json:"title"`
	URL           string `json:"url"`
	Content       string `json:"content"`
	PublishedDate string `json:"published_date"`
}

func buildCaseCollectionSearchSummary(index int, payload map[string]interface{}) string {
	if len(payload) == 0 {
		return ""
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	var normalized caseCollectionSearchSummaryPayload
	if err := json.Unmarshal(encoded, &normalized); err != nil {
		return ""
	}

	if status := strings.TrimSpace(normalized.Status); status != "" && status != "success" {
		return ""
	}

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("[search_web #%d]\n", index))
	if query := strings.TrimSpace(normalized.Query); query != "" {
		builder.WriteString("query: ")
		builder.WriteString(query)
		builder.WriteString("\n")
	}
	if answer := truncateForCaseCollectionPrompt(normalized.Answer, 220); answer != "" {
		builder.WriteString("answer: ")
		builder.WriteString(answer)
		builder.WriteString("\n")
	}

	for resultIndex, result := range normalized.Results {
		builder.WriteString(fmt.Sprintf("%d. 标题：%s\n", resultIndex+1, firstNonEmptyForLog(result.Title, "未命名结果")))
		if publishedDate := strings.TrimSpace(result.PublishedDate); publishedDate != "" {
			builder.WriteString("   时间：")
			builder.WriteString(publishedDate)
			builder.WriteString("\n")
		}
		if url := strings.TrimSpace(result.URL); url != "" {
			builder.WriteString("   链接：")
			builder.WriteString(url)
			builder.WriteString("\n")
		}
		if content := truncateForCaseCollectionPrompt(result.Content, 260); content != "" {
			builder.WriteString("   摘要：")
			builder.WriteString(content)
			builder.WriteString("\n")
		}
	}

	return strings.TrimSpace(builder.String())
}

func truncateForCaseCollectionPrompt(input string, maxLen int) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return ""
	}
	if maxLen <= 3 {
		return trimmed
	}
	runes := []rune(trimmed)
	if len(runes) <= maxLen {
		return trimmed
	}
	return string(runes[:maxLen-3]) + "..."
}

func caseCollectionHasTool(name string) bool {
	trimmedName := strings.TrimSpace(name)
	for _, registeredTool := range caseCollectionToolsProvider() {
		if registeredTool.Function == nil {
			continue
		}
		if strings.TrimSpace(registeredTool.Function.Name) == trimmedName {
			return true
		}
	}
	return false
}
