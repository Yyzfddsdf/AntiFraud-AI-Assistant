package family_system

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	loginmodel "antifraud/login_system/models"
	"antifraud/login_system/smscode"

	"gorm.io/gorm"
)

var (
	ErrFamilyAlreadyExists      = errors.New("当前用户已加入家庭")
	ErrNoFamily                 = errors.New("当前用户未加入家庭")
	ErrFamilyPermissionDenied   = errors.New("无权操作当前家庭")
	ErrInvalidInvitationCode    = errors.New("邀请码无效")
	ErrInvitationExpired        = errors.New("邀请已过期")
	ErrInvitationTargetMismatch = errors.New("当前账号与邀请目标不匹配")
	ErrInvitationProcessed      = errors.New("邀请已处理")
	ErrInvalidFamilyRole        = errors.New("无效的家庭角色")
	ErrInvalidInvitationTarget  = errors.New("邀请目标不能为空")
	ErrInvalidGuardianConfig    = errors.New("无效的守护关系配置")
	ErrFamilyMemberNotFound     = errors.New("家庭成员不存在")
	ErrGuardianLinkNotFound     = errors.New("守护关系不存在")
	ErrFamilyOwnerImmutable     = errors.New("家庭创建者不可移除或降级")
)

// Service 封装家庭系统业务能力。
type Service struct {
	db *gorm.DB
}

// NewService 创建家庭系统服务。
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// EnsureSchema 确保家庭系统表结构存在。
func EnsureSchema(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("family system db is nil")
	}
	return db.AutoMigrate(
		&FamilyGroupEntity{},
		&FamilyMemberEntity{},
		&FamilyInvitationEntity{},
		&FamilyGuardianLinkEntity{},
		&FamilyNotificationEntity{},
	)
}

// CreateFamily 创建家庭并把当前用户写为 owner。
func (s *Service) CreateFamily(ctx context.Context, userID uint, input CreateFamilyInput) (FamilyOverviewResponse, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyOverviewResponse{}, err
	}

	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}
	if _, err := s.getActiveMemberByUserID(ctx, userID); err == nil {
		return FamilyOverviewResponse{}, ErrFamilyAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return FamilyOverviewResponse{}, err
	}

	trimmedName := strings.TrimSpace(input.Name)
	if trimmedName == "" {
		trimmedName = fmt.Sprintf("%s的家庭", strings.TrimSpace(user.Username))
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		group := FamilyGroupEntity{
			Name:        trimmedName,
			OwnerUserID: userID,
			InviteCode:  newFamilyCode("FAM"),
			Status:      FamilyStatusActive,
		}
		if err := tx.Create(&group).Error; err != nil {
			return err
		}

		member := FamilyMemberEntity{
			FamilyID:  group.ID,
			UserID:    userID,
			Role:      FamilyMemberRoleOwner,
			Relation:  "家庭创建者",
			Status:    FamilyMemberStatusActive,
			CreatedBy: userID,
		}
		return tx.Create(&member).Error
	})
	if err != nil {
		return FamilyOverviewResponse{}, err
	}

	return s.GetMyFamily(ctx, userID)
}

// GetMyFamily 获取当前用户家庭总览。
func (s *Service) GetMyFamily(ctx context.Context, userID uint) (FamilyOverviewResponse, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyOverviewResponse{}, err
	}

	member, err := s.getActiveMemberByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FamilyOverviewResponse{
				Family:                  nil,
				CurrentMember:           nil,
				Members:                 []FamilyMemberView{},
				Invitations:             []FamilyInvitationView{},
				GuardianLinks:           []FamilyGuardianLinkView{},
				UnreadNotificationCount: 0,
			}, nil
		}
		return FamilyOverviewResponse{}, err
	}

	group, err := s.getFamilyGroupByID(ctx, member.FamilyID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}

	members, err := s.listFamilyMembers(ctx, group.ID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}
	invitations, err := s.listFamilyInvitations(ctx, group.ID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}
	links, err := s.listGuardianLinks(ctx, group.ID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}
	unreadCount, err := s.countUnreadNotifications(ctx, userID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}

	groupView, currentMemberView := buildFamilyOverviewViews(group, member, members)
	return FamilyOverviewResponse{
		Family:                  groupView,
		CurrentMember:           currentMemberView,
		Members:                 members,
		Invitations:             invitations,
		GuardianLinks:           links,
		UnreadNotificationCount: unreadCount,
	}, nil
}

