package family_system

import (
	"time"

	"gorm.io/gorm"
)

const (
	FamilyStatusActive = "active"

	FamilyMemberRoleOwner    = "owner"
	FamilyMemberRoleGuardian = "guardian"
	FamilyMemberRoleMember   = "member"

	FamilyMemberStatusActive = "active"

	FamilyInvitationStatusPending = "pending"
	FamilyInvitationStatusRevoked = "revoked"

	FamilyGuardianLinkStatusActive = "active"

	FamilyNotificationTypeHighRiskCase = "high_risk_case"
)

// FamilyGroupEntity 表示家庭组。
type FamilyGroupEntity struct {
	gorm.Model
	Name        string `gorm:"size:128;not null"`
	OwnerUserID uint   `gorm:"index;not null"`
	InviteCode  string `gorm:"size:32;uniqueIndex;not null"`
	Status      string `gorm:"size:32;index;not null;default:'active'"`
}

func (FamilyGroupEntity) TableName() string {
	return "family_groups"
}

// FamilyMemberEntity 表示家庭成员关系。
type FamilyMemberEntity struct {
	gorm.Model
	FamilyID  uint   `gorm:"index;not null"`
	UserID    uint   `gorm:"index;not null;uniqueIndex:idx_family_user"`
	Role      string `gorm:"size:32;index;not null"`
	Relation  string `gorm:"size:64"`
	Status    string `gorm:"size:32;index;not null;default:'active'"`
	CreatedBy uint   `gorm:"index;not null"`
}

func (FamilyMemberEntity) TableName() string {
	return "family_members"
}

// FamilyInvitationEntity 表示家庭邀请记录。
type FamilyInvitationEntity struct {
	gorm.Model
	FamilyID      uint      `gorm:"index;not null"`
	InviterUserID uint      `gorm:"index;not null"`
	InviteeEmail  *string   `gorm:"size:255;index"`
	InviteePhone  *string   `gorm:"size:32;index"`
	Role          string    `gorm:"size:32;index;not null"`
	Relation      string    `gorm:"size:64"`
	InviteCode    string    `gorm:"size:32;uniqueIndex;not null"`
	Status        string    `gorm:"size:32;index;not null;default:'pending'"`
	ExpiresAt     time.Time `gorm:"index;not null"`
}

func (FamilyInvitationEntity) TableName() string {
	return "family_invitations"
}

// FamilyGuardianLinkEntity 表示守护人配置关系。
type FamilyGuardianLinkEntity struct {
	gorm.Model
	FamilyID       uint   `gorm:"index;not null"`
	GuardianUserID uint   `gorm:"index;not null;uniqueIndex:idx_family_guardian_member"`
	MemberUserID   uint   `gorm:"index;not null;uniqueIndex:idx_family_guardian_member"`
	Status         string `gorm:"size:32;index;not null;default:'active'"`
}

func (FamilyGuardianLinkEntity) TableName() string {
	return "family_guardian_links"
}

// FamilyNotificationEntity 表示家庭通知。
type FamilyNotificationEntity struct {
	gorm.Model
	FamilyID       uint       `gorm:"index;not null"`
	TargetUserID   uint       `gorm:"index;not null"`
	ReceiverUserID uint       `gorm:"index;not null;uniqueIndex:idx_family_receiver_record_event"`
	EventType      string     `gorm:"size:64;index;not null;uniqueIndex:idx_family_receiver_record_event"`
	RecordID       string     `gorm:"size:64;index;not null;uniqueIndex:idx_family_receiver_record_event"`
	Title          string     `gorm:"size:255;not null"`
	CaseSummary    string     `gorm:"type:text"`
	ScamType       string     `gorm:"size:64;index"`
	RiskLevel      string     `gorm:"size:32;index"`
	Summary        string     `gorm:"type:text;not null"`
	EventAt        time.Time  `gorm:"index;not null"`
	ReadAt         *time.Time `gorm:"index"`
}

func (FamilyNotificationEntity) TableName() string {
	return "family_notifications"
}

// CreateFamilyInput 创建家庭请求。
type CreateFamilyInput struct {
	Name string `json:"name" binding:"required"`
}

// CreateFamilyInvitationInput 创建家庭邀请请求。
type CreateFamilyInvitationInput struct {
	InviteeEmail  string `json:"invitee_email,omitempty"`
	InviteePhone  string `json:"invitee_phone,omitempty"`
	Role          string `json:"role,omitempty"`
	Relation      string `json:"relation,omitempty"`
	ExpiresInDays int    `json:"expires_in_days,omitempty"`
}

// AcceptFamilyInvitationInput 接受邀请请求。
type AcceptFamilyInvitationInput struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

// UpdateFamilyMemberInput 更新成员请求。
type UpdateFamilyMemberInput struct {
	Role     string `json:"role,omitempty"`
	Relation string `json:"relation,omitempty"`
}

