package intercept

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestService() (*Service, *executor.MockCommandExecutor) {
	mock := executor.NewMockCommandExecutor()
	s := NewService(mock, false)
	return s, mock
}

// TestCleanup_PerformsTeardownAndIsIdempotent verifies cleanup leaves the
// intercept, quits the daemon, restores the namespace, and is a no-op on a
// second call.
func TestCleanup_PerformsTeardownAndIsIdempotent(t *testing.T) {
	s, mock := newTestService()
	s.isIntercepting = true
	s.currentService = "my-service"
	s.currentNamespace = "dev"
	s.originalNamespace = "default"

	s.cleanup()

	assert.False(t, s.isIntercepting, "cleanup must clear the intercepting flag")
	assert.True(t, mock.WasCommandExecuted("telepresence leave my-service"))
	assert.True(t, mock.WasCommandExecuted("telepresence quit"))
	assert.True(t, mock.WasCommandExecuted("telepresence connect --namespace default"))

	countAfterFirst := mock.GetCommandCount()
	s.cleanup() // second call must do nothing
	assert.Equal(t, countAfterFirst, mock.GetCommandCount(), "cleanup must be idempotent")
}

// TestSignalHandler_RunsCleanupAndClosesDone verifies the wiring: a signal makes
// the handler run cleanup exactly once and unblock waitForInterrupt via
// cleanupDone — without os.Exit.
func TestSignalHandler_RunsCleanupAndClosesDone(t *testing.T) {
	s, mock := newTestService()
	s.isIntercepting = true
	s.currentService = "svc"

	s.setupCleanupHandler("svc")

	// Simulate Ctrl-C by delivering to the signal channel directly.
	s.signalChannel <- os.Interrupt

	done := make(chan error, 1)
	go func() { done <- s.waitForInterrupt() }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("waitForInterrupt did not return after signal-driven cleanup")
	}

	assert.True(t, mock.WasCommandExecuted("telepresence quit"))
	// cleanupDone must be closed.
	select {
	case <-s.cleanupDone:
	default:
		t.Fatal("cleanupDone was not closed by the signal handler")
	}
}

// TestNoOsExitInLibrary is the structural guard for audit I6: no library file in
// this package may call os.Exit (that belongs only in main()).
func TestNoOsExitInLibrary(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Clean(name))
		require.NoError(t, err)
		// Match the call form so explanatory comments mentioning os.Exit don't trip it.
		assert.NotContainsf(t, string(data), "os.Exit(",
			"%s must not call os.Exit (library/signal-handler code)", name)
	}
}
