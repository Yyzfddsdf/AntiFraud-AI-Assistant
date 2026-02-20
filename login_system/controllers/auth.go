package controllers

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"time"

	"image_recognition/login_system/database"
	"image_recognition/login_system/models"
	"image_recognition/login_system/settings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecretFromEnv())

// getJWTSecretFromEnv 读取 JWT 密钥；未配置时使用默认值（开发环境）。
func getJWTSecretFromEnv() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "change_me_to_a_strong_secret_in_production"
	}
	return secret
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// generateJWT 生成包含用户基础标识的 JWT。
func generateJWT(userID uint, email, username string) (string, error) {
	expirationTime := time.Now().Add(settings.JWTExpireDuration)
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// RegisterHandle 处理用户注册：参数校验、验证码校验、密码强度校验、写库。
func RegisterHandle(c *gin.Context) {
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

	var existingUser models.User
	if err := database.DB.Where("email = ? OR username = ?", payload.Email, payload.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "邮箱或用户名已存在"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}

	user := models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
		return
	}

	c.JSON(http.StatusCreated, models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}

// LoginHandle 处理用户登录：参数校验、验证码校验、密码验证并签发 JWT。
func LoginHandle(c *gin.Context) {
	var payload models.LoginPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	if !verifyCaptcha(payload.CaptchaID, payload.CaptchaCode) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码错误或已过期"})
		return
	}

	var user models.User
	if err := database.DB.Where("email = ?", payload.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码不正确"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "邮箱或密码不正确"})
		return
	}

	tokenString, err := generateJWT(user.ID, user.Email, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成 Token 失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   tokenString,
		"user": models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
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
	return models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Age:      user.Age,
	}, nil
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

// GetJWTSecret 对外提供 JWT 密钥读取（供鉴权中间件使用）。
func GetJWTSecret() []byte {
	return jwtSecret
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
