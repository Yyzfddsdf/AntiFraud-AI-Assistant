package user_profile_system

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"antifraud/database"
	loginmodel "antifraud/login_system/models"
	"antifraud/multi_agent/overview"
	"antifraud/multi_agent/state"
)

const occupationsConfigPath = "config/occupations.json"

const (
	maxRecentTagsCount     = 10
	maxRecentTagRuneLength = 120
	defaultDisplayUserName = "user"
)

type UpdateProfileInput struct {
	Age        int
	Occupation string
}

type UserRiskTrendAnalysis struct {
	Interval       string `json:"interval"`
	CurrentBucket  string `json:"current_bucket"`
	PreviousBucket string `json:"previous_bucket,omitempty"`
	OverallTrend   string `json:"overall_trend"`
	HighRiskTrend  string `json:"high_risk_trend"`
	Summary        string `json:"summary"`
}

type UserRiskInfo struct {
	UserName          string                `json:"user_name"`
	Age               *int                  `json:"age"`
	Occupation        string                `json:"occupation,omitempty"`
	RecentTags        []string              `json:"recent_tags"`
	TotalCaseCount    int                   `json:"total_case_count"`
	HistoricalScore   int                   `json:"historical_score"`
	HighRiskCaseRatio float64               `json:"high_risk_case_ratio"`
	MidRiskCaseRatio  float64               `json:"mid_risk_case_ratio"`
	LowRiskCaseRatio  float64               `json:"low_risk_case_ratio"`
	RiskTrendAnalysis UserRiskTrendAnalysis `json:"risk_trend_analysis"`
}

var (
	occupationsMu    sync.RWMutex
	occupationsCache []string
)

func ListOccupations() []string {
	occupationsMu.RLock()
	if len(occupationsCache) > 0 {
		cached := append([]string{}, occupationsCache...)
		occupationsMu.RUnlock()
		return cached
	}
	occupationsMu.RUnlock()

	items := loadOccupations()

	occupationsMu.Lock()
	occupationsCache = append([]string{}, items...)
	occupationsMu.Unlock()
	return append([]string{}, items...)
}

func GetCurrentUserResponse(userID interface{}) (loginmodel.UserResponse, error) {
	user, err := getUserByAnyID(userID)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}
	return loginmodel.ToUserResponse(user), nil
}

func UpdateCurrentUserProfile(userID uint, input UpdateProfileInput) (loginmodel.UserResponse, error) {
	if database.DB == nil {
		return loginmodel.UserResponse{}, fmt.Errorf("main db is not initialized")
	}
	if input.Age < 1 || input.Age > 150 {
		return loginmodel.UserResponse{}, fmt.Errorf("age must be between 1 and 150")
	}

	occupation, err := normalizeOccupation(input.Occupation)
	if err != nil {
		return loginmodel.UserResponse{}, err
	}

	age := input.Age
	result := database.DB.Model(&loginmodel.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"age":        &age,
			"occupation": occupation,
		})
	if result.Error != nil {
		return loginmodel.UserResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return loginmodel.UserResponse{}, fmt.Errorf("user not found")
	}

	return GetCurrentUserResponse(userID)
}

func UpdateRecentTagsByStringUserID(userID string, recentTags []string) ([]string, error) {
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

	result := database.DB.Model(&loginmodel.User{}).
		Where("id = ?", uint(numericUserID)).
		Update("recent_tags", recentTagsJSON)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return append([]string{}, normalizedRecentTags...), nil
}

