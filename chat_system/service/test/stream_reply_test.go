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
        if r.URL.Path != "/responses" {
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

        input := mustArray(t, payload, "input")
        if len(input) != 2 {
            t.Fatalf("unexpected input count: %d", len(input))
        }
        tools := mustArray(t, payload, "tools")
        if len(tools) == 0 {
            t.Fatalf("expected responses tools, got empty list")
        }
        firstTool := mustObject(t, tools[0])
        if firstTool["type"] != "web_search" {
            t.Fatalf("expected first tool web_search, got: %#v", firstTool)
        }
        firstMessage := mustObject(t, input[0])
        if firstMessage["role"] != "developer" {
            t.Fatalf("expected developer role for prompt, got: %#v", firstMessage["role"])
        }

        writeSSEResponse(t, w,
            `{"type":"response.output_text.delta","delta":"Hello"}`,
            `{"type":"response.output_text.delta","delta":" world"}`,
        )
    }))
    defer server.Close()

    cfg := &appcfg.ChatConfig{BaseURL: server.URL, Model: "test-model"}
    svc := chatservice.NewChatService(cfg)

    events := make([]map[string]interface{}, 0, 3)
    reply, turnMessages, err := svc.StreamReply(
        context.Background(),
        "u1",
        "hello",
        nil,
        []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleSystem, Content: "sys"},
            {Role: openai.ChatMessageRoleUser, Content: "hello"},
        },
        func(event map[string]interface{}) error {
            events = append(events, event)
            return nil
        },
    )
    if err != nil {
        t.Fatalf("StreamReply failed: %v", err)
    }

    if reply != "Hello world" {
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
    if turnMessages[0].Role != openai.ChatMessageRoleUser || turnMessages[0].Content != "hello" {
        t.Fatalf("unexpected user turn message: %+v", turnMessages[0])
    }
    if turnMessages[1].Role != openai.ChatMessageRoleAssistant || turnMessages[1].Content != "Hello world" {
        t.Fatalf("unexpected final turn message: %+v", turnMessages[1])
    }
}

