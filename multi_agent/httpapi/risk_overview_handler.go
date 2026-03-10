package httpapi

import (
	"net/http"
	"strings"

	apimodel "antifraud/multi_agent/httpapi/models"
	"antifraud/multi_agent/overview"

	"github.com/gin-gonic/gin"
)

// GetMultimodalRiskOverviewHandle 返回当前用户风险变化趋势与风险等级统计总览。
// 可选 query 参数：
// - interval: day/week/month（默认 day）。
func GetMultimodalRiskOverviewHandle(c *gin.Context) {
	interval := strings.TrimSpace(c.DefaultQuery("interval", overview.IntervalDay))
	normalized, ok := overview.NormalizeInterval(interval)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval 仅支持 day/week/month"})
		return
	}

	userID := getCurrentUserID(c)
	result := overview.BuildUserRiskOverview(userID, normalized)

	trend := make([]apimodel.MultimodalRiskTrendItem, 0, len(result.Trend))
	for _, point := range result.Trend {
		trend = append(trend, apimodel.MultimodalRiskTrendItem{
			TimeBucket: point.TimeBucket,
			High:       point.High,
			Medium:     point.Medium,
			Low:        point.Low,
			Total:      point.Total,
		})
	}

	c.JSON(http.StatusOK, apimodel.MultimodalRiskOverviewResponse{
		Stats: apimodel.MultimodalRiskLevelStats{
			High:   result.Stats.High,
			Medium: result.Stats.Medium,
			Low:    result.Stats.Low,
			Total:  result.Stats.Total,
		},
		Trend: trend,
		Analysis: apimodel.MultimodalRiskTrendAnalysis{
			CurrentBucket:  result.Analysis.CurrentBucket,
			PreviousBucket: result.Analysis.PreviousBucket,
			OverallTrend:   result.Analysis.OverallTrend,
			HighRiskTrend:  result.Analysis.HighRiskTrend,
			Summary:        result.Analysis.Summary,
		},
	})
}
