package user_profile_system

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UpdateProfileRequest struct {
	Age        int    `json:"age"`
	Occupation string `json:"occupation"`
}

func RegisterRoutes(api gin.IRoutes) {
	api.PUT("/user/profile", UpdateCurrentUserProfileHandle)
	api.GET("/user/profile/options/occupations", GetOccupationOptionsHandle)
}

func UpdateCurrentUserProfileHandle(c *gin.Context) {
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

	userResp, err := UpdateCurrentUserProfile(numericUserID, UpdateProfileInput{
		Age:        payload.Age,
		Occupation: payload.Occupation,
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

func GetOccupationOptionsHandle(c *gin.Context) {
	options := ListOccupations()
	c.JSON(http.StatusOK, gin.H{
		"occupations": append([]string{}, options...),
		"count":       len(options),
	})
}
