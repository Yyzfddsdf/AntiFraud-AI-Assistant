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
	geoCaseMapCacheVersionKey = "cache:case_library:geo_map:v1:version"
	geoCaseMapCacheKeyPrefix  = "cache:case_library:geo_map:v1:data:"
	geoCaseMapCacheTTL        = 2 * time.Minute
)

type geoCaseJoinedRow struct {
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

type geoProvinceAggregate struct {
	geoRegionAggregate
	Cities map[string]*geoRegionAggregate
	CityDistricts map[string]map[string]*geoRegionAggregate
}

func buildGeoCaseMap() (apimodel.GeoCaseMapResponse, error) {
	cacheKey := buildGeoCaseMapCacheKey()
	var cached apimodel.GeoCaseMapResponse
	if found, err := cache.GetJSON(cacheKey, &cached); err == nil && found {
		return cached, nil
	}

	rows, err := queryGeoCaseRows()
	if err != nil {
		return apimodel.GeoCaseMapResponse{}, err
	}
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	last7dStart := todayStart.AddDate(0, 0, -6)
	prev7dStart := last7dStart.AddDate(0, 0, -7)
	last30dStart := todayStart.AddDate(0, 0, -29)
	prev30dStart := last30dStart.AddDate(0, 0, -30)

	provinces := make(map[string]*geoProvinceAggregate)
	userLocations := make(map[string]struct{})
	for _, row := range rows {
		provinceCode := strings.TrimSpace(row.ProvinceCode)
		provinceName := strings.TrimSpace(row.ProvinceName)
		cityCode := strings.TrimSpace(row.CityCode)
		cityName := strings.TrimSpace(row.CityName)
		scamType := normalizeGeoScamType(row.ScamType)
		if provinceCode == "" || provinceName == "" {
			continue
		}
		userLocations[strings.TrimSpace(row.UserID)] = struct{}{}
		province := provinces[provinceCode]
		if province == nil {
			province = &geoProvinceAggregate{
				geoRegionAggregate: newGeoRegionAggregate(provinceCode, provinceName),
				Cities:             map[string]*geoRegionAggregate{},
				CityDistricts:      map[string]map[string]*geoRegionAggregate{},
			}
			provinces[provinceCode] = province
		}
		applyGeoCaseToAggregate(&province.geoRegionAggregate, row.CreatedAt, scamType, now, todayStart, last7dStart, prev7dStart, last30dStart, prev30dStart)

		if cityCode == "" || cityName == "" {
			continue
		}
		city := province.Cities[cityCode]
		if city == nil {
			aggregate := newGeoRegionAggregate(cityCode, cityName)
			city = &aggregate
			province.Cities[cityCode] = city
		}
		applyGeoCaseToAggregate(city, row.CreatedAt, scamType, now, todayStart, last7dStart, prev7dStart, last30dStart, prev30dStart)

		districtCode := strings.TrimSpace(row.DistrictCode)
		districtName := strings.TrimSpace(row.DistrictName)
		if districtCode == "" || districtName == "" {
			continue
		}
		if province.CityDistricts[cityCode] == nil {
			province.CityDistricts[cityCode] = map[string]*geoRegionAggregate{}
		}
		district := province.CityDistricts[cityCode][districtCode]
		if district == nil {
			aggregate := newGeoRegionAggregate(districtCode, districtName)
			district = &aggregate
			province.CityDistricts[cityCode][districtCode] = district
		}
		applyGeoCaseToAggregate(district, row.CreatedAt, scamType, now, todayStart, last7dStart, prev7dStart, last30dStart, prev30dStart)
	}

	provinceItems := make([]apimodel.GeoCaseMapProvinceItem, 0, len(provinces))
	allProvinceToday := make([]int, 0, len(provinces))
	allProvince7d := make([]int, 0, len(provinces))
	allProvince30d := make([]int, 0, len(provinces))
	allProvinceAll := make([]int, 0, len(provinces))
	cityCount := 0

	for _, province := range provinces {
		cityItems := make([]apimodel.GeoCaseMapCityItem, 0, len(province.Cities))
		cityToday := make([]int, 0, len(province.Cities))
		city7d := make([]int, 0, len(province.Cities))
		city30d := make([]int, 0, len(province.Cities))
		cityAll := make([]int, 0, len(province.Cities))
		for _, city := range province.Cities {
			cityToday = append(cityToday, city.Today.Count)
			city7d = append(city7d, city.Last7d.Count)
			city30d = append(city30d, city.Last30d.Count)
			cityAll = append(cityAll, city.AllTime.Count)
		}
		for cityCode, city := range province.Cities {
			districtAggregates := province.CityDistricts[cityCode]
			districtItems := make([]apimodel.GeoCaseMapDistrictItem, 0, len(districtAggregates))
			districtToday := make([]int, 0, len(districtAggregates))
			district7d := make([]int, 0, len(districtAggregates))
			district30d := make([]int, 0, len(districtAggregates))
			districtAll := make([]int, 0, len(districtAggregates))
			for _, district := range districtAggregates {
				districtToday = append(districtToday, district.Today.Count)
				district7d = append(district7d, district.Last7d.Count)
				district30d = append(district30d, district.Last30d.Count)
				districtAll = append(districtAll, district.AllTime.Count)
			}
			for _, district := range districtAggregates {
				districtItems = append(districtItems, apimodel.GeoCaseMapDistrictItem{
					RegionCode: district.RegionCode,
					RegionName: district.RegionName,
					Stats: buildGeoRegionStats(*district,
						geoRiskLevel(district.Today.Count, districtToday),
						geoRiskLevel(district.Last7d.Count, district7d),
						geoRiskLevel(district.Last30d.Count, district30d),
						geoRiskLevel(district.AllTime.Count, districtAll),
					),
				})
			}
			sort.Slice(districtItems, func(i, j int) bool {
				if districtItems[i].Stats.AllTime.Count == districtItems[j].Stats.AllTime.Count {
					return districtItems[i].RegionCode < districtItems[j].RegionCode
				}
				return districtItems[i].Stats.AllTime.Count > districtItems[j].Stats.AllTime.Count
			})
			cityItems = append(cityItems, apimodel.GeoCaseMapCityItem{
				RegionCode: city.RegionCode,
				RegionName: city.RegionName,
				Stats: buildGeoRegionStats(*city,
					geoRiskLevel(city.Today.Count, cityToday),
					geoRiskLevel(city.Last7d.Count, city7d),
					geoRiskLevel(city.Last30d.Count, city30d),
					geoRiskLevel(city.AllTime.Count, cityAll),
				),
				Districts: districtItems,
			})
		}
		sort.Slice(cityItems, func(i, j int) bool {
			if cityItems[i].Stats.AllTime.Count == cityItems[j].Stats.AllTime.Count {
				return cityItems[i].RegionCode < cityItems[j].RegionCode
			}
			return cityItems[i].Stats.AllTime.Count > cityItems[j].Stats.AllTime.Count
		})

		allProvinceToday = append(allProvinceToday, province.Today.Count)
		allProvince7d = append(allProvince7d, province.Last7d.Count)
		allProvince30d = append(allProvince30d, province.Last30d.Count)
		allProvinceAll = append(allProvinceAll, province.AllTime.Count)
		cityCount += len(cityItems)
		provinceItems = append(provinceItems, apimodel.GeoCaseMapProvinceItem{
			RegionCode: province.RegionCode,
			RegionName: province.RegionName,
			Stats: buildGeoRegionStats(province.geoRegionAggregate,
				geoRiskLevel(province.Today.Count, allProvinceToday),
				geoRiskLevel(province.Last7d.Count, allProvince7d),
				geoRiskLevel(province.Last30d.Count, allProvince30d),
				geoRiskLevel(province.AllTime.Count, allProvinceAll),
			),
			Cities: cityItems,
		})
	}

	// Recompute province risk levels after the full distribution is known.
	for index := range provinceItems {
		provinceItems[index].Stats = buildGeoRegionStats(
			provinces[provinceItems[index].RegionCode].geoRegionAggregate,
			geoRiskLevel(provinceItems[index].Stats.Today.Count, allProvinceToday),
			geoRiskLevel(provinceItems[index].Stats.Last7d.Count, allProvince7d),
			geoRiskLevel(provinceItems[index].Stats.Last30d.Count, allProvince30d),
			geoRiskLevel(provinceItems[index].Stats.AllTime.Count, allProvinceAll),
		)
	}

	sort.Slice(provinceItems, func(i, j int) bool {
		if provinceItems[i].Stats.AllTime.Count == provinceItems[j].Stats.AllTime.Count {
			return provinceItems[i].RegionCode < provinceItems[j].RegionCode
		}
		return provinceItems[i].Stats.AllTime.Count > provinceItems[j].Stats.AllTime.Count
	})

	result := apimodel.GeoCaseMapResponse{
		GeneratedAt: now.Format(time.RFC3339),
		Summary: apimodel.GeoCaseMapSummary{
			TotalUsersWithLocation: len(userLocations),
			TotalCases:             len(rows),
			ProvinceCount:          len(provinceItems),
			CityCount:              cityCount,
		},
		Provinces: provinceItems,
	}
	_ = cache.SetJSON(cacheKey, result, geoCaseMapCacheTTL)
	return result, nil
}

func queryGeoCaseRows() ([]geoCaseJoinedRow, error) {
	if database.DB == nil {
		return nil, nil
	}
	rows := make([]geoCaseJoinedRow, 0)
	err := database.DB.Table("history_cases AS hc").
		Select(
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

func buildGeoCaseMapCacheKey() string {
	return geoCaseMapCacheKeyPrefix + GeoCaseMapCacheVersion()
}
