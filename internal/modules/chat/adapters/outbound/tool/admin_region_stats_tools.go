package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	multihttp "antifraud/internal/modules/multi_agent/adapters/inbound/http"
	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	region_system "antifraud/internal/modules/region"
	openai "antifraud/internal/platform/llm"
)

const (
	AdminQueryRegionTopScamTypesToolName = "admin_query_region_top_scam_types"
	AdminQueryRegionCaseRankingToolName  = "admin_query_region_case_ranking"
)

type AdminQueryRegionTopScamTypesInput struct {
	Level        string `json:"level"`
	ProvinceName string `json:"province_name,omitempty"`
	CityName     string `json:"city_name,omitempty"`
	DistrictName string `json:"district_name,omitempty"`
	Window       string `json:"window,omitempty"`
}

type AdminQueryRegionCaseRankingInput struct {
	Scope        string `json:"scope"`
	ProvinceName string `json:"province_name,omitempty"`
	CityName     string `json:"city_name,omitempty"`
	Window       string `json:"window,omitempty"`
	TopK         int    `json:"top_k,omitempty"`
}

var AdminQueryRegionTopScamTypesTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        AdminQueryRegionTopScamTypesToolName,
		Description: "查询某个省、市或区县在指定时间窗口下的案件Top3诈骗类型、案件数和风险等级。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"level": map[string]interface{}{
					"type":        "string",
					"description": "地区层级，可选 province、city、district。",
					"enum":        []string{"province", "city", "district"},
				},
				"province_name": map[string]interface{}{
					"type":        "string",
					"description": "省份名称。查询省级时必填；查询市级或区县时建议填写以便精确定位。",
				},
				"city_name": map[string]interface{}{
					"type":        "string",
					"description": "城市名称。查询市级时必填；查询区县时必填。",
				},
				"district_name": map[string]interface{}{
					"type":        "string",
					"description": "区县名称。查询区县时必填。",
				},
				"window": map[string]interface{}{
					"type":        "string",
					"description": "时间窗口，可选 today、last_7d、last_30d、all_time，默认 last_7d。",
					"enum":        []string{"today", "last_7d", "last_30d", "all_time"},
				},
			},
			"required": []string{"level"},
		},
	},
}

var AdminQueryRegionCaseRankingTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        AdminQueryRegionCaseRankingToolName,
		Description: "查询全国、某省或某市范围内的案件数量TopK地区排行。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"scope": map[string]interface{}{
					"type":        "string",
					"description": "排行范围，可选 country、province、city。country返回省级排行，province返回某省下城市排行，city返回某市下区县排行。",
					"enum":        []string{"country", "province", "city"},
				},
				"province_name": map[string]interface{}{
					"type":        "string",
					"description": "当 scope=province 或 scope=city 时建议填写；scope=province 时表示该省下的城市排行。",
				},
				"city_name": map[string]interface{}{
					"type":        "string",
					"description": "当 scope=city 时必填，表示该市下的区县排行。",
				},
				"window": map[string]interface{}{
					"type":        "string",
					"description": "时间窗口，可选 today、last_7d、last_30d、all_time，默认 last_7d。",
					"enum":        []string{"today", "last_7d", "last_30d", "all_time"},
				},
				"top_k": map[string]interface{}{
					"type":        "integer",
					"description": "返回前K个地区，默认5，最大20。",
				},
			},
			"required": []string{"scope"},
		},
	},
}

type AdminQueryRegionTopScamTypesHandler struct{}

func (h *AdminQueryRegionTopScamTypesHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	_ = ctx
	_ = userID

	input, err := parseAdminQueryRegionTopScamTypesInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid admin region top scam input: %v", err)}}, nil
	}

	regionItem, resolved, resolveErr := resolveAdminRegionStats(input.Level, input.ProvinceName, input.CityName, input.DistrictName)
	if resolveErr != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": resolveErr.Error()}}, nil
	}

	stats := pickGeoWindowStats(regionItem.Stats, normalizeAdminGeoWindow(input.Window))
	return ChatToolResponse{Payload: map[string]interface{}{
		"level":          resolved.Level,
		"region_code":    regionItem.RegionCode,
		"region_name":    regionItem.RegionName,
		"window":         normalizeAdminGeoWindow(input.Window),
		"case_count":     stats.Count,
		"risk_level":     stats.RiskLevel,
		"trend":          stats.Trend,
		"change_rate":    stats.ChangeRate,
		"top_scam_types": stats.TopScamTypes,
	}}, nil
}

type AdminQueryRegionCaseRankingHandler struct{}

