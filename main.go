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

 // main 是服务启动入口：初始化、路由注册与启动 HTTP 服务。
 // 不信任反向代理头，避免来源判断偏差。
 // 全局 CORS 配置。
func main() {
	database.ConnectDB()

	r := gin.Default()

	 // 全局限流。
	r.SetTrustedProxies(nil)

	 // 静态资源与主页。
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

	 // 认证相关接口（无需 JWT）。
	r.Use(middleware.RateLimitMiddleware())
	r.StaticFile("/test-login", "login_system/web/test_login.html")
	r.StaticFile("/", "login_system/web/index.html")
	r.Static("/assets", "login_system/web/assets")

	 // 业务接口（需要 JWT）。
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.GET("/captcha", controllers.GetCaptchaHandle)
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	 // 管理员升级接口（需要管理员权限）。
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
			adminCaseLibrary.GET("/options/scam-types", httpapi.GetHistoricalCaseScamTypeOptionsHandle)
			adminCaseLibrary.GET("/options/target-groups", httpapi.GetHistoricalCaseTargetGroupOptionsHandle)
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

