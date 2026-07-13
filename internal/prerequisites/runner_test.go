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

func TestCheck_ReportsWithoutInstalling(t *testing.T) {
	present, _ := prereq("docker", true, true, false)
	missing, state := prereq("k3d", false, true, false)
	r := Runner{OS: "linux"} // even where auto-install is supported, Check must not install
	res := r.Check(Set{Items: []Prerequisite{present, missing}})

	assert.Equal(t, []string{"docker"}, res.Satisfied)
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "k3d", res.Missing[0].Name)
	assert.Equal(t, "https://docs/k3d", res.Missing[0].DocsURL)
	assert.False(t, *state, "Check must never run the installer")
	assert.False(t, res.OK())
}

func TestRun_MissingWithoutInstallerIsManual(t *testing.T) {
	missing, _ := prereq("docker", false, false, false) // not installable
	r := Runner{OS: "linux"}
	res := r.Run(context.Background(), Set{Items: []Prerequisite{missing}})
	require.False(t, res.OK())
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "https://docs/docker", res.Missing[0].DocsURL)
}

// TestCheck_DetailBecomesReason: a prerequisite that supplies a Detail (e.g.
// Docker "installed but not running") must surface it as MissingItem.Reason so
// the renderer can avoid the false "not installed" wording. A prereq without a
// Detail leaves Reason empty (genuine absence → generic wording).
func TestCheck_DetailBecomesReason(t *testing.T) {
	notRunning := Prerequisite{
		Name:        "Docker",
		DocsURL:     "https://docs/docker",
		IsSatisfied: func() bool { return false },
		Detail:      func() string { return "installed but not running" },
	}
	absent := Prerequisite{
		Name:        "k3d",
		DocsURL:     "https://docs/k3d",
		IsSatisfied: func() bool { return false },
	}

	res := Runner{}.Check(Set{Items: []Prerequisite{notRunning, absent}})
	require.Len(t, res.Missing, 2)

	byName := map[string]MissingItem{}
	for _, m := range res.Missing {
		byName[m.Name] = m
	}
	assert.Equal(t, "installed but not running", byName["Docker"].Reason,
		"Detail must flow into Reason so the tool isn't mislabeled 'not installed'")
	assert.Empty(t, byName["k3d"].Reason, "a prereq with no Detail must leave Reason empty")
}

// TestRun_DetailReasonSurvivesFailedInstall: on Linux, when an auto-install
// runs but the tool is still unsatisfied (Docker installed but daemon down),
// the Detail reason must still be attached — this is the exact WSL-Alpine case
// where apk installs docker but no OpenRC starts the daemon.
func TestRun_DetailReasonSurvivesFailedInstall(t *testing.T) {
	installedButDown := Prerequisite{
		Name:        "Docker",
		IsSatisfied: func() bool { return false }, // never becomes running
		Install:     func(context.Context) error { return nil },
		Detail:      func() string { return "installed but not running" },
	}
	// OS: "linux" forces the auto-install path (the WSL-Alpine case).
	res := Runner{OS: "linux"}.Run(context.Background(), Set{Items: []Prerequisite{installedButDown}})
	require.Len(t, res.Missing, 1)
	assert.Equal(t, "installed but not running", res.Missing[0].Reason)
}
