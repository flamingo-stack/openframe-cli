package errors

import (
	stderrors "errors"
	"fmt"
	"testing"
)

// These lock the P1 contract: HandleGlobalError never calls os.Exit — it
// displays the error and RETURNS an AlreadyHandledError sentinel, so the RunE
// caller's deferred cleanup runs and cobra/main derive the exit code from the
// returned error. (If this regressed to os.Exit, these tests would kill the
// test binary and the package would report failure.)

func TestHandleGlobalError_ReturnsSentinelNotExit(t *testing.T) {
	orig := fmt.Errorf("boom")
	err := HandleGlobalError(orig, false)
	if err == nil {
		t.Fatal("expected a non-nil error so the exit code is non-zero")
	}

	var handled *AlreadyHandledError
	if !stderrors.As(err, &handled) {
		t.Fatalf("expected an AlreadyHandledError sentinel, got %T", err)
	}
	if !stderrors.Is(err, orig) {
		t.Error("sentinel must wrap the original error (errors.Is)")
	}
}

func TestHandleGlobalError_Nil(t *testing.T) {
	if err := HandleGlobalError(nil, false); err != nil {
		t.Fatalf("nil in must give nil out, got %v", err)
	}
}

func TestHandleGlobalError_InterruptionAlsoReturnsSentinel(t *testing.T) {
	// The interruption branch previously os.Exit(1); it must now also return a
	// sentinel so cleanup runs.
	err := HandleGlobalError(fmt.Errorf("operation cancelled by user"), true)
	var handled *AlreadyHandledError
	if !stderrors.As(err, &handled) {
		t.Fatalf("interruption must return a sentinel, got %T", err)
	}
}

// TestHandleError_InterruptionDoesNotExit guards the display path (used by the
// command wrapper): an interruption error must not terminate the process.
func TestHandleError_InterruptionDoesNotExit(t *testing.T) {
	NewErrorHandler(false).HandleError(fmt.Errorf("interrupt"))
	// Reaching here proves no os.Exit happened.
}
