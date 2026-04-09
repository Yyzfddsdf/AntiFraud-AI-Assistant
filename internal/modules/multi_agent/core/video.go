package multi_agent

import (
	"antifraud/internal/modules/multi_agent/adapters/outbound/tool"
	"antifraud/internal/platform/config"
	"context"
	"fmt"
	"strings"
	"sync"

	openai "antifraud/internal/platform/llm"
)

func buildVideoDataURL(videoBase64 string) (string, error) {
	trimmed := strings.TrimSpace(videoBase64)
	if strings.HasPrefix(trimmed, "data:") {
		if strings.Contains(trimmed, ";base64,") {
			return trimmed, nil
		}
		return "", fmt.Errorf("invalid video data url")
	}
	if trimmed == "" {
		return "", fmt.Errorf("empty video base64")
	}
	return fmt.Sprintf("data:video/mp4;base64,%s", trimmed), nil
}

type VideoAgent struct {
	SubAgentBase
}

func NewVideoAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *VideoAgent {
	profile := SubAgentProfile{
		Modality:     "video",
		SystemPrompt: strings.TrimSpace(systemPrompt),
		UserPrompt:   "Analyze this video and extract risk-related evidence.",
		BuildDataURL: buildVideoDataURL,
		BuildDataPart: func(dataURL string) openai.ChatMessagePart {
			return openai.ChatMessagePart{
				Type: "video_url",
				VideoURL: &openai.ChatMessageVideoURL{
					URL: dataURL,
				},
			}
		},
	}

	return &VideoAgent{
		SubAgentBase: NewSubAgentBase("VideoAgent", modelCfg, retryCfg, profile),
	}
}

func (a *VideoAgent) AnalyzeWithTranscript(ctx context.Context, videoBase64 string, transcript string, index int) (string, error) {
	displayIndex := index + 1

	if a.profile.BuildDataURL == nil {
		return "", fmt.Errorf("video %d: BuildDataURL is not configured", displayIndex)
	}
	if a.profile.BuildDataPart == nil {
		return "", fmt.Errorf("video %d: BuildDataPart is not configured", displayIndex)
	}

	dataURL, err := a.profile.BuildDataURL(videoBase64)
	if err != nil {
		return "", fmt.Errorf("video %d: %w", displayIndex, err)
	}

	dataPart := a.profile.BuildDataPart(dataURL)
	if strings.TrimSpace(dataPart.Type) == "" {
		return "", fmt.Errorf("video %d: data part type is empty", displayIndex)
	}

	userPrompt := strings.TrimSpace(a.profile.UserPrompt)
	trimmedTranscript := strings.TrimSpace(transcript)
	if trimmedTranscript != "" {
		if userPrompt != "" {
			userPrompt += "\n\n"
		}
		userPrompt += "以下是该视频音轨的 ASR 转写文本，可能存在少量识别误差。请将其作为辅助证据，与视频画面联合判断，不要脱离画面单独下结论：\n" + trimmedTranscript
	}

	parts := make([]openai.ChatMessagePart, 0, 2)
	if userPrompt != "" {
		parts = append(parts, openai.ChatMessagePart{Type: "text", Text: userPrompt})
	}
	parts = append(parts, dataPart)

	systemPrompt := a.profile.SystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are a fraud-risk analysis assistant."
	}

	req := openai.ChatCompletionRequest{
		Model:       a.modelID,
		MaxTokens:   a.MaxTokens,
		Temperature: float32(a.Temperature),
		TopP:        float32(a.TopP),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:         openai.ChatMessageRoleUser,
				MultiContent: parts,
			},
		},
		Stream:     false,
		Tools:      []openai.Tool{tool.AnalysisTool},
		ToolChoice: "required",
	}

	for key, value := range a.profile.RequestFields {
		req.SetField(key, value)
	}

	result, err := a.AnalyzeWithUnifiedTool(ctx, a.client, req, a.profile.Modality, index)
	if err != nil {
		return "", err
	}
	if trimmedTranscript == "" {
		return result, nil
	}
	return strings.TrimSpace(result + "\n\n【视频音轨ASR转写】\n" + trimmedTranscript), nil
}

func AnalyzeVideosParallel(videosBase64 []string) []string {
	cfg, err := config.LoadConfig("internal/platform/config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	ctx := context.Background()
	videoAgent := NewVideoAgent(cfg.Agents.Video, cfg.Retry, cfg.Prompts.Video)
	asrAgent := NewASRAgent(cfg.Agents.ASR, cfg.Retry)

	var wg sync.WaitGroup
	results := make([]string, len(videosBase64))

	for i, item := range videosBase64 {
		wg.Add(1)
		go func(index int, input string) {
			defer wg.Done()
			fmt.Printf("[VideoAgent] starting analysis for video %d...\n", index+1)

			videoForAnalysis, prepareErr := compressVideoForAnalysis(input)
			if prepareErr != nil {
				results[index] = fmt.Sprintf("Error: prepare video failed: %v", prepareErr)
				fmt.Printf("[VideoAgent] video prepare failed for video %d: %v\n", index+1, prepareErr)
				return
			}

			transcript, transcriptErr := transcribeVideoAudio(ctx, asrAgent, input, index)
			if transcriptErr != nil {
				fmt.Printf("[VideoAgent] asr skipped for video %d: %v\n", index+1, transcriptErr)
			}

			result, err := videoAgent.AnalyzeWithTranscript(ctx, videoForAnalysis, transcript, index)
			if err != nil {
				if strings.TrimSpace(transcript) != "" {
					results[index] = strings.TrimSpace(fmt.Sprintf("Video analysis failed: %v\n\n【视频音轨ASR转写】\n%s", err, transcript))
				} else {
					results[index] = fmt.Sprintf("Error: %v", err)
				}
			} else {
				results[index] = result
			}

			fmt.Printf("[VideoAgent] finished analysis for video %d\n", index+1)
		}(i, item)
	}

	wg.Wait()
	return results
}

func transcribeVideoAudio(ctx context.Context, asrAgent *ASRAgent, videoBase64 string, index int) (string, error) {
	if asrAgent == nil {
		return "", fmt.Errorf("ASR agent is nil")
	}

	audioData, err := extractAudioTrackForASR(videoBase64)
	if err != nil {
		return "", fmt.Errorf("extract audio track failed: %w", err)
	}

	transcript, err := asrAgent.Transcribe(ctx, audioData, index)
	if err != nil {
		return "", fmt.Errorf("transcribe audio failed: %w", err)
	}

	return strings.TrimSpace(transcript), nil
}
