<!-- source-hash: e43794c334073c70a189b926d84eace3 -->
Test suite for validating ArgoCD Helm values configuration. Contains comprehensive tests to verify the structure, content, and format of ArgoCD values YAML.

## Key Components

- **`TestGetArgoCDValues`** - Main test function that validates the `GetArgoCDValues()` function returns proper YAML configuration with expected ArgoCD settings
- **`TestGetArgoCDValuesStructure`** - Additional test that verifies the returned configuration has sufficient content and includes required Lua health check scripts

## Usage Example

```go
// Run the test suite
go test -v ./argocd

// Example of what the tests verify:
func TestCustomValidation(t *testing.T) {
    values := GetArgoCDValues()
    
    // Check for specific configuration
    if !strings.Contains(values, "fullnameOverride: argocd") {
        t.Error("Missing fullname override")
    }
    
    // Verify health check configuration
    if !strings.Contains(values, "resource.customizations.health") {
        t.Error("Missing health customizations")
    }
}
```

The tests ensure that `GetArgoCDValues()` returns a valid YAML configuration containing essential ArgoCD settings like fullname overrides, config customizations, and Lua-based health check scripts for Application resources.