package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	chattool "antifraud/internal/modules/chat/adapters/outbound/tool"
	"antifraud/internal/platform/cache"
	appcfg "antifraud/internal/platform/config"

	openai "antifraud/internal/platform/llm"
)

const conversationTTL = 5 * time.Minute

const DefaultConversationKeyPrefix = "chat:context:"
const AdminConversationKeyPrefix = "admin:chat:context:"

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
	client              *openai.Client
	model               string
	toolProvider        func() []openai.Tool
	toolHandlerResolver func(string) chattool.ChatToolHandler
}

// NewChatService 根据聊天配置创建服务实例。
func NewChatService(cfg *appcfg.ChatConfig) *ChatService {
	return newChatService(cfg, chattool.ChatTools, chattool.GetChatToolHandler)
}

// NewAdminChatService 根据聊天配置创建管理员聊天服务实例。
func NewAdminChatService(cfg *appcfg.ChatConfig) *ChatService {
	return newChatService(cfg, chattool.AdminChatTools, chattool.GetAdminChatToolHandler)
}

func newChatService(cfg *appcfg.ChatConfig, toolProvider func() []openai.Tool, toolHandlerResolver func(string) chattool.ChatToolHandler) *ChatService {
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
		client:              openai.NewClient(apiKey, baseURL),
		model:               modelID,
		toolProvider:        toolProvider,
		toolHandlerResolver: toolHandlerResolver,
	}
}

// BuildMessagesForUser 组装最终请求消息：
// 系统提示词 + Redis 历史上下文 + 当前用户输入。
func BuildMessagesForUser(systemPrompt string, userID string, currentUserInput string, currentUserImageURLs []string) ([]openai.ChatCompletionMessage, error) {
	return BuildMessagesForUserWithPrefix(DefaultConversationKeyPrefix, systemPrompt, userID, currentUserInput, currentUserImageURLs)
}

func BuildMessagesForUserWithPrefix(conversationKeyPrefix string, systemPrompt string, userID string, currentUserInput string, currentUserImageURLs []string) ([]openai.ChatCompletionMessage, error) {
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

	key := conversationKey(conversationKeyPrefix, trimmedUserID)
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
	resolved := append([]openai.ChatCompletionMessage{}, messages...)
	recorded := make([]ConversationMessage, 0)

	const maxRounds = 8
	for round := 0; round < maxRounds; round++ {
		assistantMessage, err := s.streamAssistantRound(ctx, resolved, emit)
		if err != nil {
			return "", nil, err
		}
		resolved = append(resolved, assistantMessage)

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
				ImageURLs: normalizeImageURLs(userImageURLs),
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

		for _, call := range assistantMessage.ToolCalls {
			toolPayload := s.handleToolCall(ctx, userID, call)
			if emit != nil {
				_ = emit(map[string]interface{}{
					"type": "tool_call",
					"tool": call.Function.Name,
					"id":   call.ID,
				})
				_ = emit(map[string]interface{}{
					"type":   "tool_result",
					"tool":   call.Function.Name,
					"id":     call.ID,
					"result": toolPayload,
				})
			}

			payloadBytes, _ := json.Marshal(toolPayload)
			resolved = append(resolved, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: call.ID,
				Content:    string(payloadBytes),
			})
			recorded = append(recorded, ConversationMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: call.ID,
				Content:    string(payloadBytes),
			})
		}
	}

	return "", nil, fmt.Errorf("resolve tool calls exceeded max rounds: %d", maxRounds)
}

func (s *ChatService) streamAssistantRound(ctx context.Context, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) (openai.ChatCompletionMessage, error) {
	stream, err := s.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    messages,
		Tools:       s.resolveTools(),
		ToolChoice:  "auto",
		Stream:      true,
		MaxTokens:   2048,
		Temperature: 0.7,
		TopP:        1.0,
	})
	if err != nil {
		return openai.ChatCompletionMessage{}, fmt.Errorf("create chat stream failed: %w", err)
	}
	defer stream.Close()

	var contentBuilder strings.Builder
	toolCallCollector := newStreamToolCallCollector()

	for {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				break
			}
			return openai.ChatCompletionMessage{}, fmt.Errorf("recv stream failed: %w", recvErr)
		}

		if len(resp.Choices) == 0 {
			continue
		}

		delta := resp.Choices[0].Delta
		if delta.Content != "" {
			contentBuilder.WriteString(delta.Content)
			if emit != nil {
				if err := emit(map[string]interface{}{
					"type":    "content",
					"content": delta.Content,
				}); err != nil {
					return openai.ChatCompletionMessage{}, err
				}
			}
		}
		if len(delta.ToolCalls) > 0 {
			toolCallCollector.Append(delta.ToolCalls)
		}
	}

	return openai.ChatCompletionMessage{
		Role:      openai.ChatMessageRoleAssistant,
		Content:   contentBuilder.String(),
		ToolCalls: toolCallCollector.ToolCalls(),
	}, nil
}

