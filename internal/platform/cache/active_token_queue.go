package cache

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var touchBoundedTokenQueueScript = redis.NewScript(`
local queueKey = KEYS[1]
local currentToken = ARGV[1]
local tokenKeyPrefix = ARGV[2]
local maxTokens = tonumber(ARGV[3])
local ttlMs = tonumber(ARGV[4])

if not maxTokens or maxTokens <= 0 then
  return redis.error_reply('maxTokens must be positive')
end
if not ttlMs or ttlMs <= 0 then
  return redis.error_reply('ttlMs must be positive')
end

local queuedTokens = redis.call('LRANGE', queueKey, 0, -1)
for _, tokenDigest in ipairs(queuedTokens) do
  if redis.call('EXISTS', tokenKeyPrefix .. tokenDigest) == 0 then
    redis.call('LREM', queueKey, 0, tokenDigest)
  end
end

redis.call('LREM', queueKey, 0, currentToken)
redis.call('RPUSH', queueKey, currentToken)
redis.call('SET', tokenKeyPrefix .. currentToken, '1', 'PX', ttlMs)
redis.call('PEXPIRE', queueKey, ttlMs)

local evicted = {}
while redis.call('LLEN', queueKey) > maxTokens do
  local oldest = redis.call('LPOP', queueKey)
  if not oldest then
    break
  end
  redis.call('DEL', tokenKeyPrefix .. oldest)
  table.insert(evicted, oldest)
end

return evicted
`)

var ensureTokenAllowedScript = redis.NewScript(`
local queueKey = KEYS[1]
local currentToken = ARGV[1]
local tokenKeyPrefix = ARGV[2]
local maxTokens = tonumber(ARGV[3])
local ttlMs = tonumber(ARGV[4])

if not maxTokens or maxTokens <= 0 then
  return redis.error_reply('maxTokens must be positive')
end
if not ttlMs or ttlMs <= 0 then
  return redis.error_reply('ttlMs must be positive')
end

local queuedTokens = redis.call('LRANGE', queueKey, 0, -1)
for _, tokenDigest in ipairs(queuedTokens) do
  if redis.call('EXISTS', tokenKeyPrefix .. tokenDigest) == 0 then
    redis.call('LREM', queueKey, 0, tokenDigest)
  end
end

if redis.call('EXISTS', tokenKeyPrefix .. currentToken) == 1 then
  redis.call('LREM', queueKey, 0, currentToken)
  redis.call('RPUSH', queueKey, currentToken)
  redis.call('PEXPIRE', tokenKeyPrefix .. currentToken, ttlMs)
  redis.call('PEXPIRE', queueKey, ttlMs)
  return 1
end

if redis.call('LLEN', queueKey) >= maxTokens then
  return 0
end

redis.call('RPUSH', queueKey, currentToken)
redis.call('SET', tokenKeyPrefix .. currentToken, '1', 'PX', ttlMs)
redis.call('PEXPIRE', queueKey, ttlMs)
return 1
`)

// TouchBoundedTokenQueueWithContext 在 Redis 中维护一个带 TTL 的活跃 token 队列：
// 1) 先清理已过期 token；
// 2) 把当前 token 视为最近活跃并刷新 TTL；
// 3) 当数量超过上限时按队列语义挤掉最旧 token。
func TouchBoundedTokenQueueWithContext(ctx context.Context, queueKey string, tokenKeyPrefix string, tokenID string, maxTokens int, ttl time.Duration) ([]string, error) {
	normalizedQueueKey, err := normalizeKey(queueKey)
	if err != nil {
		return nil, err
	}
	normalizedTokenPrefix, err := normalizeKey(tokenKeyPrefix)
	if err != nil {
		return nil, fmt.Errorf("token key prefix is invalid: %w", err)
	}
	normalizedTokenID := strings.TrimSpace(tokenID)
	if normalizedTokenID == "" {
		return nil, fmt.Errorf("token id is empty")
	}
	if maxTokens <= 0 {
		return nil, fmt.Errorf("maxTokens must be greater than 0")
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("ttl must be greater than 0")
	}
	ttlMS := ttl.Milliseconds()
	if ttlMS <= 0 {
		ttlMS = 1
	}

	rdb, err := getRedisClient()
	if err != nil {
		return nil, err
	}

	result, err := touchBoundedTokenQueueScript.Run(
		redisCtx(ctx),
		rdb,
		[]string{normalizedQueueKey},
		normalizedTokenID,
		normalizedTokenPrefix,
		maxTokens,
		ttlMS,
	).Result()
	if err != nil {
		return nil, fmt.Errorf("touch bounded token queue failed: %w", err)
	}

	if result == nil {
		return []string{}, nil
	}

	rawItems, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected token queue script result type: %T", result)
	}

	evicted := make([]string, 0, len(rawItems))
	for _, item := range rawItems {
		value, valueErr := redisResultToString(item)
		if valueErr != nil {
			return nil, valueErr
		}
		evicted = append(evicted, strings.TrimSpace(value))
	}
	return evicted, nil
}

func TouchBoundedTokenQueue(queueKey string, tokenKeyPrefix string, tokenID string, maxTokens int, ttl time.Duration) ([]string, error) {
	return TouchBoundedTokenQueueWithContext(context.Background(), queueKey, tokenKeyPrefix, tokenID, maxTokens, ttl)
}

// EnsureTokenAllowedWithContext 校验当前 token 是否仍允许作为活跃 token 使用：
// 1) 若 token 已活跃，则刷新其活跃时间并允许继续使用；
// 2) 若 token 未活跃但当前活跃数未达上限，则允许重新加入；
// 3) 若 token 未活跃且当前活跃数已达上限，则拒绝。
func EnsureTokenAllowedWithContext(ctx context.Context, queueKey string, tokenKeyPrefix string, tokenID string, maxTokens int, ttl time.Duration) (bool, error) {
	normalizedQueueKey, err := normalizeKey(queueKey)
	if err != nil {
		return false, err
	}
	normalizedTokenPrefix, err := normalizeKey(tokenKeyPrefix)
	if err != nil {
		return false, fmt.Errorf("token key prefix is invalid: %w", err)
	}
	normalizedTokenID := strings.TrimSpace(tokenID)
	if normalizedTokenID == "" {
		return false, fmt.Errorf("token id is empty")
	}
	if maxTokens <= 0 {
		return false, fmt.Errorf("maxTokens must be greater than 0")
	}
	if ttl <= 0 {
		return false, fmt.Errorf("ttl must be greater than 0")
	}
	ttlMS := ttl.Milliseconds()
	if ttlMS <= 0 {
		ttlMS = 1
	}

	rdb, err := getRedisClient()
	if err != nil {
		return false, err
	}

	allowed, err := ensureTokenAllowedScript.Run(
		redisCtx(ctx),
		rdb,
		[]string{normalizedQueueKey},
		normalizedTokenID,
		normalizedTokenPrefix,
		maxTokens,
		ttlMS,
	).Int64()
	if err != nil {
		return false, fmt.Errorf("ensure token allowed failed: %w", err)
	}
	return allowed == 1, nil
}

func EnsureTokenAllowed(queueKey string, tokenKeyPrefix string, tokenID string, maxTokens int, ttl time.Duration) (bool, error) {
	return EnsureTokenAllowedWithContext(context.Background(), queueKey, tokenKeyPrefix, tokenID, maxTokens, ttl)
}
