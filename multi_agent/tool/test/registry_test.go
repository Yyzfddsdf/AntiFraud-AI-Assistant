package tool_test

import (
	"testing"

	agenttool "antifraud/multi_agent/tool"
)

func TestMainAgentToolsExcludeLegacyHistoryQuery(t *testing.T) {
	for _, tool := range agenttool.MainAgentTools() {
		if tool.Function != nil && tool.Function.Name == "query_user_history_cases" {
			t.Fatalf("expected legacy history query tool to stay out of the main registry")
		}
	}

	if handler := agenttool.GetToolHandler("query_user_history_cases"); handler != nil {
		t.Fatalf("expected legacy history query handler to stay unregistered, got %#v", handler)
	}
}
