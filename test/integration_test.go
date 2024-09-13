package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const INTEGRATION_TESTS_DIR string = "integration-tests"

// NOTE: the working directory in the tests is the directory of the current
// test file.
// main.go is in the parent directory.
var mainGoBinaryPath string = filepath.Join("..", "main.go")

// prepareTempoGobin returns an absolute path to a temporary directory. It
// could serve as GOBIN for a test.
func prepareTempGobin(t *testing.T) string {
	gobin, err := os.MkdirTemp(INTEGRATION_TESTS_DIR, "test-")
	require.Nil(t, err, "could not create a temporary directory for GOBIN")

	gobin, err = filepath.Abs(gobin)
	require.Nil(t, err, "could not get the absolute directory of GOBIN")

	return gobin
}

func newTestCommand(t *testing.T, gobin, name string, args ...string) *exec.Cmd {
	command := exec.Command(name, args...)
	command.Env = os.Environ()
	// NOTE: setting the GOBIN will instruct `go install` to install binaries
	// in that GOBIN directory. Any installation during tests will not affect
	// the current user's installed binaries.
	// This also means it should be easy to remove those artifacts - just remove
	// the GOBIN.
	command.Env = append(command.Env, fmt.Sprintf("GOBIN=%s", gobin))

	return command
}

func newGoGlobalUpdateCommand(t *testing.T, gobin string, args ...string) *exec.Cmd {
	args = append([]string{"run", mainGoBinaryPath}, args...)
	return newTestCommand(t, gobin, "go", args...)
}

func installBinary(t *testing.T, gobin, pathURL string, buildTags []string) {
	args := []string{"install"}
	if len(buildTags) > 0 {
		args = append(args, "-tags", strings.Join(buildTags, ","))
	}
	args = append(args, pathURL)

	output, err := newTestCommand(t, gobin, "go", args...).CombinedOutput()
	require.Nilf(t, err, "could not install binary %s in GOBIN %s (args: %v)\noutput: %s", pathURL, gobin, args, string(output))
}

func getVersion(t *testing.T, gobin, binaryName string) (string, error) {
	versionOutput, err := newTestCommand(t, gobin, "go", "version", "-m", filepath.Join(gobin, binaryName)).Output()
	return string(versionOutput), err
}

func binaryName(baseName string) string {
	// NOTE: inspired by https://github.com/markbates/refresh/pull/4/files
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(baseName, ".exe") {
			return baseName + ".exe"
		}
	}

	return baseName
}

func ensureIntegrationTestsDirExists(t *testing.T) {
	require.Nil(t, os.MkdirAll(INTEGRATION_TESTS_DIR, os.ModePerm), "could not create a directory for integration tests at", INTEGRATION_TESTS_DIR)
}

