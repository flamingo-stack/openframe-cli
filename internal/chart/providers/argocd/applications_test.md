<!-- source-hash: cdfa92c638d91863010dcfa3b126dda9 -->
This file contains comprehensive unit tests for the ArgoCD applications manager functionality, testing application discovery, counting, and status parsing.

## Key Components

- **TestNewManager**: Tests manager initialization with mock executor
- **TestGetTotalExpectedApplications**: Tests application counting logic with multiple fallback strategies (kubectl, helm values, ApplicationSets)
- **TestParseApplications**: Tests parsing of application status from kubectl output

## Usage Example

```go
func TestApplicationManagement(t *testing.T) {
    // Create mock executor with expected responses
    mockExec := executor.NewMockCommandExecutor()
    mockExec.SetResponse("kubectl -n argocd get applications.argoproj.io", &executor.CommandResult{
        Stdout: "app1\tHealthy\tSynced\napp2\tProgressing\tSynced\n",
    })
    
    // Test manager creation and application parsing
    manager := NewManager(mockExec)
    apps, err := manager.parseApplications(context.Background(), false)
    
    assert.NoError(t, err)
    assert.Len(t, apps, 2)
    assert.Equal(t, "Healthy", apps[0].Health)
}
```

The tests verify the manager's ability to count applications through various methods (direct kubectl queries, helm values inspection, and ApplicationSet estimation) and correctly parse application health/sync status with proper error handling for edge cases.