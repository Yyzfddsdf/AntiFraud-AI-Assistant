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

var autoHeadersToStrip = []string{
	"OpenAI-Organization",
	"OpenAI-Project",
	"User-Agent",
	"X-Stainless-Arch",
	"X-Stainless-Lang",
	"X-Stainless-OS",
	"X-Stainless-Package-Version",
	"X-Stainless-Retry-Count",
	"X-Stainless-Runtime",
	"X-Stainless-Runtime-Version",
	"X-Stainless-Timeout",
}

type strippedHeaderTransport struct {
	base http.RoundTripper
}

func (t strippedHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := t.base
	if transport == nil {
		transport = http.DefaultTransport
	}

	cloned := req.Clone(req.Context())
	cloned.Header = req.Header.Clone()
	for _, headerName := range autoHeadersToStrip {
		cloned.Header.Del(headerName)
	}

	return transport.RoundTrip(cloned)
}

func newMinimalHeaderHTTPClient() *http.Client {
	return &http.Client{
		Transport: strippedHeaderTransport{base: http.DefaultTransport},
	}
}

func NewMinimalHeaderHTTPClient() *http.Client {
	return newMinimalHeaderHTTPClient()
}

// Client is a minimal Responses HTTP client.
//
// Typical usage:
//
//	client := minimalopenai.NewClient(apiKey, baseURL)
//
//	body := minimalopenai.ResponsesRequest{
//	    Model: "gpt-5.4",
//	    Input: []any{
//	        minimalopenai.Message{
//	            Type: "message",
//	            Role: "user",
//	            Content: []any{
//	                minimalopenai.InputText{
//	                    Type: "input_text",
//	                    Text: "Hello",
//	                },
//	            },
//	        },
//	    },
//	}
//
//	payload, err := client.CreateResponse(ctx, body)
//
// Enable reasoning:
//
//	body := minimalopenai.ResponsesRequest{
//	    Model: "gpt-5.4",
//	    Input: []any{
//	        minimalopenai.Message{
//	            Type: "message",
//	            Role: "user",
//	            Content: []any{
//	                minimalopenai.InputText{
//	                    Type: "input_text",
//	                    Text: "Solve this step by step.",
//	                },
//	            },
//	        },
//	    },
//	    Reasoning: &minimalopenai.Reasoning{
//	        Effort:  minimalopenai.ReasoningEffortMedium,
//	        Summary: minimalopenai.ReasoningSummaryAuto,
//	    },
//	}
//
//	Reasoning request fields:
//	- reasoning.effort
//	- reasoning.summary
//	- reasoning.generate_summary
//
//	Verified in official openai-go v3.26.0 source:
//	- Responses requests have a top-level reasoning object.
//	- The SDK exposes reasoning effort/summary enums.
//	- generate_summary exists, but is deprecated. Prefer summary.
//
//	Verified against the current gateway:
//	- Sending reasoning: {"effort":"medium","summary":"auto"} is accepted.
//	- The response includes a top-level reasoning object.
//	- The response output may also contain an item with type "reasoning".
//
//	How to configure reasoning effort:
//	- ReasoningEffortNone
//	- ReasoningEffortMinimal
//	- ReasoningEffortLow
//	- ReasoningEffortMedium
//	- ReasoningEffortHigh
//	- ReasoningEffortXhigh
//
//	How to configure reasoning summary:
//	- ReasoningSummaryAuto
//	- ReasoningSummaryConcise
//	- ReasoningSummaryDetailed
//
//	Recommended pattern:
//	- Use effort to control how much reasoning the model spends.
//	- Use summary to control how much reasoning summary the API returns.
//	- Do not set generate_summary unless you are intentionally targeting an older backend.
//
//	Model-specific caveats from official SDK comments:
//	- gpt-5.1 defaults to none, and supports none/low/medium/high.
//	- Models before gpt-5.1 default to medium and do not support none.
//	- gpt-5-pro only supports high.
//	- xhigh is only available on newer reasoning-capable models.
//
//	Do not assume every model accepts every reasoning value.
//	Check the target model/backend before hard-coding a value.
//
// Multi-modal input:
//
//	minimalopenai.Message{
//	    Type: "message",
//	    Role: "user",
//	    Content: []any{
//	        minimalopenai.InputText{
//	            Type: "input_text",
//	            Text: "Describe these images",
//	        },
//	        minimalopenai.InputImage{
//	            Type:     "input_image",
//	            ImageURL: imageURL1,
//	        },
//	        minimalopenai.InputImage{
//	            Type:     "input_image",
//	            ImageURL: imageURL2,
//	        },
//	    },
//	}
//
// Tools:
//
//	tools := []any{
//	    minimalopenai.WebSearchTool{
//	        Type:              "web_search",
//	        SearchContextSize: "medium",
//	    },
//	    minimalopenai.FunctionTool{
//	        Type:        "function",
//	        Name:        "get_weather",
//	        Description: "Get weather for a city",
//	        Parameters: map[string]any{
//	            "type": "object",
//	            "properties": map[string]any{
//	                "city": map[string]any{"type": "string"},
//	            },
//	            "required": []string{"city"},
//	        },
//	    },
//	}
//
// Function tool continuation:
//
//  1. Send ResponsesRequest.
//  2. If output contains a function_call, execute your local tool yourself.
//  3. Append both items into the next request history:
//     - minimalopenai.ResponseFunctionCall
//     - minimalopenai.ResponseFunctionCallOutput
//  4. Send the next ResponsesRequest again.
//
// Assistant history continuation:
//
//	Rebuild assistant history by hand instead of reusing raw response objects.
//	A minimal assistant history item usually looks like this:
//
//	minimalopenai.Message{
//	    Type: "message",
//	    Role: "assistant",
//	    Content: []any{
//	        minimalopenai.OutputText{
//	            Type: "output_text",
//	            Text: assistantText,
//	        },
//	    },
//	}
//
// Streaming:
//
//	stream, err := client.StreamResponse(ctx, body)
//	if err != nil {
//	    return err
//	}
//	defer stream.Close()
//
//	for stream.Next() {
//	    event := stream.Current()
//	    // event.Event may be empty on some gateways.
//	    // In that case inspect event.Data and parse the JSON type field yourself.
//	}
//	if err := stream.Err(); err != nil {
//	    return err
//	}
//
// StreamResponse automatically sets body.Stream = true and requests
// text/event-stream.
//
// Some reasoning-capable backends may also emit reasoning-related stream events,
// for example response.reasoning_text.delta or response.reasoning_summary_text.delta.
// This package leaves those events as raw JSON for caller-side handling.
//
// How to parse reasoning content:
//
//	Normal non-stream response:
//	1. Check payload["reasoning"].
//	   This is top-level reasoning metadata, not the final assistant answer.
//	2. Check payload["output"].
//	3. If an output item has item["type"] == "reasoning",
//	   that item is a reasoning item.
//	4. A reasoning item often carries item["summary"].
//	   Depending on backend/model, summary may be empty.
//	5. The final assistant answer is still usually in message/output_text,
//	   not in the reasoning item.
//
//	Typical reasoning-aware parsing rule for normal responses:
//	- payload["reasoning"]: reasoning metadata
//	- output item type == "reasoning": reasoning item
//	- output item type == "message": assistant answer
//
//	Streaming response:
//	1. Read event.Event first.
//	2. If event.Event is empty, read event.Data JSON -> type.
//	3. Treat these as reasoning events when they appear:
//	   - response.reasoning_text.delta
//	   - response.reasoning_text.done
//	   - response.reasoning_summary_text.delta
//	   - response.reasoning_summary_text.done
//	   - response.output_item.added with item.type == "reasoning"
//	   - response.output_item.done with item.type == "reasoning"
//	4. Treat response.output_text.delta as assistant answer text, not reasoning text.
//
//	Current gateway behavior observed in this workspace:
//	- It returns top-level payload["reasoning"].
//	- It may return an output item with type == "reasoning".
//	- It did not emit reasoning_text.delta in the latest verified stream run.
//
//	So in this package, the most reliable rule is:
//	- reasoning metadata: payload["reasoning"]
//	- reasoning item: output item type == "reasoning"
//	- assistant answer: message/output_text
//
// Response parsing:
//
//	This package does not provide a full strong-typed response model like the
//	official SDK. CreateResponse returns map[string]any, and StreamResponse
//	returns raw stream events. That means response parsing is done by your code.
//
//	Normal response shape:
//
//	    payload["output"] -> []any
//
//	Each output item usually has a type field, for example:
//
//	    "message"
//	    "function_call"
//	    "web_search_call"
//
//	Typical normal-response parsing flow:
//
//	1. Read payload["output"].
//	2. For each item, inspect item["type"].
//	3. If type == "message", inspect item["content"].
//	4. Inside content, inspect each content item's type:
//	   - "output_text" -> read text
//	   - "refusal" -> read refusal
//	5. If type == "function_call", read:
//	   - call_id
//	   - name
//	   - arguments
//	6. If type == "web_search_call", usually do not append it into next-turn history.
//
//	In practice a final assistant text often comes from either:
//
//	    payload["output_text"]
//
//	or from:
//
//	    payload["output"] -> message -> content -> output_text -> text
//
// Streaming response parsing:
//
//	StreamResponse gives you ResponseStream. Each Next() yields one StreamEvent.
//
//	    event := stream.Current()
//
//	On standard SSE servers:
//	    event.Event is usually filled from the SSE event: line.
//
//	On some gateways:
//	    event.Event may be empty,
//	    and the real event type is only inside event.Data JSON -> type.
//
//	Typical stream parsing flow:
//
//	1. Iterate with stream.Next().
//	2. Determine event type from:
//	   - event.Event
//	   - or event.Data.type when event.Event is empty
//	3. If type == "response.output_text.delta", read delta and print/append it.
//	4. If type == "response.output_item.done", collect item for fallback assembly.
//	5. If type == "response.completed", read completed.response as the final payload.
//	6. If no completed event arrives, you may need to build a fallback payload
//	   from collected output items and accumulated text.
//
// Function tool continuation parsing:
//
//	When the response contains a function_call, you should:
//
//	1. Read:
//	   - item["call_id"]
//	   - item["name"]
//	   - item["arguments"]
//	2. Execute your local tool yourself.
//	3. Append two items into next-turn history:
//	   - ResponseFunctionCall
//	   - ResponseFunctionCallOutput
//	4. Send ResponsesRequest again.
//
//	Example continuation history items:
//
//	    minimalopenai.ResponseFunctionCall{
//	        Type:      "function_call",
//	        CallID:    callID,
//	        Name:      name,
//	        Arguments: argumentsJSON,
//	    }
//
//	    minimalopenai.ResponseFunctionCallOutput{
//	        Type:   "function_call_output",
//	        CallID: callID,
//	        Output: outputJSON,
//	    }
//
//	For assistant history text continuation, rebuild a new Message by hand.
//	Do not directly reuse raw output message objects as next-turn history.
func NewClient(apiKey, baseURL string) *Client {
	cfg := DefaultConfig(apiKey)
	if trimmedBaseURL := strings.TrimSpace(baseURL); trimmedBaseURL != "" {
		cfg.BaseURL = trimmedBaseURL
	}
	cfg.HTTPClient = newMinimalHeaderHTTPClient()
	return NewClientWithConfig(cfg)
}

