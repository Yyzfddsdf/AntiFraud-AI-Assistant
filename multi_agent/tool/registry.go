package tool

import "github.com/sashabaranov/go-openai"

var mainAgentToolRegistry = []openai.Tool{
	CaseSearchTool,
	QueryUserHistoryCasesTool,
	QueryUserInfoTool,
	WriteUserHistoryCaseTool,
	FinalReportTool,
}

var mainAgentToolBlacklist = map[string]struct{}{
	AnalysisToolName: {},
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
