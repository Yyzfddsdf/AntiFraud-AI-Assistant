package multi_agent_test

import (
	openai "antifraud/llm"
	agenttool "antifraud/multi_agent/tool"
	_ "unsafe"
)

//go:linkname caseCollectionToolsProvider antifraud/multi_agent.caseCollectionToolsProvider
var caseCollectionToolsProvider func() []openai.Tool

//go:linkname caseCollectionToolHandlerResolver antifraud/multi_agent.caseCollectionToolHandlerResolver
var caseCollectionToolHandlerResolver func(string) agenttool.ToolHandler
