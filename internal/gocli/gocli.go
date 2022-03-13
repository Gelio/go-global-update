package gocli

import (
	"fmt"
)

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
