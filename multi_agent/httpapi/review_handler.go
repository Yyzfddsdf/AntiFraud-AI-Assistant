package httpapi

import (
	"net/http"
	"strings"
	"time"

	"antifraud/multi_agent/case_library"
	apimodel "antifraud/multi_agent/httpapi/models"

	"github.com/gin-gonic/gin"
)

// GetPendingReviewCasesHandle 返回所有待审核案件预览列表。
func GetPendingReviewCasesHandle(c *gin.Context) {
	previews, err := case_library.ListPendingReviewPreviews()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "待审核案件查询失败: " + err.Error()})
		return
	}

	items := make([]apimodel.PendingReviewPreviewItem, 0, len(previews))
	for _, p := range previews {
		items = append(items, apimodel.PendingReviewPreviewItem{
			RecordID:    p.RecordID,
			Title:       p.Title,
			TargetGroup: p.TargetGroup,
			RiskLevel:   p.RiskLevel,
			ScamType:    p.ScamType,
			CreatedAt:   p.CreatedAt.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, apimodel.PendingReviewPreviewResponse{
		Total: len(items),
		Cases: items,
	})
}
// APPEND_MARKER

// GetPendingReviewCaseDetailHandle 返回指定 recordId 的待审核案件详情。
func GetPendingReviewCaseDetailHandle(c *gin.Context) {
	recordID := strings.TrimSpace(c.Param("recordId"))
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recordId 不能为空"})
		return
	}

	record, exists, err := case_library.GetPendingReviewByID(recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "待审核案件详情查询失败: " + err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "待审核案件不存在"})
		return
	}

	c.JSON(http.StatusOK, apimodel.PendingReviewDetailResponse{
		Case: apimodel.PendingReviewDetailItem{
			RecordID:        record.RecordID,
			UserID:          record.UserID,
			Title:           record.Title,
			TargetGroup:     record.TargetGroup,
			RiskLevel:       record.RiskLevel,
			ScamType:        record.ScamType,
			CaseDescription: record.CaseDescription,
			TypicalScripts:  append([]string{}, record.TypicalScripts...),
			Keywords:        append([]string{}, record.Keywords...),
			ViolatedLaw:     record.ViolatedLaw,
			Suggestion:      record.Suggestion,
			Status:          record.Status,
			CreatedAt:       record.CreatedAt.Format(time.RFC3339),
			UpdatedAt:       record.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// ApprovePendingReviewCaseHandle 审核通过待审核案件，入库知识库。
func ApprovePendingReviewCaseHandle(c *gin.Context) {
	recordID := strings.TrimSpace(c.Param("recordId"))
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recordId 不能为空"})
		return
	}

	record, err := case_library.ApprovePendingReview(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核入库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, apimodel.ApproveReviewResponse{
		Message: "审核通过，案件已入库知识库",
		CaseID:  record.CaseID,
	})
}