package updater

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gelio/go-global-update/internal/colors"
	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gobinariestest"
	"github.com/Gelio/go-global-update/internal/goclitest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockFilesystemUtils struct{}

func (_ mockFilesystemUtils) Chdir(_ string) error {
	return nil
}

func updateMockResponse(binary gobinaries.GoBinary, output string, err error) goclitest.MockResponse {
	return goclitest.MockResponse{
		Args:   []string{"install", fmt.Sprintf("%s@latest", binary.PathURL)},
		Output: output,
		Error:  err,
	}
}

func gobinMockResponse() goclitest.MockResponse {
	return goclitest.MockResponse{
		Args:   []string{"env", "GOBIN"},
		Output: gobinariestest.GOBIN,
	}
}

func TestUpdateAllFoundBinaries(t *testing.T) {
	gofumptMockBinary := gobinariestest.GetGofumptMockBinary()
	gofumptMockBinary.Binary.LatestVersion = "v0.4.0"
	shfmtMockBinary := gobinariestest.GetShfmtMockBinary()
	shfmtMockBinary.Binary.LatestVersion = "v3.4.3"

	logger := zap.NewNop()
	options := Options{}
	var output bytes.Buffer
	lister := gobinariestest.TestSuccessDirectoryLister{
		Entries: []string{gofumptMockBinary.Binary.Name, shfmtMockBinary.Binary.Name},
	}
	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinMockResponse(),
			gobinariestest.GetLatestVersionMockResponse(gofumptMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(gofumptMockBinary),
			updateMockResponse(gofumptMockBinary.Binary, "", nil),
			gobinariestest.GetLatestVersionMockResponse(shfmtMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(shfmtMockBinary),
			updateMockResponse(shfmtMockBinary.Binary, "", nil),
		},
	}

	fsutils := mockFilesystemUtils{}
	colorsFactory := colors.NewFactory(false)

	err := UpdateBinaries(logger, options, &output, &colorsFactory, &cmdRunner, &lister, fsutils)

	assert.Nil(t, err)
	assert.Equal(t, strings.TrimSpace(`
Binary       Current version      Status
gofumpt      v0.3.0               can upgrade to v0.4.0
shfmt        v3.4.2               can upgrade to v3.4.3

Upgrading gofumpt to v0.4.0 ... ✅

Upgrading shfmt to v3.4.3 ... ✅

`), strings.TrimSpace(output.String()))
}

