package family_system

import (
	"context"
	"time"
)

// UseCase 定义家庭系统 HTTP 适配器依赖的业务端口。
type UseCase interface {
	CreateFamily(ctx context.Context, userID uint, input CreateFamilyInput) (FamilyOverviewResponse, error)
	GetMyFamily(ctx context.Context, userID uint) (FamilyOverviewResponse, error)
	CreateInvitation(ctx context.Context, userID uint, input CreateFamilyInvitationInput) (FamilyInvitationView, error)
	ListInvitations(ctx context.Context, userID uint) ([]FamilyInvitationView, error)
	ListReceivedInvitations(ctx context.Context, userID uint) ([]ReceivedFamilyInvitationView, error)
	AcceptInvitation(ctx context.Context, userID uint, input AcceptFamilyInvitationInput) (FamilyOverviewResponse, error)
	ListMembers(ctx context.Context, userID uint) ([]FamilyMemberView, error)
	UpdateMember(ctx context.Context, userID uint, memberID uint, input UpdateFamilyMemberInput) (FamilyMemberView, error)
	RemoveMember(ctx context.Context, userID uint, memberID uint) error
	CreateGuardianLink(ctx context.Context, userID uint, input CreateGuardianLinkInput) (FamilyGuardianLinkView, error)
	DeleteGuardianLink(ctx context.Context, userID uint, linkID uint) error
	ListRecentUnreadNotifications(ctx context.Context, userID uint, recentWindow time.Duration) ([]FamilyNotificationView, error)
	MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) error
}
