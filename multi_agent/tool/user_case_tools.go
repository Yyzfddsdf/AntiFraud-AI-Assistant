package tool

import (
	"context"
	"fmt"
	"strings"
	"time"

	"antifraud/multi_agent/state"
	"antifraud/multi_agent/user_history_index"
	"antifraud/user_profile_system"

	openai "antifraud/llm"
)

const QueryUserInfoToolName = "query_user_info"
const UpdateUserRecentTagsToolName = "update_user_recent_tags"
const WriteUserHistoryCaseToolName = "write_user_history_case"
const SearchUserHistoryToolName = "search_user_history"

const defaultUserHistorySearchTopK = 5

type QueryUserInfoInput struct {
	Interval string `json:"interval,omitempty"`
}

type UpdateUserRecentTagsInput struct {
	RecentTags []string `json:"recent_tags"`
}

type WriteUserHistoryCaseInput struct {
	Title       string `json:"title"`
	CaseSummary string `json:"case_summary"`
	ScamType    string `json:"scam_type"`
	RiskLevel   string `json:"risk_level"`
}

type SearchUserHistoryInput struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k,omitempty"`
}

var QueryUserInfoTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        QueryUserInfoToolName,
		Description: "查询当前绑定用户的画像信息、近期标签与风险摘要。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"interval": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"day", "week", "month"},
					"description": "可选，风险趋势分析的时间粒度。允许：day/week/month，默认 day。",
				},
			},
		},
	},
}

var UpdateUserRecentTagsTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        UpdateUserRecentTagsToolName,
		Description: "更新当前绑定用户的近期标签。标签可为简短词语或句子，用于描述用户近期状态。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"recent_tags": map[string]interface{}{
					"type":        "array",
					"items":       map[string]string{"type": "string"},
					"description": "近期标签数组。支持词语或句子；会整体覆盖当前标签。",
				},
			},
			"required": []string{"recent_tags"},
		},
	},
}

var WriteUserHistoryCaseTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        WriteUserHistoryCaseToolName,
		Description: "将当前分析案件写入用户历史记录。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":        "string",
					"description": "案件标题。",
				},
				"case_summary": map[string]interface{}{
					"type":        "string",
					"description": "案件摘要。",
				},
				"scam_type": buildScamTypeSchema("诈骗类型。必须来自 config/scam_types.json 配置。"),
				"risk_level": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"低", "中", "高"},
					"description": "风险等级，仅允许：低/中/高。",
				},
			},
			"required": []string{"title", "case_summary", "scam_type", "risk_level"},
		},
	},
}

var SearchUserHistoryTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        SearchUserHistoryToolName,
		Description: "基于语义搜索当前用户的历史案件（向量化召回）。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "搜索关键词或案件描述（语义搜索）。",
				},
				"top_k": map[string]interface{}{
					"type":        "integer",
					"description": "返回结果数量，默认 5，最大 20。",
				},
			},
			"required": []string{"query"},
		},
	},
}

func ParseQueryUserInfoInput(arguments string) (QueryUserInfoInput, error) {
	return ParseArgs[QueryUserInfoInput](arguments)
}

func ParseUpdateUserRecentTagsInput(arguments string) (UpdateUserRecentTagsInput, error) {
	return ParseArgs[UpdateUserRecentTagsInput](arguments)
}

func ParseWriteUserHistoryCaseInput(arguments string) (WriteUserHistoryCaseInput, error) {
	return ParseArgs[WriteUserHistoryCaseInput](arguments)
}

func ParseSearchUserHistoryInput(arguments string) (SearchUserHistoryInput, error) {
	return ParseArgs[SearchUserHistoryInput](arguments)
}

func noneIfEmpty(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "none"
	}
	return trimmed
}

