package prerequisites

import (
	"bytes"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/cluster/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPrerequisitesCmd_Structure(t *testing.T) {
	cmd := GetPrerequisitesCmd()
	require.NotNil(t, cmd)
	assert.Equal(t, "prerequisites", cmd.Name())
	assert.Contains(t, cmd.Aliases, "prereq")
	assert.Contains(t, cmd.Aliases, "prereqs")
	assert.NotEmpty(t, cmd.Short)

	sub := map[string]bool{}
	for _, c := range cmd.Commands() {
		sub[c.Name()] = true
	}
	assert.True(t, sub["check"], "must have a check subcommand")
	assert.True(t, sub["install"], "must have an install subcommand")
}

// TestSubcommands_TypeFlag verifies both subcommands carry the --type flag in
// the same shape as `cluster create` (-t shorthand) with the local default.
func TestSubcommands_TypeFlag(t *testing.T) {
	for _, c := range GetPrerequisitesCmd().Commands() {
		t.Run(c.Name(), func(t *testing.T) {
			flag := c.Flags().Lookup("type")
			require.NotNilf(t, flag, "%s must have a --type flag", c.Name())
			assert.Equal(t, "k3d", flag.DefValue)
			assert.Equal(t, "t", flag.Shorthand)
		})
	}
}

// TestSubcommands_UnknownType verifies an unknown --type fails with a clear
// error before any host checks or installs run.
func TestSubcommands_UnknownType(t *testing.T) {
	for _, name := range []string{"check", "install"} {
		t.Run(name, func(t *testing.T) {
			root := GetPrerequisitesCmd()
			root.SetOut(&bytes.Buffer{})
			root.SetErr(&bytes.Buffer{})
			root.SetArgs([]string{name, "--type", "minikube"})

			err := root.Execute()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "minikube")
			assert.Contains(t, err.Error(), "supported: k3d, eks, gke")
		})
	}
}

// installCommandFor must carry the selected --type: after `check --type eks`
// a bare `prerequisites install` would default back to k3d and install the
// wrong toolset. The default type stays unspoken to keep the common local
// command short.
func TestInstallCommandFor(t *testing.T) {
	assert.Equal(t, "openframe prerequisites install", installCommandFor(models.ClusterTypeK3d))
	assert.Equal(t, "openframe prerequisites install --type eks", installCommandFor(models.ClusterTypeEKS))
	assert.Equal(t, "openframe prerequisites install --type gke", installCommandFor(models.ClusterTypeGKE))
}

// End-to-end shape of the recovery hint: a typed check that finds tools
// missing must point at a typed install. HOME and PATH are pointed at empty
// temp dirs so every eks tool (terraform, aws — both PATH lookups) reads as
// missing regardless of the host.
func TestCheck_MissingCloudToolsPointAtTypedInstall(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	root := GetPrerequisitesCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"check", "--type", "eks"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "openframe prerequisites install --type eks",
		"the recovery command must keep the selected type, or install defaults back to k3d")
}

// The provider-name aliases must behave exactly like the canonical types —
// and the recovery hint must echo the CANONICAL type, teaching the shorter
// spelling as a side effect.
func TestCheck_AliasTypeMapsToCanonicalSet(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("PATH", t.TempDir())

	root := GetPrerequisitesCmd()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"check", "--type", "aws"})

	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "openframe prerequisites install --type eks",
		"--type aws must select the eks set and hint its canonical name")
}
