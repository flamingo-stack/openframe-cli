<!-- source-hash: 3f7dddbd60cc945011312ff30c8bb147 -->
Test file for validating the intercept command functionality in a Kubernetes traffic interception tool. Tests command structure, flag validation, and configuration binding using the Cobra CLI framework.

## Key Components

- **TestGetInterceptCmd()** - Validates intercept command properties, flag existence, and default values
- **TestInterceptCmd_FlagBinding()** - Tests flag value binding to the `InterceptFlags` model structure  
- **TestInterceptCmd_Examples()** - Verifies example usage text in command documentation
- **TestInterceptCmd_PreRunE()** - Confirms command execution handler setup

## Usage Example

```go
// Run the intercept command tests
func ExampleTestUsage() {
    // Test command creation and flag validation
    cmd := getInterceptCmd()
    
    // Verify port flag exists and has correct default
    port, err := cmd.Flags().GetInt("port")
    if err == nil && port == 8080 {
        // Flag properly configured
    }
    
    // Test flag binding to model
    flags := &models.InterceptFlags{}
    cmd.Flags().Set("port", "9090")
    cmd.Flags().Set("namespace", "production")
    
    // Extract values for binding
    flags.Port, _ = cmd.Flags().GetInt("port")
    flags.Namespace, _ = cmd.Flags().GetString("namespace")
}
```

The tests ensure the intercept command properly configures Telepresence integration for local development with comprehensive flag validation and model binding verification.