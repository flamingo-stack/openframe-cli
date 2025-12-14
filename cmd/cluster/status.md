Defines a CLI command for displaying Kubernetes cluster status information with optional detailed output and application filtering.

## Key Components

- `getStatusCmd()` - Creates and configures the cobra command with flags and validation
- `runClusterStatus()` - Main execution logic that handles cluster selection and status display
- Interactive cluster selection through UI when no cluster name is provided
- Integration with global flags for detailed output, application filtering, and verbosity
- Support for both direct cluster specification and interactive selection

## Usage Example

```go
// Register the status command
cmd := getStatusCmd()

// Command line usage examples:
// openframe cluster status my-cluster
// openframe cluster status --detailed
// openframe cluster status my-cluster --no-apps --verbose
```

The command validates flags during pre-execution, provides interactive cluster selection when no name is specified, and delegates actual status retrieval to the service layer with appropriate configuration options.