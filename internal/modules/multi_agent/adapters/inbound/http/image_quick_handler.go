package httpapi

import (
	"net/http"
	"strings"

	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/modules/multi_agent/core"

	"github.com/gin-gonic/gin"
)

var AnalyzeImageQuickFunc = multi_agent.AnalyzeImageQuick

// AnalyzeImageQuickHandle 同步执行单图快速风险识别并直接返回结果。
func AnalyzeImageQuickHandle(c *gin.Context) {
	var payload apimodel.ImageQuickAnalyzeRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	image := strings.TrimSpace(payload.Image)
	if image == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image 不能为空"})
		return
	}

	result, err := AnalyzeImageQuickFunc(image)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "图片快速识别失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, apimodel.ImageQuickAnalyzeResponse{
		RiskLevel: result.RiskLevel,
		Reason:    result.Reason,
	})
}
