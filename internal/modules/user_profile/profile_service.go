package user_profile_system

import (
	"context"
	"fmt"
	"time"

	loginmodel "antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	region_system "antifraud/internal/modules/region"
	"antifraud/internal/platform/cache"
	"antifraud/internal/platform/database"

	"gorm.io/gorm"
)

// UserRepository 定义用户画像服务依赖的最小用户仓储端口。
type UserRepository interface {
	GetByID(ctx context.Context, userID interface{}) (loginmodel.User, error)
	UpdateProfile(ctx context.Context, userID uint, age *int, occupation string, region region_system.RegionSelection, locationSource string) error
	UpdateRecentTags(ctx context.Context, userID uint, recentTagsJSON string) error
}

// CaseHistoryReader 定义风险画像构建所需的案件历史读取端口。
type CaseHistoryReader interface {
	GetCaseHistory(userID string) []state.CaseHistoryRecord
}

// Service 是用户画像应用服务。
type Service struct {
	userRepo      UserRepository
	historyReader CaseHistoryReader
}

// NewService 创建用户画像应用服务。
func NewService(userRepo UserRepository, historyReader CaseHistoryReader) *Service {
	return &Service{
		userRepo:      userRepo,
		historyReader: historyReader,
	}
}

// DefaultService 返回默认用户画像应用服务。
func DefaultService() *Service {
	return NewService(
		&gormUserRepository{db: database.DB},
		stateHistoryReader{},
	)
}

// GetCurrentUserResponse 返回当前用户公开信息。
func (s *Service) GetCurrentUserResponse(userID interface{}) (loginmodel.UserResponse, error) {
	user, err := s.getUserByAnyID(userID)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}
	return loginmodel.ToUserResponse(user), nil
}

// UpdateCurrentUserProfile 更新用户画像。
func (s *Service) UpdateCurrentUserProfile(userID uint, input UpdateProfileInput) (loginmodel.UserResponse, error) {
	if s == nil || s.userRepo == nil {
		return loginmodel.UserResponse{}, fmt.Errorf("用户画像服务当前不可用")
	}
	if input.Age < 1 || input.Age > 150 {
		return loginmodel.UserResponse{}, fmt.Errorf("年龄必须在 1 到 150 之间")
	}

	occupation, err := normalizeOccupation(input.Occupation)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}
	selection, err := region_system.NormalizeSelection(region_system.RegionSelection{
		ProvinceCode: input.ProvinceCode,
		ProvinceName: input.ProvinceName,
		CityCode:     input.CityCode,
		CityName:     input.CityName,
		DistrictCode: input.DistrictCode,
		DistrictName: input.DistrictName,
	})
	if err != nil {
		return loginmodel.UserResponse{}, err
	}
	locationSource, err := normalizeLocationSource(input.LocationSource)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}

	age := input.Age
	if err := s.userRepo.UpdateProfile(context.Background(), userID, &age, occupation, selection, locationSource); err != nil {
		return loginmodel.UserResponse{}, err
	}
	_ = cache.SetJSON("cache:case_library:geo_map:v1:version", fmt.Sprintf("%d", time.Now().UnixNano()), 0)
	region_system.TouchRegionCaseStatsCacheVersion()
	return s.GetCurrentUserResponse(userID)
}

// UpdateRecentTagsByStringUserID 更新用户近期标签。
func (s *Service) UpdateRecentTagsByStringUserID(userID string, recentTags []string) ([]string, error) {
	if s == nil || s.userRepo == nil {
		return nil, fmt.Errorf("用户画像服务当前不可用")
	}
	numericUserID, err := parseNumericUserID(userID)
	if err != nil {
		return nil, err
	}

	normalizedRecentTags, err := normalizeRecentTags(recentTags)
	if err != nil {
		return nil, err
	}

	recentTagsJSON, err := encodeRecentTags(normalizedRecentTags)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.UpdateRecentTags(context.Background(), uint(numericUserID), recentTagsJSON); err != nil {
		return nil, err
	}
	return append([]string{}, normalizedRecentTags...), nil
}

