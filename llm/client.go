package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ToolTypeFunction = "function"
	ToolTypeWebSearch = "web_search"

	ChatMessageRoleSystem    = "system"
	ChatMessageRoleUser      = "user"
	ChatMessageRoleAssistant = "assistant"
	ChatMessageRoleTool      = "tool"
)

type Config struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func DefaultConfig(apiKey string) Config {
	return Config{
		APIKey:  strings.TrimSpace(apiKey),
		BaseURL: "https://api.openai.com/v1",
	}
}

type Client struct {
	cfg Config
}

func NewClientWithConfig(cfg Config) *Client {
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{}
	}
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	return &Client{cfg: cfg}
}

type FunctionDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

type Tool struct {
	Type     string              `json:"type"`
	Function *FunctionDefinition `json:"function,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type ChatMessageImageURL struct {
	URL string `json:"url"`
}

type ChatMessageVideoURL struct {
	URL string `json:"url"`
}

type ChatMessagePart struct {
	Type     string               `json:"type"`
	Text     string               `json:"text,omitempty"`
	ImageURL *ChatMessageImageURL `json:"image_url,omitempty"`
	VideoURL *ChatMessageVideoURL `json:"video_url,omitempty"`
	// ExtraFields 用于透传尚未在结构体中预定义的 provider 扩展字段。
	//
	// 例如新增一个自定义 part（audio_url）可这样写：
	//   ChatMessagePart{
	//     Type: "audio_url",
	//     ExtraFields: map[string]interface{}{
	//       "audio_url": map[string]interface{}{"url": "https://example.com/a.wav"},
	//     },
	//   }
	ExtraFields map[string]interface{} `json:"-"`
}

func (p ChatMessagePart) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"type": p.Type,
	}
	if p.Text != "" {
		payload["text"] = p.Text
	}
	if p.ImageURL != nil {
		payload["image_url"] = p.ImageURL
	}
	if p.VideoURL != nil {
		payload["video_url"] = p.VideoURL
	}
	for key, value := range p.ExtraFields {
		payload[key] = value
	}
	return json.Marshal(payload)
}

type ChatCompletionMessage struct {
	Role         string                 `json:"role"`
	Content      string                 `json:"content,omitempty"`
	MultiContent []ChatMessagePart      `json:"-"`
	ToolCalls    []ToolCall             `json:"tool_calls,omitempty"`
	ToolCallID   string                 `json:"tool_call_id,omitempty"`
	ExtraFields  map[string]interface{} `json:"-"`
}

func (m ChatCompletionMessage) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"role": m.Role,
	}

	if len(m.MultiContent) > 0 {
		payload["content"] = m.MultiContent
	} else {
		payload["content"] = m.Content
	}

	if len(m.ToolCalls) > 0 {
		payload["tool_calls"] = m.ToolCalls
	}
	if strings.TrimSpace(m.ToolCallID) != "" {
		payload["tool_call_id"] = m.ToolCallID
	}
	for key, value := range m.ExtraFields {
		payload[key] = value
	}

	return json.Marshal(payload)
}

func (m *ChatCompletionMessage) UnmarshalJSON(data []byte) error {
	var aux struct {
		Role       string          `json:"role"`
		Content    json.RawMessage `json:"content"`
		ToolCalls  []ToolCall      `json:"tool_calls"`
		ToolCallID string          `json:"tool_call_id"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Role = aux.Role
	m.ToolCalls = aux.ToolCalls
	m.ToolCallID = aux.ToolCallID
	m.Content = ""
	m.MultiContent = nil

	if len(aux.Content) == 0 || string(aux.Content) == "null" {
		return nil
	}

	var text string
	if err := json.Unmarshal(aux.Content, &text); err == nil {
		m.Content = text
		return nil
	}

	var parts []ChatMessagePart
	if err := json.Unmarshal(aux.Content, &parts); err == nil {
		m.MultiContent = parts
		return nil
	}

	m.Content = string(aux.Content)
	return nil
}

type ChatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []ChatCompletionMessage `json:"messages"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature float32                 `json:"temperature,omitempty"`
	TopP        float32                 `json:"top_p,omitempty"`
	Stream      bool                    `json:"stream,omitempty"`
	Tools       []Tool                  `json:"tools,omitempty"`
	ToolChoice  interface{}             `json:"tool_choice,omitempty"`
	ExtraFields map[string]interface{}  `json:"-"`
}

func (r ChatCompletionRequest) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"model":    r.Model,
		"messages": r.Messages,
	}

	if r.MaxTokens > 0 {
		payload["max_tokens"] = r.MaxTokens
	}
	if r.Temperature != 0 {
		payload["temperature"] = r.Temperature
	}
	if r.TopP != 0 {
		payload["top_p"] = r.TopP
	}
	if r.Stream {
		payload["stream"] = true
	}
	if len(r.Tools) > 0 {
		payload["tools"] = r.Tools
	}
	if r.ToolChoice != nil {
		payload["tool_choice"] = r.ToolChoice
	}
	for key, value := range r.ExtraFields {
		payload[key] = value
	}

	return json.Marshal(payload)
}

func (r *ChatCompletionRequest) SetField(key string, value interface{}) {
	if r.ExtraFields == nil {
		r.ExtraFields = map[string]interface{}{}
	}
	r.ExtraFields[key] = value
}

type EmbeddingRequest struct {
	Model          string                 `json:"model"`
	Input          interface{}            `json:"input"`
	EncodingFormat string                 `json:"encoding_format,omitempty"`
	Dimensions     int                    `json:"dimensions,omitempty"`
	User           string                 `json:"user,omitempty"`
	ExtraFields    map[string]interface{} `json:"-"`
}

func (r EmbeddingRequest) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"model": r.Model,
		"input": r.Input,
	}
	if strings.TrimSpace(r.EncodingFormat) != "" {
		payload["encoding_format"] = strings.TrimSpace(r.EncodingFormat)
	}
	if r.Dimensions > 0 {
		payload["dimensions"] = r.Dimensions
	}
	if strings.TrimSpace(r.User) != "" {
		payload["user"] = strings.TrimSpace(r.User)
	}
	for key, value := range r.ExtraFields {
		payload[key] = value
	}
	return json.Marshal(payload)
}

func (r *EmbeddingRequest) SetField(key string, value interface{}) {
	if r.ExtraFields == nil {
		r.ExtraFields = map[string]interface{}{}
	}
	r.ExtraFields[key] = value
}

type EmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type EmbeddingResponse struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

type ChatCompletionChoice struct {
	Message ChatCompletionMessage `json:"message"`
}

type ChatCompletionResponse struct {
	Choices []ChatCompletionChoice `json:"choices"`
}

type ChatCompletionStreamFunctionCallDelta struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type ChatCompletionStreamToolCallDelta struct {
	Index    *int                                   `json:"index,omitempty"`
	ID       string                                 `json:"id,omitempty"`
	Type     string                                 `json:"type,omitempty"`
	Function *ChatCompletionStreamFunctionCallDelta `json:"function,omitempty"`
}

type ChatCompletionStreamChoiceDelta struct {
	Role      string                              `json:"role,omitempty"`
	Content   string                              `json:"content"`
	ToolCalls []ChatCompletionStreamToolCallDelta `json:"tool_calls,omitempty"`
}

type ChatCompletionStreamChoice struct {
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason,omitempty"`
}

type ChatCompletionStreamResponse struct {
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

type ChatCompletionStream struct {
	body   io.ReadCloser
	reader *bufio.Reader
}

func (s *ChatCompletionStream) Recv() (ChatCompletionStreamResponse, error) {
	for {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return ChatCompletionStreamResponse{}, io.EOF
			}
			return ChatCompletionStreamResponse{}, err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ":") {
			continue
		}
		if !strings.HasPrefix(trimmed, "data:") {
			continue
		}

		payload := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
		if payload == "" {
			continue
		}
		if payload == "[DONE]" {
			return ChatCompletionStreamResponse{}, io.EOF
		}

		var chunk ChatCompletionStreamResponse
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return ChatCompletionStreamResponse{}, err
		}
		return chunk, nil
	}
}

func (s *ChatCompletionStream) Close() error {
	if s.body != nil {
		return s.body.Close()
	}
	return nil
}

func (c *Client) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (ChatCompletionResponse, error) {
	resp, err := c.doChatCompletion(ctx, req)
	if err != nil {
		return ChatCompletionResponse{}, err
	}
	defer resp.Body.Close()

	var result ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ChatCompletionResponse{}, fmt.Errorf("decode chat completion response failed: %w", err)
	}

	return result, nil
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionStream, error) {
	req.Stream = true
	resp, err := c.doChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	return &ChatCompletionStream{
		body:   resp.Body,
		reader: bufio.NewReader(resp.Body),
	}, nil
}

func (c *Client) CreateEmbeddings(ctx context.Context, req EmbeddingRequest) (EmbeddingResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return EmbeddingResponse{}, fmt.Errorf("embedding model is empty")
	}
	if req.Input == nil {
		return EmbeddingResponse{}, fmt.Errorf("embedding input is nil")
	}

	resp, err := c.doEmbeddings(ctx, req)
	if err != nil {
		return EmbeddingResponse{}, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return EmbeddingResponse{}, fmt.Errorf("decode embedding response failed: %w", err)
	}
	return result, nil
}

func (c *Client) doChatCompletion(ctx context.Context, req ChatCompletionRequest) (*http.Response, error) {
	if c.cfg.BaseURL == "" {
		return nil, fmt.Errorf("base url is empty")
	}
	endpoint := c.cfg.BaseURL + "/chat/completions"

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode chat completion request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("build chat completion request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send chat completion request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chat completion request failed, status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return resp, nil
}

func (c *Client) doEmbeddings(ctx context.Context, req EmbeddingRequest) (*http.Response, error) {
	if c.cfg.BaseURL == "" {
		return nil, fmt.Errorf("base url is empty")
	}
	endpoint := c.cfg.BaseURL + "/embeddings"

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode embedding request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("build embedding request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send embedding request failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding request failed, status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return resp, nil
}
