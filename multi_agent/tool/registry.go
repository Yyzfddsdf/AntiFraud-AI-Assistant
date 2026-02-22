package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

type ToolHandler interface {
	Handle(ctx context.Context, args string) (ToolResponse, error)
}

// ParseArgs 泛型参数解析函数
// T: 目标结构体类型
// args: JSON字符串
func ParseArgs[T any](args string) (T, error) {
	var input T
	if err := json.Unmarshal([]byte(args), &input); err != nil {
		return input, fmt.Errorf("参数解析失败: %v", err)
	}
	return input, nil
}

type ToolResponse struct {
	Payload        map[string]interface{}
	FinalResultStr string // 最终结果字符串 (当工具认为这就是最终答案时设置)
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
