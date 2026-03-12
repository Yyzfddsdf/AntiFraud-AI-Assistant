package family_system_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"antifraud/family_system"
	loginmodel "antifraud/login_system/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestService(t *testing.T) (*family_system.Service, *gorm.DB) {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&loginmodel.User{}); err != nil {
		t.Fatalf("migrate users failed: %v", err)
	}
	if err := family_system.EnsureSchema(db); err != nil {
		t.Fatalf("migrate family system failed: %v", err)
	}
	return family_system.NewService(db), db
}

func createUser(t *testing.T, db *gorm.DB, username, email, phone string) loginmodel.User {
	t.Helper()
	userPhone := phone
	age := 28
	user := loginmodel.User{
		Username: username,
		Email:    email,
		Phone:    &userPhone,
		Age:      &age,
		Password: "hashed-password",
		Role:     "user",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	return user
}

func TestCreateFamily(t *testing.T) {
	service, db := newTestService(t)
	user := createUser(t, db, "owner_user", "owner@example.com", "13800138000")

	overview, err := service.CreateFamily(context.Background(), user.ID, family_system.CreateFamilyInput{Name: "测试家庭"})
	if err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	if overview.Family == nil || overview.Family.Name != "测试家庭" {
		t.Fatalf("unexpected family overview: %+v", overview.Family)
	}
	if overview.CurrentMember == nil || overview.CurrentMember.Role != family_system.FamilyMemberRoleOwner {
		t.Fatalf("unexpected current member: %+v", overview.CurrentMember)
	}
}

func TestInvitationAcceptFlow(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "owner_user", "owner@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}

	invitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "父亲",
	})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}

	overview, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: invitation.InviteCode,
	})
	if err != nil {
		t.Fatalf("accept invitation failed: %v", err)
	}
	if len(overview.Members) != 2 {
		t.Fatalf("expected 2 family members, got: %d", len(overview.Members))
	}
}

func TestListReceivedInvitations(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "owner_user", "owner@example.com", "13800138000")
	target := createUser(t, db, "member_user", "member@example.com", "13900139000")
	other := createUser(t, db, "other_user", "other@example.com", "13700137000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	if _, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleGuardian,
		Relation:     "母亲",
	}); err != nil {
		t.Fatalf("create target invitation failed: %v", err)
	}
	if _, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "other@example.com",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "朋友",
	}); err != nil {
		t.Fatalf("create other invitation failed: %v", err)
	}

	invitations, err := service.ListReceivedInvitations(context.Background(), target.ID)
	if err != nil {
		t.Fatalf("list received invitations failed: %v", err)
	}
	if len(invitations) != 1 {
		t.Fatalf("expected 1 received invitation, got: %d", len(invitations))
	}
	invitation := invitations[0]
	if invitation.FamilyName != "测试家庭" {
		t.Fatalf("unexpected family name: %+v", invitation)
	}
	if invitation.InviterName != "owner_user" {
		t.Fatalf("unexpected inviter name: %+v", invitation)
	}
	if invitation.InviteeEmail != "member@example.com" || invitation.InviteePhone != "13900139000" {
		t.Fatalf("unexpected invitation target: %+v", invitation)
	}
	if invitation.Role != family_system.FamilyMemberRoleGuardian || invitation.Relation != "母亲" {
		t.Fatalf("unexpected invitation role info: %+v", invitation)
	}

	otherInvitations, err := service.ListReceivedInvitations(context.Background(), other.ID)
	if err != nil {
		t.Fatalf("list other received invitations failed: %v", err)
	}
	if len(otherInvitations) != 1 {
		t.Fatalf("expected 1 invitation for other user, got: %d", len(otherInvitations))
	}
}

