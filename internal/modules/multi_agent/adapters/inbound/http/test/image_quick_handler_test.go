package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "antifraud/internal/modules/multi_agent/adapters/inbound/http"
	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/modules/multi_agent/core"

	"github.com/gin-gonic/gin"
)

func TestAnalyzeImageQuickHandle_BadRequestWhenImageEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/quick", httpapi.AnalyzeImageQuickHandle)

	req := httptest.NewRequest(http.MethodPost, "/quick", bytes.NewBufferString(`{"image":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestAnalyzeImageQuickHandle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalAnalyze := httpapi.AnalyzeImageQuickFunc
	t.Cleanup(func() {
		httpapi.AnalyzeImageQuickFunc = originalAnalyze
	})

	httpapi.AnalyzeImageQuickFunc = func(imageBase64 string) (multi_agent.ImageQuickRiskResponse, error) {
		if imageBase64 != "base64-image" {
			t.Fatalf("unexpected image payload: %q", imageBase64)
		}
		return multi_agent.ImageQuickRiskResponse{
			RiskLevel: "高",
			Reason:    "图片中出现仿冒客服页面与转账引导",
		}, nil
	}

	router := gin.New()
	router.POST("/quick", httpapi.AnalyzeImageQuickHandle)

	req := httptest.NewRequest(http.MethodPost, "/quick", bytes.NewBufferString(`{"image":"base64-image"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}

	var payload apimodel.ImageQuickAnalyzeResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.RiskLevel != "高" || payload.Reason == "" {
		t.Fatalf("unexpected response payload: %+v", payload)
	}
}

func TestAnalyzeImageQuickHandle_UpstreamFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalAnalyze := httpapi.AnalyzeImageQuickFunc
	t.Cleanup(func() {
		httpapi.AnalyzeImageQuickFunc = originalAnalyze
	})

	httpapi.AnalyzeImageQuickFunc = func(string) (multi_agent.ImageQuickRiskResponse, error) {
		return multi_agent.ImageQuickRiskResponse{}, errors.New("upstream failed")
	}

	router := gin.New()
	router.POST("/quick", httpapi.AnalyzeImageQuickHandle)

	req := httptest.NewRequest(http.MethodPost, "/quick", bytes.NewBufferString(`{"image":"base64-image"}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadGateway {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}
