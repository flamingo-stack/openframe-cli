# Contributing Guidelines

Welcome to OpenFrame CLI! This guide outlines our code style, development process, and review criteria. Following these guidelines ensures smooth collaboration and maintains high code quality.

## Code Style and Conventions

### Go Code Style

We follow standard Go conventions with additional project-specific guidelines.

#### Formatting and Organization

```go
// Package declaration with clear documentation
// Package cluster provides Kubernetes cluster management functionality
// for OpenFrame CLI, including K3d cluster lifecycle operations.
package cluster

import (
    // Standard library imports first
    "context"
    "fmt"
    "os"
    
    // Third-party imports second
    "github.com/spf13/cobra"
    "github.com/docker/docker/client"
    
    // Local imports last, grouped by module
    "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
    "github.com/flamingo-stack/openframe-cli/internal/cluster/services"
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)
```

#### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| **Packages** | lowercase, single word | `cluster`, `chart`, `bootstrap` |
| **Types** | PascalCase | `ClusterService`, `ChartConfig` |
| **Interfaces** | PascalCase, often ending in -er | `ClusterProvider`, `Installer` |
| **Functions** | PascalCase (public), camelCase (private) | `CreateCluster()`, `validateConfig()` |
| **Variables** | camelCase | `clusterName`, `httpPort` |
| **Constants** | PascalCase or ALL_CAPS | `DefaultTimeout`, `MAX_RETRIES` |

#### Documentation Standards

```go
// ClusterService manages the lifecycle of Kubernetes clusters using K3d.
// It provides creation, deletion, and status operations with interactive
// configuration support and comprehensive error handling.
type ClusterService struct {
    provider    ClusterProvider
    ui         UIService
    logger     Logger
}

// Create creates a new K3d cluster with the specified configuration.
// It validates prerequisites, configures networking, and starts all nodes.
//
// Parameters:
//   - ctx: Context for cancellation and timeout control
//   - config: Cluster configuration including name, nodes, and ports
//
// Returns:
//   - error: Any error that occurred during creation, or nil on success
//
// Example:
//   config := &ClusterConfig{Name: "my-cluster", Nodes: 3}
//   err := service.Create(ctx, config)
//   if err != nil {
//       return fmt.Errorf("failed to create cluster: %w", err)
//   }
func (s *ClusterService) Create(ctx context.Context, config *ClusterConfig) error {
    if err := s.validatePrerequisites(ctx); err != nil {
        return fmt.Errorf("prerequisites validation failed: %w", err)
    }
    
    return s.provider.Create(ctx, config)
}
```

#### Error Handling Patterns

```go
// Error wrapping for context
func (s *ClusterService) Create(ctx context.Context, config *ClusterConfig) error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid cluster configuration: %w", err)
    }
    
    if err := s.checkPrerequisites(ctx); err != nil {
        return fmt.Errorf("prerequisites check failed: %w", err)
    }
    
    if err := s.provider.Create(ctx, config); err != nil {
        return fmt.Errorf("cluster creation failed: %w", err)
    }
    
    return nil
}

// Custom error types for specific cases
type PrerequisiteError struct {
    Tool    string
    Version string
    Message string
}

func (e *PrerequisiteError) Error() string {
    return fmt.Sprintf("prerequisite %s (version %s) error: %s", 
        e.Tool, e.Version, e.Message)
}

func (e *PrerequisiteError) Is(target error) bool {
    _, ok := target.(*PrerequisiteError)
    return ok
}
```

### Project Structure Conventions

#### Package Organization

```
internal/
├── bootstrap/              # Bootstrap orchestration
│   ├── service.go         # Main service implementation
│   ├── models.go          # Data structures
│   └── service_test.go    # Tests
├── cluster/               # Cluster management
│   ├── services/          # Service implementations
│   │   ├── cluster_service.go
│   │   ├── k3d_provider.go
│   │   └── command_service.go
│   ├── models/            # Data models and configuration
│   │   ├── configuration.go
│   │   └── cluster_info.go
│   ├── ui/                # User interface components
│   │   ├── wizard.go
│   │   └── prompts.go
│   └── prerequisites/     # Prerequisite checking
│       └── checker.go
└── shared/                # Shared utilities
    ├── ui/                # Common UI components
    ├── errors/            # Error handling
    └── config/            # Configuration management
```

