<!-- source-hash: 91f187cfbac4154c51d2c53b84009b95 -->
Test suite for the AppOfAppsConfig model, verifying configuration initialization and Git URL generation functionality for ArgoCD app-of-apps pattern deployment.

## Key Components

- **TestNewAppOfAppsConfig** - Tests the default configuration constructor with predefined values
- **TestAppOfAppsConfig_GetGitURL** - Validates Git URL formatting for Helm chart references
- **TestAppOfAppsConfig_Fields** - Verifies all configuration fields are accessible and assignable
- **TestAppOfAppsConfig_CompleteConfiguration** - Tests full configuration scenarios with custom values

## Usage Example

```go
// Run all tests
go test ./models -v

// Run specific test
go test ./models -run TestNewAppOfAppsConfig

// The tests verify default configuration:
config := NewAppOfAppsConfig()
// Expects: GitHubRepo="https://github.com/flamingo-stack/openframe-oss-tenant"
// Expects: GitHubBranch="main", ChartPath="manifests/app-of-apps"

// Tests Git URL generation:
config.GetGitURL()
// Returns: "git+https://github.com/repo@path?ref=branch"
```

The test suite ensures the AppOfAppsConfig struct properly handles GitHub repository URLs, branch references, chart paths, and generates correctly formatted Git URLs for ArgoCD Helm chart deployments. It validates both default initialization and custom configuration scenarios.