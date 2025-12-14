This file serves as the entry point for the OpenFrame CLI application, initializing the command-line interface and handling any execution errors with proper exit codes.

## Key Components

- **main()** - Application entry point that executes the CLI commands and handles errors
- **cmd.Execute()** - Imported function that runs the CLI command parser and handlers
- **Error handling** - Captures and displays execution errors to stderr before exiting

## Usage Example

```go
// This is typically run as a compiled binary:
// $ openframe-cli [command] [flags]

// The main function automatically handles command execution:
func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}

// Example CLI usage would be:
// $ openframe-cli deploy --config config.yaml
// $ openframe-cli status
// $ openframe-cli version
```

This minimal main package follows Go CLI best practices by delegating command logic to a separate `cmd` package while maintaining clean error handling and proper exit codes for shell integration.