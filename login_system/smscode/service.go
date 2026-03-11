package smscode

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidPhoneFormat = errors.New("手机号格式不正确")
	ErrInvalidSMSCode     = errors.New("短信验证码错误")
)

var mainlandPhonePattern = regexp.MustCompile(`^1\d{10}$`)

// DemoCode 是当前演示环境固定使用的短信验证码。
const DemoCode = "000000"

// Service 定义短信验证码发送与校验的最小能力边界。
type Service interface {
	SendCode(ctx context.Context, phone string) error
	VerifyCode(ctx context.Context, phone string, code string) error
}

// DemoService 是短信验证码能力的占位实现。
type DemoService struct{}

// NewDemoService 创建演示短信验证码服务。
func NewDemoService() *DemoService {
	return &DemoService{}
}

// NormalizePhone 归一化并校验大陆手机号。
func NormalizePhone(raw string) (string, error) {
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "")
	normalized := replacer.Replace(strings.TrimSpace(raw))
	if !mainlandPhonePattern.MatchString(normalized) {
		return "", ErrInvalidPhoneFormat
	}
	return normalized, nil
}

// SendCode 预留短信发送能力，当前仅校验手机号格式。
func (s *DemoService) SendCode(ctx context.Context, phone string) error {
	_ = ctx
	_, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	// TODO: 接入真实短信服务，将验证码发送到手机号。
	return nil
}

// VerifyCode 校验短信验证码，当前演示环境固定为 000000。
func (s *DemoService) VerifyCode(ctx context.Context, phone string, code string) error {
	_ = ctx
	_, err := NormalizePhone(phone)
	if err != nil {
		return err
	}
	// TODO: 接入真实短信验证码校验与短期存储逻辑。
	if strings.TrimSpace(code) != DemoCode {
		return ErrInvalidSMSCode
	}
	return nil
}
