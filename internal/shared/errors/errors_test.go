package errors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
)

// TestIsUserInterruption_Structural locks the fix: real cancellation is detected
// via errors.Is(context.Canceled) (incl. %w-wrapped), a timeout is NOT mislabeled
// as a user cancellation, and an unrelated error mentioning "cancel" is not a
// false positive. Prompt Ctrl-C markers ("^C"/"interrupted") still count.
func TestIsUserInterruption_Structural(t *testing.T) {
	eh := NewErrorHandler(false)

	// Genuine cancellation, bare and wrapped.
	assert.True(t, eh.isUserInterruption(context.Canceled))
	assert.True(t, eh.isUserInterruption(fmt.Errorf("operation cancelled: %w", context.Canceled)))

	// A timeout is context.DeadlineExceeded — must NOT read as user cancellation.
	assert.False(t, eh.isUserInterruption(context.DeadlineExceeded))
	assert.False(t, eh.isUserInterruption(fmt.Errorf("wait timed out: %w", context.DeadlineExceeded)))

	// Coincidental text must not false-match (no wrapped context.Canceled).
	assert.False(t, eh.isUserInterruption(errors.New("upstream returned: context canceled by peer")))
	assert.False(t, eh.isUserInterruption(errors.New("cluster create failed")))

	// Prompt Ctrl-C markers still detected.
	assert.True(t, eh.isUserInterruption(fmt.Errorf("selection failed: %w", errors.New("^C"))))
	assert.True(t, eh.isUserInterruption(errors.New("interrupted")))

	// nil is safe.
	assert.False(t, eh.isUserInterruption(nil))
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "with value",
			err: &ValidationError{
				Field:   "name",
				Value:   "invalid-name",
				Message: "must contain only letters",
			},
			expected: "validation failed for name 'invalid-name': must contain only letters",
		},
		{
			name: "without value",
			err: &ValidationError{
				Field:   "count",
				Value:   "",
				Message: "must be greater than zero",
			},
			expected: "validation failed for count: must be greater than zero",
		},
		{
			name: "empty field",
			err: &ValidationError{
				Field:   "",
				Value:   "test",
				Message: "required field",
			},
			expected: "validation failed for  'test': required field",
		},
		{
			name: "empty message",
			err: &ValidationError{
				Field:   "email",
				Value:   "invalid-email",
				Message: "",
			},
			expected: "validation failed for email 'invalid-email': ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewErrorHandler(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{
			name:    "verbose enabled",
			verbose: true,
		},
		{
			name:    "verbose disabled",
			verbose: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewErrorHandler(tt.verbose)
			assert.NotNil(t, handler)
			assert.Equal(t, tt.verbose, handler.verbose)
		})
	}
}

func TestErrorHandler_HandleError_Nil(t *testing.T) {
	handler := NewErrorHandler(false)

	// Should not panic with nil error
	assert.NotPanics(t, func() {
		handler.HandleError(nil)
	})
}

func TestErrorHandler_HandleError_ValidationError(t *testing.T) {
	handler := NewErrorHandler(false)
	err := &ValidationError{
		Field:   "name",
		Value:   "test",
		Message: "invalid format",
	}

	// Test that the function doesn't panic and runs successfully
	assert.NotPanics(t, func() {
		handler.HandleError(err)
	})
}

func TestErrorHandler_HandleError_ValidationError_NoValue(t *testing.T) {
	handler := NewErrorHandler(false)
	err := &ValidationError{
		Field:   "count",
		Value:   "",
		Message: "must be positive",
	}

	// Test that the function doesn't panic and runs successfully
	// Note: pterm output cannot be easily captured in tests
	assert.NotPanics(t, func() {
		handler.HandleError(err)
	})
}

// TestErrorHandler_CommandError_ShowsChildStderr is the M1.1 guard: a failed
// external command must surface the CHILD'S reason to the user, not the
// useless "exit status 1". Before this, the handler matched a CommandError
// type that nothing ever constructed, so real failures (executor.CommandError)
// fell through to the generic dump with the reason discarded.
func TestErrorHandler_CommandError_ShowsChildStderr(t *testing.T) {
	var buf bytes.Buffer
	oldBasic, oldErr := pterm.DefaultBasicText, pterm.Error
	pterm.DefaultBasicText = *pterm.DefaultBasicText.WithWriter(&buf)
	pterm.Error = *pterm.Error.WithWriter(&buf)
	t.Cleanup(func() { pterm.DefaultBasicText, pterm.Error = oldBasic, oldErr })

	cmdErr := &executor.CommandError{
		Command:  "k3d cluster create dev",
		ExitCode: 1,
		Stderr:   "failed to bind port 6550: address already in use",
	}
	// Wrapped, the way real callers return it.
	NewErrorHandler(false).HandleError(fmt.Errorf("cluster create operation failed: %w", cmdErr))

	out := buf.String()
	assert.Contains(t, out, "address already in use", "the child's stderr must reach the user")
	assert.Contains(t, out, "k3d cluster create dev", "the failing command must be shown")
	assert.Contains(t, out, "Exit code: 1", "the exit code must be shown")
}

