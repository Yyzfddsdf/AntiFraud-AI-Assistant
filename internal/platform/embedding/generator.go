package embedding

import (
	"context"
	"fmt"
	"sort"
	"strings"

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

	resp, err := client.CreateEmbeddings(callCtx, req)
	if err != nil {
		return nil, "", fmt.Errorf("create embedding failed: %w", err)
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