func TestIntegration(t *testing.T) {
	type binaryToInstall struct {
		name           string
		pathAndVersion string
		beforeUpdate   func(t *testing.T, version string)
		afterUpdate    func(t *testing.T, version string)
	}

	gnosticBinaryToInstall := binaryToInstall{
		name:           "gnostic",
		pathAndVersion: "github.com/google/gnostic@v0.6.2",
		beforeUpdate: func(t *testing.T, version string) {
			require.Contains(t, version, "v0.6.2")
		},
		afterUpdate: func(t *testing.T, version string) {
			assert.NotContains(t, version, "v0.6.2", "binary was unexpectedly updated")
		},
	}
	shfmtBinaryToInstall := binaryToInstall{
		name:           "shfmt",
		pathAndVersion: "mvdan.cc/sh/v3/cmd/shfmt@v3.4.2",
		beforeUpdate: func(t *testing.T, version string) {
			require.Contains(t, version, "v3.4.2")
		},
		afterUpdate: func(t *testing.T, version string) {
			assert.NotContains(t, version, "v3.4.2", "binary was unexpectedly updated")
		},
	}

	cases := []struct {
		name              string
		binariesToInstall []binaryToInstall
		updateArgs        []string
	}{
		{
			name:              "single binary",
			binariesToInstall: []binaryToInstall{shfmtBinaryToInstall},
		},
		{
			name:              "multiple binaries",
			binariesToInstall: []binaryToInstall{gnosticBinaryToInstall, shfmtBinaryToInstall},
		},
		{
			name:       "single binary when multiple binaries installed",
			updateArgs: []string{binaryName("shfmt")},
			binariesToInstall: []binaryToInstall{
				shfmtBinaryToInstall,
				{
					name:           gnosticBinaryToInstall.name,
					pathAndVersion: gnosticBinaryToInstall.pathAndVersion,
					beforeUpdate:   gnosticBinaryToInstall.beforeUpdate,
					// NOTE: the binary should not be upgraded
					afterUpdate: gnosticBinaryToInstall.beforeUpdate,
				},
			},
		},
		{
			name:       "dry-run",
			updateArgs: []string{"--dry-run"},
			binariesToInstall: []binaryToInstall{
				{
					name:           gnosticBinaryToInstall.name,
					pathAndVersion: gnosticBinaryToInstall.pathAndVersion,
					beforeUpdate:   gnosticBinaryToInstall.beforeUpdate,
					// NOTE: the binary should not be upgraded
					afterUpdate: gnosticBinaryToInstall.beforeUpdate,
				},
				{
					name:           shfmtBinaryToInstall.name,
					pathAndVersion: shfmtBinaryToInstall.pathAndVersion,
					beforeUpdate:   shfmtBinaryToInstall.beforeUpdate,
					// NOTE: the binary should not be upgraded
					afterUpdate: shfmtBinaryToInstall.beforeUpdate,
				},
			},
		},
		{
			name:              "multiple binaries with forced colored output",
			updateArgs:        []string{"--colors"},
			binariesToInstall: []binaryToInstall{gnosticBinaryToInstall, shfmtBinaryToInstall},
		},
	}

	ensureIntegrationTestsDirExists(t)

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			gobin := prepareTempGobin(t)

			for _, binary := range c.binariesToInstall {
				installBinary(t, gobin, binary.pathAndVersion, nil)
			}

			for _, binary := range c.binariesToInstall {
				version, err := getVersion(t, gobin, binaryName(binary.name))
				require.Nilf(t, err, "could not get version of %s before updating\noutput: %s", binaryName(binary.name), string(version))
				binary.beforeUpdate(t, version)
			}

			output, err := newGoGlobalUpdateCommand(t, gobin, c.updateArgs...).CombinedOutput()
			assert.Nilf(t, err, "could not run go-global-update command with args: %v\noutput: %s", c.updateArgs, string(output))

			for _, binary := range c.binariesToInstall {
				version, err := getVersion(t, gobin, binaryName(binary.name))
				require.Nilf(t, err, "could not get version of %s after updating\noutput: %s", binaryName(binary.name), string(version))
				binary.afterUpdate(t, version)
			}

			// NOTE: only remove if the test finished successfully
			os.RemoveAll(gobin)
		})
	}
}

func TestBuiltFromSourceCommandLineArguments(t *testing.T) {
	ensureIntegrationTestsDirExists(t)

	// Verify handling of binaries built using `go build -o path main.go`
	// See https://github.com/Gelio/go-global-update/issues/3#issuecomment-1072178664

	gobin := prepareTempGobin(t)
	defer os.RemoveAll(gobin)

	builtBinaryName := binaryName("built-binary")
	output, err := newTestCommand(t, gobin, "go", "build", "-o", filepath.Join(gobin, builtBinaryName), mainGoBinaryPath).CombinedOutput()
	require.Nilf(t, err, "could not build %s\noutput: %s", mainGoBinaryPath, string(output))

	output, err = newGoGlobalUpdateCommand(t, gobin, builtBinaryName).CombinedOutput()
	assert.Nilf(t, err, "could not run go-global-update for %s\noutput: %s", builtBinaryName, string(output))
	assert.Regexp(t, fmt.Sprintf("%s\\s+\\(devel\\)\\s+cannot upgrade", builtBinaryName), string(output))
	assert.Contains(t, string(output), "binary was built from source")

	version, err := newTestCommand(t, gobin, "go", "version", "-m", filepath.Join(gobin, builtBinaryName)).Output()
	assert.Nil(t, err, "could not get version of", builtBinaryName)
	assert.Contains(t, string(version), "(devel)", "Binary was upgraded but binaries built from source should be skipped")
}

