package case_library

import (
	"fmt"
	"strings"
)

const pendingReviewDuplicateThreshold = 0.9

type DuplicateHistoricalCaseError struct {
	TopMatch SimilarCaseResult
}

func (e *DuplicateHistoricalCaseError) Error() string {
	if e == nil {
		return "duplicate historical case detected"
	}
	return fmt.Sprintf("duplicate historical case detected: top1 similarity=%.4f case_id=%s title=%s",
		e.TopMatch.Similarity,
		strings.TrimSpace(e.TopMatch.CaseID),
		strings.TrimSpace(e.TopMatch.Title),
	)
}

func IsDuplicateHistoricalCaseError(err error) bool {
	_, ok := err.(*DuplicateHistoricalCaseError)
	return ok
}

func AsDuplicateHistoricalCaseError(err error) (*DuplicateHistoricalCaseError, bool) {
	duplicateErr, ok := err.(*DuplicateHistoricalCaseError)
	return duplicateErr, ok
}

var searchHistoricalCasesByVector = SearchTopKSimilarCasesByVector

func detectDuplicateHistoricalCase(queryVector []float64) error {
	results, _, err := searchHistoricalCasesByVector(queryVector, 1)
	if err != nil {
		return fmt.Errorf("compare with historical case library failed: %w", err)
	}
	if len(results) == 0 {
		return nil
	}

	top1 := results[0]
	if top1.Similarity >= pendingReviewDuplicateThreshold {
		return &DuplicateHistoricalCaseError{TopMatch: top1}
	}
	return nil
}
