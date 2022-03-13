package gobinaries

import "fmt"

type IntrospectionResult struct {
	Binary GoBinary
	Error  error
}

func IntrospectBinaries(introspecter *Introspecter, binaryNames []string) []IntrospectionResult {
	var results []IntrospectionResult

	for _, binaryName := range binaryNames {
		binary, err := introspecter.Introspect(binaryName)
		if err != nil {
			err = fmt.Errorf("could not introspect binary %s: %w", binaryName, err)
		}

		results = append(results, IntrospectionResult{
			Binary: binary,
			Error:  err,
		})
	}

	return results
}
