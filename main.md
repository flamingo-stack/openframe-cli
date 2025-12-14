Main entry point for the OpenFrame CLI application that initializes and executes the command-line interface with proper error handling.

## Key Components

- **main()**: Primary function that bootstraps the CLI application
- **cmd.Execute()**: Core command execution from the cmd package
- **Error handling**: Stderr output and graceful exit on failures

## Usage Example

This file serves as the application entry point and is typically built into a binary:

```go
// Build the CLI application
// go build -o openframe-cli main.go

// The main function will be called when running:
// ./openframe-cli [commands and flags]

// Error handling example - if cmd.Execute() returns an error:
// Error: invalid command 'unknown-cmd'
// (program exits with code 1)
```

The main function delegates all CLI logic to the `cmd` package's `Execute()` function, which likely contains the cobra/CLI command definitions and routing logic.