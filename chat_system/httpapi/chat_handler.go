package httpapi

import (
	"fmt"
	"net/http"
	"strings"

	chatcfg "antifraud/chat_system/config"
	chatservice "antifraud/chat_system/service"

	"github.com/gin-gonic/gin"
)

type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// ChatContextResponse 返回当前用户的会话上下文快照。
type ChatContextResponse struct {
	UserID     string                            `json:"user_id"`
	HasContext bool                              `json:"has_context"`
	TTLSeconds int64                             `json:"ttl_seconds"`
	Messages   []chatservice.ConversationMessage `json:"messages"`
}

// ChatHandle 处理聊天请求，内部流程为：
// 1) 解析请求；
// 2) 组装上下文；
// 3) 调用模型流式输出；
// 4) 持久化本轮会话到 Redis。
func ChatHandle(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message 不能为空"})
		return
	}

	cfg, err := chatcfg.LoadConfig("chat_system/config/config.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载聊天配置失败: " + err.Error()})
		return
	}

	userID := getCurrentUserID(c)
	messages, err := chatservice.BuildMessagesForUser(cfg, userID, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载Redis上下文失败: " + err.Error()})
		return
	}

	service := chatservice.NewChatService(cfg)
	streamCtx := c.Request.Context()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	assistantReply, turnMessages, err := service.StreamReply(streamCtx, userID, message, messages, func(event map[string]interface{}) error {
		c.SSEvent("event", chatservice.EncodeEvent(event))
		return nil
	})
	if err != nil {
		c.SSEvent("error", chatservice.EncodeEvent(map[string]interface{}{"error": err.Error()}))
		return
	}

	if err := chatservice.PersistConversation(cfg, userID, turnMessages); err != nil {
		c.SSEvent("error", chatservice.EncodeEvent(map[string]interface{}{"error": err.Error()}))
		return
	}

	_ = assistantReply
}

// RefreshChatContextHandle 清空当前用户在 Redis 中的会话上下文。
func RefreshChatContextHandle(c *gin.Context) {
	cfg, err := chatcfg.LoadConfig("chat_system/config/config.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载聊天配置失败: " + err.Error()})
		return
	}

	userID := getCurrentUserID(c)
	if err := chatservice.ClearConversation(cfg, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "刷新对话失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"message": "对话上下文已刷新",
	})
}

// GetChatContextHandle 读取当前用户会话上下文和剩余 TTL。
func GetChatContextHandle(c *gin.Context) {
	cfg, err := chatcfg.LoadConfig("chat_system/config/config.json")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加载聊天配置失败: " + err.Error()})
		return
	}

	userID := getCurrentUserID(c)
	messages, ttlSeconds, hasContext, err := chatservice.GetConversationContext(cfg, userID)
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