// CreateInvitation 创建家庭邀请。
func (s *Service) CreateInvitation(ctx context.Context, userID uint, input CreateFamilyInvitationInput) (FamilyInvitationView, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyInvitationView{}, err
	}

	group, currentMember, err := s.mustGetOwnedFamily(ctx, userID)
	if err != nil {
		return FamilyInvitationView{}, err
	}
	_ = currentMember

	normalizedRole, err := normalizeFamilyRole(strings.TrimSpace(input.Role), false)
	if err != nil {
		return FamilyInvitationView{}, err
	}

	trimmedEmail := strings.TrimSpace(input.InviteeEmail)
	var normalizedPhone string
	if strings.TrimSpace(input.InviteePhone) != "" {
		normalizedPhone, err = smscode.NormalizePhone(input.InviteePhone)
		if err != nil {
			return FamilyInvitationView{}, err
		}
	}
	if trimmedEmail == "" && normalizedPhone == "" {
		return FamilyInvitationView{}, ErrInvalidInvitationTarget
	}

	if targetUser, err := s.findUserByInvitationTarget(ctx, trimmedEmail, normalizedPhone); err == nil {
		if _, err := s.getActiveMemberByUserID(ctx, targetUser.ID); err == nil {
			return FamilyInvitationView{}, ErrFamilyAlreadyExists
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return FamilyInvitationView{}, err
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return FamilyInvitationView{}, err
	}

	expiresInDays := input.ExpiresInDays
	if expiresInDays <= 0 {
		expiresInDays = 7
	}
	expiresAt := time.Now().Add(time.Duration(expiresInDays) * 24 * time.Hour)

	var inviteeEmail *string
	if trimmedEmail != "" {
		inviteeEmail = &trimmedEmail
	}
	var inviteePhone *string
	if normalizedPhone != "" {
		inviteePhone = &normalizedPhone
	}

	entity := FamilyInvitationEntity{
		FamilyID:      group.ID,
		InviterUserID: userID,
		InviteeEmail:  inviteeEmail,
		InviteePhone:  inviteePhone,
		Role:          normalizedRole,
		Relation:      strings.TrimSpace(input.Relation),
		InviteCode:    newFamilyCode("INV"),
		Status:        FamilyInvitationStatusPending,
		ExpiresAt:     expiresAt,
	}
	if err := s.db.WithContext(ctx).Create(&entity).Error; err != nil {
		return FamilyInvitationView{}, err
	}

	return invitationViewFromEntity(entity), nil
}

// ListInvitations 返回当前家庭邀请列表。
func (s *Service) ListInvitations(ctx context.Context, userID uint) ([]FamilyInvitationView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	group, _, err := s.mustGetFamilyByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.listFamilyInvitations(ctx, group.ID)
}

