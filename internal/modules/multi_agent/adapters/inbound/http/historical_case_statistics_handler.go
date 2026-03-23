package httpapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetHistoricalCaseStatisticsOverviewHandle 返回历史案件库多维度统计总览（仅管理员路由）。
// 可选 query 参数：
// - interval: day/week/month（默认 day）。
func GetHistoricalCaseStatisticsOverviewHandle(c *gin.Context) {
	interval := strings.TrimSpace(c.DefaultQuery("interval", historicalCaseStatsIntervalDay))
	normalized, ok := normalizeHistoricalCaseStatisticsInterval(interval)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "interval 仅支持 day/week/month"})
		return
	}

	result, err := buildHistoricalCaseStatistics(normalized)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件统计查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
