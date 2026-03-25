package region_system

import (
	"fmt"
	"sort"
	"strings"

	gb2260 "github.com/cn/GB2260.go"
)

type RegionOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type RegionSelection struct {
	ProvinceCode string `json:"province_code"`
	ProvinceName string `json:"province_name"`
	CityCode     string `json:"city_code"`
	CityName     string `json:"city_name"`
	DistrictCode string `json:"district_code"`
	DistrictName string `json:"district_name"`
}

type ResolveRegionInput struct {
	ProvinceName       string   `json:"province_name"`
	CityName           string   `json:"city_name"`
	DistrictName       string   `json:"district_name"`
	DistrictCandidates []string `json:"district_candidates"`
}

type Service struct {
	gb gb2260.GB2260
}

var specialRegionProvinceNames = map[string]string{
	"710000": "台湾省",
	"810000": "香港特别行政区",
	"820000": "澳门特别行政区",
}

var supplementalDistrictsByCity = map[string][]RegionOption{
	"330100": {
		{Code: "330113", Name: "临平区"},
		{Code: "330114", Name: "钱塘区"},
	},
}

var deprecatedDistrictCodesByCity = map[string]map[string]struct{}{
	"330100": {
		"330103": {},
		"330104": {},
	},
}

func NewService() *Service {
	return &Service{gb: gb2260.NewGB2260("")}
}

func (s *Service) ListProvinces() []RegionOption {
	if s == nil {
		return []RegionOption{}
	}
	options := make([]RegionOption, 0)
	for code, name := range s.gb.Store {
		if strings.HasSuffix(code, "0000") {
			options = append(options, RegionOption{Code: code, Name: name})
		}
	}
	sortOptions(options)
	return options
}

func (s *Service) ListCities(provinceCode string) ([]RegionOption, error) {
	if s == nil {
		return nil, fmt.Errorf("地区服务当前不可用")
	}
	provinceCode = strings.TrimSpace(provinceCode)
	if provinceCode == "" {
		return nil, fmt.Errorf("缺少省级行政区编码")
	}
	province := s.gb.Get(provinceCode)
	if province == nil || !province.IsProvince() {
		return nil, fmt.Errorf("省级行政区编码无效")
	}
	options := make([]RegionOption, 0)
	provincePrefix := provinceCode[:2]
	for code, name := range s.gb.Store {
		if !strings.HasPrefix(code, provincePrefix) || code == provinceCode {
			continue
		}
		if strings.HasSuffix(code, "00") && !strings.HasSuffix(code, "0000") {
			options = append(options, RegionOption{Code: code, Name: name})
		}
	}
	if len(options) == 0 {
		options = append(options, RegionOption{Code: province.Code, Name: province.Name})
	}
	sortOptions(options)
	return options, nil
}

func (s *Service) ListDistricts(cityCode string) ([]RegionOption, error) {
	if s == nil {
		return nil, fmt.Errorf("地区服务当前不可用")
	}
	cityCode = strings.TrimSpace(cityCode)
	if cityCode == "" {
		return nil, fmt.Errorf("缺少中间层级行政区编码")
	}
	if province := s.gb.Get(cityCode); province != nil && province.IsProvince() {
		return s.withSupplementalDistricts(province.Code, s.listDistrictsUnderProvince(province.Code)), nil
	}
	city := s.gb.Get(cityCode)
	if city == nil {
		if supplemental, ok := supplementalDistrictsByCity[cityCode]; ok {
			return append([]RegionOption{}, supplemental...), nil
		}
		return nil, fmt.Errorf("中间层级行政区编码无效")
	}
	options := make([]RegionOption, 0)
	prefix := cityCode[:4]
	for code, name := range s.gb.Store {
		if strings.HasPrefix(code, prefix) && !strings.HasSuffix(code, "00") {
			options = append(options, RegionOption{Code: code, Name: name})
		}
	}
	sortOptions(options)
	return s.withSupplementalDistricts(cityCode, options), nil
}