// AcceptInvitation 接受家庭邀请。
func (s *Service) AcceptInvitation(ctx context.Context, userID uint, input AcceptFamilyInvitationInput) (FamilyOverviewResponse, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyOverviewResponse{}, err
	}

	if _, err := s.getActiveMemberByUserID(ctx, userID); err == nil {
		return FamilyOverviewResponse{}, ErrFamilyAlreadyExists
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return FamilyOverviewResponse{}, err
	}

	user, err := s.getUserByID(ctx, userID)
	if err != nil {
		return FamilyOverviewResponse{}, err
	}

	trimmedCode := strings.TrimSpace(input.InviteCode)
	if trimmedCode == "" {
		return FamilyOverviewResponse{}, ErrInvalidInvitationCode
	}

	var invitation FamilyInvitationEntity
	if err := s.db.WithContext(ctx).Where("invite_code = ?", trimmedCode).First(&invitation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FamilyOverviewResponse{}, ErrInvalidInvitationCode
		}
		return FamilyOverviewResponse{}, err
	}
	if invitation.Status != FamilyInvitationStatusPending {
		return FamilyOverviewResponse{}, ErrInvitationProcessed
	}
	if invitation.ExpiresAt.Before(time.Now()) {
		return FamilyOverviewResponse{}, ErrInvitationExpired
	}
	if !invitationMatchesUser(invitation, user) {
		return FamilyOverviewResponse{}, ErrInvitationTargetMismatch
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		member := FamilyMemberEntity{
			FamilyID:  invitation.FamilyID,
			UserID:    userID,
			Role:      invitation.Role,
			Relation:  invitation.Relation,
			Status:    FamilyMemberStatusActive,
			CreatedBy: invitation.InviterUserID,
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		now := time.Now()
		return tx.Model(&FamilyInvitationEntity{}).
			Where("id = ?", invitation.ID).
			Updates(map[string]interface{}{
				"status":              FamilyInvitationStatusAccepted,
				"accepted_by_user_id": userID,
				"accepted_at":         now,
			}).Error
	})
	if err != nil {
		return FamilyOverviewResponse{}, err
	}

	return s.GetMyFamily(ctx, userID)
}

// ListMembers 返回当前家庭成员。
func (s *Service) ListMembers(ctx context.Context, userID uint) ([]FamilyMemberView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	group, _, err := s.mustGetFamilyByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.listFamilyMembers(ctx, group.ID)
}

// UpdateMember 更新家庭成员角色或关系。
func (s *Service) UpdateMember(ctx context.Context, userID uint, memberID uint, input UpdateFamilyMemberInput) (FamilyMemberView, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyMemberView{}, err
	}
	group, _, err := s.mustGetOwnedFamily(ctx, userID)
	if err != nil {
		return FamilyMemberView{}, err
	}

	var member FamilyMemberEntity
	if err := s.db.WithContext(ctx).Where("id = ? AND family_id = ?", memberID, group.ID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FamilyMemberView{}, ErrFamilyMemberNotFound
		}
		return FamilyMemberView{}, err
	}
	if member.Role == FamilyMemberRoleOwner {
		return FamilyMemberView{}, ErrFamilyOwnerImmutable
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(input.Role) != "" {
		role, err := normalizeFamilyRole(input.Role, false)
		if err != nil {
			return FamilyMemberView{}, err
		}
		updates["role"] = role
	}
	if input.Relation != "" {
		updates["relation"] = strings.TrimSpace(input.Relation)
	}
	if len(updates) == 0 {
		updates["updated_at"] = time.Now()
	}

	if err := s.db.WithContext(ctx).Model(&member).Updates(updates).Error; err != nil {
		return FamilyMemberView{}, err
	}

	return s.getMemberViewByID(ctx, member.ID)
}

// RemoveMember 移除家庭成员。
func (s *Service) RemoveMember(ctx context.Context, userID uint, memberID uint) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	group, _, err := s.mustGetOwnedFamily(ctx, userID)
	if err != nil {
		return err
	}

	var member FamilyMemberEntity
	if err := s.db.WithContext(ctx).Where("id = ? AND family_id = ?", memberID, group.ID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrFamilyMemberNotFound
		}
		return err
	}
	if member.Role == FamilyMemberRoleOwner {
		return ErrFamilyOwnerImmutable
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("family_id = ? AND (guardian_user_id = ? OR member_user_id = ?)", group.ID, member.UserID, member.UserID).Delete(&FamilyGuardianLinkEntity{}).Error; err != nil {
			return err
		}
		return tx.Delete(&member).Error
	})
}

