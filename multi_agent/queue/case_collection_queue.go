package queue

import (
	"log"
	"strings"

	"antifraud/multi_agent"
)

// CaseCollectionEnqueueRequest 是案件采集后台任务的最小请求结构。
type CaseCollectionEnqueueRequest struct {
	Query     string
	CaseCount int
}

// EnqueueCaseCollectionTask 启动一个内存内后台 goroutine 执行案件采集。
func EnqueueCaseCollectionTask(userID string, request CaseCollectionEnqueueRequest) error {
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		trimmedUserID = "demo-user"
	}

	trimmedQuery := strings.TrimSpace(request.Query)
	go processCaseCollectionTask(trimmedUserID, trimmedQuery, request.CaseCount)
	return nil
}

func processCaseCollectionTask(userID string, query string, caseCount int) {
	defer func() {
		if recovered := recover(); recovered != nil {
			log.Printf("[case_collection_queue] panic recovered: user=%s query=%s err=%v", userID, query, recovered)
		}
	}()

	if err := multi_agent.CollectCasesForUser(userID, query, caseCount); err != nil {
		log.Printf("[case_collection_queue] case collection failed: user=%s query=%s count=%d err=%v",
			userID, strings.TrimSpace(query), caseCount, err)
		return
	}

	log.Printf("[case_collection_queue] case collection completed: user=%s query=%s count=%d",
		userID, strings.TrimSpace(query), caseCount)
}
