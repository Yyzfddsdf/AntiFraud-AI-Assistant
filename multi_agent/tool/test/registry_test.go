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

func TestCaseCollectionToolsIncludeWebSearch(t *testing.T) {
	found := false
	for _, tool := range agenttool.CaseCollectionTools() {
		if tool.Function != nil && tool.Function.Name == agenttool.WebSearchToolName {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected %s to be registered for case collection tools", agenttool.WebSearchToolName)
	}

	if handler := agenttool.GetCaseCollectionToolHandler(agenttool.WebSearchToolName); handler == nil {
		t.Fatalf("expected %s handler to be registered for case collection tools", agenttool.WebSearchToolName)
	}
}
