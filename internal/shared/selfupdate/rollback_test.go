package selfupdate

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChangeVerb(t *testing.T) {
	cases := []struct{ current, target, want string }{
		{"v1.0.0", "v1.1.0", "Update"},
		{"v1.2.0", "v1.1.0", "Downgrade"},
		{"v1.2.0", "v1.2.0", "Reinstall"},
		{"dev", "v1.0.0", "Switch"},
	}
	for _, c := range cases {
		if got := ChangeVerb(c.current, c.target); got != c.want {
			t.Errorf("ChangeVerb(%q,%q) = %q, want %q", c.current, c.target, got, c.want)
		}
	}
}

func TestRollbackNoPrevious(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix-only")
	}
	t.Setenv("HOME", t.TempDir()) // empty state → nothing saved
	u := Updater{Current: "v2.0.0", GOOS: "linux", GOARCH: "amd64", exePath: filepath.Join(t.TempDir(), "openframe")}
	if _, ok := PreviousVersion(); ok {
		t.Fatal("expected no previous version in a fresh state dir")
	}
	if err := u.Rollback(context.Background(), nil); err == nil {
		t.Fatal("Rollback must fail when nothing was saved")
	}
}

// TestSavePreviousAndRollback saves a rollback point, then reverts the current
// executable to it and confirms the point is consumed.
func TestSavePreviousAndRollback(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("smoke test runs a /bin/sh stub binary; unix-only")
	}
	t.Setenv("HOME", t.TempDir())

	dir := t.TempDir()
	exe := filepath.Join(dir, "openframe")
	require.NoError(t, os.WriteFile(exe, []byte("#!/bin/sh\necho v2.0.0\n"), 0o755)) // current binary

	// A stand-in "previous" binary, retained as the rollback point.
	prevSrc := filepath.Join(dir, "prev-src")
	require.NoError(t, os.WriteFile(prevSrc, []byte("#!/bin/sh\necho v1.0.0\n"), 0o755))
	require.NoError(t, savePrevious(prevSrc, "v1.0.0"))

	v, ok := PreviousVersion()
	require.True(t, ok)
	require.Equal(t, "v1.0.0", v)

	u := Updater{Current: "v2.0.0", GOOS: "linux", GOARCH: "amd64", exePath: exe}
	require.NoError(t, u.Rollback(context.Background(), nil))

	got, err := os.ReadFile(exe)
	require.NoError(t, err)
	require.True(t, strings.Contains(string(got), "echo v1.0.0"), "exe not reverted: %q", got)

	// The rollback point is consumed (one level deep).
	_, ok = PreviousVersion()
	require.False(t, ok, "rollback point should be cleared after use")
	for _, leftover := range []string{exe + ".bak", exe + ".rollback"} {
		if _, err := os.Stat(leftover); !os.IsNotExist(err) {
			t.Errorf("leftover file: %s", leftover)
		}
	}
}
