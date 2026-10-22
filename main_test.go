package main

import (
	"context"
	"log/slog"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLoggerLevel(t *testing.T) {
	ctx := context.Background()

	assert.False(t, newLogger(false, false).Enabled(ctx, slog.LevelDebug))
	assert.True(t, newLogger(true, false).Enabled(ctx, slog.LevelDebug))
}

func TestLogColorsEnabled(t *testing.T) {
	env := func(vars map[string]string) func(string) string {
		return func(key string) string { return vars[key] }
	}

	assert.True(t, logColorsEnabled(false, true, env(nil)), "terminal")
	assert.False(t, logColorsEnabled(false, false, env(nil)), "not a terminal")
	assert.False(t, logColorsEnabled(false, true, env(map[string]string{"NO_COLOR": "1"})), "NO_COLOR set")
	assert.True(t, logColorsEnabled(false, true, env(map[string]string{"NO_COLOR": ""})), "NO_COLOR empty")
	assert.False(t, logColorsEnabled(false, true, env(map[string]string{"TERM": "dumb"})), "dumb terminal")
	assert.True(t, logColorsEnabled(true, false, env(map[string]string{"NO_COLOR": "1"})), "--colors forces colors")
}

func TestVersionFromBuildInfo(t *testing.T) {
	tests := []struct {
		name     string
		info     *debug.BuildInfo
		ok       bool
		expected string
	}{
		{
			name:     "no build info",
			info:     nil,
			ok:       false,
			expected: "(devel)",
		},
		{
			name:     "empty main module version",
			info:     &debug.BuildInfo{Main: debug.Module{Version: ""}},
			ok:       true,
			expected: "(devel)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, versionFromBuildInfo(tt.info, tt.ok))
		})
	}
}
