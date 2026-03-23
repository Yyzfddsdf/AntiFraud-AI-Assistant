package llm_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	openai "antifraud/internal/platform/llm"
)

func TestDefaultConfig(t *testing.T) {
	cfg := openai.DefaultConfig("  token  ")
	if cfg.APIKey != "token" {
		t.Fatalf("expected trimmed api key, got %q", cfg.APIKey)
	}
	if cfg.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected default base url: %q", cfg.BaseURL)
	}
}

func TestNewClientWithConfig(t *testing.T) {
	t.Run("default http client and normalized auth/base url", func(t *testing.T) {
		var gotAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			if r.URL.Path != "/chat/completions" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"choices":[]}`))
		}))
		defer server.Close()

		client := openai.NewClientWithConfig(openai.Config{
			APIKey:  "  token  ",
			BaseURL: " " + server.URL + "/ ",
		})

		_, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: "gpt-test",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "hi"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotAuth != "Bearer token" {
			t.Fatalf("unexpected authorization header: %q", gotAuth)
		}
	})

	t.Run("custom http client preserved", func(t *testing.T) {
		called := false
		customHTTP := &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				called = true
				if req.URL.String() != "https://example.com/v1/chat/completions" {
					t.Fatalf("unexpected request url: %s", req.URL.String())
				}
				if req.Header.Get("Authorization") != "Bearer k" {
					t.Fatalf("unexpected auth header: %q", req.Header.Get("Authorization"))
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(bytes.NewBufferString(`{"choices":[]}`)),
					Request:    req,
				}, nil
			}),
		}

		client := openai.NewClientWithConfig(openai.Config{
			APIKey:     "k",
			BaseURL:    "https://example.com/v1",
			HTTPClient: customHTTP,
		})

		_, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
			Model: "gpt-test",
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: "hi"},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatalf("expected custom transport to be called")
		}
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestChatMessagePartMarshalJSON(t *testing.T) {
	part := openai.ChatMessagePart{
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
	msg := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: "fallback text",
		MultiContent: []openai.ChatMessagePart{
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
