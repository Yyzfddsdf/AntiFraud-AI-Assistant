package multi_agent_test

import (
	"context"
	"encoding/json"
	"fmt"
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

type stubCaseCollectionToolHandlerFunc func(ctx context.Context, args string) (agenttool.ToolResponse, error)

func (f stubCaseCollectionToolHandlerFunc) Handle(ctx context.Context, args string) (agenttool.ToolResponse, error) {
	return f(ctx, args)
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
			if !strings.Contains(userContent, "最多可进行 1 轮 search_web 搜索") {
				t.Fatalf("expected user prompt to include search round limit, got %q", userContent)
			}
			if payload["tool_choice"] != "required" {
				t.Fatalf("expected required tool choice, got %#v", payload["tool_choice"])
			}

			tools, ok := payload["tools"].([]interface{})
			if !ok || len(tools) != 1 {
				t.Fatalf("unexpected tools payload: %#v", payload["tools"])
			}
			assertCaseCollectionExposedTool(t, tools[0], agenttool.WebSearchToolName)

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

		if payload["tool_choice"] != "required" {
			t.Fatalf("expected required tool choice, got %#v", payload["tool_choice"])
		}

		tools, ok := payload["tools"].([]interface{})
		if !ok || len(tools) != 1 {
			t.Fatalf("unexpected tools payload: %#v", payload["tools"])
		}
		assertCaseCollectionExposedTool(t, tools[0], agenttool.UploadHistoricalCaseToVectorDBToolName)

		if len(messages) != 2 {
			t.Fatalf("expected compressed upload-phase messages, got %#v", payload["messages"])
		}
		uploadUserContent, _ := messages[1].(map[string]interface{})["content"].(string)
		if !strings.Contains(uploadUserContent, "前置 search_web 阶段已经完成，共 1 轮") {
			t.Fatalf("expected upload prompt to mention compressed search phase, got %q", uploadUserContent)
		}
		if !strings.Contains(uploadUserContent, "案例1") {
			t.Fatalf("expected upload prompt to contain compressed search results, got %q", uploadUserContent)
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

func TestCaseCollectionAgentSearchRoundsFollowCaseCount(t *testing.T) {
	originalToolsProvider := caseCollectionToolsProvider
	originalHandlerResolver := caseCollectionToolHandlerResolver
	t.Cleanup(func() {
		caseCollectionToolsProvider = originalToolsProvider
		caseCollectionToolHandlerResolver = originalHandlerResolver
	})

	caseCollectionToolsProvider = func() []openai.Tool {
		return []openai.Tool{agenttool.WebSearchTool, agenttool.UploadHistoricalCaseToVectorDBTool}
	}

	uploadCount := 0
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
			return stubCaseCollectionToolHandlerFunc(func(ctx context.Context, args string) (agenttool.ToolResponse, error) {
				uploadCount++
				return agenttool.ToolResponse{Payload: map[string]interface{}{
					"status": "success",
					"review": map[string]interface{}{
						"record_id":  fmt.Sprintf("PREV-%03d", uploadCount),
						"user_id":    "99",
						"title":      fmt.Sprintf("冒充客服退款诈骗-%d", uploadCount),
						"risk_level": "高",
						"scam_type":  "冒充电商物流客服类",
						"created_at": "2026-03-14T20:00:00Z",
					},
				}}, nil
			})
		default:
			return nil
		}
	}

	requestCount := 0
	expectedChoices := []string{
		agenttool.WebSearchToolName,
		agenttool.WebSearchToolName,
		agenttool.UploadHistoricalCaseToVectorDBToolName,
		agenttool.UploadHistoricalCaseToVectorDBToolName,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		payload := map[string]interface{}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}

		requestCount++
		if requestCount > len(expectedChoices) {
			t.Fatalf("unexpected request count: %d", requestCount)
		}

		if payload["tool_choice"] != "required" {
			t.Fatalf("expected required tool choice, got %#v", payload["tool_choice"])
		}

		tools, ok := payload["tools"].([]interface{})
		if !ok || len(tools) != 1 {
			t.Fatalf("unexpected tools payload: %#v", payload["tools"])
		}
		assertCaseCollectionExposedTool(t, tools[0], expectedChoices[requestCount-1])
		if requestCount == 3 {
			messages, ok := payload["messages"].([]interface{})
			if !ok || len(messages) != 2 {
				t.Fatalf("expected compressed upload-phase messages on first upload round, got %#v", payload["messages"])
			}
			uploadUserContent, _ := messages[1].(map[string]interface{})["content"].(string)
			if !strings.Contains(uploadUserContent, "前置 search_web 阶段已经完成，共 2 轮") {
				t.Fatalf("expected upload prompt to mention compressed search phase, got %q", uploadUserContent)
			}
			if !strings.Contains(uploadUserContent, "案例1") {
				t.Fatalf("expected upload prompt to contain compressed search results, got %q", uploadUserContent)
			}
		}

		switch requestCount {
		case 1, 2:
			_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{{
					Message: openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleAssistant,
						ToolCalls: []openai.ToolCall{{
							ID:   fmt.Sprintf("call_search_%d", requestCount),
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      agenttool.WebSearchToolName,
								Arguments: `{"query":"冒充客服诈骗","max_results":3}`,
							},
						}},
					},
				}},
			})
		case 3, 4:
			_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{{
					Message: openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleAssistant,
						ToolCalls: []openai.ToolCall{{
							ID:   fmt.Sprintf("call_upload_%d", requestCount),
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      agenttool.UploadHistoricalCaseToVectorDBToolName,
								Arguments: `{"title":"冒充客服退款诈骗","target_group":"成年人","risk_level":"高","scam_type":"冒充电商物流客服类","case_description":"诈骗分子冒充客服，以退款理赔为由诱导受害人下载会议软件并转账。"}`,
							},
						}},
					},
				}},
			})
		}
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

	err := agent.CollectCases(context.Background(), "99", "冒充客服诈骗", 2)
	if err != nil {
		t.Fatalf("CollectCases failed: %v", err)
	}

	if requestCount != 4 {
		t.Fatalf("expected 4 model rounds for case_count=2, got %d", requestCount)
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

func TestCaseCollectionAgentAcceptsSearchSummaryWithoutStatus(t *testing.T) {
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
					"query":  "测试",
					"answer": "这是一个没有 status 的搜索摘要",
					"results": []map[string]interface{}{
						{"title": "案例A", "url": "https://example.com/a", "content": "摘要A"},
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
						"title":      "测试案件",
						"risk_level": "中",
						"scam_type":  "其他",
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
		payload := map[string]interface{}{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		requestCount++

		if requestCount == 2 {
			messages, ok := payload["messages"].([]interface{})
			if !ok || len(messages) != 2 {
				t.Fatalf("expected compressed upload-phase messages, got %#v", payload["messages"])
			}
			uploadUserContent, _ := messages[1].(map[string]interface{})["content"].(string)
			if !strings.Contains(uploadUserContent, "案例A") {
				t.Fatalf("expected upload prompt to contain compressed search results without status, got %q", uploadUserContent)
			}
		}

		if requestCount == 1 {
			_ = json.NewEncoder(w).Encode(openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{{
					Message: openai.ChatCompletionMessage{
						Role: openai.ChatMessageRoleAssistant,
						ToolCalls: []openai.ToolCall{{
							ID:   "call_search_1",
							Type: openai.ToolTypeFunction,
							Function: openai.FunctionCall{
								Name:      agenttool.WebSearchToolName,
								Arguments: `{"query":"测试","max_results":3}`,
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
							Arguments: `{"title":"测试案件","target_group":"成年人","risk_level":"中","scam_type":"其他","case_description":"测试描述"}`,
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

	if err := agent.CollectCases(context.Background(), "99", "测试", 1); err != nil {
		t.Fatalf("CollectCases failed: %v", err)
	}
}

func assertCaseCollectionExposedTool(t *testing.T, raw interface{}, expectedName string) {
	t.Helper()

	toolPayload, ok := raw.(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected tool payload: %#v", raw)
	}
	if toolPayload["type"] == openai.ToolTypeWebSearch {
		t.Fatalf("unexpected built-in web_search tool in case collection request: %#v", toolPayload)
	}

	functionPayload, ok := toolPayload["function"].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected tool function payload: %#v", toolPayload["function"])
	}
	if functionPayload["name"] != expectedName {
		t.Fatalf("unexpected exposed tool function name: got=%#v want=%q", functionPayload["name"], expectedName)
	}
}
