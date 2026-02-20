package settings

import "time"

const (
	// JWTExpireDuration: JWT 令牌默认有效期。
	JWTExpireDuration = 24 * time.Hour

	// CaptchaCodeLength: 验证码字符长度。
	CaptchaCodeLength = 5
	// CaptchaIDByteLength: 验证码 ID 随机字节长度（会转为十六进制字符串）。
	CaptchaIDByteLength = 12
	// CaptchaTTL: 验证码有效期。
	CaptchaTTL = 3 * time.Minute
	// CaptchaCleanupInterval: 过期验证码清理协程的执行间隔。
	CaptchaCleanupInterval = 1 * time.Minute

	// RateLimitWindow: 限流统计时间窗口。
	RateLimitWindow = 1 * time.Second
	// RateLimitMaxRequests: 单个 IP 在窗口期内允许的最大请求数。
	RateLimitMaxRequests = 5
)