func (c *Client) NewRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	if strings.TrimSpace(c.cfg.BaseURL) == "" {
		return nil, fmt.Errorf("base url is empty")
	}

	var requestBody io.Reader
	if body != nil {
		encodedBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		requestBody = bytes.NewReader(encodedBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, requestBody)
	if err != nil {
		return nil, err
	}
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

func (c *Client) CreateResponse(ctx context.Context, body ResponsesRequest) (map[string]any, error) {
	req, err := c.NewRequest(ctx, http.MethodPost, "/responses", body)
	if err != nil {
		return nil, err
	}

	httpClient := c.cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("POST %q: %s %s", req.URL.String(), resp.Status, string(responseBody))
	}

	var payload map[string]any
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

type StreamEvent struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type ResponseStream struct {
	body    io.ReadCloser
	scanner *bufio.Scanner
	current StreamEvent
	err     error
}

func (c *Client) StreamResponse(ctx context.Context, body ResponsesRequest) (*ResponseStream, error) {
	body.Stream = true

	req, err := c.NewRequest(ctx, http.MethodPost, "/responses", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")

	httpClient := c.cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		responseBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("POST %q: %s", req.URL.String(), resp.Status)
		}
		return nil, fmt.Errorf("POST %q: %s %s", req.URL.String(), resp.Status, string(responseBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(nil, bufio.MaxScanTokenSize<<9)
	return &ResponseStream{body: resp.Body, scanner: scanner}, nil
}

func (s *ResponseStream) Next() bool {
	if s.err != nil {
		return false
	}

	eventType := ""
	dataLines := make([]string, 0)
	for s.scanner.Scan() {
		line := s.scanner.Text()
		if line == "" {
			if len(dataLines) == 0 && eventType == "" {
				continue
			}
			data := strings.Join(dataLines, "\n")
			if strings.TrimSpace(data) == "[DONE]" {
				return false
			}
			s.current = StreamEvent{
				Event: eventType,
				Data:  json.RawMessage(data),
			}
			return true
		}

		if strings.HasPrefix(line, "event:") {
			eventType = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			continue
		}
		if strings.HasPrefix(line, "data:") {
			dataLines = append(dataLines, strings.TrimSpace(strings.TrimPrefix(line, "data:")))
			continue
		}

		// Some gateways return a plain JSON response body even when Accept is
		// text/event-stream. Keep the raw lines so callers can fall back.
		dataLines = append(dataLines, line)
	}

	if err := s.scanner.Err(); err != nil {
		s.err = err
		return false
	}

	if len(dataLines) > 0 {
		s.current = StreamEvent{
			Event: eventType,
			Data:  json.RawMessage(strings.Join(dataLines, "\n")),
		}
		return true
	}
	return false
}

func (s *ResponseStream) Current() StreamEvent {
	return s.current
}

func (s *ResponseStream) Err() error {
	return s.err
}

func (s *ResponseStream) Close() error {
	if s.body == nil {
		return nil
	}
	return s.body.Close()
}

type Message struct {
	Type    string `json:"type"`
	Role    string `json:"role"`
	Content []any  `json:"content"`
}

type ResponsesRequest struct {
	Model     string     `json:"model"`
	Input     []any      `json:"input"`
	Tools     []any      `json:"tools,omitempty"`
	Stream    bool       `json:"stream,omitempty"`
	Reasoning *Reasoning `json:"reasoning,omitempty"`
}

type Reasoning struct {
	Effort          ReasoningEffort          `json:"effort,omitempty"`
	GenerateSummary ReasoningGenerateSummary `json:"generate_summary,omitempty"`
	Summary         ReasoningSummary         `json:"summary,omitempty"`
}

type ReasoningEffort string

const (
	ReasoningEffortNone    ReasoningEffort = "none"
	ReasoningEffortMinimal ReasoningEffort = "minimal"
	ReasoningEffortLow     ReasoningEffort = "low"
	ReasoningEffortMedium  ReasoningEffort = "medium"
	ReasoningEffortHigh    ReasoningEffort = "high"
	ReasoningEffortXhigh   ReasoningEffort = "xhigh"
)

type ReasoningGenerateSummary string

const (
	ReasoningGenerateSummaryAuto     ReasoningGenerateSummary = "auto"
	ReasoningGenerateSummaryConcise  ReasoningGenerateSummary = "concise"
	ReasoningGenerateSummaryDetailed ReasoningGenerateSummary = "detailed"
)

type ReasoningSummary string

const (
	ReasoningSummaryAuto     ReasoningSummary = "auto"
	ReasoningSummaryConcise  ReasoningSummary = "concise"
	ReasoningSummaryDetailed ReasoningSummary = "detailed"
)

type InputText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type InputImage struct {
	Type     string `json:"type"`
	ImageURL string `json:"image_url"`
	Detail   string `json:"detail,omitempty"`
}

type OutputText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type FunctionTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type WebSearchTool struct {
	Type              string `json:"type"`
	SearchContextSize string `json:"search_context_size,omitempty"`
}

type ResponseFunctionCall struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ResponseFunctionCallOutput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}
