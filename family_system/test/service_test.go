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
