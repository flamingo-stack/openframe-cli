package cluster

import (
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/utils"
	"github.com/flamingo-stack/openframe-cli/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// Freezes the public CLI contract of the `cluster` command tree.

func TestClusterContract_RootShape(t *testing.T) {
	utils.InitGlobalFlags()
	t.Cleanup(utils.ResetGlobalFlags)
	cluster := GetClusterCmd()

	assert.Equal(t, "cluster", cluster.Name())
	assert.ElementsMatch(t, []string{"k"}, cluster.Aliases, "k alias is part of the contract")

	testutil.AssertSubcommands(t, cluster, "create", "list", "delete", "status", "use", "cleanup")
}

func TestClusterContract_Flags(t *testing.T) {
	utils.InitGlobalFlags()
	t.Cleanup(utils.ResetGlobalFlags)
	cluster := GetClusterCmd()

	create := testutil.FindSubcommand(t, cluster, "create")
	testutil.AssertFlags(t, create, []testutil.FlagSpec{
		{Name: "type", Shorthand: "t", Type: "string", Default: ""},
		{Name: "nodes", Shorthand: "n", Type: "int", Default: "3"},
		{Name: "version", Type: "string", Default: ""},
		{Name: "skip-wizard", Type: "bool", Default: "false"},
	})

	list := testutil.FindSubcommand(t, cluster, "list")
	testutil.AssertFlag(t, list, testutil.FlagSpec{Name: "output", Shorthand: "o", Type: "string", Default: "text"})

	del := testutil.FindSubcommand(t, cluster, "delete")
	testutil.AssertFlag(t, del, testutil.FlagSpec{Name: "force", Shorthand: "f", Type: "bool", Default: "false"})

	status := testutil.FindSubcommand(t, cluster, "status")
	testutil.AssertFlags(t, status, []testutil.FlagSpec{
		{Name: "detailed", Shorthand: "d", Type: "bool", Default: "false"},
		{Name: "no-apps", Type: "bool", Default: "false"},
	})

	cleanup := testutil.FindSubcommand(t, cluster, "cleanup")
	assert.ElementsMatch(t, []string{"c"}, cleanup.Aliases, "cleanup keeps the c alias")
	testutil.AssertFlag(t, cleanup, testutil.FlagSpec{Name: "force", Shorthand: "f", Type: "bool", Default: "false"})
}