func TestAcceptInvitationDeletesAllInvitationsForUser(t *testing.T) {
	service, db := newTestService(t)
	ownerA := createUser(t, db, "owner_a", "ownera@example.com", "13800138000")
	ownerB := createUser(t, db, "owner_b", "ownerb@example.com", "13700137000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), ownerA.ID, family_system.CreateFamilyInput{Name: "家庭A"}); err != nil {
		t.Fatalf("create family A failed: %v", err)
	}
	if _, err := service.CreateFamily(context.Background(), ownerB.ID, family_system.CreateFamilyInput{Name: "家庭B"}); err != nil {
		t.Fatalf("create family B failed: %v", err)
	}

	invitationA, err := service.CreateInvitation(context.Background(), ownerA.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "家人",
	})
	if err != nil {
		t.Fatalf("create invitation A failed: %v", err)
	}
	if _, err := service.CreateInvitation(context.Background(), ownerB.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleGuardian,
		Relation:     "监护人",
	}); err != nil {
		t.Fatalf("create invitation B failed: %v", err)
	}

	if _, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: invitationA.InviteCode,
	}); err != nil {
		t.Fatalf("accept invitation failed: %v", err)
	}

	received, err := service.ListReceivedInvitations(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("list received invitations failed: %v", err)
	}
	if len(received) != 0 {
		t.Fatalf("expected no received invitations after accept, got: %d", len(received))
	}

	ownerAInvitations, err := service.ListInvitations(context.Background(), ownerA.ID)
	if err != nil {
		t.Fatalf("list owner A invitations failed: %v", err)
	}
	if len(ownerAInvitations) != 0 {
		t.Fatalf("expected owner A invitations to be deleted, got: %d", len(ownerAInvitations))
	}

	ownerBInvitations, err := service.ListInvitations(context.Background(), ownerB.ID)
	if err != nil {
		t.Fatalf("list owner B invitations failed: %v", err)
	}
	if len(ownerBInvitations) != 0 {
		t.Fatalf("expected owner B invitations to be deleted, got: %d", len(ownerBInvitations))
	}

	var count int64
	if err := db.Unscoped().
		Model(&family_system.FamilyInvitationEntity{}).
		Where("LOWER(invitee_email) = LOWER(?) OR invitee_phone = ?", "member@example.com", "13900139000").
		Count(&count).Error; err != nil {
		t.Fatalf("count invitation rows failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected invitation rows to be physically deleted, got: %d", count)
	}
}

func TestExpiredInvitationIsDeletedAutomatically(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "owner_user", "owner@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	invitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "家人",
	})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	if err := db.Model(&family_system.FamilyInvitationEntity{}).
		Where("id = ?", invitation.ID).
		Update("expires_at", time.Now().Add(-2*time.Hour)).Error; err != nil {
		t.Fatalf("expire invitation failed: %v", err)
	}

	received, err := service.ListReceivedInvitations(context.Background(), member.ID)
	if err != nil {
		t.Fatalf("list received invitations failed: %v", err)
	}
	if len(received) != 0 {
		t.Fatalf("expected expired invitation to be hidden, got: %d", len(received))
	}

	ownerInvitations, err := service.ListInvitations(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("list owner invitations failed: %v", err)
	}
	if len(ownerInvitations) != 0 {
		t.Fatalf("expected expired invitation to be deleted for owner, got: %d", len(ownerInvitations))
	}

	var count int64
	if err := db.Unscoped().
		Model(&family_system.FamilyInvitationEntity{}).
		Where("id = ?", invitation.ID).
		Count(&count).Error; err != nil {
		t.Fatalf("count invitation rows failed: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected expired invitation to be physically deleted, got: %d", count)
	}
}

func TestRemovedMemberCanRejoin(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "owner_user", "owner@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	firstInvitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "家人",
	})
	if err != nil {
		t.Fatalf("create first invitation failed: %v", err)
	}
	overview, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: firstInvitation.InviteCode,
	})
	if err != nil {
		t.Fatalf("accept first invitation failed: %v", err)
	}

	var memberView family_system.FamilyMemberView
	for _, row := range overview.Members {
		if row.UserID == member.ID {
			memberView = row
			break
		}
	}
	if memberView.MemberID == 0 {
		t.Fatalf("member view not found after first join")
	}

	if err := service.RemoveMember(context.Background(), owner.ID, memberView.MemberID); err != nil {
		t.Fatalf("remove member failed: %v", err)
	}

	secondInvitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleGuardian,
		Relation:     "再次加入",
	})
	if err != nil {
		t.Fatalf("create second invitation failed: %v", err)
	}
	rejoinOverview, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: secondInvitation.InviteCode,
	})
	if err != nil {
		t.Fatalf("rejoin should succeed after removal, got: %v", err)
	}

	if len(rejoinOverview.Members) != 2 {
		t.Fatalf("expected 2 family members after rejoin, got: %d", len(rejoinOverview.Members))
	}
}

func TestDeletedGuardianLinkCanRecreate(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "owner_user", "owner@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	invitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteeEmail: "member@example.com",
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "家人",
	})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	if _, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: invitation.InviteCode,
	}); err != nil {
		t.Fatalf("accept invitation failed: %v", err)
	}

	link, err := service.CreateGuardianLink(context.Background(), owner.ID, family_system.CreateGuardianLinkInput{
		GuardianUserID: owner.ID,
		MemberUserID:   member.ID,
	})
	if err != nil {
		t.Fatalf("create guardian link failed: %v", err)
	}

	if err := service.DeleteGuardianLink(context.Background(), owner.ID, link.ID); err != nil {
		t.Fatalf("delete guardian link failed: %v", err)
	}

	recreated, err := service.CreateGuardianLink(context.Background(), owner.ID, family_system.CreateGuardianLinkInput{
		GuardianUserID: owner.ID,
		MemberUserID:   member.ID,
	})
	if err != nil {
		t.Fatalf("recreate guardian link failed: %v", err)
	}
	if recreated.ID == 0 {
		t.Fatalf("expected recreated guardian link id")
	}
}