// CreateGuardianLink 创建守护关系。
func (s *Service) CreateGuardianLink(ctx context.Context, userID uint, input CreateGuardianLinkInput) (FamilyGuardianLinkView, error) {
	if err := s.ensureReady(); err != nil {
		return FamilyGuardianLinkView{}, err
	}
	group, _, err := s.mustGetOwnedFamily(ctx, userID)
	if err != nil {
		return FamilyGuardianLinkView{}, err
	}
	if input.GuardianUserID == input.MemberUserID {
		return FamilyGuardianLinkView{}, ErrInvalidGuardianConfig
	}

	guardianMember, err := s.getFamilyMemberByUser(ctx, group.ID, input.GuardianUserID)
	if err != nil {
		return FamilyGuardianLinkView{}, ErrInvalidGuardianConfig
	}
	memberMember, err := s.getFamilyMemberByUser(ctx, group.ID, input.MemberUserID)
	if err != nil {
		return FamilyGuardianLinkView{}, ErrInvalidGuardianConfig
	}
	if guardianMember.Role != FamilyMemberRoleOwner && guardianMember.Role != FamilyMemberRoleGuardian {
		return FamilyGuardianLinkView{}, ErrInvalidGuardianConfig
	}
	if memberMember.Role == FamilyMemberRoleOwner {
		return FamilyGuardianLinkView{}, ErrInvalidGuardianConfig
	}

	entity := FamilyGuardianLinkEntity{
		FamilyID:       group.ID,
		GuardianUserID: input.GuardianUserID,
		MemberUserID:   input.MemberUserID,
		Status:         FamilyGuardianLinkStatusActive,
	}
	if err := s.db.WithContext(ctx).Where("family_id = ? AND guardian_user_id = ? AND member_user_id = ?", group.ID, input.GuardianUserID, input.MemberUserID).FirstOrCreate(&entity).Error; err != nil {
		return FamilyGuardianLinkView{}, err
	}
	return s.getGuardianLinkViewByID(ctx, entity.ID)
}

// DeleteGuardianLink 删除守护关系。
func (s *Service) DeleteGuardianLink(ctx context.Context, userID uint, linkID uint) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	group, _, err := s.mustGetOwnedFamily(ctx, userID)
	if err != nil {
		return err
	}

	result := s.db.WithContext(ctx).Where("id = ? AND family_id = ?", linkID, group.ID).Delete(&FamilyGuardianLinkEntity{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrGuardianLinkNotFound
	}
	return nil
}

