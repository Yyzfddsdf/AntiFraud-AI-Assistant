package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"antifraud/internal/modules/login/domain/settings"
	"antifraud/internal/platform/cache"

	"github.com/gin-gonic/gin"
)

type captchaEntry struct {
	Code string `json:"code"`
}

const captchaCachePrefix = "cache:captcha:"

// GetCaptchaHandle 生成并返回彩色 SVG 验证码（Data URL）。
func GetCaptchaHandle(c *gin.Context) {
	captchaID, code, err := generateCaptchaCode(settings.CaptchaCodeLength)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码生成失败"})
		return
	}

	if err := storeCaptcha(captchaID, code, settings.CaptchaTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证码存储失败"})
		return
	}
	svg := buildCaptchaSVG(code)

	c.JSON(http.StatusOK, gin.H{
		"captchaId":    captchaID,
		"captchaImage": "data:image/svg+xml;utf8," + svg,
		"expiresIn":    int(settings.CaptchaTTL / time.Second),
	})
}

// verifyCaptcha 校验验证码：一次性消费，校验后立即删除。
func verifyCaptcha(captchaID, captchaCode string) bool {
	trimmedID := strings.TrimSpace(captchaID)
	provided := strings.TrimSpace(strings.ToUpper(captchaCode))
	if trimmedID == "" || provided == "" {
		return false
	}

	var entry captchaEntry
	found, err := cache.GetDelJSON(captchaCacheKey(trimmedID), &entry)
	if err != nil {
		log.Printf("[captcha] verify failed: id=%s err=%v", trimmedID, err)
		return false
	}
	if !found {
		return false
	}
	return strings.EqualFold(entry.Code, provided)
}

// storeCaptcha 保存验证码到 Redis，并设置 TTL。
func storeCaptcha(captchaID, code string, ttl time.Duration) error {
	trimmedID := strings.TrimSpace(captchaID)
	if trimmedID == "" {
		return fmt.Errorf("captcha id is empty")
	}

	entry := captchaEntry{
		Code: strings.ToUpper(strings.TrimSpace(code)),
	}
	return cache.SetJSON(captchaCacheKey(trimmedID), entry, ttl)
}

func captchaCacheKey(captchaID string) string {
	return captchaCachePrefix + strings.TrimSpace(captchaID)
}

// generateCaptchaCode 生成随机验证码 ID 与验证码文本。
func generateCaptchaCode(length int) (string, string, error) {
	if length <= 0 {
		length = 5
	}

	idBytes := make([]byte, settings.CaptchaIDByteLength)
	if _, err := rand.Read(idBytes); err != nil {
		return "", "", err
	}
	captchaID := hex.EncodeToString(idBytes)

	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", "", err
		}
		code[i] = charset[n.Int64()]
	}

	return captchaID, string(code), nil
}

// buildCaptchaSVG 构建带干扰线与噪点的彩色 SVG 验证码。
func buildCaptchaSVG(code string) string {
	bg := randomColor()
	line1 := randomColor()
	line2 := randomColor()
	line3 := randomColor()

	svg := fmt.Sprintf("<svg xmlns='http://www.w3.org/2000/svg' width='180' height='60' viewBox='0 0 180 60'>"+
		"<rect width='180' height='60' fill='%s'/>"+
		"<path d='M0 20 Q45 5 90 20 T180 20' stroke='%s' fill='none' stroke-width='2'/>"+
		"<path d='M0 40 Q45 55 90 40 T180 40' stroke='%s' fill='none' stroke-width='2'/>"+
		"<path d='M0 30 Q60 10 120 35 T180 30' stroke='%s' fill='none' stroke-width='1.5'/>",
		bg, line1, line2, line3)

	for i, r := range code {
		x := 20 + i*30
		y := 38 + randInt(-8, 8)
		rotate := randInt(-25, 25)
		svg += fmt.Sprintf("<text x='%d' y='%d' font-size='30' font-family='Arial, sans-serif' font-weight='700' fill='%s' transform='rotate(%d %d %d)'>%s</text>",
			x, y, randomColor(), rotate, x, y, html.EscapeString(string(r)))
	}

	for i := 0; i < 25; i++ {
		x := randInt(0, 180)
		y := randInt(0, 60)
		svg += fmt.Sprintf("<circle cx='%d' cy='%d' r='1.5' fill='%s' opacity='0.8'/>", x, y, randomColor())
	}

	svg += "</svg>"
	return svg
}

// randomColor 生成随机 RGB 颜色。
func randomColor() string {
	r := randInt(30, 220)
	g := randInt(30, 220)
	b := randInt(30, 220)
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
}

// randInt 返回 [min, max] 的随机整数。
func randInt(min, max int) int {
	if max <= min {
		return min
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return min
	}
	return min + int(n.Int64())
}
