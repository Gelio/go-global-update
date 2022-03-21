package gobinaries

type GoBinary struct {
	// ModuleURL is the `mod` URL from `go version -m`
	ModuleURL string
	// Path is the `path` URL from `go version -m`
	PathURL string
	Name    string
	// Path is the filesystem path to this binary.
	Path          string
	Version       string
	LatestVersion string
}

func (b *GoBinary) UpgradePossible() bool {
	return b.Version != b.LatestVersion
}

// BuiltFromSource determines whether the binary was built or installed from source.
func (b *GoBinary) BuiltFromSource() bool {
	return b.Version == "(devel)"
}

func (b *GoBinary) BuiltWithGoBuild() bool {
	// Binaries built with `go build main.go` have a very distinct path.
	// See https://github.com/Gelio/go-global-update/issues/3#issuecomment-1071566068
	return b.PathURL == "command-line-arguments"
}
