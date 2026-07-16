package update

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests freeze the public CLI contract of the `update` command tree.
// This surface is T0-critical — it replaces the running binary — so a renamed
// flag, a dropped shorthand, or a vanished subcommand must fail loudly here
// (audit B6: cmd/update previously had zero contract coverage).

func TestUpdateContract_RootShape(t *testing.T) {
	cmd := GetUpdateCmd("v1.0.0")

	assert.Equal(t, "update", cmd.Name())
	assert.Equal(t, "update [version]", cmd.Use, "the optional [version] arg (switch to a specific release) is part of the contract")
	require.NotNil(t, cmd.RunE, "update must have a RunE")

	testutil.AssertSubcommands(t, cmd, "check", "rollback")

	testutil.AssertFlags(t, cmd, []testutil.FlagSpec{
		{Name: "yes", Shorthand: "y", Type: "bool", Default: "false"},
		{Name: "force", Type: "bool", Default: "false"},
	})
}

func TestUpdateContract_CheckFlags(t *testing.T) {
	check := testutil.FindSubcommand(t, GetUpdateCmd("v1.0.0"), "check")

	require.NotNil(t, check.RunE, "check must have a RunE")
	testutil.AssertFlags(t, check, []testutil.FlagSpec{
		{Name: "output", Shorthand: "o", Type: "string", Default: "text"},
	})
}

func TestUpdateContract_RollbackFlags(t *testing.T) {
	rollback := testutil.FindSubcommand(t, GetUpdateCmd("v1.0.0"), "rollback")

	require.NotNil(t, rollback.RunE, "rollback must have a RunE")
	testutil.AssertFlags(t, rollback, []testutil.FlagSpec{
		{Name: "yes", Shorthand: "y", Type: "bool", Default: "false"},
	})
}
