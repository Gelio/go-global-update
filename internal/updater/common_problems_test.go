package updater

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindCommonProblems(t *testing.T) {
	cases := []struct {
		name                 string
		goInstallOutput      string
		expectedProblemNames []string
	}{
		{
			name: "E002 module found but does not contain package cobra",
			goInstallOutput: `
go: downloading github.com/spf13/cobra v1.4.0
go install: github.com/spf13/cobra/cobra@latest: module github.com/spf13/cobra@latest found (v1.4.0), but does not contain package github.com/spf13/cobra/cobra
`,
			expectedProblemNames: []string{"E002"},
		},
		{
			name: "E002 module found but does not contain package StevenACoffman checker",
			goInstallOutput: `
go: downloading github.com/StevenACoffman/toolbox v0.0.0-20210809155116-d52e63616b7a
go install: github.com/StevenACoffman/toolbox/cmd/checker@latest: module github.com/StevenACoffman/toolbox@latest found (v0.0.0-20210809155116-d52e63616b7a), but does not contain package github.com/StevenACoffman/toolbox/cmd/checker
`,
			expectedProblemNames: []string{"E002"},
		},
		{
			name: "E003 go.mod path mismatch",
			goInstallOutput: `
go: downloading github.com/googleapis/gnostic v0.6.6
go install: github.com/googleapis/gnostic/generate-gnostic@latest: github.com/googleapis/gnostic@v0.6.6: parsing go.mod:
	module declares its path as: github.com/google/gnostic
	        but was required as: github.com/googleapis/gnostic
`,
			expectedProblemNames: []string{"E003"},
		},
		{
			name: "E004 go.mod contains replace directive",
			goInstallOutput: `
go: downloading github.com/wagoodman/dive v0.10.0
go install: github.com/wagoodman/dive@latest (in github.com/wagoodman/dive@v0.10.0):
	The go.mod file for the module providing named packages contains one or
	more replace directives. It must not contain directives that would cause
	it to be interpreted differently than if it were the main module.
`,
			expectedProblemNames: []string{"E004"},
		},
		{
			name: "malformed module path command-line-arguments",
			goInstallOutput: `
go install: command-line-arguments@latest: malformed module path "command-line-arguments": missing dot in first path element
`,
			expectedProblemNames: []string{},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			problemsFound := FindCommonUpdateProblems(c.goInstallOutput)
			var problemNames []string
			for _, p := range problemsFound {
				problemNames = append(problemNames, p.name)
			}

			assert.ElementsMatch(t, problemNames, c.expectedProblemNames, "mismatched errors found (listA) while expected listB")
		})
	}
}

func TestCommonProblemHashes(t *testing.T) {
	workingDirectory, err := os.Getwd()
	require.Nil(t, err)
	troubleshootingFilePath := filepath.Join(workingDirectory, "..", "..", "TROUBLESHOOTING.md")
	troubleshootingFileBuffer, err := os.ReadFile(troubleshootingFilePath)
	require.Nil(t, err)
	troubleshootingFileContents := string(troubleshootingFileBuffer)

	hashRegexp := regexp.MustCompile(`^- \[.*\]\((.*)\)$`)
	var knownProblemHashes []string
	endOfLineSequence := "\n"
	if runtime.GOOS == "windows" {
		endOfLineSequence = "\r\n"
	}

	for _, line := range strings.Split(troubleshootingFileContents, endOfLineSequence) {
		matches := hashRegexp.FindStringSubmatch(line)
		if matches != nil {
			assert.Len(t, matches, 2, "hashRegexp matched an unexpected line")
			knownProblemHashes = append(knownProblemHashes, matches[1])
		}
	}

	require.GreaterOrEqualf(t, len(knownProblemHashes), 4,
		"could not find hashes in %s\nfile contents: %s",
		troubleshootingFilePath, troubleshootingFileContents)

	allCommonProblems := append(commonUpdateProblems, binaryBuiltFromSourceProblem)
	for _, p := range allCommonProblems {
		assert.Containsf(t, knownProblemHashes, p.troubleshootingHeadingHash,
			"not a known hash. Check if there is a typo in the heading hash for problem %v", p)
	}
}
