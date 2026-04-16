package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetGeoCaseMapHandle 返回管理员全国反诈案件地理可视化统计数据。
func GetGeoCaseMapHandle(c *gin.Context) {
	result, err := buildGeoCaseMapOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "全国地理统计查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetGeoCaseMapChildrenHandle 返回指定父级行政区的子级地区统计。
func GetGeoCaseMapChildrenHandle(c *gin.Context) {
	parentCode := strings.TrimSpace(c.Query("parent_code"))
	level := strings.TrimSpace(c.Query("level"))
	if parentCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "parent_code 不能为空"})
		return
	}

	result, err := buildGeoCaseMapChildren(parentCode, level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "子级地区统计查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetGeoCaseRegionCasesHandle 返回指定地区在当前时间窗口下的案件摘要列表。
func GetGeoCaseRegionCasesHandle(c *gin.Context) {
	regionCode := strings.TrimSpace(c.Query("region_code"))
	window := strings.TrimSpace(c.DefaultQuery("window", "last_7d"))
	if regionCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "region_code 不能为空"})
		return
	}
	page := parsePositiveInt(c.DefaultQuery("page", "1"), 1)
	pageSize := parsePositiveInt(c.DefaultQuery("page_size", "10"), 10)

	result, err := buildGeoCaseRegionCases(regionCode, window, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "地区案件摘要查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
