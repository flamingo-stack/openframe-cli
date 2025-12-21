# Contributing Guidelines

Welcome to the OpenFrame CLI project! This guide outlines our development standards, workflows, and best practices for contributing code, documentation, and improvements.

## Getting Started

### Before You Contribute

1. **Read the Documentation**:
   - [Architecture Overview](../architecture/overview.md) - Understand the system design
   - [Local Development](../setup/local-development.md) - Set up your development environment
   - [Testing Overview](../testing/overview.md) - Learn our testing practices

2. **Set Up Your Environment**:
   - Complete [Environment Setup](../setup/environment.md)
   - Fork the repository on GitHub
   - Clone your fork locally

3. **Join the Community**:
   - Check existing issues and discussions
   - Introduce yourself in discussions
   - Ask questions if you need clarification

## Code Standards

### Go Code Style

We follow standard Go conventions with additional project-specific guidelines:

#### 1. Formatting and Imports

```go
// Use goimports for automatic formatting and import management
//go:generate goimports -w .

// Standard import order: stdlib, external, internal
import (
    "context"
    "fmt"
    "time"
    
    "github.com/spf13/cobra"
    "k8s.io/client-go/kubernetes"
    
    "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)
```

#### 2. Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| **Packages** | lowercase, short, descriptive | `cluster`, `models`, `ui` |
| **Variables** | camelCase | `clusterConfig`, `nodeCount` |
| **Constants** | PascalCase or ALL_CAPS | `DefaultTimeout`, `MAX_RETRIES` |
| **Functions** | PascalCase (exported), camelCase (private) | `CreateCluster()`, `validateConfig()` |
| **Interfaces** | PascalCase ending in -er when possible | `ClusterProvider`, `ConfigValidator` |
| **Structs** | PascalCase | `ClusterConfig`, `ServiceOptions` |

#### 3. Function Design

```go
// Good: Clear function signature with context
func CreateCluster(ctx context.Context, config ClusterConfig) error {
    if err := config.Validate(); err != nil {
        return fmt.Errorf("invalid configuration: %w", err)
    }
    
    // Implementation...
    return nil
}

// Good: Use functional options for complex configurations
func NewClusterService(opts ...ServiceOption) *ClusterService {
    s := &ClusterService{
        timeout: DefaultTimeout,
        retries: DefaultRetries,
    }
    
    for _, opt := range opts {
        opt(s)
    }
    
    return s
}

type ServiceOption func(*ClusterService)

func WithTimeout(timeout time.Duration) ServiceOption {
    return func(s *ClusterService) {
        s.timeout = timeout
    }
}
```

#### 4. Error Handling

```go
// Use custom error types for domain-specific errors
type ClusterError struct {
    Type    ErrorType
    Message string
    Cause   error
}

func (e ClusterError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}

func (e ClusterError) Unwrap() error {
    return e.Cause
}

// Wrap errors with context
func (s *ClusterService) CreateCluster(config ClusterConfig) error {
    if err := s.provider.Create(config); err != nil {
        return ClusterError{
            Type:    ErrorTypeCreation,
            Message: "failed to create cluster",
            Cause:   err,
        }
    }
    return nil
}
```

#### 5. Documentation

```go
// Package documentation
// Package cluster provides Kubernetes cluster management functionality.
// It supports creating, deleting, and managing local K3d clusters for development.
package cluster

// Function documentation with examples
// CreateCluster creates a new K3d cluster with the specified configuration.
//
// Example:
//   config := ClusterConfig{
//       Name: "my-cluster",
//       Nodes: 3,
//   }
//   if err := service.CreateCluster(config); err != nil {
//       return fmt.Errorf("cluster creation failed: %w", err)
//   }
func CreateCluster(config ClusterConfig) error {
    // Implementation...
}
```

### File Organization

```text
internal/cluster/
├── cluster.go              # Package documentation and main types
├── models/
│   ├── config.go          # Configuration structures
│   ├── config_test.go     # Configuration tests
│   ├── validation.go      # Validation logic
│   └── validation_test.go # Validation tests
├── services/
│   ├── cluster.go         # Cluster service implementation
│   ├── cluster_test.go    # Service tests
│   └── interfaces.go      # Service interfaces
├── ui/
│   ├── prompts.go         # Interactive prompts
│   └── prompts_test.go    # UI tests
└── utils/
    ├── helpers.go         # Utility functions
    └── helpers_test.go    # Utility tests
```

## Development Workflow

### 1. Issue-Based Development

All contributions should be linked to a GitHub issue:

```bash
# 1. Find or create an issue
# 2. Comment on the issue to indicate you're working on it
# 3. Create a feature branch
git checkout -b feature/issue-123-add-cluster-scaling
```

### 2. Branch Naming

| Type | Format | Example |
|------|--------|---------|
| **Feature** | `feature/issue-number-description` | `feature/45-add-multi-node-support` |
| **Bug Fix** | `bugfix/issue-number-description` | `bugfix/67-fix-cluster-deletion` |
| **Documentation** | `docs/description` | `docs/update-contributing-guide` |
| **Hotfix** | `hotfix/description` | `hotfix/security-vulnerability` |