func (s *Service) ResolveByNames(input ResolveRegionInput) (RegionSelection, error) {
	if s == nil {
		return RegionSelection{}, fmt.Errorf("地区服务当前不可用")
	}
	provinceName := normalizeRegionName(input.ProvinceName)
	cityName := normalizeRegionName(input.CityName)
	districtCandidates := normalizeDistrictCandidates(input.DistrictName, input.DistrictCandidates)
	if len(districtCandidates) == 0 {
		return RegionSelection{}, fmt.Errorf("缺少末级行政区名称")
	}

	var fallback RegionSelection
	provinces := s.ListProvinces()
	for _, province := range provinces {
		if provinceName != "" && normalizeRegionName(province.Name) != provinceName {
			continue
		}
		cities, _ := s.ListCities(province.Code)
		for _, city := range cities {
			if cityName != "" && normalizeRegionName(city.Name) != cityName {
				// allow fallback for direct-admin cities later
			}
			districts, _ := s.ListDistricts(city.Code)
			for _, district := range districts {
				if !matchesAnyDistrictCandidate(district.Name, districtCandidates) {
					continue
				}
				selection := RegionSelection{
					ProvinceCode: province.Code,
					ProvinceName: province.Name,
					CityCode:     city.Code,
					CityName:     city.Name,
					DistrictCode: district.Code,
					DistrictName: district.Name,
				}
				if cityName == "" || normalizeRegionName(city.Name) == cityName {
					return selection, nil
				}
				if fallback.DistrictCode == "" {
					fallback = selection
				}
			}
		}
	}
	if fallback.DistrictCode != "" {
		return fallback, nil
	}
	return RegionSelection{}, fmt.Errorf("无法根据当前位置解析出标准行政区")
}

func NormalizeSelection(selection RegionSelection) (RegionSelection, error) {
	service := NewService()
	if service == nil {
		return RegionSelection{}, fmt.Errorf("地区服务当前不可用")
	}

	trimmedProvinceCode := strings.TrimSpace(selection.ProvinceCode)
	trimmedProvinceName := strings.TrimSpace(selection.ProvinceName)
	trimmedCityCode := strings.TrimSpace(selection.CityCode)
	trimmedCityName := strings.TrimSpace(selection.CityName)
	trimmedDistrictCode := strings.TrimSpace(selection.DistrictCode)
	trimmedDistrictName := strings.TrimSpace(selection.DistrictName)

	if expectedProvinceName, isSpecial := specialRegionProvinceNames[trimmedProvinceCode]; isSpecial && trimmedDistrictCode == "" {
		if trimmedProvinceName != "" && trimmedProvinceName != expectedProvinceName {
			return RegionSelection{}, fmt.Errorf("省级行政区名称与编码不匹配")
		}
		return RegionSelection{
			ProvinceCode: trimmedProvinceCode,
			ProvinceName: expectedProvinceName,
			CityCode:     trimmedCityCode,
			CityName:     trimmedCityName,
			DistrictCode: trimmedDistrictCode,
			DistrictName: trimmedDistrictName,
		}, nil
	}

	if selection, ok := supplementalDistrictSelection(trimmedDistrictCode); ok {
		if trimmedProvinceCode != selection.ProvinceCode && trimmedProvinceCode != "" {
			return RegionSelection{}, fmt.Errorf("省级行政区编码无效")
		}
		return validateSupplementalSelection(selection, RegionSelection{
			ProvinceCode: trimmedProvinceCode,
			ProvinceName: trimmedProvinceName,
			CityCode:     trimmedCityCode,
			CityName:     trimmedCityName,
			DistrictCode: trimmedDistrictCode,
			DistrictName: trimmedDistrictName,
		})
	}

	province := service.gb.Get(trimmedProvinceCode)
	district := service.gb.Get(trimmedDistrictCode)
	if province == nil || district == nil {
		return RegionSelection{}, fmt.Errorf("省级、中间层级和末级行政区编码不能为空")
	}
	if !district.IsCountry() {
		return RegionSelection{}, fmt.Errorf("末级行政区编码必须指向可选的最后一级行政区")
	}
	p := district.Province()
	if p == nil {
		return RegionSelection{}, fmt.Errorf("末级行政区编码无效")
	}
	var cityCode string
	var cityName string
	if prefecture := district.Prefecture(); prefecture != nil {
		cityCode = prefecture.Code
		cityName = prefecture.Name
	} else {
		cityCode = p.Code
		cityName = p.Name
	}
	if p.Code != province.Code || cityCode != trimmedCityCode {
		return RegionSelection{}, fmt.Errorf("省、市和末级行政区选择不一致")
	}
	if trimmedProvinceName != "" && trimmedProvinceName != p.Name {
		return RegionSelection{}, fmt.Errorf("省级行政区名称与编码不匹配")
	}
	if trimmedCityName != "" && trimmedCityName != cityName {
		return RegionSelection{}, fmt.Errorf("中间层级行政区名称与编码不匹配")
	}
	if trimmedDistrictName != "" && trimmedDistrictName != district.Name {
		return RegionSelection{}, fmt.Errorf("末级行政区名称与编码不匹配")
	}
	return RegionSelection{
		ProvinceCode: p.Code,
		ProvinceName: p.Name,
		CityCode:     cityCode,
		CityName:     cityName,
		DistrictCode: district.Code,
		DistrictName: district.Name,
	}, nil
}

