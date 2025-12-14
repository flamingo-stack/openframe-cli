<!-- source-hash: 8361fce1efa2cd457ee8a67925ea2da6 -->
This file contains comprehensive unit tests for a mock command executor implementation used for testing purposes.

## Key Components

- **Test Functions**: Complete test coverage for MockCommandExecutor including:
  - Constructor testing (`TestNewMockCommandExecutor`)
  - Command execution scenarios (`TestMockCommandExecutor_Execute*`)
  - Response configuration (`TestMockCommandExecutor_SetResponse`, `TestMockCommandExecutor_SetDefaultResult`)
  - Command tracking (`TestMockCommandExecutor_GetExecutedCommands`, `TestMockCommandExecutor_WasCommandExecuted`)
  - Pattern matching and edge cases
  - Concurrent access and benchmark tests

- **Test Scenarios**: Covers success cases, failure simulation, default behaviors, pattern matching, and edge cases
- **Benchmark Tests**: Performance testing for execute and response setting operations
- **Interface Compliance**: Verifies MockCommandExecutor implements the CommandExecutor interface

## Usage Example

```go
// Run tests for the mock executor
func TestMyComponent(t *testing.T) {
    // Run all mock executor tests
    go test ./executor -run "TestMock*"
    
    // Run specific test
    go test ./executor -run "TestMockCommandExecutor_Execute"
    
    // Run benchmarks
    go test ./executor -bench "BenchmarkMock*"
    
    // Run with coverage
    go test ./executor -cover -run "TestMock*"
}
```

The tests validate that the mock executor correctly simulates command execution, tracks executed commands, handles pattern matching for responses, and provides proper error simulation capabilities for testing command-based workflows.