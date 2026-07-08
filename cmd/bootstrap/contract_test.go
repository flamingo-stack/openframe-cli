package bootstrap

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// Freezes the public CLI contract of `bootstrap`. The CLI supports only the OSS
// (oss-tenant) deployment, so `bootstrap --non-interactive` reuses the existing
// helm-values.yaml with no deployment-mode flag.

func TestBootstrapContract_Flags(t *testing.T) {
	cmd := GetBootstrapCmd()

	assert.Equal(t, "bootstrap", cmd.Name())
	testutil.AssertFlags(t, cmd, []testutil.FlagSpec{
		{Name: "non-interactive", Type: "bool", Default: "false"},
		// verbose/-v is now inherited from the root persistent flag, not local.
	})
	assert.Nil(t, cmd.Flags().Lookup("deployment-mode"), "--deployment-mode must be removed")
}

func TestBootstrapContract_AcceptsAtMostOneArg(t *testing.T) {
	cmd := GetBootstrapCmd()

	assert.NotNil(t, cmd.RunE, "bootstrap must be wired to a RunE")
	assert.NoError(t, cmd.Args(cmd, []string{"one"}), "a single cluster-name arg is allowed")
	assert.Error(t, cmd.Args(cmd, []string{"one", "two"}), "more than one positional arg must be rejected")
}

// TestBootstrapContract_RejectsUnsafeClusterName drives the command end-to-end.
// This is now possible because the error path RETURNS instead of calling
// os.Exit (P1) — previously Execute() killed the test process here.
func TestBootstrapContract_RejectsUnsafeClusterName(t *testing.T) {
	cmd := GetBootstrapCmd()
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"bad;rm -rf /"})

	err := cmd.Execute()
	assert.Error(t, err, "an unsafe cluster name must be rejected before Execute reaches the cluster")
}