// ListNotifications 返回当前用户的家庭通知。
func (s *Service) ListNotifications(ctx context.Context, userID uint) ([]FamilyNotificationView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	rows := make([]FamilyNotificationEntity, 0)
	if err := s.db.WithContext(ctx).Where("receiver_user_id = ?", userID).Order("event_at desc, created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.buildNotificationViews(ctx, rows)
}

// ListRecentUnreadNotifications 返回当前用户最近窗口内的未读家庭通知。
func (s *Service) ListRecentUnreadNotifications(ctx context.Context, userID uint, recentWindow time.Duration) ([]FamilyNotificationView, error) {
	if err := s.ensureReady(); err != nil {
		return nil, err
	}
	if recentWindow <= 0 {
		recentWindow = time.Hour
	}
	cutoff := time.Now().Add(-recentWindow)
	rows := make([]FamilyNotificationEntity, 0)
	if err := s.db.WithContext(ctx).
		Where("receiver_user_id = ? AND read_at IS NULL AND event_at >= ?", userID, cutoff).
		Order("event_at desc, created_at desc").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.buildNotificationViews(ctx, rows)
}

// MarkNotificationRead 将通知标记为已读。
func (s *Service) MarkNotificationRead(ctx context.Context, userID uint, notificationID uint) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	now := time.Now()
	result := s.db.WithContext(ctx).Model(&FamilyNotificationEntity{}).
		Where("id = ? AND receiver_user_id = ? AND read_at IS NULL", notificationID, userID).
		Update("read_at", now)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// HandleRiskEvent 在高风险事件产生后为守护人创建通知。
func (s *Service) HandleRiskEvent(ctx context.Context, event RiskEvent) error {
	if err := s.ensureReady(); err != nil {
		return err
	}
	if event.TargetUserID == 0 || strings.TrimSpace(event.RecordID) == "" || strings.TrimSpace(event.RiskLevel) != "高" {
		return nil
	}

	targetMember, err := s.getActiveMemberByUserID(ctx, event.TargetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	links := make([]FamilyGuardianLinkEntity, 0)
	if err := s.db.WithContext(ctx).
		Where("family_id = ? AND member_user_id = ? AND status = ?", targetMember.FamilyID, event.TargetUserID, FamilyGuardianLinkStatusActive).
		Find(&links).Error; err != nil {
		return err
	}
	if len(links) == 0 {
		return nil
	}

	targetUser, err := s.getUserByID(ctx, event.TargetUserID)
	if err != nil {
		return err
	}

	summary := fmt.Sprintf("家庭成员 %s 触发高风险案件，请及时核查。", strings.TrimSpace(targetUser.Username))
	for _, link := range links {
		entity := FamilyNotificationEntity{
			FamilyID:       targetMember.FamilyID,
			TargetUserID:   event.TargetUserID,
			ReceiverUserID: link.GuardianUserID,
			EventType:      FamilyNotificationTypeHighRiskCase,
			RecordID:       strings.TrimSpace(event.RecordID),
			Title:          strings.TrimSpace(event.Title),
			CaseSummary:    strings.TrimSpace(event.CaseSummary),
			ScamType:       strings.TrimSpace(event.ScamType),
			RiskLevel:      "高",
			Summary:        summary,
			EventAt:        event.CreatedAt,
		}
		if err := s.db.WithContext(ctx).Where(
			"family_id = ? AND receiver_user_id = ? AND event_type = ? AND record_id = ?",
			entity.FamilyID,
			entity.ReceiverUserID,
			entity.EventType,
			entity.RecordID,
		).FirstOrCreate(&entity).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureReady() error {
	if s == nil || s.db == nil {
		return fmt.Errorf("family system service is unavailable")
	}
	return EnsureSchema(s.db)
}

func (s *Service) getUserByID(ctx context.Context, userID uint) (loginmodel.User, error) {
	var user loginmodel.User
	err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error
	return user, err
}

func (s *Service) findUserByInvitationTarget(ctx context.Context, email string, phone string) (loginmodel.User, error) {
	var user loginmodel.User
	query := s.db.WithContext(ctx).Model(&loginmodel.User{})
	if email != "" && phone != "" {
		err := query.Where("email = ? OR phone = ?", email, phone).First(&user).Error
		return user, err
	}
	if email != "" {
		err := query.Where("email = ?", email).First(&user).Error
		return user, err
	}
	err := query.Where("phone = ?", phone).First(&user).Error
	return user, err
}

func (s *Service) getActiveMemberByUserID(ctx context.Context, userID uint) (FamilyMemberEntity, error) {
	var member FamilyMemberEntity
	err := s.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, FamilyMemberStatusActive).First(&member).Error
	return member, err
}

func (s *Service) getFamilyMemberByUser(ctx context.Context, familyID uint, userID uint) (FamilyMemberEntity, error) {
	var member FamilyMemberEntity
	err := s.db.WithContext(ctx).Where("family_id = ? AND user_id = ? AND status = ?", familyID, userID, FamilyMemberStatusActive).First(&member).Error
	return member, err
}

func (s *Service) getFamilyGroupByID(ctx context.Context, familyID uint) (FamilyGroupEntity, error) {
	var group FamilyGroupEntity
	err := s.db.WithContext(ctx).Where("id = ? AND status = ?", familyID, FamilyStatusActive).First(&group).Error
	return group, err
}

func (s *Service) mustGetFamilyByUser(ctx context.Context, userID uint) (FamilyGroupEntity, FamilyMemberEntity, error) {
	member, err := s.getActiveMemberByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return FamilyGroupEntity{}, FamilyMemberEntity{}, ErrNoFamily
		}
		return FamilyGroupEntity{}, FamilyMemberEntity{}, err
	}
	group, err := s.getFamilyGroupByID(ctx, member.FamilyID)
	if err != nil {
		return FamilyGroupEntity{}, FamilyMemberEntity{}, err
	}
	return group, member, nil
}

func (s *Service) mustGetOwnedFamily(ctx context.Context, userID uint) (FamilyGroupEntity, FamilyMemberEntity, error) {
	group, member, err := s.mustGetFamilyByUser(ctx, userID)
	if err != nil {
		return FamilyGroupEntity{}, FamilyMemberEntity{}, err
	}
	if member.Role != FamilyMemberRoleOwner {
		return FamilyGroupEntity{}, FamilyMemberEntity{}, ErrFamilyPermissionDenied
	}
	return group, member, nil
}

func (s *Service) listFamilyMembers(ctx context.Context, familyID uint) ([]FamilyMemberView, error) {
	rows := make([]FamilyMemberEntity, 0)
	if err := s.db.WithContext(ctx).Where("family_id = ? AND status = ?", familyID, FamilyMemberStatusActive).Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	users, err := s.loadUsersByIDs(ctx, collectUserIDsFromMembers(rows))
	if err != nil {
		return nil, err
	}
	result := make([]FamilyMemberView, 0, len(rows))
	for _, row := range rows {
		result = append(result, memberViewFromEntity(row, users[row.UserID]))
	}
	return result, nil
}

func (s *Service) listFamilyInvitations(ctx context.Context, familyID uint) ([]FamilyInvitationView, error) {
	rows := make([]FamilyInvitationEntity, 0)
	if err := s.db.WithContext(ctx).Where("family_id = ?", familyID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]FamilyInvitationView, 0, len(rows))
	for _, row := range rows {
		if row.Status == FamilyInvitationStatusPending && row.ExpiresAt.Before(time.Now()) {
			row.Status = FamilyInvitationStatusExpired
		}
		result = append(result, invitationViewFromEntity(row))
	}
	return result, nil
}

