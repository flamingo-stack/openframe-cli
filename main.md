The main entry point for the OpenFrame CLI application that initializes and executes the command-line interface, handling any errors that occur during execution.

## Key Components

- **main()**: Primary entry function that calls the CLI executor and handles error reporting

## Usage Example

This file serves as the application entry point and is typically not imported directly. When the CLI is built and run:

```go
// Build the CLI
go build -o openframe main.go

// Run the CLI (examples)
./openframe --help
./openframe deploy
./openframe config set --key value
```

The main function delegates all command handling to the `cmd` package's `Execute()` function, which manages the complete CLI functionality including argument parsing, command routing, and business logic execution.