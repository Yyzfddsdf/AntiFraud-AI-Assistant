package tool_test

import (
	"context"

	case_library "antifraud/multi_agent/case_library"
	_ "unsafe"
)

//go:linkname noneFallback antifraud/multi_agent/tool.noneFallback
func noneFallback(text string) string

//go:linkname normalizeViolatedLaw antifraud/multi_agent/tool.normalizeViolatedLaw
func normalizeViolatedLaw(text string) string

//go:linkname createPendingReview antifraud/multi_agent/tool.createPendingReview
var createPendingReview func(context.Context, string, case_library.CreateHistoricalCaseInput) (case_library.PendingReviewRecord, error)
