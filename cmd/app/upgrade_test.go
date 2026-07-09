package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpgradeIsChangeRef locks the Mode 1 (change-ref) vs Mode 2 (force-sync)
// decision: any ref/branch change selects Mode 1 unless --sync forces Mode 2;
// a bare invocation defaults to Mode 2.
func TestUpgradeIsChangeRef(t *testing.T) {
	cases := []struct {
		name                      string
		refChanged, branchChanged bool
		sync                      bool
		wantChangeRef             bool
	}{
		{"bare -> force-sync", false, false, false, false},
		{"--sync -> force-sync", false, false, true, false},
		{"--ref -> change-ref", true, false, false, true},
		{"--github-branch -> change-ref", false, true, false, true},
		{"--ref with --sync -> force-sync", true, false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.wantChangeRef, upgradeIsChangeRef(tc.refChanged, tc.branchChanged, tc.sync))
		})
	}
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

// TestDeploymentModeForUpgrade covers OSS auto-inference and the SaaS-requires-flag
// guard for Mode-1 upgrades.
func TestDeploymentModeForUpgrade(t *testing.T) {
	t.Run("no helm-values → auto oss-tenant", func(t *testing.T) {
		t.Chdir(t.TempDir())
		mode, err := deploymentModeForUpgrade()
		require.NoError(t, err)
		assert.Equal(t, "oss-tenant", mode)
	})

	t.Run("oss-enabled values → oss-tenant", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "helm-values.yaml"),
			[]byte("deployment:\n  oss:\n    enabled: true\n"), 0o600))
		t.Chdir(dir)
		mode, err := deploymentModeForUpgrade()
		require.NoError(t, err)
		assert.Equal(t, "oss-tenant", mode)
	})

	t.Run("saas-enabled values → require explicit flag", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "helm-values.yaml"),
			[]byte("deployment:\n  saas:\n    enabled: true\n"), 0o600))
		t.Chdir(dir)
		_, err := deploymentModeForUpgrade()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deployment-mode")
	})
}
