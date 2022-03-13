package updater

import (
	"bytes"
	"fmt"
	"testing"

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

	err := UpdateBinaries(logger, options, &output, &cmdRunner, &lister, fsutils)

	assert.Nil(t, err)
	assert.Equal(t, `gofumpt (version: v0.3.0, can upgrade to v0.4.0)
shfmt (version: v3.4.2, can upgrade to v3.4.3)

Upgrading gofumpt to v0.4.0 ... ✅

Upgrading shfmt to v3.4.3 ... ✅

`, output.String())
}
