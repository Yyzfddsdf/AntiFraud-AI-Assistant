package case_library

import (
	"context"
	"fmt"
	"strings"

	"antifraud/embedding"
)

type preparedHistoricalCaseInput struct {
	normalizedInput CreateHistoricalCaseInput
	vector          []float64
	modelName       string
}

var generateCaseEmbedding = embedding.GenerateVector

func prepareHistoricalCaseInput(ctx context.Context, input CreateHistoricalCaseInput) (preparedHistoricalCaseInput, error) {
	normalizedInput, err := normalizeAndValidateInput(input)
	if err != nil {
		return preparedHistoricalCaseInput{}, err
	}

	embeddingText := BuildEmbeddingInput(normalizedInput)
	vector, modelName, err := generateCaseEmbedding(ctx, embeddingText)
	if err != nil {
		return preparedHistoricalCaseInput{}, fmt.Errorf("generate embedding failed: %w", err)
	}

	return preparedHistoricalCaseInput{
		normalizedInput: normalizedInput,
		vector:          append([]float64{}, vector...),
		modelName:       strings.TrimSpace(modelName),
	}, nil
}
