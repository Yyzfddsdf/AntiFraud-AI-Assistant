package httpapi

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	"antifraud/internal/platform/cache"
)

const (
	defaultHistoricalCaseGraphTopK = 5
	maxHistoricalCaseGraphTopK     = 10
	historicalCaseGraphCacheTTL    = 5 * time.Minute
)

const historicalCaseGraphCacheKeyPrefix = "cache:case_library:graph:v1:"

type historicalCaseGraphAggregate struct {
	ScamType         string
	CaseCount        int
	High             int
	Medium           int
	Low              int
	TargetGroupCount map[string]int
	KeywordCount     map[string]int
	KeywordSet       map[string]struct{}
	TargetGroupSet   map[string]struct{}
	EmbeddingVectors [][]float64
}

func normalizeHistoricalCaseGraphTopK(topK int) int {
	if topK <= 0 {
		return defaultHistoricalCaseGraphTopK
	}
	if topK > maxHistoricalCaseGraphTopK {
		return maxHistoricalCaseGraphTopK
	}
	return topK
}

func buildHistoricalCaseGraph(focusType string, focusGroup string, topK int) (apimodel.HistoricalCaseGraphResponse, error) {
	normalizedFocusType := strings.TrimSpace(focusType)
	normalizedFocusGroup := strings.TrimSpace(focusGroup)
	appliedTopK := normalizeHistoricalCaseGraphTopK(topK)
	cacheKey := buildHistoricalCaseGraphCacheKey(normalizedFocusType, normalizedFocusGroup, appliedTopK, case_library.HistoricalCaseGraphCacheVersion())
	var cached apimodel.HistoricalCaseGraphResponse
	found, cacheErr := cache.GetJSON(cacheKey, &cached)
	if cacheErr == nil && found {
		return cached, nil
	}

	aggregates := make(map[string]*historicalCaseGraphAggregate)
	err := case_library.StreamAllHistoricalCases(func(record case_library.HistoricalCaseRecord) error {
		updateHistoricalCaseGraphAggregate(aggregates, record)
		return nil
	})
	if err != nil {
		return apimodel.HistoricalCaseGraphResponse{}, err
	}

	result := BuildHistoricalCaseGraphFromAggregates(aggregates, normalizedFocusType, normalizedFocusGroup, appliedTopK)
	if cacheErr := cache.SetJSON(cacheKey, result, historicalCaseGraphCacheTTL); cacheErr != nil {
		// 缓存失败不影响主流程，保持只读分析接口可用。
	}
	return result, nil
}

// BuildHistoricalCaseGraphFromAggregates 基于预先聚合的数据构建诈骗类型画像与相似关系图谱。
func BuildHistoricalCaseGraphFromAggregates(aggregates map[string]*historicalCaseGraphAggregate, focusType string, focusGroup string, topK int) apimodel.HistoricalCaseGraphResponse {
	appliedTopK := normalizeHistoricalCaseGraphTopK(topK)
	focus := strings.TrimSpace(focusType)
	groupFocus := strings.TrimSpace(focusGroup)
	filtered := filterHistoricalCaseGraphAggregates(aggregates, focus)
	profiles, graph, targetGroupCount, keywordCount := buildHistoricalCaseGraphArtifacts(filtered, appliedTopK)
	targetGroupTopScamTypes := buildHistoricalCaseTargetGroupTopScamTypes(filtered, groupFocus, appliedTopK)

	return apimodel.HistoricalCaseGraphResponse{
		Summary: apimodel.HistoricalCaseGraphSummary{
			FocusType:        focus,
			FocusGroup:       groupFocus,
			TopK:             appliedTopK,
			TotalCases:       countHistoricalCaseGraphCases(filtered),
			ScamTypeCount:    len(filtered),
			TargetGroupCount: targetGroupCount,
			KeywordCount:     keywordCount,
		},
		Profiles:                profiles,
		Graph:                   graph,
		TargetGroupTopScamTypes: targetGroupTopScamTypes,
	}
}

