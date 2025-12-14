<!-- source-hash: c4b59e3ebba24ecffd58e526a134440a -->
Test suite for the operations UI component that validates cluster selection logic and UI display methods.

## Key Components

**Test Functions:**
- `TestOperationsUI_SelectClusterForOperation` - Tests cluster selection with various input scenarios (provided args, empty names, no clusters, whitespace handling)
- `TestNewOperationsUI` - Validates constructor returns non-nil instance
- `TestOperationsUI_ShowOperationStart` - Tests operation start display for different operation types
- `TestOperationsUI_ShowOperationSuccess` - Tests success message display functionality
- `TestOperationsUI_ShowOperationError` - Tests error message display with error handling
- `TestOperationsUI_ShowNoResourcesMessage` - Tests "no resources" message display

**Test Coverage:**
- Cluster selection validation and edge cases
- UI method panic prevention
- Error handling scenarios
- Empty/whitespace input validation

## Usage Example

```go
// Run all tests
go test ./internal/ui/operations_test.go

// Run specific test
go test -run TestOperationsUI_SelectClusterForOperation

// Test cluster selection scenarios
clusters := []models.ClusterInfo{
    {Name: "test-cluster", Type: models.ClusterTypeK3d},
}
args := []string{"test-cluster"}
ui := NewOperationsUI()
result, err := ui.SelectClusterForOperation(clusters, args, "cleanup")
```

The tests focus on ensuring robust input validation and preventing UI method panics during cluster operations.