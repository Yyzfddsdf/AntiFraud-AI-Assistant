package user_profile_system_test

import (
	"fmt"
	"testing"

	loginmodel "antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	userprofile "antifraud/internal/modules/user_profile"
	"antifraud/internal/platform/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&loginmodel.User{}); err != nil {
		t.Fatalf("migrate users failed: %v", err)
	}

	originalDB := database.DB
	database.DB = db
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func createUser(t *testing.T, db *gorm.DB, username string) loginmodel.User {
	t.Helper()

	phone := "13800138000"
	age := 28
	user := loginmodel.User{
		Username: username,
		Email:    username + "@example.com",
		Phone:    &phone,
		Age:      &age,
		Password: "hashed-password",
		Role:     "user",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user failed: %v", err)
	}
	return user
}

func TestUpdateCurrentUserProfile(t *testing.T) {
	db := newTestDB(t)
	user := createUser(t, db, "profile_user")

	resp, err := userprofile.UpdateCurrentUserProfile(user.ID, userprofile.UpdateProfileInput{
		Age:            32,
		Occupation:     "企业职员",
		ProvinceCode:   "310000",
		ProvinceName:   "上海市",
		CityCode:       "310000",
		CityName:       "上海市",
		DistrictCode:   "310115",
		DistrictName:   "浦东新区",
		LocationSource: "manual",
	})
	if err != nil {
		t.Fatalf("update profile failed: %v", err)
	}
	recentTags, err := userprofile.UpdateRecentTagsByStringUserID(fmt.Sprintf("%d", user.ID), []string{"近期频繁网购", "正在找工作", "近期频繁网购"})
	if err != nil {
		t.Fatalf("update recent tags failed: %v", err)
	}

	if resp.Age == nil || *resp.Age != 32 {
		t.Fatalf("unexpected age: %+v", resp.Age)
	}
	if resp.Occupation != "企业职员" {
		t.Fatalf("unexpected occupation: %q", resp.Occupation)
	}
	if resp.DistrictCode != "310115" || resp.DistrictName != "浦东新区" {
		t.Fatalf("unexpected district: %s %s", resp.DistrictCode, resp.DistrictName)
	}
	if resp.LocationSource != "manual" {
		t.Fatalf("unexpected location_source: %q", resp.LocationSource)
	}
	if len(recentTags) != 2 {
		t.Fatalf("expected deduped recent tags, got %+v", recentTags)
	}
}

func TestBuildUserRiskInfoIncludesRatios(t *testing.T) {
	db := newTestDB(t)
	user := createUser(t, db, "risk_user")

	if _, err := userprofile.UpdateCurrentUserProfile(user.ID, userprofile.UpdateProfileInput{
		Age:            28,
		Occupation:     "学生",
		ProvinceCode:   "110000",
		ProvinceName:   "北京市",
		CityCode:       "110000",
		CityName:       "北京市",
		DistrictCode:   "110108",
		DistrictName:   "海淀区",
		LocationSource: "auto",
	}); err != nil {
		t.Fatalf("update profile failed: %v", err)
	}
	if _, err := userprofile.UpdateRecentTagsByStringUserID(fmt.Sprintf("%d", user.ID), []string{"近期准备求职"}); err != nil {
		t.Fatalf("update recent tags failed: %v", err)
	}

	state.AddCaseHistory(fmt.Sprintf("%d", user.ID), "TASK-1", "高风险案件", "summary", "其他诈骗类", "高", 82, `{"score":82}`, state.TaskPayload{}, "report")
	state.AddCaseHistory(fmt.Sprintf("%d", user.ID), "TASK-2", "中风险案件", "summary", "其他诈骗类", "中", 56, `{"score":56}`, state.TaskPayload{}, "report")
	state.AddCaseHistory(fmt.Sprintf("%d", user.ID), "TASK-3", "低风险案件", "summary", "其他诈骗类", "低", 24, `{"score":24}`, state.TaskPayload{}, "report")

	info, err := userprofile.BuildUserRiskInfo(fmt.Sprintf("%d", user.ID), "day")
	if err != nil {
		t.Fatalf("build user risk info failed: %v", err)
	}

	if info.UserName != "risk_user" {
		t.Fatalf("unexpected user_name: %q", info.UserName)
	}
	if info.Occupation != "学生" {
		t.Fatalf("unexpected occupation: %q", info.Occupation)
	}
	if info.DistrictCode != "110108" || info.DistrictName != "海淀区" {
		t.Fatalf("unexpected district: %s %s", info.DistrictCode, info.DistrictName)
	}
	if info.LocationSource != "auto" {
		t.Fatalf("unexpected location_source: %q", info.LocationSource)
	}
	if len(info.RecentTags) != 1 || info.RecentTags[0] != "近期准备求职" {
		t.Fatalf("unexpected recent_tags: %+v", info.RecentTags)
	}
	if info.TotalCaseCount != 3 {
		t.Fatalf("unexpected total_case_count: %d", info.TotalCaseCount)
	}
	if info.HistoricalScore <= 0 {
		t.Fatalf("expected historical_score > 0, got %d", info.HistoricalScore)
	}
	if info.HighRiskCaseRatio != 0.3333 || info.MidRiskCaseRatio != 0.3333 || info.LowRiskCaseRatio != 0.3333 {
		t.Fatalf("unexpected ratios: high=%v mid=%v low=%v", info.HighRiskCaseRatio, info.MidRiskCaseRatio, info.LowRiskCaseRatio)
	}
}