func (h *AdminQueryRegionCaseRankingHandler) Handle(ctx context.Context, userID string, args string) (ChatToolResponse, error) {
	_ = ctx
	_ = userID

	input, err := parseAdminQueryRegionCaseRankingInput(args)
	if err != nil {
		return ChatToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid admin region ranking input: %v", err)}}, nil
	}

	scope := normalizeAdminScope(input.Scope)
	window := normalizeAdminGeoWindow(input.Window)
	topK := normalizeAdminTopK(input.TopK)

	var regions []apimodel.GeoCaseMapRegionItem
	var parentName string

	switch scope {
	case "country":
		overview, overviewErr := multihttp.BuildGeoCaseMapOverview()
		if overviewErr != nil {
			return ChatToolResponse{Payload: map[string]interface{}{"error": overviewErr.Error()}}, nil
		}
		regions = overview.Regions
		parentName = "全国"
	case "province":
		provinceCode, provinceName, provinceErr := resolveProvinceCodeByName(input.ProvinceName)
		if provinceErr != nil {
			return ChatToolResponse{Payload: map[string]interface{}{"error": provinceErr.Error()}}, nil
		}
		children, childrenErr := multihttp.BuildGeoCaseMapChildren(provinceCode, "city")
		if childrenErr != nil {
			return ChatToolResponse{Payload: map[string]interface{}{"error": childrenErr.Error()}}, nil
		}
		regions = children.Regions
		parentName = provinceName
	case "city":
		cityCode, cityName, cityErr := resolveCityCodeByName(input.ProvinceName, input.CityName)
		if cityErr != nil {
			return ChatToolResponse{Payload: map[string]interface{}{"error": cityErr.Error()}}, nil
		}
		children, childrenErr := multihttp.BuildGeoCaseMapChildren(cityCode, "district")
		if childrenErr != nil {
			return ChatToolResponse{Payload: map[string]interface{}{"error": childrenErr.Error()}}, nil
		}
		regions = children.Regions
		parentName = cityName
	}

	ranked := rankAdminRegionsByWindow(regions, window, topK)
	return ChatToolResponse{Payload: map[string]interface{}{
		"scope":       scope,
		"window":      window,
		"top_k":       topK,
		"parent_name": parentName,
		"regions":     ranked,
	}}, nil
}

type adminResolvedRegion struct {
	Level string
}

func parseAdminQueryRegionTopScamTypesInput(args string) (AdminQueryRegionTopScamTypesInput, error) {
	if strings.TrimSpace(args) == "" {
		return AdminQueryRegionTopScamTypesInput{}, nil
	}
	var input AdminQueryRegionTopScamTypesInput
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return AdminQueryRegionTopScamTypesInput{}, err
	}
	return input, nil
}

func parseAdminQueryRegionCaseRankingInput(args string) (AdminQueryRegionCaseRankingInput, error) {
	if strings.TrimSpace(args) == "" {
		return AdminQueryRegionCaseRankingInput{}, nil
	}
	var input AdminQueryRegionCaseRankingInput
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return AdminQueryRegionCaseRankingInput{}, err
	}
	return input, nil
}

func resolveAdminRegionStats(level string, provinceName string, cityName string, districtName string) (apimodel.GeoCaseMapRegionItem, adminResolvedRegion, error) {
	switch normalizeAdminLevel(level) {
	case "province":
		provinceCode, _, err := resolveProvinceCodeByName(provinceName)
		if err != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, err
		}
		overview, overviewErr := multihttp.BuildGeoCaseMapOverview()
		if overviewErr != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, overviewErr
		}
		for _, item := range overview.Regions {
			if item.RegionCode == provinceCode {
				return item, adminResolvedRegion{Level: "province"}, nil
			}
		}
		return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, fmt.Errorf("未找到省份 %s 的案件统计", strings.TrimSpace(provinceName))
	case "city":
		provinceCode, _, provinceErr := resolveProvinceCodeByName(provinceName)
		if provinceErr != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, provinceErr
		}
		children, childrenErr := multihttp.BuildGeoCaseMapChildren(provinceCode, "city")
		if childrenErr != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, childrenErr
		}
		target := normalizeRegionKeyword(cityName)
		for _, item := range children.Regions {
			if normalizeRegionKeyword(item.RegionName) == target || strings.Contains(normalizeRegionKeyword(item.RegionName), target) {
				return item, adminResolvedRegion{Level: "city"}, nil
			}
		}
		return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, fmt.Errorf("未找到城市 %s 的案件统计", strings.TrimSpace(cityName))
	default:
		cityCode, _, cityErr := resolveCityCodeByName(provinceName, cityName)
		if cityErr != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, cityErr
		}
		children, childrenErr := multihttp.BuildGeoCaseMapChildren(cityCode, "district")
		if childrenErr != nil {
			return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, childrenErr
		}
		target := normalizeRegionKeyword(districtName)
		for _, item := range children.Regions {
			if normalizeRegionKeyword(item.RegionName) == target || strings.Contains(normalizeRegionKeyword(item.RegionName), target) {
				return item, adminResolvedRegion{Level: "district"}, nil
			}
		}
		return apimodel.GeoCaseMapRegionItem{}, adminResolvedRegion{}, fmt.Errorf("未找到区县 %s 的案件统计", strings.TrimSpace(districtName))
	}
}

