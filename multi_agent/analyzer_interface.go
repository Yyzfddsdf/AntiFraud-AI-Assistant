package multi_agent

import (
	"context"
	"fmt"
	"sync"
)

type Base64Analyzer interface {
	Analyze(ctx context.Context, dataBase64 string, index int) (string, error)
}

func analyzeBatchInParallel(ctx context.Context, analyzer Base64Analyzer, inputs []string, modality string) []string {
	var wg sync.WaitGroup
	results := make([]string, len(inputs))

	for i, item := range inputs {
		wg.Add(1)
		go func(index int, input string) {
			defer wg.Done()
			fmt.Printf("Starting analysis for %s %d...\n", modality, index+1)
			res, err := analyzer.Analyze(ctx, input, index)
			if err != nil {
				results[index] = fmt.Sprintf("Error: %v", err)
			} else {
				results[index] = res
			}
			fmt.Printf("Finished analysis for %s %d\n", modality, index+1)
		}(i, item)
	}

	wg.Wait()
	return results
}