### 3. Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Commit Types

| Type | Description | Example |
|------|-------------|---------|
| **feat** | New feature | `feat(cluster): add multi-node cluster support` |
| **fix** | Bug fix | `fix(chart): resolve ArgoCD installation timeout` |
| **docs** | Documentation | `docs(readme): update installation instructions` |
| **style** | Code style changes | `style: format code with goimports` |
| **refactor** | Code refactoring | `refactor(services): extract common validation logic` |
| **test** | Test additions/changes | `test(cluster): add integration tests for deletion` |
| **chore** | Maintenance | `chore: update dependencies to latest versions` |

#### Examples

```bash
# Good commit messages
feat(cluster): add support for custom K8s versions
fix(bootstrap): handle missing Docker daemon gracefully
docs(api): add examples for cluster configuration
test(integration): add end-to-end bootstrap tests
refactor(ui): extract common prompt patterns

# Include issue references
feat(cluster): add cluster scaling functionality

- Add scale up/down commands
- Implement node management
- Add validation for scaling operations

Closes #123
```

### 4. Pull Request Process

#### Before Creating a PR

```bash
# 1. Ensure your branch is up to date
git checkout main
git pull upstream main
git checkout feature/your-branch
git rebase main

# 2. Run all checks
make test
make lint
make build

# 3. Verify integration tests pass
make test-integration

# 4. Check test coverage
make test-coverage
```

#### PR Template

Create your PR with this information:

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Related Issue
Closes #123

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Test coverage maintained/improved

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Code is documented
- [ ] Tests added for new functionality
- [ ] All CI checks pass

## Screenshots (if applicable)
```

#### PR Review Process

1. **Automated Checks**: All CI checks must pass
2. **Code Review**: At least one maintainer review required
3. **Testing**: Reviewer tests the changes locally
4. **Documentation**: Verify docs are updated if needed
5. **Merge**: Squash and merge when approved

## Testing Requirements

### Required Test Coverage

All contributions must include appropriate tests:

| Change Type | Required Tests | Coverage Target |
|-------------|----------------|----------------|
| **New Features** | Unit + Integration | 85% |
| **Bug Fixes** | Reproduction test + fix verification | 80% |
| **Refactoring** | Existing tests pass + new tests if needed | Maintain existing |
| **Documentation** | Examples work as documented | N/A |

### Test Implementation

```go
// Example: Adding a new feature with tests

