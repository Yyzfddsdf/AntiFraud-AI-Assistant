package multi_agent

import (
	"antifraud/internal/platform/config"
	"context"
	"fmt"
	"strings"
	"sync"

	openai "antifraud/internal/platform/llm"
)

func buildASRAudioDataURL(audioBase64 string) (string, error) {
	trimmed := strings.TrimSpace(audioBase64)
	if strings.HasPrefix(trimmed, "data:") {
		if strings.Contains(trimmed, ";base64,") {
			return trimmed, nil
		}
		return "", fmt.Errorf("invalid asr audio data url")
	}
	if trimmed == "" {
		return "", fmt.Errorf("empty asr audio base64")
	}
	return fmt.Sprintf("data:audio/mpeg;base64,%s", trimmed), nil
}

type ASRAgent struct {
	CommonAgent
	client  *openai.Client
	modelID string
}

func NewASRAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig) *ASRAgent {
	common := NewCommonAgent("ASRAgent", modelCfg, retryCfg)
	return &ASRAgent{
		CommonAgent: common,
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  common.APIKey,
			BaseURL: common.BaseURL,
		}),
		modelID: strings.TrimSpace(modelCfg.Model),
	}
}

func (a *ASRAgent) Transcribe(ctx context.Context, audioBase64 string, index int) (string, error) {
	displayIndex := index + 1

	dataURL, err := buildASRAudioDataURL(audioBase64)
	if err != nil {
		return "", fmt.Errorf("audio %d: %w", displayIndex, err)
	}

	parts := []openai.ChatMessagePart{{
		Type: "input_audio",
		ExtraFields: map[string]interface{}{
			"input_audio": map[string]interface{}{
				"data": dataURL,
			},
		},
	}}

	messages := []openai.ChatCompletionMessage{{
		Role:         openai.ChatMessageRoleUser,
		MultiContent: parts,
	}}

	req := openai.ChatCompletionRequest{
		Model:       a.modelID,
		Messages:    messages,
		MaxTokens:   a.MaxTokens,
		Temperature: float32(a.Temperature),
		TopP:        float32(a.TopP),
		Stream:      false,
	}
	req.SetField("modalities", []string{"text"})
	req.SetField("asr_options", map[string]interface{}{
		"enable_itn": false,
	})

	action := fmt.Sprintf("create chat completion for asr audio %d", displayIndex)
	var resp openai.ChatCompletionResponse
	if err := a.Retry(action, func() error {
		var callErr error
		resp, callErr = a.client.CreateChatCompletion(ctx, req)
		return callErr
	}); err != nil {
		return "", fmt.Errorf("audio %d: ASR API error: %w", displayIndex, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("audio %d: no choices in ASR response", displayIndex)
	}

	transcript := strings.TrimSpace(extractMessageText(resp.Choices[0].Message))
	if transcript == "" {
		return "", fmt.Errorf("audio %d: empty ASR transcript", displayIndex)
	}
	return transcript, nil
}

func (a *ASRAgent) TranscribeBatchInParallel(ctx context.Context, inputs []string) []string {
	var wg sync.WaitGroup
	results := make([]string, len(inputs))

	for i, item := range inputs {
		wg.Add(1)
		go func(index int, input string) {
			defer wg.Done()
			fmt.Printf("[%s] starting transcription for audio %d...\n", a.Name(), index+1)
			res, err := a.Transcribe(ctx, input, index)
			if err != nil {
				results[index] = fmt.Sprintf("Error: %v", err)
			} else {
				results[index] = res
			}
			fmt.Printf("[%s] finished transcription for audio %d\n", a.Name(), index+1)
		}(i, item)
	}

	wg.Wait()
	return results
}

func extractMessageText(msg openai.ChatCompletionMessage) string {
	if strings.TrimSpace(msg.Content) != "" {
		return strings.TrimSpace(msg.Content)
	}

	if len(msg.MultiContent) == 0 {
		return ""
	}

	texts := make([]string, 0, len(msg.MultiContent))
	for _, part := range msg.MultiContent {
		if strings.TrimSpace(part.Text) != "" {
			texts = append(texts, strings.TrimSpace(part.Text))
		}
	}
	return strings.TrimSpace(strings.Join(texts, "\n"))
}

func TranscribeAudiosParallel(audiosBase64 []string) []string {
	cfg, err := config.LoadConfig("internal/platform/config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	ctx := context.Background()
	agent := NewASRAgent(cfg.Agents.ASR, cfg.Retry)
	return agent.TranscribeBatchInParallel(ctx, audiosBase64)
}
