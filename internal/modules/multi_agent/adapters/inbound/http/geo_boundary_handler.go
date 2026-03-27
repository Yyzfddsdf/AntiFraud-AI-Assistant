package httpapi

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var fetchGeoBoundaryGeoJSONFunc = fetchGeoBoundaryGeoJSON

// GetGeoBoundaryGeoJSONHandle 返回地图边界 GeoJSON，供前端同源加载。
func GetGeoBoundaryGeoJSONHandle(c *gin.Context) {
	payload, err := fetchGeoBoundaryGeoJSONFunc(c.Request.Context(), c.Query("code"))
	if err != nil {
		var serviceErr *geoBoundaryServiceError
		if errors.As(err, &serviceErr) {
			c.JSON(serviceErr.statusCode, gin.H{"error": serviceErr.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "地图边界数据查询失败: " + err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json; charset=utf-8", payload)
}
