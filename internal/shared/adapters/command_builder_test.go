package adapters

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommandBuilder_BuildsCommandWithFieldsAndFlags(t *testing.T) {
	cmd := NewCommandBuilder("thing [name]", "manage things").
		Long("long description").
		Aliases([]string{"t"}).
		Args(cobra.MaximumNArgs(1)).
		AddBoolFlag("force", "f", false, "force it").
		AddStringFlag("repo", "r", "default-repo", "repo url").
		AddIntFlag("nodes", "n", 3, "node count").
		Build()

	assert.Equal(t, "thing [name]", cmd.Use)
	assert.Equal(t, "manage things", cmd.Short)
	assert.Equal(t, "long description", cmd.Long)
	assert.Contains(t, cmd.Aliases, "t")

	require.NotNil(t, cmd.Flags().Lookup("force"))
	require.NotNil(t, cmd.Flags().Lookup("repo"))
	require.NotNil(t, cmd.Flags().Lookup("nodes"))

	repo, _ := cmd.Flags().GetString("repo")
	assert.Equal(t, "default-repo", repo)
	nodes, _ := cmd.Flags().GetInt("nodes")
	assert.Equal(t, 3, nodes)
}

func TestFlagExtractor_ReadsFlags(t *testing.T) {
	cmd := NewCommandBuilder("x", "x").
		AddBoolFlag("force", "f", false, "").
		AddStringFlag("repo", "r", "", "").
		AddIntFlag("nodes", "n", 0, "").
		Build()

	require.NoError(t, cmd.Flags().Set("force", "true"))
	require.NoError(t, cmd.Flags().Set("repo", "acme/repo"))
	require.NoError(t, cmd.Flags().Set("nodes", "5"))

	fe := NewFlagExtractor(cmd)

	b, err := fe.GetBool("force")
	require.NoError(t, err)
	assert.True(t, b)

	s, err := fe.GetString("repo")
	require.NoError(t, err)
	assert.Equal(t, "acme/repo", s)

	n, err := fe.GetInt("nodes")
	require.NoError(t, err)
	assert.Equal(t, 5, n)

	assert.True(t, fe.FlagChanged("force"))
	assert.False(t, fe.FlagChanged("missing"))
}

func TestValidationResult(t *testing.T) {
	vr := NewValidationResult()
	assert.True(t, vr.IsValid)
	assert.False(t, vr.HasErrors())
	assert.Nil(t, vr.GetFirstError())

	first := errors.New("first")
	vr.AddError(first)
	vr.AddError(errors.New("second"))
	assert.False(t, vr.IsValid)
	assert.True(t, vr.HasErrors())
	assert.Equal(t, first, vr.GetFirstError())
	assert.Len(t, vr.Errors, 2)
}

func TestExampleBuilder(t *testing.T) {
	out := NewExampleBuilder().
		Add("create a cluster", "cluster create").
		Add("delete a cluster", "cluster delete").
		Build("openframe")

	assert.Contains(t, out, "Examples:")
	assert.Contains(t, out, "openframe")
	assert.Contains(t, out, "cluster create")
	assert.Contains(t, out, "create a cluster")

	assert.Empty(t, NewExampleBuilder().Build("openframe"), "no examples → empty string")
}

func TestPreRunEChain_ExecutesInOrder(t *testing.T) {
	var order []string
	err := NewPreRunEChain().
		Add(func(*cobra.Command, []string) error { order = append(order, "a"); return nil }).
		Add(func(*cobra.Command, []string) error { order = append(order, "b"); return nil }).
		Execute(nil, nil)

	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, order)
}

func TestPreRunEChain_StopsOnFirstError(t *testing.T) {
	var ran []string
	chain := NewPreRunEChain().
		Add(func(*cobra.Command, []string) error { ran = append(ran, "a"); return errors.New("stop") }).
		Add(func(*cobra.Command, []string) error { ran = append(ran, "b"); return nil })

	err := chain.Build()(nil, nil) // Build() returns Execute
	require.Error(t, err)
	assert.Equal(t, []string{"a"}, ran, "second function must not run after an error")
}

func TestFlagErrors(t *testing.T) {
	assert.Contains(t, (&RequiredFlagError{FlagName: "deployment-mode"}).Error(), "deployment-mode")
	assert.Contains(t, (&InvalidFlagError{FlagName: "nodes", Reason: "must be positive"}).Error(), "must be positive")
}
