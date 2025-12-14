<!-- source-hash: a30b19537539542cb9cf4f01335a6ec6 -->
A mock implementation of CommandExecutor designed for testing, allowing simulation of command execution without running actual external commands.

## Key Components

- **MockCommandExecutor**: Main mock struct that implements CommandExecutor interface
- **NewMockCommandExecutor()**: Factory function to create a new mock executor
- **SetShouldFail()**: Configure the mock to simulate command failures
- **SetResponse()**: Set custom responses for specific command patterns
- **Execute/ExecuteWithOptions()**: Mock command execution methods
- **GetExecutedCommands()**: Retrieve list of all executed commands for verification
- **WasCommandExecuted()**: Check if a command pattern was executed
- **Reset()**: Clear all execution history and responses

## Usage Example

```go
// Create mock executor for testing
mock := NewMockCommandExecutor()

// Configure custom response for git commands
gitResult := &CommandResult{
    ExitCode: 0,
    Stdout:   "commit abc123",
    Duration: 50 * time.Millisecond,
}
mock.SetResponse("git", gitResult)

// Execute command
result, err := mock.Execute(context.Background(), "git", "rev-parse", "HEAD")

// Verify execution in tests
assert.True(t, mock.WasCommandExecuted("git rev-parse"))
assert.Equal(t, "git rev-parse HEAD", mock.GetLastCommand())

// Simulate failures
mock.SetShouldFail(true, "network error")
_, err = mock.Execute(context.Background(), "curl", "example.com")
assert.Error(t, err)
```