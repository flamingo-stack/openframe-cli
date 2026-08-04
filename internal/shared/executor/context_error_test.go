package executor

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A context-killed child dies with "signal: killed" — the exec error alone
// never wraps ctx.Err(). These tests pin the executor joining the context
// error into the CommandError cause, because downstream both the Ctrl-C
// detection (errors.Is context.Canceled) and the retry classifier (errors.Is
// context.DeadlineExceeded) depend on it being in the chain.

func TestExecuteWithOptions_TimeoutSurfacesDeadlineExceeded(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on the sleep binary")
	}
	e := NewRealCommandExecutor(false, false)

	_, err := e.ExecuteWithOptions(context.Background(), ExecuteOptions{
		Command: "sleep",
		Args:    []string{"5"},
		Timeout: 50 * time.Millisecond,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded),
		"timeout must be visible via errors.Is, got: %v", err)
	var cmdErr *CommandError
	assert.True(t, errors.As(err, &cmdErr), "still a CommandError for exit-code fidelity")
}

func TestExecuteWithOptions_CancelSurfacesCanceled(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("relies on the sleep binary")
	}
	e := NewRealCommandExecutor(false, false)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := e.Execute(ctx, "sleep", "5")

	require.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled),
		"cancellation must be visible via errors.Is, got: %v", err)
}
