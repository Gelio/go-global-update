package gocli

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type GoCmdRunner interface {
	RunGoCommand(args ...string) (string, error)
}

type RealGoCmdRunner struct{}

func (runner *RealGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}

func New(cmdRunner GoCmdRunner) GoCLI {
	return GoCLI{
		cmdRunner,
	}
}

type GoCLI struct {
	cmdRunner GoCmdRunner
}

func (cli *GoCLI) GetEnvVar(name string) (string, error) {
	return cli.cmdRunner.RunGoCommand("env", name)
}

func (cli *GoCLI) GetModuleURL(binaryName string) (string, error) {
	moduleOutput, err := cli.cmdRunner.RunGoCommand("version", "-m", binaryName)
	if err != nil {
		return "", err
	}
	moduleURL := findModuleURLInModuleOutput(moduleOutput)
	if len(moduleURL) == 0 {
		return "", errors.New(fmt.Sprintf("could not find module URL for binary %s", binaryName))
	}
	return moduleURL, nil
}

func findModuleURLInModuleOutput(output string) string {
	r := regexp.MustCompile(`\bpath\s+(.*)$`)

	for _, l := range strings.Split(output, "\n") {
		matches := r.FindStringSubmatch(l)
		if len(matches) != 0 {
			return matches[1]
		}
	}

	return ""
}

func (cli *GoCLI) UpgradePackage(name string) error {
	packageNameWithVersion := fmt.Sprintf("%s@latest", name)
	_, err := cli.cmdRunner.RunGoCommand("install", packageNameWithVersion)

	return err
}
