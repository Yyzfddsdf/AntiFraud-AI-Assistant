package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetGeoCaseMapHandle 返回管理员全国反诈案件地理可视化统计数据。
func GetGeoCaseMapHandle(c *gin.Context) {
	result, err := buildGeoCaseMap()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "全国地理统计查询失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
