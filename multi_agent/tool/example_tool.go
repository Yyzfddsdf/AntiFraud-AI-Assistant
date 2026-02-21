package tool

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

const ExampleToolName = "example_tool"

type ExampleInput struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

var ExampleTool = openai.Tool{
	Type: openai.ToolTypeFunction,
	Function: &openai.FunctionDefinition{
		Name:        ExampleToolName,
		Description: "示例工具：处理消息并返回重复结果",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "要处理的字符串消息",
				},
				"count": map[string]interface{}{
					"type":        "integer",
					"description": "重复次数",
				},
			},
			"required": []string{"message", "count"},
		},
	},
}

func ParseExampleInput(arguments string) (ExampleInput, error) {
	var input ExampleInput
	err := json.Unmarshal([]byte(arguments), &input)
	return input, err
}

func ExecuteExampleLogic(ctx context.Context, input ExampleInput) (map[string]interface{}, error) {
	userID := CurrentUserID(ctx)
	taskID := CurrentTaskID(ctx)

	result := ""
	for i := 0; i < input.Count; i++ {
		result += input.Message + " "
	}

	return map[string]interface{}{
		"user_id": userID,
		"task_id": taskID,
		"result":  result,
		"status":  "success",
	}, nil
}

type ExampleHandler struct{}

func (h *ExampleHandler) Handle(ctx context.Context, args string) (ToolResponse, error) {
	input, err := ParseExampleInput(args)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": fmt.Sprintf("invalid example args: %v", err)}}, nil
	}

	result, err := ExecuteExampleLogic(ctx, input)
	if err != nil {
		return ToolResponse{Payload: map[string]interface{}{"error": err.Error()}}, nil
	}

	return ToolResponse{Payload: result}, nil
}
