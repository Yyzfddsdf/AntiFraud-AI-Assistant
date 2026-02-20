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

const AudioSystemPrompt = `你是一位精通语音风控的AI专家。你的任务是深度分析音频内容，通过语调、话术模式及关键词识别，捕捉潜在的诈骗、博彩或非法违规风险。

请重点关注以下维度：
1. **声音来源判定**：首先辨别声音是“自然人声”还是“AI合成/机械音”。重点关注语调的自然度、停顿呼吸感以及是否存在电子合成痕迹。
2. **话术分析**：是否存在典型的诈骗脚本特征，如“内幕消息”、“安全账户”、“低风险高回报”、“公检法办案”等。
3. **语态与情绪**：说话人是否刻意营造紧迫感（催促行动）、恐吓感（威胁后果）或过度热情（诱导信任）。
4. **环境背景**：背景音是否异常（如伪造的办公环境音、嘈杂的呼叫中心声）。

请严格按照以下格式输出分析结果（不要包含任何开场白或结束语）：

1. **音频摘要与场景描述**：
   - 开篇明确指出声音性质（如：自然人声、AI合成语音、机械录音）。
   - 概括音频的主要话题及对话场景（如：推销理财、冒充客服、虚假办案）。
   - 描述说话人的语气特征（如：机械读稿、强硬恐吓、急促焦虑）。

2. **关键信息提取（客观事实）**：
   - 提取音频中提及的关键实体信息（人名、机构名、银行账号、电话号码、网址、金额）。
   - 记录要求受害者执行的关键动作或时间限制。

3. **可疑点清单（客观列举）**：
   - 逐条列出音频中符合诈骗/违规套路的话术或逻辑漏洞（如：索要验证码/密码、要求私下转账、以保密为由禁止挂断电话）。
   - 若未发现明显异常，请注明“未发现明显语音风险特征”。`

type AliChatCompletionRequest struct {
	Model         string            `json:"model"`
	Messages      []AliMessage      `json:"messages"`
	Modalities    []string          `json:"modalities,omitempty"`
	Audio         *AliAudioOutput   `json:"audio,omitempty"`
	Stream        bool              `json:"stream,omitempty"`
	StreamOptions *AliStreamOptions `json:"stream_options,omitempty"`
	Tools         []AliTool         `json:"tools,omitempty"`
	ToolChoice    string            `json:"tool_choice,omitempty"`
}

type AliTool struct {
	Type     string          `json:"type"`
	Function AliToolFunction `json:"function"`
}

type AliToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type AliMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type AliContentPart struct {
	Type       string         `json:"type"`
	Text       string         `json:"text,omitempty"`
	InputAudio *AliInputAudio `json:"input_audio,omitempty"`
}

type AliInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format,omitempty"`
}

type AliAudioOutput struct {
	Voice  string `json:"voice,omitempty"`
	Format string `json:"format,omitempty"`
}

type AliStreamOptions struct {
	IncludeUsage bool `json:"include_usage,omitempty"`
}

type AliChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content   string          `json:"content"`
			ToolCalls []tool.ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
	} `json:"choices"`
	Usage interface{} `json:"usage,omitempty"`
}

type AliClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewAliClient(apiKey, baseURL string) *AliClient {
	return &AliClient{
		APIKey:     apiKey,
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *AliClient) CreateChatCompletion(ctx context.Context, req AliChatCompletionRequest) (string, error) {
	url := c.BaseURL + "/chat/completions"
	req.Stream = false
	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var completion AliChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	msg := completion.Choices[0].Message
	if len(msg.ToolCalls) > 0 {
		result, err := tool.ParseAnalysisResult(msg.ToolCalls[0].Function.Arguments)
		if err != nil {
			return "", fmt.Errorf("parse tool call error: %w", err)
		}
		// Use custom formatter for audio
		return FormatAudioAnalysisResult(result), nil
	}
	return msg.Content, nil
}

func FormatAudioAnalysisResult(result tool.AnalysisResult) string {
	output := "【音频摘要与场景描述】\n" + result.VisualImpression + "\n\n"
	output += "【关键信息提取（客观信息）】\n" + result.KeyContent + "\n\n"
	output += "【可疑点清单（仅列出，不判断）】\n"
	if len(result.SuspiciousPoints) == 0 {
		output += "- 未发现明显语音异常\n"
	} else {
		for i, point := range result.SuspiciousPoints {
			output += fmt.Sprintf("%d. %s\n", i+1, point)
		}
	}
	return output
}

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

func AnalyzeAudio(ctx context.Context, client *AliClient, modelID string, audioBase64 string, index int) (string, error) {
	dataURL, err := buildAudioDataURL(audioBase64)
	if err != nil {
		return "", fmt.Errorf("audio %d: %w", index, err)
	}

	req := AliChatCompletionRequest{
		Model: modelID,
		Messages: []AliMessage{
			{
				Role:    "system",
				Content: AudioSystemPrompt,
			},
			{
				Role: "user",
				Content: []AliContentPart{
					{
						Type: "input_audio",
						InputAudio: &AliInputAudio{
							Data:   dataURL,
							Format: "mp3",
						},
					},
					{
						Type: "text",
						Text: "请分析这段音频的内容，提取关键信息并指出可疑之处。",
					},
				},
			},
		},
		Modalities: []string{"text"},
		Stream:     false,
		Tools: []AliTool{
			{
				Type: "function",
				Function: AliToolFunction{
					Name:        tool.AnalysisToolName,
					Description: "提交分析结果，包含音频摘要、关键内容提取和可疑点清单",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"visual_impression": map[string]interface{}{
								"type":        "string",
								"description": "音频摘要与场景描述：声音来源判定（真人/AI）、主要话题、语气特征",
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

	action := fmt.Sprintf("create chat completion for audio %d", index+1)
	resp, err := callWithRetry[string]("AudioAgent", action, func() (string, error) {
		return client.CreateChatCompletion(ctx, req)
	})
	if err != nil {
		return "", fmt.Errorf("audio %d: API error: %w", index, err)
	}
	return resp, nil
}

type audioAnalyzer struct {
	client  *AliClient
	modelID string
}

func (a audioAnalyzer) Analyze(ctx context.Context, dataBase64 string, index int) (string, error) {
	return AnalyzeAudio(ctx, a.client, a.modelID, dataBase64, index)
}

func AnalyzeAudiosParallel(audiosBase64 []string) []string {
	cfg, err := config.LoadConfig("config/config_ali.json")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return []string{fmt.Sprintf("Error loading config: %v", err)}
	}

	apiKey := cfg.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("DASHSCOPE_API_KEY")
	}
	if apiKey == "" {
		return []string{"Error: API Key not found in config or environment variables."}
	}

	modelID := cfg.AudioModel
	if strings.TrimSpace(modelID) == "" {
		return []string{"Error: audio_model is empty in config/config_ali.json"}
	}

	client := NewAliClient(apiKey, cfg.BaseURL)
	ctx := context.Background()

	analyzer := audioAnalyzer{client: client, modelID: modelID}
	return analyzeBatchInParallel(ctx, analyzer, audiosBase64, "audio")
}

func RunAudioCLI() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . audio <path_to_audio1> [path_to_audio2 ...]")
		return
	}

	audioPaths := os.Args[1:]
	var audiosBase64 []string

	for _, path := range audioPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", path, err)
			os.Exit(1)
		}
		audiosBase64 = append(audiosBase64, base64.StdEncoding.EncodeToString(data))
	}

	fmt.Printf("Processing %d audios in parallel...\n", len(audiosBase64))
	results := AnalyzeAudiosParallel(audiosBase64)

	fmt.Println("\n--- Results ---")
	for i, res := range results {
		fmt.Printf("\n[Audio %s]:\n%s\n", audioPaths[i], res)
	}
}
