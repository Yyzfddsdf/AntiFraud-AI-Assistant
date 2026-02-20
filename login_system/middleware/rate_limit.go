package middleware

import (
	"net/http"
	"sync"
	"time"

	"image_recognition/login_system/settings"

	"github.com/gin-gonic/gin"
)

type ipBucket struct {
	windowStart time.Time
	count       int
}

var (
	rateMu     sync.Mutex
	rateWindow = settings.RateLimitWindow
	rateLimit  = settings.RateLimitMaxRequests
	ipBuckets  = map[string]*ipBucket{}
)

// RateLimitMiddleware 基于 IP + 时间窗口做轻量限流。
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		rateMu.Lock()
		bucket, exists := ipBuckets[ip]
		if !exists || now.Sub(bucket.windowStart) >= rateWindow {
			ipBuckets[ip] = &ipBucket{windowStart: now, count: 1}
			rateMu.Unlock()
			c.Next()
			return
		}

		if bucket.count >= rateLimit {
			rateMu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁"})
			c.Abort()
			return
		}

		bucket.count++
		rateMu.Unlock()
		c.Next()
	}
}
