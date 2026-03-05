package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"antifraud/cache"
	"antifraud/login_system/settings"

	"github.com/gin-gonic/gin"
)

var (
	rateWindow = settings.RateLimitWindow
	rateLimit  = settings.RateLimitMaxRequests
)

const rateLimitIPPrefix = "cache:rate_limit:ip:"

// RateLimitMiddleware 基于 IP + 时间窗口做轻量限流。
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := rateLimitCacheKey(ip)

		count, err := cache.IncrWithinWindow(key, rateWindow)
		if err != nil {
			// 限流缓存异常时默认放行，避免缓存故障扩大为全站不可用。
			log.Printf("[rate_limit] incr failed: ip=%s err=%v", ip, err)
			c.Next()
			return
		}

		if count > int64(rateLimit) {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func rateLimitCacheKey(ip string) string {
	trimmedIP := strings.TrimSpace(ip)
	if trimmedIP == "" {
		trimmedIP = "unknown"
	}
	return fmt.Sprintf("%s%s:%dms", rateLimitIPPrefix, trimmedIP, rateWindow/time.Millisecond)
}
