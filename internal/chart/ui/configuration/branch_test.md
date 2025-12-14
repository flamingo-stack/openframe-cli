<!-- source-hash: f0ff30dd7ad9632bc99d1a57149c1d7c -->
Test suite for the BranchConfigurator component that validates Git branch configuration functionality in Helm chart deployments.

## Key Components

- **TestNewBranchConfigurator** - Tests constructor initialization and dependency injection
- **TestBranchConfigurator_Configure_KeepExisting** - Validates preserving existing branch configurations
- **TestBranchConfigurator_Configure_CustomBranch** - Tests custom branch assignment for OSS deployments
- **TestBranchConfigurator_Configure_WithEmptyValues** - Handles empty configuration scenarios with default fallbacks
- **TestBranchConfigurator_Configure_BranchValidation** - Validates various branch name formats and patterns
- **TestBranchConfigurator_Configure_NoChanges** - Ensures unchanged configurations remain intact

## Usage Example

```go
// Run specific test
go test -run TestBranchConfigurator_Configure_CustomBranch

// Test branch validation scenarios
func TestCustomBranchFormat(t *testing.T) {
    modifier := templates.NewHelmValuesModifier()
    config := &types.ChartConfiguration{
        Branch: &"feature/new-api",
        DeploymentMode: &types.DeploymentModeOSS,
        ModifiedSections: []string{"branch"},
    }
    
    err := modifier.ApplyConfiguration(values, config)
    assert.NoError(t, err)
}

// Run all branch configurator tests
go test ./configuration -v -run "TestBranch"
```

The tests cover branch configuration workflows including validation, deployment structure creation, and value preservation across different Git branch naming conventions.