func TestInstalledFromSource(t *testing.T) {
	ensureIntegrationTestsDirExists(t)

	// Verify handling of binaries installed using `go install` in a local repository
	// See https://github.com/Gelio/go-global-update/issues/3#issuecomment-1071221182

	gobin := prepareTempGobin(t)
	defer os.RemoveAll(gobin)

	repositoryDir, err := filepath.Abs("..")
	require.Nil(t, err, "could not get the absolute filepath of the repository")

	installCommand := newTestCommand(t, gobin, "go", "install")
	installCommand.Dir = repositoryDir

	err = installCommand.Run()
	require.Nil(t, err, "could not run go install for the go-global-update repository")

	builtBinaryName := binaryName("go-global-update")
	output, err := newGoGlobalUpdateCommand(t, gobin, builtBinaryName).Output()
	assert.Nil(t, err, "could not run go-global-update for", builtBinaryName)
	assert.Regexp(t, fmt.Sprintf("%s\\s+\\(devel\\)\\s+can upgrade to v", builtBinaryName), string(output))
	assert.Contains(t, string(output), "binary was installed from source")

	version, err := newTestCommand(t, gobin, "go", "version", "-m", filepath.Join(gobin, builtBinaryName)).Output()
	assert.Nil(t, err, "could not get version of", builtBinaryName)
	assert.Contains(t, string(version), "(devel)", "Binary was upgraded but binaries built from source should be skipped")
}

// TestDetectCommonProblems checks that common problems are detected and
// reported.
func TestDetectCommonProblems(t *testing.T) {
	ensureIntegrationTestsDirExists(t)

	cases := []struct {
		name                string
		installCommand      func(t *testing.T, gobin string) *exec.Cmd
		binaryName          string
		expectedProblemCode string
	}{
		{
			// NOTE: this seems like the only "stable" error that will keep
			// happening in the future.
			//
			// E003 repository being moved would require installing the
			// package before the repository was moved first, which means
			// going back in time or mocking the results of `go install`,
			// which is not intended in integration tests.
			//
			// E004 go.mod containing `replace` directives can be changed in the
			// latest version so tests will fail.
			name: "cobra moved to cobra-cli",
			installCommand: func(t *testing.T, gobin string) *exec.Cmd {
				return newTestCommand(t, gobin, "go", "install", "github.com/spf13/cobra/cobra@v1.3.0")
			},
			binaryName:          "cobra",
			expectedProblemCode: "E002",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			gobin := prepareTempGobin(t)
			defer os.RemoveAll(gobin)

			err := c.installCommand(t, gobin).Run()
			require.Nil(t, err, "could not run install command")

			builtBinaryName := binaryName(c.binaryName)
			output, err := newGoGlobalUpdateCommand(t, gobin, builtBinaryName).Output()

			assert.NotNil(t, err, "expected the update to fail due to a common update problem")
			assert.Contains(t, string(output), c.expectedProblemCode, "problem code not found")
		})
	}
}

func TestPreserveBuildTags(t *testing.T) {
	ensureIntegrationTestsDirExists(t)
	gobin := prepareTempGobin(t)

	installBinary(t, gobin, "mvdan.cc/sh/v3/cmd/shfmt@v3.4.2", []string{"tagA", "tagB"})
	builtBinaryName := binaryName("shfmt")

	output, err := getVersion(t, gobin, builtBinaryName)
	require.Nilf(t, err, "could not get version of shfmt\noutput: %s", string(output))
	require.Contains(t, output, "v3.4.2", "installed wrong version of the binary")
	require.Contains(t, output, "-tags=tagA,tagB", "expected build tags not found")

	outputBytes, err := newGoGlobalUpdateCommand(t, gobin, builtBinaryName).Output()
	assert.Nilf(t, err, "could not run go-global-update for shfmt\noutput: %s", string(output))
	assert.Contains(t, string(outputBytes), "(build tags: tagA,tagB)", "expected build tags to appear in the go-global-update output")

	output, err = getVersion(t, gobin, builtBinaryName)
	assert.Nilf(t, err, "could not get version of shfmt after updating\noutput: %s", string(output))
	assert.Contains(t, output, "-tags=tagA,tagB", "expected build tags not found after updating")
	assert.NotContains(t, output, "v3.4.2", "binary was not updated")
}
