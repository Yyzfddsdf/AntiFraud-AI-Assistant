package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	chatapi "antifraud/internal/modules/chat/adapters/inbound/http"
	chatapp "antifraud/internal/modules/chat/application"
	"antifraud/internal/modules/family"
	"antifraud/internal/modules/login/adapters/inbound/http/controllers"
	"antifraud/internal/modules/login/adapters/inbound/http/middleware"
	"antifraud/internal/modules/login/adapters/outbound/session"
	"antifraud/internal/modules/login/adapters/outbound/smscode"
	multihttp "antifraud/internal/modules/multi_agent/adapters/inbound/http"
	"antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	region_system "antifraud/internal/modules/region"
	"antifraud/internal/modules/scam_simulation"
	"antifraud/internal/modules/user_profile"
	"antifraud/internal/platform/cache"
	appcfg "antifraud/internal/platform/config"
	"antifraud/internal/platform/database"

	"github.com/gin-gonic/gin"
)

const defaultConfigPath = "internal/platform/internal/platform/config/config.json"

// BuildRouter 创建并装配 HTTP 服务。
func BuildRouter() (*gin.Engine, error) {
	if _, err := appcfg.LoadConfig(defaultConfigPath); err != nil {
		return nil, err
	}
	if err := database.InitPersistence(); err != nil {
		return nil, err
	}
	if err := case_library.WarmupHistoricalCaseVectorCache(); err != nil {
		log.Printf("warmup historical case vector cache failed: %v", err)
	}

	authUserReader := middleware.NewGormAuthUserReader(database.DB)
	activeTokenManager := session.NewDefaultRedisActiveTokenManager()
	smsCodeService := smscode.NewDemoService()
	familyService := family_system.NewService(database.DB)
	userProfileService := user_profile_system.DefaultService()
	regionService := region_system.NewService()
	simulationService := scam_simulation.NewService()
	authHandler := controllers.NewDefaultAuthHandler(activeTokenManager, smsCodeService)
	chatHandler := chatapi.NewHandler(chatapp.NewDefaultUseCase(defaultConfigPath))

	state.RegisterHistoryObserver(func(record state.CaseHistoryRecord) {
		_ = cache.SetJSON("cache:case_library:geo_map:v1:version", fmt.Sprintf("%d", time.Now().UnixNano()), 0)
		region_system.TouchRegionCaseStatsCacheVersion()
		if record.RiskLevel != "高" {
			return
		}
		userID, err := strconv.ParseUint(record.UserID, 10, 64)
		if err != nil {
			return
		}
		_ = familyService.HandleRiskEvent(context.Background(), family_system.RiskEvent{
			TargetUserID: uint(userID),
			RecordID:     record.RecordID,
			Title:        record.Title,
			CaseSummary:  record.CaseSummary,
			ScamType:     record.ScamType,
			RiskLevel:    record.RiskLevel,
			CreatedAt:    record.CreatedAt,
		})
	})

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.Use(corsMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	registerAuthRoutes(r, authHandler, smsCodeService)
	registerProtectedRoutes(r, authUserReader, activeTokenManager, authHandler, userProfileService, familyService, regionService, simulationService, chatHandler)

	return r, nil
}

// Run 启动 HTTP 服务。
func Run() error {
	r, err := BuildRouter()
	if err != nil {
		return err
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	return r.Run(":" + port)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func registerAuthRoutes(r *gin.Engine, authHandler *controllers.AuthHandler, smsCodeService smscode.Service) {
	authRoutes := r.Group("/api/auth")
	authRoutes.GET("/captcha", controllers.GetCaptchaHandle)
	authRoutes.POST("/sms-code", controllers.SendSMSCodeHandle(smsCodeService))
	authRoutes.POST("/register", authHandler.RegisterHandle)
	authRoutes.POST("/login", authHandler.LoginHandle)
}

func registerProtectedRoutes(
	r *gin.Engine,
	authUserReader middleware.AuthUserReader,
	activeTokenManager session.ActiveTokenManager,
	authHandler *controllers.AuthHandler,
	userProfileService *user_profile_system.Service,
	familyService *family_system.Service,
	regionService *region_system.Service,
	simulationService *scam_simulation.Service,
	chatHandler *chatapi.Handler,
) {
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authUserReader), middleware.ActiveTokenLimitMiddleware(activeTokenManager))

	api.GET("/user", authHandler.GetCurrentUserHandle)
	api.DELETE("/user", authHandler.DeleteCurrentUserHandle)
	api.GET("/users", middleware.AdminMiddleware(authUserReader), authHandler.GetAllUsersHandle)
	api.POST("/upgrade", authHandler.UpgradeUserHandle)
	user_profile_system.RegisterRoutes(api, userProfileService)
	region_system.RegisterRoutes(api, regionService)
	scam_simulation.RegisterRoutes(api, simulationService)
	chatapi.RegisterRoutes(api, chatHandler)
	api.GET("/alert/ws", multihttp.AlertWebSocketHandle)
	family_system.RegisterRoutes(api, familyService)
	api.POST("/scam/image/quick-analyze", multihttp.AnalyzeImageQuickHandle)
	api.POST("/scam/multimodal/analyze", multihttp.AnalyzeMultimodalScamHandle)
	api.GET("/scam/multimodal/tasks", multihttp.GetMultimodalTaskStateHandle)
	api.GET("/scam/multimodal/history", multihttp.GetMultimodalHistoryHandle)
	api.GET("/scam/multimodal/history/overview", multihttp.GetMultimodalRiskOverviewHandle)
	api.DELETE("/scam/multimodal/history/:recordId", multihttp.DeleteMultimodalHistoryHandle)
	api.GET("/scam/multimodal/tasks/:taskId", multihttp.GetMultimodalTaskDetailHandle)

	adminCaseLibrary := api.Group("/scam/case-library")
	adminCaseLibrary.Use(middleware.AdminMiddleware(authUserReader))
	adminCaseLibrary.POST("/cases", multihttp.CreateHistoricalCaseHandle)
	adminCaseLibrary.GET("/cases", multihttp.GetHistoricalCasePreviewHandle)
	adminCaseLibrary.GET("/cases/overview", multihttp.GetHistoricalCaseStatisticsOverviewHandle)
	adminCaseLibrary.GET("/cases/graph", multihttp.GetHistoricalCaseGraphHandle)
	adminCaseLibrary.GET("/cases/geo-map", multihttp.GetGeoCaseMapHandle)
	adminCaseLibrary.GET("/maps/geojson", multihttp.GetGeoBoundaryGeoJSONHandle)
	adminCaseLibrary.GET("/options/scam-types", multihttp.GetHistoricalCaseScamTypeOptionsHandle)
	adminCaseLibrary.GET("/options/target-groups", multihttp.GetHistoricalCaseTargetGroupOptionsHandle)
	adminCaseLibrary.GET("/cases/:caseId", multihttp.GetHistoricalCaseDetailHandle)
	adminCaseLibrary.DELETE("/cases/:caseId", multihttp.DeleteHistoricalCaseHandle)

	adminReview := api.Group("/scam/review")
	adminReview.Use(middleware.AdminMiddleware(authUserReader))
	adminReview.GET("/cases", multihttp.GetPendingReviewCasesHandle)
	adminReview.GET("/cases/:recordId", multihttp.GetPendingReviewCaseDetailHandle)
	adminReview.POST("/cases/:recordId/approve", multihttp.ApprovePendingReviewCaseHandle)
	adminReview.POST("/cases/:recordId/reject", multihttp.RejectPendingReviewCaseHandle)

	adminCaseCollection := api.Group("/scam/case-collection")
	adminCaseCollection.Use(middleware.AdminMiddleware(authUserReader))
	adminCaseCollection.POST("/search", multihttp.CollectCaseCollectionHandle)
}
