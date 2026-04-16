package embedding

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	appcfg "antifraud/internal/platform/config"
	openai "antifraud/internal/platform/llm"
)

const defaultConfigPath = "internal/platform/config/config.json"

// GenerateVector 调用统一 embeddings 接口，并返回单条输入的向量与模型名。
func GenerateVector(ctx context.Context, inputText string) ([]float64, string, error) {
	trimmedInput := strings.TrimSpace(inputText)
	if trimmedInput == "" {
		return nil, "", fmt.Errorf("input text is empty")
	}

	cfg, err := appcfg.LoadConfig(defaultConfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("load config failed: %w", err)
	}

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  cfg.Embedding.APIKey,
		BaseURL: cfg.Embedding.BaseURL,
	})

	req := openai.EmbeddingRequest{
		Model:          cfg.Embedding.Model,
		Input:          []string{trimmedInput},
		EncodingFormat: "float",
	}
	req.SetField("truncate", "NONE")

	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}

	resp, err := createEmbeddingsWithRetry(callCtx, client, cfg, req)
	if err != nil {
		return nil, "", err
	}
	if len(resp.Data) == 0 {
		return nil, "", fmt.Errorf("embedding response is empty")
	}

	sort.Slice(resp.Data, func(i, j int) bool {
		return resp.Data[i].Index < resp.Data[j].Index
	})

	vector := append([]float64{}, resp.Data[0].Embedding...)
	if len(vector) == 0 {
		return nil, "", fmt.Errorf("embedding vector is empty")
	}

	modelName := strings.TrimSpace(resp.Model)
	if modelName == "" {
		modelName = strings.TrimSpace(cfg.Embedding.Model)
	}

	return vector, modelName, nil
}

func createEmbeddingsWithRetry(ctx context.Context, client *openai.Client, cfg *appcfg.Config, req openai.EmbeddingRequest) (openai.EmbeddingResponse, error) {
	maxRetries := 1
	retryDelay := time.Duration(0)
	if cfg != nil {
		if cfg.Retry.MaxRetries > 0 {
			maxRetries = cfg.Retry.MaxRetries
		}
		if cfg.Retry.RetryDelayMS > 0 {
			retryDelay = time.Duration(cfg.Retry.RetryDelayMS) * time.Millisecond
		}
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.CreateEmbeddings(ctx, req)
		if err == nil {
			if attempt > 1 {
				fmt.Printf("[Embedding] retry succeeded: attempt=%d\n", attempt)
			}
			return resp, nil
		}

		lastErr = err
		fmt.Printf("[Embedding] call failed: attempt=%d/%d err=%v\n", attempt, maxRetries, err)
		if attempt >= maxRetries || !isRetryableEmbeddingError(err) {
			break
		}

		backoff := time.Duration(attempt) * retryDelay
		if backoff > 0 {
			select {
			case <-ctx.Done():
				return openai.EmbeddingResponse{}, fmt.Errorf("create embedding failed: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}
	}

	return openai.EmbeddingResponse{}, fmt.Errorf("create embedding failed after %d attempts: %w", maxRetries, lastErr)
}

func isRetryableEmbeddingError(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	if message == "" {
		return false
	}

	if strings.Contains(message, "context deadline exceeded") ||
		strings.Contains(message, "timeout") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "connection refused") ||
		strings.Contains(message, "temporarily unavailable") ||
		strings.Contains(message, "eof") {
		return true
	}

	statusMarker := "status="
	index := strings.Index(message, statusMarker)
	if index < 0 {
		return false
	}

	statusStart := index + len(statusMarker)
	statusEnd := statusStart
	for statusEnd < len(message) && message[statusEnd] >= '0' && message[statusEnd] <= '9' {
		statusEnd++
	}
	if statusEnd == statusStart {
		return false
	}

	statusCode, parseErr := strconv.Atoi(message[statusStart:statusEnd])
	if parseErr != nil {
		return false
	}

	if statusCode == 408 || statusCode == 409 || statusCode == 429 {
		return true
	}
	return statusCode >= 500
}
