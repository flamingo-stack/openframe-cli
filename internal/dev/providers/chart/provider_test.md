<!-- source-hash: 6ca274ab39748b98bceab7abc582e46c -->
Test file for the chart provider package, containing comprehensive unit tests for Helm chart management functionality including provider initialization, values file validation, and chart installation workflows.

## Key Components

**Test Functions:**
- `TestNewProvider` - Tests provider instantiation with executor and verbose settings
- `TestProvider_ValidateHelmValuesFile` - Tests validation of Helm values files (empty paths, non-existent files, valid files)
- `TestProvider_PrepareDevHelmValues` - Tests preparation of development Helm values files
- `TestProvider_GetDefaultDevValues` - Tests default values file resolution logic
- `TestProvider_InstallCharts` - Tests chart installation with various scenarios
- `TestProvider_VerboseLogging` - Tests verbose mode functionality

**Test Utilities:**
- Uses `testutil.NewTestMockExecutor()` for mocking command execution
- Creates temporary files and directories for file system operations
- Implements table-driven tests for comprehensive scenario coverage

## Usage Example

```go
// Run specific test
go test -run TestProvider_ValidateHelmValuesFile

// Run all chart provider tests
go test ./provider_test.go

// Example test pattern used in the file
func TestProvider_ValidateHelmValuesFile(t *testing.T) {
    mockExecutor := testutil.NewTestMockExecutor()
    provider := NewProvider(mockExecutor, false)
    
    err := provider.validateHelmValuesFile("valid-values.yaml")
    assert.NoError(t, err)
}
```

The tests cover error handling, file validation, provider configuration, and ensure proper integration with the mock executor for isolated testing.