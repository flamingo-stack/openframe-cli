package executor

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/redact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout runs fn while capturing everything written to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	fn()
	_ = w.Close()
	os.Stdout = orig
	return <-done
}

// TestVerboseLogging_RedactsRegisteredSecret is the I4 wiring guard: a registered
// secret appearing in a command must not be printed by verbose command logging.
func TestVerboseLogging_RedactsRegisteredSecret(t *testing.T) {
	redact.ClearSecrets()
	defer redact.ClearSecrets()

	secret := "ghp_executorVerboseSecret123"
	redact.RegisterSecret(secret)

	// dry-run + verbose hits the "Would run" log path without executing anything.
	exec := NewRealCommandExecutor(true, true)

	out := captureStdout(t, func() {
		_, err := exec.Execute(context.Background(), "git", "clone", "https://x-access-token:"+secret+"@github.com/org/repo")
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Would run:", "verbose dry-run should log the command")
	assert.NotContains(t, out, secret, "registered secret must be redacted from verbose output")
	assert.Contains(t, out, "***", "redaction marker expected")
}
