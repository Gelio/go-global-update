package gobinaries_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gobinariestest"
	"github.com/Gelio/go-global-update/internal/goclitest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestExtractValidModuleURL(t *testing.T) {
	for _, mockBinary := range []gobinariestest.MockBinary{
		gobinariestest.GetGofumptMockBinary(),
		gobinariestest.GetShfmtMockBinary(),
	} {
		mockBinary := mockBinary
		t.Run(mockBinary.Binary.Name, func(t *testing.T) {
			cmdRunner := goclitest.TestGoCmdRunner{
				Responses: []goclitest.MockResponse{
					gobinariestest.GetModuleInfoMockResponse(mockBinary),
					gobinariestest.GetLatestVersionMockResponse(mockBinary.Binary),
				},
			}

			introspecter := gobinaries.NewIntrospecter(&cmdRunner, gobinariestest.GOBIN, zap.NewNop())

			binary, err := introspecter.Introspect(mockBinary.Binary.Name)
			assert.Nil(t, err)
			assert.Equal(t, mockBinary.Binary, binary)
		})
	}
}

func TestMissingModuleURL(t *testing.T) {
	gobin := "~/go/bin"
	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			goclitest.GetModuleInfoMockResponse(gobinariestest.GOBIN, "shfmt", `
shfmt: go1.17
`),
		},
	}
	introspecter := gobinaries.NewIntrospecter(&cmdRunner, gobin, zap.NewNop())

	_, err := introspecter.Introspect("shfmt")
	assert.NotNil(t, err)
}

func TestExtractLatestVersion(t *testing.T) {
	mockBinary := gobinariestest.GetGofumptMockBinary()
	mockBinary.Binary.LatestVersion = "v0.4.0"

	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinariestest.GetModuleInfoMockResponse(mockBinary),
			gobinariestest.GetLatestVersionMockResponse(mockBinary.Binary),
		},
	}

	introspecter := gobinaries.NewIntrospecter(&cmdRunner, gobinariestest.GOBIN, zap.NewNop())
	binary, err := introspecter.Introspect(mockBinary.Binary.Name)
	assert.Nil(t, err)
	assert.Equal(t, mockBinary.Binary, binary)
	assert.True(t, binary.UpgradePossible())
}

func TestBuiltFromSourceOnGo116And117(t *testing.T) {
	mockBinary := gobinariestest.MockBinary{
		Binary: gobinaries.GoBinary{
			Name:      "go-global-update",
			Path:      filepath.Join(gobinariestest.GOBIN, "go-global-update"),
			PathURL:   "command-line-arguments",
			ModuleURL: "github.com/Gelio/go-global-update",
			Version:   "(devel)",
		},
		ModuleInfo: `
go-global-update: go1.17
    path    command-line-arguments
    mod     github.com/Gelio/go-global-update       (devel)
    dep     github.com/cpuguy83/go-md2man/v2        v2.0.0-20190314233015-f79a8a8ca69d      h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
    dep     github.com/russross/blackfriday/v2      v2.0.1  h1:lPqVAte+HuHNfhJ/0LC98ESWRz8afy9tM/0RK8m9o+Q=
    dep     github.com/shurcooL/sanitized_anchor_name       v1.0.0  h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=
    dep     github.com/urfave/cli/v2        v2.3.0  h1:qph92Y649prgesehzOrQjdWyxFOp/QVM+6imKHad91M=
    dep     go.uber.org/atomic      v1.9.0  h1:ECmE8Bn/WFTYwEW/bpKD3M8VtR/zQVbavAoalC1PYyE=
    dep     go.uber.org/multierr    v1.8.0  h1:dg6GjLku4EH+249NNmoIciG9N/jURbDG+pFlTkhzIC8=
    dep     go.uber.org/zap v1.21.0 h1:WefMeulhovoZ2sYXz7st6K0sLj7bBhpiFaud4r4zST8=
    build   -compiler=gc
    build   CGO_ENABLED=1
    build   CGO_CFLAGS=
    build   CGO_CPPFLAGS=
    build   CGO_CXXFLAGS=
    build   CGO_LDFLAGS=
    build   GOARCH=amd64
    build   GOOS=linux
    build   GOAMD64=v1
`,
	}

	latestVersionMockResponse := goclitest.GetLatestVersionMockResponse(mockBinary.Binary.PathURL, "go: command-line-arguments@latest: malformed module path \"command-line-arguments\": missing dot in first path element")
	latestVersionMockResponse.Error = fmt.Errorf("exit code 1")

	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinariestest.GetModuleInfoMockResponse(mockBinary),
			latestVersionMockResponse,
		},
	}

	introspecter := gobinaries.NewIntrospecter(&cmdRunner, gobinariestest.GOBIN, zap.NewNop())
	binary, err := introspecter.Introspect(mockBinary.Binary.Name)
	assert.Nil(t, err)
	assert.Equal(t, mockBinary.Binary, binary)
	assert.True(t, binary.BuiltFromSource())
	assert.True(t, binary.BuiltWithGoBuild())
}

