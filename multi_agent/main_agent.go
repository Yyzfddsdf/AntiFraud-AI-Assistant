package multi_agent

import (
	"antifraud/config"
	"antifraud/multi_agent/state"
	"antifraud/multi_agent/tool"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	openai "antifraud/llm"
)

type modalityResult struct {
	Text          string
	Image         string
	Video         string
	Audio         string
	ImageInsights []string
	VideoInsights []string
	AudioInsights []string
}

// MainAgent 负责聚合多模态子智能体结果并驱动工具调用闭环。
type MainAgent struct {
	CommonAgent
	client       *openai.Client
	modelID      string
	systemPrompt string
}

// NewMainAgent 按配置创建主智能体实例。
func NewMainAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *MainAgent {
	common := NewCommonAgent("MainAgent", modelCfg, retryCfg)

	return &MainAgent{
		CommonAgent: common,
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  common.APIKey,
			BaseURL: common.BaseURL,
		}),
		modelID:      strings.TrimSpace(modelCfg.Model),
		systemPrompt: strings.TrimSpace(systemPrompt),
	}
}

// AnalyzeMainReport 提供默认用户上下文的主流程入口。
func AnalyzeMainReport(text string, videosBase64 []string, audiosBase64 []string, imagesBase64 []string) (string, error) {
	return AnalyzeMainReportForUser("demo-user", "", text, videosBase64, audiosBase64, imagesBase64)
}

// AnalyzeMainReportForUser 是主流程入口：
// 并行执行子模态分析 -> 组装主输入 -> 调用主智能体输出最终报告。
func AnalyzeMainReportForUser(userID string, taskID string, text string, videosBase64 []string, audiosBase64 []string, imagesBase64 []string) (string, error) {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return "", fmt.Errorf("load main config failed: %w", err)
	}
	mainAgent := NewMainAgent(cfg.Agents.Main, cfg.Retry, cfg.Prompts.Main)

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}
	trimmedTaskID := strings.TrimSpace(taskID)
	trimmedText := strings.TrimSpace(text)

	videosBase64 = normalizeBase64List(videosBase64)
	audiosBase64 = normalizeBase64List(audiosBase64)
	imagesBase64 = normalizeBase64List(imagesBase64)
	fmt.Printf("[MainAgent] start analyze: user=%s task=%s text_len=%d video=%d audio=%d image=%d\n",
		trimmedUserID, firstNonEmptyForLog(trimmedTaskID, "<empty>"), len(trimmedText), len(videosBase64), len(audiosBase64), len(imagesBase64))

	results := modalityResult{
		Text:  trimmedText,
		Image: "No image input provided.",
		Video: "No video input provided.",
		Audio: "No audio input provided.",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if len(imagesBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("[MainAgent] image sub-agent start, count=%d\n", len(imagesBase64))
			parallelResults := AnalyzeImagesParallel(imagesBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Image = "Image analysis failed: empty result"
				results.ImageInsights = []string{"Image analysis failed: empty result"}
				return
			}
			results.Image = formatModalityBatchResult("Image", parallelResults)
			results.ImageInsights = append([]string{}, parallelResults...)
			fmt.Printf("[MainAgent] image sub-agent done, result_count=%d\n", len(parallelResults))
		}()
	}

	if len(videosBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("[MainAgent] video sub-agent start, count=%d\n", len(videosBase64))
			parallelResults := AnalyzeVideosParallel(videosBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Video = "Video analysis failed: empty result"
				results.VideoInsights = []string{"Video analysis failed: empty result"}
				return
			}
			results.Video = formatModalityBatchResult("Video", parallelResults)
			results.VideoInsights = append([]string{}, parallelResults...)
			fmt.Printf("[MainAgent] video sub-agent done, result_count=%d\n", len(parallelResults))
		}()
	}

	if len(audiosBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("[MainAgent] audio sub-agent start, count=%d\n", len(audiosBase64))
			parallelResults := AnalyzeAudiosParallel(audiosBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Audio = "Audio analysis failed: empty result"
				results.AudioInsights = []string{"Audio analysis failed: empty result"}
				return
			}
			results.Audio = formatModalityBatchResult("Audio", parallelResults)
			results.AudioInsights = append([]string{}, parallelResults...)
			fmt.Printf("[MainAgent] audio sub-agent done, result_count=%d\n", len(parallelResults))
		}()
	}

	wg.Wait()
	fmt.Printf("[MainAgent] sub-agents complete: image_insights=%d video_insights=%d audio_insights=%d\n",
		len(results.ImageInsights), len(results.VideoInsights), len(results.AudioInsights))

	if trimmedTaskID != "" {
		state.UpdateTaskInsights(trimmedUserID, trimmedTaskID, results.VideoInsights, results.AudioInsights, results.ImageInsights)
	}

	finalInput := buildMainAgentInput(results)
	// 外层先写入子模态洞察（insights）。
	// generateReport 入口会继续补齐 user/task/payload，确保工具上下文完整。
	ctx := context.Background()
	ctx = tool.BindTaskInsights(ctx, results.VideoInsights, results.AudioInsights, results.ImageInsights)

	report, err := mainAgent.generateReport(ctx, finalInput, trimmedUserID, trimmedTaskID, trimmedText, videosBase64, audiosBase64, imagesBase64)
	if err != nil {
		return "", err
	}
	fmt.Printf("[MainAgent] final report generated, len=%d\n", len(strings.TrimSpace(report)))
	return report, nil
}

