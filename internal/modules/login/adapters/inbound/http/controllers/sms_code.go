package controllers

import (
	"errors"
	"net/http"

	"antifraud/internal/modules/login/adapters/outbound/smscode"
	"antifraud/internal/modules/login/domain/models"

	"github.com/gin-gonic/gin"
)

// SendSMSCodeHandle 返回短信验证码发送处理器。
func SendSMSCodeHandle(service smscode.Service) gin.HandlerFunc {
	if service == nil {
		service = smscode.NewDemoService()
	}

	return func(c *gin.Context) {
		var payload models.SendSMSCodePayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		if err := service.SendCode(c.Request.Context(), payload.Phone); err != nil {
			if errors.Is(err, smscode.ErrInvalidPhoneFormat) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "手机号格式不正确，请输入 11 位大陆手机号"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "短信验证码发送失败"})
			return
		}

		c.JSON(http.StatusOK, models.SendSMSCodeResponse{
			Message: "短信验证码已发送，当前演示环境请使用 000000",
		})
	}
}
