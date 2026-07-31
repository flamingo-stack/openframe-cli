package app

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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

// Cobra runs only the CLOSEST parent's PersistentPreRunE, so the app group's
// hook shadows the root's — it must apply --verbose itself, or every `app`
// subcommand has zero debug output while `bootstrap --verbose` prints it all.
func TestAppGroup_PersistentPreRunEHonorsVerbose(t *testing.T) {
	t.Cleanup(pterm.DisableDebugMessages)
	dummy := &cobra.Command{Use: "status"}
	dummy.Flags().Bool("silent", false, "")
	dummy.Flags().Bool("verbose", true, "")
	dummy.Flags().String("output", "json", "") // machine mode: skip the logo
	if err := GetAppCmd().PersistentPreRunE(dummy, nil); err != nil {
		t.Fatal(err)
	}
	if !pterm.PrintDebugMessages {
		t.Fatal("--verbose must survive the app group's shadowing hook")
	}
}
