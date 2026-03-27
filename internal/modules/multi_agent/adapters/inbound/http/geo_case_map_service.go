package httpapi

import (
	"fmt"
	"sort"
	"strings"
	"time"

	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/platform/cache"
	"antifraud/internal/platform/database"
)

const (
	geoCaseMapCacheVersionKey        = "cache:case_library:geo_map:v2:version"
	geoCaseMapOverviewCacheKeyPrefix = "cache:case_library:geo_map:v2:overview:"
	geoCaseMapChildrenCacheKeyPrefix = "cache:case_library:geo_map:v2:children:"
	geoCaseRegionCasesCacheKeyPrefix = "cache:case_library:geo_map:v2:region_cases:"
	geoCaseMapOverviewCacheTTL       = 2 * time.Minute
	geoCaseMapChildrenCacheTTL       = 2 * time.Minute
	geoCaseRegionCasesCacheTTL       = 90 * time.Second
	geoCaseMapLevelProvince          = "province"
	geoCaseMapLevelCity              = "city"
	geoCaseMapLevelDistrict          = "district"
)

type geoCaseJoinedRow struct {
	RecordID     string    `gorm:"column:record_id"`
	Title        string    `gorm:"column:title"`
	CaseSummary  string    `gorm:"column:case_summary"`
	RiskLevel    string    `gorm:"column:risk_level"`
	UserID       string    `gorm:"column:user_id"`
	ScamType     string    `gorm:"column:scam_type"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	ProvinceCode string    `gorm:"column:province_code"`
	ProvinceName string    `gorm:"column:province_name"`
	CityCode     string    `gorm:"column:city_code"`
	CityName     string    `gorm:"column:city_name"`
	DistrictCode string    `gorm:"column:district_code"`
	DistrictName string    `gorm:"column:district_name"`
}

type geoWindowAccumulator struct {
	Count         int
	PreviousCount int
	ScamTypeCount map[string]int
}

type geoRegionAggregate struct {
	RegionCode string
	RegionName string
	Today      geoWindowAccumulator
	Last7d     geoWindowAccumulator
	Last30d    geoWindowAccumulator
	AllTime    geoWindowAccumulator
}

func buildGeoCaseMapOverview() (apimodel.GeoCaseMapResponse, error) {
	cacheKey := buildGeoCaseMapOverviewCacheKey()
	var cached apimodel.GeoCaseMapResponse
	if found, err := cache.GetJSON(cacheKey, &cached); err == nil && found {
		return cached, nil
	}

	rows, err := queryGeoCaseRows()
	if err != nil {
		return apimodel.GeoCaseMapResponse{}, err
	}

	now := time.Now()
	result := apimodel.GeoCaseMapResponse{
		GeneratedAt: now.Format(time.RFC3339),
		Level:       geoCaseMapLevelProvince,
		Summary:     buildGeoCaseMapSummary(rows),
		Regions:     aggregateGeoRegions(rows, geoCaseMapLevelProvince, ""),
	}
	_ = cache.SetJSON(cacheKey, result, geoCaseMapOverviewCacheTTL)
	return result, nil
}

func buildGeoCaseMapChildren(parentCode string, level string) (apimodel.GeoCaseMapChildrenResponse, error) {
	normalizedLevel := normalizeGeoLevel(level)
	normalizedParentCode := strings.TrimSpace(parentCode)
	cacheKey := buildGeoCaseMapChildrenCacheKey(normalizedLevel, normalizedParentCode)

	var cached apimodel.GeoCaseMapChildrenResponse
	if found, err := cache.GetJSON(cacheKey, &cached); err == nil && found {
		return cached, nil
	}

	rows, err := queryGeoCaseRows()
	if err != nil {
		return apimodel.GeoCaseMapChildrenResponse{}, err
	}

	parentName := resolveGeoParentName(rows, normalizedLevel, normalizedParentCode)
	regions := aggregateGeoRegions(rows, normalizedLevel, normalizedParentCode)
	result := apimodel.GeoCaseMapChildrenResponse{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Level:       normalizedLevel,
		ParentCode:  normalizedParentCode,
		ParentName:  parentName,
		RegionCount: len(regions),
		Regions:     regions,
	}
	_ = cache.SetJSON(cacheKey, result, geoCaseMapChildrenCacheTTL)
	return result, nil
}

func buildGeoCaseRegionCases(regionCode string, window string, page int, pageSize int) (apimodel.GeoCaseMapRegionCasesResponse, error) {
	normalizedRegionCode := strings.TrimSpace(regionCode)
	normalizedWindow := normalizeGeoWindow(window)
	normalizedPageSize := normalizeGeoRegionCasesPageSize(pageSize)
	normalizedPage := page
	if normalizedPage <= 0 {
		normalizedPage = 1
	}

	cacheKey := buildGeoCaseRegionCasesCacheKey(normalizedRegionCode, normalizedWindow, normalizedPage, normalizedPageSize)
	var cached apimodel.GeoCaseMapRegionCasesResponse
	if found, err := cache.GetJSON(cacheKey, &cached); err == nil && found {
		return cached, nil
	}

	rows, err := queryGeoCaseRows()
	if err != nil {
		return apimodel.GeoCaseMapRegionCasesResponse{}, err
	}

	now := time.Now()
	items := make([]apimodel.GeoCaseMapCaseSummaryItem, 0)
	regionName := ""
	for _, row := range rows {
		matchedName := matchGeoRegionName(row, normalizedRegionCode)
		if matchedName == "" {
			continue
		}
		if !geoCaseInWindow(row.CreatedAt, normalizedWindow, now) {
			continue
		}
		if regionName == "" {
			regionName = matchedName
		}
		items = append(items, apimodel.GeoCaseMapCaseSummaryItem{
			RecordID:    strings.TrimSpace(row.RecordID),
			Title:       strings.TrimSpace(row.Title),
			CaseSummary: strings.TrimSpace(row.CaseSummary),
			ScamType:    normalizeGeoScamType(row.ScamType),
			RiskLevel:   strings.TrimSpace(row.RiskLevel),
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].RecordID > items[j].RecordID
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})

	total := len(items)
	totalPages := 0
	if total > 0 {
		totalPages = (total + normalizedPageSize - 1) / normalizedPageSize
		if normalizedPage > totalPages {
			normalizedPage = totalPages
		}
	} else {
		normalizedPage = 1
	}

	start := 0
	end := 0
	if total > 0 {
		start = (normalizedPage - 1) * normalizedPageSize
		if start > total {
			start = total
		}
		end = start + normalizedPageSize
		if end > total {
			end = total
		}
	}
	pagedItems := items[start:end]

	result := apimodel.GeoCaseMapRegionCasesResponse{
		GeneratedAt: now.Format(time.RFC3339),
		RegionCode:  normalizedRegionCode,
		RegionName:  regionName,
		Window:      normalizedWindow,
		CaseCount:   total,
		Page:        normalizedPage,
		PageSize:    normalizedPageSize,
		Total:       total,
		TotalPages:  totalPages,
		HasPrev:     totalPages > 0 && normalizedPage > 1,
		HasNext:     totalPages > 0 && normalizedPage < totalPages,
		Cases:       pagedItems,
	}
	_ = cache.SetJSON(cacheKey, result, geoCaseRegionCasesCacheTTL)
	return result, nil
}

func queryGeoCaseRows() ([]geoCaseJoinedRow, error) {
	if database.DB == nil {
		return nil, nil
	}

	rows := make([]geoCaseJoinedRow, 0)
	err := database.DB.Table("history_cases AS hc").
		Select(
			"hc.record_id AS record_id, hc.title AS title, hc.case_summary AS case_summary, hc.risk_level AS risk_level, "+
				"hc.user_id AS user_id, hc.scam_type AS scam_type, hc.created_at AS created_at, "+
				"u.province_code AS province_code, u.province_name AS province_name, "+
				"u.city_code AS city_code, u.city_name AS city_name, "+
				"u.district_code AS district_code, u.district_name AS district_name",
		).
		Joins("JOIN users AS u ON CAST(u.id AS TEXT) = hc.user_id").
		Where("hc.status = ?", "completed").
		Scan(&rows).Error
	return rows, err
}

func buildGeoCaseMapSummary(rows []geoCaseJoinedRow) apimodel.GeoCaseMapSummary {
	userLocations := make(map[string]struct{})
	provinces := make(map[string]struct{})
	cities := make(map[string]struct{})

	for _, row := range rows {
		if strings.TrimSpace(row.ProvinceCode) != "" && strings.TrimSpace(row.UserID) != "" {
			userLocations[strings.TrimSpace(row.UserID)] = struct{}{}
			provinces[strings.TrimSpace(row.ProvinceCode)] = struct{}{}
		}
		if strings.TrimSpace(row.CityCode) != "" {
			cities[strings.TrimSpace(row.CityCode)] = struct{}{}
		}
	}

	return apimodel.GeoCaseMapSummary{
		TotalUsersWithLocation: len(userLocations),
		TotalCases:             len(rows),
		ProvinceCount:          len(provinces),
		CityCount:              len(cities),
	}
}

func aggregateGeoRegions(rows []geoCaseJoinedRow, level string, parentCode string) []apimodel.GeoCaseMapRegionItem {
	normalizedLevel := normalizeGeoLevel(level)
	normalizedParentCode := strings.TrimSpace(parentCode)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last7dStart := todayStart.AddDate(0, 0, -6)
	prev7dStart := last7dStart.AddDate(0, 0, -7)
	last30dStart := todayStart.AddDate(0, 0, -29)
	prev30dStart := last30dStart.AddDate(0, 0, -30)

	aggregates := map[string]*geoRegionAggregate{}
	for _, row := range rows {
		regionCode, regionName, include := pickGeoRegion(row, normalizedLevel, normalizedParentCode)
		if !include {
			continue
		}
		aggregate := aggregates[regionCode]
		if aggregate == nil {
			nextAggregate := newGeoRegionAggregate(regionCode, regionName)
			aggregate = &nextAggregate
			aggregates[regionCode] = aggregate
		}
		applyGeoCaseToAggregate(aggregate, row.CreatedAt, normalizeGeoScamType(row.ScamType), now, todayStart, last7dStart, prev7dStart, last30dStart, prev30dStart)
	}

	items := make([]apimodel.GeoCaseMapRegionItem, 0, len(aggregates))
	distributionToday := make([]int, 0, len(aggregates))
	distribution7d := make([]int, 0, len(aggregates))
	distribution30d := make([]int, 0, len(aggregates))
	distributionAll := make([]int, 0, len(aggregates))
	for _, aggregate := range aggregates {
		distributionToday = append(distributionToday, aggregate.Today.Count)
		distribution7d = append(distribution7d, aggregate.Last7d.Count)
		distribution30d = append(distribution30d, aggregate.Last30d.Count)
		distributionAll = append(distributionAll, aggregate.AllTime.Count)
	}

	for _, aggregate := range aggregates {
		items = append(items, apimodel.GeoCaseMapRegionItem{
			RegionCode: aggregate.RegionCode,
			RegionName: aggregate.RegionName,
			Stats: buildGeoRegionStats(
				*aggregate,
				geoRiskLevel(aggregate.Today.Count, distributionToday),
				geoRiskLevel(aggregate.Last7d.Count, distribution7d),
				geoRiskLevel(aggregate.Last30d.Count, distribution30d),
				geoRiskLevel(aggregate.AllTime.Count, distributionAll),
			),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Stats.AllTime.Count == items[j].Stats.AllTime.Count {
			return items[i].RegionCode < items[j].RegionCode
		}
		return items[i].Stats.AllTime.Count > items[j].Stats.AllTime.Count
	})
	return items
}

func pickGeoRegion(row geoCaseJoinedRow, level string, parentCode string) (string, string, bool) {
	switch normalizeGeoLevel(level) {
	case geoCaseMapLevelCity:
		if strings.TrimSpace(row.ProvinceCode) != parentCode {
			return "", "", false
		}
		code := strings.TrimSpace(row.CityCode)
		name := strings.TrimSpace(row.CityName)
		return code, name, code != "" && name != ""
	case geoCaseMapLevelDistrict:
		if strings.TrimSpace(row.CityCode) != parentCode {
			return "", "", false
		}
		code := strings.TrimSpace(row.DistrictCode)
		name := strings.TrimSpace(row.DistrictName)
		return code, name, code != "" && name != ""
	default:
		code := strings.TrimSpace(row.ProvinceCode)
		name := strings.TrimSpace(row.ProvinceName)
		return code, name, code != "" && name != ""
	}
}

func resolveGeoParentName(rows []geoCaseJoinedRow, level string, parentCode string) string {
	normalizedLevel := normalizeGeoLevel(level)
	normalizedParentCode := strings.TrimSpace(parentCode)
	for _, row := range rows {
		switch normalizedLevel {
		case geoCaseMapLevelCity:
			if strings.TrimSpace(row.ProvinceCode) == normalizedParentCode {
				return strings.TrimSpace(row.ProvinceName)
			}
		case geoCaseMapLevelDistrict:
			if strings.TrimSpace(row.CityCode) == normalizedParentCode {
				return strings.TrimSpace(row.CityName)
			}
		}
	}
	return ""
}

func newGeoRegionAggregate(code string, name string) geoRegionAggregate {
	return geoRegionAggregate{
		RegionCode: code,
		RegionName: name,
		Today:      geoWindowAccumulator{ScamTypeCount: map[string]int{}},
		Last7d:     geoWindowAccumulator{ScamTypeCount: map[string]int{}},
		Last30d:    geoWindowAccumulator{ScamTypeCount: map[string]int{}},
		AllTime:    geoWindowAccumulator{ScamTypeCount: map[string]int{}},
	}
}

func applyGeoCaseToAggregate(aggregate *geoRegionAggregate, createdAt time.Time, scamType string, now, todayStart, last7dStart, prev7dStart, last30dStart, prev30dStart time.Time) {
	if aggregate == nil {
		return
	}

	created := createdAt.In(now.Location())
	aggregate.AllTime.Count++
	aggregate.AllTime.ScamTypeCount[scamType]++
	if !created.Before(todayStart) {
		aggregate.Today.Count++
		aggregate.Today.ScamTypeCount[scamType]++
	} else if created.After(todayStart.AddDate(0, 0, -1).Add(-time.Nanosecond)) {
		aggregate.Today.PreviousCount++
	}
	if !created.Before(last7dStart) {
		aggregate.Last7d.Count++
		aggregate.Last7d.ScamTypeCount[scamType]++
	} else if !created.Before(prev7dStart) {
		aggregate.Last7d.PreviousCount++
	}
	if !created.Before(last30dStart) {
		aggregate.Last30d.Count++
		aggregate.Last30d.ScamTypeCount[scamType]++
	} else if !created.Before(prev30dStart) {
		aggregate.Last30d.PreviousCount++
	}
}

func buildGeoRegionStats(aggregate geoRegionAggregate, todayRisk, sevenRisk, thirtyRisk, allRisk string) apimodel.GeoCaseMapRegionStats {
	return apimodel.GeoCaseMapRegionStats{
		Today:   buildGeoWindowStats(aggregate.Today, todayRisk),
		Last7d:  buildGeoWindowStats(aggregate.Last7d, sevenRisk),
		Last30d: buildGeoWindowStats(aggregate.Last30d, thirtyRisk),
		AllTime: buildGeoWindowStats(aggregate.AllTime, allRisk),
	}
}

func buildGeoWindowStats(acc geoWindowAccumulator, risk string) apimodel.GeoCaseMapWindowStats {
	return apimodel.GeoCaseMapWindowStats{
		Count:         acc.Count,
		PreviousCount: acc.PreviousCount,
		ChangeRate:    geoChangeRate(acc.Count, acc.PreviousCount),
		Trend:         geoTrend(acc.Count, acc.PreviousCount),
		RiskLevel:     risk,
		TopScamTypes:  geoTopScamTypes(acc.ScamTypeCount, 3),
	}
}

func geoTopScamTypes(counter map[string]int, limit int) []apimodel.GeoCaseMapTopScamType {
	items := make([]apimodel.GeoCaseMapTopScamType, 0, len(counter))
	for scamType, count := range counter {
		items = append(items, apimodel.GeoCaseMapTopScamType{ScamType: scamType, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].ScamType < items[j].ScamType
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

func geoChangeRate(current, previous int) float64 {
	if previous <= 0 {
		if current <= 0 {
			return 0
		}
		return 1
	}
	return float64(current-previous) / float64(previous)
}

func geoTrend(current, previous int) string {
	switch {
	case current > previous:
		return "上升"
	case current < previous:
		return "下降"
	default:
		return "持平"
	}
}

func geoRiskLevel(count int, distribution []int) string {
	if count <= 0 {
		return "低"
	}
	positive := make([]int, 0, len(distribution))
	for _, item := range distribution {
		if item > 0 {
			positive = append(positive, item)
		}
	}
	if len(positive) == 0 {
		return "低"
	}
	sort.Ints(positive)
	lowThreshold := positive[(len(positive)-1)/3]
	highThreshold := positive[((len(positive)-1)*2)/3]
	switch {
	case count >= highThreshold && highThreshold > 0:
		return "高"
	case count >= lowThreshold:
		return "中"
	default:
		return "低"
	}
}

func normalizeGeoScamType(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "未知类型"
	}
	return trimmed
}

func normalizeGeoLevel(raw string) string {
	switch strings.TrimSpace(raw) {
	case geoCaseMapLevelCity:
		return geoCaseMapLevelCity
	case geoCaseMapLevelDistrict:
		return geoCaseMapLevelDistrict
	default:
		return geoCaseMapLevelProvince
	}
}

func normalizeGeoWindow(raw string) string {
	switch strings.TrimSpace(raw) {
	case "today":
		return "today"
	case "last_30d":
		return "last_30d"
	case "all_time":
		return "all_time"
	default:
		return "last_7d"
	}
}

func normalizeGeoRegionCasesPageSize(raw int) int {
	switch {
	case raw <= 10:
		return 10
	case raw <= 20:
		return 20
	case raw <= 50:
		return 50
	default:
		return 50
	}
}

func matchGeoRegionName(row geoCaseJoinedRow, regionCode string) string {
	normalizedCode := strings.TrimSpace(regionCode)
	if normalizedCode == "" {
		return ""
	}
	if strings.TrimSpace(row.DistrictCode) == normalizedCode {
		return strings.TrimSpace(row.DistrictName)
	}
	if strings.TrimSpace(row.CityCode) == normalizedCode {
		return strings.TrimSpace(row.CityName)
	}
	if strings.TrimSpace(row.ProvinceCode) == normalizedCode {
		return strings.TrimSpace(row.ProvinceName)
	}
	return ""
}

func geoCaseInWindow(createdAt time.Time, window string, now time.Time) bool {
	created := createdAt.In(now.Location())
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	switch normalizeGeoWindow(window) {
	case "today":
		return !created.Before(todayStart)
	case "last_30d":
		return !created.Before(todayStart.AddDate(0, 0, -29))
	case "all_time":
		return true
	default:
		return !created.Before(todayStart.AddDate(0, 0, -6))
	}
}

func GeoCaseMapCacheVersion() string {
	var version string
	found, err := cache.GetJSON(geoCaseMapCacheVersionKey, &version)
	if err != nil || !found || strings.TrimSpace(version) == "" {
		return "0"
	}
	return strings.TrimSpace(version)
}

func TouchGeoCaseMapCacheVersion() {
	_ = cache.SetJSON(geoCaseMapCacheVersionKey, fmt.Sprintf("%d", time.Now().UnixNano()), 0)
}

func buildGeoCaseMapOverviewCacheKey() string {
	return geoCaseMapOverviewCacheKeyPrefix + GeoCaseMapCacheVersion()
}

func buildGeoCaseMapChildrenCacheKey(level string, parentCode string) string {
	return geoCaseMapChildrenCacheKeyPrefix + normalizeGeoLevel(level) + ":" + strings.TrimSpace(parentCode) + ":" + GeoCaseMapCacheVersion()
}

func buildGeoCaseRegionCasesCacheKey(regionCode string, window string, page int, pageSize int) string {
	return fmt.Sprintf("%s%s:%s:%d:%d:%s",
		geoCaseRegionCasesCacheKeyPrefix,
		strings.TrimSpace(regionCode),
		normalizeGeoWindow(window),
		page,
		pageSize,
		GeoCaseMapCacheVersion(),
	)
}
