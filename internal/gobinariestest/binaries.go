package gobinariestest

import (
	"path/filepath"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/goclitest"
)

var GOBIN = "/home/test/go/bin"

type MockBinary struct {
	Binary     gobinaries.GoBinary
	ModuleInfo string
}

func GetModuleInfoMockResponse(mockBinary MockBinary) goclitest.MockResponse {
	return goclitest.GetModuleInfoMockResponse(GOBIN, mockBinary.Binary.Name, mockBinary.ModuleInfo)
}

func GetLatestVersionMockResponse(binary gobinaries.GoBinary) goclitest.MockResponse {
	return goclitest.GetLatestVersionMockResponse(binary.ModuleURL, binary.LatestVersion)
}

func GetShfmtMockBinary() MockBinary {
	return MockBinary{
		Binary: gobinaries.GoBinary{
			Name:          "shfmt",
			PathURL:       "mvdan.cc/sh/v3/cmd/shfmt",
			ModuleURL:     "mvdan.cc/sh/v3",
			Path:          filepath.Join(GOBIN, "shfmt"),
			Version:       "v3.4.2",
			LatestVersion: "v3.4.2",
		},
		ModuleInfo: `
shfmt: go1.17
        path    mvdan.cc/sh/v3/cmd/shfmt
        mod     mvdan.cc/sh/v3  v3.4.2  h1:d3TKODXfZ1bjWU/StENN+GDg5xOzNu5+C8AEu405E5U=
        dep     github.com/google/renameio      v1.0.1  h1:Lh/jXZmvZxb0BBeSY5VKEfidcbcbenKjZFzM/q0fSeU=
        dep     github.com/pkg/diff     v0.0.0-20210226163009-20ebb0f2a09e      h1:aoZm08cpOy4WuID//EZDgcC4zIxODThtZNPirFr42+A=
        dep     golang.org/x/sys        v0.0.0-20210925032602-92d5a993a665      h1:QOQNt6vCjMpXE7JSK5VvAzJC1byuN3FgTNSBwf+CJgI=
        dep     golang.org/x/term       v0.0.0-20210916214954-140adaaadfaf      h1:Ihq/mm/suC88gF8WFcVwk+OV6Tq+wyA1O0E5UEvDglI=
        dep     mvdan.cc/editorconfig   v0.2.0  h1:XL+7ys6ls/RKrkUNFQvEwIvNHh+JKx8Mj1pUV5wQxQE=
`,
	}
}

func GetGofumptMockBinary() MockBinary {
	return MockBinary{
		Binary: gobinaries.GoBinary{
			Name:          "gofumpt",
			ModuleURL:     "mvdan.cc/gofumpt",
			PathURL:       "mvdan.cc/gofumpt",
			Path:          filepath.Join(GOBIN, "gofumpt"),
			Version:       "v0.3.0",
			LatestVersion: "v0.3.0",
		},
		ModuleInfo: `
gofumpt: go1.17
        path    mvdan.cc/gofumpt
        mod     mvdan.cc/gofumpt        v0.3.0  h1:kTojdZo9AcEYbQYhGuLf/zszYthRdhDNDUi2JKTxas4=
        dep     github.com/google/go-cmp        v0.5.7  h1:81/ik6ipDQS2aGcBfIN5dHDB36BwrStyeAQquSYCV4o=
        dep     golang.org/x/mod        v0.5.1  h1:OJxoQ/rynoF0dcCdI7cLPktw/hR2cueqYfjm43oqK38=
        dep     golang.org/x/sync       v0.0.0-20210220032951-036812b2e83c      h1:5KslGYwFpkhGh+Q16bwMP3cOontH8FOep7tGV86Y7SQ=
        dep     golang.org/x/sys        v0.0.0-20220209214540-3681064d5158      h1:rm+CHSpPEEW2IsXUib1ThaHIjuBVZjxNgSKmBLFfD4c=
        dep     golang.org/x/tools      v0.1.9  h1:j9KsMiaP1c3B0OTQGth0/k+miLGTgLsAFUCrF2vLcF8=
`,
	}
}