func QueryUserInfo(ctx context.Context, interval string) (map[string]interface{}, error) {
	info, err := user_profile_system.BuildUserRiskInfo(CurrentUserID(ctx), interval)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"user_name":            info.UserName,
		"age":                  info.Age,
		"occupation":           info.Occupation,
		"recent_tags":          info.RecentTags,
		"total_case_count":     info.TotalCaseCount,
		"historical_risk":      info.HistoricalRisk,
		"high_risk_case_ratio": info.HighRiskCaseRatio,
		"mid_risk_case_ratio":  info.MidRiskCaseRatio,
		"low_risk_case_ratio":  info.LowRiskCaseRatio,
		"risk_trend_analysis":  info.RiskTrendAnalysis,
	}, nil
}

// WriteUserHistoryCase 把当前任务归档到 history_cases。
// 归档数据来源：
// 1) 原始输入（text/videos/audios/images）来自 CurrentTaskPayload(ctx)
// 2) 子模态洞察来自 CurrentTaskInsights(ctx)
// 3) 最终报告来自 CurrentFinalReport(ctx)
func WriteUserHistoryCase(ctx context.Context, input WriteUserHistoryCaseInput) (map[string]interface{}, error) {
	normalizedScamType, scamTypeErr := normalizeAndValidateScamType(input.ScamType)
	if scamTypeErr != nil {
		return nil, fmt.Errorf("invalid scam_type: %v", scamTypeErr)
	}

	payload := CurrentTaskPayload(ctx)
	insights := CurrentTaskInsights(ctx)
	record := state.AddCaseHistory(CurrentUserID(ctx), CurrentTaskID(ctx), input.Title, input.CaseSummary, normalizedScamType, input.RiskLevel, state.TaskPayload{
		Text:          payload.Text,
		Videos:        append([]string{}, payload.Videos...),
		Audios:        append([]string{}, payload.Audios...),
		Images:        append([]string{}, payload.Images...),
		VideoInsights: append([]string{}, insights.VideoInsights...),
		AudioInsights: append([]string{}, insights.AudioInsights...),
		ImageInsights: append([]string{}, insights.ImageInsights...),
	}, CurrentFinalReport(ctx))

	result := map[string]interface{}{
		"status":       "success",
		"record_id":    record.RecordID,
		"user_id":      record.UserID,
		"message":      "history case persisted",
		"title":        record.Title,
		"created_at":   record.CreatedAt.Format(time.RFC3339),
		"case_summary": record.CaseSummary,
		"scam_type":    record.ScamType,
		"report":       record.Report,
		"stored_level": record.RiskLevel,
	}

	indexRecord, indexErr := user_history_index.UpsertHistoryVector(ctx, user_history_index.ArchiveInput{
		RecordID:    record.RecordID,
		UserID:      record.UserID,
		Title:       record.Title,
		CaseSummary: record.CaseSummary,
		ScamType:    record.ScamType,
		CreatedAt:   record.CreatedAt,
	})
	if indexErr != nil {
		result["message"] = "history case persisted, but vector index is unavailable"
		result["vector_index_status"] = "failed"
		result["vector_index_error"] = indexErr.Error()
		return result, nil
	}

	result["vector_index_status"] = "success"
	result["embedding_model"] = indexRecord.EmbeddingModel
	result["embedding_dimension"] = indexRecord.EmbeddingDimension
	result["vector_index_updated_at"] = indexRecord.UpdatedAt.Format(time.RFC3339)
	return result, nil
}

