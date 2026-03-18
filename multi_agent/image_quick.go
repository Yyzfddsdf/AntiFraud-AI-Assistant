package multi_agent

import (
	"context"
	"fmt"
	"strings"

	"antifraud/config"
	openai "antifraud/llm"
	"antifraud/multi_agent/tool"
)

type ImageQuickRiskResponse struct {
	RiskLevel string `json:"risk_level"`
	Reason    string `json:"reason"`
}

type ImageQuickAgent struct {
	SubAgentBase
}

func NewImageQuickAgent(modelCfg config.ModelConfig, retryCfg config.RetryConfig, systemPrompt string) *ImageQuickAgent {
	profile := SubAgentProfile{
		Modality:     "image",
		SystemPrompt: strings.TrimSpace(systemPrompt),
		UserPrompt:   "请快速识别这张图片的风险等级，并通过指定工具提交标准化结果。",
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

	return &ImageQuickAgent{
		SubAgentBase: NewSubAgentBase("ImageQuickAgent", modelCfg, retryCfg, profile),
	}
}

func (a *ImageQuickAgent) AnalyzeQuick(ctx context.Context, imageBase64 string) (ImageQuickRiskResponse, error) {
	if a.profile.BuildDataURL == nil || a.profile.BuildDataPart == nil {
		return ImageQuickRiskResponse{}, fmt.Errorf("image quick agent profile is not configured")
	}

	dataURL, err := a.profile.BuildDataURL(imageBase64)
	if err != nil {
		return ImageQuickRiskResponse{}, err
	}

	req := openai.ChatCompletionRequest{
		Model:       a.modelID,
		MaxTokens:   a.MaxTokens,
		Temperature: float32(a.Temperature),
		TopP:        float32(a.TopP),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: a.profile.SystemPrompt,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{Type: "text", Text: a.profile.UserPrompt},
					a.profile.BuildDataPart(dataURL),
				},
			},
		},
		Stream:     false,
		Tools:      []openai.Tool{tool.ImageQuickRiskTool},
		ToolChoice: "required",
	}

	action := "create chat completion for image quick risk"
	var resp openai.ChatCompletionResponse
	if err := a.Retry(action, func() error {
		var callErr error
		resp, callErr = a.client.CreateChatCompletion(ctx, req)
		return callErr
	}); err != nil {
		return ImageQuickRiskResponse{}, fmt.Errorf("image quick api error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return ImageQuickRiskResponse{}, fmt.Errorf("image quick returned empty choices")
	}

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) == 0 {
		return ImageQuickRiskResponse{}, fmt.Errorf("image quick returned no tool call")
	}

	result, err := tool.ParseImageQuickRiskResult(msg.ToolCalls[0].Function.Arguments)
	if err != nil {
		return ImageQuickRiskResponse{}, fmt.Errorf("parse image quick tool result failed: %w", err)
	}

	return ImageQuickRiskResponse{
		RiskLevel: result.RiskLevel,
		Reason:    result.Reason,
	}, nil
}

func AnalyzeImageQuick(imageBase64 string) (ImageQuickRiskResponse, error) {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		return ImageQuickRiskResponse{}, fmt.Errorf("load image quick config failed: %w", err)
	}

	agent := NewImageQuickAgent(cfg.Agents.ImageQuick, cfg.Retry, cfg.Prompts.ImageQuick)
	return agent.AnalyzeQuick(context.Background(), imageBase64)
}
