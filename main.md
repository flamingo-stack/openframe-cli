The main entry point for the OpenFrame CLI application that initializes and executes the command-line interface with proper error handling.

## Key Components

- **main()**: Application entry point that delegates execution to the cmd package and handles any errors by printing to stderr and exiting with status code 1

## Usage Example

```go
// This file is typically built and run as a binary
// Build the CLI:
// go build -o openframe main.go

// The main function will execute commands defined in the cmd package
// Error handling ensures clean exit with appropriate error messages
func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

This follows the standard pattern for CLI applications where the main function serves as a thin wrapper around the actual command implementation, providing consistent error handling and exit behavior.