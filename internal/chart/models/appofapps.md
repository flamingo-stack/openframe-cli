<!-- source-hash: 4398314b889120d50b16b9089d61437b -->
This file defines configuration for ArgoCD app-of-apps pattern installation using Helm with Git repositories. It provides structured configuration management and URL formatting for the helm-git plugin.

## Key Components

**AppOfAppsConfig struct**: Main configuration struct containing Git repository settings, certificate directory, values file path, and Helm deployment parameters.

**NewAppOfAppsConfig()**: Constructor function that creates a new configuration instance with sensible defaults (main branch, argocd namespace, 60m timeout).

**GetGitURL()**: Method that formats the Git repository URL according to helm-git plugin v1.4.0 specification for Helm chart installation.

## Usage Example

```go
// Create configuration with defaults
config := NewAppOfAppsConfig()

// Customize configuration
config.GitHubBranch = "develop"
config.CertDir = "/etc/ssl/certs"
config.ValuesFile = "custom-values.yaml"

// Get formatted Git URL for helm-git plugin
gitURL := config.GetGitURL()
// Returns: "git+https://github.com/flamingo-stack/openframe-oss-tenant@manifests/app-of-apps?ref=develop"

// Use with custom repository
customConfig := &AppOfAppsConfig{
    GitHubRepo:   "https://github.com/myorg/myrepo",
    GitHubBranch: "feature-branch",
    ChartPath:    "charts/apps",
    Namespace:    "gitops",
    Timeout:      "30m",
}
```