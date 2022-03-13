package goclitest

import (
	"fmt"
	"path"
)

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

		match := true
		for i, arg := range args {
			if arg != v.Args[i] {
				match = false
				break
			}
		}
		if match {
			return v.Output, v.Error
		}
	}

	return "", fmt.Errorf("could not match args: %v", args)
}

func GetLatestVersionMockResponse(pathURL, version string) MockResponse {
	return MockResponse{
		Args:   []string{"list", "-m", "-f", "{{.Version}}", fmt.Sprintf("%s@latest", pathURL)},
		Output: version,
	}
}

func GetModuleInfoMockResponse(gobin, binaryName, output string) MockResponse {
	return MockResponse{
		Args:   []string{"version", "-m", path.Join(gobin, binaryName)},
		Output: output,
	}
}
