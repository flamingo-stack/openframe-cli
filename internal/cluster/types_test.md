Test file for the cluster package that validates the FlagContainer, domain models, error types, and interface implementations.

## Key Components

- **TestNewFlagContainer**: Tests creation and default values of the flag container
- **TestFlagContainer_SyncGlobalFlags**: Validates synchronization of global flags across all command flags
- **TestFlagContainer_Reset**: Tests resetting all flags to zero values
- **TestDomainTypes**: Validates cluster domain types and constants
- **TestClusterConfig/Info/NodeInfo**: Tests cluster data model structures
- **TestErrorTypes**: Comprehensive testing of custom error types with proper error wrapping
- **TestInterface_***: Validates that implementations properly satisfy required interfaces
- **TestFlagTypes**: Tests various flag structures used across commands

## Usage Example

```go
// Run tests for flag container functionality
go test -v -run TestNewFlagContainer

// Test error type handling
go test -v -run TestErrorTypes

// Validate interface compliance
go test -v -run TestInterface_ClusterService

// Test all domain model types
go test -v -run TestDomainTypes
```

The tests ensure proper initialization of flag containers with defaults, correct global flag propagation, error type assertions using `errors.As()`, and interface compliance for cluster service implementations.