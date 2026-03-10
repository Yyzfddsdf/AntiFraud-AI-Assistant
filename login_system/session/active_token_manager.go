package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"antifraud/cache"
)

const (
	defaultActiveTokenLimit = 2
	defaultActiveTokenTTL   = 5 * time.Minute
)

// ActiveTokenManager 定义活跃 token 管理所需的最小能力。
type ActiveTokenManager interface {
	RegisterToken(ctx context.Context, userID uint, tokenString string) error
	AllowRequestToken(ctx context.Context, userID uint, tokenString string) (bool, error)
}

type redisActiveTokenManager struct {
	maxTokens int
	ttl       time.Duration
}

// NewDefaultRedisActiveTokenManager 创建默认活跃 token 管理器。
func NewDefaultRedisActiveTokenManager() ActiveTokenManager {
	return NewRedisActiveTokenManager(defaultActiveTokenLimit, defaultActiveTokenTTL)
}

// NewRedisActiveTokenManager 创建基于 Redis 的活跃 token 管理器。
func NewRedisActiveTokenManager(maxTokens int, ttl time.Duration) ActiveTokenManager {
	if maxTokens <= 0 {
		maxTokens = defaultActiveTokenLimit
	}
	if ttl <= 0 {
		ttl = defaultActiveTokenTTL
	}
	return &redisActiveTokenManager{maxTokens: maxTokens, ttl: ttl}
}

func (m *redisActiveTokenManager) RegisterToken(ctx context.Context, userID uint, tokenString string) error {
	if m == nil {
		return fmt.Errorf("active token manager is nil")
	}
	if userID == 0 {
		return fmt.Errorf("user id is invalid")
	}
	trimmedToken := strings.TrimSpace(tokenString)
	if trimmedToken == "" {
		return fmt.Errorf("token string is empty")
	}

	tokenDigest := digestActiveToken(trimmedToken)
	queueKey := fmt.Sprintf("cache:auth:active_tokens:user:%d:queue", userID)
	tokenKeyPrefix := fmt.Sprintf("cache:auth:active_tokens:user:%d:token:", userID)

	_, err := cache.TouchBoundedTokenQueueWithContext(ctx, queueKey, tokenKeyPrefix, tokenDigest, m.maxTokens, m.ttl)
	if err != nil {
		return fmt.Errorf("register active token failed: %w", err)
	}
	return nil
}

func (m *redisActiveTokenManager) AllowRequestToken(ctx context.Context, userID uint, tokenString string) (bool, error) {
	if m == nil {
		return false, fmt.Errorf("active token manager is nil")
	}
	if userID == 0 {
		return false, fmt.Errorf("user id is invalid")
	}
	trimmedToken := strings.TrimSpace(tokenString)
	if trimmedToken == "" {
		return false, fmt.Errorf("token string is empty")
	}

	tokenDigest := digestActiveToken(trimmedToken)
	queueKey := fmt.Sprintf("cache:auth:active_tokens:user:%d:queue", userID)
	tokenKeyPrefix := fmt.Sprintf("cache:auth:active_tokens:user:%d:token:", userID)

	allowed, err := cache.EnsureTokenAllowedWithContext(ctx, queueKey, tokenKeyPrefix, tokenDigest, m.maxTokens, m.ttl)
	if err != nil {
		return false, fmt.Errorf("ensure request token allowed failed: %w", err)
	}
	return allowed, nil
}

func digestActiveToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}
