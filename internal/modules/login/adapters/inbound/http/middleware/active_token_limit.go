package middleware

import (
	"log"
	"net/http"
	"strings"

	"antifraud/internal/modules/login/adapters/outbound/session"

	"github.com/gin-gonic/gin"
)

// ActiveTokenLimitMiddleware 维护单用户活跃 token 队列。
// 设计策略：
// 1) 基于最近活跃时间维护最多 2 个 token；
// 2) TTL 固定 5 分钟，活跃访问会刷新；
// 3) Redis 异常时采用 fail-open，避免把鉴权链路整体拖垮。
func ActiveTokenLimitMiddleware(activeTokenManager session.ActiveTokenManager) gin.HandlerFunc {
	if activeTokenManager == nil {
		activeTokenManager = session.NewDefaultRedisActiveTokenManager()
	}

	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		userID, err := normalizeContextUserID(userIDValue)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户标识无效"})
			c.Abort()
			return
		}

		tokenValue, exists := c.Get("authToken")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌缺失"})
			c.Abort()
			return
		}

		tokenString, ok := tokenValue.(string)
		if !ok || strings.TrimSpace(tokenString) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "认证令牌无效"})
			c.Abort()
			return
		}

		allowed, err := activeTokenManager.AllowRequestToken(c.Request.Context(), userID, tokenString)
		if err != nil {
			log.Printf("active token limiter degraded, allow request: user_id=%d err=%v", userID, err)
			c.Next()
			return
		}
		if !allowed {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "当前登录已在其他设备被挤下线，请重新登录"})
			c.Abort()
			return
		}

		c.Next()
	}
}
