package prerequisites

import (
	"bytes"
	"testing"

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
			assert.Contains(t, err.Error(), "k3d, eks, or gke")
		})
	}
}
