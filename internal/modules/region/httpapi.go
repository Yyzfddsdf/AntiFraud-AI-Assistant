package region_system

import (
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
