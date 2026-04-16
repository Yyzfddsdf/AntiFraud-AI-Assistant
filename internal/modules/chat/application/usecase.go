package application

import (
	"context"
	"fmt"

	chatservice "antifraud/internal/modules/chat/adapters/outbound/service"
	appcfg "antifraud/internal/platform/config"
	openai "antifraud/internal/platform/llm"
)

const defaultConfigPath = "internal/platform/config/config.json"

// ConfigProvider 定义聊天用例所需的最小配置读取能力。
type ConfigProvider interface {
	LoadChatConfig(ctx context.Context) (appcfg.ChatConfig, error)
}

// MessageBuilder 定义聊天消息组装端口。
type MessageBuilder interface {
	Build(systemPrompt string, userID string, userInput string, userImageURLs []string) ([]openai.ChatCompletionMessage, error)
}

// ConversationResponder 定义模型流式响应端口。
type ConversationResponder interface {
	StreamReply(ctx context.Context, userID string, userInput string, userImageURLs []string, messages []openai.ChatCompletionMessage, emit func(event map[string]interface{}) error) (string, []chatservice.ConversationMessage, error)
}

// ConversationResponderFactory 负责按配置创建响应器。
type ConversationResponderFactory interface {
	New(chatCfg appcfg.ChatConfig) ConversationResponder
}

// ConversationStore 定义会话持久化端口。
type ConversationStore interface {
	Persist(userID string, newMessages []chatservice.ConversationMessage) error
	Clear(userID string) error
	GetContext(userID string) ([]chatservice.ConversationMessage, int64, bool, error)
}

// UseCase 负责编排聊天请求。
type UseCase struct {
	configProvider    ConfigProvider
	messageBuilder    MessageBuilder
	responderFactory  ConversationResponderFactory
	conversationStore ConversationStore
}

// NewUseCase 创建聊天应用服务。
func NewUseCase(configProvider ConfigProvider, messageBuilder MessageBuilder, responderFactory ConversationResponderFactory, conversationStore ConversationStore) *UseCase {
	return &UseCase{
		configProvider:    configProvider,
		messageBuilder:    messageBuilder,
		responderFactory:  responderFactory,
		conversationStore: conversationStore,
	}
}

// NewDefaultUseCase 创建默认聊天应用服务。
func NewDefaultUseCase(configPath string) *UseCase {
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return NewUseCase(
		fileConfigProvider{path: configPath},
		defaultMessageBuilder{},
		defaultResponderFactory{},
		defaultConversationStore{},
	)
}

// NewAdminUseCase 创建管理员聊天应用服务。
func NewAdminUseCase(configPath string) *UseCase {
	if configPath == "" {
		configPath = defaultConfigPath
	}
	return NewUseCase(
		adminConfigProvider{path: configPath},
		adminMessageBuilder{},
		adminResponderFactory{},
		adminConversationStore{},
	)
}

// HandleChat 执行一轮聊天请求。
func (u *UseCase) HandleChat(ctx context.Context, userID string, message string, images []string, emit func(event map[string]interface{}) error) (string, error) {
	if u == nil {
		return "", fmt.Errorf("chat use case is unavailable")
	}
	if u.configProvider == nil || u.messageBuilder == nil || u.responderFactory == nil || u.conversationStore == nil {
		return "", fmt.Errorf("chat use case dependencies are incomplete")
	}

	cfg, err := u.configProvider.LoadChatConfig(ctx)
	if err != nil {
		return "", err
	}

	messages, err := u.messageBuilder.Build(cfg.Prompt, userID, message, images)
	if err != nil {
		return "", err
	}

	responder := u.responderFactory.New(cfg)
	if responder == nil {
		return "", fmt.Errorf("chat responder is unavailable")
	}

	reply, turnMessages, err := responder.StreamReply(ctx, userID, message, images, messages, emit)
	if err != nil {
		return "", err
	}
	if err := u.conversationStore.Persist(userID, turnMessages); err != nil {
		return "", err
	}
	return reply, nil
}

// RefreshConversation 清空指定用户对话上下文。
func (u *UseCase) RefreshConversation(userID string) error {
	if u == nil || u.conversationStore == nil {
		return fmt.Errorf("chat use case is unavailable")
	}
	return u.conversationStore.Clear(userID)
}

// GetConversationContext 查询指定用户对话上下文。
func (u *UseCase) GetConversationContext(userID string) ([]chatservice.ConversationMessage, int64, bool, error) {
	if u == nil || u.conversationStore == nil {
		return nil, 0, false, fmt.Errorf("chat use case is unavailable")
	}
	return u.conversationStore.GetContext(userID)
}

type fileConfigProvider struct {
	path string
}

func (p fileConfigProvider) LoadChatConfig(ctx context.Context) (appcfg.ChatConfig, error) {
	_ = ctx
	cfg, err := appcfg.LoadConfig(p.path)
	if err != nil {
		return appcfg.ChatConfig{}, fmt.Errorf("加载聊天配置失败: %w", err)
	}
	return cfg.Chat, nil
}

type defaultMessageBuilder struct{}

func (defaultMessageBuilder) Build(systemPrompt string, userID string, userInput string, userImageURLs []string) ([]openai.ChatCompletionMessage, error) {
	return chatservice.BuildMessagesForUser(systemPrompt, userID, userInput, userImageURLs)
}

type defaultResponderFactory struct{}

func (defaultResponderFactory) New(chatCfg appcfg.ChatConfig) ConversationResponder {
	return chatservice.NewChatService(&chatCfg)
}

type defaultConversationStore struct{}

func (defaultConversationStore) Persist(userID string, newMessages []chatservice.ConversationMessage) error {
	return chatservice.PersistConversation(userID, newMessages)
}

func (defaultConversationStore) Clear(userID string) error {
	return chatservice.ClearConversation(userID)
}

func (defaultConversationStore) GetContext(userID string) ([]chatservice.ConversationMessage, int64, bool, error) {
	return chatservice.GetConversationContext(userID)
}

type adminConfigProvider struct {
	path string
}

func (p adminConfigProvider) LoadChatConfig(ctx context.Context) (appcfg.ChatConfig, error) {
	_ = ctx
	cfg, err := appcfg.LoadConfig(p.path)
	if err != nil {
		return appcfg.ChatConfig{}, err
	}
	return cfg.AdminChat, nil
}

type adminMessageBuilder struct{}

func (adminMessageBuilder) Build(systemPrompt string, userID string, userInput string, userImageURLs []string) ([]openai.ChatCompletionMessage, error) {
	return chatservice.BuildMessagesForUserWithPrefix(chatservice.AdminConversationKeyPrefix, systemPrompt, userID, userInput, userImageURLs)
}

type adminResponderFactory struct{}

func (adminResponderFactory) New(chatCfg appcfg.ChatConfig) ConversationResponder {
	return chatservice.NewAdminChatService(&chatCfg)
}

type adminConversationStore struct{}

func (adminConversationStore) Persist(userID string, newMessages []chatservice.ConversationMessage) error {
	return chatservice.PersistConversationWithPrefix(chatservice.AdminConversationKeyPrefix, userID, newMessages)
}

func (adminConversationStore) Clear(userID string) error {
	return chatservice.ClearConversationWithPrefix(chatservice.AdminConversationKeyPrefix, userID)
}

func (adminConversationStore) GetContext(userID string) ([]chatservice.ConversationMessage, int64, bool, error) {
	return chatservice.GetConversationContextWithPrefix(chatservice.AdminConversationKeyPrefix, userID)
}
