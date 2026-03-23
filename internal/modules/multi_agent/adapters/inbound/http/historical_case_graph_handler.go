package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetHistoricalCaseGraphHandle 返回案件知识库 V1 图谱分析（仅管理员）。
// 可选 query 参数：
// - focus_type: 仅查看某个诈骗类型的画像与局部图谱；
// - focus_group: 仅查看指定目标人群的人群 TopK 统计；
// - top_k: 控制每个诈骗类型返回的 top 目标人群/关键词/相似类型数量。
func GetHistoricalCaseGraphHandle(c *gin.Context) {
	focusType := strings.TrimSpace(c.Query("focus_type"))
	focusGroup := strings.TrimSpace(c.Query("focus_group"))
	if focusGroup == "" {
		focusGroup = strings.TrimSpace(c.Query("focus_gropu"))
	}
	topKRaw := strings.TrimSpace(c.DefaultQuery("top_k", strconv.Itoa(defaultHistoricalCaseGraphTopK)))
	topK, err := strconv.Atoi(topKRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "top_k 必须是整数"})
		return
	}

	result, graphErr := buildHistoricalCaseGraph(focusType, focusGroup, topK)
	if graphErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件图谱分析失败: " + graphErr.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
