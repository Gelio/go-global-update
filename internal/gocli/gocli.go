package gocli

import (
	"fmt"
	"os/exec"
	"strings"
)

type GoCmdRunner interface {
	RunGoCommand(args ...string) (string, error)
}

type RealGoCmdRunner struct{}

func (runner *RealGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	bytes, err := cmd.CombinedOutput()

	// NOTE: always include output for displaying errors
	return strings.TrimSpace(string(bytes)), err
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

func (cli *GoCLI) UpgradePackage(name string) (string, error) {
	packageNameWithVersion := fmt.Sprintf("%s@latest", name)
	return cli.cmdRunner.RunGoCommand("install", packageNameWithVersion)
}
