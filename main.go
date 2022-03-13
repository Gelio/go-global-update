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
		Name: "go-global-update",
		Usage: `Update globally installed go binaries.

   By default it will update all upgradable binaries in GOBIN.
   If arguments are provided, only those binaries will be updated.

   Examples:

   * go-global-update gofumpt gopls shfmt
   * go-global-update --dry-run`,
		Version:   "v0.1.0",
		ArgsUsage: "[binaries to update...]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Display debug information",
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "Check which binaries are upgradable without actually installing new versions",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Include more detailed information in the logs",
			},
		},
		Action: func(c *cli.Context) error {
			binariesToUpdate := c.Args().Slice()
			return updateBinaries(
				logger,
				c.Bool("dry-run"),
				c.Bool("verbose"),
				binariesToUpdate,
			)
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
	if c.Bool("debug") {
		logLevel = zap.DebugLevel
	}

	loggerConfig.Level.SetLevel(logLevel)
}

// updateBinaries updates binaries in GOBIN
//
// If binariesToUpdate is empty, the command will attempt to update all
// found binaries in GOBIN.
func updateBinaries(logger *zap.Logger, dryRun, verbose bool, binariesToUpdate []string) error {
	goCmdRunner := gocli.NewCmdRunner(logger)
	goCLI := gocli.New(&goCmdRunner)
	gobin, err := getExecutableBinariesPath(&goCLI)
	if err != nil {
		fmt.Println("Error while trying to determine the GOBIN path", err)
		os.Exit(1)
	}

	logger.Debug("found GOBIN path", zap.String("GOBIN", gobin))

	if err := os.Chdir(gobin); err != nil {
		fmt.Println("Error when changing directory to GOBIN", err)
		os.Exit(1)
	}

	introspecter := gobinaries.NewIntrospecter(&goCmdRunner, gobin, logger)
	if len(binariesToUpdate) == 0 {
		lister := gobinaries.FilesystemDirectoryLister{}
		binariesToUpdate, err = lister.ListDirectoryEntries(gobin)
		if err != nil {
			fmt.Printf("Error while trying to list GOBIN (%s) entries: %v", gobin, err)
			os.Exit(1)
		}
	}

	introspectionResults, err := gobinaries.IntrospectBinaries(&introspecter, binariesToUpdate)
	if err != nil {
		fmt.Println("Error while trying to find go binaries to update", err)
		os.Exit(1)
	}

	var upgradeErrors []error

	for _, result := range introspectionResults {
		if result.Error != nil {
			fmt.Println(result.Error)
			continue
		}

		binary := result.Binary
		if !binary.UpgradePossible() {
			fmt.Printf("%s (current version %s, latest)\n", binary.Name, binary.Version)
			continue
		}

		if dryRun {
			fmt.Printf("%s (current version %s, would upgrade to %s)\n", binary.Name, binary.Version, binary.LatestVersion)
			continue
		}

		fmt.Printf("%s (current version %s, upgrading to %s) ... ", binary.Name, binary.Version, binary.LatestVersion)
		upgradeOutput, err := goCLI.UpgradePackage(binary.PathURL)
		if err != nil {
			upgradeErrors = append(upgradeErrors, err)
			fmt.Println("❌")
			fmt.Println("\tCould not upgrade package")
		} else {
			fmt.Println("✅")
		}

		if len(upgradeOutput) > 0 && (verbose || err != nil) {
			fmt.Println(upgradeOutput)
		}
		fmt.Println()
	}

	if len(upgradeErrors) > 0 {
		return cli.Exit("Some packages failed to upgrade", 1)
	}

	return nil
}
