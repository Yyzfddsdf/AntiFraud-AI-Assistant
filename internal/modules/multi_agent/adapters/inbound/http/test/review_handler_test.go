package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "antifraud/internal/modules/multi_agent/adapters/inbound/http"

	"github.com/gin-gonic/gin"
)

func TestRejectPendingReviewCaseHandleSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalReject := rejectPendingReview
	t.Cleanup(func() {
		rejectPendingReview = originalReject
	})

	calledRecordID := ""
	rejectPendingReview = func(_ context.Context, recordID string) error {
		calledRecordID = recordID
		return nil
	}

	router := gin.New()
	router.POST("/cases/:recordId/reject", httpapi.RejectPendingReviewCaseHandle)

	req := httptest.NewRequest(http.MethodPost, "/cases/PR-001/reject", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
	if calledRecordID != "PR-001" {
		t.Fatalf("unexpected record id: %q", calledRecordID)
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["record_id"] != "PR-001" {
		t.Fatalf("unexpected response payload: %+v", payload)
	}
}

func TestRejectPendingReviewCaseHandleNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalReject := rejectPendingReview
	t.Cleanup(func() {
		rejectPendingReview = originalReject
	})

	rejectPendingReview = func(_ context.Context, recordID string) error {
		return errors.New("pending review case not found or already processed")
	}

	router := gin.New()
	router.POST("/cases/:recordId/reject", httpapi.RejectPendingReviewCaseHandle)

	req := httptest.NewRequest(http.MethodPost, "/cases/PR-404/reject", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}
