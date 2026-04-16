package httpapi

import (
	"net/http"
	"strings"
	"time"

	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/modules/multi_agent/adapters/outbound/case_library"

	"github.com/gin-gonic/gin"
)

var reviewCaseLibraryService = case_library.DefaultService()
var rejectPendingReview = reviewCaseLibraryService.RejectPendingReview

// GetPendingReviewCasesHandle 返回所有待审核案件预览列表。
func GetPendingReviewCasesHandle(c *gin.Context) {
	previews, err := reviewCaseLibraryService.ListPendingReviewPreviews()
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
			ViolatedLaw: p.ViolatedLaw,
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

	record, exists, err := reviewCaseLibraryService.GetPendingReviewByID(recordID)
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

	record, err := reviewCaseLibraryService.ApprovePendingReview(c.Request.Context(), recordID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核入库失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, apimodel.ApproveReviewResponse{
		Message: "审核通过，案件已入库知识库",
		CaseID:  record.CaseID,
	})
}

// RejectPendingReviewCaseHandle 审核拒绝待审核案件，从待审核列表移除。
func RejectPendingReviewCaseHandle(c *gin.Context) {
	recordID := strings.TrimSpace(c.Param("recordId"))
	if recordID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recordId 不能为空"})
		return
	}

	if err := rejectPendingReview(c.Request.Context(), recordID); err != nil {
		if strings.Contains(err.Error(), "not found or already processed") {
			c.JSON(http.StatusNotFound, gin.H{"error": "待审核案件不存在或已处理"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审核拒绝失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, apimodel.RejectReviewResponse{
		Message:  "审核拒绝，案件已从待审核列表移除",
		RecordID: recordID,
	})
}
