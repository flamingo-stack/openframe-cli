package app

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

func init() {
	testutil.InitializeTestMode()
}

func TestChartRootCommand(t *testing.T) {
	// Test the root chart command - using basic structure test since TestClusterCommand
	// is designed specifically for cluster commands
	cmd := GetAppCmd()

	// Test basic structure
	assert.Equal(t, "app", cmd.Name(), "Command name should match")
	assert.NotEmpty(t, cmd.Short, "Command should have short description")
	assert.NotEmpty(t, cmd.Long, "Command should have long description")
	assert.NotNil(t, cmd.RunE, "App root command should have RunE function")

	// The "chart" alias was removed — only "openframe app" is supported.
	assert.Empty(t, cmd.Aliases, "app must have no aliases")

	// Test that help contains expected content
	assert.Contains(t, cmd.Short, "OpenFrame application")
	assert.Contains(t, cmd.Long, "Install the OpenFrame application")
}