func TestSkipUpgradingBuiltFromSource(t *testing.T) {
	builtFromSourceMockBinary := gobinariestest.MockBinary{
		Binary: gobinaries.GoBinary{
			Name:          "built-from-source",
			ModuleURL:     "github.com/Gelio/built-from-source",
			PathURL:       "command-line-arguments",
			Path:          filepath.Join(gobinariestest.GOBIN, "built-from-source"),
			Version:       "(devel)",
			LatestVersion: "v0.1.0",
		},
		ModuleInfo: `
built-from-source: go1.17
    path    command-line-arguments
    mod     github.com/Gelio/built-from-source      (devel)
    dep     github.com/cpuguy83/go-md2man/v2        v2.0.0-20190314233015-f79a8a8ca69d      h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
	}
	installedFromSourceMockBinary := gobinariestest.MockBinary{
		Binary: gobinaries.GoBinary{
			Name:          "installed-from-source",
			ModuleURL:     "github.com/Gelio/installed-from-source",
			PathURL:       "github.com/Gelio/installed-from-source",
			Path:          filepath.Join(gobinariestest.GOBIN, "installed-from-source"),
			Version:       "(devel)",
			LatestVersion: "v0.1.0",
		},
		ModuleInfo: `
installed-from-source: go1.17
    path    github.com/Gelio/built-from-source
    mod     github.com/Gelio/built-from-source      (devel)
    dep     github.com/cpuguy83/go-md2man/v2        v2.0.0-20190314233015-f79a8a8ca69d      h1:U+s90UTSYgptZMwQh2aRr3LuazLJIa+Pg3Kc1ylSYVY=
`,
	}

	logger := zap.NewNop()
	options := Options{}
	var output bytes.Buffer
	lister := gobinariestest.TestSuccessDirectoryLister{
		Entries: []string{builtFromSourceMockBinary.Binary.Name, installedFromSourceMockBinary.Binary.Name},
	}
	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinMockResponse(),
			gobinariestest.GetLatestVersionMockResponse(builtFromSourceMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(builtFromSourceMockBinary),
			updateMockResponse(builtFromSourceMockBinary.Binary, "", nil),
			gobinariestest.GetLatestVersionMockResponse(installedFromSourceMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(installedFromSourceMockBinary),
			updateMockResponse(installedFromSourceMockBinary.Binary, "", nil),
		},
	}

	fsutils := mockFilesystemUtils{}
	colorsFactory := colors.NewFactory(false)

	err := UpdateBinaries(logger, options, &output, &colorsFactory, &cmdRunner, &lister, fsutils)

	assert.Nil(t, err)
	assert.Equal(t, strings.TrimSpace(`
Binary                     Current version      Status
built-from-source          (devel)              cannot upgrade
installed-from-source      (devel)              can upgrade to v0.1.0

Skipping upgrading built-from-source
    The binary was built from source (probably using "go build") and the binary path is unknown.
    Install the binary using "go install repositoryPath@latest" instead.
    This seems like a known problem E001.
    See https://github.com/Gelio/go-global-update/blob/main/TROUBLESHOOTING.md#e001---binaries-built-from-source
    for more information.

Skipping upgrading installed-from-source
    The binary was installed from source (probably using "go install" in the cloned repository).
    Install the binary using "go install github.com/Gelio/built-from-source@latest" instead.
    This seems like a known problem E001.
    See https://github.com/Gelio/go-global-update/blob/main/TROUBLESHOOTING.md#e001---binaries-built-from-source
    for more information.

  `), strings.TrimSpace(output.String()))
}

func TestForceReinstallingBinaries(t *testing.T) {
	gofumptMockBinary := gobinariestest.GetGofumptMockBinary()
	shfmtMockBinary := gobinariestest.GetShfmtMockBinary()
	shfmtMockBinary.Binary.LatestVersion = "v3.4.3"

	logger := zap.NewNop()
	options := Options{
		ForceReinstall: true,
	}
	var output bytes.Buffer
	lister := gobinariestest.TestSuccessDirectoryLister{
		Entries: []string{gofumptMockBinary.Binary.Name, shfmtMockBinary.Binary.Name},
	}
	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinMockResponse(),
			gobinariestest.GetLatestVersionMockResponse(gofumptMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(gofumptMockBinary),
			updateMockResponse(gofumptMockBinary.Binary, "", nil),
			gobinariestest.GetLatestVersionMockResponse(shfmtMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(shfmtMockBinary),
			updateMockResponse(shfmtMockBinary.Binary, "", nil),
		},
	}

	fsutils := mockFilesystemUtils{}
	colorsFactory := colors.NewFactory(false)

	err := UpdateBinaries(logger, options, &output, &colorsFactory, &cmdRunner, &lister, fsutils)

	assert.Nil(t, err)
	assert.Equal(t, strings.TrimSpace(`
Binary       Current version      Status
gofumpt      v0.3.0               up-to-date
shfmt        v3.4.2               can upgrade to v3.4.3

Force-reinstalling gofumpt v0.3.0 ... ✅

Upgrading shfmt to v3.4.3 ... ✅

`), strings.TrimSpace(output.String()))
}

func TestPreserveBuildTags(t *testing.T) {
	shfmtMockBinary := gobinariestest.GetShfmtMockBinary()
	shfmtMockBinary.Binary.BuildTags = []string{"a", "b", "c"}
	shfmtMockBinary.Binary.LatestVersion = "v3.4.3"
	shfmtMockBinary.ModuleInfo = fmt.Sprintf(`%s
  build  -tags=a,b,c`, shfmtMockBinary.ModuleInfo)

	logger := zap.NewNop()
	options := Options{
		ForceReinstall: true,
	}
	var output bytes.Buffer
	lister := gobinariestest.TestSuccessDirectoryLister{
		Entries: []string{shfmtMockBinary.Binary.Name},
	}
	cmdRunner := goclitest.TestGoCmdRunner{
		Responses: []goclitest.MockResponse{
			gobinMockResponse(),
			gobinariestest.GetLatestVersionMockResponse(shfmtMockBinary.Binary),
			gobinariestest.GetModuleInfoMockResponse(shfmtMockBinary),
			{
				Args:   []string{"install", "-tags", "a,b,c", fmt.Sprintf("%s@latest", shfmtMockBinary.Binary.PathURL)},
				Output: "",
				Error:  nil,
			},
		},
	}

	fsutils := mockFilesystemUtils{}
	colorsFactory := colors.NewFactory(false)

	err := UpdateBinaries(logger, options, &output, &colorsFactory, &cmdRunner, &lister, fsutils)

	assert.Nil(t, err)
	assert.Equal(t, strings.TrimSpace(`
Binary      Current version      Status
shfmt       v3.4.2               can upgrade to v3.4.3

Upgrading shfmt to v3.4.3 (build tags: a,b,c) ... ✅

`), strings.TrimSpace(output.String()))
}