// BuildHistoricalCaseGraphFromRecords 基于案件记录构建诈骗类型画像与相似关系图谱（仅用于向后兼容或测试）。
func BuildHistoricalCaseGraphFromRecords(records []case_library.HistoricalCaseRecord, focusType string, focusGroup string, topK int) apimodel.HistoricalCaseGraphResponse {
	aggregates := buildHistoricalCaseGraphAggregates(records)
	return BuildHistoricalCaseGraphFromAggregates(aggregates, focusType, focusGroup, topK)
}

func buildHistoricalCaseGraphCacheKey(focusType string, focusGroup string, topK int, version string) string {
	normalizedFocusType := strings.TrimSpace(focusType)
	if normalizedFocusType == "" {
		normalizedFocusType = "all"
	}
	normalizedFocusGroup := strings.TrimSpace(focusGroup)
	if normalizedFocusGroup == "" {
		normalizedFocusGroup = "all_groups"
	}
	trimmedVersion := strings.TrimSpace(version)
	if trimmedVersion == "" {
		trimmedVersion = "0"
	}
	replacer := strings.NewReplacer(":", "_", " ", "_", "/", "_", "\\", "_")
	return historicalCaseGraphCacheKeyPrefix + replacer.Replace(normalizedFocusType) + fmt.Sprintf(":group:%s:topk:%d:version:%s", replacer.Replace(normalizedFocusGroup), topK, trimmedVersion)
}

func buildHistoricalCaseGraphAggregates(records []case_library.HistoricalCaseRecord) map[string]*historicalCaseGraphAggregate {
	aggregates := make(map[string]*historicalCaseGraphAggregate)
	for _, record := range records {
		updateHistoricalCaseGraphAggregate(aggregates, record)
	}
	return aggregates
}

func updateHistoricalCaseGraphAggregate(aggregates map[string]*historicalCaseGraphAggregate, record case_library.HistoricalCaseRecord) {
	scamType := strings.TrimSpace(record.ScamType)
	if scamType == "" {
		scamType = historicalCaseStatsUnknownCategory
	}
	agg, exists := aggregates[scamType]
	if !exists {
		agg = &historicalCaseGraphAggregate{
			ScamType:         scamType,
			TargetGroupCount: map[string]int{},
			KeywordCount:     map[string]int{},
			KeywordSet:       map[string]struct{}{},
			TargetGroupSet:   map[string]struct{}{},
			EmbeddingVectors: make([][]float64, 0),
		}
		aggregates[scamType] = agg
	}

	agg.CaseCount++
	switch strings.TrimSpace(record.RiskLevel) {
	case "高":
		agg.High++
	case "低":
		agg.Low++
	default:
		agg.Medium++
	}

	targetGroup := strings.TrimSpace(record.TargetGroup)
	if targetGroup == "" {
		targetGroup = historicalCaseStatsUnknownCategory
	}
	agg.TargetGroupCount[targetGroup]++
	agg.TargetGroupSet[targetGroup] = struct{}{}

	for _, keyword := range record.Keywords {
		trimmedKeyword := strings.TrimSpace(keyword)
		if trimmedKeyword == "" {
			continue
		}
		agg.KeywordCount[trimmedKeyword]++
		agg.KeywordSet[trimmedKeyword] = struct{}{}
	}

	if len(record.EmbeddingVector) > 0 {
		agg.EmbeddingVectors = append(agg.EmbeddingVectors, append([]float64{}, record.EmbeddingVector...))
	}
}

func filterHistoricalCaseGraphAggregates(aggregates map[string]*historicalCaseGraphAggregate, focusType string) []*historicalCaseGraphAggregate {
	items := make([]*historicalCaseGraphAggregate, 0, len(aggregates))
	for _, aggregate := range aggregates {
		if focusType != "" && aggregate.ScamType != focusType {
			continue
		}
		items = append(items, aggregate)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].CaseCount == items[j].CaseCount {
			return items[i].ScamType < items[j].ScamType
		}
		return items[i].CaseCount > items[j].CaseCount
	})
	return items
}

