package main

import (
	"net/http"
	"os"

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
