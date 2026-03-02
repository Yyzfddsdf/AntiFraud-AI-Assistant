package multi_agent

import (
	"context"
	"fmt"
	"image_recognition/config"
	"image_recognition/multi_agent/tool"
	"strings"
	"sync"
	"time"

	openai "image_recognition/llm"
)

// CommonAgent 保存所有智能体共享的模型参数与重试配置。
type CommonAgent struct {
	name        string
	APIKey      string
	BaseURL     string
	MaxTokens   int
	TopP        float64
	Temperature float64
	RetryMax    int
	RetryDelay  time.Duration
}

func NewCommonAgent(name string, modelCfg config.ModelConfig, retryCfg config.RetryConfig) CommonAgent {
	return CommonAgent{
		name:        name,
		APIKey:      modelCfg.APIKey,
		BaseURL:     modelCfg.BaseURL,
		MaxTokens:   modelCfg.MaxTokens,
		TopP:        modelCfg.TopP,
		Temperature: modelCfg.Temperature,
		RetryMax:    retryCfg.MaxRetries,
		RetryDelay:  time.Duration(retryCfg.RetryDelayMS) * time.Millisecond,
	}
}

func (a CommonAgent) Name() string {
	return a.name
}

// Retry 使用线性退避执行重试。
func (a CommonAgent) Retry(action string, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= a.RetryMax; attempt++ {
		if err := fn(); err == nil {
			if attempt > 1 {
				fmt.Printf("[%s] retry succeeded: action=%s, attempt=%d\n", a.Name(), action, attempt)
			}
			return nil
		} else {
			lastErr = err
			fmt.Printf("[%s] call failed: action=%s, attempt=%d/%d, err=%v\n", a.Name(), action, attempt, a.RetryMax, err)
		}

		if attempt < a.RetryMax {
			backoff := time.Duration(attempt) * a.RetryDelay
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", action, a.RetryMax, lastErr)
}

// SubAgentProfile 定义模态子智能体的请求构造参数。
type SubAgentProfile struct {
	Modality      string
	SystemPrompt  string
	UserPrompt    string
	BuildDataURL  func(raw string) (string, error)
	BuildDataPart func(dataURL string) openai.ChatMessagePart
	RequestFields map[string]interface{}
}

// SubAgentBase 是图像/视频/音频子智能体的通用执行基类。
type SubAgentBase struct {
	CommonAgent
	client  *openai.Client
	modelID string
	profile SubAgentProfile
}

func NewSubAgentBase(name string, modelCfg config.ModelConfig, retryCfg config.RetryConfig, profile SubAgentProfile) SubAgentBase {
	modality := strings.TrimSpace(profile.Modality)
	if modality == "" {
		modality = "input"
	}
	profile.Modality = modality
	profile.SystemPrompt = strings.TrimSpace(profile.SystemPrompt)
	profile.UserPrompt = strings.TrimSpace(profile.UserPrompt)
	if profile.RequestFields == nil {
		profile.RequestFields = map[string]interface{}{}
	}

	return SubAgentBase{
		CommonAgent: NewCommonAgent(name, modelCfg, retryCfg),
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  modelCfg.APIKey,
			BaseURL: modelCfg.BaseURL,
		}),
		modelID: modelCfg.Model,
		profile: profile,
	}
}

func (a *SubAgentBase) Analyze(ctx context.Context, dataBase64 string, index int) (string, error) {
	displayIndex := index + 1

	if a.profile.BuildDataURL == nil {
		return "", fmt.Errorf("%s %d: BuildDataURL is not configured", a.profile.Modality, displayIndex)
	}
	if a.profile.BuildDataPart == nil {
		return "", fmt.Errorf("%s %d: BuildDataPart is not configured", a.profile.Modality, displayIndex)
	}

	dataURL, err := a.profile.BuildDataURL(dataBase64)
	if err != nil {
		return "", fmt.Errorf("%s %d: %w", a.profile.Modality, displayIndex, err)
	}

	dataPart := a.profile.BuildDataPart(dataURL)
	if strings.TrimSpace(dataPart.Type) == "" {
		return "", fmt.Errorf("%s %d: data part type is empty", a.profile.Modality, displayIndex)
	}

	parts := make([]openai.ChatMessagePart, 0, 2)
	if a.profile.UserPrompt != "" {
		parts = append(parts, openai.ChatMessagePart{Type: "text", Text: a.profile.UserPrompt})
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

	return a.AnalyzeWithUnifiedTool(ctx, a.client, req, a.profile.Modality, index)
}

func (a *SubAgentBase) AnalyzeBatchInParallel(ctx context.Context, inputs []string) []string {
	var wg sync.WaitGroup
	results := make([]string, len(inputs))

	for i, item := range inputs {
		wg.Add(1)
		go func(index int, input string) {
			defer wg.Done()
			fmt.Printf("[%s] starting analysis for %s %d...\n", a.Name(), a.profile.Modality, index+1)
			res, err := a.Analyze(ctx, input, index)
			if err != nil {
				results[index] = fmt.Sprintf("Error: %v", err)
			} else {
				results[index] = res
			}
			fmt.Printf("[%s] finished analysis for %s %d\n", a.Name(), a.profile.Modality, index+1)
		}(i, item)
	}

	wg.Wait()
	return results
}

// AnalyzeWithUnifiedTool 统一封装重试、模型调用、工具解析与输出格式化流程。
func (a *SubAgentBase) AnalyzeWithUnifiedTool(ctx context.Context, client *openai.Client, req openai.ChatCompletionRequest, modality string, index int) (string, error) {
	displayIndex := index + 1
	action := fmt.Sprintf("create chat completion for %s %d", modality, displayIndex)

	var resp openai.ChatCompletionResponse
	if err := a.Retry(action, func() error {
		var callErr error
		resp, callErr = client.CreateChatCompletion(ctx, req)
		return callErr
	}); err != nil {
		return "", fmt.Errorf("%s %d: API error: %w", modality, displayIndex, err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("%s %d: no choices in response", modality, displayIndex)
	}

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) > 0 {
		result, err := tool.ParseAnalysisResult(msg.ToolCalls[0].Function.Arguments)
		if err != nil {
			return "", fmt.Errorf("%s %d: parse tool call error: %w", modality, displayIndex, err)
		}
		return tool.FormatAnalysisResult(result), nil
	}

	return msg.Content, nil
}
