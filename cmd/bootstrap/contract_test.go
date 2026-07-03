package bootstrap

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// Freezes the public CLI contract of `bootstrap`. The flags below are the
// hard-contract entrypoint (`bootstrap --deployment-mode=oss-tenant
// --non-interactive` must keep working at every step of the refactor).

func TestBootstrapContract_Flags(t *testing.T) {
	cmd := GetBootstrapCmd()

	assert.Equal(t, "bootstrap", cmd.Name())
	testutil.AssertFlags(t, cmd, []testutil.FlagSpec{
		{Name: "deployment-mode", Type: "string", Default: ""},
		{Name: "non-interactive", Type: "bool", Default: "false"},
		{Name: "verbose", Shorthand: "v", Type: "bool", Default: "false"},
	})
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
