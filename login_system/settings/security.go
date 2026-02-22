package settings

import (
	"os"
	"time"
)

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
	RateLimitMaxRequests = 100

	// DefaultJWTSecret: JWT 签名密钥的默认值。
	// 当环境变量 JWT_SECRET 未设置时，将使用此值。
	// 生产环境请务必设置环境变量以保证安全。
	DefaultJWTSecret = "change_me_to_a_strong_secret_in_production"

	// DefaultAdminInviteCode: 管理员账户升级邀请码的默认值。
	// 当环境变量 INVITE_CODE_ADMIN 未设置时，将使用此值。
	// 建议在生产环境中通过环境变量设置复杂的邀请码。
	DefaultAdminInviteCode = "Secret_Admin_Invite_Code_2026"
)

// GetJWTSecret 获取 JWT 签名密钥。
// 优先读取环境变量 JWT_SECRET；若未设置，则返回 DefaultJWTSecret。
func GetJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}
	return DefaultJWTSecret
}

// GetAdminInviteCode 获取管理员升级邀请码。
// 优先读取环境变量 INVITE_CODE_ADMIN；若未设置，则返回 DefaultAdminInviteCode。
func GetAdminInviteCode() string {
	if code := os.Getenv("INVITE_CODE_ADMIN"); code != "" {
		return code
	}
	return DefaultAdminInviteCode
}
