package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"antifraud/internal/platform/cache"
)

const (
	geoBoundaryDefaultCode       = "100000"
	geoBoundaryCacheKeyPrefix    = "cache:case_library:geo_boundary:v1:"
	geoBoundaryUpstreamURLFormat = "https://geo.datav.aliyun.com/areas_v3/bound/%s_full.json"
	geoBoundaryCacheTTL          = 6 * time.Hour
	geoBoundaryMaxPayloadBytes   = 8 << 20
)

var (
	geoBoundaryCodePattern = regexp.MustCompile(`^\d{6}$`)
	geoBoundaryHTTPClient  = &http.Client{Timeout: 12 * time.Second}
)

type geoBoundaryServiceError struct {
	statusCode int
	message    string
	cause      error
}

func (e *geoBoundaryServiceError) Error() string {
	if e == nil {
		return ""
	}
	if e.cause == nil {
		return e.message
	}
	return e.message + ": " + e.cause.Error()
}

func (e *geoBoundaryServiceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

func fetchGeoBoundaryGeoJSON(ctx context.Context, rawCode string) (json.RawMessage, error) {
	code, err := normalizeGeoBoundaryCode(rawCode)
	if err != nil {
		return nil, err
	}

	if cached, found := getCachedGeoBoundary(code); found {
		return cached, nil
	}

	payload, err := fetchGeoBoundaryFromUpstream(ctx, code)
	if err != nil {
		return nil, err
	}
	storeGeoBoundaryCache(code, payload)
	return payload, nil
}

func normalizeGeoBoundaryCode(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return geoBoundaryDefaultCode, nil
	}
	if !geoBoundaryCodePattern.MatchString(trimmed) {
		return "", &geoBoundaryServiceError{
			statusCode: http.StatusBadRequest,
			message:    "code 参数必须为 6 位行政区划编码",
		}
	}
	return trimmed, nil
}

func getCachedGeoBoundary(code string) (json.RawMessage, bool) {
	var payload json.RawMessage
	found, err := cache.GetJSON(buildGeoBoundaryCacheKey(code), &payload)
	if err != nil || !found || len(payload) == 0 || !json.Valid(payload) {
		return nil, false
	}
	return append(json.RawMessage(nil), payload...), true
}

func storeGeoBoundaryCache(code string, payload json.RawMessage) {
	if len(payload) == 0 || !json.Valid(payload) {
		return
	}
	_ = cache.SetJSON(buildGeoBoundaryCacheKey(code), payload, geoBoundaryCacheTTL)
}

func buildGeoBoundaryCacheKey(code string) string {
	return geoBoundaryCacheKeyPrefix + code
}

func buildGeoBoundaryUpstreamURL(code string) string {
	return fmt.Sprintf(geoBoundaryUpstreamURLFormat, code)
}

func fetchGeoBoundaryFromUpstream(ctx context.Context, code string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, buildGeoBoundaryUpstreamURL(code), nil)
	if err != nil {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusInternalServerError,
			message:    "地图边界请求构造失败",
			cause:      err,
		}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AntiFraud-AI-Assistant/geo-boundary")

	resp, err := geoBoundaryHTTPClient.Do(req)
	if err != nil {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusBadGateway,
			message:    "地图边界数据拉取失败",
			cause:      err,
		}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, geoBoundaryMaxPayloadBytes+1))
	if err != nil {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusBadGateway,
			message:    "地图边界数据读取失败",
			cause:      err,
		}
	}
	if len(body) > geoBoundaryMaxPayloadBytes {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusBadGateway,
			message:    "地图边界数据体积超出限制",
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusBadGateway,
			message:    fmt.Sprintf("地图边界服务返回异常状态: %d", resp.StatusCode),
		}
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return nil, &geoBoundaryServiceError{
			statusCode: http.StatusBadGateway,
			message:    "地图边界数据格式无效",
		}
	}
	return append(json.RawMessage(nil), trimmed...), nil
}
