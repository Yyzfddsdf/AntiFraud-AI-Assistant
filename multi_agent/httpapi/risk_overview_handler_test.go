package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetMultimodalRiskOverviewHandle_InvalidInterval(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "u-test")
		c.Next()
	})
	router.GET("/overview", GetMultimodalRiskOverviewHandle)

	req := httptest.NewRequest(http.MethodGet, "/overview?interval=year", nil)
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

func TestGetMultimodalRiskOverviewHandle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "u-test")
		c.Next()
	})
	router.GET("/overview", GetMultimodalRiskOverviewHandle)

	req := httptest.NewRequest(http.MethodGet, "/overview?interval=day", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d want=%d", resp.Code, http.StatusOK)
	}

	var payload MultimodalRiskOverviewResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}

	if payload.Stats.Total != payload.Stats.High+payload.Stats.Medium+payload.Stats.Low {
		t.Fatalf("inconsistent stats: %+v", payload.Stats)
	}
	if payload.Trend == nil {
		t.Fatalf("trend should not be nil")
	}
}
