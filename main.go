package main

import (
	"log"
	"os"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"github.com/Gelio/go-global-update/internal/updater"
	"go.uber.org/zap"

	"github.com/urfave/cli/v2"
)

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
			cmdRunner := gocli.NewCmdRunner(logger)

			err := updater.UpdateBinaries(
				logger,
				updater.Options{
					Debug:            c.Bool("debug"),
					DryRun:           c.Bool("dry-run"),
					Verbose:          c.Bool("verbose"),
					BinariesToUpdate: c.Args().Slice(),
				},
				os.Stdout,
				&cmdRunner,
				&gobinaries.FilesystemDirectoryLister{},
				&updater.Filesystem{},
			)
			return err
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
