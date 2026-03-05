package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	appcfg "antifraud/config"

	"github.com/redis/go-redis/v9"
)

const (
	defaultRedisAddr = "127.0.0.1:6379"
)

var (
	redisClientOnce sync.Once
	redisClient     *redis.Client
	redisClientErr  error

	getDelScript = redis.NewScript(`
local value = redis.call('GET', KEYS[1])
if not value then
  return nil
end
redis.call('DEL', KEYS[1])
return value
`)

	incrWithWindowScript = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
return count
`)
)

func redisCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func getRedisClient() (*redis.Client, error) {
	redisClientOnce.Do(func() {
		cfg, err := appcfg.LoadConfig("config/config.json")
		if err != nil {
			redisClientErr = fmt.Errorf("load config for redis failed: %w", err)
			return
		}

		addr := strings.TrimSpace(cfg.Redis.Addr)
		if addr == "" {
			addr = defaultRedisAddr
		}

		db := cfg.Redis.DB
		if db < 0 {
			db = 0
		}

		client := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: strings.TrimSpace(cfg.Redis.Password),
			DB:       db,
		})

		ctx := context.Background()
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			redisClientErr = fmt.Errorf("ping redis failed: %w", err)
			return
		}

		redisClient = client
	})

	if redisClientErr != nil {
		return nil, redisClientErr
	}
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is not initialized")
	}
	return redisClient, nil
}

func normalizeKey(key string) (string, error) {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return "", fmt.Errorf("cache key is empty")
	}
	return trimmed, nil
}

func SetJSON(key string, value interface{}, ttl time.Duration) error {
	return SetJSONWithContext(context.Background(), key, value, ttl)
}

func SetJSONWithContext(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return err
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal cache value failed: %w", err)
	}

	rdb, err := getRedisClient()
	if err != nil {
		return err
	}

	if err := rdb.Set(redisCtx(ctx), normalizedKey, payload, ttl).Err(); err != nil {
		return fmt.Errorf("set cache value failed: %w", err)
	}
	return nil
}

func GetJSON(key string, out interface{}) (bool, error) {
	return GetJSONWithContext(context.Background(), key, out)
}

func GetJSONWithContext(ctx context.Context, key string, out interface{}) (bool, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return false, err
	}
	if out == nil {
		return false, fmt.Errorf("cache output receiver is nil")
	}

	rdb, err := getRedisClient()
	if err != nil {
		return false, err
	}

	raw, err := rdb.Get(redisCtx(ctx), normalizedKey).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("get cache value failed: %w", err)
	}

	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return false, fmt.Errorf("unmarshal cache value failed: %w", err)
	}
	return true, nil
}

func GetDelJSON(key string, out interface{}) (bool, error) {
	return GetDelJSONWithContext(context.Background(), key, out)
}

func GetDelJSONWithContext(ctx context.Context, key string, out interface{}) (bool, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return false, err
	}
	if out == nil {
		return false, fmt.Errorf("cache output receiver is nil")
	}

	rdb, err := getRedisClient()
	if err != nil {
		return false, err
	}

	result, err := getDelScript.Run(redisCtx(ctx), rdb, []string{normalizedKey}).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("get and delete cache value failed: %w", err)
	}
	if result == nil {
		return false, nil
	}

	raw, err := redisResultToString(result)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return false, fmt.Errorf("unmarshal get-del cache value failed: %w", err)
	}
	return true, nil
}

func Delete(key string) error {
	return DeleteWithContext(context.Background(), key)
}

func DeleteWithContext(ctx context.Context, key string) error {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return err
	}

	rdb, err := getRedisClient()
	if err != nil {
		return err
	}

	if err := rdb.Del(redisCtx(ctx), normalizedKey).Err(); err != nil {
		return fmt.Errorf("delete cache key failed: %w", err)
	}
	return nil
}

func IncrWithinWindow(key string, window time.Duration) (int64, error) {
	return IncrWithinWindowWithContext(context.Background(), key, window)
}

func IncrWithinWindowWithContext(ctx context.Context, key string, window time.Duration) (int64, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return 0, err
	}
	if window <= 0 {
		return 0, fmt.Errorf("window must be greater than 0")
	}

	windowMS := window.Milliseconds()
	if windowMS <= 0 {
		windowMS = 1
	}

	rdb, err := getRedisClient()
	if err != nil {
		return 0, err
	}

	count, err := incrWithWindowScript.Run(redisCtx(ctx), rdb, []string{normalizedKey}, windowMS).Int64()
	if err != nil {
		return 0, fmt.Errorf("incr cache counter failed: %w", err)
	}
	return count, nil
}

func TTL(key string) (time.Duration, error) {
	return TTLWithContext(context.Background(), key)
}

func TTLWithContext(ctx context.Context, key string) (time.Duration, error) {
	normalizedKey, err := normalizeKey(key)
	if err != nil {
		return 0, err
	}

	rdb, err := getRedisClient()
	if err != nil {
		return 0, err
	}

	ttl, err := rdb.TTL(redisCtx(ctx), normalizedKey).Result()
	if err != nil {
		return 0, fmt.Errorf("read cache ttl failed: %w", err)
	}
	return ttl, nil
}

func HashSetJSON(hashKey string, field string, value interface{}) error {
	return HashSetJSONWithContext(context.Background(), hashKey, field, value)
}

func HashSetJSONWithContext(ctx context.Context, hashKey string, field string, value interface{}) error {
	normalizedHashKey, err := normalizeKey(hashKey)
	if err != nil {
		return err
	}
	normalizedField, err := normalizeKey(field)
	if err != nil {
		return fmt.Errorf("cache hash field is invalid: %w", err)
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal hash cache value failed: %w", err)
	}

	rdb, err := getRedisClient()
	if err != nil {
		return err
	}

	if err := rdb.HSet(redisCtx(ctx), normalizedHashKey, normalizedField, payload).Err(); err != nil {
		return fmt.Errorf("set hash cache value failed: %w", err)
	}
	return nil
}

func HashGetAll(hashKey string) (map[string]string, error) {
	return HashGetAllWithContext(context.Background(), hashKey)
}

func HashGetAllWithContext(ctx context.Context, hashKey string) (map[string]string, error) {
	normalizedHashKey, err := normalizeKey(hashKey)
	if err != nil {
		return nil, err
	}

	rdb, err := getRedisClient()
	if err != nil {
		return nil, err
	}

	values, err := rdb.HGetAll(redisCtx(ctx), normalizedHashKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get hash cache values failed: %w", err)
	}
	if values == nil {
		return map[string]string{}, nil
	}
	return values, nil
}

func HashDelete(hashKey string, field string) error {
	return HashDeleteWithContext(context.Background(), hashKey, field)
}

func HashDeleteWithContext(ctx context.Context, hashKey string, field string) error {
	normalizedHashKey, err := normalizeKey(hashKey)
	if err != nil {
		return err
	}
	normalizedField, err := normalizeKey(field)
	if err != nil {
		return fmt.Errorf("cache hash field is invalid: %w", err)
	}

	rdb, err := getRedisClient()
	if err != nil {
		return err
	}

	if err := rdb.HDel(redisCtx(ctx), normalizedHashKey, normalizedField).Err(); err != nil {
		return fmt.Errorf("delete hash cache field failed: %w", err)
	}
	return nil
}

func redisResultToString(value interface{}) (string, error) {
	switch typed := value.(type) {
	case string:
		return typed, nil
	case []byte:
		return string(typed), nil
	default:
		return "", fmt.Errorf("unexpected redis script result type: %T", value)
	}
}