func TestHighRiskEventCreatesNotification(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "guardian_user", "guardian@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	invitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "父亲",
	})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	if _, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: invitation.InviteCode,
	}); err != nil {
		t.Fatalf("accept invitation failed: %v", err)
	}

	if _, err := service.CreateGuardianLink(context.Background(), owner.ID, family_system.CreateGuardianLinkInput{
		GuardianUserID: owner.ID,
		MemberUserID:   member.ID,
	}); err != nil {
		t.Fatalf("create guardian link failed: %v", err)
	}

	err = service.HandleRiskEvent(context.Background(), family_system.RiskEvent{
		TargetUserID: member.ID,
		RecordID:     "TASK-001",
		Title:        "疑似高风险案件",
		CaseSummary:  "存在高风险诈骗迹象",
		ScamType:     "冒充客服类",
		RiskLevel:    "高",
		CreatedAt:    time.Now(),
	})
	if err != nil {
		t.Fatalf("handle risk event failed: %v", err)
	}

	notifications, err := service.ListNotifications(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("list notifications failed: %v", err)
	}
	if len(notifications) != 1 {
		t.Fatalf("expected 1 notification, got: %d", len(notifications))
	}
	if notifications[0].TargetUserID != member.ID {
		t.Fatalf("unexpected notification target: %+v", notifications[0])
	}
}

func TestRemoveMemberClearsRelatedNotifications(t *testing.T) {
	service, db := newTestService(t)
	owner := createUser(t, db, "guardian_user", "guardian@example.com", "13800138000")
	member := createUser(t, db, "member_user", "member@example.com", "13900139000")

	if _, err := service.CreateFamily(context.Background(), owner.ID, family_system.CreateFamilyInput{Name: "测试家庭"}); err != nil {
		t.Fatalf("create family failed: %v", err)
	}
	invitation, err := service.CreateInvitation(context.Background(), owner.ID, family_system.CreateFamilyInvitationInput{
		InviteePhone: "13900139000",
		Role:         family_system.FamilyMemberRoleMember,
		Relation:     "父亲",
	})
	if err != nil {
		t.Fatalf("create invitation failed: %v", err)
	}
	if _, err := service.AcceptInvitation(context.Background(), member.ID, family_system.AcceptFamilyInvitationInput{
		InviteCode: invitation.InviteCode,
	}); err != nil {
		t.Fatalf("accept invitation failed: %v", err)
	}

	if _, err := service.CreateGuardianLink(context.Background(), owner.ID, family_system.CreateGuardianLinkInput{
		GuardianUserID: owner.ID,
		MemberUserID:   member.ID,
	}); err != nil {
		t.Fatalf("create guardian link failed: %v", err)
	}

	if err := service.HandleRiskEvent(context.Background(), family_system.RiskEvent{
		TargetUserID: member.ID,
		RecordID:     "TASK-001",
		Title:        "疑似高风险案件",
		CaseSummary:  "存在高风险诈骗迹象",
		ScamType:     "冒充客服类",
		RiskLevel:    "高",
		CreatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("handle risk event failed: %v", err)
	}

	members, err := service.ListMembers(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("list members failed: %v", err)
	}
	var memberID uint
	for _, item := range members {
		if item.UserID == member.ID {
			memberID = item.MemberID
			break
		}
	}
	if memberID == 0 {
		t.Fatalf("member id not found for user %d", member.ID)
	}

	if err := service.RemoveMember(context.Background(), owner.ID, memberID); err != nil {
		t.Fatalf("remove member failed: %v", err)
	}

	notifications, err := service.ListNotifications(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("list notifications failed: %v", err)
	}
	if len(notifications) != 0 {
		t.Fatalf("expected notifications to be removed, got: %d", len(notifications))
	}

	recent, err := service.ListRecentUnreadNotifications(context.Background(), owner.ID, time.Hour)
	if err != nil {
		t.Fatalf("list recent unread notifications failed: %v", err)
	}
	if len(recent) != 0 {
		t.Fatalf("expected no recent notifications, got: %d", len(recent))
	}

	if err := service.HandleRiskEvent(context.Background(), family_system.RiskEvent{
		TargetUserID: member.ID,
		RecordID:     "TASK-002",
		Title:        "再次触发高风险案件",
		CaseSummary:  "成员已移除后不应再推送",
		ScamType:     "冒充客服类",
		RiskLevel:    "高",
		CreatedAt:    time.Now(),
	}); err != nil {
		t.Fatalf("handle risk event after removal failed: %v", err)
	}

	notifications, err = service.ListNotifications(context.Background(), owner.ID)
	if err != nil {
		t.Fatalf("list notifications after removal failed: %v", err)
	}
	if len(notifications) != 0 {
		t.Fatalf("expected no notifications after removal, got: %d", len(notifications))
	}
}
