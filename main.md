Entry point for the OpenFrame CLI application that initializes and executes the command-line interface with basic error handling.

## Key Components

- **main()** - Application entry point that delegates to the cmd package's Execute function
- **Error handling** - Captures and displays any execution errors to stderr before exiting

## Usage Example

This file serves as the application bootstrap and is typically not imported by other packages. To run the CLI:

```go
// Build and run the application
go build -o openframe-cli main.go
./openframe-cli [commands and flags]
```

The main function follows the standard Go CLI pattern by:
1. Calling the root command executor from the cmd package
2. Handling any returned errors by printing them to stderr
3. Exiting with status code 1 on failure

```go
// The execution flow delegates to cmd.Execute()
// which contains the actual CLI logic and command definitions
if err := cmd.Execute(); err != nil {
    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    os.Exit(1)
}
```