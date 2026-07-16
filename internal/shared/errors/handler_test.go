package errors

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWrapConfirmationError locks the confirmation-error contract: nil stays nil,
// a prompt interruption flows up unchanged (so main prints "cancelled" and exits
// non-zero — no os.Exit), and any other error is wrapped with context.
func TestWrapConfirmationError(t *testing.T) {
	// nil → nil.
	assert.NoError(t, WrapConfirmationError(nil, "ctx"))

	// A real error gets the context prefix.
	wrapped := WrapConfirmationError(fmt.Errorf("some error"), "test context")
	require.Error(t, wrapped)
	assert.Equal(t, "test context: some error", wrapped.Error())

	// Interruptions are returned as-is (not wrapped), so main can detect them.
	for _, e := range []error{
		errors.New("interrupted"),
		fmt.Errorf("selection failed: %w", errors.New("^C")),
		context.Canceled,
		fmt.Errorf("operation cancelled: %w", context.Canceled),
	} {
		got := WrapConfirmationError(e, "test context")
		assert.Same(t, e, got, "interruption %q must be returned unchanged", e)
	}

	// A network error that merely mentions "interrupted" is NOT an interruption —
	// it must be wrapped with context, not passed through as a cancellation.
	netErr := fmt.Errorf("connection was interrupted unexpectedly")
	gotNet := WrapConfirmationError(netErr, "network error")
	assert.Contains(t, gotNet.Error(), "network error", "non-interruption must be wrapped")
}
