package gobinaries

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/Gelio/go-global-update/internal/gocli"
)

type GoBinary struct {
	ModuleURL     string
	Name          string
	AbsPath       string
	Version       string
	LatestVersion string
}

func (b *GoBinary) UpgradePossible() bool {
	return b.Version != b.LatestVersion
}

type GoBinariesFinder interface {
	FindGoBinaries(gobin string) ([]string, error)
}

type RealGoBinariesFinder struct {
	cmdRunner       gocli.GoCmdRunner
	directoryLister DirectoryLister
}

func NewFinder(cmdRunner gocli.GoCmdRunner, directoryLister DirectoryLister) RealGoBinariesFinder {
	return RealGoBinariesFinder{
		cmdRunner,
		directoryLister,
	}
}

func (f *RealGoBinariesFinder) FindGoBinaries(gobin string) ([]GoBinary, error) {
	binaryNames, err := f.directoryLister.ListDirectoryEntries(gobin)
	if err != nil {
		return nil, err
	}

	var goBinaries []GoBinary

	for _, binaryName := range binaryNames {
		absPath := path.Join(gobin, binaryName)

		moduleInfo, err := f.getModuleInfo(absPath)
		if err != nil {
			return nil, fmt.Errorf("could not get module info about %v: %w", absPath, err)
		}

		latestVersion, err := f.getLatestModuleVersion(moduleInfo.moduleURL)
		if err != nil {
			return nil, fmt.Errorf("could not get latest version of %v: %w", moduleInfo.moduleURL, err)
		}

		goBinaries = append(goBinaries, GoBinary{
			ModuleURL:     moduleInfo.moduleURL,
			Version:       moduleInfo.version,
			Name:          binaryName,
			AbsPath:       absPath,
			LatestVersion: latestVersion,
		})
	}

	return goBinaries, nil
}

type parsedGoModuleInfo struct {
	moduleURL string
	version   string
}

func (f *RealGoBinariesFinder) getModuleInfo(binaryPath string) (*parsedGoModuleInfo, error) {
	moduleOutput, err := f.cmdRunner.RunGoCommand("version", "-m", binaryPath)
	if err != nil {
		return nil, err
	}
	goModuleInfo := findModuleURLInModuleOutput(moduleOutput)
	if goModuleInfo == nil {
		return nil, errors.New(fmt.Sprintf("could not parse module information for binary %s", binaryPath))
	}

	return goModuleInfo, nil
}

func (f *RealGoBinariesFinder) getLatestModuleVersion(moduleURL string) (string, error) {
	latestVersionModule := fmt.Sprintf("%s@latest", moduleURL)
	return f.cmdRunner.RunGoCommand("list", "-m", "-f", "{{.Version}}", latestVersionModule)
}

func findModuleURLInModuleOutput(output string) *parsedGoModuleInfo {
	r := regexp.MustCompile(`\s(path|mod)\s+([^\s]+)(\s+([^\s]+))?`)

	goModuleInfo := parsedGoModuleInfo{}

	for _, l := range strings.Split(output, "\n") {
		matches := r.FindStringSubmatch(l)
		if len(matches) == 0 {
			continue
		}

		switch matches[1] {
		case "path":
			goModuleInfo.moduleURL = matches[2]
		case "mod":
			goModuleInfo.version = matches[4]
		default:
			panic(fmt.Sprintf("unknown match %s", matches[0]))
		}
	}

	if len(goModuleInfo.version) == 0 || len(goModuleInfo.moduleURL) == 0 {
		return nil
	}

	return &goModuleInfo
}
