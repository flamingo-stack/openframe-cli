<!-- source-hash: c864d6b5dcef83b5665dddb316774eaa -->
Defines domain-specific error types for cluster management operations with structured error handling and contextual information.

## Key Components

**Error Types:**
- `ErrClusterNotFound` - Cluster does not exist
- `ErrProviderNotFound` - No provider available for cluster type
- `ErrInvalidClusterConfig` - Invalid cluster configuration
- `ErrClusterAlreadyExists` - Duplicate cluster name
- `ErrClusterOperation` - General operation failure with error wrapping

**Constructor Functions:**
- `NewClusterNotFoundError()` - Creates cluster not found error
- `NewProviderNotFoundError()` - Creates provider not found error
- `NewInvalidConfigError()` - Creates configuration validation error
- `NewClusterAlreadyExistsError()` - Creates duplicate cluster error
- `NewClusterOperationError()` - Creates operation failure error

## Usage Example

```go
// Check for specific error types
if err := clusterService.GetCluster("my-cluster"); err != nil {
    var notFoundErr ErrClusterNotFound
    if errors.As(err, &notFoundErr) {
        log.Printf("Cluster %s does not exist", notFoundErr.Name)
    }
}

// Create and return domain errors
func ValidateCluster(name string) error {
    if name == "" {
        return NewInvalidConfigError("name", name, "cannot be empty")
    }
    return nil
}

// Wrap underlying errors
if err := provider.CreateCluster(config); err != nil {
    return NewClusterOperationError("create", config.Name, err)
}
```