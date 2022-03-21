package test

import (
	"fmt"
	"io/ioutil"
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
	gobin, err := ioutil.TempDir(INTEGRATION_TESTS_DIR, "test-")
	require.Nil(t, err)

	gobin, err = filepath.Abs(gobin)
	require.Nil(t, err)

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

func installBinary(t *testing.T, gobin, pathURL string) {
	err := newTestCommand(t, gobin, "go", "install", pathURL).Run()
	require.Nil(t, err)
}

func getVersion(t *testing.T, gobin, binaryName string) (string, error) {
	// NOTE: assumes that `./binary --version` will return the current version
	version, err := newTestCommand(t, gobin, filepath.Join(gobin, binaryName), "--version").Output()
	return string(version), err
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
	require.Nil(t, os.MkdirAll(INTEGRATION_TESTS_DIR, os.ModePerm))
}

func TestIntegration(t *testing.T) {
	type binaryToInstall struct {
		name           string
		pathAndVersion string
		beforeUpdate   func(t *testing.T, version string)
		afterUpdate    func(t *testing.T, version string)
	}

	gofumptBinaryToInstall := binaryToInstall{
		name:           "gofumpt",
		pathAndVersion: "mvdan.cc/gofumpt@v0.2.0",
		beforeUpdate: func(t *testing.T, version string) {
			require.Contains(t, version, "v0.2.0")
		},
		afterUpdate: func(t *testing.T, version string) {
			assert.NotContains(t, version, "v0.2.0")
		},
	}
	shfmtBinaryToInstall := binaryToInstall{
		name:           "shfmt",
		pathAndVersion: "mvdan.cc/sh/v3/cmd/shfmt@v3.4.2",
		beforeUpdate: func(t *testing.T, version string) {
			require.Contains(t, version, "v3.4.2")
		},
		afterUpdate: func(t *testing.T, version string) {
			assert.NotContains(t, version, "v3.4.2")
		},
	}

	cases := []struct {
		name              string
		binariesToInstall []binaryToInstall
		updateArgs        []string
	}{
		{
			name:              "single binary",
			binariesToInstall: []binaryToInstall{gofumptBinaryToInstall},
		},
		{
			name:              "multiple binary",
			binariesToInstall: []binaryToInstall{gofumptBinaryToInstall, shfmtBinaryToInstall},
		},
		{
			name:       "single binary when multiple binaries installed",
			updateArgs: []string{binaryName("gofumpt")},
			binariesToInstall: []binaryToInstall{
				gofumptBinaryToInstall,
				{
					name:           shfmtBinaryToInstall.name,
					pathAndVersion: shfmtBinaryToInstall.pathAndVersion,
					beforeUpdate:   shfmtBinaryToInstall.beforeUpdate,
					// NOTE: the shfmt binary should not be upgraded
					afterUpdate: shfmtBinaryToInstall.beforeUpdate,
				},
			},
		},
		{
			name:       "dry-run",
			updateArgs: []string{"--dry-run"},
			binariesToInstall: []binaryToInstall{
				{
					name:           gofumptBinaryToInstall.name,
					pathAndVersion: gofumptBinaryToInstall.pathAndVersion,
					beforeUpdate:   gofumptBinaryToInstall.beforeUpdate,
					// NOTE: the binary should not be upgraded
					afterUpdate: gofumptBinaryToInstall.beforeUpdate,
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
	}

	ensureIntegrationTestsDirExists(t)

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			gobin := prepareTempGobin(t)

			for _, binary := range c.binariesToInstall {
				installBinary(t, gobin, binary.pathAndVersion)
			}

			for _, binary := range c.binariesToInstall {
				version, err := getVersion(t, gobin, binary.name)
				require.Nil(t, err)
				binary.beforeUpdate(t, version)
			}

			err := newGoGlobalUpdateCommand(t, gobin, c.updateArgs...).Run()
			assert.Nil(t, err)

			for _, binary := range c.binariesToInstall {
				version, err := getVersion(t, gobin, binary.name)
				require.Nil(t, err)
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
	err := newTestCommand(t, gobin, "go", "build", "-o", filepath.Join(gobin, builtBinaryName), mainGoBinaryPath).Run()
	require.Nil(t, err)

	output, err := newGoGlobalUpdateCommand(t, gobin, builtBinaryName).Output()
	assert.Nil(t, err)
	assert.Contains(t, string(output), fmt.Sprintf("%s (version: (devel), cannot upgrade)", builtBinaryName))
	assert.Contains(t, string(output), "binary was built from source")

	version, err := newTestCommand(t, gobin, "go", "version", "-m", filepath.Join(gobin, builtBinaryName)).Output()
	assert.Nil(t, err)
	assert.Contains(t, string(version), "(devel)", "Binary was upgraded but binaries built from source should be skipped")
}

func TestInstalledFromSource(t *testing.T) {
	ensureIntegrationTestsDirExists(t)

	// Verify handling of binaries installed using `go install` in a local repository
	// See https://github.com/Gelio/go-global-update/issues/3#issuecomment-1071221182

	gobin := prepareTempGobin(t)
	defer os.RemoveAll(gobin)

	repositoryDir, err := filepath.Abs("..")
	require.Nil(t, err)

	installCommand := newTestCommand(t, gobin, "go", "install")
	installCommand.Dir = repositoryDir

	err = installCommand.Run()
	require.Nil(t, err)

	builtBinaryName := binaryName("go-global-update")
	output, err := newGoGlobalUpdateCommand(t, gobin, builtBinaryName).Output()
	assert.Nil(t, err)
	assert.Contains(t, string(output), fmt.Sprintf("%s (version: (devel), can upgrade to v", builtBinaryName))
	assert.Contains(t, string(output), "binary was installed from source")

	version, err := newTestCommand(t, gobin, "go", "version", "-m", filepath.Join(gobin, builtBinaryName)).Output()
	assert.Nil(t, err)
	assert.Contains(t, string(version), "(devel)", "Binary was upgraded but binaries built from source should be skipped")
}
