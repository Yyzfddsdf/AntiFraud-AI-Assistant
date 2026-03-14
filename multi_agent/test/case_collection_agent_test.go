package multi_agent_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"antifraud/config"
	openai "antifraud/llm"
	"antifraud/multi_agent"
	agenttool "antifraud/multi_agent/tool"
)

type stubCaseCollectionToolHandler struct {
	payload agenttool.ToolResponse
	err     error
}

func (h stubCaseCollectionToolHandler) Handle(ctx context.Context, args string) (agenttool.ToolResponse, error) {
	return h.payload, h.err
}

func TestCaseCollectionAgentCollectCasesSuccess(t *testing.T) {
	originalToolsProvider := caseCollectionToolsProvider
	originalHandlerResolver := caseCollectionToolHandlerResolver
	t.Cleanup(func() {
		caseCollectionToolsProvider = originalToolsProvider
		caseCollectionToolHandlerResolver = originalHandlerResolver
	})

	caseCollectionToolsProvider = func() []openai.Tool {
		return []openai.Tool{agenttool.WebSearchTool, agenttool.UploadHistoricalCaseToVectorDBTool}
	}
	caseCollectionToolHandlerResolver = func(name string) agenttool.ToolHandler {
		switch name {
		case agenttool.WebSearchToolName:
			return stubCaseCollectionToolHandler{
				payload: agenttool.ToolResponse{Payload: map[string]interface{}{
					"status": "success",
					"query":  "冒充客服诈骗",
					"results": []map[string]interface{}{
						{"title": "案例1", "url": "https://example.com/1", "content": "内容"},
					},
				}},
			}
		case agenttool.UploadHistoricalCaseToVectorDBToolName:
			return stubCaseCollectionToolHandler{
				payload: agenttool.ToolResponse{Payload: map[string]interface{}{
					"status": "success",
					"review": map[string]interface{}{
						"record_id":  "PREV-001",
						"user_id":    "99",
						"title":      "冒充客服退款诈骗",
						"risk_level": "高",
						"scam_type":  "冒充电商物流客服类",
						"status":     "pending_review",
						"created_at": "2026-03-14T20:00:00Z",
					},
				}},
			}
		default:
			return nil
		}
	}

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		payload := map[string]interface{}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		requestCount++

		messages, ok := payload["messages"].([]interface{})
		if !ok || len(messages) < 2 {
			t.Fatalf("unexpected messages payload: %#v", payload["messages"])
		}

		if requestCount == 1 {
			systemMessage := messages[0].(map[string]interface{})
			userMessage := messages[1].(map[string]interface{})
			if systemMessage["content"] != "固定系统提示词" {
				t.Fatalf("unexpected system prompt: %#v", systemMessage["content"])
			}
			userContent, _ := userMessage["content"].(string)
			if !strings.Contains(userContent, "提交 1 个待审核案件") {
				t.Fatalf("expected user prompt to include case count, got %q", userContent)
			}
			if payload["tool_choice"] != "required" {
				t.Fatalf("expected required tool choice, got %#v", payload["tool_choice"])
			}

			_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{{
					Message: openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleAssistant,
						ToolCalls: []openai.ToolCall{{
							ID:   "call_search_1",
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      agenttool.WebSearchToolName,
								Arguments: `{"query":"冒充客服诈骗","max_results":3}`,
							},
						}},
					},
				}},
			})
			return
		}

		_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{{
				Message: openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleAssistant,
					ToolCalls: []openai.ToolCall{{
						ID:   "call_upload_1",
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      agenttool.UploadHistoricalCaseToVectorDBToolName,
							Arguments: `{"title":"冒充客服退款诈骗","target_group":"成年人","risk_level":"高","scam_type":"冒充电商物流客服类","case_description":"诈骗分子冒充客服，以退款理赔为由诱导受害人下载会议软件并转账。"}`,
						},
					}},
				},
			}},
		})
	}))
	defer server.Close()

	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  "key",
		BaseURL: server.URL,
	})
	agent := multi_agent.NewCaseCollectionAgentWithClient(
		config.ModelConfig{Model: "deepseek-chat", APIKey: "key", BaseURL: server.URL, MaxTokens: 1024, TopP: 1, Temperature: 0.3},
		config.RetryConfig{MaxRetries: 1, RetryDelayMS: 1},
		"固定系统提示词",
		client,
	)

	err := agent.CollectCases(context.Background(), "99", "冒充客服诈骗", 1)
	if err != nil {
		t.Fatalf("CollectCases failed: %v", err)
	}

	if requestCount != 2 {
		t.Fatalf("expected 2 model rounds, got %d", requestCount)
	}
}

func TestCaseCollectionAgentRejectsInvalidCaseCount(t *testing.T) {
	client := openai.NewClientWithConfig(openai.Config{
		APIKey:  "key",
		BaseURL: "https://example.com",
	})
	agent := multi_agent.NewCaseCollectionAgentWithClient(
		config.ModelConfig{Model: "deepseek-chat", APIKey: "key", BaseURL: "https://example.com", MaxTokens: 1024, TopP: 1, Temperature: 0.3},
		config.RetryConfig{MaxRetries: 1, RetryDelayMS: 1},
		"固定系统提示词",
		client,
	)

	err := agent.CollectCases(context.Background(), "99", "冒充客服诈骗", 0)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "case_count") {
		t.Fatalf("unexpected error: %v", err)
	}
}
