package utils

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/errors"
	"github.com/spf13/cobra"
)

// TestWrapCommandWithCommonSetup_ErrorYieldsNonNil locks the exit-code fix: any
// error from the wrapped RunF must surface as a non-nil (handled) error so the
// process exits non-zero. Previously non-matching errors returned nil → exit 0
// despite a failure.
func TestWrapCommandWithCommonSetup_ErrorYieldsNonNil(t *testing.T) {
	InitGlobalFlags()
	t.Cleanup(ResetGlobalFlags)

	run := WrapCommandWithCommonSetup(func(*cobra.Command, []string) error {
		return fmt.Errorf("some arbitrary failure not matched by any string")
	})

	err := run(&cobra.Command{Use: "x"}, nil)
	if err == nil {
		t.Fatal("a failing command must return a non-nil error (non-zero exit)")
	}
	var handled *errors.AlreadyHandledError
	if !stderrors.As(err, &handled) {
		t.Fatalf("wrapped error must be marked handled, got %T", err)
	}
}

func TestWrapCommandWithCommonSetup_NilStaysNil(t *testing.T) {
	InitGlobalFlags()
	t.Cleanup(ResetGlobalFlags)

	run := WrapCommandWithCommonSetup(func(*cobra.Command, []string) error { return nil })
	if err := run(&cobra.Command{Use: "x"}, nil); err != nil {
		t.Fatalf("a successful command must return nil, got %v", err)
	}
}

// TestWrapCommandWithCommonSetup_PreservesSentinel: an error already displayed
// inside the RunF (returned as AlreadyHandledError) must propagate unchanged, so
// main does not re-print it.
func TestWrapCommandWithCommonSetup_PreservesSentinel(t *testing.T) {
	InitGlobalFlags()
	t.Cleanup(ResetGlobalFlags)

	orig := fmt.Errorf("already shown")
	run := WrapCommandWithCommonSetup(func(*cobra.Command, []string) error {
		return &errors.AlreadyHandledError{OriginalError: orig}
	})

	err := run(&cobra.Command{Use: "x"}, nil)
	if !stderrors.Is(err, orig) {
		t.Fatalf("sentinel must be preserved, got %v", err)
	}
}
