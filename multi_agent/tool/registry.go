package tool

import (
	"context"
	"encoding/json"
	"fmt"

	openai "antifraud/llm"
)

type ToolHandler interface {
	Handle(ctx context.Context, args string) (ToolResponse, error)
}

// ParseArgs 将 JSON 参数字符串解析为指定类型的输入结构体。
// T: 目标输入结构体类型。
// args: JSON 参数载荷。
func ParseArgs[T any](args string) (T, error) {
	var input T
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return input, fmt.Errorf("parse arguments failed: %v", err)
	}
	return input, nil
}

type ToolResponse struct {
	Payload        map[string]interface{}
	FinalResultStr string // 当工具输出为终态时，这里携带最终报告文本。
	ContextMutator func(context.Context) context.Context
}

var mainAgentToolRegistry = []openai.Tool{
	CaseSearchTool,
	QueryUserInfoTool,
	RiskAssessmentTool,
	DynamicRiskLevelTool,
	UpdateUserRecentTagsTool,
	SearchUserHistoryTool,
	UploadHistoricalCaseToVectorDBTool,
	WriteUserHistoryCaseTool,
	FinalReportTool,
	ExampleTool,
}

var mainAgentToolBlacklist = map[string]struct{}{
	AnalysisToolName: {},
}

var mainAgentToolHandlers = map[string]ToolHandler{
	CaseSearchToolName:                     &CaseSearchHandler{},
	QueryUserInfoToolName:                  &QueryUserInfoHandler{},
	RiskAssessmentToolName:                 &RiskAssessmentHandler{},
	DynamicRiskLevelToolName:               &DynamicRiskLevelHandler{},
	UpdateUserRecentTagsToolName:           &UpdateUserRecentTagsHandler{},
	SearchUserHistoryToolName:              &SearchUserHistoryHandler{},
	WriteUserHistoryCaseToolName:           &WriteUserHistoryCaseHandler{},
	FinalReportToolName:                    &FinalReportHandler{},
	ExampleToolName:                        &ExampleHandler{},
	UploadHistoricalCaseToVectorDBToolName: &UploadHistoricalCaseToVectorDBHandler{},
}

var caseCollectionToolRegistry = []openai.Tool{
	WebSearchTool,
	UploadHistoricalCaseToVectorDBTool,
}

var caseCollectionToolHandlers = map[string]ToolHandler{
	WebSearchToolName:                      &WebSearchHandler{},
	UploadHistoricalCaseToVectorDBToolName: &UploadHistoricalCaseToVectorDBHandler{},
}

// MainAgentTools 返回主智能体可用工具列表（已过滤黑名单）。
func MainAgentTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(mainAgentToolRegistry))
	for _, registeredTool := range mainAgentToolRegistry {
		if registeredTool.Function != nil {
			if _, blocked := mainAgentToolBlacklist[registeredTool.Function.Name]; blocked {
				continue
			}
		}
		tools = append(tools, registeredTool)
	}
	return tools
}

func GetToolHandler(name string) ToolHandler {
	return mainAgentToolHandlers[name]
}

// CaseCollectionTools 返回案件采集智能体可用工具列表。
func CaseCollectionTools() []openai.Tool {
	tools := make([]openai.Tool, 0, len(caseCollectionToolRegistry))
	tools = append(tools, caseCollectionToolRegistry...)
	return tools
}

// GetCaseCollectionToolHandler 返回案件采集智能体的工具处理器。
func GetCaseCollectionToolHandler(name string) ToolHandler {
	return caseCollectionToolHandlers[name]
}
