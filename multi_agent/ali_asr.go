package multi_agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"antifraud/config"
	"strings"

	openai "antifraud/llm"
)

func buildAudioDataURL(audioBase64 string) (string, error) {
	trimmed := strings.TrimSpace(audioBase64)
	if strings.HasPrefix(trimmed, "data:") {
		if strings.Contains(trimmed, ";base64,") {
			return trimmed, nil
		}
		return "", fmt.Errorf("invalid audio data url")
	}
	if trimmed == "" {
		return "", fmt.Errorf("empty audio base64")
	}
	raw, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", fmt.Errorf("invalid audio base64: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	return fmt.Sprintf("data:;base64,%s", encoded), nil
}

type AudioAgent struct {
	SubAgentBase
}

func NewAudioAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *AudioAgent {
	profile := SubAgentProfile{
		Modality:     "audio",
		SystemPrompt: strings.TrimSpace(systemPrompt),
		UserPrompt:   "Analyze this audio and extract key risk evidence.",
		BuildDataURL: buildAudioDataURL,
		BuildDataPart: func(dataURL string) openai.ChatMessagePart {
			return openai.ChatMessagePart{
				Type: "input_audio",
				ExtraFields: map[string]interface{}{
					"input_audio": map[string]interface{}{
						"data":   dataURL,
						"format": "mp3",
					},
				},
			}
		},
		RequestFields: map[string]interface{}{
			"modalities": []string{"text"},
		},
	}

	return &AudioAgent{
		SubAgentBase: NewSubAgentBase("AudioAgent", modelCfg, retryCfg, profile),
	}
}

func AnalyzeAudiosParallel(audiosBase64 []string) []string {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	ctx := context.Background()
	agent := NewAudioAgent(cfg.Agents.Audio, cfg.Retry, cfg.Prompts.Audio)
	return agent.AnalyzeBatchInParallel(ctx, audiosBase64)
}
