package main

import (
	"net/http"
	"os"

	chatapi "image_recognition/chat_system/httpapi"
	"image_recognition/login_system/controllers"
	"image_recognition/login_system/database"
	"image_recognition/login_system/middleware"
	"image_recognition/multi_agent/httpapi"

	"github.com/gin-gonic/gin"
)

// main 启动登录系统 HTTP 服务，完成数据库初始化、路由挂载和中间件注册。
func main() {
	database.ConnectDB()

	r := gin.Default()

	// 禁用代理信任警告：在开发环境中，如果你不使用反向代理（如 Nginx），设置为 nil 即可。
	// 这表示不信任任何代理服务器发送的头部信息（如 X-Forwarded-For）。
	r.SetTrustedProxies(nil)

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	r.Use(middleware.RateLimitMiddleware())
	r.StaticFile("/test-login", "login_system/web/test_login.html")
	r.StaticFile("/", "login_system/web/index.html")

	authRoutes := r.Group("/api/auth")
	{
		authRoutes.GET("/captcha", controllers.GetCaptchaHandle)
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/user", controllers.GetCurrentUserHandle)
		api.DELETE("/user", controllers.DeleteCurrentUserHandle)
		api.POST("/chat", chatapi.ChatHandle)
		api.GET("/chat/context", chatapi.GetChatContextHandle)
		api.POST("/chat/refresh", chatapi.RefreshChatContextHandle)
		api.PUT("/scam/multimodal/user/age", httpapi.UpdateUserAgeHandle)
		api.POST("/scam/multimodal/analyze", httpapi.AnalyzeMultimodalScamHandle)
		api.GET("/scam/multimodal/tasks", httpapi.GetMultimodalTaskStateHandle)
		api.GET("/scam/multimodal/history", httpapi.GetMultimodalHistoryHandle)
		api.GET("/scam/multimodal/tasks/:taskId", httpapi.GetMultimodalTaskDetailHandle)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	r.Run(":" + port)
}
