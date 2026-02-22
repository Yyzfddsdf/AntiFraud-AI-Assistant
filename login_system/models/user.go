package models

import "gorm.io/gorm"

// User 用户表模型。
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Email    string `gorm:"unique;not null" json:"email"`
	Age      *int   `json:"age"`
	Password string `gorm:"not null" json:"-"`
	Role     string `gorm:"default:'user'" json:"role"` // 用户身份，默认为 "user"
}

// LoginPayload 登录请求参数。
type LoginPayload struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

// RegisterPayload 注册请求参数。
type RegisterPayload struct {
	Username    string `json:"username" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	CaptchaID   string `json:"captchaId" binding:"required"`
	CaptchaCode string `json:"captchaCode" binding:"required"`
}

// UpgradePayload 升级账户请求参数。
type UpgradePayload struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// UserResponse 对外返回的用户信息（不包含密码）。
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Age      *int   `json:"age"`
	Role     string `json:"role"`
}