func (s *Service) listGuardianLinks(ctx context.Context, familyID uint) ([]FamilyGuardianLinkView, error) {
	rows := make([]FamilyGuardianLinkEntity, 0)
	if err := s.db.WithContext(ctx).Where("family_id = ? AND status = ?", familyID, FamilyGuardianLinkStatusActive).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	return s.buildGuardianLinkViews(ctx, rows)
}

func (s *Service) countUnreadNotifications(ctx context.Context, userID uint) (int, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&FamilyNotificationEntity{}).Where("receiver_user_id = ? AND read_at IS NULL", userID).Count(&count).Error
	return int(count), err
}

func (s *Service) getMemberViewByID(ctx context.Context, memberID uint) (FamilyMemberView, error) {
	var row FamilyMemberEntity
	if err := s.db.WithContext(ctx).Where("id = ?", memberID).First(&row).Error; err != nil {
		return FamilyMemberView{}, err
	}
	user, err := s.getUserByID(ctx, row.UserID)
	if err != nil {
		return FamilyMemberView{}, err
	}
	return memberViewFromEntity(row, user), nil
}

func (s *Service) getGuardianLinkViewByID(ctx context.Context, linkID uint) (FamilyGuardianLinkView, error) {
	var row FamilyGuardianLinkEntity
	if err := s.db.WithContext(ctx).Where("id = ?", linkID).First(&row).Error; err != nil {
		return FamilyGuardianLinkView{}, err
	}
	views, err := s.buildGuardianLinkViews(ctx, []FamilyGuardianLinkEntity{row})
	if err != nil {
		return FamilyGuardianLinkView{}, err
	}
	if len(views) == 0 {
		return FamilyGuardianLinkView{}, ErrGuardianLinkNotFound
	}
	return views[0], nil
}

