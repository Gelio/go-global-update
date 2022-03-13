package gobinaries

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
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

type testDirectoryLister struct{ entries []string }

func (l *testDirectoryLister) ListDirectoryEntries(path string) ([]string, error) {
	return l.entries, nil
}

func TestExtractValidModuleURL(t *testing.T) {
	gobin := "~/go/bin"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{
			{
				args: []string{"version", "-m", "~/go/bin/shfmt"},
				output: `
shfmt: go1.17
        path    mvdan.cc/sh/v3/cmd/shfmt
        mod     mvdan.cc/sh/v3  v3.4.2  h1:d3TKODXfZ1bjWU/StENN+GDg5xOzNu5+C8AEu405E5U=
        dep     github.com/google/renameio      v1.0.1  h1:Lh/jXZmvZxb0BBeSY5VKEfidcbcbenKjZFzM/q0fSeU=
        dep     github.com/pkg/diff     v0.0.0-20210226163009-20ebb0f2a09e      h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=
        dep     golang.org/x/sys        v0.0.0-20210925032602-92d5a993a665      h1:QOQNt6vCjMpXE7JSK5VvAzJC1byuN3FgTNSBwf+CJgI=
        dep     golang.org/x/term       v0.0.0-20210916214954-140adaaadfaf      h1:Ihq/mm/suC88gF8WFcVwk+OV6Tq+wyA1O0E5UEvDglI=
        dep     mvdan.cc/editorconfig   v0.2.0  h1:XL+7ys6ls/RKrkUNFQvEwIvNHh+JKx8Mj1pUV5wQxQE=
`,
			},
			{
				args:   []string{"list", "-m", "-f", "{{.Version}}", "mvdan.cc/sh/v3/cmd/shfmt@latest"},
				output: "v3.4.2",
			},
		},
	}

	binariesFinder := NewIntrospecter(&cmdRunner, gobin, zap.NewNop())

	binary, err := binariesFinder.Introspect("shfmt")
	assert.Nil(t, err)
	assert.Equal(t, binary, GoBinary{
		Name:          "shfmt",
		PathURL:       "mvdan.cc/sh/v3/cmd/shfmt",
		ModuleURL:     "mvdan.cc/sh/v3",
		Path:          "~/go/bin/shfmt",
		Version:       "v3.4.2",
		LatestVersion: "v3.4.2",
	})
}

func TestExtractValidModuleURLFromGofumpt(t *testing.T) {
	gobin := "~/go/bin"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{
			{
				args: []string{"version", "-m", "~/go/bin/gofumpt"},
				output: `
    bin/gofumpt: go1.17
        path    mvdan.cc/gofumpt
        mod     mvdan.cc/gofumpt        v0.3.0  h1:kTojdZo9AcEYbQYhGuLf/zszYthRdhDNDUi2JKTxas4=
        dep     github.com/google/go-cmp        v0.5.7  h1:81/ik6ipDQS2aGcBfIN5dHDB36BwrStyeAQquSYCV4o=
        dep     golang.org/x/mod        v0.5.1  h1:OJxoQ/rynoF0dcCdI7cLPktw/hR2cueqYfjm43oqK38=
        dep     golang.org/x/sync       v0.0.0-20210220032951-036812b2e83c      h1:5KslGYwFpkhGh+Q16bwMP3cOontH8FOep7tGV86Y7SQ=
        dep     golang.org/x/sys        v0.0.0-20220209214540-3681064d5158      h1:rm+CHSpPEEW2IsXUib1ThaHIjuBVZjxNgSKmBLFfD4c=
        dep     golang.org/x/tools      v0.1.9  h1:j9KsMiaP1c3B0OTQGth0/k+miLGTgLsAFUCrF2vLcF8=
`,
			},
			{
				args:   []string{"list", "-m", "-f", "{{.Version}}", "mvdan.cc/gofumpt@latest"},
				output: "v0.3.0",
			},
		},
	}

	binariesFinder := NewIntrospecter(&cmdRunner, gobin, zap.NewNop())

	binary, err := binariesFinder.Introspect("gofumpt")
	assert.Nil(t, err)
	assert.Equal(t, binary, GoBinary{
		Name:          "gofumpt",
		ModuleURL:     "mvdan.cc/gofumpt",
		PathURL:       "mvdan.cc/gofumpt",
		Path:          "~/go/bin/gofumpt",
		Version:       "v0.3.0",
		LatestVersion: "v0.3.0",
	})
}

func TestMissingModuleURL(t *testing.T) {
	gobin := "~/go/bin"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{{args: []string{"version", "-m", "~/go/bin/shfmt"}, output: `
shfmt: go1.17
`}},
	}
	binariesFinder := NewIntrospecter(&cmdRunner, gobin, zap.NewNop())

	_, err := binariesFinder.Introspect("shfmt")
	assert.NotNil(t, err)
}

func TestExtractLatestVersion(t *testing.T) {
	gobin := "~/go/bin"
	cmdRunner := testGoCmdRunner{
		responses: []mockResponse{
			{
				args: []string{"version", "-m", "~/go/bin/gofumpt"},
				output: `
    bin/gofumpt: go1.17
        path    mvdan.cc/gofumpt
        mod     mvdan.cc/gofumpt        v0.3.0  h1:kTojdZo9AcEYbQYhGuLf/zszYthRdhDNDUi2JKTxas4=
        dep     github.com/google/go-cmp        v0.5.7  h1:81/ik6ipDQS2aGcBfIN5dHDB36BwrStyeAQquSYCV4o=
        dep     golang.org/x/mod        v0.5.1  h1:OJxoQ/rynoF0dcCdI7cLPktw/hR2cueqYfjm43oqK38=
        dep     golang.org/x/sync       v0.0.0-20210220032951-036812b2e83c      h1:5KslGYwFpkhGh+Q16bwMP3cOontH8FOep7tGV86Y7SQ=
        dep     golang.org/x/sys        v0.0.0-20220209214540-3681064d5158      h1:rm+CHSpPEEW2IsXUib1ThaHIjuBVZjxNgSKmBLFfD4c=
        dep     golang.org/x/tools      v0.1.9  h1:j9KsMiaP1c3B0OTQGth0/k+miLGTgLsAFUCrF2vLcF8=
`,
			},
			{
				args:   []string{"list", "-m", "-f", "{{.Version}}", "mvdan.cc/gofumpt@latest"},
				output: "v0.4.0",
			},
		},
	}

	binariesFinder := NewIntrospecter(&cmdRunner, gobin, zap.NewNop())
	binary, err := binariesFinder.Introspect("gofumpt")
	assert.Nil(t, err)
	assert.Equal(t, binary, GoBinary{
		Name:          "gofumpt",
		ModuleURL:     "mvdan.cc/gofumpt",
		PathURL:       "mvdan.cc/gofumpt",
		Path:          "~/go/bin/gofumpt",
		Version:       "v0.3.0",
		LatestVersion: "v0.4.0",
	})
	assert.True(t, binary.UpgradePossible())
}