// CreateGuardianLinkInput 配置守护关系请求。
type CreateGuardianLinkInput struct {
	GuardianUserID uint `json:"guardian_user_id" binding:"required"`
	MemberUserID   uint `json:"member_user_id" binding:"required"`
}

// FamilyGroupView 是家庭组返回结构。
type FamilyGroupView struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	OwnerUserID   uint   `json:"owner_user_id"`
	OwnerName     string `json:"owner_name"`
	OwnerEmail    string `json:"owner_email"`
	OwnerPhone    string `json:"owner_phone,omitempty"`
	InviteCode    string `json:"invite_code"`
	Status        string `json:"status"`
	MemberCount   int    `json:"member_count"`
	GuardianCount int    `json:"guardian_count"`
}

// FamilyMemberView 是家庭成员返回结构。
type FamilyMemberView struct {
	MemberID  uint   `json:"member_id"`
	FamilyID  uint   `json:"family_id"`
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role"`
	Relation  string `json:"relation,omitempty"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// FamilyInvitationView 是邀请返回结构。
type FamilyInvitationView struct {
	ID            uint   `json:"id"`
	FamilyID      uint   `json:"family_id"`
	InviterUserID uint   `json:"inviter_user_id"`
	InviteeEmail  string `json:"invitee_email,omitempty"`
	InviteePhone  string `json:"invitee_phone,omitempty"`
	Role          string `json:"role"`
	Relation      string `json:"relation,omitempty"`
	InviteCode    string `json:"invite_code"`
	Status        string `json:"status"`
	ExpiresAt     string `json:"expires_at"`
}

// ReceivedFamilyInvitationView 是被邀请人收到的邀请返回结构。
type ReceivedFamilyInvitationView struct {
	ID            uint   `json:"id"`
	FamilyID      uint   `json:"family_id"`
	FamilyName    string `json:"family_name"`
	InviterUserID uint   `json:"inviter_user_id"`
	InviterName   string `json:"inviter_name"`
	InviterEmail  string `json:"inviter_email,omitempty"`
	InviterPhone  string `json:"inviter_phone,omitempty"`
	InviteeEmail  string `json:"invitee_email,omitempty"`
	InviteePhone  string `json:"invitee_phone,omitempty"`
	Role          string `json:"role"`
	Relation      string `json:"relation,omitempty"`
	InviteCode    string `json:"invite_code"`
	Status        string `json:"status"`
	ExpiresAt     string `json:"expires_at"`
}

// FamilyGuardianLinkView 是守护关系返回结构。
type FamilyGuardianLinkView struct {
	ID             uint   `json:"id"`
	FamilyID       uint   `json:"family_id"`
	GuardianUserID uint   `json:"guardian_user_id"`
	GuardianName   string `json:"guardian_name"`
	GuardianEmail  string `json:"guardian_email"`
	GuardianPhone  string `json:"guardian_phone,omitempty"`
	MemberUserID   uint   `json:"member_user_id"`
	MemberName     string `json:"member_name"`
	MemberEmail    string `json:"member_email"`
	MemberPhone    string `json:"member_phone,omitempty"`
	Status         string `json:"status"`
}

// FamilyNotificationView 是家庭通知返回结构。
type FamilyNotificationView struct {
	ID             uint   `json:"id"`
	FamilyID       uint   `json:"family_id"`
	TargetUserID   uint   `json:"target_user_id"`
	TargetName     string `json:"target_name"`
	ReceiverUserID uint   `json:"receiver_user_id"`
	EventType      string `json:"event_type"`
	RecordID       string `json:"record_id"`
	Title          string `json:"title"`
	CaseSummary    string `json:"case_summary"`
	ScamType       string `json:"scam_type,omitempty"`
	RiskLevel      string `json:"risk_level,omitempty"`
	Summary        string `json:"summary"`
	EventAt        string `json:"event_at"`
	ReadAt         string `json:"read_at,omitempty"`
}

// FamilyOverviewResponse 是家庭中心总览返回结构。
type FamilyOverviewResponse struct {
	Family                  *FamilyGroupView         `json:"family"`
	CurrentMember           *FamilyMemberView        `json:"current_member"`
	Members                 []FamilyMemberView       `json:"members"`
	Invitations             []FamilyInvitationView   `json:"invitations"`
	GuardianLinks           []FamilyGuardianLinkView `json:"guardian_links"`
	UnreadNotificationCount int                      `json:"unread_notification_count"`
}

// RiskEvent 是家庭通知依赖的最小风险事件载荷。
type RiskEvent struct {
	TargetUserID uint
	RecordID     string
	Title        string
	CaseSummary  string
	ScamType     string
	RiskLevel    string
	CreatedAt    time.Time
}
