package models

type GeoCaseMapTopScamType struct {
	ScamType string `json:"scam_type"`
	Count    int    `json:"count"`
}

type GeoCaseMapCaseSummaryItem struct {
	RecordID    string `json:"record_id"`
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	ScamType    string `json:"scam_type,omitempty"`
	RiskLevel   string `json:"risk_level"`
	CreatedAt   string `json:"created_at"`
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

type GeoCaseMapRegionItem struct {
	RegionCode string                `json:"region_code"`
	RegionName string                `json:"region_name"`
	Stats      GeoCaseMapRegionStats `json:"stats"`
}

type GeoCaseMapSummary struct {
	TotalUsersWithLocation int `json:"total_users_with_location"`
	TotalCases             int `json:"total_cases"`
	ProvinceCount          int `json:"province_count"`
	CityCount              int `json:"city_count"`
}

type GeoCaseMapResponse struct {
	GeneratedAt string                 `json:"generated_at"`
	Level       string                 `json:"level"`
	Summary     GeoCaseMapSummary      `json:"summary"`
	Regions     []GeoCaseMapRegionItem `json:"regions"`
}

type GeoCaseMapChildrenResponse struct {
	GeneratedAt string                 `json:"generated_at"`
	Level       string                 `json:"level"`
	ParentCode  string                 `json:"parent_code"`
	ParentName  string                 `json:"parent_name"`
	RegionCount int                    `json:"region_count"`
	Regions     []GeoCaseMapRegionItem `json:"regions"`
}

type GeoCaseMapRegionCasesResponse struct {
	GeneratedAt string                      `json:"generated_at"`
	RegionCode  string                      `json:"region_code"`
	RegionName  string                      `json:"region_name"`
	Window      string                      `json:"window"`
	CaseCount   int                         `json:"case_count"`
	Page        int                         `json:"page"`
	PageSize    int                         `json:"page_size"`
	Total       int                         `json:"total"`
	TotalPages  int                         `json:"total_pages"`
	HasPrev     bool                        `json:"has_prev"`
	HasNext     bool                        `json:"has_next"`
	Cases       []GeoCaseMapCaseSummaryItem `json:"cases"`
}
