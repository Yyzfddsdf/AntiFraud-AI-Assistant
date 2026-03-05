package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"antifraud/cache"
	chattool "antifraud/chat_system/tool"
	appcfg "antifraud/config"

	openai "antifraud/llm"
)

const conversationTTL = 5 * time.Minute
const defaultChatSystemPrompt = `你是“反诈对话助手”，同时保持简洁、友好、专业，默认使用中文回复。

你的核心目标是主动引导用户防骗，而不是只被动答题。每轮对话按以下原则执行：
1) 先识别风险信号：冒充公检法/客服、诱导转账、索要验证码、要求屏幕共享/远程控制、投资拉群、刷单返利、交友裸聊敲诈、虚假链接等。
2) 信息不足时，优先提出 2-4 个关键澄清问题（联系渠道、对方身份说法、是否已转账、金额/时间、是否泄露验证码或银行卡信息）。
3) 给出明确风险分级（低/中/高/紧急）及理由。
4) 给出可执行步骤，优先“立刻能做”的止损动作：
   - 未转账：立即停止操作，不点击链接、不提供验证码、不下载远控软件。
   - 已转账：立即联系银行/支付平台申请止付或冻结；第一时间报警（110）并联系国家反诈专线（96110）；保留聊天、转账、账号、链接等证据。
   - 已泄露敏感信息：立刻改密码、开启二次验证、冻结相关账户并关注异常登录与扣款。
5) 高风险或紧急场景下，直接明确劝阻并给出“先断联、先止损、再核验”的顺序。
6) 可利用现有工具查询用户相关信息（如用户画像、历史案例）并结合结果回答；若工具失败，明确不确定性并提供通用安全建议。

输出要求：
- 结论先行，步骤清晰，尽量使用短句和编号。
- 不制造恐慌，不做法律定性，不编造机构联系方式。`

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
		modelID = "qwen/qwen3.5-397b-a17b"
	}

	return &ChatService{
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  apiKey,
			BaseURL: baseURL,
		}),
		model: modelID,
	}
}

// BuildMessagesForUser 组装最终请求消息：
// 系统提示词 + Redis 历史上下文 + 当前用户输入。
func BuildMessagesForUser(systemPrompt string, userID string, currentUserInput string) ([]openai.ChatCompletionMessage, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: resolveChatSystemPrompt(systemPrompt),
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

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: strings.TrimSpace(currentUserInput),
	})

	return messages, nil
}

// StreamReply 使用全流式回合处理：首轮即 stream=true，边接收边向前端推送内容；
// 若出现 tool_calls，则在参数拼接完成后执行工具并继续下一轮流式请求。
func (s *ChatService) StreamReply(ctx context.Context, userID string, userInput string, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) (string, []ConversationMessage, error) {
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
				Role:    openai.ChatMessageRoleUser,
				Content: strings.TrimSpace(userInput),
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
		Tools:       chattool.ChatTools(),
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
		return openai.ChatCompletionMessage{Role: role, Content: item.Content}, true
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

func resolveChatSystemPrompt(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return defaultChatSystemPrompt
	}
	return trimmed
}
