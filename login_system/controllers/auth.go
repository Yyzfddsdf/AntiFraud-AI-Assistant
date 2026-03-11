package controllers

import (
	"antifraud/database"
	authcore "antifraud/login_system/auth"
	"antifraud/login_system/models"
	"antifraud/login_system/session"
	"antifraud/login_system/settings"
	"antifraud/login_system/smscode"
	"context"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type activeTokenRegistrar interface {
	RegisterToken(ctx context.Context, userID uint, tokenString string) error
}

type loginMode string

const (
	loginModePassword loginMode = "password"
	loginModeSMS      loginMode = "sms"
)

// RegisterHandle 处理用户注册：参数校验、验证码校验、密码强度校验、写库。
func RegisterHandle(c *gin.Context) {
	RegisterHandleWithSMSCodeService(smscode.NewDemoService())(c)
}

// RegisterHandleWithSMSCodeService 使用注入的短信验证码服务处理注册。
func RegisterHandleWithSMSCodeService(smsService smscode.Service) gin.HandlerFunc {
	if smsService == nil {
		smsService = smscode.NewDemoService()
	}

	return func(c *gin.Context) {
		var payload models.RegisterPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		if !verifyCaptcha(payload.CaptchaID, payload.CaptchaCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
			return
		}

		if err := validatePasswordComplexity(payload.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		normalizedPhone, err := smscode.NormalizePhone(payload.Phone)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
			return
		}
		if err := smsService.VerifyCode(c.Request.Context(), normalizedPhone, payload.SMSCode); err != nil {
			if errors.Is(err, smscode.ErrInvalidSMSCode) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "短信验证码错误"})
				return
			}
			if errors.Is(err, smscode.ErrInvalidPhoneFormat) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "短信验证码校验失败"})
			return
		}

		var existingUser models.User
		if err := database.DB.Where("email = ? OR username = ? OR phone = ?", payload.Email, payload.Username, normalizedPhone).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "邮箱、手机号或用户名已存在"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
			return
		}

		defaultAge := 28
		userPhone := normalizedPhone
		user := models.User{
			Username: payload.Username,
			Email:    payload.Email,
			Phone:    &userPhone,
			Age:      &defaultAge,
			Password: string(hashedPassword),
			Role:     "user",
		}

		if err := database.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
			return
		}

		c.JSON(http.StatusCreated, models.ToUserResponse(user))
	}
}

// LoginHandle 处理用户登录：参数校验、验证码校验、密码验证并签发 JWT。
func LoginHandle(c *gin.Context) {
	LoginHandleWithActiveTokenManagerAndSMSCodeService(
		session.NewDefaultRedisActiveTokenManager(),
		smscode.NewDemoService(),
	)(c)
}

// LoginHandleWithActiveTokenManager 使用注入的活跃 token 管理器处理登录。
func LoginHandleWithActiveTokenManager(activeTokenManager activeTokenRegistrar) gin.HandlerFunc {
	return LoginHandleWithActiveTokenManagerAndSMSCodeService(activeTokenManager, smscode.NewDemoService())
}

// LoginHandleWithActiveTokenManagerAndSMSCodeService 使用注入的 token 管理器与短信验证码服务处理登录。
func LoginHandleWithActiveTokenManagerAndSMSCodeService(activeTokenManager activeTokenRegistrar, smsService smscode.Service) gin.HandlerFunc {
	if activeTokenManager == nil {
		activeTokenManager = session.NewDefaultRedisActiveTokenManager()
	}
	if smsService == nil {
		smsService = smscode.NewDemoService()
	}

	return func(c *gin.Context) {
		var payload models.LoginPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		var user models.User
		mode, err := resolveLoginMode(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		switch mode {
		case loginModeSMS:
			normalizedPhone, err := smscode.NormalizePhone(payload.Phone)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
				return
			}
			if err := smsService.VerifyCode(c.Request.Context(), normalizedPhone, payload.SMSCode); err != nil {
				if errors.Is(err, smscode.ErrInvalidSMSCode) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "手机号或短信验证码不正确"})
					return
				}
				if errors.Is(err, smscode.ErrInvalidPhoneFormat) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": "短信验证码校验失败"})
				return
			}
			if err := database.DB.Where("phone = ?", normalizedPhone).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "手机号或短信验证码不正确"})
				return
			}
		default:
			if !verifyCaptcha(payload.CaptchaID, payload.CaptchaCode) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
				return
			}

			account := resolvePasswordLoginAccount(payload)
			queryField, queryValue := resolveAccountLookup(account)
			if err := database.DB.Where(queryField+" = ?", queryValue).First(&user).Error; err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码不正确"})
				return
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码不正确"})
				return
			}
		}

		tokenString, err := authcore.IssueToken(user.ID, user.Email, user.Username)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
			return
		}
		if err := activeTokenManager.RegisterToken(c.Request.Context(), user.ID, tokenString); err != nil {
			log.Printf("register active login token degraded, allow login: user_id=%d err=%v", user.ID, err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "登录成功",
			"token":   tokenString,
			"user":    models.ToUserResponse(user),
		})
	}
}

