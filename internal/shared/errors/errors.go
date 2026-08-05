package errors

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/flamingo-stack/openframe-cli/internal/shared/executor"
	"github.com/flamingo-stack/openframe-cli/internal/shared/ui"
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

// handleGenericError renders the structured failure panel:
//
//	✖ create failed for cluster big-gke
//	  cause   googleapi: Error 403: quota exceeded
//	  hint    💡 Request more quota, lower --nodes, or pick another region
//	  resume  openframe cluster create big-gke
//
// One glance answers what failed, why, and what to do next — instead of the
// previous single wall-of-text "❌ Operation failed / Error: a: b: c: d".
func (eh *ErrorHandler) handleGenericError(err error) {
	// Handle user interruptions (Ctrl+C). Do NOT os.Exit here — returning lets
	// the caller's deferred cleanup run and the process exit via the normal
	// error-return path.
	if eh.isUserInterruption(err) {
		printInterruption(err)
		return
	}

	headline, cause := splitCause(err)
	// In GitHub Actions the failure also becomes a job/PR annotation, so the
	// cause is visible without opening the 40-minute log.
	ui.ErrorAnnotation(headline, firstLine(cause))
	// No manual failure glyph: the themed pterm.Error prefix carries it
	// (✖ interactively, the "error" tag otherwise).
	pterm.Error.Printf("%s\n", headline)
	if cause != "" {
		panelRow("cause", cause)
	}
	if eh.verbose && cause != "" && err.Error() != headline+": "+cause {
		// The full chain names every wrapping layer — noise by default,
		// useful when debugging which layer swallowed what.
		panelRow("chain", err.Error())
	}
	if hint := genericHint(err); hint != "" {
		panelRow("hint", "💡 "+hint)
	}
	var rh interface{ ResumeHint() string }
	if stderrors.As(err, &rh) {
		if hint := rh.ResumeHint(); hint != "" {
			panelRow("resume", hint)
		}
	}
}

// genericHint merges the pattern-matched friendly hint with the k3d-create
// special case that used to be a separate hardcoded block.
func genericHint(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "cluster create operation failed") &&
		strings.Contains(msg, "exit status 1") && strings.Contains(msg, "k3d cluster create") {
		// 6550 is only the preferred API port; k3d falls back to 6551/6552
		// when it is taken (providers/k3d/ports.go), and that fallback is
		// exactly what a port conflict looks like.
		return "Check Docker is running (docker info) and the API ports are free (lsof -i :6550-6552), then retry — or try another name"
	}
	return friendlyHint(err)
}

// splitCause splits an error chain into the outer description and the root
// cause. "create failed for cluster X: quota exceeded" → ("create failed for
// cluster X", "quota exceeded"). A chain of one keeps everything in the
// headline (first line) with any remaining lines as the cause block.
func splitCause(err error) (headline, cause string) {
	full := err.Error()
	deepest := err
	unwrapped := false
	for {
		u := stderrors.Unwrap(deepest)
		if u == nil {
			break
		}
		deepest = u
		unwrapped = true
	}
	if unwrapped {
		causeText := deepest.Error()
		if idx := strings.LastIndex(full, causeText); idx > 0 && causeText != "" {
			head := strings.TrimRight(strings.TrimSuffix(full[:idx], ": "), ": \n")
			if head != "" {
				return head, causeText
			}
		}
	}
	lines := strings.SplitN(full, "\n", 2)
	if len(lines) == 2 {
		return lines[0], lines[1]
	}
	return full, ""
}

// firstLine trims a possibly multi-line cause (pod logs, terraform output)
// down to its first line for the one-line annotation surface.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// panelRow prints one aligned "key value" row of the failure panel; multi-line
// values (pod logs, terraform output) stay indented under their key.
func panelRow(key, value string) {
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	pterm.DefaultBasicText.Printf("  %s %s\n", pterm.FgGray.Sprintf("%-7s", key), lines[0])
	for _, l := range lines[1:] {
		pterm.DefaultBasicText.Printf("  %-7s %s\n", "", l)
	}
}

// isInterruption reports whether err represents a user interruption (Ctrl+C).
//
// It is structural first: errors.Is(context.Canceled) matches the signal-
// cancelled root context and anything that %w-wraps ctx.Err() (e.g. "operation
// cancelled: <ctx.Err()>"). Crucially it does NOT match context.DeadlineExceeded,
// so a real timeout is not mislabeled as a user cancellation — and it won't
// false-match an unrelated error that merely mentions "context canceled" in its
// text. ui.ErrPromptInterrupted matches structurally too, so a %w-wrapped
// aborted prompt ("cluster selection failed: interrupted") still counts. The
// remaining string checks cover the exact "interrupted" some prompt sites
// return and legacy "^C" strings. "interrupted" is matched exactly (not as a
// substring) so an unrelated "connection was interrupted" network error is not
// mislabeled as a user cancellation.
func isInterruption(err error) bool {
	if err == nil {
		return false
	}
	if stderrors.Is(err, context.Canceled) || stderrors.Is(err, ui.ErrPromptInterrupted) {
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

	// A sentinel anywhere in the chain means an inner layer already displayed
	// this error (the chart workflow calls this handler itself before its
	// result travels back through the cmd layer, which calls it again) —
	// re-rendering printed the same failure twice. Return the error as-is so
	// the sentinel still maps to a non-zero exit without a second print.
	var handled *AlreadyHandledError
	if stderrors.As(err, &handled) {
		return err
	}

	handler := NewErrorHandler(verbose)

	// Display the error (interruptions get a friendly "cancelled" message). We
	// return an AlreadyHandledError rather than calling os.Exit: the RunE caller
	// returns it, cobra/main map it to a non-zero exit code, and every deferred
	// cleanup (signal.Stop, cancel, temp-file restore) still runs. main.go
	// recognises the sentinel and does not re-print the message.
	if handler.isUserInterruption(err) {
		printInterruption(err)
	} else {
		handler.HandleError(err)
	}
	return &AlreadyHandledError{OriginalError: err}
}

// printInterruption prints the "cancelled" notice, plus any resume hint an
// error in the chain carries structurally. On an interruption the plain
// err.Error() text is deliberately not shown, so a hint wrapped only as text
// (e.g. "re-run create to resume") would be lost — a value implementing
// ResumeHint() carries it through the swallow.
func printInterruption(err error) {
	fmt.Println()
	pterm.Info.Println("Operation cancelled by user.")
	var rh interface{ ResumeHint() string }
	if stderrors.As(err, &rh) {
		if hint := rh.ResumeHint(); hint != "" {
			pterm.Info.Println(hint)
		}
	}
}