// BuildUserRiskInfo 构建用户风险画像。
func (s *Service) BuildUserRiskInfo(userID string, interval string) (UserRiskInfo, error) {
	trimmedUserID := normalizeDefaultUserID(userID)
	history := []state.CaseHistoryRecord{}
	if s != nil && s.historyReader != nil {
		history = s.historyReader.GetCaseHistory(trimmedUserID)
	}

	userName := fmt.Sprintf("%s-%s", defaultDisplayUserName, trimmedUserID)
	var age *int
	occupation := ""
	provinceCode := ""
	provinceName := ""
	cityCode := ""
	cityName := ""
	districtCode := ""
	districtName := ""
	locationSource := ""
	recentTags := []string{}

	if numericID, err := parseNumericUserID(trimmedUserID); err == nil {
		if user, queryErr := s.getUserByAnyID(numericID); queryErr == nil {
			userName = firstNonEmpty(user.Username, userName)
			age = user.Age
			occupation = user.Occupation
			provinceCode = user.ProvinceCode
			provinceName = user.ProvinceName
			cityCode = user.CityCode
			cityName = user.CityName
			districtCode = user.DistrictCode
			districtName = user.DistrictName
			locationSource = user.LocationSource
			recentTags = append([]string{}, loginmodel.ToUserResponse(user).RecentTags...)
		}
	}

	riskCaseCount := map[string]int{
		"低": 0,
		"中": 0,
		"高": 0,
	}
	for _, item := range history {
		level := normalizeRiskLevel(item.RiskLevel)
		riskCaseCount[level]++
	}

	totalCaseCount := len(history)
	riskOverview := buildRiskOverview(trimmedUserID, history, interval)
	historicalScore := calculateHistoricalScore(history)

	return UserRiskInfo{
		UserName:          userName,
		Age:               age,
		Occupation:        occupation,
		ProvinceCode:      provinceCode,
		ProvinceName:      provinceName,
		CityCode:          cityCode,
		CityName:          cityName,
		DistrictCode:      districtCode,
		DistrictName:      districtName,
		LocationSource:    locationSource,
		RecentTags:        recentTags,
		TotalCaseCount:    totalCaseCount,
		HistoricalScore:   historicalScore,
		HighRiskCaseRatio: calculateRatio(riskCaseCount["高"], totalCaseCount),
		MidRiskCaseRatio:  calculateRatio(riskCaseCount["中"], totalCaseCount),
		LowRiskCaseRatio:  calculateRatio(riskCaseCount["低"], totalCaseCount),
		RiskTrendAnalysis: UserRiskTrendAnalysis{
			Interval:       riskOverview.Interval,
			CurrentBucket:  riskOverview.Analysis.CurrentBucket,
			PreviousBucket: riskOverview.Analysis.PreviousBucket,
			OverallTrend:   riskOverview.Analysis.OverallTrend,
			HighRiskTrend:  riskOverview.Analysis.HighRiskTrend,
			Summary:        riskOverview.Analysis.Summary,
		},
	}, nil
}

func (s *Service) getUserByAnyID(userID interface{}) (loginmodel.User, error) {
	if s == nil || s.userRepo == nil {
		return loginmodel.User{}, fmt.Errorf("用户画像服务当前不可用")
	}
	return s.userRepo.GetByID(context.Background(), userID)
}

type gormUserRepository struct {
	db *gorm.DB
}

func (r *gormUserRepository) GetByID(ctx context.Context, userID interface{}) (loginmodel.User, error) {
	if r == nil || r.db == nil {
		return loginmodel.User{}, fmt.Errorf("主业务数据库未初始化")
	}
	var user loginmodel.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return loginmodel.User{}, err
	}
	return user, nil
}

func (r *gormUserRepository) UpdateProfile(ctx context.Context, userID uint, age *int, occupation string, region region_system.RegionSelection, locationSource string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("主业务数据库未初始化")
	}
	result := r.db.WithContext(ctx).Model(&loginmodel.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"age":             age,
			"occupation":      occupation,
			"province_code":   region.ProvinceCode,
			"province_name":   region.ProvinceName,
			"city_code":       region.CityCode,
			"city_name":       region.CityName,
			"district_code":   region.DistrictCode,
			"district_name":   region.DistrictName,
			"location_source": locationSource,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

func (r *gormUserRepository) UpdateRecentTags(ctx context.Context, userID uint, recentTagsJSON string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("主业务数据库未初始化")
	}
	result := r.db.WithContext(ctx).Model(&loginmodel.User{}).
		Where("id = ?", userID).
		Update("recent_tags", recentTagsJSON)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("用户不存在")
	}
	return nil
}

type stateHistoryReader struct{}

func (stateHistoryReader) GetCaseHistory(userID string) []state.CaseHistoryRecord {
	return state.GetCaseHistory(userID)
}