// 1. Write the test first (TDD approach)
func TestClusterService_ScaleCluster(t *testing.T) {
    tests := []struct {
        name        string
        initialNodes int
        targetNodes  int
        wantErr     bool
    }{
        {
            name:        "scale up",
            initialNodes: 1,
            targetNodes:  3,
            wantErr:     false,
        },
        {
            name:        "scale down",
            initialNodes: 3,
            targetNodes:  1,
            wantErr:     false,
        },
        {
            name:        "invalid scale to zero",
            initialNodes: 3,
            targetNodes:  0,
            wantErr:     true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}

// 2. Implement the feature
func (s *ClusterService) ScaleCluster(name string, nodeCount int) error {
    if nodeCount < 1 {
        return ErrInvalidNodeCount
    }
    // Implementation...
}

// 3. Add integration test
//go:build integration
func TestClusterScaling_Integration(t *testing.T) {
    // Real cluster scaling test
}
```

## Code Review Guidelines

### For Contributors

#### Before Requesting Review

- [ ] **Self-review**: Review your own code first
- [ ] **Test thoroughly**: All tests pass locally
- [ ] **Document changes**: Update relevant documentation
- [ ] **Clean commits**: Squash fixup commits
- [ ] **Descriptive PR**: Clear description and context

#### Responding to Feedback

```bash
# Address feedback with additional commits
git add .
git commit -m "address review feedback: improve error messages"

# If major changes are needed, consider rebasing
git rebase -i HEAD~3  # Interactive rebase to clean up commits
```

### For Reviewers

#### Review Checklist

- [ ] **Functionality**: Does the code do what it's supposed to?
- [ ] **Tests**: Are there adequate tests with good coverage?
- [ ] **Performance**: Are there any obvious performance issues?
- [ ] **Security**: Are there any security concerns?
- [ ] **Style**: Does the code follow project conventions?
- [ ] **Documentation**: Is the code and functionality documented?
- [ ] **Error Handling**: Are errors handled appropriately?

#### Review Comments

```markdown
# Constructive feedback examples:

## Good feedback:
Consider using a context with timeout here to prevent hanging:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

## Suggestion for improvement:
This validation logic could be extracted to a separate function for reusability:
```go
func validateClusterName(name string) error {
    // validation logic
}
```

## Nitpick (non-blocking):
Consider using a more descriptive variable name: `clusterConfig` instead of `config`.
```

## Performance and Security

### Performance Guidelines

1. **Efficient Resource Usage**:
```go
// Good: Use buffered channels appropriately
results := make(chan Result, 10)

// Good: Proper context usage
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// Good: Avoid unnecessary allocations
var buf strings.Builder
for _, item := range items {
    buf.WriteString(item.String())
}
return buf.String()
```

2. **Concurrent Operations**:
```go
// Good: Use sync.WaitGroup for concurrent operations
var wg sync.WaitGroup
for _, cluster := range clusters {
    wg.Add(1)
    go func(c Cluster) {
        defer wg.Done()
        processCluster(c)
    }(cluster)
}
wg.Wait()
```

### Security Guidelines

1. **Input Validation**:
```go
func ValidateClusterName(name string) error {
    if len(name) == 0 {
        return ErrEmptyName
    }
    
    // Validate characters (alphanumeric, hyphens only)
    if !regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(name) {
        return ErrInvalidCharacters
    }
    
    return nil
}
```

2. **Command Injection Prevention**:
```go
// Good: Use proper command construction
cmd := exec.Command("k3d", "cluster", "create", name)

// Bad: Never construct commands from strings
// cmd := exec.Command("sh", "-c", fmt.Sprintf("k3d cluster create %s", name))
```

## Documentation Standards

### Code Documentation

```go
// Package-level documentation
// Package cluster provides functionality for managing Kubernetes clusters.
//
// This package supports creating, deleting, and managing K3d clusters
// for local development environments. It provides both programmatic
// access and CLI command integration.
//
// Example usage:
//   service := cluster.NewService()
//   config := cluster.Config{Name: "my-cluster", Nodes: 3}
//   if err := service.Create(config); err != nil {
//       log.Fatal(err)
//   }
package cluster

// Function documentation with examples
// CreateCluster creates a new K3d cluster with the specified configuration.
//
// The function validates the configuration, checks prerequisites, and
// creates the cluster using the K3d provider. It returns an error if
// the cluster already exists or if creation fails.
//
// Example:
//   config := ClusterConfig{
//       Name: "dev-cluster",
//       Nodes: 3,
//       K8sVersion: "v1.27.3",
//   }
//   err := service.CreateCluster(config)
//   if err != nil {
//       return fmt.Errorf("failed to create cluster: %w", err)
//   }
func CreateCluster(config ClusterConfig) error {
    // Implementation...
}
```

### User-Facing Documentation

When adding new features, update relevant documentation:

1. **Command Help Text**: Update command descriptions and examples
2. **User Guides**: Add to getting-started or development guides
3. **API Documentation**: Document new configuration options
4. **Examples**: Provide working examples

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Changelog

Maintain CHANGELOG.md with each release:

```markdown
## [1.2.0] - 2023-12-01

### Added
- Cluster scaling functionality (#123)
- Support for custom K8s versions (#145)

### Fixed
- ArgoCD installation timeout issues (#167)
- Cluster deletion race condition (#189)

### Changed
- Improved error messages for validation failures (#134)

### Deprecated
- Old cluster configuration format (will be removed in v2.0.0)
```

## Getting Help

### For Contributors

- **GitHub Issues**: Technical questions and bug reports
- **GitHub Discussions**: Feature ideas and general questions  
- **Code Comments**: Questions about specific implementations
- **PR Reviews**: Get feedback on your contributions

### Communication Guidelines

- **Be respectful**: Treat all community members with respect
- **Be constructive**: Provide actionable feedback and suggestions
- **Be patient**: Reviews and responses may take time
- **Be thorough**: Provide context and details in issues and PRs

## Examples of Good Contributions

### Small Bug Fix Example

```markdown
## Bug Fix: Handle empty cluster name gracefully

### Problem
When users provide an empty cluster name, the CLI crashes instead 
of showing a helpful error message.

### Solution
- Add validation in the command layer
- Return user-friendly error message
- Add test case for empty name validation

### Testing
- Added unit test for validation function
- Tested CLI command manually
- Verified error message is clear and actionable
```

### Feature Addition Example

```markdown
## Feature: Add cluster scaling support

### Motivation
Users need to scale their development clusters up/down based on 
resource requirements and testing scenarios.

### Implementation
- Added `cluster scale` command
- Implemented scaling logic in cluster service
- Added validation for scaling parameters
- Updated documentation with examples

### Testing
- Unit tests for scaling logic
- Integration tests with real K3d clusters
- End-to-end tests for CLI command
- Performance testing for large scale operations

### Documentation
- Updated CLI help text
- Added examples to user guide
- Updated architecture documentation
```

Ready to contribute? Start by checking our [open issues](https://github.com/flamingo-stack/openframe-cli/issues) and find something that interests you! 🚀

## Quick Contribution Checklist

- [ ] Read this contributing guide
- [ ] Set up development environment
- [ ] Find an issue to work on (or create one)
- [ ] Create feature branch
- [ ] Write tests first (TDD)
- [ ] Implement the feature/fix
- [ ] Update documentation
- [ ] Run all checks locally
- [ ] Create descriptive PR
- [ ] Respond to review feedback
- [ ] Celebrate your contribution! 🎉