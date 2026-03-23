package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	chatservice "antifraud/internal/modules/chat/adapters/outbound/service"
	"antifraud/internal/modules/chat/application"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message string   `json:"message"`
	Images  []string `json:"images,omitempty"`
}

// ChatContextResponse 返回当前用户的会话上下文快照。
type ChatContextResponse struct {
	UserID     string                            `json:"user_id"`
	HasContext bool                              `json:"has_context"`
	TTLSeconds int64                             `json:"ttl_seconds"`
	Messages   []chatservice.ConversationMessage `json:"messages"`
}

// Handler 是聊天 HTTP 适配器。
type Handler struct {
	useCase *application.UseCase
}

// NewHandler 创建聊天 HTTP 处理器。
func NewHandler(useCase *application.UseCase) *Handler {
	if useCase == nil {
		useCase = application.NewDefaultUseCase("")
	}
	return &Handler{useCase: useCase}
}

var defaultHandler = NewHandler(nil)

// RegisterRoutes 注册聊天相关路由。
func RegisterRoutes(api gin.IRoutes, handler *Handler) {
	if api == nil {
		return
	}
	if handler == nil {
		handler = defaultHandler
	}
	api.POST("/chat", handler.ChatHandle)
	api.GET("/chat/context", handler.GetChatContextHandle)
	api.POST("/chat/refresh", handler.RefreshChatContextHandle)
}

// ChatHandle 处理聊天请求，内部流程为：
// 1) 解析请求；
// 2) 组装上下文；
// 3) 调用模型流式输出；
// 4) 持久化本轮会话到 Redis。
func (h *Handler) ChatHandle(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" && len(req.Images) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message 和 images 不能同时为空"})
		return
	}

	userID := getCurrentUserID(c)
	streamCtx := c.Request.Context()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	emitSSE := func(eventType string, payload map[string]interface{}) {
		c.SSEvent(eventType, chatservice.EncodeEvent(payload))
		c.Writer.Flush()
	}

	assistantReply, err := h.useCase.HandleChat(streamCtx, userID, message, req.Images, func(event map[string]interface{}) error {
		emitSSE("event", event)
		return nil
	})
	if err != nil {
		emitSSE("error", map[string]interface{}{"error": err.Error()})
		return
	}

	_ = assistantReply
}

// RefreshChatContextHandle 清空当前用户在 Redis 中的会话上下文。
func (h *Handler) RefreshChatContextHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	if err := h.useCase.RefreshConversation(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刷新对话失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "对话上下文已刷新",
	})
}

// GetChatContextHandle 读取当前用户会话上下文和剩余 TTL。
func (h *Handler) GetChatContextHandle(c *gin.Context) {
	userID := getCurrentUserID(c)
	messages, ttlSeconds, hasContext, err := h.useCase.GetConversationContext(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询对话上下文失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, ChatContextResponse{
		UserID:     userID,
		HasContext: hasContext,
		TTLSeconds: ttlSeconds,
		Messages:   messages,
	})
}

// ChatHandle 保持旧入口兼容。
func ChatHandle(c *gin.Context) {
	defaultHandler.ChatHandle(c)
}

// RefreshChatContextHandle 保持旧入口兼容。
func RefreshChatContextHandle(c *gin.Context) {
	defaultHandler.RefreshChatContextHandle(c)
}

// GetChatContextHandle 保持旧入口兼容。
func GetChatContextHandle(c *gin.Context) {
	defaultHandler.GetChatContextHandle(c)
}

// getCurrentUserID 从鉴权上下文提取用户 ID，未命中时回退 demo-user。
func getCurrentUserID(c *gin.Context) string {
	userIDValue, exists := c.Get("userID")
	if !exists {
		return "demo-user"
	}
	if value, ok := userIDValue.(uint); ok {
		return fmt.Sprintf("%d", value)
	}
	return strings.TrimSpace(fmt.Sprintf("%v", userIDValue))
}
