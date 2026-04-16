package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"antifraud/internal/modules/login/adapters/outbound/session"
	"antifraud/internal/modules/login/adapters/outbound/smscode"
	authcore "antifraud/internal/modules/login/domain/auth"
	"antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/login/domain/settings"

	"golang.org/x/crypto/bcrypt"
)

// HTTPError 用于在应用服务层携带 HTTP 语义。
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

type LoginResult struct {
	Message string              `json:"message"`
	Token   string              `json:"token"`
	User    models.UserResponse `json:"user"`
}

// AuthService 是登录系统应用服务。
type AuthService struct {
	users              UserRepository
	activeTokenManager activeTokenRegistrar
	smsService         smscode.Service
}

func NewAuthService(users UserRepository, activeTokenManager activeTokenRegistrar, smsService smscode.Service) *AuthService {
	if users == nil {
		users = defaultUserRepository()
	}
	if activeTokenManager == nil {
		activeTokenManager = session.NewDefaultRedisActiveTokenManager()
	}
	if smsService == nil {
		smsService = smscode.NewDemoService()
	}
	return &AuthService{
		users:              users,
		activeTokenManager: activeTokenManager,
		smsService:         smsService,
	}
}

func (s *AuthService) Register(ctx context.Context, payload models.RegisterPayload) (models.UserResponse, error) {
	if !verifyCaptcha(payload.CaptchaID, payload.CaptchaCode) {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "验证码错误或已过期"}
	}
	if err := validatePasswordComplexity(payload.Password); err != nil {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	normalizedPhone, err := smscode.NormalizePhone(payload.Phone)
	if err != nil {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "手机号格式不正确，请输入 11 位大陆手机号"}
	}
	if err := s.smsService.VerifyCode(ctx, normalizedPhone, payload.SMSCode); err != nil {
		switch {
		case errors.Is(err, smscode.ErrInvalidSMSCode):
			return models.UserResponse{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "短信验证码错误"}
		case errors.Is(err, smscode.ErrInvalidPhoneFormat):
			return models.UserResponse{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "手机号格式不正确，请输入 11 位大陆手机号"}
		default:
			return models.UserResponse{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "短信验证码校验失败"}
		}
	}

	exists, err := s.users.ExistsByIdentity(ctx, payload.Email, payload.Username, normalizedPhone)
	if err != nil {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "用户查重失败"}
	}
	if exists {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusConflict, Message: "邮箱、手机号或用户名已存在"}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "密码加密失败"}
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
	if err := s.users.Create(ctx, &user); err != nil {
		return models.UserResponse{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "用户创建失败"}
	}
	return models.ToUserResponse(user), nil
}

func (s *AuthService) Login(ctx context.Context, payload models.LoginPayload) (LoginResult, error) {
	var user models.User

	mode, err := resolveLoginMode(payload)
	if err != nil {
		return LoginResult{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: err.Error()}
	}

	switch mode {
	case loginModeSMS:
		normalizedPhone, err := smscode.NormalizePhone(payload.Phone)
		if err != nil {
			return LoginResult{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "手机号格式不正确，请输入 11 位大陆手机号"}
		}
		if err := s.smsService.VerifyCode(ctx, normalizedPhone, payload.SMSCode); err != nil {
			switch {
			case errors.Is(err, smscode.ErrInvalidSMSCode):
				return LoginResult{}, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "手机号或短信验证码不正确"}
			case errors.Is(err, smscode.ErrInvalidPhoneFormat):
				return LoginResult{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "手机号格式不正确，请输入 11 位大陆手机号"}
			default:
				return LoginResult{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "短信验证码校验失败"}
			}
		}
		user, err = s.users.FindByPhone(ctx, normalizedPhone)
		if err != nil {
			return LoginResult{}, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "手机号或短信验证码不正确"}
		}
	default:
		if !verifyCaptcha(payload.CaptchaID, payload.CaptchaCode) {
			return LoginResult{}, &HTTPError{StatusCode: http.StatusBadRequest, Message: "验证码错误或已过期"}
		}
		account := resolvePasswordLoginAccount(payload)
		queryField, queryValue := resolveAccountLookup(account)
		user, err = s.users.FindByAccount(ctx, queryField, queryValue)
		if err != nil {
			return LoginResult{}, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "账号或密码不正确"}
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
			return LoginResult{}, &HTTPError{StatusCode: http.StatusUnauthorized, Message: "账号或密码不正确"}
		}
	}

	tokenString, err := authcore.IssueToken(user.ID, user.Email, user.Username)
	if err != nil {
		return LoginResult{}, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "生成 Token 失败"}
	}
	if err := s.activeTokenManager.RegisterToken(ctx, user.ID, tokenString); err != nil {
		log.Printf("register active login token degraded, allow login: user_id=%d err=%v", user.ID, err)
	}

	return LoginResult{
		Message: "登录成功",
		Token:   tokenString,
		User:    models.ToUserResponse(user),
	}, nil
}

func (s *AuthService) DeleteCurrentUser(ctx context.Context, userID interface{}) error {
	if err := s.users.DeleteByID(ctx, userID); err != nil {
		return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "删除用户失败"}
	}
	return nil
}

func (s *AuthService) UpgradeUser(ctx context.Context, userID interface{}, inviteCode string) error {
	if inviteCode != settings.GetAdminInviteCode() {
		return &HTTPError{StatusCode: http.StatusForbidden, Message: "无效的邀请码"}
	}
	if err := s.users.UpdateRole(ctx, userID, "admin"); err != nil {
		return &HTTPError{StatusCode: http.StatusInternalServerError, Message: "升级失败"}
	}
	return nil
}

func (s *AuthService) ListUsers(ctx context.Context, query string) ([]models.UserResponse, error) {
	users, err := s.users.List(ctx, query)
	if err != nil {
		return nil, &HTTPError{StatusCode: http.StatusInternalServerError, Message: "获取用户列表失败"}
	}
	result := make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, models.ToUserResponse(user))
	}
	return result, nil
}
