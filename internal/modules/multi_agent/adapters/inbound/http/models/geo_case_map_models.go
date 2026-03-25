package models

type GeoCaseMapTopScamType struct {
	ScamType string `json:"scam_type"`
	Count    int    `json:"count"`
}

type GeoCaseMapWindowStats struct {
	Count         int                     `json:"count"`
	PreviousCount int                     `json:"previous_count"`
	ChangeRate    float64                 `json:"change_rate"`
	Trend         string                  `json:"trend"`
	RiskLevel     string                  `json:"risk_level"`
	TopScamTypes  []GeoCaseMapTopScamType `json:"top_scam_types"`
}

type GeoCaseMapRegionStats struct {
	Today   GeoCaseMapWindowStats `json:"today"`
	Last7d  GeoCaseMapWindowStats `json:"last_7d"`
	Last30d GeoCaseMapWindowStats `json:"last_30d"`
	AllTime GeoCaseMapWindowStats `json:"all_time"`
}

type GeoCaseMapCityItem struct {
	RegionCode string                `json:"region_code"`
	RegionName string                `json:"region_name"`
	Stats      GeoCaseMapRegionStats `json:"stats"`
	Districts  []GeoCaseMapDistrictItem `json:"districts"`
}

type GeoCaseMapDistrictItem struct {
	RegionCode string                `json:"region_code"`
	RegionName string                `json:"region_name"`
	Stats      GeoCaseMapRegionStats `json:"stats"`
}

type GeoCaseMapProvinceItem struct {
	RegionCode string                `json:"region_code"`
	RegionName string                `json:"region_name"`
	Stats      GeoCaseMapRegionStats `json:"stats"`
	Cities     []GeoCaseMapCityItem  `json:"cities"`
}

type GeoCaseMapSummary struct {
	TotalUsersWithLocation int `json:"total_users_with_location"`
	TotalCases             int `json:"total_cases"`
	ProvinceCount          int `json:"province_count"`
	CityCount              int `json:"city_count"`
}

type GeoCaseMapResponse struct {
	GeneratedAt string                   `json:"generated_at"`
	Summary     GeoCaseMapSummary        `json:"summary"`
	Provinces   []GeoCaseMapProvinceItem `json:"provinces"`
}
