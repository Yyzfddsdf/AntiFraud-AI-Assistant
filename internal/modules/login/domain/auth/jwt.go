package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"antifraud/internal/modules/login/domain/settings"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken 表示令牌缺失、签名无效或结构非法。
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken 表示令牌已过期。
	ErrExpiredToken = errors.New("token expired")
)

var jwtSecret = []byte(settings.GetJWTSecret())

// Claims 定义 JWT 中承载的用户身份字段。
type Claims struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// IssueToken 生成包含用户基础标识的 JWT。
func IssueToken(userID uint, email, username string) (string, error) {
	now := time.Now()
	tokenID, err := generateTokenID()
	if err != nil {
		return "", err
	}
	claims := &Claims{
		UserID:   userID,
		Email:    strings.TrimSpace(email),
		Username: strings.TrimSpace(username),
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(now.Add(settings.JWTExpireDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析并校验 JWT，返回标准化 claims。
func ParseToken(tokenString string) (*Claims, error) {
	trimmedToken := strings.TrimSpace(tokenString)
	if trimmedToken == "" {
		return nil, ErrInvalidToken
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(trimmedToken, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	if token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func generateTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token id failed: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
