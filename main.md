This is the main entry point for the OpenFrame CLI application that initializes and executes the command-line interface with error handling.

## Key Components

- **main()** - Application entry point that calls the command executor and handles any errors by printing to stderr and exiting with status code 1

## Usage Example

```go
// The application is typically run from the command line
// $ openframe-cli [command] [flags]

// Example commands might include:
// $ openframe-cli init
// $ openframe-cli deploy
// $ openframe-cli status

// If any command fails, the application will:
// 1. Print the error to stderr
// 2. Exit with status code 1
```

The main function follows Go CLI best practices by delegating command logic to a separate `cmd` package and providing proper error handling with non-zero exit codes for failures.