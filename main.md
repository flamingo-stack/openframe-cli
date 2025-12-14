This file serves as the entry point for the OpenFrame CLI application, handling command execution and error management.

## Key Components

- **main()** - Entry point function that executes CLI commands and handles errors
- **cmd.Execute()** - External command execution function from the cmd package
- **Error handling** - Writes errors to stderr and exits with status code 1 on failure

## Usage Example

```go
// This is typically run as a compiled binary
// The main function automatically executes when the program starts

// If cmd.Execute() succeeds, the program exits normally
// If it fails, error output goes to stderr:
// Error: command failed
// (program exits with code 1)
```

The application follows standard CLI patterns by delegating actual command logic to the `cmd` package while keeping the main function minimal and focused on error handling.