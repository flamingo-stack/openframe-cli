Entry point for the OpenFrame CLI application that initializes and executes the command-line interface with proper error handling.

## Key Components

- **main()**: Application entry point that calls the root command executor and handles any errors by printing them to stderr and exiting with status code 1

## Usage Example

```go
// This file is typically run as a compiled binary
// Build and run the CLI:
go build -o openframe-cli main.go
./openframe-cli [command] [flags]

// Error handling example (internal behavior):
if err := cmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```

The main function delegates all CLI functionality to the `cmd` package's `Execute()` function, following the common pattern for Go CLI applications using frameworks like Cobra. Any errors encountered during command execution are properly formatted and output to stderr before terminating the program with a non-zero exit status.