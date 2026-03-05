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
}

var mainAgentToolRegistry = []openai.Tool{
	CaseSearchTool,
	QueryUserHistoryCasesTool,
	QueryUserInfoTool,
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
	QueryUserHistoryCasesToolName:          &QueryUserHistoryCasesHandler{},
	QueryUserInfoToolName:                  &QueryUserInfoHandler{},
	WriteUserHistoryCaseToolName:           &WriteUserHistoryCaseHandler{},
	FinalReportToolName:                    &FinalReportHandler{},
	ExampleToolName:                        &ExampleHandler{},
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
