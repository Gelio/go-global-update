package gocli

import (
	"log/slog"
	"os/exec"
	"strings"
)

type GoCmdRunner interface {
	RunGoCommand(args ...string) (string, error)
}

type RealGoCmdRunner struct {
	logger *slog.Logger
}

func NewCmdRunner(logger *slog.Logger) RealGoCmdRunner {
	return RealGoCmdRunner{
		logger,
	}
}

func (runner *RealGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	bytes, err := cmd.CombinedOutput()

	attrs := []any{"args", args, "output", string(bytes)}
	if err != nil {
		attrs = append(attrs, "error", err)
	}

	runner.logger.Debug("go command output", attrs...)

	// NOTE: always include output for displaying errors
	return strings.TrimSpace(string(bytes)), err
}