func (s *Service) buildGuardianLinkViews(ctx context.Context, rows []FamilyGuardianLinkEntity) ([]FamilyGuardianLinkView, error) {
	userIDs := make([]uint, 0, len(rows)*2)
	for _, row := range rows {
		userIDs = append(userIDs, row.GuardianUserID, row.MemberUserID)
	}
	users, err := s.loadUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	result := make([]FamilyGuardianLinkView, 0, len(rows))
	for _, row := range rows {
		guardian := users[row.GuardianUserID]
		member := users[row.MemberUserID]
		result = append(result, FamilyGuardianLinkView{
			ID:             row.ID,
			FamilyID:       row.FamilyID,
			GuardianUserID: row.GuardianUserID,
			GuardianName:   strings.TrimSpace(guardian.Username),
			GuardianEmail:  strings.TrimSpace(guardian.Email),
			GuardianPhone:  derefString(guardian.Phone),
			MemberUserID:   row.MemberUserID,
			MemberName:     strings.TrimSpace(member.Username),
			MemberEmail:    strings.TrimSpace(member.Email),
			MemberPhone:    derefString(member.Phone),
			Status:         strings.TrimSpace(row.Status),
		})
	}
	return result, nil
}

func (s *Service) buildNotificationViews(ctx context.Context, rows []FamilyNotificationEntity) ([]FamilyNotificationView, error) {
	userIDs := make([]uint, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.TargetUserID)
	}
	users, err := s.loadUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	result := make([]FamilyNotificationView, 0, len(rows))
	for _, row := range rows {
		target := users[row.TargetUserID]
		view := FamilyNotificationView{
			ID:             row.ID,
			FamilyID:       row.FamilyID,
			TargetUserID:   row.TargetUserID,
			TargetName:     strings.TrimSpace(target.Username),
			ReceiverUserID: row.ReceiverUserID,
			EventType:      strings.TrimSpace(row.EventType),
			RecordID:       strings.TrimSpace(row.RecordID),
			Title:          strings.TrimSpace(row.Title),
			CaseSummary:    strings.TrimSpace(row.CaseSummary),
			ScamType:       strings.TrimSpace(row.ScamType),
			RiskLevel:      strings.TrimSpace(row.RiskLevel),
			Summary:        strings.TrimSpace(row.Summary),
			EventAt:        row.EventAt.Format(time.RFC3339),
		}
		if row.ReadAt != nil {
			view.ReadAt = row.ReadAt.Format(time.RFC3339)
		}
		result = append(result, view)
	}
	return result, nil
}

func (s *Service) loadUsersByIDs(ctx context.Context, ids []uint) (map[uint]loginmodel.User, error) {
	result := map[uint]loginmodel.User{}
	uniqueIDs := uniqueUint(ids)
	if len(uniqueIDs) == 0 {
		return result, nil
	}
	rows := make([]loginmodel.User, 0)
	if err := s.db.WithContext(ctx).Where("id IN ?", uniqueIDs).Find(&rows).Error; err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ID] = row
	}
	return result, nil
}

