package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	chatcfg "image_recognition/chat_system/config"
	chattool "image_recognition/chat_system/tool"

	"github.com/redis/go-redis/v9"

	"github.com/sashabaranov/go-openai"
)

const conversationTTL = 5 * time.Minute
const chatSystemPrompt = "你是一个简洁、友好的中文聊天助手。必要时可调用工具查询用户信息或用户案件历史后再回答。"

type ConversationToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ConversationMessage struct {
	Role       string                 `json:"role"`
	Content    string                 `json:"content"`
	ToolCallID string                 `json:"tool_call_id,omitempty"`
	ToolCalls  []ConversationToolCall `json:"tool_calls,omitempty"`
}

type ChatService struct {
	client *openai.Client
	model  string
}

func NewChatService(cfg *chatcfg.Config) *ChatService {
	openaiCfg := openai.DefaultConfig(cfg.APIKey)
	openaiCfg.BaseURL = cfg.BaseURL

	modelID := strings.TrimSpace(cfg.ChatModel)
	if modelID == "" {
		modelID = "qwen/qwen3.5-397b-a17b"
	}

	return &ChatService{
		client: openai.NewClientWithConfig(openaiCfg),
		model:  modelID,
	}
}

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

func conversationKey(userID string) string {
	return "chat:context:" + strings.TrimSpace(userID)
}

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

func EncodeEvent(event map[string]interface{}) string {
	data, _ := json.Marshal(event)
	return string(data)
}
