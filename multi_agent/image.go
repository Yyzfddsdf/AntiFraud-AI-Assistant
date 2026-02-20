package multi_agent

import (
	"context"
	"encoding/base64"
	"fmt"
	"image_recognition/config"
	"image_recognition/multi_agent/tool"
	"net/http"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// API Configuration
const (
	SystemPrompt = `你是一位精通视觉风控的AI专家。你的核心任务是深入分析图像内容，精准识别其中可能存在的诈骗、博彩或非法违规特征，并提取关键的客观信息。

请遵循以下分析逻辑：
1. **画面性质判定**：首先明确区分图片是“现实拍摄”（Real World Photography）、“屏幕翻拍”（Screen Photograph）、“数字合成/游戏画面”（Digital/Game Render）还是“UI界面截图”。特别注意区分逼真的游戏画面与真实场景。
2. **全局视觉扫描**：评估图片的整体设计风格、配色方案及排版布局，判断是否具有高风险网站/应用的典型视觉特征（如高饱和度色彩冲击、杂乱的弹窗/悬浮窗、粗糙的模仿痕迹）。
3. **关键要素提取**：仔细识别并提取图片中的文字信息（如APP名称、URL、金额、联系方式、机构名称）及核心场景元素。
4. **风险特征排查**：重点检测是否存在诱导性内容（如“点击领取”、“稳赚不赔”、“美女荷官”）、紧迫感营造（如倒计时、名额限制）或其他可疑的社会工程学套路。

**重要执行要求**：
- 必须调用 'submit_analysis_result' 工具提交你的分析结果。
- 在 'visual_impression' 中必须包含对画面性质（现实/游戏/屏幕翻拍）的明确描述。
- 严禁直接输出文本分析，所有结果必须通过工具参数传递。
- 在“可疑点清单”中，请客观列出观察到的异常特征，无需给出最终定性结论（如“这是诈骗”），保持客观中立。`
)

// RecognizeImage processes a single image and returns the result string.
// It is designed to be called concurrently.
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

func AnalyzeImage(ctx context.Context, client *openai.Client, modelID string, imageBase64 string, index int) (string, error) {
	dataURL, err := buildImageDataURL(imageBase64)
	if err != nil {
		return "", fmt.Errorf("image %d: %w", index, err)
	}

	// 3. Create Request
	req := openai.ChatCompletionRequest{
		Model: modelID,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: SystemPrompt,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: openai.ChatMessagePartTypeText,
						Text: "提取图片中的文字、场景和可疑视觉特征",
					},
					{
						Type: openai.ChatMessagePartTypeImageURL,
						ImageURL: &openai.ChatMessageImageURL{
							URL: dataURL,
						},
					},
				},
			},
		},
		Stream:      false,
		MaxTokens:   1024,
		TopP:        1.0,
		Temperature: 0.5,
		Tools: []openai.Tool{
			{
				Type: "function",
				Function: &openai.FunctionDefinition{
					Name:        tool.AnalysisToolName,
					Description: "提交分析结果，包含整体视觉感受、关键内容提取和可疑点清单",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"visual_impression": map[string]interface{}{
								"type":        "string",
								"description": "整体视觉感受（主观特征）：描述整体风格、高风险视觉特征",
							},
							"key_content": map[string]interface{}{
								"type":        "string",
								"description": "关键内容提取（客观信息）：提取文字信息和核心场景描述",
							},
							"suspicious_points": map[string]interface{}{
								"type":        "array",
								"items":       map[string]string{"type": "string"},
								"description": "可疑点清单",
							},
						},
						"required": []string{"visual_impression", "key_content", "suspicious_points"},
					},
				},
			},
		},
		ToolChoice: "required",
	}

	action := fmt.Sprintf("create chat completion for image %d", index+1)
	resp, err := callWithRetry[openai.ChatCompletionResponse]("ImageAgent", action, func() (openai.ChatCompletionResponse, error) {
		return client.CreateChatCompletion(ctx, req)
	})
	if err != nil {
		return "", fmt.Errorf("image %d: API error: %w", index, err)
	}

	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		if len(msg.ToolCalls) > 0 {
			result, err := tool.ParseAnalysisResult(msg.ToolCalls[0].Function.Arguments)
			if err != nil {
				return "", fmt.Errorf("image %d: parse tool call error: %w", index, err)
			}
			return tool.FormatAnalysisResult(result), nil
		}
		return msg.Content, nil
	}
	return "", fmt.Errorf("image %d: no content returned", index)
}

type imageAnalyzer struct {
	client  *openai.Client
	modelID string
}

func (a imageAnalyzer) Analyze(ctx context.Context, dataBase64 string, index int) (string, error) {
	return AnalyzeImage(ctx, a.client, a.modelID, dataBase64, index)
}

// RecognizeImagesParallel takes a slice of image data and processes them concurrently.
// It returns a slice of results in the same order as the input.
func AnalyzeImagesParallel(imagesBase64 []string) []string {
	// Load config
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	// Configure Client once
	openaiConfig := openai.DefaultConfig(cfg.APIKey)
	openaiConfig.BaseURL = cfg.BaseURL
	client := openai.NewClientWithConfig(openaiConfig)
	ctx := context.Background()

	analyzer := imageAnalyzer{client: client, modelID: cfg.ImageModel}
	return analyzeBatchInParallel(ctx, analyzer, imagesBase64, "image")
}

func RunImageCLI() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . image <path_to_image1> [path_to_image2 ...]")
		return
	}

	imagePaths := os.Args[1:]
	var imagesBase64 []string

	// Read all images first
	for _, path := range imagePaths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			os.Exit(1)
		}
		imagesBase64 = append(imagesBase64, base64.StdEncoding.EncodeToString(data))
	}

	fmt.Printf("Processing %d images in parallel...\n", len(imagesBase64))
	results := AnalyzeImagesParallel(imagesBase64)

	fmt.Println("\n--- Results ---")
	for i, res := range results {
		fmt.Printf("\n[Image %s]:\n%s\n", imagePaths[i], res)
	}
}
