package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Gelio/go-global-update/internal/colors"
	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"github.com/Gelio/go-global-update/internal/updater"
	"github.com/fatih/color"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/urfave/cli/v2"
)

func main() {
	loggerConfig := zap.NewDevelopmentConfig()

	// NOTE: re-define the version flag to use `V` instead of `v` as the alias
	// `-v` belongs to the `--verbose` flag
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	app := &cli.App{
		Name: "go-global-update",
		Usage: `Update globally installed go binaries.

   By default it will update all upgradable binaries in GOBIN.
   If arguments are provided, only those binaries will be updated.

   Examples:

   * go-global-update gofumpt gopls shfmt
   * go-global-update --dry-run`,
		Version:   "v0.2.0",
		ArgsUsage: "[binaries to update...]",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Display debug information",
			},
			&cli.BoolFlag{
				Name:    "dry-run",
				Aliases: []string{"n"},
				Usage:   "Check which binaries are upgradable without actually installing new versions",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Include more detailed information in the logs",
			},
			&cli.BoolFlag{
				Name:  "colors",
				Usage: "Force using ANSI color codes in the output even if the output is not a TTY.\n\t\tSet the NO_COLOR environment variable if you want to force-disable colors (see https://no-color.org/).",
			},
		},
		Action: func(c *cli.Context) error {
			forceColors := c.Bool("colors")
			colorsDecoratorFactory := colors.NewFactory(forceColors)

			logger, err := loggerConfig.Build()
			if err != nil {
				return fmt.Errorf("cannot initialize zap logger: %w", err)
			}
			defer logger.Sync()

			cmdRunner := gocli.NewCmdRunner(logger)

			err = updater.UpdateBinaries(
				logger,
				updater.Options{
					DryRun:           c.Bool("dry-run"),
					Verbose:          c.Bool("verbose"),
					BinariesToUpdate: c.Args().Slice(),
				},
				os.Stdout,
				&colorsDecoratorFactory,
				&cmdRunner,
				&gobinaries.FilesystemDirectoryLister{},
				&updater.Filesystem{},
			)
			return err
		},
		Before: func(c *cli.Context) error {
			debugMode := c.Bool("debug")
			updateLoggerLevel(&loggerConfig, debugMode)
			forceColors := c.Bool("colors")
			updateLoggerColors(&loggerConfig, forceColors)

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("could not run command: %v", err)
	}
}

func updateLoggerLevel(loggerConfig *zap.Config, debugMode bool) {
	logLevel := zap.InfoLevel
	if debugMode {
		logLevel = zap.DebugLevel
	}

	loggerConfig.Level.SetLevel(logLevel)
}

func updateLoggerColors(loggerConfig *zap.Config, forceColors bool) {
	if forceColors || !color.NoColor {
		loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
}
