package multi_agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image_recognition/config"
	"image_recognition/multi_agent/state"
	"image_recognition/multi_agent/tool"
	"os"
	"strings"
	"sync"

	"github.com/sashabaranov/go-openai"
)

const mainAgentSystemPrompt = `你是一位多模态风控总分析专家。你将接收：
1) 用户提供的文本描述；
2) 图像子智能体分析结果；
3) 视频子智能体分析结果；
4) 音频子智能体分析结果。

你的任务是：先完成检索与用户信息补全，提交结构化最终报告，并在报告后补充写入用户案件历史。

【强制工具调用规则】
1. 第一阶段（必须）：
	- 必须先调用 search_similar_cases。
	- query 必须是完整案件描述，不得过短，需包含：可疑行为、话术特征、关键实体（金额/联系方式/平台名/账号）、场景线索。
2. 第二阶段（按需但强烈建议）：
	- 优先调用：
	  a) query_user_info
	  b) query_user_history_cases
	- 这三个用户相关工具均不需要输入 user_id，user_id 由服务端 HTTP 上下文自动获取（当前为占位实现）。
3. 第三阶段（必须）：
	- 在调用 submit_final_report 之后，必须调用 write_user_history_case 写入案件摘要（补充用户案件历史）。
4. 结束阶段（必须最后一步）：
	- 只能在完成至少一次 search_similar_cases 后，调用 submit_final_report。
	- submit_final_report 后必须再调用 write_user_history_case，完成写入后流程才结束。
	- 禁止在 submit_final_report 之前直接给自然语言结论。

【输出约束】
- 仅输出中文；
- 你不应直接输出正文报告，最终必须通过 submit_final_report 返回；
- 不得编造未出现的事实；
- 若某模态未提供，相关字段应明确“未提供该模态数据”。

【submit_final_report字段要求】
- summary: 综合摘要
- text_finding/image_finding/video_finding/audio_finding: 各模态关键发现
- risk_signals: 风险信号清单（数组）
- risk_level: 低/中/高
- risk_reason: 风险等级理由
- next_actions: 建议的下一步核查动作（数组）`

type modalityResult struct {
	Text          string
	Image         string
	Video         string
	Audio         string
	ImageInsights []string
	VideoInsights []string
	AudioInsights []string
}

// AnalyzeMainReport 封装主分析智能体。
// 入参共四个：文本、视频base64列表、音频base64列表、图片base64列表（后3者可为空列表）。
func AnalyzeMainReport(text string, videosBase64 []string, audiosBase64 []string, imagesBase64 []string) (string, error) {
	return AnalyzeMainReportForUser("demo-user", "", text, videosBase64, audiosBase64, imagesBase64)
}

// AnalyzeMainReportForUser 封装主分析智能体（带用户上下文）。
func AnalyzeMainReportForUser(userID string, taskID string, text string, videosBase64 []string, audiosBase64 []string, imagesBase64 []string) (string, error) {
	mainCfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return "", fmt.Errorf("load main config failed: %w", err)
	}

	videosBase64 = normalizeBase64List(videosBase64)
	audiosBase64 = normalizeBase64List(audiosBase64)
	imagesBase64 = normalizeBase64List(imagesBase64)

	results := modalityResult{
		Text:  strings.TrimSpace(text),
		Image: "未提供该模态数据",
		Video: "未提供该模态数据",
		Audio: "未提供该模态数据",
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	if len(imagesBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			parallelResults := AnalyzeImagesParallel(imagesBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Image = "图像分析失败: 未返回结果"
				results.ImageInsights = []string{"图像分析失败: 未返回结果"}
				return
			}
			results.Image = formatModalityBatchResult("图像", parallelResults)
			results.ImageInsights = append([]string{}, parallelResults...)
		}()
	}

	if len(videosBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			parallelResults := AnalyzeVideosParallel(videosBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Video = "视频分析失败: 未返回结果"
				results.VideoInsights = []string{"视频分析失败: 未返回结果"}
				return
			}
			results.Video = formatModalityBatchResult("视频", parallelResults)
			results.VideoInsights = append([]string{}, parallelResults...)
		}()
	}

	if len(audiosBase64) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			parallelResults := AnalyzeAudiosParallel(audiosBase64)
			mu.Lock()
			defer mu.Unlock()
			if len(parallelResults) == 0 {
				results.Audio = "音频分析失败: 未返回结果"
				results.AudioInsights = []string{"音频分析失败: 未返回结果"}
				return
			}
			results.Audio = formatModalityBatchResult("音频", parallelResults)
			results.AudioInsights = append([]string{}, parallelResults...)
		}()
	}

	wg.Wait()

	if strings.TrimSpace(taskID) != "" {
		state.UpdateTaskInsights(strings.TrimSpace(userID), strings.TrimSpace(taskID), results.VideoInsights, results.AudioInsights, results.ImageInsights)
	}

	finalInput := buildMainAgentInput(results)
	ctx := context.Background()
	ctx = tool.BindTaskInsights(ctx, results.VideoInsights, results.AudioInsights, results.ImageInsights)
	report, err := generateMainReport(ctx, mainCfg, finalInput, strings.TrimSpace(userID), strings.TrimSpace(taskID), strings.TrimSpace(text), videosBase64, audiosBase64, imagesBase64)
	if err != nil {
		return "", err
	}
	return report, nil
}

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

