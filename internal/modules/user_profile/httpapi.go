package user_profile_system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	Age            int    `json:"age"`
	Occupation     string `json:"occupation"`
	ProvinceCode   string `json:"province_code"`
	ProvinceName   string `json:"province_name"`
	CityCode       string `json:"city_code"`
	CityName       string `json:"city_name"`
	DistrictCode   string `json:"district_code"`
	DistrictName   string `json:"district_name"`
	LocationSource string `json:"location_source"`
}

func RegisterRoutes(api gin.IRoutes, service *Service) {
	if api == nil {
		return
	}
	if service == nil {
		service = DefaultService()
	}
	api.PUT("/user/profile", updateCurrentUserProfileHandle(service))
	api.GET("/user/profile/options/occupations", getOccupationOptionsHandle())
}

func updateCurrentUserProfileHandle(service *Service) gin.HandlerFunc {
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

		var payload UpdateProfileRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		userResp, err := service.UpdateCurrentUserProfile(numericUserID, UpdateProfileInput{
			Age:            payload.Age,
			Occupation:     payload.Occupation,
			ProvinceCode:   payload.ProvinceCode,
			ProvinceName:   payload.ProvinceName,
			CityCode:       payload.CityCode,
			CityName:       payload.CityName,
			DistrictCode:   payload.DistrictCode,
			DistrictName:   payload.DistrictName,
			LocationSource: payload.LocationSource,
		})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "用户画像更新成功",
			"user":    userResp,
		})
	}
}

func getOccupationOptionsHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		options := ListOccupations()
		c.JSON(http.StatusOK, gin.H{
			"occupations": append([]string{}, options...),
			"count":       len(options),
		})
	}
}

func UpdateCurrentUserProfileHandle(c *gin.Context) {
	updateCurrentUserProfileHandle(DefaultService())(c)
}

func GetOccupationOptionsHandle(c *gin.Context) {
	getOccupationOptionsHandle()(c)
}
