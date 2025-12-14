<!-- source-hash: 63461f4db97b3797377d12f223b6c6b4 -->
Test file for the ConfigurationWizard's SaaS branch extraction functionality, providing comprehensive test coverage for parsing repository branch configurations from nested value maps.

## Key Components

- **TestConfigurationWizard_GetSaaSBranchFromValues**: Main test function validating branch extraction from various input scenarios including nil/empty values, complete/incomplete structures, and different branch names
- **TestConfigurationWizard_SaaSBranchExtraction**: Tests complex nested configuration structures and validates both SaaS and OSS branch extraction
- **TestConfigurationWizard_SaaSConfigStructure**: Verifies proper handling of SaaS configuration structure edge cases

## Usage Example

```go
// Run specific test function
go test -run TestConfigurationWizard_GetSaaSBranchFromValues

// Run all SaaS-related tests
go test -run TestConfigurationWizard_SaaS

// Test with verbose output
go test -v ./configuration -run SaaS
```

The tests validate that the `getSaaSBranchFromValues()` method correctly extracts branch names from the nested path `deployment.saas.repository.branch`, returning "main" as the default when the structure is incomplete or missing. The test scenarios cover edge cases like nil values, missing nested keys, and mixed SaaS/OSS configurations.