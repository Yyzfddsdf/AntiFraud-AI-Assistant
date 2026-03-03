package main

import (
	"log"
	"net/http"
	"os"

	appcfg "antifraud/config"
	chatapi "antifraud/chat_system/httpapi"
	"antifraud/login_system/controllers"
	"antifraud/login_system/database"
	"antifraud/login_system/middleware"
	"antifraud/multi_agent/case_library"
	"antifraud/multi_agent/httpapi"

	"github.com/gin-gonic/gin"
)

// main 是服务启动入口：完成数据库初始化、路由注册并启动 HTTP 服务。
func main() {
	// 启动时预加载主配置，避免首次分析请求触发读盘与解析。
	if _, err := appcfg.LoadConfig("config/config.json"); err != nil {
		log.Fatalf("load app config failed: %v", err)
	}

	database.ConnectDB()
	if err := case_library.InitHistoricalCaseDB(); err != nil {
		log.Fatalf("init historical case db failed: %v", err)
	}

	r := gin.Default()

	// 不信任反向代理头，避免来源判断偏差。
	r.SetTrustedProxies(nil)

	// 全局 CORS 配置。
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

	// 全局限流。
	r.Use(middleware.RateLimitMiddleware())

	// 静态资源与主页。
	r.StaticFile("/test-login", "login_system/web/test_login.html")
	r.StaticFile("/", "login_system/web/index.html")
	r.Static("/assets", "login_system/web/assets")

	// 认证相关接口（无需 JWT）。
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.GET("/captcha", controllers.GetCaptchaHandle)
		authRoutes.POST("/register", controllers.RegisterHandle)
		authRoutes.POST("/login", controllers.LoginHandle)
	}

	// 业务接口（需要 JWT）。
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

		// 管理员案件库接口（需要管理员权限）。
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

