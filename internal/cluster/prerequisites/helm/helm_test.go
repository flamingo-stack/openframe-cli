package helm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHelmInstaller_And_InstallHelp(t *testing.T) {
	h := NewHelmInstaller()
	require.NotNil(t, h)

	help := h.GetInstallHelp()
	assert.NotEmpty(t, help)
	assert.Contains(t, help, "helm")
	assert.Contains(t, help, "https://helm.sh", "install help must point to the official docs")
}

func TestCommandExists(t *testing.T) {
	// The Go toolchain is running these tests, so `go` is guaranteed on PATH.
	assert.True(t, commandExists("go"))
	assert.False(t, commandExists("definitely-not-a-real-command-9z8x7"))
}