func TestStreamReply_StreamToolCallsRemainComplete(t *testing.T) {
    requestCount := 0
    streamFlags := make([]bool, 0, 2)

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestCount++
        if r.URL.Path != "/responses" {
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
            input := mustArray(t, payload, "input")
            if len(input) != 2 {
                t.Fatalf("unexpected first-round input: %#v", payload["input"])
            }
            tools := mustArray(t, payload, "tools")
            if len(tools) == 0 {
                t.Fatalf("expected responses tools, got empty list")
            }
            firstTool := mustObject(t, tools[0])
            if firstTool["type"] != "web_search" {
                t.Fatalf("expected first tool web_search, got: %#v", firstTool)
            }
            firstMessage := mustObject(t, input[0])
            if firstMessage["role"] != "developer" {
                t.Fatalf("expected first-round developer prompt, got: %#v", firstMessage)
            }

            writeSSEResponse(t, w,
                `{"type":"response.output_item.done","item":{"type":"function_call","call_id":"call_1","name":"unknown_tool","arguments":"{}"}}`,
            )
        case 2:
            input := mustArray(t, payload, "input")
            if len(input) != 4 {
                t.Fatalf("unexpected second-round input: %#v", payload["input"])
            }
            functionCallItem := mustObject(t, input[2])
            if functionCallItem["type"] != "function_call" || functionCallItem["call_id"] != "call_1" {
                t.Fatalf("unexpected function_call continuation item: %#v", functionCallItem)
            }
            functionOutputItem := mustObject(t, input[3])
            if functionOutputItem["type"] != "function_call_output" || functionOutputItem["call_id"] != "call_1" {
                t.Fatalf("unexpected function_call_output continuation item: %#v", functionOutputItem)
            }

            writeSSEResponse(t, w,
                `{"type":"response.output_text.delta","delta":"Recorded. Call police immediately."}`,
            )
        default:
            t.Fatalf("unexpected extra round: %d", requestCount)
        }
    }))
    defer server.Close()

    cfg := &appcfg.ChatConfig{BaseURL: server.URL, Model: "test-model"}
    svc := chatservice.NewChatService(cfg)

    events := make([]map[string]interface{}, 0, 6)
    reply, turnMessages, err := svc.StreamReply(
        context.Background(),
        "u1",
        "I may be scammed",
        nil,
        []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleSystem, Content: "sys"},
            {Role: openai.ChatMessageRoleUser, Content: "I may be scammed"},
        },
        func(event map[string]interface{}) error {
            events = append(events, event)
            return nil
        },
    )
    if err != nil {
        t.Fatalf("StreamReply failed: %v", err)
    }

    if reply != "Recorded. Call police immediately." {
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

func TestStreamReply_IncludesInputImagesInResponsesRequest(t *testing.T) {
    image1 := "data:image/png;base64,AAA111"
    image2 := "data:image/jpeg;base64,BBB222"

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/responses" {
            t.Fatalf("unexpected path: %s", r.URL.Path)
        }

        payload := decodeJSONBody(t, r)
        input := mustArray(t, payload, "input")
        if len(input) != 2 {
            t.Fatalf("unexpected input count: %d", len(input))
        }

        userMessage := mustObject(t, input[1])
        if userMessage["role"] != "user" {
            t.Fatalf("expected user role, got: %#v", userMessage["role"])
        }
        content := mustObjectArray(t, userMessage, "content")
        if len(content) != 3 {
            t.Fatalf("expected one text item and two image items, got: %#v", userMessage["content"])
        }
        if content[0]["type"] != "input_text" || content[0]["text"] != "Describe these images" {
            t.Fatalf("unexpected first content item: %#v", content[0])
        }
        if content[1]["type"] != "input_image" || content[1]["image_url"] != image1 {
            t.Fatalf("unexpected second content item: %#v", content[1])
        }
        if content[2]["type"] != "input_image" || content[2]["image_url"] != image2 {
            t.Fatalf("unexpected third content item: %#v", content[2])
        }

        writeSSEResponse(t, w,
            `{"type":"response.output_text.delta","delta":"Images received."}`,
        )
    }))
    defer server.Close()

    cfg := &appcfg.ChatConfig{BaseURL: server.URL, Model: "test-model"}
    svc := chatservice.NewChatService(cfg)

    reply, turnMessages, err := svc.StreamReply(
        context.Background(),
        "u1",
        "Describe these images",
        []string{image1, image2},
        []openai.ChatCompletionMessage{
            {Role: openai.ChatMessageRoleSystem, Content: "sys"},
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Describe these images",
                MultiContent: []openai.ChatMessagePart{
                    {Type: "text", Text: "Describe these images"},
                    {Type: "image_url", ImageURL: &openai.ChatMessageImageURL{URL: image1}},
                    {Type: "image_url", ImageURL: &openai.ChatMessageImageURL{URL: image2}},
                },
            },
        },
        nil,
    )
    if err != nil {
        t.Fatalf("StreamReply failed: %v", err)
    }
    if reply != "Images received." {
        t.Fatalf("unexpected reply: %q", reply)
    }
    if len(turnMessages) != 2 {
        t.Fatalf("unexpected turn message count: %d", len(turnMessages))
    }
    if len(turnMessages[0].ImageURLs) != 2 || turnMessages[0].ImageURLs[0] != image1 || turnMessages[0].ImageURLs[1] != image2 {
        t.Fatalf("expected image urls to be stored on user turn, got: %+v", turnMessages[0])
    }
}

func TestBuildMessagesForUser_EmptySystemPromptReturnsError(t *testing.T) {
    _, err := chatservice.BuildMessagesForUser("   ", "u1", "hello", nil)
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

func mustArray(t *testing.T, payload map[string]interface{}, key string) []interface{} {
    t.Helper()
    value, ok := payload[key].([]interface{})
    if !ok {
        t.Fatalf("expected array field %q, got: %#v", key, payload[key])
    }
    return value
}

func mustObject(t *testing.T, value interface{}) map[string]interface{} {
    t.Helper()
    obj, ok := value.(map[string]interface{})
    if !ok {
        t.Fatalf("expected object, got: %#v", value)
    }
    return obj
}

func mustObjectArray(t *testing.T, payload map[string]interface{}, key string) []map[string]interface{} {
    t.Helper()
    raw := mustArray(t, payload, key)
    result := make([]map[string]interface{}, 0, len(raw))
    for _, item := range raw {
        result = append(result, mustObject(t, item))
    }
    return result
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
