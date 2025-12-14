Application entry point for the OpenFrame CLI tool that initializes and executes the command-line interface with proper error handling.

## Key Components

- **main()** - Entry function that executes the CLI command tree and handles any errors by printing to stderr and exiting with status code 1
- **cmd.Execute()** - Imported function from the cmd package that runs the CLI application logic

## Usage Example

```go
// Run the CLI application
func main() {
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

This is a standard Go CLI application pattern where the main function serves as a thin wrapper around the actual command execution logic. The real CLI functionality is implemented in the `cmd` package, keeping the main function focused solely on error handling and process exit codes.