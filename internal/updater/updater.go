package updater

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/Gelio/go-global-update/internal/colors"
	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"github.com/fatih/color"
	"go.uber.org/zap"
)

type Options struct {
	Verbose bool
	DryRun  bool
	// List of binary names to update.
	// If empty, will update all binaries in GOBIN
	BinariesToUpdate []string
	// Whether to force reinstalling/updating all binaries.
	ForceReinstall bool
}

// UpdateBinaries updates binaries in GOBIN
//
// If binariesToUpdate is empty, the command will attempt to update all
// found binaries in GOBIN.
func UpdateBinaries(
	logger *zap.Logger,
	options Options,
	out io.Writer,
	colorsFactory *colors.DecoratorFactory,
	cmdRunner gocli.GoCmdRunner,
	lister gobinaries.DirectoryLister,
	fs FilesystemUtils,
) error {
	goCLI := gocli.New(cmdRunner)
	gobin, err := getExecutableBinariesPath(&goCLI)
	if err != nil {
		return fmt.Errorf("could not determine GOBIN path: %w", err)
	}

	logger.Debug("found GOBIN path", zap.String("GOBIN", gobin))

	if err := fs.Chdir(gobin); err != nil {
		return fmt.Errorf("could not change directory to GOBIN (%s): %w", gobin, err)
	}

	introspecter := gobinaries.NewIntrospecter(cmdRunner, gobin, logger)
	binaryNames, err := resolveBinaryNames(options.BinariesToUpdate, lister, gobin)
	if err != nil {
		return err
	}

	introspectionResults := gobinaries.IntrospectBinaries(&introspecter, binaryNames)
	printBinariesSummary(introspectionResults, out, colorsFactory, options.Verbose)

	if !options.DryRun {
		return updateBinaries(introspectionResults, &goCLI, out, colorsFactory, options)
	}

	return nil
}

func resolveBinaryNames(binariesToUpdate []string, lister gobinaries.DirectoryLister, gobin string) ([]string, error) {
	if len(binariesToUpdate) > 0 {
		return binariesToUpdate, nil
	}

	binaryNames, err := lister.ListDirectoryEntries(gobin)
	if err != nil {
		err = fmt.Errorf("could not list GOBIN (%s) entries: %w", gobin, err)
	}

	return binaryNames, err
}

func printBinariesSummary(
	introspectionResults []gobinaries.IntrospectionResult,
	out io.Writer,
	colorsFactory *colors.DecoratorFactory,
	verbose bool,
) {
	tabWriter := tabwriter.NewWriter(out, 0, 0, 6, ' ', tabwriter.StripEscape)
	fmt.Fprintln(tabWriter, "Binary\tCurrent version\tStatus")
	defer tabWriter.Flush()

	for _, result := range introspectionResults {
		if result.Error != nil {
			fmt.Fprintln(out, result.Error)
			continue
		}

		binary := result.Binary
		var latestVersionInfo string
		if binary.LatestVersion != "" {
			if binary.UpgradePossible() {
				latestVersionInfo = fmt.Sprintf("can upgrade to %s", colorsFactory.NewDecorator(color.FgGreen)(binary.LatestVersion))
			} else {
				latestVersionInfo = "up-to-date"
			}
		} else {
			latestVersionInfo = colorsFactory.NewDecorator(color.FgYellow)("cannot upgrade")
		}

		name := binary.Name
		if verbose {
			name = binary.PathURL
		}

		// NOTE: only the last column can safely use ANSI color codes. Otherwise,
		// column widths can be mismatched due to color codes used only in some
		// rows.
		// @see https://stackoverflow.com/questions/35398497/how-do-i-get-colors-to-work-with-golang-tabwriter
		fmt.Fprintf(tabWriter, "%s\t%s\t%s\n", name, binary.Version, latestVersionInfo)
	}
}

func updateBinaries(
	introspectionResults []gobinaries.IntrospectionResult,
	goCLI *gocli.GoCLI,
	out io.Writer,
	colorsFactory *colors.DecoratorFactory,
	options Options,
) error {
	var upgradeErrors []error
	var binariesToUpdate []gobinaries.GoBinary

	fmt.Fprintln(out)

	binaryNameFormatter := colorsFactory.NewDecorator(color.FgCyan)
	faintFormatter := colorsFactory.NewDecorator(color.Faint)

	for _, result := range introspectionResults {
		if result.Error != nil {
			continue
		}
		if !result.Binary.UpgradePossible() && !options.ForceReinstall {
			continue
		}

		if binary := result.Binary; binary.BuiltFromSource() {
			verb := "reinstalling"
			if result.Binary.UpgradePossible() {
				verb = "upgrading"
			}
			fmt.Fprintf(out, "Skipping %s %s\n    ", verb, binaryNameFormatter(binary.Name))
			if binary.BuiltWithGoBuild() {
				fmt.Fprintf(out, "The binary was built from source (probably using \"%s\") and the binary path is unknown.\n",
					faintFormatter("go build"))
			} else {
				fmt.Fprintf(out, "The binary was installed from source (probably using \"%s\" in the cloned repository).\n",
					faintFormatter("go install"))
			}
			pathURL := binary.PathURL
			if binary.BuiltWithGoBuild() {
				// NOTE: binaries built with `go build` have `command-line-arguments`
				// as their `path` which would not make sense in help message.
				pathURL = "repositoryPath"
			}

			fmt.Fprintf(out, "    Install the binary using \"%s\" instead.\n",
				faintFormatter(fmt.Sprintf("go install %s@latest", pathURL)))
			fmt.Fprintf(out, "%s\n\n", binaryBuiltFromSourceProblem.String(colorsFactory))
			continue
		}

		binariesToUpdate = append(binariesToUpdate, result.Binary)
	}

	if len(binariesToUpdate) == 0 {
		return nil
	}

	latestVersionFormatter := colorsFactory.NewDecorator(color.FgGreen)

	for _, binary := range binariesToUpdate {
		var buildTagsInfo string
		if len(binary.BuildTags) > 0 {
			buildTagsInfo = fmt.Sprintf(" (build tags: %s)", faintFormatter(strings.Join(binary.BuildTags, ",")))
		}

		if binary.UpgradePossible() {
			fmt.Fprintf(out, "Upgrading %s to %s%s ... ", binaryNameFormatter(binary.Name),
				latestVersionFormatter(binary.LatestVersion), buildTagsInfo)
		} else {
			fmt.Fprintf(out, "Force-reinstalling %s %s%s ... ", binaryNameFormatter(binary.Name),
				latestVersionFormatter(binary.LatestVersion), buildTagsInfo)
		}
		upgradeOutput, err := goCLI.UpgradePackage(binary.PathURL, binary.BuildTags)
		if err != nil {
			upgradeErrors = append(upgradeErrors, err)
			fmt.Fprintln(out, "❌")
			fmt.Fprintln(out, "    Could not install package")
		} else {
			fmt.Fprintln(out, "✅")
		}

		if len(upgradeOutput) > 0 && (options.Verbose || err != nil) {
			fmt.Fprintln(out, upgradeOutput)

			for _, problem := range FindCommonUpdateProblems(upgradeOutput) {
				fmt.Fprintf(out, "%s\n", problem.String(colorsFactory))
			}
		}
		fmt.Fprintln(out)
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("could not install %s package(s)",
			colorsFactory.NewDecorator(color.FgRed, color.Bold)(len(upgradeErrors)))
	}

	return nil
}

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

	gobin = filepath.Join(gopath, "bin")

	return gobin, nil
}
