package multi_agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"antifraud/config"
	"net/http"
	"strings"

	openai "antifraud/llm"
)

func buildImageDataURL(imageBase64 string) (string, error) {
	trimmed := strings.TrimSpace(imageBase64)
	if strings.HasPrefix(trimmed, "data:") {
		if strings.Contains(trimmed, ";base64,") {
			return trimmed, nil
		}
		return "", fmt.Errorf("invalid image data url")
	}
	if trimmed == "" {
		return "", fmt.Errorf("empty image base64")
	}
	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid image base64: %w", err)
	}
	mimeType := http.DetectContentType(raw)
	if mimeType == "application/octet-stream" {
		mimeType = "image/jpeg"
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, encoded), nil
}

type ImageAgent struct {
	SubAgentBase
}

func NewImageAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *ImageAgent {
	profile := SubAgentProfile{
		Modality:     "image",
		SystemPrompt: strings.TrimSpace(systemPrompt),
		UserPrompt:   "Analyze this image and extract risk-related evidence.",
		BuildDataURL: buildImageDataURL,
		BuildDataPart: func(dataURL string) openai.ChatMessagePart {
			return openai.ChatMessagePart{
				Type: "image_url",
				ImageURL: &openai.ChatMessageImageURL{
					URL: dataURL,
				},
			}
		},
	}

	return &ImageAgent{
		SubAgentBase: NewSubAgentBase("ImageAgent", modelCfg, retryCfg, profile),
	}
}

func AnalyzeImagesParallel(imagesBase64 []string) []string {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	ctx := context.Background()
	agent := NewImageAgent(cfg.Agents.Image, cfg.Retry, cfg.Prompts.Image)
	return agent.AnalyzeBatchInParallel(ctx, imagesBase64)
}
