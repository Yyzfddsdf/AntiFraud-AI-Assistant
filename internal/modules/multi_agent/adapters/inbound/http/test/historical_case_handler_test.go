package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "antifraud/internal/modules/multi_agent/adapters/inbound/http"

	"github.com/gin-gonic/gin"
)

func TestGetHistoricalCasePreviewHandle_InvalidPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/cases", httpapi.GetHistoricalCasePreviewHandle)

	req := httptest.NewRequest(http.MethodGet, "/cases?page=0&page_size=abc", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusBadRequest)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["error"] == nil {
		t.Fatalf("expected error field, got: %+v", payload)
	}
}
