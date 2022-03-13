package gocli

import (
	"os/exec"
	"strings"
)

type GoCmdRunner interface {
	RunGoCommand(args ...string) (string, error)
}

type RealGoCmdRunner struct{}

func (runner *RealGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	bytes, err := cmd.CombinedOutput()

	// NOTE: always include output for displaying errors
	return strings.TrimSpace(string(bytes)), err
}
