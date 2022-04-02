package colors

import (
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisabledByDefault(t *testing.T) {
	c := NewFactory(false)
	output := c.NewDecorator(color.BgBlack)("test")

	assert.Equal(t, "test", output)
}

func TestForceEnableColors(t *testing.T) {
	c := NewFactory(true)
	output := c.NewDecorator(color.BgBlack)("test")

	assert.NotEqual(t, "test", output)
}

func TestColorsWhenNoColorIsDisabled(t *testing.T) {
	require.True(t, color.NoColor, "NoColor should be enabled by default")
	color.NoColor = false
	defer (func() { color.NoColor = true })()

	c := NewFactory(false)
	output := c.NewDecorator(color.BgBlack)("test")

	assert.NotEqual(t, "test", output)
}
