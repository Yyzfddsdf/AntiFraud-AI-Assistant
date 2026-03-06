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

type ResponsesInputPart struct {
	Type        string                 `json:"type"`
	Text        string                 `json:"text,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"`
	FileID      string                 `json:"file_id,omitempty"`
	Detail      string                 `json:"detail,omitempty"`
	ExtraFields map[string]interface{} `json:"-"`
}

func (p ResponsesInputPart) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"type": p.Type,
	}
	if strings.TrimSpace(p.Text) != "" {
		payload["text"] = p.Text
	}
	if strings.TrimSpace(p.ImageURL) != "" {
		payload["image_url"] = p.ImageURL
	}
	if strings.TrimSpace(p.FileID) != "" {
		payload["file_id"] = p.FileID
	}
	if strings.TrimSpace(p.Detail) != "" {
		payload["detail"] = p.Detail
	}
	for key, value := range p.ExtraFields {
		payload[key] = value
	}
	return json.Marshal(payload)
}

type ResponsesMessagePart = ResponsesInputPart

type ResponsesInputItem struct {
	Type        string                 `json:"type"`
	Role        string                 `json:"role,omitempty"`
	Content     []ResponsesInputPart   `json:"content,omitempty"`
	Status      string                 `json:"status,omitempty"`
	ExtraFields map[string]interface{} `json:"-"`
}

func (i ResponsesInputItem) MarshalJSON() ([]byte, error) {
	payload := map[string]interface{}{
		"type": i.Type,
	}
	if strings.TrimSpace(i.Role) != "" {
		payload["role"] = i.Role
	}
	if len(i.Content) > 0 {
		payload["content"] = i.Content
	}
	if strings.TrimSpace(i.Status) != "" {
		payload["status"] = i.Status
	}
	for key, value := range i.ExtraFields {
		payload[key] = value
	}
	return json.Marshal(payload)
}

type ResponsesMessage = ResponsesInputItem

func MakeResponsesInputText(role, text string) ResponsesInputItem {
	return ResponsesInputItem{
		Type: "message",
		Role: role,
		Content: []ResponsesInputPart{{
			Type: "input_text",
			Text: text,
		}},
	}
}

func MakeResponsesMessageText(role, text string) ResponsesMessage {
	return MakeResponsesInputText(role, text)
}

func MakeResponsesOutputText(role, text string) ResponsesInputItem {
	return ResponsesInputItem{
		Type: "message",
		Role: role,
		Content: []ResponsesInputPart{{
			Type: "output_text",
			Text: text,
		}},
	}
}

func MakeResponsesMessageOutputText(role, text string) ResponsesMessage {
	return MakeResponsesOutputText(role, text)
}

func MakeResponsesInputImageURL(role, imageURL string) ResponsesInputItem {
	return ResponsesInputItem{
		Type: "message",
		Role: role,
		Content: []ResponsesInputPart{{
			Type:     "input_image",
			ImageURL: imageURL,
		}},
	}
}

func MakeResponsesMessageImageURL(role, imageURL string) ResponsesMessage {
	return MakeResponsesInputImageURL(role, imageURL)
}

type ResponsesRequest struct {
	Model           string                 `json:"model"`
	Messages        []ResponsesMessage     `json:"-"`
	Input           interface{}            `json:"input"`
	Tools           []Tool                 `json:"tools,omitempty"`
	Stream          bool                   `json:"stream,omitempty"`
	ToolChoice      interface{}            `json:"tool_choice,omitempty"`
	Temperature     float32                `json:"-"`
	TopP            float32                `json:"-"`
	MaxOutputTokens int                    `json:"-"`
	ExtraFields     map[string]interface{} `json:"-"`
}

func (r ResponsesRequest) MarshalJSON() ([]byte, error) {
	input := r.Input
	if input == nil && len(r.Messages) > 0 {
		input = r.Messages
	}

	payload := map[string]interface{}{
		"model": r.Model,
		"input": input,
	}
	if len(r.Tools) > 0 {
		payload["tools"] = r.Tools
	}
	if r.Stream {
		payload["stream"] = true
	}
	if r.ToolChoice != nil {
		payload["tool_choice"] = r.ToolChoice
	}
	if r.Temperature != 0 {
		payload["temperature"] = r.Temperature
	}
	if r.TopP != 0 {
		payload["top_p"] = r.TopP
	}
	if r.MaxOutputTokens > 0 {
		payload["max_output_tokens"] = r.MaxOutputTokens
	}
	for key, value := range r.ExtraFields {
		payload[key] = value
	}
	return json.Marshal(payload)
}

func (r *ResponsesRequest) SetField(key string, value interface{}) {
	if r.ExtraFields == nil {
		r.ExtraFields = map[string]interface{}{}
	}
	r.ExtraFields[key] = value
}

type ResponsesOutputContent struct {
	Type        string `json:"type"`
	Text        string `json:"text,omitempty"`
	Annotations []any  `json:"annotations,omitempty"`
	Logprobs    []any  `json:"logprobs,omitempty"`
}

