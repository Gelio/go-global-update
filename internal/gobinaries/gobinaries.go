package gobinaries

import "fmt"

type GoBinariesFinder interface {
	FindGoBinaries(gobin string) ([]string, error)
}

type RealGoBinariesFinder struct {
	introspecter    GoBinaryIntrospecter
	directoryLister DirectoryLister
}

func NewFinder(introspecter GoBinaryIntrospecter, directoryLister DirectoryLister) RealGoBinariesFinder {
	return RealGoBinariesFinder{
		introspecter,
		directoryLister,
	}
}

type IntrospectionResult struct {
	Binary GoBinary
	Error  error
}

func (f *RealGoBinariesFinder) FindGoBinaries(gobin string) ([]IntrospectionResult, error) {
	binaryNames, err := f.directoryLister.ListDirectoryEntries(gobin)
	if err != nil {
		return nil, fmt.Errorf("could not list entries in %s: %w", gobin, err)
	}

	var results []IntrospectionResult

	for _, binaryName := range binaryNames {
		binary, err := f.introspecter.Introspect(binaryName)
		if err != nil {
			err = fmt.Errorf("could not introspect binary %s: %w", binaryName, err)
		}

		results = append(results, IntrospectionResult{
			Binary: binary,
			Error:  err,
		})
	}

	return results, nil
}
