package errors

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/pterm/pterm"
)

// ValidationError represents validation failures
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Value != "" {
		return fmt.Sprintf("validation failed for %s '%s': %s", e.Field, e.Value, e.Message)
	}
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// NOTE: there is deliberately no CommandError type here. There used to be one,
// with a polished handler — but nothing ever constructed it, so real command
// failures (executor.CommandError) fell through to the generic error dump. The
// handler now matches the type the executor actually returns.

// AlreadyHandledError wraps errors that have already been displayed to the user
type AlreadyHandledError struct {
	OriginalError error
}

func (e *AlreadyHandledError) Error() string {
	return e.OriginalError.Error()
}

func (e *AlreadyHandledError) Unwrap() error {
	return e.OriginalError
}

// ErrorHandler provides standardized error handling
type ErrorHandler struct {
	verbose bool
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(verbose bool) *ErrorHandler {
	return &ErrorHandler{verbose: verbose}
}

// HandleError processes and displays errors consistently
func (eh *ErrorHandler) HandleError(err error) {
	if err == nil {
		return
	}

	var validationErr *ValidationError
	var commandErr *executor.CommandError
	var branchErr *BranchNotFoundError
	switch {
	case stderrors.As(err, &validationErr):
		eh.handleValidationError(validationErr)
	case stderrors.As(err, &commandErr):
		eh.handleCommandError(commandErr, err)
	case stderrors.As(err, &branchErr):
		eh.handleBranchNotFoundError(branchErr)
	default:
		eh.handleGenericError(err)
	}
}

func (eh *ErrorHandler) handleValidationError(err *ValidationError) {
	pterm.Error.Printf("⚠️ Validation failed\n")
	pterm.Printf("  Field: %s\n", pterm.Yellow(err.Field))
	if err.Value != "" {
		pterm.Printf("  Value: %s\n", pterm.Red(err.Value))
	}
	pterm.Printf("  Issue: %s\n", err.Message)
}

// handleCommandError renders a failed external command: what ran, how it
// failed, and — crucially — what the child process actually said. Before this,
// the handler matched a errors.CommandError type that was never constructed
// anywhere, so real failures (executor.CommandError) fell through to the
// generic dump and the user saw "exit status 1" with no reason.
//
// outer is the full error chain, used for the friendly hint (which matches on
// wrapper text such as "cluster create operation failed").
func (eh *ErrorHandler) handleCommandError(err *executor.CommandError, outer error) {
	// DefaultBasicText, not bare pterm.Printf: the latter writes straight to
	// stdout, bypassing --silent redirection (and any test capture).
	pterm.Error.Printf("Command failed\n")
	pterm.DefaultBasicText.Printf("  Command:   %s\n", pterm.Yellow(err.Command))
	pterm.DefaultBasicText.Printf("  Exit code: %d\n", err.ExitCode)

	if reason := strings.TrimSpace(err.Stderr); reason != "" {
		pterm.DefaultBasicText.Printf("  Output:\n")
		for _, line := range strings.Split(reason, "\n") {
			pterm.DefaultBasicText.Printf("    %s\n", pterm.Red(line))
		}
	} else {
		pterm.DefaultBasicText.Printf("  Error:     %v\n", err)
	}

	if hint := friendlyHint(outer); hint != "" {
		pterm.Info.Printf("%s\n", hint)
	}
}

// handleBranchNotFoundError names the ref that could not be found. The advice
// alone ("check if the branch name is correct") was useless when the ref came
// from a config file or a default rather than from something the user typed.
func (eh *ErrorHandler) handleBranchNotFoundError(err *BranchNotFoundError) {
	pterm.Error.Printfln("Branch %q does not exist in the chart repository", err.Branch)
	pterm.Info.Println("Check the ref, or pass an existing one with --ref (e.g. --ref main)")
}

func (eh *ErrorHandler) handleGenericError(err error) {
	// Clean up common error patterns for better user experience
	errorMsg := err.Error()

	// Handle user interruptions (Ctrl+C). Do NOT os.Exit here — returning lets
	// the caller's deferred cleanup run and the process exit via the normal
	// error-return path.
	if eh.isUserInterruption(err) {
		fmt.Println()
		pterm.Info.Println("Operation cancelled by user.")
		return
	}

	// Extract meaningful error from complex error chains
	if strings.Contains(errorMsg, "cluster create operation failed") {
		pterm.Error.Printf("❌ Failed to create cluster\n")

		// Try to extract the actual k3d error and give helpful advice
		if strings.Contains(errorMsg, "exit status 1") && strings.Contains(errorMsg, "k3d cluster create") {
			pterm.Printf("  Issue: k3d cluster creation failed\n")
			fmt.Println()
			pterm.Info.Printf("🔧 Troubleshooting steps:\n")
			pterm.Printf("  1. Check Docker is running: docker info\n")
			// 6550 is only the preferred API port; k3d falls back to 6551/6552
			// when it is taken (providers/k3d/ports.go), and that fallback is
			// exactly what a port conflict looks like.
			pterm.Printf("  2. Check the API ports are free: lsof -i :6550-6552\n")
			pterm.Printf("  3. Try with different name: openframe cluster create my-test\n")
			pterm.Printf("  4. Check k3d directly: k3d version\n")
		} else {
			pterm.Printf("  Details: %s\n", errorMsg)
		}
	} else {
		// Generic error handling
		pterm.Error.Printf("❌ Operation failed\n")
		if eh.verbose {
			pterm.Printf("  Details: %v\n", err)
			pterm.Printf("  Type: %T\n", err)
		} else {
			// Show only the essential error message
			pterm.Printf("  Error: %s\n", errorMsg)
		}
		// Add a plain-language next step for common failures (req 30).
		if hint := friendlyHint(err); hint != "" {
			pterm.Info.Printf("💡 %s\n", hint)
		}
	}
}

// isInterruption reports whether err represents a user interruption (Ctrl+C).
//
// It is structural first: errors.Is(context.Canceled) matches the signal-
// cancelled root context and anything that %w-wraps ctx.Err() (e.g. "operation
// cancelled: <ctx.Err()>"). Crucially it does NOT match context.DeadlineExceeded,
// so a real timeout is not mislabeled as a user cancellation — and it won't
// false-match an unrelated error that merely mentions "context canceled" in its
// text. The remaining string checks cover promptui's Ctrl-C at an interactive
// prompt ("^C") and the exact "interrupted" some prompt sites return. "interrupted"
// is matched exactly (not as a substring) so an unrelated "connection was
// interrupted" network error is not mislabeled as a user cancellation.
func isInterruption(err error) bool {
	if err == nil {
		return false
	}
	if stderrors.Is(err, context.Canceled) {
		return true
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	return msg == "interrupted" || strings.Contains(msg, "^c")
}

// isUserInterruption reports whether err is a user interruption (Ctrl+C), so the
// handler can print a friendly "cancelled" message instead of a failure.
func (eh *ErrorHandler) isUserInterruption(err error) bool {
	return isInterruption(err)
}

// BranchNotFoundError represents a branch not found error
type BranchNotFoundError struct {
	Branch string
}

func (e *BranchNotFoundError) Error() string {
	return fmt.Sprintf("branch '%s' does not exist in repository. Please check if the branch name is correct or use 'main' branch", e.Branch)
}

// NewBranchNotFoundError creates a new branch not found error
func NewBranchNotFoundError(branch string) *BranchNotFoundError {
	return &BranchNotFoundError{Branch: branch}
}

// HandleGlobalError provides a global error handling entry point
// This should be used by all command RunE functions to ensure consistent error handling
func HandleGlobalError(err error, verbose bool) error {
	if err == nil {
		return nil
	}

	handler := NewErrorHandler(verbose)

	// Display the error (interruptions get a friendly "cancelled" message). We
	// return an AlreadyHandledError rather than calling os.Exit: the RunE caller
	// returns it, cobra/main map it to a non-zero exit code, and every deferred
	// cleanup (signal.Stop, cancel, temp-file restore) still runs. main.go
	// recognises the sentinel and does not re-print the message.
	if handler.isUserInterruption(err) {
		fmt.Println()
		pterm.Info.Println("Operation cancelled by user.")
	} else {
		handler.HandleError(err)
	}
	return &AlreadyHandledError{OriginalError: err}
}
