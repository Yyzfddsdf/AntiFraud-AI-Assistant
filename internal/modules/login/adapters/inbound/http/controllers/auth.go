package controllers

import (
	"antifraud/internal/modules/login/adapters/outbound/session"
	"antifraud/internal/modules/login/adapters/outbound/smscode"
	"antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/user_profile"
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
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
	return NewDefaultAuthHandler(nil, smsService).RegisterHandle
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
	return NewDefaultAuthHandler(activeTokenManager, smsService).LoginHandle
}

// GetCurrentUserHandle 返回当前鉴权用户信息。
func GetCurrentUserHandle(c *gin.Context) {
	NewDefaultAuthHandler(nil, nil).GetCurrentUserHandle(c)
}

func queryCurrentUserResponse(userID interface{}) (models.UserResponse, error) {
	return user_profile_system.DefaultService().GetCurrentUserResponse(userID)
}

// DeleteCurrentUserHandle 删除当前登录用户。
func DeleteCurrentUserHandle(c *gin.Context) {
	NewDefaultAuthHandler(nil, nil).DeleteCurrentUserHandle(c)
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
	NewDefaultAuthHandler(nil, nil).UpgradeUserHandle(c)
}

// GetAllUsersHandle 获取所有用户列表（仅管理员可用）。
// 支持通过 ?query=xxx 搜索用户名、邮箱或手机号。
func GetAllUsersHandle(c *gin.Context) {
	NewDefaultAuthHandler(nil, nil).GetAllUsersHandle(c)
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
