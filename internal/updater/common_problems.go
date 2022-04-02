package updater

import (
	"fmt"
	"regexp"

	"github.com/Gelio/go-global-update/internal/colors"
	"github.com/fatih/color"
)

var troubleshootingBaseURL string = "https://github.com/Gelio/go-global-update/blob/main/TROUBLESHOOTING.md"

type CommonUpdateProblem struct {
	troubleshootingHeadingHash string
	name                       string
	occurs                     func(goInstallOutput string) bool
}

func (p *CommonUpdateProblem) troubleshootingURL() string {
	return troubleshootingBaseURL + p.troubleshootingHeadingHash
}

var binaryBuiltFromSourceProblem CommonUpdateProblem = CommonUpdateProblem{
	troubleshootingHeadingHash: "#e001---binaries-built-from-source",
	name:                       "E001",
}

func (p *CommonUpdateProblem) String(f *colors.DecoratorFactory) string {
	errorNameFormatter := f.NewDecorator(color.Bold)
	urlFormatter := f.NewDecorator(color.Faint)
	return fmt.Sprintf(`    This seems like a known problem %s.
    See %s
    for more information.`, errorNameFormatter(p.name), urlFormatter(p.troubleshootingURL()))
}

var commonUpdateProblems []CommonUpdateProblem = []CommonUpdateProblem{
	// E001 is handled out-of-band since the update is not even attempted
	// if E001 is detected.

	{
		name:                       "E002",
		troubleshootingHeadingHash: "#e002---module-found-but-does-not-contain-package",
		occurs: func(goInstallOutput string) bool {
			r := regexp.MustCompile("module .* found .*, but does not contain package .*")
			return r.MatchString(goInstallOutput)
		},
	},

	{
		name:                       "E003",
		troubleshootingHeadingHash: "#e003---module-declares-its-path-as--but-was-required-as-",
		occurs: func(goInstallOutput string) bool {
			r := regexp.MustCompile("(?s)module declares its path as: .*\\n\\s+but was required as: .*")
			return r.MatchString(goInstallOutput)
		},
	},

	{
		name:                       "E004",
		troubleshootingHeadingHash: "#e004---gomod-contains-replace-directives",
		occurs: func(goInstallOutput string) bool {
			r := regexp.MustCompile("(?s)The go.mod file .* contains .* replace directives.")
			return r.MatchString(goInstallOutput)
		},
	},
}

func FindCommonUpdateProblems(goInstallOutput string) []CommonUpdateProblem {
	var problems []CommonUpdateProblem

	for _, p := range commonUpdateProblems {
		if p.occurs(goInstallOutput) {
			problems = append(problems, p)
		}
	}

	return problems
}
