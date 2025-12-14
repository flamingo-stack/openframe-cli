<!-- source-hash: 7bbadf5e4f9688d6f0b9ea11b1951768 -->
Provides a flexible abstraction layer for executing external system commands with support for dry-run mode, timeouts, custom environments, and detailed result tracking.

## Key Components

- **CommandExecutor** - Interface for command execution with dependency injection support
- **RealCommandExecutor** - Concrete implementation that executes actual system commands
- **CommandResult** - Struct containing execution results (exit code, stdout, stderr, duration)
- **ExecuteOptions** - Configuration struct for advanced command execution settings

## Usage Example

```go
// Create executor with dry-run and verbose logging
executor := NewRealCommandExecutor(false, true)

// Simple command execution
result, err := executor.Execute(ctx, "ls", "-la", "/tmp")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Exit code: %d\nOutput: %s\n", result.ExitCode, result.Stdout)

// Advanced execution with custom options
options := ExecuteOptions{
    Command: "git",
    Args:    []string{"clone", "https://github.com/user/repo.git"},
    Dir:     "/workspace",
    Env:     map[string]string{"GIT_TOKEN": "secret"},
    Timeout: 30 * time.Second,
}
result, err = executor.ExecuteWithOptions(ctx, options)
```

The executor supports dry-run mode for testing, verbose logging for debugging, and handles timeouts, working directories, and environment variables seamlessly.