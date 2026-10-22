package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Gelio/go-global-update/internal/colors"
	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"github.com/Gelio/go-global-update/internal/updater"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v2"
)

func main() {
	// NOTE: re-define the version flag to use `V` instead of `v` as the alias
	// `-v` belongs to the `--verbose` flag
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"V"},
		Usage:   "print the version",
	}

	app := &cli.App{
		Name:  "go-global-update",
		Usage: "Update globally installed go binaries.",
		UsageText: `go-global-update [global options] [binaries to update...]

Examples:

go-global-update gofumpt gopls shfmt
go-global-update --dry-run`,
		Description: `By default it will update all upgradable binaries in GOBIN.
If arguments are provided, only those binaries will be updated.`,
		Version:                "v0.2.5",
		ArgsUsage:              "[binaries to update...]",
		UseShortOptionHandling: true,
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
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force reinstall all binaries, even if they do not need to be updated",
			},
		},
		Action: func(c *cli.Context) error {
			forceColors := c.Bool("colors")
			colorsDecoratorFactory := colors.NewFactory(forceColors)

			logger := newLogger(c.Bool("debug"), forceColors)

			cmdRunner := gocli.NewCmdRunner(logger)

			options := updater.Options{
				DryRun:           c.Bool("dry-run"),
				Verbose:          c.Bool("verbose"),
				ForceReinstall:   c.Bool("force"),
				BinariesToUpdate: c.Args().Slice(),
			}

			if options.DryRun && options.ForceReinstall {
				return fmt.Errorf("--dry-run and --force options cannot be used together")
			}

			err := updater.UpdateBinaries(
				logger,
				options,
				os.Stdout,
				&colorsDecoratorFactory,
				&cmdRunner,
				&gobinaries.FilesystemDirectoryLister{},
				&updater.Filesystem{},
			)
			return err
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("could not run command: %v", err)
	}
}

func newLogger(verbose, forceColors bool) *slog.Logger {
	lvl := slog.LevelInfo
	if verbose {
		lvl = slog.LevelDebug
	}

	stderrIsTerminal := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())

	return slog.New(tint.NewTextHandler(os.Stderr, &tint.Options{
		Level:   lvl,
		NoColor: !logColorsEnabled(forceColors, stderrIsTerminal, os.Getenv),
	}))
}

// logColorsEnabled decides whether log lines are colored, following the same
// rules github.com/fatih/color applies to the summary table: `--colors` forces
// colors, otherwise they are used only on a terminal when NO_COLOR is unset or
// empty (https://no-color.org/) and TERM is not "dumb".
func logColorsEnabled(forceColors, stderrIsTerminal bool, getenv func(string) string) bool {
	if forceColors {
		return true
	}

	return stderrIsTerminal && getenv("NO_COLOR") == "" && getenv("TERM") != "dumb"
}
