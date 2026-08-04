package cluster

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func init() {
	testutil.InitializeTestMode()
}

func TestClusterRootCommand(t *testing.T) {
	// Test the root cluster command (no setup needed for root command)
	testutil.TestClusterCommand(t, "cluster", GetClusterCmd, nil, nil)
}

// Cobra runs only the CLOSEST parent's PersistentPreRunE, so the cluster
// group's hook shadows the root's — it must apply --verbose itself, or every
// `cluster` subcommand has zero debug output.
func TestClusterGroup_PersistentPreRunEHonorsVerbose(t *testing.T) {
	t.Cleanup(pterm.DisableDebugMessages)
	dummy := &cobra.Command{Use: "status"}
	dummy.Flags().Bool("silent", false, "")
	dummy.Flags().Bool("verbose", true, "")
	dummy.Flags().String("output", "json", "") // machine mode: skip logo and gates
	if err := GetClusterCmd().PersistentPreRunE(dummy, nil); err != nil {
		t.Fatal(err)
	}
	if !pterm.PrintDebugMessages {
		t.Fatal("--verbose must survive the cluster group's shadowing hook")
	}
}

// delete and cleanup are type-aware like create: a cloud cluster needs
// terraform, not Docker/k3d, and the generic gate may install k3d/helm as a
// side effect — so they must bypass the group-level gate and run their own
// after DetectClusterType. Unknown (future) subcommands keep the generic gate.
func TestClusterGroup_TypeAwareCommandsSkipGenericGate(t *testing.T) {
	for _, name := range []string{"create", "use", "status", "list", "delete", "cleanup"} {
		if !skipsGenericPrerequisiteGate(name) {
			t.Errorf("%s must skip the generic k3d prerequisite gate", name)
		}
	}
	if skipsGenericPrerequisiteGate("some-future-command") {
		t.Error("unknown subcommands must default to the generic gate")
	}
}
