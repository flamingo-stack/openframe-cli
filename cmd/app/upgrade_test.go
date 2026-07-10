package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeIsChangeRef locks the Mode 1 (change-ref) vs Mode 2 (force-sync)
// decision: a ref/branch change selects Mode 1; a bare invocation defaults to
// Mode 2. The --ref + --sync combination never reaches this function — it is
// rejected up front (F5 guard below).
func TestUpgradeIsChangeRef(t *testing.T) {
	cases := []struct {
		name          string
		refChanged    bool
		sync          bool
		wantChangeRef bool
	}{
		{"bare -> force-sync", false, false, false},
		{"--sync -> force-sync", false, true, false},
		{"--ref -> change-ref", true, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantChangeRef, upgradeIsChangeRef(tc.refChanged, tc.sync))
		})
	}
}

// TestUpgradeRejectsRefWithSync is the F5 regression guard: `upgrade --ref X
// --sync` used to silently discard --ref and force-sync the CURRENT ref — the
// user believed X was deployed. The combination must be rejected loudly.
func TestUpgradeRejectsRefWithSync(t *testing.T) {
	cmd := getUpgradeCmd()
	cmd.SetArgs([]string{"--ref=v1.2.3", "--sync"})
	err := cmd.Execute()
	require.Error(t, err, "--ref with --sync must be rejected")
	assert.Contains(t, err.Error(), "mutually exclusive")
}

// TestUpgradeCommandShape verifies the command is wired with RunE and the --sync
// flag alongside the shared install flags.
func TestUpgradeCommandShape(t *testing.T) {
	cmd := getUpgradeCmd()
	require.NotNil(t, cmd.RunE, "upgrade must have a RunE")
	assert.Equal(t, "upgrade [cluster-name]", cmd.Use)

	require.NotNil(t, cmd.Flags().Lookup("sync"), "upgrade must have --sync")
	require.NotNil(t, cmd.Flags().Lookup("ref"), "upgrade must have --ref")
	require.NotNil(t, cmd.Flags().Lookup("dry-run"), "upgrade must have --dry-run")
}

func TestClusterNameArg(t *testing.T) {
	assert.Equal(t, "", clusterNameArg(nil))
	assert.Equal(t, "", clusterNameArg([]string{}))
	assert.Equal(t, "my-cluster", clusterNameArg([]string{"my-cluster"}))
}
