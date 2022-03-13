package gocli

import (
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

type GoCmdRunner interface {
	RunGoCommand(args ...string) (string, error)
}

type RealGoCmdRunner struct {
	logger *zap.Logger
}

func NewCmdRunner(logger *zap.Logger) RealGoCmdRunner {
	return RealGoCmdRunner{
		logger,
	}
}

func (runner *RealGoCmdRunner) RunGoCommand(args ...string) (string, error) {
	cmd := exec.Command("go", args...)
	bytes, err := cmd.CombinedOutput()

	runner.logger.Debug(
		"go command output",
		zap.Strings("args", args),
		zap.String("output", string(bytes)),
		zap.Error(err),
	)

	// NOTE: always include output for displaying errors
	return strings.TrimSpace(string(bytes)), err
}