#### File Naming

- **Source files**: `snake_case.go` (e.g., `cluster_service.go`)
- **Test files**: `*_test.go` (e.g., `cluster_service_test.go`)
- **Interface files**: Often `interfaces.go` or `*_interface.go`
- **Mock files**: `mock_*.go` in `mocks/` directory

### Configuration and Flags

#### Flag Definition Standards

```go
// Define flags with consistent naming and descriptions
func addClusterFlags(cmd *cobra.Command) {
    flags := cmd.Flags()
    
    // Use kebab-case for flag names
    flags.StringP("cluster-name", "n", "", 
        "Name of the cluster to create (required)")
    flags.IntP("nodes", "", 3,
        "Number of worker nodes (default: 3)")
    flags.IntP("http-port", "", 80,
        "HTTP port for ingress (default: 80)")
    flags.BoolP("enable-registry", "r", true,
        "Enable local container registry (default: true)")
    
    // Mark required flags
    cmd.MarkFlagRequired("cluster-name")
    
    // Set up flag completion
    cmd.RegisterFlagCompletionFunc("cluster-name", 
        completeClusterNames)
}
```

#### Configuration Structure

```go
// Use embedded structs for configuration composition
type ClusterConfig struct {
    // Core cluster settings
    Name    string `yaml:"name" validate:"required,min=3,max=63"`
    Nodes   int    `yaml:"nodes" validate:"min=1,max=10"`
    
    // Network configuration
    Network NetworkConfig `yaml:"network"`
    
    // Feature flags
    Features FeatureConfig `yaml:"features"`
}

type NetworkConfig struct {
    HTTPPort     int    `yaml:"httpPort" validate:"min=1,max=65535"`
    HTTPSPort    int    `yaml:"httpsPort" validate:"min=1,max=65535"`
    KubeAPIPort  int    `yaml:"kubeApiPort" validate:"min=1,max=65535"`
    Registry     bool   `yaml:"registry"`
    RegistryPort int    `yaml:"registryPort,omitempty"`
}

// Implement validation methods
func (c *ClusterConfig) Validate() error {
    if err := validator.New().Struct(c); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Custom validation logic
    if c.Network.HTTPPort == c.Network.HTTPSPort {
        return errors.New("HTTP and HTTPS ports must be different")
    }
    
    return nil
}
```

## Branch Naming and Git Workflow

### Branch Naming Convention

| Type | Pattern | Example | Description |
|------|---------|---------|-------------|
| **Feature** | `feature/description` | `feature/cluster-registry-support` | New functionality |
| **Bugfix** | `fix/issue-description` | `fix/cluster-deletion-timeout` | Bug fixes |
| **Hotfix** | `hotfix/critical-issue` | `hotfix/security-vulnerability` | Critical production fixes |
| **Documentation** | `docs/topic` | `docs/contributing-guidelines` | Documentation updates |
| **Refactor** | `refactor/component` | `refactor/cluster-service` | Code restructuring |
| **Test** | `test/feature-area` | `test/integration-coverage` | Test improvements |

### Git Workflow

```mermaid
gitgraph
    commit id: "main"
    branch feature/cluster-registry
    checkout feature/cluster-registry
    commit id: "Add registry config"
    commit id: "Implement registry logic"
    commit id: "Add tests"
    checkout main
    commit id: "Other changes"
    merge feature/cluster-registry
    commit id: "Release v1.2.0"
```

#### Development Process

1. **Create Feature Branch**
```bash
# Start from updated main branch
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/cluster-registry-support
```

2. **Development Cycle**
```bash
# Make changes and commit frequently
git add .
git commit -m "Add registry configuration struct"

# Push branch regularly
git push origin feature/cluster-registry-support
```

3. **Prepare for PR**
```bash
# Update from main and rebase if needed
git checkout main
git pull origin main
git checkout feature/cluster-registry-support
git rebase main

# Run tests and linting
make test
make lint

# Push final changes
git push origin feature/cluster-registry-support --force-with-lease
```

## Commit Message Format

We follow conventional commit format for clear, semantic commit history.