func buildHistoricalCaseGraphArtifacts(aggregates []*historicalCaseGraphAggregate, topK int) ([]apimodel.HistoricalCaseGraphProfile, apimodel.HistoricalCaseGraphData, int, int) {
	profiles := make([]apimodel.HistoricalCaseGraphProfile, 0, len(aggregates))
	nodes := make([]apimodel.HistoricalCaseGraphNode, 0)
	edges := make([]apimodel.HistoricalCaseGraphEdge, 0)
	nodeSeen := make(map[string]struct{})
	targetGroupsSeen := make(map[string]struct{})
	keywordsSeen := make(map[string]struct{})

	for _, aggregate := range aggregates {
		similarities := buildHistoricalCaseGraphSimilarities(aggregate, aggregates, topK)
		profile := apimodel.HistoricalCaseGraphProfile{
			ScamType:  aggregate.ScamType,
			CaseCount: aggregate.CaseCount,
			RiskDistribution: apimodel.HistoricalCaseGraphRiskStats{
				High:   aggregate.High,
				Medium: aggregate.Medium,
				Low:    aggregate.Low,
				Total:  aggregate.CaseCount,
			},
			TopTargetGroups: topHistoricalCaseGraphNamedCounts(aggregate.TargetGroupCount, topK),
			TopKeywords:     topHistoricalCaseGraphNamedCounts(aggregate.KeywordCount, topK),
			SimilarTypes:    similarities,
		}
		profiles = append(profiles, profile)

		scamNodeID := buildHistoricalCaseGraphNodeID("scam_type", aggregate.ScamType)
		appendHistoricalCaseGraphNode(&nodes, nodeSeen, apimodel.HistoricalCaseGraphNode{
			ID:       scamNodeID,
			NodeType: "scam_type",
			Label:    aggregate.ScamType,
			Weight:   aggregate.CaseCount,
		})

		for _, item := range profile.TopTargetGroups {
			targetNodeID := buildHistoricalCaseGraphNodeID("target_group", item.Name)
			appendHistoricalCaseGraphNode(&nodes, nodeSeen, apimodel.HistoricalCaseGraphNode{
				ID:       targetNodeID,
				NodeType: "target_group",
				Label:    item.Name,
				Weight:   item.Count,
			})
			targetGroupsSeen[item.Name] = struct{}{}
			edges = append(edges, apimodel.HistoricalCaseGraphEdge{
				Source:   scamNodeID,
				Target:   targetNodeID,
				Relation: "targets",
				Score:    safeRatio(item.Count, aggregate.CaseCount),
			})
		}

		for _, item := range profile.TopKeywords {
			keywordNodeID := buildHistoricalCaseGraphNodeID("keyword", item.Name)
			appendHistoricalCaseGraphNode(&nodes, nodeSeen, apimodel.HistoricalCaseGraphNode{
				ID:       keywordNodeID,
				NodeType: "keyword",
				Label:    item.Name,
				Weight:   item.Count,
			})
			keywordsSeen[item.Name] = struct{}{}
			edges = append(edges, apimodel.HistoricalCaseGraphEdge{
				Source:   scamNodeID,
				Target:   keywordNodeID,
				Relation: "keyword",
				Score:    safeRatio(item.Count, aggregate.CaseCount),
			})
		}

		for _, item := range profile.SimilarTypes {
			targetNodeID := buildHistoricalCaseGraphNodeID("scam_type", item.ScamType)
			appendHistoricalCaseGraphNode(&nodes, nodeSeen, apimodel.HistoricalCaseGraphNode{
				ID:       targetNodeID,
				NodeType: "scam_type",
				Label:    item.ScamType,
				Weight:   caseCountForScamType(aggregates, item.ScamType),
			})
			edges = append(edges, apimodel.HistoricalCaseGraphEdge{
				Source:   scamNodeID,
				Target:   targetNodeID,
				Relation: "similar",
				Score:    item.Score,
			})
		}
	}

	return profiles, apimodel.HistoricalCaseGraphData{Nodes: nodes, Edges: edges}, len(targetGroupsSeen), len(keywordsSeen)
}

