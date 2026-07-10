package executor

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDryRun_PrintsCommandWithoutVerbose is the B6 guard for dry-run
// visibility: `--dry-run` must print the command it WOULD run even without
// --verbose. Previously a non-verbose dry-run produced no output at all and
// exited 0 — indistinguishable from a real successful run.
func TestDryRun_PrintsCommandWithoutVerbose(t *testing.T) {
	var buf bytes.Buffer
	old := pterm.Info
	pterm.Info = *pterm.Info.WithWriter(&buf)
	t.Cleanup(func() { pterm.Info = old })

	exec := NewRealCommandExecutor(true, false) // dry-run, NOT verbose
	result, err := exec.Execute(context.Background(), "k3d", "cluster", "delete", "test")
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)

	out := buf.String()
	assert.Contains(t, out, "Would run:", "dry-run must announce the command")
	assert.Contains(t, out, "k3d cluster delete test", "dry-run must show the full command line")
}

// TestDryRun_RedactsSecrets: the announced command line goes through the
// redactor like every other print.
func TestDryRun_RedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	old := pterm.Info
	pterm.Info = *pterm.Info.WithWriter(&buf)
	t.Cleanup(func() { pterm.Info = old })

	exec := NewRealCommandExecutor(true, false)
	_, err := exec.Execute(context.Background(), "git", "clone", "https://user:supersecret@github.com/org/repo.git")
	require.NoError(t, err)

	if strings.Contains(buf.String(), "supersecret") {
		t.Fatalf("dry-run output leaks a URL credential:\n%s", buf.String())
	}
}
