package errors

import (
	"fmt"
)

// WrapConfirmationError wraps an error from an interactive confirmation prompt.
// A prompt interruption (Ctrl-C) is returned as-is so it flows up to main, where
// HandleGlobalError prints a friendly "Operation cancelled by user." and exits
// non-zero with deferred cleanup intact — never os.Exit here. Any other error
// gets the context prefix; nil stays nil.
func WrapConfirmationError(err error, context string) error {
	if err == nil {
		return nil
	}
	if isInterruption(err) {
		return err
	}
	return fmt.Errorf("%s: %w", context, err)
}
