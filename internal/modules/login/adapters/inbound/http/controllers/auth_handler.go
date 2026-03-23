package controllers

import (
	"net/http"

	"antifraud/internal/modules/login/adapters/outbound/session"
	"antifraud/internal/modules/login/adapters/outbound/smscode"
	"antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/user_profile"

	"github.com/gin-gonic/gin"
)

type profileReader interface {
	GetCurrentUserResponse(userID interface{}) (models.UserResponse, error)
}

// AuthHandler 是登录系统 HTTP 适配器。
type AuthHandler struct {
	authService    *AuthService
	profileService profileReader
}

func NewAuthHandler(authService *AuthService, profileService profileReader) *AuthHandler {
	if authService == nil {
		authService = NewAuthService(defaultUserRepository(), session.NewDefaultRedisActiveTokenManager(), smscode.NewDemoService())
	}
	if profileService == nil {
		profileService = user_profile_system.DefaultService()
	}
	return &AuthHandler{
		authService:    authService,
		profileService: profileService,
	}
}

func NewDefaultAuthHandler(activeTokenManager activeTokenRegistrar, smsService smscode.Service) *AuthHandler {
	return NewAuthHandler(NewAuthService(defaultUserRepository(), activeTokenManager, smsService), user_profile_system.DefaultService())
}

func (h *AuthHandler) RegisterHandle(c *gin.Context) {
	var payload models.RegisterPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), payload)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) LoginHandle(c *gin.Context) {
	var payload models.LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), payload)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GetCurrentUserHandle(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userResp, err := h.profileService.GetCurrentUserResponse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
		return
	}
	c.JSON(http.StatusOK, userResp)
}

func (h *AuthHandler) DeleteCurrentUserHandle(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}
	if err := h.authService.DeleteCurrentUser(c.Request.Context(), userID); err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

func (h *AuthHandler) UpgradeUserHandle(c *gin.Context) {
	var payload models.UpgradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}
	if err := h.authService.UpgradeUser(c.Request.Context(), userID, payload.InviteCode); err != nil {
		writeAuthError(c, err)
		return
	}

	userResp, err := h.profileService.GetCurrentUserResponse(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "升级成功，但获取最新信息失败", "role": "admin"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "账户已升级为管理员",
		"user":    userResp,
	})
}

func (h *AuthHandler) GetAllUsersHandle(c *gin.Context) {
	query := c.Query("query")
	users, err := h.authService.ListUsers(c.Request.Context(), query)
	if err != nil {
		writeAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

func writeAuthError(c *gin.Context, err error) {
	if httpErr, ok := err.(*HTTPError); ok {
		c.JSON(httpErr.StatusCode, gin.H{"error": httpErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "认证服务处理失败"})
}
