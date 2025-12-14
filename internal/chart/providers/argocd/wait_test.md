<!-- source-hash: 16a6352a9401b58c28609bc92781d59c -->
Test file that validates the ArgoCD wait functionality, ensuring applications reach healthy states with proper context handling and dry-run support.

## Key Components

- **TestWaitForApplications_DryRun** - Verifies dry-run mode skips kubectl commands
- **TestWaitForApplications_ContextCancellation** - Tests context timeout handling with short deadlines
- **TestWaitForApplications_AllAppsHealthy** - Validates successful completion when all apps are healthy (skipped due to 30s sleep)
- **TestWaitForApplications_ParseError** - Tests error handling scenarios (skipped, incomplete implementation)

## Usage Example

```go
// Run the test suite
go test -v ./wait_test.go

// Run specific test
go test -run TestWaitForApplications_DryRun

// Run with context cancellation test
go test -run TestWaitForApplications_ContextCancellation

// Test dry-run behavior
mockExec := executor.NewMockCommandExecutor()
manager := NewManager(mockExec)
config := config.ChartInstallConfig{DryRun: true}
err := manager.WaitForApplications(context.Background(), config)
// Should return no error and make no kubectl calls
```

The tests use mock command executors to simulate kubectl responses and validate the manager's behavior under different scenarios including dry runs, context cancellation, and application health states.