func formatModalityBatchResult(modality string, results []string) string {
	if len(results) == 0 {
		return fmt.Sprintf("%s分析失败: 未返回结果", modality)
	}

	var builder strings.Builder
	for i, result := range results {
		if i > 0 {
			builder.WriteString("\n\n")
		}
		builder.WriteString(fmt.Sprintf("【%s #%d】\n", modality, i+1))
		builder.WriteString(strings.TrimSpace(result))
	}
	return strings.TrimSpace(builder.String())
}

func buildMainAgentInput(results modalityResult) string {
	textInput := results.Text
	if textInput == "" {
		textInput = "未提供文本说明"
	}

	return fmt.Sprintf(
		"【用户文本输入】\n%s\n\n【图像子智能体结果】\n%s\n\n【视频子智能体结果】\n%s\n\n【音频子智能体结果】\n%s",
		textInput,
		results.Image,
		results.Video,
		results.Audio,
	)
}

func generateMainReport(ctx context.Context, cfg *config.Config, finalInput string, userID string, taskID string, rawText string, rawVideos []string, rawAudios []string, rawImages []string) (string, error) {
	ctx = tool.BindUserID(ctx, userID)
	ctx = tool.BindTaskID(ctx, taskID)
	ctx = tool.BindTaskPayload(ctx, rawText, rawVideos, rawAudios, rawImages)

	openaiCfg := openai.DefaultConfig(cfg.APIKey)
	openaiCfg.BaseURL = cfg.BaseURL
	client := openai.NewClientWithConfig(openaiCfg)

	modelID := cfg.MainModel
	if strings.TrimSpace(modelID) == "" {
		modelID = "z-ai/glm-5"
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: mainAgentSystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: finalInput,
		},
	}

	const maxRounds = 8
	hasCaseSearch := false
	var finalReportPayload *tool.FinalReportPayload
	hasHistoryWriteAfterFinal := false
	for round := 0; round < maxRounds; round++ {
		fmt.Printf("[MainAgent][Round %d] 开始请求模型\n", round+1)
		action := fmt.Sprintf("create chat completion round %d", round+1)
		resp, err := callWithRetry[openai.ChatCompletionResponse]("MainAgent", action, func() (openai.ChatCompletionResponse, error) {
			return client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
				Model:       modelID,
				Messages:    messages,
				Tools:       tool.MainAgentTools(),
				ToolChoice:  "required",
				Stream:      false,
				MaxTokens:   2048,
				Temperature: 0.3,
				TopP:        1.0,
			})
		})
		if err != nil {
			return "", fmt.Errorf("main agent api error: %w", err)
		}

		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("main agent returned empty choices")
		}

		msg := resp.Choices[0].Message
		fmt.Printf("[MainAgent][Round %d] 模型回复: %s\n", round+1, truncateForLog(msg.Content, 240))
		if len(msg.ToolCalls) > 0 {
			fmt.Printf("[MainAgent][Round %d] 触发工具数: %d\n", round+1, len(msg.ToolCalls))
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		if len(msg.ToolCalls) == 0 {
			if finalReportPayload != nil && !hasHistoryWriteAfterFinal {
				return "", fmt.Errorf("final report already submitted, but write_user_history_case was not called after report")
			}
			fallback := tool.FinalReportPayload{
				Summary:      strings.TrimSpace(msg.Content),
				TextFinding:  "未提供结构化结果",
				ImageFinding: "未提供结构化结果",
				VideoFinding: "未提供结构化结果",
				AudioFinding: "未提供结构化结果",
				RiskSignals:  []string{"模型未返回工具调用，已降级为文本输出"},
				RiskLevel:    "中",
				RiskReason:   "返回格式不符合工具约束",
				NextActions:  []string{"重试请求并检查模型工具调用兼容性"},
			}
			return tool.FormatFinalReport(fallback), nil
		}

		toolResponseAdded := false
		appendToolResponse := func(callID string, payload map[string]interface{}) {
			toolPayload, _ := json.Marshal(payload)
			fmt.Printf("[MainAgent][Round %d] 工具返回: %s\n", round+1, truncateForLog(string(toolPayload), 320))
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: callID,
				Content:    string(toolPayload),
			})
			toolResponseAdded = true
		}
		for _, call := range msg.ToolCalls {
			fmt.Printf("[MainAgent][Round %d] 调用工具: %s, 参数: %s\n", round+1, call.Function.Name, truncateForLog(call.Function.Arguments, 240))
			handler := tool.GetToolHandler(call.Function.Name)
			if handler == nil {
				appendToolResponse(call.ID, map[string]interface{}{"error": "unsupported tool"})
				fmt.Printf("[MainAgent][Round %d] 未支持工具: %s\n", round+1, call.Function.Name)
				continue
			}

			// 特殊检查 for final report
			if call.Function.Name == tool.FinalReportToolName && !hasCaseSearch {
				appendToolResponse(call.ID, map[string]interface{}{
					"error": "必须先调用 search_similar_cases 再调用 submit_final_report",
				})
				fmt.Printf("[MainAgent][Round %d] 拦截最终报告: 尚未完成案件检索\n", round+1)
				continue
			}

			response, err := handler.Handle(ctx, call.Function.Arguments)
			if err != nil {
				appendToolResponse(call.ID, map[string]interface{}{"error": err.Error()})
				continue
			}

			appendToolResponse(call.ID, response.Payload)

			// 处理标志
			if response.SetCaseSearch {
				hasCaseSearch = true
			}
			if response.SetFinalReport {
				finalReportPayload = response.FinalReportPayload
				ctx = tool.BindFinalReport(ctx, tool.FormatFinalReport(*response.FinalReportPayload))
				hasHistoryWriteAfterFinal = false
				fmt.Printf("[MainAgent][Round %d] 收到最终报告工具，等待写入用户历史后结束\n", round+1)
			}
			if response.SetHistoryWriteAfterFinal {
				hasHistoryWriteAfterFinal = true
				fmt.Printf("[MainAgent][Round %d] 已在最终报告后完成历史写入，流程可结束\n", round+1)
			}
		}

		if !toolResponseAdded {
			return "", fmt.Errorf("tool calls returned but no tool response was added")
		}
		if finalReportPayload != nil && hasHistoryWriteAfterFinal {
			return tool.FormatFinalReport(*finalReportPayload), nil
		}
		fmt.Printf("[MainAgent][Round %d] 工具结果已回填，进入下一轮\n", round+1)
	}

	return "", fmt.Errorf("main agent exceeded max tool rounds (%d) without final report", maxRounds)
}

