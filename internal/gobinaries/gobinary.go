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
