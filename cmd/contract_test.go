package cmd

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// Freezes the root command's persistent-flag and top-level subcommand contract.

func TestRootContract_PersistentFlags(t *testing.T) {
	root := GetRootCmd(VersionInfo{Version: "t", Commit: "t", Date: "t"})

	verbose := root.PersistentFlags().Lookup("verbose")
	if assert.NotNil(t, verbose, "root must expose a persistent --verbose") {
		assert.Equal(t, "v", verbose.Shorthand)
		assert.Equal(t, "bool", verbose.Value.Type())
		assert.Equal(t, "false", verbose.DefValue)
	}

	silent := root.PersistentFlags().Lookup("silent")
	if assert.NotNil(t, silent, "root must expose a persistent --silent") {
		assert.Equal(t, "bool", silent.Value.Type())
		assert.Equal(t, "false", silent.DefValue)
	}
}

func TestRootContract_TopLevelSubcommands(t *testing.T) {
	root := GetRootCmd(VersionInfo{Version: "t", Commit: "t", Date: "t"})

	// Subset check (cobra may inject help/completion), so assert each is present
	// rather than an exact count. `update` is here too: it rewrites the running
	// binary, so its surface must never drift or vanish unnoticed.
	for _, name := range []string{"cluster", "app", "bootstrap", "prerequisites", "update"} {
		testutil.FindSubcommand(t, root, name)
	}
}
