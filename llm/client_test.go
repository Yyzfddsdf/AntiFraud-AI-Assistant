package llm

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("  token  ")
	if cfg.APIKey != "token" {
		t.Fatalf("expected trimmed api key, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected default base url: %q", cfg.BaseURL)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	client := NewClientWithConfig(Config{
		APIKey:  "  token  ",
		BaseURL: " https://example.com/v1/ ",
	})
	if client.cfg.APIKey != "token" {
		t.Fatalf("expected trimmed api key, got %q", client.cfg.APIKey)
	}
	if client.cfg.BaseURL != "https://example.com/v1" {
		t.Fatalf("expected trimmed base url without trailing slash, got %q", client.cfg.BaseURL)
	}
	if client.cfg.HTTPClient == nil {
		t.Fatalf("expected default http client to be initialized")
	}

	customHTTP := &http.Client{}
	client2 := NewClientWithConfig(Config{
		APIKey:     "k",
		BaseURL:    "https://example.com/v1",
		HTTPClient: customHTTP,
	})
	if client2.cfg.HTTPClient != customHTTP {
		t.Fatalf("expected custom http client to be preserved")
	}
}

func TestChatMessagePartMarshalJSON(t *testing.T) {
	part := ChatMessagePart{
		Type: "text",
		Text: "hello",
		ExtraFields: map[string]interface{}{
			"cache_control": map[string]interface{}{"type": "ephemeral"},
		},
	}
	raw, err := json.Marshal(part)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if payload["type"] != "text" || payload["text"] != "hello" {
		t.Fatalf("unexpected payload fields: %+v", payload)
	}
	if _, ok := payload["cache_control"]; !ok {
		t.Fatalf("expected extra field cache_control in payload")
	}
}

func TestChatCompletionMessageMarshalJSON_MultiContentAndToolCallID(t *testing.T) {
	msg := ChatCompletionMessage{
		Role:    ChatMessageRoleUser,
		Content: "fallback text",
		MultiContent: []ChatMessagePart{
			{Type: "text", Text: "part-text"},
		},
		ToolCallID: "   ",
	}
	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	content, ok := payload["content"].([]interface{})
	if !ok || len(content) != 1 {
		t.Fatalf("expected multi content array, got: %#v", payload["content"])
	}
	if _, exists := payload["tool_call_id"]; exists {
		t.Fatalf("did not expect tool_call_id when empty/blank")
	}
}