// normalizeBase64List 过滤空输入并保留有效 Base64 项。
func normalizeBase64List(items []string) []string {
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

// formatModalityBatchResult 将同一模态的并行结果拼接为统一文本块。
func formatModalityBatchResult(modality string, results []string) string {
	if len(results) == 0 {
		return fmt.Sprintf("%s analysis failed: empty result", modality)
	}

	var builder strings.Builder
	for i, result := range results {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("[%s #%d]\n", modality, i+1))
		builder.WriteString(strings.TrimSpace(result))
	}
	return strings.TrimSpace(builder.String())
}

// buildMainAgentInput 构建主智能体用户输入载荷。
func buildMainAgentInput(results modalityResult) string {
	textInput := results.Text
	if textInput == "" {
		textInput = "No text input provided."
	}

	return fmt.Sprintf(
		"[User Text]\n%s\n\n[Image Insights]\n%s\n\n[Video Insights]\n%s\n\n[Audio Insights]\n%s",
		textInput,
		results.Image,
		results.Video,
		results.Audio,
	)
}

// generateReport 驱动主智能体工具调用循环，直到拿到终态报告。
func (a *MainAgent) generateReport(ctx context.Context, finalInput string, userID string, taskID string, rawText string, rawVideos []string, rawAudios []string, rawImages []string) (string, error) {
	// 工具执行依赖的关键上下文在这里统一绑定：
	// - user_id / task_id：用于用户查询与任务级归档定位
	// - 原始 payload（text/videos/audios/images）：用于落库保留原始输入
	ctx = tool.BindUserID(ctx, userID)
	ctx = tool.BindTaskID(ctx, taskID)
	ctx = tool.BindTaskPayload(ctx, rawText, rawVideos, rawAudios, rawImages)

	if a.client == nil {
		return "", fmt.Errorf("main agent client is not initialized")
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: a.systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: finalInput,
		},
	}

	const maxRounds = 8
	var finalResult string
	finalReportSubmitted := false
	historyCaseWritten := false

	for round := 0; round < maxRounds; round++ {
		action := fmt.Sprintf("create chat completion round %d", round+1)
		fmt.Printf("[MainAgent][Round %d] request model=%s messages=%d\n", round+1, strings.TrimSpace(a.modelID), len(messages))
		var resp openai.ChatCompletionResponse
		if err := a.Retry(action, func() error {
			var callErr error
			resp, callErr = a.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       a.modelID,
				Messages:    messages,
				Tools:       tool.MainAgentTools(),
				ToolChoice:  "required",
				Stream:      false,
				MaxTokens:   a.MaxTokens,
				Temperature: float32(a.Temperature),
				TopP:        float32(a.TopP),
			})
			return callErr
		}); err != nil {
			return "", fmt.Errorf("main agent api error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("main agent returned empty choices")
		}

		msg := resp.Choices[0].Message
		fmt.Printf("[MainAgent][Round %d] ai_reply=%s\n", round+1, truncateForLog(msg.Content, 320))
		if len(msg.ToolCalls) > 0 {
			fmt.Printf("[MainAgent][Round %d] tool_calls=%d\n", round+1, len(msg.ToolCalls))
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		if len(msg.ToolCalls) == 0 {
			if finalReportSubmitted && historyCaseWritten && finalResult != "" {
				fmt.Printf("[MainAgent][Round %d] finish: final report + history done\n", round+1)
				return finalResult, nil
			}
			if finalResult != "" {
				fmt.Printf("[MainAgent][Round %d] finish: return cached final result\n", round+1)
				return finalResult, nil
			}
			if msg.Content != "" {
				fmt.Printf("[MainAgent][Round %d] finish: no tool call, return ai content\n", round+1)
				return msg.Content, nil
			}
			return "", fmt.Errorf("model returned no tool calls and no content")
		}

		toolResponseAdded := false
		appendToolResponse := func(callID string, payload map[string]interface{}) {
			toolPayload, _ := json.Marshal(payload)
			fmt.Printf("[MainAgent][Round %d] tool_result call_id=%s payload=%s\n", round+1, strings.TrimSpace(callID), truncateForLog(string(toolPayload), 360))
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: callID,
				Content:    string(toolPayload),
			})
			toolResponseAdded = true
		}

		for _, call := range msg.ToolCalls {
			fmt.Printf("[MainAgent][Round %d] tool_call name=%s args=%s\n", round+1, call.Function.Name, truncateForLog(call.Function.Arguments, 240))
			if call.Function.Name == tool.FinalReportToolName {
				finalReportSubmitted = true
				fmt.Printf("[MainAgent][Round %d] mark submit_final_report=true\n", round+1)
			}
			if call.Function.Name == tool.WriteUserHistoryCaseToolName {
				historyCaseWritten = true
				fmt.Printf("[MainAgent][Round %d] mark write_user_history_case=true\n", round+1)
			}

			handler := tool.GetToolHandler(call.Function.Name)
			if handler == nil {
				appendToolResponse(call.ID, map[string]interface{}{"error": "unsupported tool"})
				fmt.Printf("[MainAgent][Round %d] unsupported tool: %s\n", round+1, call.Function.Name)
				continue
			}

			response, err := handler.Handle(ctx, call.Function.Arguments)
			if err != nil {
				appendToolResponse(call.ID, map[string]interface{}{"error": err.Error()})
				fmt.Printf("[MainAgent][Round %d] tool handler error: %v\n", round+1, err)
				continue
			}

			appendToolResponse(call.ID, response.Payload)
			if response.FinalResultStr != "" {
				finalResult = response.FinalResultStr
				// submit_final_report 返回最终报告后，写入 ctx，
				// 供 write_user_history_case 在归档时读取并持久化。
				ctx = tool.BindFinalReport(ctx, finalResult)
				fmt.Printf("[MainAgent][Round %d] final_result updated, len=%d\n", round+1, len(strings.TrimSpace(finalResult)))
			}
		}

		if !toolResponseAdded {
			return "", fmt.Errorf("tool calls returned but no tool response was added")
		}

		if finalReportSubmitted && historyCaseWritten && finalResult != "" {
			fmt.Printf("[MainAgent][Round %d] finish: final report archived\n", round+1)
			return finalResult, nil
		}
	}

	if finalResult != "" {
		fmt.Printf("[MainAgent] max rounds reached, return existing final result len=%d\n", len(strings.TrimSpace(finalResult)))
		return finalResult, nil
	}
	return "", fmt.Errorf("main agent exceeded max tool rounds (%d) without final result", maxRounds)
}

func truncateForLog(input string, maxLen int) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "<empty>"
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

func firstNonEmptyForLog(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
