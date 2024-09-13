package gocli

import (
	"fmt"
	"strings"
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

func (cli *GoCLI) UpgradePackage(name string, buildTags []string) (string, error) {
	args := []string{"install"}

	if len(buildTags) > 0 {
		args = append(args, "-tags", strings.Join(buildTags, ","))
	}

	packageNameWithVersion := fmt.Sprintf("%s@latest", name)
	args = append(args, packageNameWithVersion)

	return cli.cmdRunner.RunGoCommand(args...)
}
