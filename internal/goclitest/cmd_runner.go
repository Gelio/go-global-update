package goclitest

import "fmt"

type MockResponse struct {
	Args   []string
	Output string
	Error  error
}

type TestGoCmdRunner struct {
	Responses []MockResponse
}

func (r *TestGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	for _, v := range r.Responses {
		if len(args) != len(v.Args) {
			continue
		}

		for i, arg := range args {
			if arg != v.Args[i] {
				continue
			}
		}

		return v.Output, v.Error
	}

	return "", fmt.Errorf("could not match args: %v", args)
}

type testDirectoryLister struct{ entries []string }

func (l *testDirectoryLister) ListDirectoryEntries(path string) ([]string, error) {
	return l.entries, nil
}
