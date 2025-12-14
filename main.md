Entry point for the OpenFrame CLI application that initializes and executes the command-line interface with error handling.

## Key Components

- **main()** - Application entry point that calls the command executor and handles any errors by printing to stderr and exiting with status code 1

## Usage Example

This file serves as the application bootstrap and is typically built into an executable:

```go
// Build the CLI application
go build -o openframe main.go

// The main function will execute when the binary is run
./openframe [command] [flags]
```

The main function delegates all command processing to the `cmd` package's `Execute()` function, ensuring clean separation of concerns between application startup and command logic. Any errors returned from command execution are properly formatted and cause the program to exit with a non-zero status code for proper shell integration.