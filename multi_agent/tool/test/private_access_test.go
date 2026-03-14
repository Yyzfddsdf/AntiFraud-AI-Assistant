package tool_test

import _ "unsafe"

//go:linkname noneFallback antifraud/multi_agent/tool.noneFallback
func noneFallback(text string) string

//go:linkname normalizeViolatedLaw antifraud/multi_agent/tool.normalizeViolatedLaw
func normalizeViolatedLaw(text string) string
