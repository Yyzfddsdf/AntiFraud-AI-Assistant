package multi_agent

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image_recognition/config"
	"image_recognition/multi_agent/tool"
	"io"
	"net/http"
	"os"
	"strings"
)

const VideoSystemPrompt = `你是一位精通视频内容风控的AI专家。你的任务是全方位分析视频的视觉画面与行为逻辑，识别潜在的诈骗、博彩或非法违规风险。

请重点关注以下维度：
1. **画面真实性判定**：首先明确视频内容是“真实拍摄”、“游戏录屏/CG动画”还是“手机/电脑屏幕翻拍”。对于高拟真的游戏画面，需仔细甄别其物理光影和人物动作的自然度。
2. **视觉呈现**：是否存在高饱和度色彩、夸张的动态特效、满屏的弹窗广告或模仿知名应用的伪造界面。
3. **内容逻辑**：视频内容是否包含诱导性承诺（如“高额回报”、“立即提现”）、紧迫感制造（如倒计时、限时优惠）或展示虚假的高消费生活/大量现金。
4. **关键信息**：提取视频中出现的文字、网址、联系方式及特定的引导性话术。

请严格按照以下格式输出分析结果（不要包含任何开场白或结束语）：

1. **视频摘要与场景描述**：
   - 开篇明确指出画面性质（如：FPS游戏录屏、真人出镜拍摄、手机屏幕翻拍）。
   - 概括视频的核心内容、场景设定及整体风格（如：粗制滥造的营销视频、伪装的教程演示）。

2. **关键视觉特征（主观与客观）**：
   - 描述识别到的高风险视觉元素（颜色、特效、布局）。
   - 列出提取的关键文字信息（APP名、网址、电话、金额等）。

3. **可疑点清单（客观列举）**：
   - 逐条列出视频中不符合常理或具有欺诈嫌疑的特征（如：承诺不切实际的收益、索要敏感权限/信息、界面与正版应用存在明显差异）。
   - 若未发现明显异常，请注明“未发现明显视觉风险特征”。`

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
	Tools       []Tool    `json:"tools,omitempty"`
	ToolChoice  string    `json:"tool_choice,omitempty"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type ContentPart struct {
	Type     string      `json:"type"`
	Text     string      `json:"text,omitempty"`
	ImageURL *FileObject `json:"image_url,omitempty"`
	VideoURL *FileObject `json:"video_url,omitempty"`
}

type FileObject struct {
	URL string `json:"url"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content   string          `json:"content"`
			ToolCalls []tool.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
}

type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewClient(apiKey, baseURL string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	url := c.BaseURL + "/chat/completions"
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}

func AnalyzeVideo(ctx context.Context, client *Client, modelID string, videoBase64 string, index int) (string, error) {
	dataURL, err := buildVideoDataURL(videoBase64)
	if err != nil {
		return "", fmt.Errorf("video %d: %w", index, err)
	}

	req := ChatCompletionRequest{
		Model: modelID,
		Messages: []Message{
			{
				Role:    "system",
				Content: VideoSystemPrompt,
			},
			{
				Role: "user",
				Content: []ContentPart{
					{
						Type: "text",
						Text: "请分析这个视频的内容",
					},
					{
						Type: "video_url",
						VideoURL: &FileObject{
							URL: dataURL,
						},
					},
				},
			},
		},
		MaxTokens:   1024,
		TopP:        1.0,
		Temperature: 0.5,
		Stream:      false,
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
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

	action := fmt.Sprintf("create chat completion for video %d", index+1)
	resp, err := callWithRetry[*ChatCompletionResponse]("VideoAgent", action, func() (*ChatCompletionResponse, error) {
		return client.CreateChatCompletion(ctx, req)
	})
	if err != nil {
		return "", fmt.Errorf("video %d: API error: %w", index, err)
	}

	if len(resp.Choices) > 0 {
		msg := resp.Choices[0].Message
		if len(msg.ToolCalls) > 0 {
			result, err := tool.ParseAnalysisResult(msg.ToolCalls[0].Function.Arguments)
			if err != nil {
				return "", fmt.Errorf("video %d: parse tool call error: %w", index, err)
			}
			return tool.FormatAnalysisResult(result), nil
		}
		return msg.Content, nil
	}
	return "", fmt.Errorf("video %d: no content returned", index)
}

type videoAnalyzer struct {
	client  *Client
	modelID string
}

func (a videoAnalyzer) Analyze(ctx context.Context, dataBase64 string, index int) (string, error) {
	return AnalyzeVideo(ctx, a.client, a.modelID, dataBase64, index)
}

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

func AnalyzeVideosParallel(videosBase64 []string) []string {
	cfg, err := config.LoadConfig("config/config.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	client := NewClient(cfg.APIKey, cfg.BaseURL)
	ctx := context.Background()

	analyzer := videoAnalyzer{client: client, modelID: cfg.ImageModel}
	return analyzeBatchInParallel(ctx, analyzer, videosBase64, "video")
}

func RunVideoCLI() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . video <path_to_video1> [path_to_video2 ...]")
		return
	}

	videoPaths := os.Args[1:]
	var videosBase64 []string

	for _, path := range videoPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading video file %s: %v\n", path, err)
			os.Exit(1)
		}
		videosBase64 = append(videosBase64, base64.StdEncoding.EncodeToString(data))
	}

	fmt.Printf("Processing %d videos in parallel...\n", len(videoPaths))
	results := AnalyzeVideosParallel(videosBase64)

	fmt.Println("\n--- Results ---")
	for i, res := range results {
		fmt.Printf("\n[Video %s]:\n%s\n", videoPaths[i], res)
	}
}
