package httpapi

import (
	"net/http"
	"strings"
	"time"

	"antifraud/multi_agent/case_library"
	apimodel "antifraud/multi_agent/httpapi/models"

	"github.com/gin-gonic/gin"
)

// CreateHistoricalCaseHandle 上传历史案件并自动生成 embedding 向量后入库。
// 数据会写入独立的 historical_case_library.db，不占用现有业务数据库文件。
func CreateHistoricalCaseHandle(c *gin.Context) {
	var payload apimodel.CreateHistoricalCaseRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	record, err := case_library.CreateHistoricalCase(c.Request.Context(), getCurrentUserID(c), case_library.CreateHistoricalCaseInput{
		Title:           payload.Title,
		TargetGroup:     payload.TargetGroup,
		RiskLevel:       payload.RiskLevel,
		ScamType:        payload.ScamType,
		CaseDescription: payload.CaseDescription,
		TypicalScripts:  payload.TypicalScripts,
		Keywords:        payload.Keywords,
		ViolatedLaw:     payload.ViolatedLaw,
		Suggestion:      payload.Suggestion,
	})
	if err != nil {
		if case_library.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":                 err.Error(),
				"allowed_target_groups": append([]string{}, case_library.ListTargetGroups()...),
				"allowed_risk_levels":   append([]string{}, case_library.FixedRiskLevels...),
				"allowed_scam_types":    append([]string{}, case_library.ListScamTypes()...),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件入库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, apimodel.CreateHistoricalCaseResponse{
		Message: "historical case stored",
		Case: apimodel.HistoricalCaseItem{
			CaseID:             record.CaseID,
			CreatedBy:          strings.TrimSpace(record.CreatedBy),
			Title:              record.Title,
			TargetGroup:        record.TargetGroup,
			RiskLevel:          record.RiskLevel,
			ScamType:           record.ScamType,
			CaseDescription:    record.CaseDescription,
			TypicalScripts:     append([]string{}, record.TypicalScripts...),
			Keywords:           append([]string{}, record.Keywords...),
			ViolatedLaw:        record.ViolatedLaw,
			Suggestion:         record.Suggestion,
			EmbeddingModel:     record.EmbeddingModel,
			EmbeddingDimension: record.EmbeddingDimension,
			CreatedAt:          record.CreatedAt.Format(time.RFC3339),
		},
	})
}

// GetHistoricalCasePreviewHandle 返回历史案件预览列表。
// 仅包含标题、目标人群、风险等级以及 case_id（便于前端点详情）。
func GetHistoricalCasePreviewHandle(c *gin.Context) {
	previews, err := case_library.ListHistoricalCasePreviews()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件预览查询失败: " + err.Error()})
		return
	}

	items := make([]apimodel.HistoricalCasePreviewItem, 0, len(previews))
	for _, preview := range previews {
		items = append(items, apimodel.HistoricalCasePreviewItem{
			CaseID:      preview.CaseID,
			Title:       preview.Title,
			TargetGroup: preview.TargetGroup,
			RiskLevel:   preview.RiskLevel,
			ScamType:    preview.ScamType,
		})
	}

	c.JSON(http.StatusOK, apimodel.HistoricalCasePreviewResponse{
		Total: len(items),
		Cases: items,
	})
}

// GetHistoricalCaseDetailHandle 返回指定 case_id 的完整历史案件详情（包含 embedding 向量）。
func GetHistoricalCaseDetailHandle(c *gin.Context) {
	caseID := strings.TrimSpace(c.Param("caseId"))
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caseId 不能为空"})
		return
	}

	record, exists, err := case_library.GetHistoricalCaseByID(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件详情查询失败: " + err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "历史案件不存在"})
		return
	}

	c.JSON(http.StatusOK, apimodel.HistoricalCaseDetailResponse{
		Case: apimodel.HistoricalCaseDetailItem{
			CaseID:             record.CaseID,
			CreatedBy:          strings.TrimSpace(record.CreatedBy),
			Title:              record.Title,
			TargetGroup:        record.TargetGroup,
			RiskLevel:          record.RiskLevel,
			ScamType:           record.ScamType,
			CaseDescription:    record.CaseDescription,
			TypicalScripts:     append([]string{}, record.TypicalScripts...),
			Keywords:           append([]string{}, record.Keywords...),
			ViolatedLaw:        record.ViolatedLaw,
			Suggestion:         record.Suggestion,
			EmbeddingVector:    append([]float64{}, record.EmbeddingVector...),
			EmbeddingModel:     record.EmbeddingModel,
			EmbeddingDimension: record.EmbeddingDimension,
			CreatedAt:          record.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          record.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// DeleteHistoricalCaseHandle 删除指定 case_id 的历史案件。
func DeleteHistoricalCaseHandle(c *gin.Context) {
	caseID := strings.TrimSpace(c.Param("caseId"))
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "caseId 不能为空"})
		return
	}

	deleted, err := case_library.DeleteHistoricalCaseByID(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "历史案件删除失败: " + err.Error()})
		return
	}
	if !deleted {
		c.JSON(http.StatusNotFound, gin.H{"error": "历史案件不存在"})
		return
	}

	c.JSON(http.StatusOK, apimodel.DeleteHistoricalCaseResponse{
		CaseID:  caseID,
		Message: "历史案件删除成功",
	})
}

// GetHistoricalCaseScamTypeOptionsHandle 返回可选诈骗类型列表（仅管理员）。
func GetHistoricalCaseScamTypeOptionsHandle(c *gin.Context) {
	options := case_library.ListScamTypes()
	c.JSON(http.StatusOK, gin.H{
		"total":   len(options),
		"options": options,
	})
}

// GetHistoricalCaseTargetGroupOptionsHandle 返回可选目标人群列表（仅管理员）。
func GetHistoricalCaseTargetGroupOptionsHandle(c *gin.Context) {
	options := case_library.ListTargetGroups()
	c.JSON(http.StatusOK, gin.H{
		"total":   len(options),
		"options": options,
	})
}
