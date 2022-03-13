package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"go.uber.org/zap"

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
	loggerConfig := zap.NewDevelopmentConfig()
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatalf("cannot initialize zap logger: %v", err)
	}
	defer logger.Sync()

	app := &cli.App{
		Name:    "go-global-update",
		Usage:   "Update globally installed go binaries",
		Version: "v0.1.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name: "verbose",
			},
		},
		Action: func(c *cli.Context) error {
			return updateBinaries(logger)
		},
		Before: func(c *cli.Context) error {
			updateLoggerLevel(&loggerConfig, c)
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("could not run command: %v", err)
	}
}

func updateLoggerLevel(loggerConfig *zap.Config, c *cli.Context) {
	logLevel := zap.InfoLevel
	if c.Bool("verbose") {
		logLevel = zap.DebugLevel
	}

	loggerConfig.Level.SetLevel(logLevel)
}

func updateBinaries(logger *zap.Logger) error {
	goCmdRunner := gocli.NewCmdRunner(logger)
	goCLI := gocli.New(&goCmdRunner)
	gobin, err := getExecutableBinariesPath(&goCLI)
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
		upgradeOutput, err := goCLI.UpgradePackage(binary.PathURL)
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
		return cli.Exit("Some packages failed to upgrade", 1)
	}

	return nil
}