### Format Structure

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Commit Types

| Type | Usage | Example |
|------|-------|---------|
| **feat** | New features | `feat(cluster): add registry support` |
| **fix** | Bug fixes | `fix(chart): resolve ArgoCD timeout issue` |
| **docs** | Documentation | `docs: update contributing guidelines` |
| **test** | Testing | `test(cluster): add integration tests` |
| **refactor** | Code refactoring | `refactor(ui): simplify prompt logic` |
| **perf** | Performance improvements | `perf(cluster): optimize node creation` |
| **build** | Build system | `build: update Go version to 1.21` |
| **ci** | CI/CD changes | `ci: add integration test workflow` |

### Examples

```bash
# Feature addition
git commit -m "feat(cluster): add support for custom registry ports

- Allow users to specify custom registry port
- Add validation for port conflicts
- Update configuration schema

Closes #123"

# Bug fix
git commit -m "fix(chart): resolve ArgoCD installation timeout

The ArgoCD installation was timing out due to insufficient
resource limits. Increased memory limits and added retry logic.

Fixes #456"

# Documentation
git commit -m "docs(development): add testing overview guide

- Document test structure and organization
- Add examples for unit, integration, and E2E tests
- Include coverage requirements and best practices"

# Breaking change
git commit -m "feat(cluster)!: change default node count to 3

BREAKING CHANGE: The default number of nodes has changed from 1 to 3
for better high-availability support. Users who need single-node
clusters must explicitly specify --nodes=1."
```

## Pull Request Process

### PR Template

We use a standard PR template to ensure consistency:

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Refactoring (no functional changes)

## Testing
- [ ] Unit tests pass (`make test`)
- [ ] Integration tests pass (`make test-integration`)
- [ ] E2E tests pass (if applicable)
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines (`make lint`)
- [ ] Self-review completed
- [ ] Documentation updated (if needed)
- [ ] Tests added/updated for changes
- [ ] No breaking changes (or properly documented)

## Screenshots/Logs (if applicable)
Include relevant screenshots or log output.

## Related Issues
Closes #[issue number]
```

### PR Workflow

1. **Create Pull Request**
   - Use descriptive title following commit convention
   - Fill out PR template completely
   - Assign appropriate labels and reviewers

2. **Review Process**
   - Address reviewer feedback promptly
   - Update documentation if requested
   - Ensure CI checks pass

3. **Merge Requirements**
   - At least one approved review
   - All CI checks passing
   - No conflicts with main branch
   - Branch up to date with main

### Review Checklist

#### For Authors
- [ ] PR description clearly explains changes
- [ ] Code follows project conventions
- [ ] Tests cover new functionality
- [ ] Documentation updated if needed
- [ ] No unnecessary changes included
- [ ] CI checks pass

#### For Reviewers
- [ ] Code is readable and maintainable
- [ ] Logic is sound and efficient
- [ ] Error handling is appropriate
- [ ] Tests are comprehensive
- [ ] Documentation is accurate
- [ ] Breaking changes are justified

## Code Review Guidelines

### What to Look For

#### Code Quality
```go
// ❌ Poor: No error handling, unclear naming
func createCluster(name string) {
    k3d.Create(name)
    fmt.Println("Done")
}

