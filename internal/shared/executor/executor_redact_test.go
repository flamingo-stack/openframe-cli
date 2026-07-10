package executor

import (
	"bytes"
	"context"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVerboseLogging_RedactsRegisteredSecret is the I4 wiring guard: a registered
// secret appearing in a command must not be printed by verbose command logging.
func TestVerboseLogging_RedactsRegisteredSecret(t *testing.T) {
	redact.ClearSecrets()
	defer redact.ClearSecrets()

	secret := "ghp_executorVerboseSecret123"
	redact.RegisterSecret(secret)

	// dry-run hits the "Would run" log path without executing anything. The
	// line now goes through pterm.Info (so --silent can suppress it), so the
	// capture swaps pterm's writer rather than os.Stdout.
	var buf bytes.Buffer
	old := pterm.Info
	pterm.Info = *pterm.Info.WithWriter(&buf)
	t.Cleanup(func() { pterm.Info = old })

	exec := NewRealCommandExecutor(true, true)
	_, err := exec.Execute(context.Background(), "git", "clone", "https://x-access-token:"+secret+"@github.com/org/repo")
	require.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "Would run:", "dry-run should log the command")
	assert.NotContains(t, out, secret, "registered secret must be redacted from verbose output")
	assert.Contains(t, out, "***", "redaction marker expected")
}
