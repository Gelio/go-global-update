package gobinaries

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Gelio/go-global-update/internal/gocli"
	"go.uber.org/zap"
)

type Introspecter struct {
	cmdRunner gocli.GoCmdRunner
	gobin     string
	logger    *zap.Logger
}

func NewIntrospecter(cmdRunner gocli.GoCmdRunner, gobin string, logger *zap.Logger) Introspecter {
	return Introspecter{
		cmdRunner,
		gobin,
		logger,
	}
}

func (i *Introspecter) Introspect(binaryName string) (GoBinary, error) {
	binaryPath := filepath.Join(i.gobin, binaryName)
	moduleInfo, err := i.getModuleInfo(binaryPath)
	if err != nil {
		return GoBinary{}, fmt.Errorf("could not get module info about %v: %w", binaryPath, err)
	}

	latestVersion := ""
	// NOTE: module URL may be missing on go 1.18 for binaries built using `go build`
	// In case the package is built from source (path is
	// "command-line-arguments"), behave consistently on all go versions
	if moduleInfo.moduleURL != "" && moduleInfo.pathURL != "command-line-arguments" {
		latestVersion, err = i.getLatestModuleVersion(moduleInfo.moduleURL)
		if err != nil {
			return GoBinary{}, fmt.Errorf("could not get latest version of %v: %w", moduleInfo.moduleURL, err)
		}
	}

	goBinary := GoBinary{
		ModuleURL:     moduleInfo.moduleURL,
		PathURL:       moduleInfo.pathURL,
		Version:       moduleInfo.version,
		Name:          binaryName,
		Path:          binaryPath,
		LatestVersion: latestVersion,
		BuildTags:     moduleInfo.buildTags,
	}
	i.logger.Sugar().Debugf("introspected binary %s: %+v", binaryName, goBinary)

	return goBinary, nil
}

type parsedGoModuleInfo struct {
	moduleURL string
	pathURL   string
	version   string
	buildTags []string
}

func (i *Introspecter) getModuleInfo(binaryPath string) (*parsedGoModuleInfo, error) {
	moduleOutput, err := i.cmdRunner.RunGoCommand("version", "-m", binaryPath)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve version information about binary %s: %w\n%v", binaryPath, err, moduleOutput)
	}
	goModuleInfo := findModuleURLInModuleOutput(moduleOutput)
	if goModuleInfo == nil {
		return nil, fmt.Errorf("could not parse module information for binary %s", binaryPath)
	}
	goModuleInfo.buildTags = findBuildTagsInModuleOutput(moduleOutput)
	if len(goModuleInfo.buildTags) > 0 {
		i.logger.Sugar().Debugf("found build tags for binary %s: %v", binaryPath, goModuleInfo.buildTags)
	} else {
		i.logger.Sugar().Debugf("no build tags found for binary %s", binaryPath)
	}

	return goModuleInfo, nil
}

func (i *Introspecter) getLatestModuleVersion(moduleURL string) (string, error) {
	latestVersionModule := fmt.Sprintf("%s@latest", moduleURL)
	return i.cmdRunner.RunGoCommand("list", "-m", "-f", "{{.Version}}", latestVersionModule)
}

func findModuleURLInModuleOutput(output string) *parsedGoModuleInfo {
	r := regexp.MustCompile(`\s(path|mod)\s+([^\s]+)(\s+([^\s]+))?`)

	goModuleInfo := parsedGoModuleInfo{}

	matchedPath, matchedMod := false, false

	for _, l := range strings.Split(output, "\n") {
		matches := r.FindStringSubmatch(l)
		if len(matches) == 0 {
			continue
		}

		switch matches[1] {
		case "path":
			goModuleInfo.pathURL = matches[2]
			matchedPath = true
		case "mod":
			goModuleInfo.moduleURL = matches[2]
			goModuleInfo.version = matches[4]
			matchedMod = true
		default:
			panic(fmt.Sprintf("unknown match %s", matches[0]))
		}
	}

	if !matchedPath {
		return nil
	}

	if !matchedMod {
		// NOTE: Binaries built from source (using `go build`) do not contain a
		// `mod` entry on go 1.18.
		// If that happens, we behave as on older go versions (the version is
		// "(devel)")
		goModuleInfo.version = "(devel)"
	}

	return &goModuleInfo
}

func findBuildTagsInModuleOutput(output string) []string {
	r := regexp.MustCompile((`^\s+build\s+-tags=(([^,]+)(,[^,]+)*)$`))

	var buildTags []string

	for _, l := range strings.Split(output, "\n") {
		matches := r.FindStringSubmatch(l)
		if len(matches) == 0 {
			continue
		}

		buildTags = strings.Split(matches[1], ",")
		break
	}

	return buildTags
}