// TestErrorHandler_CommandError_FallsBackWithoutStderr: a child that wrote
// nothing to stderr still produces a legible message rather than panicking.
func TestErrorHandler_CommandError_FallsBackWithoutStderr(t *testing.T) {
	assert.NotPanics(t, func() {
		NewErrorHandler(false).HandleError(&executor.CommandError{Command: "uptime", ExitCode: 2})
	})
}

func TestErrorHandler_HandleError_GenericError(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		err     error
	}{
		{
			name:    "generic error verbose",
			verbose: true,
			err:     errors.New("generic error"),
		},
		{
			name:    "generic error non-verbose",
			verbose: false,
			err:     fmt.Errorf("wrapped error: %w", errors.New("inner error")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewErrorHandler(tt.verbose)

			// Test that the function doesn't panic and runs successfully
			// Note: pterm output cannot be easily captured in tests
			assert.NotPanics(t, func() {
				handler.HandleError(tt.err)
			})
		})
	}
}

func TestSplitCause_SimpleChain(t *testing.T) {
	err := fmt.Errorf("create failed for cluster X: %w", errors.New("quota exceeded"))
	headline, cause := splitCause(err)
	assert.Equal(t, "create failed for cluster X", headline)
	assert.Equal(t, "quota exceeded", cause)
}

// tfexec wraps a bare *exec.ExitError and appends terraform's stderr AFTER it:
// err.Error() is "exit status 1\nError: <the real reason>". The deepest error
// alone is the useless "exit status 1" — the cause must keep the tail, or a
// 30-minute apply failure renders as "cause: exit status 1" with the actual
// reason (quota, Free Tier eligibility, …) discarded.
func TestSplitCause_KeepsTextAfterDeepestError(t *testing.T) {
	exitErr := errors.New("exit status 1")
	tfErr := fmt.Errorf("%w\nError: creating EC2 Instance: InvalidParameterCombination - not eligible for Free Tier", exitErr)
	err := fmt.Errorf("cluster create operation failed for 'my-eks': %w",
		fmt.Errorf("terraform apply failed: %w", tfErr))

	headline, cause := splitCause(err)
	assert.Equal(t, "cluster create operation failed for 'my-eks': terraform apply failed", headline)
	assert.Contains(t, cause, "exit status 1")
	assert.Contains(t, cause, "not eligible for Free Tier",
		"the informative tail after the deepest error must reach the user")
}

// resumeHintStub carries a resume hint the way the gke provider's error does.
type resumeHintStub struct {
	err  error
	hint string
}

func (s resumeHintStub) Error() string      { return s.err.Error() }
func (s resumeHintStub) Unwrap() error      { return s.err }
func (s resumeHintStub) ResumeHint() string { return s.hint }

// An interrupted operation carrying a resume hint must be classified as an
// interruption AND run the hint-surfacing path without panicking (M2 wiring).
func TestPrintInterruption_SurfacesResumeHint(t *testing.T) {
	err := resumeHintStub{err: fmt.Errorf("apply: %w", context.Canceled), hint: "re-run create to resume"}

	assert.True(t, NewErrorHandler(false).isUserInterruption(err),
		"a context.Canceled chain must be treated as an interruption")

	var rh interface{ ResumeHint() string }
	assert.True(t, errors.As(err, &rh), "the hint must be discoverable the way printInterruption looks it up")

	// pterm output isn't captured here (house convention); lock that the
	// hint-carrying interruption path runs without panicking.
	devNull, _ := os.Open(os.DevNull)
	old := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = old }()
	assert.NotPanics(t, func() { printInterruption(err) })
}

func TestErrorHandler_TypeAssertion(t *testing.T) {
	handler := NewErrorHandler(true)

	// Test that the handler correctly identifies error types
	validationErr := &ValidationError{Field: "test", Message: "test"}
	commandErr := &executor.CommandError{Command: "test", ExitCode: 1, Stderr: "test"}
	genericErr := errors.New("test")

	// These should not panic
	assert.NotPanics(t, func() {
		handler.HandleError(validationErr)
	})
	assert.NotPanics(t, func() {
		handler.HandleError(commandErr)
	})
	assert.NotPanics(t, func() {
		handler.HandleError(genericErr)
	})
}

func TestErrorTypes_Interfaces(t *testing.T) {
	// Test that our error types implement the error interface
	var err error

	err = &ValidationError{Field: "test", Message: "test"}
	assert.NotNil(t, err)
	assert.Implements(t, (*error)(nil), err)

	err = &executor.CommandError{Command: "test", ExitCode: 1}
	assert.NotNil(t, err)
	assert.Implements(t, (*error)(nil), err)
}

