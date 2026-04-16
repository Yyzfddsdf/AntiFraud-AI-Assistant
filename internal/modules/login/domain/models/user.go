package models

import (
	"encoding/json"
	"strings"

	"gorm.io/gorm"
)

// User 用户表模型。
type User struct {
	gorm.Model
	Username       string  `gorm:"unique;not null" json:"username"`
	Email          string  `gorm:"unique;not null" json:"email"`
	Phone          *string `gorm:"uniqueIndex" json:"phone,omitempty"`
	Age            *int    `gorm:"default:28" json:"age"`
	Occupation     string  `gorm:"size:64" json:"occupation,omitempty"`
	ProvinceCode   string  `gorm:"size:6" json:"province_code,omitempty"`
	ProvinceName   string  `gorm:"size:32" json:"province_name,omitempty"`
	CityCode       string  `gorm:"size:6" json:"city_code,omitempty"`
	CityName       string  `gorm:"size:32" json:"city_name,omitempty"`
	DistrictCode   string  `gorm:"size:6" json:"district_code,omitempty"`
	DistrictName   string  `gorm:"size:32" json:"district_name,omitempty"`
	LocationSource string  `gorm:"size:16" json:"location_source,omitempty"`
	RecentTagsRaw  string  `gorm:"column:recent_tags;type:text" json:"-"`
	Password       string  `gorm:"not null" json:"-"`
	Role           string  `gorm:"default:'user'" json:"role"` // 用户身份，默认为 "user"
}

// LoginPayload 登录请求参数。
type LoginPayload struct {
	Account     string `json:"account,omitempty"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Password    string `json:"password,omitempty"`
	CaptchaID   string `json:"captchaId,omitempty"`
	CaptchaCode string `json:"captchaCode,omitempty"`
	SMSCode     string `json:"smsCode,omitempty"`
}

// RegisterPayload 注册请求参数。
type RegisterPayload struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
	SMSCode     string `json:"smsCode" binding:"required"`
}

// UpgradePayload 升级账户请求参数。
type UpgradePayload struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// SendSMSCodePayload 发送短信验证码请求参数。
type SendSMSCodePayload struct {
	Phone string `json:"phone" binding:"required"`
}

// SendSMSCodeResponse 发送短信验证码响应。
type SendSMSCodeResponse struct {
	Message string `json:"message"`
}

// UserResponse 对外返回的用户信息（不包含密码）。
type UserResponse struct {
	ID             uint     `json:"id"`
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	Phone          *string  `json:"phone,omitempty"`
	Age            *int     `json:"age"`
	Occupation     string   `json:"occupation,omitempty"`
	ProvinceCode   string   `json:"province_code,omitempty"`
	ProvinceName   string   `json:"province_name,omitempty"`
	CityCode       string   `json:"city_code,omitempty"`
	CityName       string   `json:"city_name,omitempty"`
	DistrictCode   string   `json:"district_code,omitempty"`
	DistrictName   string   `json:"district_name,omitempty"`
	LocationSource string   `json:"location_source,omitempty"`
	RecentTags     []string `json:"recent_tags"`
	Role           string   `json:"role"`
}

// ToUserResponse 将用户模型转换为公开响应结构。
func ToUserResponse(user User) UserResponse {
	return UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		Phone:          user.Phone,
		Age:            user.Age,
		Occupation:     strings.TrimSpace(user.Occupation),
		ProvinceCode:   strings.TrimSpace(user.ProvinceCode),
		ProvinceName:   strings.TrimSpace(user.ProvinceName),
		CityCode:       strings.TrimSpace(user.CityCode),
		CityName:       strings.TrimSpace(user.CityName),
		DistrictCode:   strings.TrimSpace(user.DistrictCode),
		DistrictName:   strings.TrimSpace(user.DistrictName),
		LocationSource: strings.TrimSpace(user.LocationSource),
		RecentTags:     decodeRecentTags(user.RecentTagsRaw),
		Role:           user.Role,
	}
}

func decodeRecentTags(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return []string{}
	}

	var tags []string
	if err := json.Unmarshal([]byte(trimmed), &tags); err != nil {
		return []string{}
	}

	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, item := range tags {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
	}
	return normalized
}