type ResponsesOutputItem struct {
	ID        string                   `json:"id,omitempty"`
	Type      string                   `json:"type"`
	Status    string                   `json:"status,omitempty"`
	Role      string                   `json:"role,omitempty"`
	Content   []ResponsesOutputContent `json:"content,omitempty"`
	CallID    string                   `json:"call_id,omitempty"`
	Name      string                   `json:"name,omitempty"`
	Arguments string                   `json:"arguments,omitempty"`
	Output    interface{}              `json:"output,omitempty"`
	Action    map[string]interface{}   `json:"action,omitempty"`
	Results   []map[string]interface{} `json:"results,omitempty"`
	Error     interface{}              `json:"error,omitempty"`
}

type ResponsesUsage struct {
	InputTokens         int                    `json:"input_tokens,omitempty"`
	OutputTokens        int                    `json:"output_tokens,omitempty"`
	TotalTokens         int                    `json:"total_tokens,omitempty"`
	InputTokensDetails  map[string]interface{} `json:"input_tokens_details,omitempty"`
	OutputTokensDetails map[string]interface{} `json:"output_tokens_details,omitempty"`
}

type ResponsesResponse struct {
	ID          string                 `json:"id"`
	Object      string                 `json:"object"`
	CreatedAt   int64                  `json:"created_at,omitempty"`
	CompletedAt int64                  `json:"completed_at,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Model       string                 `json:"model,omitempty"`
	Error       interface{}            `json:"error"`
	Output      []ResponsesOutputItem  `json:"output,omitempty"`
	Usage       *ResponsesUsage        `json:"usage,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ResponsesStreamEvent struct {
	Type           string               `json:"type"`
	Delta          string               `json:"delta,omitempty"`
	Item           *ResponsesOutputItem `json:"item,omitempty"`
	ItemID         string               `json:"item_id,omitempty"`
	OutputIndex    *int                 `json:"output_index,omitempty"`
	ContentIndex   *int                 `json:"content_index,omitempty"`
	SequenceNumber *int                 `json:"sequence_number,omitempty"`
	Name           string               `json:"name,omitempty"`
	Arguments      string               `json:"arguments,omitempty"`
	Response       *ResponsesResponse   `json:"response,omitempty"`
	Error          interface{}          `json:"error,omitempty"`
}

type ResponsesStream struct {
	body   io.ReadCloser
	reader *bufio.Reader
}

func (s *ResponsesStream) Recv() (ResponsesStreamEvent, error) {
	dataLines := make([]string, 0, 2)

	for {
		line, err := s.reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return ResponsesStreamEvent{}, err
		}

		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			if len(dataLines) == 0 {
				if err == io.EOF {
					return ResponsesStreamEvent{}, io.EOF
				}
				continue
			}
			break
		}

		if strings.HasPrefix(trimmed, ":") || strings.HasPrefix(trimmed, "event:") {
			if err == io.EOF && len(dataLines) == 0 {
				return ResponsesStreamEvent{}, io.EOF
			}
			continue
		}
		if strings.HasPrefix(trimmed, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(trimmed, "data:")))
		}

		if err == io.EOF {
			break
		}
	}

	if len(dataLines) == 0 {
		return ResponsesStreamEvent{}, io.EOF
	}

	payload := strings.Join(dataLines, "\n")
	if payload == "[DONE]" {
		return ResponsesStreamEvent{}, io.EOF
	}

	var event ResponsesStreamEvent
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		return ResponsesStreamEvent{}, fmt.Errorf("decode responses stream event failed: %w", err)
	}
	return event, nil
}

func (s *ResponsesStream) Close() error {
	if s.body != nil {
		return s.body.Close()
	}
	return nil
}

func (c *Client) CreateResponses(ctx context.Context, req ResponsesRequest) (ResponsesResponse, error) {
	resp, err := c.doResponses(ctx, req)
	if err != nil {
		return ResponsesResponse{}, err
	}
	defer resp.Body.Close()

	var result ResponsesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return ResponsesResponse{}, fmt.Errorf("decode responses response failed: %w", err)
	}
	return result, nil
}

func (c *Client) CreateResponsesStream(ctx context.Context, req ResponsesRequest) (*ResponsesStream, error) {
	req.Stream = true
	resp, err := c.doResponses(ctx, req)
	if err != nil {
		return nil, err
	}
	return &ResponsesStream{
		body:   resp.Body,
		reader: bufio.NewReader(resp.Body),
	}, nil
}

func (c *Client) doResponses(ctx context.Context, req ResponsesRequest) (*http.Response, error) {
	if strings.TrimSpace(req.Model) == "" {
		return nil, fmt.Errorf("responses model is empty")
	}
	if req.Input == nil && len(req.Messages) == 0 {
		return nil, fmt.Errorf("responses input is nil")
	}
	if c.cfg.BaseURL == "" {
		return nil, fmt.Errorf("base url is empty")
	}

	endpoint := c.cfg.BaseURL + "/responses"
	payload, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encode responses request failed: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("build responses request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.cfg.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send responses request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("responses request failed, status=%d, body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return resp, nil
}
