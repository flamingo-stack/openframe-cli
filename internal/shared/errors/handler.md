<!-- source-hash: 63d0d3bc5f4a4371fbd6eacdf54af1ac -->
This file provides centralized error handling for user confirmation prompts, specifically managing interruption signals and graceful program termination.

## Key Components

- **HandleConfirmationError(err error) bool** - Processes confirmation errors and handles user interruptions (Ctrl+C) by displaying a cancellation message and exiting the program. Returns true if error was handled.

- **WrapConfirmationError(err error, context string) error** - Combines error handling with contextual wrapping. First checks for interruptions, then wraps non-interruption errors with additional context.

## Usage Example

```go
package main

import (
    "errors"
    "github.com/pterm/pterm"
)

func confirmDeletion() error {
    result, err := pterm.DefaultInteractiveConfirm.
        WithDefaultText("Delete all files?").
        Show()
    
    // Handle interruption gracefully
    if errors.HandleConfirmationError(err) {
        return nil // Program will exit
    }
    
    if !result {
        return errors.New("operation cancelled")
    }
    return nil
}

func deleteWithContext() error {
    _, err := pterm.DefaultInteractiveConfirm.
        WithDefaultText("Proceed with deletion?").
        Show()
    
    // Wrap with context while handling interruptions
    return errors.WrapConfirmationError(err, "failed to get user confirmation")
}
```

The handlers ensure consistent behavior when users interrupt confirmation prompts, providing clean exit messages instead of raw error output.