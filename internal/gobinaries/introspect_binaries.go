package gobinaries

import (
	"fmt"
	"runtime"
	"sync"
)

type IntrospectionResult struct {
	Binary GoBinary
	Error  error
}

func IntrospectBinaries(introspecter *Introspecter, binaryNames []string) []IntrospectionResult {
	results := make([]IntrospectionResult, len(binaryNames))

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.GOMAXPROCS(0))
	for i, binaryName := range binaryNames {
		i, binaryName := i, binaryName

		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			binary, err := introspecter.Introspect(binaryName)
			if err != nil {
				err = fmt.Errorf("could not introspect binary %s: %w", binaryName, err)
			}

			results[i] = IntrospectionResult{
				Binary: binary,
				Error:  err,
			}
		}()
	}
	wg.Wait()

	return results
}
