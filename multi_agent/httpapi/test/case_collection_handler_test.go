package httpapi_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpapi "antifraud/multi_agent/httpapi"
	apimodel "antifraud/multi_agent/httpapi/models"
	"antifraud/multi_agent/queue"

	"github.com/gin-gonic/gin"
)

type stubCaseCollectionEnqueuer struct {
	lastUserID string
	lastReq    queue.CaseCollectionEnqueueRequest
	err        error
}

func (s *stubCaseCollectionEnqueuer) EnqueueCaseCollectionTask(userID string, request queue.CaseCollectionEnqueueRequest) error {
	s.lastUserID = userID
	s.lastReq = request
	return s.err
}

func TestNewCollectCaseCollectionHandleSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	enqueuer := &stubCaseCollectionEnqueuer{}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("userID", "u-test")
		c.Next()
	})
	router.POST("/search", httpapi.NewCollectCaseCollectionHandle(enqueuer))

	body, _ := json.Marshal(apimodel.CaseCollectionRequest{
		Query:     "冒充客服诈骗",
		CaseCount: 2,
	})
	req := httptest.NewRequest(http.MethodPost, "/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
	if enqueuer.lastUserID != "u-test" {
		t.Fatalf("unexpected user id: %q", enqueuer.lastUserID)
	}
	if enqueuer.lastReq.Query != "冒充客服诈骗" || enqueuer.lastReq.CaseCount != 2 {
		t.Fatalf("unexpected enqueue request: %+v", enqueuer.lastReq)
	}

	payload := map[string]interface{}{}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload["message"] != "案件采集任务已在后台启动" {
		t.Fatalf("unexpected response payload: %+v", payload)
	}
}

func TestNewCollectCaseCollectionHandleBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/search", httpapi.NewCollectCaseCollectionHandle(&stubCaseCollectionEnqueuer{}))

	body, _ := json.Marshal(apimodel.CaseCollectionRequest{
		Query:     "冒充客服诈骗",
		CaseCount: 0,
	})
	req := httptest.NewRequest(http.MethodPost, "/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestNewCollectCaseCollectionHandleEnqueueError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/search", httpapi.NewCollectCaseCollectionHandle(&stubCaseCollectionEnqueuer{
		err: errors.New("enqueue failed"),
	}))

	body, _ := json.Marshal(apimodel.CaseCollectionRequest{
		Query:     "冒充客服诈骗",
		CaseCount: 1,
	})
	req := httptest.NewRequest(http.MethodPost, "/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected status: got=%d body=%s", resp.Code, resp.Body.String())
	}
}
