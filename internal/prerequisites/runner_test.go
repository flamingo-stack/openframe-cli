package prerequisites

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// prereq is a small test helper: `installed` flips to true once Install runs.
func prereq(name string, present bool, installable bool, installFails bool) (Prerequisite, *bool) {
	state := present
	p := Prerequisite{
		Name:        name,
		DocsURL:     "https://docs/" + name,
		IsSatisfied: func() bool { return state },
	}
	if installable {
		p.Install = func(context.Context) error {
			if installFails {
				return errors.New("install failed")
			}
			state = true
			return nil
		}
	}
	return p, &state
}

func TestRun_AllSatisfied(t *testing.T) {
	a, _ := prereq("docker", true, true, false)
	b, _ := prereq("helm", true, true, false)
	r := Runner{OS: "linux"}
	res := r.Run(context.Background(), Set{Name: "cluster", Items: []Prerequisite{a, b}})
	assert.True(t, res.OK())
	assert.ElementsMatch(t, []string{"docker", "helm"}, res.Satisfied)
	assert.Empty(t, res.Installed)
}

func TestRun_AutoInstallsOnLinux(t *testing.T) {
	missing, state := prereq("k3d", false, true, false)
	r := Runner{OS: "linux"}
	res := r.Run(context.Background(), Set{Items: []Prerequisite{missing}})
	assert.True(t, res.OK(), "should auto-install and succeed")
	assert.Equal(t, []string{"k3d"}, res.Installed)
	assert.True(t, *state, "installer must have run")
}

func TestRun_InstallFailureIsMissing(t *testing.T) {
	missing, _ := prereq("k3d", false, true, true) // install fails
	r := Runner{OS: "darwin"}
	res := r.Run(context.Background(), Set{Items: []Prerequisite{missing}})
	require.False(t, res.OK())
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "k3d", res.Missing[0].Name)
	assert.Error(t, res.Missing[0].Err)
	assert.Equal(t, "https://docs/k3d", res.Missing[0].DocsURL)
}

func TestRun_WindowsNeverAutoInstalls(t *testing.T) {
	missing, state := prereq("docker", false, true, false)
	r := Runner{OS: "windows"}
	assert.False(t, r.AutoInstalls())
	res := r.Run(context.Background(), Set{Items: []Prerequisite{missing}})
	require.False(t, res.OK())
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "https://docs/docker", res.Missing[0].DocsURL, "Windows must point to docs")
	assert.NoError(t, res.Missing[0].Err, "no install attempted, so no error")
	assert.False(t, *state, "installer must NOT run on Windows")
}

func TestRun_MissingWithoutInstallerIsManual(t *testing.T) {
	missing, _ := prereq("docker", false, false, false) // not installable
	r := Runner{OS: "linux"}
	res := r.Run(context.Background(), Set{Items: []Prerequisite{missing}})
	require.False(t, res.OK())
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "https://docs/docker", res.Missing[0].DocsURL)
}
