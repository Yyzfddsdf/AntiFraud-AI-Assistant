package multi_agent

import (
	"fmt"
	"time"
)

const (
	apiMaxRetries = 3
	apiRetryDelay = 2 * time.Second
)

func callWithRetry[T any](agentName string, action string, fn func() (T, error)) (T, error) {
	var zero T
	var lastErr error

	for attempt := 1; attempt <= apiMaxRetries; attempt++ {
		result, err := fn()
		if err == nil {
			if attempt > 1 {
				fmt.Printf("[%s] 重试成功: action=%s, attempt=%d\n", agentName, action, attempt)
			}
			return result, nil
		}

		lastErr = err
		fmt.Printf("[%s] 调用失败: action=%s, attempt=%d/%d, err=%v\n", agentName, action, attempt, apiMaxRetries, err)
		if attempt < apiMaxRetries {
			backoff := time.Duration(attempt) * apiRetryDelay
			time.Sleep(backoff)
		}
	}

	return zero, fmt.Errorf("%s failed after %d attempts: %w", action, apiMaxRetries, lastErr)
}