// GetCurrentUserHandle 返回当前鉴权用户信息。
func GetCurrentUserHandle(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	userResp, err := queryCurrentUserResponse(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在或已被删除"})
		return
	}

	c.JSON(http.StatusOK, userResp)
}

func queryCurrentUserResponse(userID interface{}) (models.UserResponse, error) {
	var user models.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return models.UserResponse{}, err
	}
	return models.ToUserResponse(user), nil
}

// DeleteCurrentUserHandle 删除当前登录用户。
func DeleteCurrentUserHandle(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	if err := database.DB.Unscoped().Where("id = ?", userID).Delete(&models.User{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户已删除"})
}

var (
	// 密码复杂度规则：至少一个大写、一个小写、一个符号。
	passwordUppercasePattern = regexp.MustCompile(`[A-Z]`)
	passwordLowercasePattern = regexp.MustCompile(`[a-z]`)
	passwordSymbolPattern    = regexp.MustCompile(`[^A-Za-z0-9]`)
)

// validatePasswordComplexity 校验密码复杂度策略。
func validatePasswordComplexity(password string) error {
	if !passwordUppercasePattern.MatchString(password) {
		return errors.New("密码必须包含至少一个大写字母")
	}
	if !passwordLowercasePattern.MatchString(password) {
		return errors.New("密码必须包含至少一个小写字母")
	}
	if !passwordSymbolPattern.MatchString(password) {
		return errors.New("密码必须包含至少一个符号")
	}
	return nil
}

// UpgradeUserHandle 处理用户升级请求。
func UpgradeUserHandle(c *gin.Context) {
	var payload models.UpgradePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	inviteCode := settings.GetAdminInviteCode()

	if payload.InviteCode != inviteCode {
		c.JSON(http.StatusForbidden, gin.H{"error": "无效的邀请码"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
		return
	}

	// 更新用户角色为 admin
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("role", "admin").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "升级失败"})
		return
	}

	// 重新获取用户信息以返回
	userResp, err := queryCurrentUserResponse(userID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "升级成功，但获取最新信息失败", "role": "admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "账户已升级为管理员",
		"user":    userResp,
	})
}

// GetAllUsersHandle 获取所有用户列表（仅管理员可用）。
// 支持通过 ?query=xxx 搜索用户名、邮箱或手机号。
func GetAllUsersHandle(c *gin.Context) {
	// 1. 处理搜索参数
	query := c.Query("query")
	var users []models.User
	db := database.DB.Model(&models.User{})

	if query != "" {
		db = db.Where("username LIKE ? OR email LIKE ? OR phone LIKE ?", "%"+query+"%", "%"+query+"%", "%"+query+"%")
	}

	// 2. 查询数据库
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 3. 构建响应（过滤敏感信息）
	var userResponses []models.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, models.ToUserResponse(user))
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
		"count": len(userResponses),
	})
}

func resolveLoginMode(payload models.LoginPayload) (loginMode, error) {
	hasPassword := strings.TrimSpace(payload.Password) != ""
	hasCaptchaID := strings.TrimSpace(payload.CaptchaID) != ""
	hasCaptchaCode := strings.TrimSpace(payload.CaptchaCode) != ""
	hasSMSCode := strings.TrimSpace(payload.SMSCode) != ""
	hasPhone := strings.TrimSpace(payload.Phone) != ""
	hasPasswordAccount := strings.TrimSpace(resolvePasswordLoginAccount(payload)) != ""

	if hasSMSCode {
		if hasPassword || hasCaptchaID || hasCaptchaCode || strings.TrimSpace(payload.Account) != "" || strings.TrimSpace(payload.Email) != "" {
			return "", errors.New("请只选择一种登录方式")
		}
		if !hasPhone {
			return "", errors.New("短信登录需要手机号和短信验证码")
		}
		return loginModeSMS, nil
	}

	if hasPassword || hasCaptchaID || hasCaptchaCode || hasPasswordAccount {
		if !hasPasswordAccount || !hasPassword || !hasCaptchaID || !hasCaptchaCode {
			return "", errors.New("密码登录需要账号、密码和图形验证码")
		}
		return loginModePassword, nil
	}

	return "", errors.New("登录参数不完整")
}

func resolvePasswordLoginAccount(payload models.LoginPayload) string {
	if account := strings.TrimSpace(payload.Account); account != "" {
		return account
	}
	if email := strings.TrimSpace(payload.Email); email != "" {
		return email
	}
	return strings.TrimSpace(payload.Phone)
}

func resolveAccountLookup(raw string) (string, string) {
	trimmed := strings.TrimSpace(raw)
	if normalizedPhone, err := smscode.NormalizePhone(trimmed); err == nil {
		return "phone", normalizedPhone
	}
	return "email", trimmed
}