func resolveProvinceCodeByName(provinceName string) (string, string, error) {
	service := region_system.NewService()
	target := normalizeRegionKeyword(provinceName)
	if target == "" {
		return "", "", fmt.Errorf("province_name 不能为空")
	}
	for _, item := range service.ListProvinces() {
		if normalizeRegionKeyword(item.Name) == target || strings.Contains(normalizeRegionKeyword(item.Name), target) {
			return item.Code, item.Name, nil
		}
	}
	return "", "", fmt.Errorf("未找到省份 %s", strings.TrimSpace(provinceName))
}

func resolveCityCodeByName(provinceName string, cityName string) (string, string, error) {
	service := region_system.NewService()
	targetCity := normalizeRegionKeyword(cityName)
	if targetCity == "" {
		return "", "", fmt.Errorf("city_name 不能为空")
	}

	if strings.TrimSpace(provinceName) != "" {
		provinceCode, _, err := resolveProvinceCodeByName(provinceName)
		if err != nil {
			return "", "", err
		}
		cities, err := service.ListCities(provinceCode)
		if err != nil {
			return "", "", err
		}
		for _, item := range cities {
			if normalizeRegionKeyword(item.Name) == targetCity || strings.Contains(normalizeRegionKeyword(item.Name), targetCity) {
				return item.Code, item.Name, nil
			}
		}
		return "", "", fmt.Errorf("未在省份 %s 下找到城市 %s", strings.TrimSpace(provinceName), strings.TrimSpace(cityName))
	}

	for _, province := range service.ListProvinces() {
		cities, err := service.ListCities(province.Code)
		if err != nil {
			continue
		}
		for _, item := range cities {
			if normalizeRegionKeyword(item.Name) == targetCity || strings.Contains(normalizeRegionKeyword(item.Name), targetCity) {
				return item.Code, item.Name, nil
			}
		}
	}
	return "", "", fmt.Errorf("未找到城市 %s", strings.TrimSpace(cityName))
}

func normalizeAdminLevel(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "province":
		return "province"
	case "city":
		return "city"
	case "district":
		return "district"
	default:
		return "district"
	}
}

func normalizeAdminScope(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "province":
		return "province"
	case "city":
		return "city"
	default:
		return "country"
	}
}

func normalizeAdminGeoWindow(raw string) string {
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

func normalizeAdminTopK(raw int) int {
	switch {
	case raw <= 0:
		return 5
	case raw > 20:
		return 20
	default:
		return raw
	}
}

func pickGeoWindowStats(stats apimodel.GeoCaseMapRegionStats, window string) apimodel.GeoCaseMapWindowStats {
	switch normalizeAdminGeoWindow(window) {
	case "today":
		return stats.Today
	case "last_30d":
		return stats.Last30d
	case "all_time":
		return stats.AllTime
	default:
		return stats.Last7d
	}
}

func rankAdminRegionsByWindow(regions []apimodel.GeoCaseMapRegionItem, window string, topK int) []map[string]interface{} {
	fieldWindow := normalizeAdminGeoWindow(window)
	items := append([]apimodel.GeoCaseMapRegionItem{}, regions...)
	sort.Slice(items, func(i, j int) bool {
		left := pickGeoWindowStats(items[i].Stats, fieldWindow)
		right := pickGeoWindowStats(items[j].Stats, fieldWindow)
		if left.Count == right.Count {
			return items[i].RegionCode < items[j].RegionCode
		}
		return left.Count > right.Count
	})
	if len(items) > topK {
		items = items[:topK]
	}

	result := make([]map[string]interface{}, 0, len(items))
	for index, item := range items {
		stats := pickGeoWindowStats(item.Stats, fieldWindow)
		result = append(result, map[string]interface{}{
			"rank":           index + 1,
			"region_code":    item.RegionCode,
			"region_name":    item.RegionName,
			"case_count":     stats.Count,
			"risk_level":     stats.RiskLevel,
			"trend":          stats.Trend,
			"change_rate":    stats.ChangeRate,
			"top_scam_types": stats.TopScamTypes,
		})
	}
	return result
}

func normalizeRegionKeyword(raw string) string {
	replacer := strings.NewReplacer(
		" ", "",
		"省", "",
		"市", "",
		"区", "",
		"县", "",
		"盟", "",
		"自治州", "",
		"地区", "",
		"特别行政区", "",
		"壮族自治区", "",
		"回族自治区", "",
		"维吾尔自治区", "",
		"自治区", "",
	)
	return replacer.Replace(strings.TrimSpace(raw))
}
