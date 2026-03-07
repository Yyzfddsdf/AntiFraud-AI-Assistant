package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"antifraud/cache"
	chattool "antifraud/chat_system/tool"
	appcfg "antifraud/config"

	openai "antifraud/llm"
)

const conversationTTL = 5 * time.Minute

const responsesMessageRoleDeveloper = "developer"

// ConversationToolCall 是对工具调用的持久化表示。
type ConversationToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ConversationMessage 是会话上下文在 Redis 中的序列化结构。
type ConversationMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	ImageURLs  []string               `json:"image_urls,omitempty"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	ToolCalls  []ConversationToolCall `json:"tool_calls,omitempty"`
}

// ChatService 封装聊天模型客户端与模型 ID。
type ChatService struct {
	client *openai.Client
	model  string
}

// NewChatService 根据聊天配置创建服务实例。
func NewChatService(cfg *appcfg.ChatConfig) *ChatService {
	modelID := ""
	apiKey := ""
	baseURL := ""
	if cfg != nil {
		modelID = strings.TrimSpace(cfg.Model)
		apiKey = strings.TrimSpace(cfg.APIKey)
		baseURL = strings.TrimSpace(cfg.BaseURL)
	}
	if modelID == "" {
		modelID = "gpt-5.4"
	}

	return &ChatService{
		client: openai.NewClient(apiKey, baseURL),
		model:  modelID,
	}
}

// BuildMessagesForUser 组装最终请求消息：
// 系统提示词 + Redis 历史上下文 + 当前用户输入。
func BuildMessagesForUser(systemPrompt string, userID string, currentUserInput string, currentUserImageURLs []string) ([]openai.ChatCompletionMessage, error) {
	trimmedSystemPrompt := strings.TrimSpace(systemPrompt)
	if trimmedSystemPrompt == "" {
		return nil, fmt.Errorf("chat system prompt is empty")
	}

	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: trimmedSystemPrompt,
		},
	}

	key := conversationKey(trimmedUserID)
	history := make([]ConversationMessage, 0)
	found, err := cache.GetJSON(key, &history)
	if err != nil {
		return nil, fmt.Errorf("load conversation from redis failed: %w", err)
	}
	if found {
		// 读取历史后刷新 TTL，保持活跃会话不被过早回收。
		if err := cache.SetJSON(key, history, conversationTTL); err != nil {
			log.Printf("[chat] refresh conversation ttl failed: user=%s err=%v", trimmedUserID, err)
		}
	}

	for _, item := range history {
		msg, ok := conversationToOpenAIMessage(item)
		if !ok {
			continue
		}
		messages = append(messages, msg)
	}

	currentMessage, ok := buildUserChatMessage(currentUserInput, currentUserImageURLs)
	if !ok {
		return nil, fmt.Errorf("chat message and images are both empty")
	}
	messages = append(messages, currentMessage)

	return messages, nil
}

// StreamReply 使用全流式回合处理：首轮即 stream=true，边接收边向前端推送内容；
// 若出现 tool_calls，则在参数拼接完成后执行工具并继续下一轮流式请求。
func (s *ChatService) StreamReply(ctx context.Context, userID string, userInput string, userImageURLs []string, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) (string, []ConversationMessage, error) {
	responseInput := chatMessagesToResponsesInput(messages)
	recorded := make([]ConversationMessage, 0)
	normalizedUserImageURLs := normalizeImageURLs(userImageURLs)

	const maxRounds = 8
	for round := 0; round < maxRounds; round++ {
		roundResult, err := s.streamAssistantRound(ctx, responseInput, emit)
		if err != nil {
			return "", nil, err
		}

		assistantToolCalls := responseFunctionCallsToOpenAIToolCalls(roundResult.FunctionCalls)
		assistantMessage := openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   roundResult.Content,
			ToolCalls: assistantToolCalls,
		}

		if len(assistantMessage.ToolCalls) == 0 {
			if emit != nil {
				if err := emit(map[string]interface{}{"type": "done", "reason": "stop"}); err != nil {
					return "", nil, err
				}
			}

			finalReply := strings.TrimSpace(assistantMessage.Content)
			turnMessages := make([]ConversationMessage, 0, len(recorded)+2)
			turnMessages = append(turnMessages, ConversationMessage{
				Role:      openai.ChatMessageRoleUser,
				Content:   strings.TrimSpace(userInput),
				ImageURLs: normalizedUserImageURLs,
			})
			turnMessages = append(turnMessages, recorded...)
			turnMessages = append(turnMessages, ConversationMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: finalReply,
			})
			return finalReply, turnMessages, nil
		}

		recorded = append(recorded, ConversationMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   strings.TrimSpace(assistantMessage.Content),
			ToolCalls: openAIToolCallsToConversation(assistantMessage.ToolCalls),
		})

		if strings.TrimSpace(assistantMessage.Content) != "" {
			responseInput = append(responseInput, newResponsesAssistantMessage(assistantMessage.Content))
		}

		for _, call := range roundResult.FunctionCalls {
			toolCall := responseFunctionCallToOpenAIToolCall(call)
			if strings.TrimSpace(toolCall.ID) == "" || strings.TrimSpace(toolCall.Function.Name) == "" {
				continue
			}

			toolPayload := s.handleToolCall(ctx, userID, toolCall)
			if emit != nil {
				_ = emit(map[string]interface{}{
					"type": "tool_call",
					"tool": toolCall.Function.Name,
					"id":   toolCall.ID,
				})
				_ = emit(map[string]interface{}{
					"type": "tool_result",
					"tool": toolCall.Function.Name,
					"id":   toolCall.ID,
				})
			}

			payloadBytes, _ := json.Marshal(toolPayload)
			responseInput = append(responseInput,
				normalizeResponseFunctionCall(call),
				openai.ResponseFunctionCallOutput{
					Type:   "function_call_output",
					CallID: toolCall.ID,
					Output: string(payloadBytes),
				},
			)
			recorded = append(recorded, ConversationMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
				Content:    string(payloadBytes),
			})
		}
	}

	return "", nil, fmt.Errorf("resolve tool calls exceeded max rounds: %d", maxRounds)
}

type responseRoundResult struct {
	Content       string
	FunctionCalls []openai.ResponseFunctionCall
}

func (s *ChatService) streamAssistantRound(ctx context.Context, input []any, emit func(event map[string]interface{}) error) (responseRoundResult, error) {
	stream, err := s.client.StreamResponse(ctx, openai.ResponsesRequest{
		Model: s.model,
		Input: input,
		Tools: responseToolsFromChatTools(chattool.ChatTools()),
	})
	if err != nil {
		return responseRoundResult{}, fmt.Errorf("create responses stream failed: %w", err)
	}
	defer stream.Close()

	var contentBuilder strings.Builder
	outputItems := make([]map[string]any, 0)
	var completedResponse map[string]any

	for stream.Next() {
		eventType, payload, parseErr := parseResponseStreamEvent(stream.Current())
		if parseErr != nil {
			return responseRoundResult{}, fmt.Errorf("parse responses stream event failed: %w", parseErr)
		}

		switch eventType {
		case "response.output_text.delta":
			delta, _ := payload["delta"].(string)
			if delta == "" {
				continue
			}
			contentBuilder.WriteString(delta)
			if emit != nil {
				if err := emit(map[string]interface{}{
					"type":    "content",
					"content": delta,
				}); err != nil {
					return responseRoundResult{}, err
				}
			}
		case "response.output_item.done":
			item, ok := payload["item"].(map[string]any)
			if ok {
				outputItems = append(outputItems, item)
			}
		case "response.completed":
			response, ok := payload["response"].(map[string]any)
			if ok {
				completedResponse = response
			}
		}
	}
	if err := stream.Err(); err != nil {
		return responseRoundResult{}, fmt.Errorf("recv responses stream failed: %w", err)
	}

	content, functionCalls := parseResponsesOutput(completedResponse, outputItems, contentBuilder.String())
	return responseRoundResult{
		Content:       content,
		FunctionCalls: functionCalls,
	}, nil
}

func (s *ChatService) handleToolCall(ctx context.Context, userID string, call openai.ToolCall) map[string]interface{} {
	toolPayload := map[string]interface{}{
		"error": fmt.Sprintf("unsupported tool: %s", call.Function.Name),
	}

	handler := chattool.GetChatToolHandler(call.Function.Name)
	if handler == nil {
		return toolPayload
	}

	result, err := handler.Handle(ctx, userID, call.Function.Arguments)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return result.Payload
}

func responseToolsFromChatTools(tools []openai.Tool) []any {
	result := make([]any, 0, len(tools)+1)
	result = append(result, openai.WebSearchTool{
		Type:              "web_search",
		SearchContextSize: "medium",
	})

	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		result = append(result, openai.FunctionTool{
			Type:        openai.ToolTypeFunction,
			Name:        strings.TrimSpace(tool.Function.Name),
			Description: strings.TrimSpace(tool.Function.Description),
			Parameters:  tool.Function.Parameters,
		})
	}
	return result
}

func chatMessagesToResponsesInput(messages []openai.ChatCompletionMessage) []any {
	result := make([]any, 0, len(messages))
	for _, message := range messages {
		role := strings.TrimSpace(message.Role)
		switch role {
		case openai.ChatMessageRoleSystem:
			if strings.TrimSpace(message.Content) != "" {
				result = append(result, newResponsesInputMessage(responsesMessageRoleDeveloper, message.Content))
			}
		case openai.ChatMessageRoleUser:
			if responseMessage, ok := newResponsesUserMessage(message); ok {
				result = append(result, responseMessage)
			}
		case openai.ChatMessageRoleAssistant:
			if strings.TrimSpace(message.Content) != "" {
				result = append(result, newResponsesAssistantMessage(message.Content))
			}
			for _, toolCall := range message.ToolCalls {
				if call, ok := openAIToolCallToResponseFunctionCall(toolCall); ok {
					result = append(result, call)
				}
			}
		case openai.ChatMessageRoleTool:
			if strings.TrimSpace(message.ToolCallID) == "" {
				continue
			}
			result = append(result, openai.ResponseFunctionCallOutput{
				Type:   "function_call_output",
				CallID: strings.TrimSpace(message.ToolCallID),
				Output: message.Content,
			})
		}
	}
	return result
}

func newResponsesInputMessage(role string, text string) openai.Message {
	trimmedText := strings.TrimSpace(text)
	return openai.Message{
		Type: "message",
		Role: role,
		Content: []any{
			openai.InputText{
				Type: "input_text",
				Text: trimmedText,
			},
		},
	}
}

func newResponsesUserMessage(message openai.ChatCompletionMessage) (openai.Message, bool) {
	content := make([]any, 0, len(message.MultiContent)+1)
	if len(message.MultiContent) > 0 {
		for _, part := range message.MultiContent {
			switch strings.TrimSpace(part.Type) {
			case "text":
				trimmedText := strings.TrimSpace(part.Text)
				if trimmedText == "" {
					continue
				}
				content = append(content, openai.InputText{
					Type: "input_text",
					Text: trimmedText,
				})
			case "image_url":
				if part.ImageURL == nil {
					continue
				}
				imageURL := strings.TrimSpace(part.ImageURL.URL)
				if imageURL == "" {
					continue
				}
				content = append(content, openai.InputImage{
					Type:     "input_image",
					ImageURL: imageURL,
				})
			}
		}
	} else {
		trimmedText := strings.TrimSpace(message.Content)
		if trimmedText != "" {
			content = append(content, openai.InputText{
				Type: "input_text",
				Text: trimmedText,
			})
		}
	}

	if len(content) == 0 {
		return openai.Message{}, false
	}

	return openai.Message{
		Type:    "message",
		Role:    openai.ChatMessageRoleUser,
		Content: content,
	}, true
}

func newResponsesAssistantMessage(text string) openai.Message {
	return openai.Message{
		Type: "message",
		Role: openai.ChatMessageRoleAssistant,
		Content: []any{
			openai.OutputText{
				Type: "output_text",
				Text: text,
			},
		},
	}
}

func openAIToolCallToResponseFunctionCall(call openai.ToolCall) (openai.ResponseFunctionCall, bool) {
	id := strings.TrimSpace(call.ID)
	name := strings.TrimSpace(call.Function.Name)
	if id == "" || name == "" {
		return openai.ResponseFunctionCall{}, false
	}
	return openai.ResponseFunctionCall{
		Type:      "function_call",
		CallID:    id,
		Name:      name,
		Arguments: call.Function.Arguments,
	}, true
}

func normalizeResponseFunctionCall(call openai.ResponseFunctionCall) openai.ResponseFunctionCall {
	call.Type = "function_call"
	call.CallID = strings.TrimSpace(call.CallID)
	call.Name = strings.TrimSpace(call.Name)
	return call
}

func responseFunctionCallsToOpenAIToolCalls(items []openai.ResponseFunctionCall) []openai.ToolCall {
	result := make([]openai.ToolCall, 0, len(items))
	for _, item := range items {
		toolCall := responseFunctionCallToOpenAIToolCall(item)
		if strings.TrimSpace(toolCall.ID) == "" || strings.TrimSpace(toolCall.Function.Name) == "" {
			continue
		}
		result = append(result, toolCall)
	}
	return result
}

func responseFunctionCallToOpenAIToolCall(call openai.ResponseFunctionCall) openai.ToolCall {
	normalized := normalizeResponseFunctionCall(call)
	return openai.ToolCall{
		ID:   normalized.CallID,
		Type: openai.ToolTypeFunction,
		Function: openai.FunctionCall{
			Name:      normalized.Name,
			Arguments: normalized.Arguments,
		},
	}
}

func parseResponseStreamEvent(event openai.StreamEvent) (string, map[string]any, error) {
	payload := map[string]any{}
	if len(event.Data) > 0 {
		if err := json.Unmarshal(event.Data, &payload); err != nil {
			return "", nil, err
		}
	}

	eventType := strings.TrimSpace(event.Event)
	if eventType == "" {
		if payloadType, _ := payload["type"].(string); payloadType != "" {
			eventType = payloadType
		}
	}
	return eventType, payload, nil
}

func parseResponsesOutput(completedResponse map[string]any, outputItems []map[string]any, streamedText string) (string, []openai.ResponseFunctionCall) {
	if completedResponse != nil {
		text, calls := parseResponsesPayload(completedResponse)
		if text != "" || len(calls) > 0 {
			if text == "" {
				text = streamedText
			}
			return text, calls
		}
	}

	text, calls := parseResponsesOutputItems(outputItems)
	if text == "" {
		text = streamedText
	}
	return text, calls
}

func parseResponsesPayload(payload map[string]any) (string, []openai.ResponseFunctionCall) {
	if payload == nil {
		return "", nil
	}

	text, _ := payload["output_text"].(string)
	outputItems, _ := payload["output"].([]any)
	parsedText, calls := parseResponsesAnyOutputItems(outputItems)
	if text == "" {
		text = parsedText
	}
	return text, calls
}

func parseResponsesOutputItems(items []map[string]any) (string, []openai.ResponseFunctionCall) {
	wrapped := make([]any, 0, len(items))
	for _, item := range items {
		wrapped = append(wrapped, item)
	}
	return parseResponsesAnyOutputItems(wrapped)
}

func parseResponsesAnyOutputItems(items []any) (string, []openai.ResponseFunctionCall) {
	var textBuilder strings.Builder
	functionCalls := make([]openai.ResponseFunctionCall, 0)

	for _, rawItem := range items {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}

		switch strings.TrimSpace(stringValue(item["type"])) {
		case "message":
			textBuilder.WriteString(parseResponsesMessageText(item))
		case "function_call":
			callID := strings.TrimSpace(stringValue(item["call_id"]))
			name := strings.TrimSpace(stringValue(item["name"]))
			if callID == "" || name == "" {
				continue
			}
			functionCalls = append(functionCalls, openai.ResponseFunctionCall{
				Type:      "function_call",
				CallID:    callID,
				Name:      name,
				Arguments: stringValue(item["arguments"]),
			})
		case "web_search_call":
			continue
		}
	}

	return textBuilder.String(), functionCalls
}

func buildUserChatMessage(text string, imageURLs []string) (openai.ChatCompletionMessage, bool) {
	trimmedText := strings.TrimSpace(text)
	normalizedImageURLs := normalizeImageURLs(imageURLs)
	if trimmedText == "" && len(normalizedImageURLs) == 0 {
		return openai.ChatCompletionMessage{}, false
	}

	message := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: trimmedText,
	}
	if len(normalizedImageURLs) == 0 {
		return message, true
	}

	multiContent := make([]openai.ChatMessagePart, 0, len(normalizedImageURLs)+1)
	if trimmedText != "" {
		multiContent = append(multiContent, openai.ChatMessagePart{
			Type: "text",
			Text: trimmedText,
		})
	}
	for _, imageURL := range normalizedImageURLs {
		multiContent = append(multiContent, openai.ChatMessagePart{
			Type:     "image_url",
			ImageURL: &openai.ChatMessageImageURL{URL: imageURL},
		})
	}
	message.MultiContent = multiContent
	return message, true
}

func normalizeImageURLs(imageURLs []string) []string {
	result := make([]string, 0, len(imageURLs))
	seen := map[string]struct{}{}
	for _, imageURL := range imageURLs {
		trimmed := strings.TrimSpace(imageURL)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func parseResponsesMessageText(item map[string]any) string {
	contentItems, _ := item["content"].([]any)
	var textBuilder strings.Builder
	for _, rawContent := range contentItems {
		contentItem, ok := rawContent.(map[string]any)
		if !ok {
			continue
		}
		switch strings.TrimSpace(stringValue(contentItem["type"])) {
		case "output_text":
			textBuilder.WriteString(stringValue(contentItem["text"]))
		case "refusal":
			textBuilder.WriteString(stringValue(contentItem["refusal"]))
		}
	}
	return textBuilder.String()
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}

// PersistConversation 将本轮新增消息追加写入 Redis，并重置 TTL。
func PersistConversation(userID string, newMessages []ConversationMessage) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}
	if len(newMessages) == 0 {
		return nil
	}

	key := conversationKey(trimmedUserID)

	history := make([]ConversationMessage, 0)
	_, err := cache.GetJSON(key, &history)
	if err != nil {
		return fmt.Errorf("load conversation before persist failed: %w", err)
	}

	history = append(history, sanitizeConversationMessages(newMessages)...)

	if err := cache.SetJSON(key, history, conversationTTL); err != nil {
		return fmt.Errorf("save conversation to redis failed: %w", err)
	}
	return nil
}

// ClearConversation 清空指定用户会话上下文。
func ClearConversation(userID string) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	key := conversationKey(trimmedUserID)
	if err := cache.Delete(key); err != nil {
		return fmt.Errorf("clear conversation from redis failed: %w", err)
	}

	return nil
}

// GetConversationContext 查询上下文、剩余 TTL 和上下文是否存在。
func GetConversationContext(userID string) ([]ConversationMessage, int64, bool, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	key := conversationKey(trimmedUserID)

	history := make([]ConversationMessage, 0)
	found, err := cache.GetJSON(key, &history)
	if err != nil {
		return nil, 0, false, fmt.Errorf("load conversation context failed: %w", err)
	}
	if !found {
		return []ConversationMessage{}, 0, false, nil
	}

	ttl, err := cache.TTL(key)
	if err != nil {
		return nil, 0, true, fmt.Errorf("read conversation ttl failed: %w", err)
	}

	ttlSeconds := int64(ttl / time.Second)
	if ttlSeconds < 0 {
		ttlSeconds = 0
	}

	return history, ttlSeconds, true, nil
}

// conversationKey 生成用户会话在 Redis 中的 key。
func conversationKey(userID string) string {
	return "chat:context:" + strings.TrimSpace(userID)
}

// sanitizeConversationMessages 过滤非法角色并裁剪内容。
func sanitizeConversationMessages(items []ConversationMessage) []ConversationMessage {
	result := make([]ConversationMessage, 0, len(items))
	for _, item := range items {
		role := strings.TrimSpace(item.Role)
		if role != openai.ChatMessageRoleUser && role != openai.ChatMessageRoleAssistant && role != openai.ChatMessageRoleTool {
			continue
		}
		trimmed := ConversationMessage{
			Role:       role,
			Content:    strings.TrimSpace(item.Content),
			ImageURLs:  normalizeImageURLs(item.ImageURLs),
			ToolCallID: strings.TrimSpace(item.ToolCallID),
			ToolCalls:  append([]ConversationToolCall{}, item.ToolCalls...),
		}
		result = append(result, trimmed)
	}
	return result
}

// openAIToolCallsToConversation 将 SDK 工具调用结构转换为持久化结构。
func openAIToolCallsToConversation(items []openai.ToolCall) []ConversationToolCall {
	result := make([]ConversationToolCall, 0, len(items))
	for _, item := range items {
		name := ""
		arguments := ""
		if item.Function.Name != "" {
			name = strings.TrimSpace(item.Function.Name)
		}
		if item.Function.Arguments != "" {
			arguments = item.Function.Arguments
		}
		result = append(result, ConversationToolCall{
			ID:        strings.TrimSpace(item.ID),
			Name:      name,
			Arguments: arguments,
		})
	}
	return result
}

// conversationToolCallsToOpenAI 将持久化结构转换回 SDK 工具调用结构。
func conversationToolCallsToOpenAI(items []ConversationToolCall) []openai.ToolCall {
	result := make([]openai.ToolCall, 0, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		name := strings.TrimSpace(item.Name)
		if id == "" || name == "" {
			continue
		}
		result = append(result, openai.ToolCall{
			ID:   id,
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      name,
				Arguments: item.Arguments,
			},
		})
	}
	return result
}

// conversationToOpenAIMessage 把会话记录恢复为模型请求消息。
func conversationToOpenAIMessage(item ConversationMessage) (openai.ChatCompletionMessage, bool) {
	role := strings.TrimSpace(item.Role)
	switch role {
	case openai.ChatMessageRoleUser:
		return buildUserChatMessage(item.Content, item.ImageURLs)
	case openai.ChatMessageRoleAssistant:
		msg := openai.ChatCompletionMessage{Role: role, Content: item.Content}
		if len(item.ToolCalls) > 0 {
			msg.ToolCalls = conversationToolCallsToOpenAI(item.ToolCalls)
		}
		return msg, true
	case openai.ChatMessageRoleTool:
		if strings.TrimSpace(item.ToolCallID) == "" {
			return openai.ChatCompletionMessage{}, false
		}
		return openai.ChatCompletionMessage{Role: role, ToolCallID: item.ToolCallID, Content: item.Content}, true
	default:
		return openai.ChatCompletionMessage{}, false
	}
}

// EncodeEvent 将 SSE 事件负载编码为 JSON 字符串。
func EncodeEvent(event map[string]interface{}) string {
	data, _ := json.Marshal(event)
	return string(data)
}
