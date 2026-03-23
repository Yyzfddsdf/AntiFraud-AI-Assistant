package tool_test

import (
	"context"

	case_library "antifraud/internal/modules/multi_agent/adapters/outbound/case_library"
	_ "unsafe"
)

//go:linkname noneFallback antifraud/internal/modules/multi_agent/adapters/outbound/tool.noneFallback
func noneFallback(text string) string

//go:linkname normalizeViolatedLaw antifraud/internal/modules/multi_agent/adapters/outbound/tool.normalizeViolatedLaw
func normalizeViolatedLaw(text string) string

//go:linkname createPendingReview antifraud/internal/modules/multi_agent/adapters/outbound/tool.createPendingReview
var createPendingReview func(context.Context, string, case_library.CreateHistoricalCaseInput) (case_library.PendingReviewRecord, error)
