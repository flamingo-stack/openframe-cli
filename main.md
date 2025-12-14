<!-- source-hash: c3c3b33add246a43fb1377252f70b38b -->
This is the entry point for the OpenFrame CLI application that initializes and executes the command-line interface, handling any execution errors with proper exit codes.

## Key Components

- **main()** - Application entry point that executes the CLI and handles errors
- **cmd.Execute()** - Imported function that runs the command-line interface
- Error handling with stderr output and exit code 1 on failure

## Usage Example

```go
// This file is typically not imported but serves as the application entry point
// To run the CLI application:
// go run main.go [command] [flags]

// Example CLI usage that this main.go would handle:
// ./openframe-cli init --project myapp
// ./openframe-cli deploy --env production
// ./openframe-cli status

// The main function ensures proper error handling:
// - Successful commands exit with code 0
// - Failed commands print error to stderr and exit with code 1
```

This main.go follows Go CLI best practices by delegating command logic to a separate cmd package while providing clean error handling and appropriate exit codes for shell integration.