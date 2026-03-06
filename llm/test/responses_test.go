package llm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "antifraud/llm"
)

func TestCreateResponses(t *testing.T) {
	var gotAuth string
	var gotBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"resp_1","object":"response","status":"completed","output":[{"id":"ws_1","type":"web_search_call","status":"completed","action":{"type":"search","query":"weather: Hangzhou, China"}},{"id":"fc_1","type":"function_call","status":"completed","call_id":"call_1","name":"get_weather","arguments":"{\"city\":\"Hangzhou\"}"},{"id":"msg_1","type":"message","role":"assistant","content":[{"type":"output_text","text":"ok"}]}],"usage":{"input_tokens":10,"output_tokens":20,"total_tokens":30}}`))
	}))
	defer server.Close()

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  "token",
		BaseURL: server.URL,
	})

	resp, err := client.CreateResponses(context.Background(), openai.ResponsesRequest{
		Model: "gpt-test",
		Messages: []openai.ResponsesMessage{
			openai.MakeResponsesMessageText("user", "hello"),
		},
		Tools: []openai.Tool{{Type: "web_search"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAuth != "Bearer token" {
		t.Fatalf("unexpected authorization header: %q", gotAuth)
	}
	if gotBody["model"] != "gpt-test" {
		t.Fatalf("unexpected model payload: %#v", gotBody["model"])
	}
	input, ok := gotBody["input"].([]interface{})
	if !ok || len(input) != 1 {
		t.Fatalf("expected messages to marshal into input array, got: %#v", gotBody["input"])
	}
	if resp.ID != "resp_1" || len(resp.Output) != 3 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Output[0].Type != "web_search_call" || resp.Output[0].Action["query"] != "weather: Hangzhou, China" {
		t.Fatalf("unexpected tool call output: %+v", resp.Output[0])
	}
	if resp.Output[1].Type != "function_call" || resp.Output[1].CallID != "call_1" || resp.Output[1].Name != "get_weather" || resp.Output[1].Arguments != `{"city":"Hangzhou"}` {
		t.Fatalf("unexpected function call output: %+v", resp.Output[1])
	}
	if resp.Usage == nil || resp.Usage.InputTokens != 10 || resp.Usage.OutputTokens != 20 || resp.Usage.TotalTokens != 30 {
		t.Fatalf("unexpected usage: %+v", resp.Usage)
	}
}

func TestCreateResponsesStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/responses" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.output_item.added\",\"item\":{\"id\":\"fc_1\",\"type\":\"function_call\",\"status\":\"in_progress\",\"call_id\":\"call_1\",\"name\":\"get_weather\",\"arguments\":\"\"}}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.function_call_arguments.delta\",\"item_id\":\"fc_1\",\"output_index\":0,\"delta\":\"{\\\"city\\\":\",\"sequence_number\":1}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.function_call_arguments.done\",\"item_id\":\"fc_1\",\"output_index\":0,\"arguments\":\"{\\\"city\\\":\\\"Hangzhou\\\"}\",\"name\":\"get_weather\",\"sequence_number\":2}\n\n")
		_, _ = io.WriteString(w, "data: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\",\"object\":\"response\",\"status\":\"completed\",\"output\":[{\"id\":\"fc_1\",\"type\":\"function_call\",\"status\":\"completed\",\"call_id\":\"call_1\",\"name\":\"get_weather\",\"arguments\":\"{\\\"city\\\":\\\"Hangzhou\\\"}\"}],\"usage\":{\"input_tokens\":10,\"output_tokens\":20,\"total_tokens\":30}}}\n\n")
		_, _ = io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  "token",
		BaseURL: server.URL,
	})

	stream, err := client.CreateResponsesStream(context.Background(), openai.ResponsesRequest{
		Model: "gpt-test",
		Messages: []openai.ResponsesMessage{
			openai.MakeResponsesMessageText("user", "hello"),
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer stream.Close()

	event, err := stream.Recv()
	if err != nil {
		t.Fatalf("unexpected recv error: %v", err)
	}
	if event.Type != "response.output_item.added" || event.Item == nil || event.Item.Type != "function_call" || event.Item.CallID != "call_1" {
		t.Fatalf("unexpected event: %+v", event)
	}

	event, err = stream.Recv()
	if err != nil {
		t.Fatalf("unexpected recv error: %v", err)
	}
	if event.Type != "response.function_call_arguments.delta" || event.ItemID != "fc_1" || event.Delta != `{"city":` {
		t.Fatalf("unexpected function call delta event: %+v", event)
	}

	event, err = stream.Recv()
	if err != nil {
		t.Fatalf("unexpected recv error: %v", err)
	}
	if event.Type != "response.function_call_arguments.done" || event.Name != "get_weather" || event.Arguments != `{"city":"Hangzhou"}` {
		t.Fatalf("unexpected function call done event: %+v", event)
	}

	event, err = stream.Recv()
	if err != nil {
		t.Fatalf("unexpected recv error: %v", err)
	}
	if event.Type != "response.completed" || event.Response == nil || event.Response.Usage == nil || event.Response.Usage.TotalTokens != 30 {
		t.Fatalf("unexpected completed event: %+v", event)
	}
	if len(event.Response.Output) != 1 || event.Response.Output[0].Type != "function_call" || event.Response.Output[0].Arguments != `{"city":"Hangzhou"}` {
		t.Fatalf("unexpected completed response output: %+v", event.Response)
	}

	_, err = stream.Recv()
	if err != io.EOF {
		t.Fatalf("expected EOF, got: %v", err)
	}
}

func TestResponsesRequestMarshalJSON_ExtraFields(t *testing.T) {
	req := openai.ResponsesRequest{
		Model: "gpt-test",
		Messages: []openai.ResponsesMessage{
			openai.MakeResponsesMessageText("developer", "你是助手"),
		},
		Temperature:     0.7,
		TopP:            0.95,
		MaxOutputTokens: 1024,
		ExtraFields: map[string]interface{}{
			"temperature": 0.3,
		},
	}

	raw, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	payload := map[string]interface{}{}
	if err := json.NewDecoder(bytes.NewReader(raw)).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if payload["temperature"] != 0.3 {
		t.Fatalf("expected extra field temperature, got: %#v", payload["temperature"])
	}
	if payload["top_p"] != 0.95 {
		t.Fatalf("expected built-in field top_p, got: %#v", payload["top_p"])
	}
	if payload["max_output_tokens"] != float64(1024) {
		t.Fatalf("expected built-in field max_output_tokens, got: %#v", payload["max_output_tokens"])
	}
}
