package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	authcore "antifraud/login_system/auth"
	"antifraud/login_system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthUserReader 定义鉴权中间件需要的最小用户读取能力。
type AuthUserReader interface {
	GetUserByID(userID uint) (models.User, error)
}

type gormAuthUserReader struct {
	db *gorm.DB
}

// NewGormAuthUserReader 使用 gorm DB 构建鉴权用户读取实现。
func NewGormAuthUserReader(db *gorm.DB) AuthUserReader {
	return &gormAuthUserReader{db: db}
}

func (r *gormAuthUserReader) GetUserByID(userID uint) (models.User, error) {
	if r == nil || r.db == nil {
		return models.User{}, fmt.Errorf("auth user reader db is nil")
	}
	var user models.User
	if err := r.db.Where("id = ?", userID).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

// AuthMiddleware 校验 Authorization Bearer JWT，并将用户信息写入上下文。
func AuthMiddleware(userReader AuthUserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userReader == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "认证服务不可用"})
			c.Abort()
			return
		}

		tokenString, tokenErr := extractBearerToken(c)
		if tokenErr != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权 Token"})
			c.Abort()
			return
		}

		claims, err := authcore.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的 Token"})
			c.Abort()
			return
		}

		user, err := userReader.GetUserByID(claims.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
			c.Abort()
			return
		}

		if user.Email != claims.Email || user.Username != claims.Username {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户信息不匹配，Token可能已失效"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

// extractBearerToken 解析鉴权 Token：
// 1) 常规 HTTP 请求：仅支持 Authorization: Bearer <token>。
// 2) WebSocket 升级请求：在缺失 Authorization 时，允许使用 query 参数 token。
func extractBearerToken(c *gin.Context) (string, error) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" || strings.TrimSpace(parts[1]) == "" {
			return "", fmt.Errorf("invalid authorization header")
		}
		return strings.TrimSpace(parts[1]), nil
	}

	upgrade := strings.ToLower(strings.TrimSpace(c.GetHeader("Upgrade")))
	if upgrade == "websocket" {
		queryToken := strings.TrimSpace(c.Query("token"))
		if queryToken != "" {
			return queryToken, nil
		}
	}

	return "", fmt.Errorf("token is empty")
}

// AdminMiddleware 确保用户拥有管理员权限
func AdminMiddleware(userReader AuthUserReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userReader == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "认证服务不可用"})
			c.Abort()
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			c.Abort()
			return
		}

		numericUserID, err := normalizeContextUserID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户标识无效"})
			c.Abort()
			return
		}

		user, err := userReader.GetUserByID(numericUserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			c.Abort()
			return
		}

		if user.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func normalizeContextUserID(raw interface{}) (uint, error) {
	switch value := raw.(type) {
	case uint:
		return value, nil
	case int:
		if value < 0 {
			return 0, fmt.Errorf("negative user id")
		}
		return uint(value), nil
	case int64:
		if value < 0 {
			return 0, fmt.Errorf("negative user id")
		}
		return uint(value), nil
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return 0, fmt.Errorf("empty user id")
		}
		parsed, err := strconv.ParseUint(trimmed, 10, 64)
		if err != nil {
			return 0, err
		}
		return uint(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported user id type: %T", raw)
	}
}