func truncateForLog(input string, maxLen int) string {
	text := strings.TrimSpace(input)
	if text == "" {
		return "<empty>"
	}
	if maxLen <= 3 {
		return text
	}
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen-3]) + "..."
}

func RunMainAgentCLI() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . multimodal <text> [video_paths] [audio_paths] [image_paths]")
		fmt.Println("Tip: 每个模态支持多个文件，使用逗号分隔，例如: video1.mp4,video2.mp4")
		return
	}

	text := os.Args[1]
	var videosBase64, audiosBase64, imagesBase64 []string

	if len(os.Args) >= 3 && strings.TrimSpace(os.Args[2]) != "" {
		paths := parsePathListArg(os.Args[2])
		converted, err := encodeFilesToBase64(paths)
		if err != nil {
			fmt.Printf("Error reading video files: %v\n", err)
			os.Exit(1)
		}
		videosBase64 = converted
	}

	if len(os.Args) >= 4 && strings.TrimSpace(os.Args[3]) != "" {
		paths := parsePathListArg(os.Args[3])
		converted, err := encodeFilesToBase64(paths)
		if err != nil {
			fmt.Printf("Error reading audio files: %v\n", err)
			os.Exit(1)
		}
		audiosBase64 = converted
	}

	if len(os.Args) >= 5 && strings.TrimSpace(os.Args[4]) != "" {
		paths := parsePathListArg(os.Args[4])
		converted, err := encodeFilesToBase64(paths)
		if err != nil {
			fmt.Printf("Error reading image files: %v\n", err)
			os.Exit(1)
		}
		imagesBase64 = converted
	}

	report, err := AnalyzeMainReport(text, videosBase64, audiosBase64, imagesBase64)
	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n--- Multimodal Analysis Report ---")
	fmt.Println(report)
}

func parsePathListArg(arg string) []string {
	parts := strings.Split(arg, ",")
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		paths = append(paths, trimmed)
	}
	return paths
}

func encodeFilesToBase64(paths []string) ([]string, error) {
	results := make([]string, 0, len(paths))
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read file %s failed: %w", path, err)
		}
		results = append(results, base64.StdEncoding.EncodeToString(data))
	}
	return results, nil
}
