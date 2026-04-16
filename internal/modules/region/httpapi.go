package region_system

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(api gin.IRoutes, service *Service) {
	if api == nil {
		return
	}
	if service == nil {
		service = NewService()
	}
	api.GET("/regions/provinces", listProvincesHandle(service))
	api.GET("/regions/cities", listCitiesHandle(service))
	api.GET("/regions/districts", listDistrictsHandle(service))
	api.POST("/regions/resolve", resolveRegionHandle(service))
	api.GET("/regions/cases/stats/current", currentRegionCaseStatsHandle(service))
}

func listProvincesHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		provinces := service.ListProvinces()
		c.JSON(http.StatusOK, gin.H{"provinces": provinces, "count": len(provinces)})
	}
}

func listCitiesHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		cities, err := service.ListCities(c.Query("province_code"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"cities": cities, "count": len(cities)})
	}
}

func listDistrictsHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		districts, err := service.ListDistricts(c.Query("city_code"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"districts": districts, "count": len(districts)})
	}
}

func resolveRegionHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload ResolveRegionInput
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}
		selection, err := service.ResolveByNames(payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"region": selection})
	}
}

func currentRegionCaseStatsHandle(service *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDValue, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			return
		}

		numericUserID, ok := userIDValue.(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户未认证"})
			return
		}

		result, err := service.GetCurrentUserRegionCaseStats(c.Request.Context(), numericUserID)
		if err != nil {
			switch {
			case errors.Is(err, ErrUserNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			case errors.Is(err, ErrUserRegionNotSet):
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			case errors.Is(err, ErrRegionServiceUnavailable):
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": "地区案件统计查询失败"})
			}
			return
		}

		c.JSON(http.StatusOK, result)
	}
}