// ✅ Good: Proper error handling, clear intent
func (s *ClusterService) Create(ctx context.Context, config *ClusterConfig) error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    if err := s.provider.Create(ctx, config); err != nil {
        return fmt.Errorf("cluster creation failed: %w", err)
    }
    
    s.logger.Info("cluster created successfully", 
        "name", config.Name, 
        "nodes", config.Nodes)
    return nil
}
```

#### Testing Coverage
```go
// ✅ Good: Comprehensive test with multiple scenarios
func TestClusterServiceCreate(t *testing.T) {
    tests := []struct {
        name        string
        config      *ClusterConfig
        providerErr error
        expectError bool
    }{
        {
            name:   "valid configuration",
            config: validClusterConfig(),
            expectError: false,
        },
        {
            name:   "invalid configuration",
            config: invalidClusterConfig(),
            expectError: true,
        },
        {
            name:        "provider error",
            config:      validClusterConfig(),
            providerErr: errors.New("provider failed"),
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation...
        })
    }
}
```

#### Documentation
```go
// ✅ Good: Clear documentation with examples
// ClusterProvider defines the interface for cluster management operations.
// Implementations must handle cluster lifecycle, networking configuration,
// and resource cleanup.
//
// Example usage:
//   provider := NewK3dProvider()
//   config := &ClusterConfig{Name: "my-cluster", Nodes: 3}
//   if err := provider.Create(ctx, config); err != nil {
//       return err
//   }
type ClusterProvider interface {
    // Create creates a new cluster with the specified configuration.
    // Returns an error if creation fails or if prerequisites are not met.
    Create(ctx context.Context, config *ClusterConfig) error
    
    // Delete removes the specified cluster and all associated resources.
    // Returns an error if deletion fails or if the cluster doesn't exist.
    Delete(ctx context.Context, name string) error
}
```

### Review Process

#### Feedback Guidelines

**Constructive Feedback**
```
# ❌ Poor feedback
"This is wrong."

# ✅ Good feedback
"Consider using context.WithTimeout here to prevent indefinite blocking. 
The cluster creation could hang if K3d becomes unresponsive."
```

**Specific Suggestions**
```
# ❌ Vague
"Improve error handling."

# ✅ Specific
"Wrap this error with additional context about which cluster operation failed:
return fmt.Errorf('failed to delete cluster %s: %w', name, err)"
```

#### Review Categories

Use these categories for feedback:

- **🔴 Critical**: Must be addressed before merge
- **🟡 Suggestion**: Improvement recommendation
- **🟢 Nit**: Minor style/preference issue
- **💭 Question**: Clarification needed
- **👍 Praise**: Positive feedback on good practices

### Common Review Issues

#### Performance Concerns
```go
// ❌ Avoid: Inefficient operations in loops
for _, cluster := range clusters {
    status := k3d.GetStatus(cluster.Name) // N API calls
}

// ✅ Better: Batch operations when possible
statuses := k3d.GetAllStatuses() // 1 API call
for _, cluster := range clusters {
    status := statuses[cluster.Name]
}
```

#### Security Considerations
```go
// ❌ Security risk: Command injection
cmd := exec.Command("k3d", "cluster", "create", userInput)

// ✅ Safe: Validate and sanitize input
if !isValidClusterName(userInput) {
    return errors.New("invalid cluster name")
}
cmd := exec.Command("k3d", "cluster", "create", userInput)
```

## Development Environment

### Required Tools

Ensure these tools are installed and configured:

```bash
# Go development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golang/mock/mockgen@latest

# Pre-commit hooks
pip install pre-commit
pre-commit install

# Documentation tools
go install golang.org/x/tools/cmd/godoc@latest
```

### IDE Configuration

#### VS Code Settings

```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.useLanguageServer": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "go.testFlags": ["-v"],
  "go.coverOnSave": true
}
```

#### Git Hooks

```bash
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt -w
        language: system
        files: \.go$
      - id: go-lint
        name: go lint
        entry: golangci-lint run
        language: system
        files: \.go$
        pass_filenames: false
```

## Release Process

### Version Management

We use semantic versioning (SemVer):

- **Major** (X.0.0): Breaking changes
- **Minor** (0.X.0): New features, backward compatible
- **Patch** (0.0.X): Bug fixes, backward compatible

### Release Checklist

- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version bumped in code
- [ ] Release notes prepared
- [ ] Security review completed (if needed)

---

## Summary

Contributing to OpenFrame CLI involves:

1. **Following Code Standards**: Consistent style and clear documentation
2. **Using Git Workflow**: Proper branching and commit messages
3. **Creating Quality PRs**: Complete descriptions and thorough testing
4. **Participating in Reviews**: Constructive feedback and collaboration
5. **Maintaining Quality**: Tests, documentation, and best practices

Your contributions help make OpenFrame CLI better for everyone. Thank you for contributing!

**Next Steps:**
- Review the [Architecture Overview](../architecture/overview.md) to understand the codebase
- Check out the [Testing Overview](../testing/overview.md) for testing practices
- Start with good first issues labeled `good-first-issue` in the repository

Happy contributing! 🚀