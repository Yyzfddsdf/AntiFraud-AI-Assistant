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

	loginmodel "antifraud/internal/modules/login/domain/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/state"
	"antifraud/internal/modules/multi_agent/domain/overview"
)

const occupationsConfigPath = "internal/platform/config/occupations.json"

const (
	maxRecentTagsCount     = 10
	maxRecentTagRuneLength = 120
	defaultDisplayUserName = "user"
)

type UpdateProfileInput struct {
	Age            int
	Occupation     string
	ProvinceCode   string
	ProvinceName   string
	CityCode       string
	CityName       string
	DistrictCode   string
	DistrictName   string
	LocationSource string
}

type UserRiskTrendAnalysis struct {
	Interval       string `json:"interval"`
	CurrentWindow  string `json:"current_window"`
	PreviousWindow string `json:"previous_window,omitempty"`
	OverallTrend   string `json:"overall_trend"`
	HighRiskTrend  string `json:"high_risk_trend"`
	Summary        string `json:"summary"`
}

type UserRiskInfo struct {
	UserName          string                `json:"user_name"`
	Age               *int                  `json:"age"`
	Occupation        string                `json:"occupation,omitempty"`
	ProvinceCode      string                `json:"province_code,omitempty"`
	ProvinceName      string                `json:"province_name,omitempty"`
	CityCode          string                `json:"city_code,omitempty"`
	CityName          string                `json:"city_name,omitempty"`
	DistrictCode      string                `json:"district_code,omitempty"`
	DistrictName      string                `json:"district_name,omitempty"`
	LocationSource    string                `json:"location_source,omitempty"`
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
	return loadCachedOptions(&occupationsMu, &occupationsCache, occupationsConfigPath)
}

func GetCurrentUserResponse(userID interface{}) (loginmodel.UserResponse, error) {
	return DefaultService().GetCurrentUserResponse(userID)
}

func UpdateCurrentUserProfile(userID uint, input UpdateProfileInput) (loginmodel.UserResponse, error) {
	return DefaultService().UpdateCurrentUserProfile(userID, input)
}

func UpdateRecentTagsByStringUserID(userID string, recentTags []string) ([]string, error) {
	return DefaultService().UpdateRecentTagsByStringUserID(userID, recentTags)
}

func BuildUserRiskInfo(userID string, interval string) (UserRiskInfo, error) {
	return DefaultService().BuildUserRiskInfo(userID, interval)
}

func buildRiskOverview(userID string, history []state.CaseHistoryRecord, interval string) overview.UserRiskOverview {
	return overview.BuildRiskOverviewFromHistory(normalizeDefaultUserID(userID), history, strings.TrimSpace(interval))
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
	return "", fmt.Errorf("职业无效，可选值为：%s", strings.Join(ListOccupations(), ", "))
}

func normalizeLocationSource(raw string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(raw))
	if trimmed == "" {
		return "", nil
	}
	switch trimmed {
	case "manual", "auto":
		return trimmed, nil
	default:
		return "", fmt.Errorf("位置来源无效，可选值为：manual、auto")
	}
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
			return nil, fmt.Errorf("近期标签数量不能超过 %d 个", maxRecentTagsCount)
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
	return DefaultService().getUserByAnyID(userID)
}

func parseNumericUserID(raw string) (uint64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, fmt.Errorf("user_id is empty")
	}
	userID, err := strconv.ParseUint(trimmed, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("用户 ID 无效")
	}
	return userID, nil
}

func loadOccupations() []string {
	return loadStringOptions(occupationsConfigPath)
}

func loadCachedOptions(mu *sync.RWMutex, cache *[]string, configPath string) []string {
	mu.RLock()
	if len(*cache) > 0 {
		cached := append([]string{}, (*cache)...)
		mu.RUnlock()
		return cached
	}
	mu.RUnlock()

	items := loadStringOptions(configPath)

	mu.Lock()
	*cache = append([]string{}, items...)
	mu.Unlock()
	return append([]string{}, items...)
}

func loadStringOptions(configPath string) []string {
	content, err := os.ReadFile(resolveConfigPath(configPath))
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
		currentDir := filepath.Dir(currentFile)
		for i := 0; i < 8; i++ {
			candidate := filepath.Join(currentDir, path)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			parent := filepath.Dir(currentDir)
			if parent == currentDir {
				break
			}
			currentDir = parent
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

func normalizeDefaultUserID(userID string) string {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return "demo-user"
	}
	return trimmed
}
