<!-- source-hash: e2f013d63119a98f0ba9330bf432bc68 -->
Implements the CLI command for installing ArgoCD and app-of-apps charts on Kubernetes clusters. Provides both interactive and non-interactive modes with support for various deployment configurations.

## Key Components

- `getInstallCmd()` - Creates the Cobra command with usage examples and flag definitions
- `runInstallCommand()` - Main execution handler that processes flags and delegates to installation service
- `InstallFlags` struct - Encapsulates all command-line flags for installation
- `extractInstallFlags()` - Extracts and validates flags with deployment mode validation
- `addInstallFlags()` - Configures all available command flags

## Usage Example

```go
// Create the install command
installCmd := getInstallCmd()

// The command supports multiple usage patterns:
// Interactive mode
// openframe chart install

// Specific cluster
// openframe chart install my-cluster

// Non-interactive with deployment mode
// openframe chart install --deployment-mode=oss-tenant --non-interactive

// With custom repository and branch
// openframe chart install --github-repo https://github.com/custom/repo --github-branch develop

// Flags are automatically validated:
flags, err := extractInstallFlags(cmd)
if err != nil {
    return err // Returns validation errors for invalid deployment modes
}
```

The command integrates with the services package for actual chart installation and uses shared error handling for consistent CLI experience.