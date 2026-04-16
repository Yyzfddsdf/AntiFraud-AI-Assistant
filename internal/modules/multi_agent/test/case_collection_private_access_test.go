package multi_agent_test

import (
	agenttool "antifraud/internal/modules/multi_agent/adapters/outbound/tool"
	openai "antifraud/internal/platform/llm"
	_ "unsafe"
)

//go:linkname caseCollectionToolsProvider antifraud/internal/modules/multi_agent/core.caseCollectionToolsProvider
var caseCollectionToolsProvider func() []openai.Tool

//go:linkname caseCollectionToolHandlerResolver antifraud/internal/modules/multi_agent/core.caseCollectionToolHandlerResolver
var caseCollectionToolHandlerResolver func(string) agenttool.ToolHandler
