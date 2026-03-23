package httpapi

import (
	"net/http"
	"strings"

	apimodel "antifraud/internal/modules/multi_agent/adapters/inbound/http/models"
	"antifraud/internal/modules/multi_agent/application/queue"

	"github.com/gin-gonic/gin"
)

// CaseCollectionEnqueuer 定义案件采集处理器依赖的后台入队接口。
type CaseCollectionEnqueuer interface {
	EnqueueCaseCollectionTask(userID string, request queue.CaseCollectionEnqueueRequest) error
}

type caseCollectionEnqueuerFunc func(userID string, request queue.CaseCollectionEnqueueRequest) error

func (f caseCollectionEnqueuerFunc) EnqueueCaseCollectionTask(userID string, request queue.CaseCollectionEnqueueRequest) error {
	return f(userID, request)
}

// CollectCaseCollectionHandle 是默认案件采集处理器。
func CollectCaseCollectionHandle(c *gin.Context) {
	NewCollectCaseCollectionHandle(caseCollectionEnqueuerFunc(func(userID string, request queue.CaseCollectionEnqueueRequest) error {
		return queue.EnqueueCaseCollectionTask(userID, request)
	}))(c)
}

// NewCollectCaseCollectionHandle 创建可注入后台入队器的案件采集处理器。
func NewCollectCaseCollectionHandle(enqueuer CaseCollectionEnqueuer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload apimodel.CaseCollectionRequest
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
			return
		}

		if strings.TrimSpace(payload.Query) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query 不能为空"})
			return
		}
		if payload.CaseCount <= 0 || payload.CaseCount > 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "case_count 取值范围应为 1-20"})
			return
		}
		if enqueuer == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "案件采集入队器未初始化"})
			return
		}

		if err := enqueuer.EnqueueCaseCollectionTask(getCurrentUserID(c), queue.CaseCollectionEnqueueRequest{
			Query:     payload.Query,
			CaseCount: payload.CaseCount,
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "案件采集入队失败: " + err.Error()})
			return
		}

		c.JSON(http.StatusAccepted, gin.H{
			"message": "案件采集任务已在后台启动",
		})
	}
}
