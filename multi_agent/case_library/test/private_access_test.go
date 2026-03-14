package case_library_test

import (
	"context"

	case_library "antifraud/multi_agent/case_library"
	_ "unsafe"
)

//go:linkname generateCaseEmbedding antifraud/multi_agent/case_library.generateCaseEmbedding
var generateCaseEmbedding func(context.Context, string) ([]float64, string, error)

//go:linkname searchHistoricalCasesByVector antifraud/multi_agent/case_library.searchHistoricalCasesByVector
var searchHistoricalCasesByVector func([]float64, int) ([]case_library.SimilarCaseResult, int, error)
