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
	ModuleURL string
	Name      string
	AbsPath   string
	Version   string
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
			return nil, err
		}

		goBinaries = append(goBinaries, GoBinary{
			ModuleURL: moduleInfo.moduleURL,
			Version:   moduleInfo.version,
			Name:      binaryName,
			AbsPath:   absPath,
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
