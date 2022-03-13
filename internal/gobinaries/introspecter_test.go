package gobinaries_test

import (
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
