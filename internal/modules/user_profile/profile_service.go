package user_profile_system

import (
	"context"
	"fmt"

	loginmodel "antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/platform/database"

	"gorm.io/gorm"
)

// UserRepository 定义用户画像服务依赖的最小用户仓储端口。
type UserRepository interface {
	GetByID(ctx context.Context, userID interface{}) (loginmodel.User, error)
	UpdateProfile(ctx context.Context, userID uint, age *int, occupation string) error
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
		return loginmodel.UserResponse{}, fmt.Errorf("user profile service is unavailable")
	}
	if input.Age < 1 || input.Age > 150 {
		return loginmodel.UserResponse{}, fmt.Errorf("age must be between 1 and 150")
	}

	occupation, err := normalizeOccupation(input.Occupation)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}

	age := input.Age
	if err := s.userRepo.UpdateProfile(context.Background(), userID, &age, occupation); err != nil {
		return loginmodel.UserResponse{}, err
	}
	return s.GetCurrentUserResponse(userID)
}

// UpdateRecentTagsByStringUserID 更新用户近期标签。
func (s *Service) UpdateRecentTagsByStringUserID(userID string, recentTags []string) ([]string, error) {
	if s == nil || s.userRepo == nil {
		return nil, fmt.Errorf("user profile service is unavailable")
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
	recentTags := []string{}

	if numericID, err := parseNumericUserID(trimmedUserID); err == nil {
		if user, queryErr := s.getUserByAnyID(numericID); queryErr == nil {
			userName = firstNonEmpty(user.Username, userName)
			age = user.Age
			occupation = user.Occupation
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
		return loginmodel.User{}, fmt.Errorf("user profile service is unavailable")
	}
	return s.userRepo.GetByID(context.Background(), userID)
}

type gormUserRepository struct {
	db *gorm.DB
}

func (r *gormUserRepository) GetByID(ctx context.Context, userID interface{}) (loginmodel.User, error) {
	if r == nil || r.db == nil {
		return loginmodel.User{}, fmt.Errorf("main db is not initialized")
	}
	var user loginmodel.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		return loginmodel.User{}, err
	}
	return user, nil
}

func (r *gormUserRepository) UpdateProfile(ctx context.Context, userID uint, age *int, occupation string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("main db is not initialized")
	}
	result := r.db.WithContext(ctx).Model(&loginmodel.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"age":        age,
			"occupation": occupation,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *gormUserRepository) UpdateRecentTags(ctx context.Context, userID uint, recentTagsJSON string) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("main db is not initialized")
	}
	result := r.db.WithContext(ctx).Model(&loginmodel.User{}).
		Where("id = ?", userID).
		Update("recent_tags", recentTagsJSON)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

type stateHistoryReader struct{}

func (stateHistoryReader) GetCaseHistory(userID string) []state.CaseHistoryRecord {
	return state.GetCaseHistory(userID)
}
