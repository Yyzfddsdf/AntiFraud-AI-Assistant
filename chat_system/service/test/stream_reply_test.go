package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chatservice "antifraud/chat_system/service"
	appcfg "antifraud/config"
	openai "antifraud/llm"
)

func TestStreamReply_AllRoundsUseStreamMode(t *testing.T) {
	requestCount := 0
	streamFlags := make([]bool, 0, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		payload := decodeJSONBody(t, r)
		flag, ok := payload["stream"].(bool)
		if !ok {
			t.Fatalf("stream flag missing or invalid type: %#v", payload["stream"])
		}
		streamFlags = append(streamFlags, flag)

		if requestCount != 1 {
			t.Fatalf("unexpected request count: %d", requestCount)
		}

		writeSSEResponse(t, w,
			`{"choices":[{"delta":{"content":"你好"}}]}`,
			`{"choices":[{"delta":{"content":"，世界"}}]}`,
		)
	}))
	defer server.Close()

	cfg := &appcfg.ChatConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}
	svc := chatservice.NewChatService(cfg)

	events := make([]map[string]interface{}, 0, 3)
	reply, turnMessages, err := svc.StreamReply(
		context.Background(),
		"u1",
		"你好",
		[]openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "sys"},
			{Role: openai.ChatMessageRoleUser, Content: "你好"},
		},
		func(event map[string]interface{}) error {
			events = append(events, event)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("StreamReply failed: %v", err)
	}

	if reply != "你好，世界" {
		t.Fatalf("unexpected reply: %q", reply)
	}
	if requestCount != 1 {
		t.Fatalf("unexpected request count: %d", requestCount)
	}
	if len(streamFlags) != 1 || !streamFlags[0] {
		t.Fatalf("all requests should be stream mode, got: %#v", streamFlags)
	}
	assertEventTypes(t, events, []string{"content", "content", "done"})

	if len(turnMessages) != 2 {
		t.Fatalf("unexpected turn message count: %d", len(turnMessages))
	}
	if turnMessages[1].Role != openai.ChatMessageRoleAssistant || turnMessages[1].Content != "你好，世界" {
		t.Fatalf("unexpected final turn message: %+v", turnMessages[1])
	}
}

func TestStreamReply_StreamToolCallsRemainComplete(t *testing.T) {
	requestCount := 0
	streamFlags := make([]bool, 0, 2)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}

		payload := decodeJSONBody(t, r)
		flag, ok := payload["stream"].(bool)
		if !ok {
			t.Fatalf("stream flag missing or invalid type: %#v", payload["stream"])
		}
		streamFlags = append(streamFlags, flag)

		switch requestCount {
		case 1:
			writeSSEResponse(t, w,
				`{"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"unknown_tool"}}]}}]}`,
				`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{"}}]}}]}`,
				`{"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"}"}}]}}]}`,
			)
		case 2:
			writeSSEResponse(t, w,
				`{"choices":[{"delta":{"content":"已记录，建议立即报警并冻结支付渠道。"}}]}`,
			)
		default:
			t.Fatalf("unexpected extra round: %d", requestCount)
		}
	}))
	defer server.Close()

	cfg := &appcfg.ChatConfig{
		BaseURL: server.URL,
		Model:   "test-model",
	}
	svc := chatservice.NewChatService(cfg)

	events := make([]map[string]interface{}, 0, 6)
	reply, turnMessages, err := svc.StreamReply(
		context.Background(),
		"u1",
		"我可能被骗了",
		[]openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "sys"},
			{Role: openai.ChatMessageRoleUser, Content: "我可能被骗了"},
		},
		func(event map[string]interface{}) error {
			events = append(events, event)
			return nil
		},
	)
	if err != nil {
		t.Fatalf("StreamReply failed: %v", err)
	}

	if reply != "已记录，建议立即报警并冻结支付渠道。" {
		t.Fatalf("unexpected reply: %q", reply)
	}
	if requestCount != 2 {
		t.Fatalf("expected 2 rounds, got %d", requestCount)
	}
	for _, flag := range streamFlags {
		if !flag {
			t.Fatalf("found non-stream request in flags: %#v", streamFlags)
		}
	}

	if len(turnMessages) != 4 {
		t.Fatalf("unexpected turn message count: %d", len(turnMessages))
	}
	if len(turnMessages[1].ToolCalls) != 1 {
		t.Fatalf("assistant tool call message malformed: %+v", turnMessages[1])
	}
	if turnMessages[1].ToolCalls[0].Arguments != "{}" {
		t.Fatalf("tool arguments should be reconstructed as {}, got: %q", turnMessages[1].ToolCalls[0].Arguments)
	}
	if turnMessages[2].Role != openai.ChatMessageRoleTool || turnMessages[2].ToolCallID != "call_1" {
		t.Fatalf("unexpected tool turn message: %+v", turnMessages[2])
	}

	assertEventTypes(t, events, []string{"tool_call", "tool_result", "content", "done"})
	toolCallEvent := findEventByType(events, "tool_call")
	if toolCallEvent["id"] != "call_1" || toolCallEvent["tool"] != "unknown_tool" {
		t.Fatalf("unexpected tool_call event: %+v", toolCallEvent)
	}
	toolResultEvent := findEventByType(events, "tool_result")
	resultPayload, ok := toolResultEvent["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("tool_result payload type invalid: %#v", toolResultEvent["result"])
	}
	msg, _ := resultPayload["error"].(string)
	if !strings.Contains(msg, "unsupported tool: unknown_tool") {
		t.Fatalf("unexpected tool_result error: %+v", toolResultEvent)
	}
}

func TestBuildMessagesForUser_EmptySystemPromptReturnsError(t *testing.T) {
	_, err := chatservice.BuildMessagesForUser("   ", "u1", "你好")
	if err == nil {
		t.Fatal("expected error for empty system prompt")
	}
	if !strings.Contains(err.Error(), "chat system prompt is empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func decodeJSONBody(t *testing.T, r *http.Request) map[string]interface{} {
	t.Helper()
	defer r.Body.Close()

	payload := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		t.Fatalf("decode request body failed: %v", err)
	}
	return payload
}

func writeSSEResponse(t *testing.T, w http.ResponseWriter, chunks ...string) {
	t.Helper()
	w.Header().Set("Content-Type", "text/event-stream")

	flusher, ok := w.(http.Flusher)
	if !ok {
		t.Fatal("response writer is not flusher")
	}

	for _, chunk := range chunks {
		if _, err := fmt.Fprintf(w, "data: %s\n\n", chunk); err != nil {
			t.Fatalf("write sse chunk failed: %v", err)
		}
		flusher.Flush()
	}
	if _, err := fmt.Fprint(w, "data: [DONE]\n\n"); err != nil {
		t.Fatalf("write sse done failed: %v", err)
	}
	flusher.Flush()
}

func assertEventTypes(t *testing.T, events []map[string]interface{}, expected []string) {
	t.Helper()
	if len(events) != len(expected) {
		t.Fatalf("event length mismatch: got=%d expected=%d events=%+v", len(events), len(expected), events)
	}
	for idx, expectType := range expected {
		gotType, _ := events[idx]["type"].(string)
		if gotType != expectType {
			t.Fatalf("event[%d] type mismatch: got=%q expected=%q full=%+v", idx, gotType, expectType, events[idx])
		}
	}
}

func findEventByType(events []map[string]interface{}, eventType string) map[string]interface{} {
	for _, item := range events {
		if itemType, _ := item["type"].(string); itemType == eventType {
			return item
		}
	}
	return map[string]interface{}{}
}
