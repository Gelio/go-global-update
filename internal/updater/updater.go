package updater

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Gelio/go-global-update/internal/gobinaries"
	"github.com/Gelio/go-global-update/internal/gocli"
	"go.uber.org/zap"
)

type Options struct {
	Debug   bool
	Verbose bool
	DryRun  bool
	// List of binary names to update.
	// If empty, will update all binaries in GOBIN
	BinariesToUpdate []string
}

// UpdateBinaries updates binaries in GOBIN
//
// If binariesToUpdate is empty, the command will attempt to update all
// found binaries in GOBIN.
func UpdateBinaries(
	logger *zap.Logger,
	options Options,
	out io.Writer,
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
	printBinariesSummary(introspectionResults, out, options.Verbose)

	if !options.DryRun {
		return updateBinaries(introspectionResults, &goCLI, out, options.Verbose)
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
	verbose bool,
) {
	for _, result := range introspectionResults {
		if result.Error != nil {
			fmt.Fprintln(out, result.Error)
			continue
		}

		binary := result.Binary
		var latestVersionInfo string
		if binary.LatestVersion != "" {
			if binary.UpgradePossible() {
				latestVersionInfo = fmt.Sprintf("can upgrade to %s", binary.LatestVersion)
			} else {
				latestVersionInfo = "up-to-date"
			}
		} else {
			latestVersionInfo = "cannot upgrade"
		}

		name := binary.Name
		if verbose {
			name = binary.PathURL
		}

		fmt.Fprintf(out, "%s (version: %s, %s)\n", name, binary.Version, latestVersionInfo)
	}
}

func updateBinaries(
	introspectionResults []gobinaries.IntrospectionResult,
	goCLI *gocli.GoCLI,
	out io.Writer,
	verbose bool,
) error {
	var upgradeErrors []error
	var binariesToUpdate []gobinaries.GoBinary

	fmt.Fprintln(out)

	for _, result := range introspectionResults {
		if result.Error != nil || !result.Binary.UpgradePossible() {
			continue
		}

		if binary := result.Binary; binary.BuiltFromSource() {
			fmt.Fprintf(out, "Skipping upgrading %s\n    ", binary.Name)
			if binary.BuiltWithGoBuild() {
				fmt.Fprintln(out, `The binary was built from source (probably using "go build") and the binary path is unknown.`)
			} else {
				fmt.Fprintln(out, `The binary was installed from source (probably using "go install" in the cloned repository).`)
			}
			fmt.Fprintln(out, "    Install the binary using \"go install repositoryPath@latest\" instead.")
			fmt.Fprintln(out)
			continue
		}

		binariesToUpdate = append(binariesToUpdate, result.Binary)
	}

	if len(binariesToUpdate) == 0 {
		return nil
	}

	for _, binary := range binariesToUpdate {
		fmt.Fprintf(out, "Upgrading %s to %s ... ", binary.Name, binary.LatestVersion)
		upgradeOutput, err := goCLI.UpgradePackage(binary.PathURL)
		if err != nil {
			upgradeErrors = append(upgradeErrors, err)
			fmt.Fprintln(out, "❌")
			fmt.Fprintln(out, "    Could not upgrade package")
		} else {
			fmt.Fprintln(out, "✅")
		}

		if len(upgradeOutput) > 0 && (verbose || err != nil) {
			fmt.Fprintln(out, upgradeOutput)
		}
		fmt.Fprintln(out)
	}

	if len(upgradeErrors) > 0 {
		return fmt.Errorf("could not upgrade %d packages", len(upgradeErrors))
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
