package gocli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockResponse struct {
	args   []string
	output string
	error  error
}

type testGoCmdRunner struct {
	responses []mockResponse
}

func (r *testGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	for _, v := range r.responses {
		if len(args) != len(v.args) {
			continue
		}

		for i, arg := range args {
			if arg != v.args[i] {
				continue
			}
		}

		return v.output, v.error
	}

	return "", errors.New(fmt.Sprintf("could not match args: %v", args))
}

func TestExtractValidModuleURL(t *testing.T) {
	binaryName := "go-test-binary"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{{args: []string{"version", "-m", binaryName}, output: `
shfmt: go1.17
        path    mvdan.cc/sh/v3/cmd/shfmt
        mod     mvdan.cc/sh/v3  v3.4.2  h1:d3TKODXfZ1bjWU/StENN+GDg5xOzNu5+C8AEu405E5U=
        dep     github.com/google/renameio      v1.0.1  h1:Lh/jXZmvZxb0BBeSY5VKEfidcbcbenKjZFzM/q0fSeU=
        dep     github.com/pkg/diff     v0.0.0-20210226163009-20ebb0f2a09e      h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=
        dep     golang.org/x/sys        v0.0.0-20210925032602-92d5a993a665      h1:QOQNt6vCjMpXE7JSK5VvAzJC1byuN3FgTNSBwf+CJgI=
        dep     golang.org/x/term       v0.0.0-20210916214954-140adaaadfaf      h1:Ihq/mm/suC88gF8WFcVwk+OV6Tq+wyA1O0E5UEvDglI=
        dep     mvdan.cc/editorconfig   v0.2.0  h1:XL+7ys6ls/RKrkUNFQvEwIvNHh+JKx8Mj1pUV5wQxQE=
`}},
	}
	cli := New(&cmdRunner)

	moduleURL, err := cli.GetModuleURL(binaryName)
	assert.Nil(t, err)
	assert.Equal(t, moduleURL, "mvdan.cc/sh/v3/cmd/shfmt")
}

func TestMissingModuleURL(t *testing.T) {
	binaryName := "go-test-binary"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{{args: []string{"version", "-m", binaryName}, output: `
shfmt: go1.17
`}},
	}
	cli := New(&cmdRunner)

	_, err := cli.GetModuleURL(binaryName)
	assert.NotNil(t, err)
}
