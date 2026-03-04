package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	chatcfg "antifraud/chat_system/config"
	chattool "antifraud/chat_system/tool"

	"github.com/redis/go-redis/v9"

	openai "antifraud/llm"
)

const conversationTTL = 5 * time.Minute
const chatSystemPrompt = `你是“反诈对话助手”，同时保持简洁、友好、专业，默认使用中文回复。

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
func NewChatService(cfg *chatcfg.Config) *ChatService {
	modelID := strings.TrimSpace(cfg.ChatModel)
	if modelID == "" {
		modelID = "qwen/qwen3.5-397b-a17b"
	}

	return &ChatService{
		client: openai.NewClientWithConfig(openai.Config{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
		}),
		model: modelID,
	}
}

// BuildMessagesForUser 组装最终请求消息：
// 系统提示词 + Redis 历史上下文 + 当前用户输入。
func BuildMessagesForUser(cfg *chatcfg.Config, userID string, currentUserInput string) ([]openai.ChatCompletionMessage, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: chatSystemPrompt,
		},
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx := context.Background()
	key := conversationKey(trimmedUserID)
	raw, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, fmt.Errorf("load conversation from redis failed: %w", err)
		}
		raw = ""
	}

	if raw != "" {
		_ = rdb.Expire(ctx, key, conversationTTL).Err()
	}

	history := make([]ConversationMessage, 0)
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &history); err != nil {
			return nil, fmt.Errorf("decode conversation failed: %w", err)
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

// StreamReply 先完成工具调用，再进入纯文本流式回复。
func (s *ChatService) StreamReply(ctx context.Context, userID string, userInput string, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) (string, []ConversationMessage, error) {
	finalMessages, toolMessages, err := s.resolveToolCalls(ctx, userID, messages, emit)
	if err != nil {
		return "", nil, err
	}

	stream, err := s.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       s.model,
		Messages:    finalMessages,
		Stream:      true,
		MaxTokens:   2048,
		Temperature: 0.7,
		TopP:        1.0,
	})
	if err != nil {
		return "", nil, fmt.Errorf("create chat stream failed: %w", err)
	}
	defer stream.Close()

	var responseBuilder strings.Builder

	for {
		resp, recvErr := stream.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				break
			}
			return "", nil, fmt.Errorf("recv stream failed: %w", recvErr)
		}

		if len(resp.Choices) == 0 {
			continue
		}

		delta := resp.Choices[0].Delta
		if strings.TrimSpace(delta.Content) == "" && delta.Content == "" {
			continue
		}

		responseBuilder.WriteString(delta.Content)
		if emit != nil {
			if err := emit(map[string]interface{}{
				"type":    "content",
				"content": delta.Content,
			}); err != nil {
				return "", nil, err
			}
		}
	}

	if emit != nil {
		if err := emit(map[string]interface{}{"type": "done", "reason": "stop"}); err != nil {
			return "", nil, err
		}
	}

	finalReply := strings.TrimSpace(responseBuilder.String())

	turnMessages := make([]ConversationMessage, 0, len(toolMessages)+2)
	turnMessages = append(turnMessages, ConversationMessage{Role: openai.ChatMessageRoleUser, Content: strings.TrimSpace(userInput)})
	turnMessages = append(turnMessages, toolMessages...)
	turnMessages = append(turnMessages, ConversationMessage{Role: openai.ChatMessageRoleAssistant, Content: finalReply})

	return finalReply, turnMessages, nil
}

// resolveToolCalls 通过非流式轮询，执行模型触发的工具调用并回填 tool 消息。
func (s *ChatService) resolveToolCalls(ctx context.Context, userID string, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) ([]openai.ChatCompletionMessage, []ConversationMessage, error) {
	resolved := append([]openai.ChatCompletionMessage{}, messages...)
	recorded := make([]ConversationMessage, 0)

	for {
		resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       s.model,
			Messages:    resolved,
			Tools:       chattool.ChatTools(),
			ToolChoice:  "auto",
			Stream:      false,
			MaxTokens:   1024,
			Temperature: 0.3,
			TopP:        1.0,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("resolve tool calls failed: %w", err)
		}

		if len(resp.Choices) == 0 {
			return resolved, recorded, nil
		}

		msg := resp.Choices[0].Message
		resolved = append(resolved, openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		if len(msg.ToolCalls) == 0 {
			return resolved, recorded, nil
		}

		recorded = append(recorded, ConversationMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   strings.TrimSpace(msg.Content),
			ToolCalls: openAIToolCallsToConversation(msg.ToolCalls),
		})

		toolResponseAdded := false
		for _, call := range msg.ToolCalls {
			handler := chattool.GetChatToolHandler(call.Function.Name)
			toolPayload := map[string]interface{}{
				"error": fmt.Sprintf("unsupported tool: %s", call.Function.Name),
			}

			if handler != nil {
				result, handleErr := handler.Handle(ctx, userID, call.Function.Arguments)
				if handleErr != nil {
					toolPayload = map[string]interface{}{"error": handleErr.Error()}
				} else {
					toolPayload = result.Payload
				}
			}

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
			toolResponseAdded = true
		}

		if !toolResponseAdded {
			return resolved, recorded, nil
		}
	}
}

// PersistConversation 将本轮新增消息追加写入 Redis，并重置 TTL。
func PersistConversation(cfg *chatcfg.Config, userID string, newMessages []ConversationMessage) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}
	if len(newMessages) == 0 {
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx := context.Background()
	key := conversationKey(trimmedUserID)

	history := make([]ConversationMessage, 0)
	raw, err := rdb.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("load conversation before persist failed: %w", err)
	}
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &history); err != nil {
			return fmt.Errorf("decode conversation before persist failed: %w", err)
		}
	}

	history = append(history, sanitizeConversationMessages(newMessages)...)

	payload, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("encode conversation failed: %w", err)
	}

	if err := rdb.Set(ctx, key, payload, conversationTTL).Err(); err != nil {
		return fmt.Errorf("save conversation to redis failed: %w", err)
	}
	return nil
}

// ClearConversation 清空指定用户会话上下文。
func ClearConversation(cfg *chatcfg.Config, userID string) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx := context.Background()
	key := conversationKey(trimmedUserID)
	if err := rdb.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("clear conversation from redis failed: %w", err)
	}

	return nil
}

// GetConversationContext 查询上下文、剩余 TTL 和上下文是否存在。
func GetConversationContext(cfg *chatcfg.Config, userID string) ([]ConversationMessage, int64, bool, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPwd,
		DB:       cfg.RedisDB,
	})
	defer rdb.Close()

	ctx := context.Background()
	key := conversationKey(trimmedUserID)

	raw, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return []ConversationMessage{}, 0, false, nil
		}
		return nil, 0, false, fmt.Errorf("load conversation context failed: %w", err)
	}

	history := make([]ConversationMessage, 0)
	if strings.TrimSpace(raw) != "" {
		if err := json.Unmarshal([]byte(raw), &history); err != nil {
			return nil, 0, true, fmt.Errorf("decode conversation context failed: %w", err)
		}
	}

	ttl, err := rdb.TTL(ctx, key).Result()
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
