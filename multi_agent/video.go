package multi_agent

import (
	"context"
	"fmt"
	"image_recognition/config"
	"strings"

	openai "image_recognition/llm"
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

func AnalyzeVideosParallel(videosBase64 []string) []string {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	ctx := context.Background()
	agent := NewVideoAgent(cfg.Agents.Video, cfg.Retry, cfg.Prompts.Video)
	return agent.AnalyzeBatchInParallel(ctx, videosBase64)
}
