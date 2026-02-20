package middleware

import (
	"net/http"
	"strings"

	"image_recognition/login_system/controllers"
	"image_recognition/login_system/database"
	"image_recognition/login_system/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware 校验 Authorization Bearer JWT，并将用户信息写入上下文。
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供授权 Token"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 格式错误"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims := &controllers.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return controllers.GetJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效或过期的 Token"})
			c.Abort()
			return
		}

		var user models.User
		if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
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
		c.Set("user", models.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email, Age: user.Age})
		c.Next()
	}
}