func (s *Service) listDistrictsUnderProvince(provinceCode string) []RegionOption {
	options := make([]RegionOption, 0)
	provincePrefix := provinceCode[:2]
	for code, name := range s.gb.Store {
		if !strings.HasPrefix(code, provincePrefix) || code == provinceCode || strings.HasSuffix(code, "00") {
			continue
		}
		prefectureCode := code[:4] + "00"
		if _, exists := s.gb.Store[prefectureCode]; exists {
			continue
		}
		options = append(options, RegionOption{Code: code, Name: name})
	}
	sortOptions(options)
	return options
}

func sortOptions(options []RegionOption) {
	sort.Slice(options, func(i, j int) bool {
		return options[i].Code < options[j].Code
	})
}

func (s *Service) withSupplementalDistricts(cityCode string, base []RegionOption) []RegionOption {
	deprecated := deprecatedDistrictCodesByCity[cityCode]
	options := make([]RegionOption, 0, len(base)+len(supplementalDistrictsByCity[cityCode]))
	seen := make(map[string]struct{}, len(base)+len(supplementalDistrictsByCity[cityCode]))
	for _, item := range base {
		if _, removed := deprecated[item.Code]; removed {
			continue
		}
		if _, exists := seen[item.Code]; exists {
			continue
		}
		seen[item.Code] = struct{}{}
		options = append(options, item)
	}
	for _, item := range supplementalDistrictsByCity[cityCode] {
		if _, exists := seen[item.Code]; exists {
			continue
		}
		seen[item.Code] = struct{}{}
		options = append(options, item)
	}
	sortOptions(options)
	return options
}

func supplementalDistrictSelection(districtCode string) (RegionSelection, bool) {
	switch districtCode {
	case "330113":
		return RegionSelection{
			ProvinceCode: "330000",
			ProvinceName: "浙江省",
			CityCode:     "330100",
			CityName:     "杭州市",
			DistrictCode: "330113",
			DistrictName: "临平区",
		}, true
	case "330114":
		return RegionSelection{
			ProvinceCode: "330000",
			ProvinceName: "浙江省",
			CityCode:     "330100",
			CityName:     "杭州市",
			DistrictCode: "330114",
			DistrictName: "钱塘区",
		}, true
	default:
		return RegionSelection{}, false
	}
}

func validateSupplementalSelection(expected RegionSelection, input RegionSelection) (RegionSelection, error) {
	if input.ProvinceCode != "" && strings.TrimSpace(input.ProvinceCode) != expected.ProvinceCode {
		return RegionSelection{}, fmt.Errorf("省、市和末级行政区选择不一致")
	}
	if input.CityCode != "" && strings.TrimSpace(input.CityCode) != expected.CityCode {
		return RegionSelection{}, fmt.Errorf("省、市和末级行政区选择不一致")
	}
	if input.ProvinceName != "" && strings.TrimSpace(input.ProvinceName) != expected.ProvinceName {
		return RegionSelection{}, fmt.Errorf("省级行政区名称与编码不匹配")
	}
	if input.CityName != "" && strings.TrimSpace(input.CityName) != expected.CityName {
		return RegionSelection{}, fmt.Errorf("中间层级行政区名称与编码不匹配")
	}
	if input.DistrictName != "" && strings.TrimSpace(input.DistrictName) != expected.DistrictName {
		return RegionSelection{}, fmt.Errorf("末级行政区名称与编码不匹配")
	}
	return expected, nil
}

func normalizeRegionName(name string) string {
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
	return replacer.Replace(strings.TrimSpace(name))
}

func normalizeDistrictCandidates(primary string, extras []string) []string {
	candidates := make([]string, 0, 1+len(extras))
	seen := make(map[string]struct{}, 1+len(extras))
	appendCandidate := func(value string) {
		normalized := normalizeRegionName(value)
		if normalized == "" {
			return
		}
		if _, exists := seen[normalized]; exists {
			return
		}
		seen[normalized] = struct{}{}
		candidates = append(candidates, normalized)
	}
	appendCandidate(primary)
	for _, item := range extras {
		appendCandidate(item)
	}
	return candidates
}

func matchesAnyDistrictCandidate(districtName string, candidates []string) bool {
	normalizedDistrict := normalizeRegionName(districtName)
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if normalizedDistrict == candidate {
			return true
		}
		if strings.Contains(candidate, normalizedDistrict) || strings.Contains(normalizedDistrict, candidate) {
			return true
		}
	}
	return false
}
