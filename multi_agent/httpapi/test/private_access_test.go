package httpapi_test

import (
	"context"
	_ "unsafe"
)

//go:linkname rejectPendingReview antifraud/multi_agent/httpapi.rejectPendingReview
var rejectPendingReview func(context.Context, string) error
