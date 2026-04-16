package httpapi_test

import (
	"context"
	_ "unsafe"
)

//go:linkname rejectPendingReview antifraud/internal/modules/multi_agent/adapters/inbound/http.rejectPendingReview
var rejectPendingReview func(context.Context, string) error
