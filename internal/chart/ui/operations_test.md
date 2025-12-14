<!-- source-hash: b28e3f7487bc6057b020348d2875bfd8 -->
Comprehensive test suite for the operations UI package, providing test coverage for cluster selection, installation flow UI methods, and error handling scenarios.

## Key Components

**Test Functions:**
- `TestNewOperationsUI` - Tests UI instance creation
- `TestSelectClusterForInstall_WithClusterArgument` - Tests cluster selection with command-line arguments
- `TestSelectClusterForInstall_InteractiveMode` - Tests interactive cluster selection mode
- `TestSelectClusterForInstall_EmptyClusterList` - Tests behavior with no available clusters
- `TestShowOperationCancelled` - Tests operation cancellation message display
- `TestShowNoClusterMessage` - Tests no cluster available message
- `TestShowInstallationStart/Complete/Error` - Tests installation status messages
- `TestOperationsUI_Integration` - End-to-end flow testing

**Test Utilities:**
- `testError` - Custom error implementation for testing error handling scenarios
- Table-driven tests for various cluster selection scenarios

## Usage Example

```go
// Run specific test
go test -run TestSelectClusterForInstall_WithClusterArgument

// Run all operations UI tests
go test ./ui/operations_test.go

// Test with verbose output
go test -v -run TestOperationsUI_Integration

// Run tests with coverage
go test -cover ./ui/...
```

The test suite validates UI method functionality, error handling, and ensures methods don't panic during output operations. Interactive tests are appropriately skipped due to user input requirements.