func BuildUserRiskInfo(userID string, interval string) (UserRiskInfo, error) {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	history := state.GetCaseHistory(trimmedUserID)
	userName := fmt.Sprintf("%s-%s", defaultDisplayUserName, trimmedUserID)
	var age *int
	occupation := ""
	recentTags := []string{}

	if numericID, err := parseNumericUserID(trimmedUserID); err == nil {
		if user, queryErr := getUserByAnyID(numericID); queryErr == nil {
			userName = firstNonEmpty(strings.TrimSpace(user.Username), userName)
			age = user.Age
			occupation = strings.TrimSpace(user.Occupation)
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
	riskOverview := overview.BuildRiskOverviewFromHistory(trimmedUserID, history, strings.TrimSpace(interval))
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

func normalizeOccupation(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}

	for _, item := range ListOccupations() {
		if strings.TrimSpace(item) == trimmed {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("occupation is invalid, allowed values: %s", strings.Join(ListOccupations(), ", "))
}

func normalizeRecentTags(tags []string) ([]string, error) {
	normalized := make([]string, 0, len(tags))
	seen := make(map[string]struct{}, len(tags))
	for _, item := range tags {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		if utf8.RuneCountInString(tag) > maxRecentTagRuneLength {
			return nil, fmt.Errorf("recent_tag is too long, max %d characters", maxRecentTagRuneLength)
		}
		if _, exists := seen[tag]; exists {
			continue
		}
		seen[tag] = struct{}{}
		normalized = append(normalized, tag)
		if len(normalized) > maxRecentTagsCount {
			return nil, fmt.Errorf("recent_tags count must be <= %d", maxRecentTagsCount)
		}
	}
	return normalized, nil
}

func encodeRecentTags(tags []string) (string, error) {
	if len(tags) == 0 {
		return "[]", nil
	}
	payload, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func calculateRatio(count int, total int) float64 {
	if total <= 0 {
		return 0
	}
	ratio := float64(count) / float64(total)
	return float64(int(ratio*10000+0.5)) / 10000
}

func calculateHistoricalScore(history []state.CaseHistoryRecord) int {
	if len(history) == 0 {
		return 0
	}

	now := time.Now().UTC()
	totalWeighted := 0.0
	for _, item := range history {
		baseScore := item.RiskScore
		if baseScore <= 0 {
			baseScore = fallbackRiskScoreByLevel(item.RiskLevel)
		}
		ageHours := now.Sub(item.CreatedAt.UTC()).Hours()
		if ageHours < 0 {
			ageHours = 0
		}
		timeDecay := 1.0 / (1.0 + ageHours/240.0)
		severityBoost := 1.0
		switch normalizeRiskLevel(item.RiskLevel) {
		case "高":
			severityBoost = 1.2
		case "中":
			severityBoost = 1.0
		default:
			severityBoost = 0.7
		}
		totalWeighted += float64(baseScore) * timeDecay * severityBoost
	}

	historicalScore := int(100 * (1 - math.Exp(-totalWeighted/140.0)))
	if historicalScore < 0 {
		return 0
	}
	if historicalScore > 100 {
		return 100
	}
	return historicalScore
}

func fallbackRiskScoreByLevel(level string) int {
	switch normalizeRiskLevel(level) {
	case "高":
		return 78
	case "中":
		return 52
	default:
		return 26
	}
}

func getUserByAnyID(userID interface{}) (loginmodel.User, error) {
	if database.DB == nil {
		return loginmodel.User{}, fmt.Errorf("main db is not initialized")
	}

	var user loginmodel.User
	if err := database.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return loginmodel.User{}, err
	}
	return user, nil
}

func parseNumericUserID(raw string) (uint64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("user_id is empty")
	}
	userID, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("user_id is invalid")
	}
	return userID, nil
}

func loadOccupations() []string {
	content, err := os.ReadFile(resolveConfigPath(occupationsConfigPath))
	if err != nil {
		return []string{}
	}

	var raw []string
	if err := json.Unmarshal(content, &raw); err != nil {
		return []string{}
	}

	normalized := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func resolveConfigPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	_, currentFile, _, ok := runtime.Caller(0)
	if ok {
		projectRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), ".."))
		candidate := filepath.Join(projectRoot, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return path
}

func normalizeRiskLevel(raw string) string {
	switch strings.TrimSpace(raw) {
	case "高":
		return "高"
	case "低":
		return "低"
	default:
		return "中"
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
