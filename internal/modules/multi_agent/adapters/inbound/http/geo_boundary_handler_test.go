package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestNormalizeGeoBoundaryCode(t *testing.T) {
	code, err := normalizeGeoBoundaryCode("")
	if err != nil {
		t.Fatalf("unexpected error for default code: %v", err)
	}
	if code != geoBoundaryDefaultCode {
		t.Fatalf("unexpected default code: got=%q want=%q", code, geoBoundaryDefaultCode)
	}

	if _, err := normalizeGeoBoundaryCode("abc"); err == nil {
		t.Fatalf("expected validation error for invalid code")
	}
}

func TestFetchGeoBoundaryFromUpstream(t *testing.T) {
	originalClient := geoBoundaryHTTPClient
	t.Cleanup(func() {
		geoBoundaryHTTPClient = originalClient
	})

	geoBoundaryHTTPClient = &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", req.Method)
			}
			if req.URL.String() != buildGeoBoundaryUpstreamURL("100000") {
				t.Fatalf("unexpected url: %s", req.URL.String())
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"type":"FeatureCollection","features":[]}`)),
			}, nil
		}),
	}

	payload, err := fetchGeoBoundaryFromUpstream(context.Background(), "100000")
	if err != nil {
		t.Fatalf("fetch upstream failed: %v", err)
	}
	if !json.Valid(payload) {
		t.Fatalf("expected valid json payload, got=%s", string(payload))
	}
}

func TestGetGeoBoundaryGeoJSONHandle_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/geojson", GetGeoBoundaryGeoJSONHandle)

	req := httptest.NewRequest(http.MethodGet, "/geojson?code=abc", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestGetGeoBoundaryGeoJSONHandle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalFetch := fetchGeoBoundaryGeoJSONFunc
	t.Cleanup(func() {
		fetchGeoBoundaryGeoJSONFunc = originalFetch
	})

	fetchGeoBoundaryGeoJSONFunc = func(ctx context.Context, rawCode string) (json.RawMessage, error) {
		if strings.TrimSpace(rawCode) != "100000" {
			t.Fatalf("unexpected code: %q", rawCode)
		}
		return json.RawMessage(`{"type":"FeatureCollection","features":[]}`), nil
	}

	router := gin.New()
	router.GET("/geojson", GetGeoBoundaryGeoJSONHandle)

	req := httptest.NewRequest(http.MethodGet, "/geojson?code=100000", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
	if !strings.Contains(resp.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("unexpected content type: %s", resp.Header().Get("Content-Type"))
	}
	if !json.Valid(resp.Body.Bytes()) {
		t.Fatalf("expected valid json response, got=%s", resp.Body.String())
	}
}
