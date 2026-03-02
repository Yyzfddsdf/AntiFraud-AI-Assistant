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

// main 是整个服务的启动入口：
// 1) 初始化数据库；
// 2) 注册全局中间件；
// 3) 挂载认证、对话、多模态相关路由；
// 4) 启动 HTTP 服务。
func main() {
	database.ConnectDB()

	r := gin.Default()

	// 不信任任何反向代理头，避免在本地开发时出现代理来源误判。
	r.SetTrustedProxies(nil)

	// 全局 CORS：前端测试页面与本地调试接口都走同一套跨域策略。
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

	// 全局限流，防止暴力请求压垮服务。
	r.Use(middleware.RateLimitMiddleware())
	r.StaticFile("/test-login", "login_system/web/test_login.html")
	r.StaticFile("/", "login_system/web/index.html")
	r.Static("/assets", "login_system/web/assets")

	// 登录与注册相关接口（无需 JWT）。
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.GET("/captcha", controllers.GetCaptchaHandle)
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	// 业务接口（需 JWT）。
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/user", controllers.GetCurrentUserHandle)
		api.DELETE("/user", controllers.DeleteCurrentUserHandle)
		api.GET("/users", middleware.AdminMiddleware(), controllers.GetAllUsersHandle)
		api.POST("/upgrade", controllers.UpgradeUserHandle)
		api.POST("/chat", chatapi.ChatHandle)
		api.GET("/chat/context", chatapi.GetChatContextHandle)
		api.POST("/chat/refresh", chatapi.RefreshChatContextHandle)
		api.PUT("/scam/multimodal/user/age", httpapi.UpdateUserAgeHandle)
		api.POST("/scam/multimodal/analyze", httpapi.AnalyzeMultimodalScamHandle)
		api.GET("/scam/multimodal/tasks", httpapi.GetMultimodalTaskStateHandle)
		api.GET("/scam/multimodal/history", httpapi.GetMultimodalHistoryHandle)
		api.DELETE("/scam/multimodal/history/:recordId", httpapi.DeleteMultimodalHistoryHandle)
		api.GET("/scam/multimodal/tasks/:taskId", httpapi.GetMultimodalTaskDetailHandle)

		adminCaseLibrary := api.Group("/scam/case-library")
		adminCaseLibrary.Use(middleware.AdminMiddleware())
		{
			adminCaseLibrary.POST("/cases", httpapi.CreateHistoricalCaseHandle)
			adminCaseLibrary.GET("/cases", httpapi.GetHistoricalCasePreviewHandle)
			adminCaseLibrary.GET("/cases/:caseId", httpapi.GetHistoricalCaseDetailHandle)
			adminCaseLibrary.DELETE("/cases/:caseId", httpapi.DeleteHistoricalCaseHandle)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	r.Run(":" + port)
}