func SearchUserHistory(ctx context.Context, input SearchUserHistoryInput) ([]string, int, error) {
	trimmedQuery := strings.TrimSpace(input.Query)
	if trimmedQuery == "" {
		return nil, 0, fmt.Errorf("query is empty")
	}

	requestedTopK := input.TopK
	if requestedTopK <= 0 {
		requestedTopK = defaultUserHistorySearchTopK
	}

	results, appliedTopK, err := user_history_index.SearchTopKSimilarHistoryByQuery(ctx, CurrentUserID(ctx), trimmedQuery, requestedTopK)
	if err != nil {
		return nil, appliedTopK, err
	}
	if len(results) == 0 {
		return []string{}, appliedTopK, nil
	}

	historyRecords := state.GetCaseHistory(CurrentUserID(ctx))
	historyByRecordID := make(map[string]state.CaseHistoryRecord, len(historyRecords))
	for _, item := range historyRecords {
		trimmedRecordID := strings.TrimSpace(item.RecordID)
		if trimmedRecordID == "" {
			continue
		}
		historyByRecordID[trimmedRecordID] = item
	}

	formatted := make([]string, 0, len(results))
	for index, item := range results {
		record, ok := historyByRecordID[strings.TrimSpace(item.RecordID)]
		if !ok {
			continue
		}
		report := strings.TrimSpace(record.Report)
		if report == "" {
			report = "none"
		}
		formatted = append(formatted, fmt.Sprintf(
			"TOP%d | score:%.4f | %s | title: %s | summary: %s | scam_type: %s | risk: %s | report: %s",
			index+1,
			item.Similarity,
			record.CreatedAt.Format("2006-01-02 15:04:05"),
			noneIfEmpty(record.Title),
			noneIfEmpty(record.CaseSummary),
			noneIfEmpty(record.ScamType),
			noneIfEmpty(record.RiskLevel),
			report,
		))
	}
	return formatted, appliedTopK, nil
}

type QueryUserInfoHandler struct{}

func (h *QueryUserInfoHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseQueryUserInfoInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid query user info args: %v", err), "user": map[string]interface{}{"user_name": "demo-user"}}}, nil
	}
	userInfo, queryErr := QueryUserInfo(ctx, input.Interval)
	if queryErr != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": queryErr.Error(), "user": map[string]interface{}{"user_name": "user"}}}, nil
	}
	return ToolResponse{Payload: map[string]interface{}{"user": userInfo}}, nil
}

type UpdateUserRecentTagsHandler struct{}

func (h *UpdateUserRecentTagsHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseUpdateUserRecentTagsInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("invalid update user recent tags args: %v", err),
		}}, nil
	}

	recentTags, updateErr := user_profile_system.UpdateRecentTagsByStringUserID(CurrentUserID(ctx), input.RecentTags)
	if updateErr != nil {
		return ToolResponse{Payload: map[string]interface{}{
			"status": "failed",
			"error":  updateErr.Error(),
		}}, nil
	}

	return ToolResponse{Payload: map[string]interface{}{
		"status":      "success",
		"user_id":     CurrentUserID(ctx),
		"recent_tags": recentTags,
		"message":     "user recent tags updated",
	}}, nil
}

type WriteUserHistoryCaseHandler struct{}

func (h *WriteUserHistoryCaseHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseWriteUserHistoryCaseInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid write user history case args: %v", err), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "invalid input"}}}, nil
	}
	payload, writeErr := WriteUserHistoryCase(ctx, input)
	if writeErr != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": writeErr.Error(), "status": "failed", "record": map[string]interface{}{"record_id": "CASE-WRITE-0001", "message": "persist failed"}}}, nil
	}
	payload["system_instruction"] = "CRITICAL: Case archiving is the FINAL step. All tasks are completed. You MUST STOP calling any tools now and end the conversation immediately."
	return ToolResponse{Payload: payload}, nil
}

type SearchUserHistoryHandler struct{}

func (h *SearchUserHistoryHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseSearchUserHistoryInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid search user history args: %v", err), "status": "failed", "cases": []string{}}}, nil
	}

	results, appliedTopK, searchErr := SearchUserHistory(ctx, input)
	if searchErr != nil {
		boundUserID := CurrentUserID(ctx)
		return ToolResponse{Payload: map[string]interface{}{
			"status":        "failed",
			"user_id":       boundUserID,
			"query":         strings.TrimSpace(input.Query),
			"applied_top_k": appliedTopK,
			"error":         searchErr.Error(),
			"cases":         []string{},
		}}, nil
	}

	boundUserID := CurrentUserID(ctx)
	return ToolResponse{
		Payload: map[string]interface{}{
			"status":        "success",
			"user_id":       boundUserID,
			"query":         strings.TrimSpace(input.Query),
			"applied_top_k": appliedTopK,
			"cases":         results,
		},
	}, nil
}