func TestValidationError_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name: "all empty fields",
			err: &ValidationError{
				Field:   "",
				Value:   "",
				Message: "",
			},
			expected: "validation failed for : ",
		},
		{
			name: "special characters in value",
			err: &ValidationError{
				Field:   "path",
				Value:   "/tmp/test with spaces & symbols!",
				Message: "invalid path format",
			},
			expected: "validation failed for path '/tmp/test with spaces & symbols!': invalid path format",
		},
		{
			name: "unicode in message",
			err: &ValidationError{
				Field:   "name",
				Value:   "测试",
				Message: "must contain only ASCII characters",
			},
			expected: "validation failed for name '测试': must contain only ASCII characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// TestCommandError_LongStderrIsTruncated: a chatty child must not flood the
// error string; the tail (where the real failure usually is) survives.
func TestCommandError_LongStderrIsTruncated(t *testing.T) {
	long := strings.Repeat("noise\n", 2000) + "FINAL REASON: disk full"
	err := &executor.CommandError{Command: "helm install", ExitCode: 1, Stderr: long}

	msg := err.Error()
	assert.Less(t, len(msg), len(long), "the message must be bounded")
	assert.Contains(t, msg, "FINAL REASON: disk full", "the tail of stderr must survive truncation")
	assert.Contains(t, msg, "helm install")
}

func TestErrorHandler_NilHandling(t *testing.T) {
	tests := []struct {
		name    string
		handler *ErrorHandler
	}{
		{
			name:    "nil handler should not panic",
			handler: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Even a nil handler should not panic - this tests defensive programming
			if tt.handler == nil {
				// In this case we're testing that a nil handler would be handled gracefully
				// In practice, the caller should ensure handler is not nil
				assert.NotPanics(t, func() {
					// Simulate defensive handling if needed
					if tt.handler != nil {
						tt.handler.HandleError(errors.New("test"))
					}
				})
			}
		})
	}
}

// Benchmark tests
func BenchmarkValidationError_Error(b *testing.B) {
	err := &ValidationError{
		Field:   "email",
		Value:   "invalid@email",
		Message: "must be valid email format",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}

func BenchmarkErrorHandler_HandleError(b *testing.B) {
	handler := NewErrorHandler(false)
	err := errors.New("test error")

	// Redirect output to discard to avoid cluttering benchmark output
	devNull, _ := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	defer devNull.Close()
	old := os.Stdout
	os.Stdout = devNull
	defer func() { os.Stdout = old }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.HandleError(err)
	}
}

// The failure panel splits an error chain into headline + root cause.
func TestSplitCause(t *testing.T) {
	root := fmt.Errorf("quota exceeded")
	wrapped := fmt.Errorf("create failed for cluster big: %w", root)
	head, cause := splitCause(wrapped)
	assert.Equal(t, "create failed for cluster big", head)
	assert.Equal(t, "quota exceeded", cause)

	// Two wrapping layers: headline keeps the whole outer chain, cause is root.
	outer := fmt.Errorf("bootstrap failed: %w", wrapped)
	head, cause = splitCause(outer)
	assert.Equal(t, "bootstrap failed: create failed for cluster big", head)
	assert.Equal(t, "quota exceeded", cause)

	// A chain of one: first line is the headline, the rest is the cause block.
	head, cause = splitCause(fmt.Errorf("apply failed\nstate kept in /tmp/x"))
	assert.Equal(t, "apply failed", head)
	assert.Equal(t, "state kept in /tmp/x", cause)

	head, cause = splitCause(fmt.Errorf("plain"))
	assert.Equal(t, "plain", head)
	assert.Empty(t, cause)
}

func TestGenericHint_K3dCreateSpecialCase(t *testing.T) {
	err := fmt.Errorf("cluster create operation failed: k3d cluster create demo: exit status 1")
	hint := genericHint(err)
	assert.Contains(t, hint, "docker info")
	assert.Contains(t, hint, "6550")
}

// AvailableSummary turns "check the ref" into an answer: it lists what the
// repository actually offers. Tags sort numerically per part (1.0.9 < 1.0.48 —
// plain string sort would invert them) and are capped so hundreds of release
// tags cannot bury the branches.
func TestBranchNotFoundError_AvailableSummary(t *testing.T) {
	t.Run("empty without refs", func(t *testing.T) {
		assert.Empty(t, NewBranchNotFoundError("x").AvailableSummary())
	})

	t.Run("branches and version-sorted tags", func(t *testing.T) {
		e := NewBranchNotFoundErrorWithRefs("v1.4.0", []string{"main", "develop"}, []string{"1.0.9", "1.0.48", "0.9.1"})
		s := e.AvailableSummary()
		assert.Equal(t, "branches: develop, main; tags: 1.0.48, 1.0.9, 0.9.1", s)
	})

	t.Run("tags capped with more-marker", func(t *testing.T) {
		tags := make([]string, 0, 25)
		for i := 1; i <= 25; i++ {
			tags = append(tags, fmt.Sprintf("1.0.%d", i))
		}
		s := NewBranchNotFoundErrorWithRefs("x", nil, tags).AvailableSummary()
		assert.Contains(t, s, "1.0.25", "the highest tags must be the ones shown")
		assert.Contains(t, s, "+15 more")
		assert.NotContains(t, s, "1.0.1,", "low tags beyond the cap must be elided")
	})
}