func buildHistoricalCaseGraphSimilarities(current *historicalCaseGraphAggregate, aggregates []*historicalCaseGraphAggregate, topK int) []apimodel.HistoricalCaseGraphSimilarityItem {
	items := make([]apimodel.HistoricalCaseGraphSimilarityItem, 0, len(aggregates))
	for _, candidate := range aggregates {
		if candidate == nil || candidate.ScamType == current.ScamType {
			continue
		}
		score := calculateHistoricalCaseGraphSimilarity(current, candidate)
		if score <= 0 {
			continue
		}
		items = append(items, apimodel.HistoricalCaseGraphSimilarityItem{ScamType: candidate.ScamType, Score: score})
	}
	sort.Slice(items, func(i, j int) bool {
		if math.Abs(items[i].Score-items[j].Score) < 1e-12 {
			return items[i].ScamType < items[j].ScamType
		}
		return items[i].Score > items[j].Score
	})
	if len(items) > topK {
		items = items[:topK]
	}
	return items
}

func calculateHistoricalCaseGraphSimilarity(left *historicalCaseGraphAggregate, right *historicalCaseGraphAggregate) float64 {
	if left == nil || right == nil {
		return 0
	}
	vectorScore := cosineSimilarity(normalizeMeanVector(left.EmbeddingVectors), normalizeMeanVector(right.EmbeddingVectors))
	if vectorScore < 0 {
		vectorScore = 0
	}
	keywordScore := jaccardSimilarity(left.KeywordSet, right.KeywordSet)
	targetGroupScore := jaccardSimilarity(left.TargetGroupSet, right.TargetGroupSet)
	score := 0.6*vectorScore + 0.25*keywordScore + 0.15*targetGroupScore
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func topHistoricalCaseGraphNamedCounts(counter map[string]int, topK int) []apimodel.HistoricalCaseGraphNamedCountItem {
	items := make([]apimodel.HistoricalCaseGraphNamedCountItem, 0, len(counter))
	for name, count := range counter {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" || count <= 0 {
			continue
		}
		items = append(items, apimodel.HistoricalCaseGraphNamedCountItem{Name: trimmedName, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Name < items[j].Name
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > topK {
		items = items[:topK]
	}
	return items
}

func buildHistoricalCaseTargetGroupTopScamTypes(aggregates []*historicalCaseGraphAggregate, focusGroup string, topK int) []apimodel.HistoricalCaseGraphTargetGroupScamTypeTopKItem {
	normalizedFocusGroup := strings.TrimSpace(focusGroup)
	byTargetGroup := make(map[string]map[string]int)
	totalCasesByTargetGroup := make(map[string]int)

	for _, aggregate := range aggregates {
		if aggregate == nil {
			continue
		}
		for targetGroup, count := range aggregate.TargetGroupCount {
			normalizedTargetGroup := strings.TrimSpace(targetGroup)
			if normalizedTargetGroup == "" {
				normalizedTargetGroup = historicalCaseStatsUnknownCategory
			}
			if count <= 0 {
				continue
			}
			counter, exists := byTargetGroup[normalizedTargetGroup]
			if !exists {
				counter = make(map[string]int)
				byTargetGroup[normalizedTargetGroup] = counter
			}
			counter[aggregate.ScamType] += count
			totalCasesByTargetGroup[normalizedTargetGroup] += count
		}
	}

	items := make([]apimodel.HistoricalCaseGraphTargetGroupScamTypeTopKItem, 0, len(byTargetGroup))
	for targetGroup, counter := range byTargetGroup {
		if normalizedFocusGroup != "" && targetGroup != normalizedFocusGroup {
			continue
		}
		items = append(items, apimodel.HistoricalCaseGraphTargetGroupScamTypeTopKItem{
			TargetGroup:  targetGroup,
			TotalCases:   totalCasesByTargetGroup[targetGroup],
			TopScamTypes: topHistoricalCaseGraphScamTypeScores(counter, totalCasesByTargetGroup[targetGroup], topK),
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalCases == items[j].TotalCases {
			return items[i].TargetGroup < items[j].TargetGroup
		}
		return items[i].TotalCases > items[j].TotalCases
	})
	return items
}

func topHistoricalCaseGraphScamTypeScores(counter map[string]int, totalCases int, topK int) []apimodel.HistoricalCaseGraphTargetGroupScamTypeScoreItem {
	items := make([]apimodel.HistoricalCaseGraphTargetGroupScamTypeScoreItem, 0, len(counter))
	for scamType, score := range counter {
		trimmedScamType := strings.TrimSpace(scamType)
		if trimmedScamType == "" {
			trimmedScamType = historicalCaseStatsUnknownCategory
		}
		if score <= 0 {
			continue
		}
		items = append(items, apimodel.HistoricalCaseGraphTargetGroupScamTypeScoreItem{
			ScamType: trimmedScamType,
			Score:    safeRatio(score, totalCases),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if math.Abs(items[i].Score-items[j].Score) < 1e-12 {
			return items[i].ScamType < items[j].ScamType
		}
		return items[i].Score > items[j].Score
	})
	if len(items) > topK {
		items = items[:topK]
	}
	return items
}

func normalizeMeanVector(vectors [][]float64) []float64 {
	if len(vectors) == 0 {
		return nil
	}
	maxLen := 0
	for _, vector := range vectors {
		if len(vector) > maxLen {
			maxLen = len(vector)
		}
	}
	if maxLen == 0 {
		return nil
	}
	mean := make([]float64, maxLen)
	for _, vector := range vectors {
		for index, value := range vector {
			mean[index] += sanitizeHistoricalCaseGraphFloat(value)
		}
	}
	count := float64(len(vectors))
	for index := range mean {
		mean[index] = mean[index] / count
	}
	return normalizeL2VectorHistoricalCaseGraph(mean)
}

func normalizeL2VectorHistoricalCaseGraph(vector []float64) []float64 {
	if len(vector) == 0 {
		return nil
	}
	result := make([]float64, len(vector))
	var norm2 float64
	for index, value := range vector {
		cleanValue := sanitizeHistoricalCaseGraphFloat(value)
		result[index] = cleanValue
		norm2 += cleanValue * cleanValue
	}
	if norm2 <= 0 {
		return nil
	}
	norm := math.Sqrt(norm2)
	for index := range result {
		result[index] = result[index] / norm
	}
	return result
}

func cosineSimilarity(left []float64, right []float64) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	var dot float64
	for index := 0; index < limit; index++ {
		dot += left[index] * right[index]
	}
	if dot < -1 {
		return -1
	}
	if dot > 1 {
		return 1
	}
	return dot
}

func jaccardSimilarity(left map[string]struct{}, right map[string]struct{}) float64 {
	if len(left) == 0 || len(right) == 0 {
		return 0
	}
	union := make(map[string]struct{}, len(left)+len(right))
	intersection := 0
	for key := range left {
		union[key] = struct{}{}
		if _, exists := right[key]; exists {
			intersection++
		}
	}
	for key := range right {
		union[key] = struct{}{}
	}
	if len(union) == 0 {
		return 0
	}
	return float64(intersection) / float64(len(union))
}

func sanitizeHistoricalCaseGraphFloat(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return value
}

func countHistoricalCaseGraphCases(aggregates []*historicalCaseGraphAggregate) int {
	total := 0
	for _, aggregate := range aggregates {
		total += aggregate.CaseCount
	}
	return total
}

func buildHistoricalCaseGraphNodeID(nodeType string, label string) string {
	return fmt.Sprintf("%s:%s", strings.TrimSpace(nodeType), strings.TrimSpace(label))
}

func appendHistoricalCaseGraphNode(nodes *[]apimodel.HistoricalCaseGraphNode, seen map[string]struct{}, node apimodel.HistoricalCaseGraphNode) {
	if nodes == nil || seen == nil {
		return
	}
	if _, exists := seen[node.ID]; exists {
		return
	}
	seen[node.ID] = struct{}{}
	*nodes = append(*nodes, node)
}

func caseCountForScamType(aggregates []*historicalCaseGraphAggregate, scamType string) int {
	for _, aggregate := range aggregates {
		if aggregate.ScamType == scamType {
			return aggregate.CaseCount
		}
	}
	return 0
}

func safeRatio(count int, total int) float64 {
	if total <= 0 || count <= 0 {
		return 0
	}
	ratio := float64(count) / float64(total)
	if ratio > 1 {
		return 1
	}
	return ratio
}