func buildFamilyOverviewViews(group FamilyGroupEntity, currentMember FamilyMemberEntity, members []FamilyMemberView) (*FamilyGroupView, *FamilyMemberView) {
	var ownerView *FamilyMemberView
	var currentView *FamilyMemberView
	guardianCount := 0
	for index := range members {
		member := members[index]
		if member.Role == FamilyMemberRoleOwner {
			ownerView = &members[index]
		}
		if member.Role == FamilyMemberRoleGuardian || member.Role == FamilyMemberRoleOwner {
			guardianCount++
		}
		if member.UserID == currentMember.UserID {
			currentView = &members[index]
		}
	}
	if ownerView == nil {
		ownerView = &FamilyMemberView{}
	}
	return &FamilyGroupView{
		ID:            group.ID,
		Name:          strings.TrimSpace(group.Name),
		OwnerUserID:   group.OwnerUserID,
		OwnerName:     strings.TrimSpace(ownerView.Username),
		OwnerEmail:    strings.TrimSpace(ownerView.Email),
		OwnerPhone:    strings.TrimSpace(ownerView.Phone),
		InviteCode:    strings.TrimSpace(group.InviteCode),
		Status:        strings.TrimSpace(group.Status),
		MemberCount:   len(members),
		GuardianCount: guardianCount,
	}, currentView
}

func invitationMatchesUser(invitation FamilyInvitationEntity, user loginmodel.User) bool {
	if invitation.InviteeEmail != nil && strings.TrimSpace(*invitation.InviteeEmail) != "" {
		if !strings.EqualFold(strings.TrimSpace(*invitation.InviteeEmail), strings.TrimSpace(user.Email)) {
			return false
		}
	}
	if invitation.InviteePhone != nil && strings.TrimSpace(*invitation.InviteePhone) != "" {
		if !strings.EqualFold(strings.TrimSpace(*invitation.InviteePhone), derefString(user.Phone)) {
			return false
		}
	}
	return true
}

func normalizeFamilyRole(raw string, allowOwner bool) (string, error) {
	switch strings.TrimSpace(raw) {
	case "", FamilyMemberRoleMember:
		return FamilyMemberRoleMember, nil
	case FamilyMemberRoleGuardian:
		return FamilyMemberRoleGuardian, nil
	case FamilyMemberRoleOwner:
		if allowOwner {
			return FamilyMemberRoleOwner, nil
		}
	}
	return "", ErrInvalidFamilyRole
}

func memberViewFromEntity(entity FamilyMemberEntity, user loginmodel.User) FamilyMemberView {
	return FamilyMemberView{
		MemberID:  entity.ID,
		FamilyID:  entity.FamilyID,
		UserID:    entity.UserID,
		Username:  strings.TrimSpace(user.Username),
		Email:     strings.TrimSpace(user.Email),
		Phone:     derefString(user.Phone),
		Role:      strings.TrimSpace(entity.Role),
		Relation:  strings.TrimSpace(entity.Relation),
		Status:    strings.TrimSpace(entity.Status),
		CreatedAt: entity.CreatedAt.Format(time.RFC3339),
	}
}

func invitationViewFromEntity(entity FamilyInvitationEntity) FamilyInvitationView {
	return FamilyInvitationView{
		ID:               entity.ID,
		FamilyID:         entity.FamilyID,
		InviterUserID:    entity.InviterUserID,
		InviteeEmail:     derefString(entity.InviteeEmail),
		InviteePhone:     derefString(entity.InviteePhone),
		Role:             strings.TrimSpace(entity.Role),
		Relation:         strings.TrimSpace(entity.Relation),
		InviteCode:       strings.TrimSpace(entity.InviteCode),
		Status:           strings.TrimSpace(entity.Status),
		ExpiresAt:        entity.ExpiresAt.Format(time.RFC3339),
		AcceptedByUserID: entity.AcceptedByUserID,
	}
}

func collectUserIDsFromMembers(rows []FamilyMemberEntity) []uint {
	result := make([]uint, 0, len(rows))
	for _, row := range rows {
		result = append(result, row.UserID)
	}
	return result
}

func uniqueUint(values []uint) []uint {
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func newFamilyCode(prefix string) string {
	bytes := make([]byte, 6)
	if _, err := rand.Read(bytes); err != nil {
		return strings.TrimSpace(prefix) + "-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return strings.TrimSpace(prefix) + "-" + strings.ToUpper(hex.EncodeToString(bytes))
}
