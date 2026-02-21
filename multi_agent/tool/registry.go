package tool

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

type ToolHandler interface {
	Handle(ctx context.Context, args string) (ToolResponse, error)
}

type ToolResponse struct {
	Payload                   map[string]interface{}
	SetCaseSearch             bool                // 标记是否完成了案件检索
	SetFinalReport            bool                // 标记是否提交了最终报告
	FinalReportPayload        *FinalReportPayload // 最终报告payload
	SetHistoryWriteAfterFinal bool                // 标记是否在最终报告后写了历史
}

var mainAgentToolRegistry = []openai.Tool{
	CaseSearchTool,
	QueryUserHistoryCasesTool,
	QueryUserInfoTool,
	WriteUserHistoryCaseTool,
	FinalReportTool,
	ExampleTool,
}

var mainAgentToolBlacklist = map[string]struct{}{
	AnalysisToolName: {},
}

var mainAgentToolHandlers = map[string]ToolHandler{
	CaseSearchToolName:            &CaseSearchHandler{},
	QueryUserHistoryCasesToolName: &QueryUserHistoryCasesHandler{},
	QueryUserInfoToolName:         &QueryUserInfoHandler{},
	WriteUserHistoryCaseToolName:  &WriteUserHistoryCaseHandler{},
	FinalReportToolName:           &FinalReportHandler{},
	ExampleToolName:               &ExampleHandler{},
}

// MainAgentTools 自动挂载主智能体可用工具。
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
