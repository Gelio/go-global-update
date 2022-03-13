package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"

	"github.com/urfave/cli/v2"
)

func getExecutableBinariesPath(cli *gocli.GoCLI) (string, error) {
	gobin, err := cli.GetEnvVar("GOBIN")
	if err != nil {
		return "", nil
	}
	if len(gobin) > 0 {
		return gobin, nil
	}

	gopath, err := cli.GetEnvVar("GOPATH")
	if err != nil {
		return "", nil
	}
	if len(gopath) == 0 {
		return "", errors.New("GOPATH and GOPATH are not defined in 'go env' command")
	}

	gobin = path.Join(gopath, "bin")

	return gobin, nil
}

func main() {
	app := &cli.App{
		Name:    "go-global-update",
		Usage:   "Update globally installed go binaries",
		Version: "v0.1.0",
		Action: func(c *cli.Context) error {
			updateBinaries()
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func updateBinaries() {
	goCmdRunner := gocli.RealGoCmdRunner{}
	cli := gocli.New(&goCmdRunner)
	gobin, err := getExecutableBinariesPath(&cli)
	if err != nil {
		fmt.Println("Error while trying to determine the executable binaries path", err)
		os.Exit(1)
	}

	if err := os.Chdir(gobin); err != nil {
		fmt.Println("Error when changing directory to GOBIN", err)
		os.Exit(1)
	}

	goBinariesFinder := gobinaries.NewFinder(&goCmdRunner, &gobinaries.FilesystemDirectoryLister{})
	goBinaries, err := goBinariesFinder.FindGoBinaries(gobin)
	if err != nil {
		fmt.Println("Error while trying to find go binaries to update", err)
		os.Exit(1)
	}

	var upgradeErrors []error

	for _, goBinary := range goBinaries {
		fmt.Printf("%s (current version %s) ... ", goBinary.Name, goBinary.Version)
		upgradeOutput, err := cli.UpgradePackage(goBinary.ModuleURL)
		if err != nil {
			upgradeErrors = append(upgradeErrors, err)
			fmt.Println("❌")
			fmt.Println("\tCould not upgrade package")
		} else {
			fmt.Println("✅")
		}

		if len(upgradeOutput) > 0 {
			fmt.Println(upgradeOutput)
		}
		fmt.Println()
	}

	if len(upgradeErrors) > 0 {
		os.Exit(1)
	}
}