func (s *ChatService) handleToolCall(ctx context.Context, userID string, call openai.ToolCall) map[string]interface{} {
	toolPayload := map[string]interface{}{
		"error": fmt.Sprintf("unsupported tool: %s", call.Function.Name),
	}

	handler := s.resolveToolHandler(call.Function.Name)
	if handler == nil {
		return toolPayload
	}

	result, err := handler.Handle(ctx, userID, call.Function.Arguments)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	return result.Payload
}

type streamToolCallCollector struct {
	calls   map[int]*openai.ToolCall
	order   []int
	lastIdx int
	hasLast bool
}

func newStreamToolCallCollector() *streamToolCallCollector {
	return &streamToolCallCollector{
		calls: map[int]*openai.ToolCall{},
		order: make([]int, 0),
	}
}

func (c *streamToolCallCollector) Append(deltas []openai.ChatCompletionStreamToolCallDelta) {
	for _, delta := range deltas {
		idx := c.resolveIndex(delta.Index)
		call := c.ensure(idx)

		if strings.TrimSpace(delta.ID) != "" {
			call.ID = delta.ID
		}
		if strings.TrimSpace(delta.Type) != "" {
			call.Type = delta.Type
		}
		if delta.Function != nil {
			if strings.TrimSpace(delta.Function.Name) != "" {
				call.Function.Name = delta.Function.Name
			}
			if delta.Function.Arguments != "" {
				call.Function.Arguments += delta.Function.Arguments
			}
		}
	}
}

func (c *streamToolCallCollector) ToolCalls() []openai.ToolCall {
	result := make([]openai.ToolCall, 0, len(c.order))
	for _, idx := range c.order {
		call := *c.calls[idx]
		if strings.TrimSpace(call.Type) == "" {
			call.Type = openai.ToolTypeFunction
		}
		if strings.TrimSpace(call.ID) == "" {
			call.ID = fmt.Sprintf("tool_call_%d", idx)
		}
		result = append(result, call)
	}
	return result
}

func (c *streamToolCallCollector) resolveIndex(index *int) int {
	if index != nil {
		c.lastIdx = *index
		c.hasLast = true
		return *index
	}
	if c.hasLast {
		return c.lastIdx
	}
	if len(c.order) > 0 {
		return c.order[len(c.order)-1]
	}
	return 0
}

func (c *streamToolCallCollector) ensure(index int) *openai.ToolCall {
	if call, ok := c.calls[index]; ok {
		return call
	}
	call := &openai.ToolCall{
		Type: openai.ToolTypeFunction,
	}
	c.calls[index] = call
	c.order = append(c.order, index)
	return call
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
	return PersistConversationWithPrefix(DefaultConversationKeyPrefix, userID, newMessages)
}

func PersistConversationWithPrefix(conversationKeyPrefix string, userID string, newMessages []ConversationMessage) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}
	if len(newMessages) == 0 {
		return nil
	}

	key := conversationKey(conversationKeyPrefix, trimmedUserID)

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
	return ClearConversationWithPrefix(DefaultConversationKeyPrefix, userID)
}

func ClearConversationWithPrefix(conversationKeyPrefix string, userID string) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	key := conversationKey(conversationKeyPrefix, trimmedUserID)
	if err := cache.Delete(key); err != nil {
		return fmt.Errorf("clear conversation from redis failed: %w", err)
	}

	return nil
}

// GetConversationContext 查询上下文、剩余 TTL 和上下文是否存在。
func GetConversationContext(userID string) ([]ConversationMessage, int64, bool, error) {
	return GetConversationContextWithPrefix(DefaultConversationKeyPrefix, userID)
}

func GetConversationContextWithPrefix(conversationKeyPrefix string, userID string) ([]ConversationMessage, int64, bool, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	key := conversationKey(conversationKeyPrefix, trimmedUserID)

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
func conversationKey(prefix string, userID string) string {
	trimmedPrefix := strings.TrimSpace(prefix)
	if trimmedPrefix == "" {
		trimmedPrefix = DefaultConversationKeyPrefix
	}
	return trimmedPrefix + strings.TrimSpace(userID)
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

func (s *ChatService) resolveTools() []openai.Tool {
	if s != nil && s.toolProvider != nil {
		return s.toolProvider()
	}
	return chattool.ChatTools()
}

func (s *ChatService) resolveToolHandler(name string) chattool.ChatToolHandler {
	if s != nil && s.toolHandlerResolver != nil {
		return s.toolHandlerResolver(name)
	}
	return chattool.GetChatToolHandler(name)
}
