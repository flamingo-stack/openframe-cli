package errors

import (
	stderrors "errors"
	"fmt"
	"strings"

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

// CommandError represents command execution errors
type CommandError struct {
	Command string
	Args    []string
	Err     error
}

// AlreadyHandledError wraps errors that have already been displayed to the user
type AlreadyHandledError struct {
	OriginalError error
}

func (e *CommandError) Error() string {
	return fmt.Sprintf("command '%s %v' failed: %v", e.Command, e.Args, e.Err)
}

func (e *CommandError) Unwrap() error {
	return e.Err
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
	var commandErr *CommandError
	var branchErr *BranchNotFoundError
	switch {
	case stderrors.As(err, &validationErr):
		eh.handleValidationError(validationErr)
	case stderrors.As(err, &commandErr):
		eh.handleCommandError(commandErr)
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

func (eh *ErrorHandler) handleCommandError(err *CommandError) {
	pterm.Error.Printf("❌ Command execution failed\n")
	pterm.Printf("  Command: %s\n", pterm.Yellow(err.Command))
	if len(err.Args) > 0 {
		pterm.Printf("  Arguments: %v\n", err.Args)
	}

	if eh.verbose {
		pterm.Printf("  Details: %v\n", err.Err)
	} else {
		pterm.Printf("  Error: %v\n", err.Err)
	}
}

func (eh *ErrorHandler) handleBranchNotFoundError(err *BranchNotFoundError) {
	pterm.Error.Println("Please check if the branch name is correct or use 'main' branch")
}

func (eh *ErrorHandler) handleGenericError(err error) {
	// Clean up common error patterns for better user experience
	errorMsg := err.Error()

	// Handle user interruptions (Ctrl+C). Do NOT os.Exit here — returning lets
	// the caller's deferred cleanup run and the process exit via the normal
	// error-return path.
	if eh.isUserInterruption(errorMsg) {
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
			pterm.Printf("  2. Check available ports: lsof -i :6550\n")
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

// isUserInterruption checks if the error represents a user interruption (Ctrl+C)
func (eh *ErrorHandler) isUserInterruption(errorMsg string) bool {
	// Common interruption patterns
	interruptions := []string{
		"interrupted",
		"interrupt",
		"^C",
		"cluster selection failed: ^C",
		"selection failed: ^C",
		"confirmation failed: ^C",
		"operation cancelled",
		"user cancelled",
		"context canceled",
	}

	errorLower := strings.ToLower(errorMsg)
	for _, pattern := range interruptions {
		if strings.Contains(errorLower, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// CreateValidationError creates a new validation error
func CreateValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
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

// CreateCommandError creates a new command error
func CreateCommandError(command string, args []string, err error) *CommandError {
	return &CommandError{
		Command: command,
		Args:    args,
		Err:     err,
	}
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	var target *ValidationError
	return stderrors.As(err, &target)
}

// IsCommandError checks if an error is a command error
func IsCommandError(err error) bool {
	var target *CommandError
	return stderrors.As(err, &target)
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
	if handler.isUserInterruption(err.Error()) {
		fmt.Println()
		pterm.Info.Println("Operation cancelled by user.")
	} else {
		handler.HandleError(err)
	}
	return &AlreadyHandledError{OriginalError: err}
}