func TestMissingModLineOnGo118(t *testing.T) {
	mockBinary := gobinariestest.MockBinary{
		Binary: gobinaries.GoBinary{
			Name:    "go-global-update",
			Path:    filepath.Join(gobinariestest.GOBIN, "go-global-update"),
			PathURL: "command-line-arguments",
			Version: "(devel)",
		},
		ModuleInfo: `
go-global-update: go1.18
    path    command-line-arguments
    dep     github.com/Gelio/go-global-update       (devel)
    dep     github.com/cpuguy83/go-md2man/v2        v2.0.0-20190314233015-f79a8a8ca69d      h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
    dep     github.com/russross/blackfriday/v2      v2.0.1  h1:lPqVAte+HuHNfhJ/0LC98ESWRz8afy9tM/0RK8m9o+Q=
    dep     github.com/shurcooL/sanitized_anchor_name       v1.0.0  h1:PdmoCO6wvbs+7yrJyMORt4/BmY5IYyJwS/kOiWx8mHo=
    dep     github.com/urfave/cli/v2        v2.3.0  h1:qph92Y649prgesehzOrQjdWyxFOp/QVM+6imKHad91M=
    dep     go.uber.org/atomic      v1.9.0  h1:ECmE8Bn/WFTYwEW/bpKD3M8VtR/zQVbavAoalC1PYyE=
    dep     go.uber.org/multierr    v1.8.0  h1:dg6GjLku4EH+249NNmoIciG9N/jURbDG+pFlTkhzIC8=
    dep     go.uber.org/zap v1.21.0 h1:WefMeulhovoZ2sYXz7st6K0sLj7bBhpiFaud4r4zST8=
    build   -compiler=gc
    build   CGO_ENABLED=1
    build   CGO_CFLAGS=
    build   CGO_CPPFLAGS=
    build   CGO_CXXFLAGS=
    build   CGO_LDFLAGS=
    build   GOARCH=amd64
    build   GOOS=linux
    build   GOAMD64=v1
`,
	}

	latestVersionMockResponse := goclitest.GetLatestVersionMockResponse(mockBinary.Binary.PathURL, "go: command-line-arguments@latest: malformed module path \"command-line-arguments\": missing dot in first path element")
	latestVersionMockResponse.Error = fmt.Errorf("exit code 1")

	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinariestest.GetModuleInfoMockResponse(mockBinary),
			latestVersionMockResponse,
		},
	}

	introspecter := gobinaries.NewIntrospecter(&cmdRunner, gobinariestest.GOBIN, zap.NewNop())
	binary, err := introspecter.Introspect(mockBinary.Binary.Name)
	assert.Nil(t, err)
	assert.Equal(t, mockBinary.Binary, binary)
	assert.True(t, binary.BuiltFromSource())
	assert.True(t, binary.BuiltWithGoBuild())
}
