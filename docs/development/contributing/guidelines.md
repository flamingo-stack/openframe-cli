# Contributing Guidelines

Welcome to the OpenFrame CLI project! We're excited that you want to contribute. This guide provides everything you need to know about contributing code, documentation, and features to make the process smooth and productive for everyone.

## Getting Started

### Before You Start

1. **Read the documentation**: Familiarize yourself with OpenFrame CLI by reading the [Introduction](../../getting-started/introduction.md) and trying the [Quick Start](../../getting-started/quick-start.md)
2. **Set up your environment**: Follow the [Development Environment Setup](../setup/environment.md) guide
3. **Run the project locally**: Complete the [Local Development](../setup/local-development.md) setup
4. **Understand the architecture**: Review the [Architecture Overview](../architecture/overview.md)

### Ways to Contribute

| Contribution Type | Examples | Best For |
|------------------|----------|----------|
| **Bug Reports** | Found a bug? Report it with reproduction steps | All contributors |
| **Feature Requests** | Ideas for new functionality or improvements | Users and developers |
| **Bug Fixes** | Fix reported issues or bugs you've found | Developers |
| **New Features** | Implement requested features or your own ideas | Experienced contributors |
| **Documentation** | Improve guides, fix typos, add examples | Writers and developers |
| **Tests** | Add test coverage, improve test quality | Developers focused on quality |

## Code Contribution Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub (via web interface)

# Clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git

# Verify remotes
git remote -v
```

### 2. Create a Feature Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description

# Examples:
git checkout -b feature/add-cluster-templates
git checkout -b fix/bootstrap-timeout-error
git checkout -b docs/improve-installation-guide
```

### 3. Make Your Changes

Follow our coding standards and best practices:

```bash
# Make your changes
vim internal/cluster/service.go

# Build and test frequently
go build -o bin/openframe main.go
go test ./internal/cluster/...

# Run full test suite before committing
go test ./...
```

### 4. Commit Your Changes

We follow [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
# Stage your changes
git add .

# Commit with conventional format
git commit -m "type(scope): description"

# Examples:
git commit -m "feat(cluster): add cluster template support"
git commit -m "fix(bootstrap): handle timeout errors gracefully"
git commit -m "docs: improve installation instructions"
git commit -m "test: add integration tests for chart installation"
```

#### Commit Types

| Type | Use Case | Examples |
|------|----------|----------|
| `feat` | New feature | `feat(cli): add --dry-run flag to bootstrap` |
| `fix` | Bug fix | `fix(cluster): resolve port conflict issue` |
| `docs` | Documentation | `docs(readme): update installation instructions` |
| `style` | Code style changes | `style: format code with gofmt` |
| `refactor` | Code refactoring | `refactor(service): simplify error handling` |
| `test` | Adding tests | `test(integration): add k3d cluster tests` |
| `chore` | Maintenance tasks | `chore: update dependencies` |

### 5. Push and Create Pull Request

```bash
# Push your branch
git push origin feature/your-feature-name

# Create pull request via GitHub web interface
```

## Pull Request Guidelines

### PR Title and Description

**Title Format**: Follow the same conventional commit format
```
feat(cluster): add support for cluster templates
```

**Description Template**:
```markdown
## Description
Brief description of what this PR does and why.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Screenshots/Examples (if applicable)
Include screenshots or example command outputs

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review of code completed
- [ ] Tests added for new functionality
- [ ] Documentation updated (if applicable)
- [ ] All tests pass locally
```

### PR Requirements

**Before submitting a PR:**

✅ **Code Quality**
- [ ] Code follows Go best practices and project conventions
- [ ] All tests pass: `go test ./...`
- [ ] No linting errors: `golangci-lint run`
- [ ] Code is properly formatted: `go fmt ./...`

✅ **Testing**
- [ ] Unit tests added for new functionality
- [ ] Integration tests updated if needed
- [ ] Manual testing completed
- [ ] Test coverage maintained or improved

✅ **Documentation**
- [ ] Code comments added for complex logic
- [ ] Public functions have GoDoc comments
- [ ] User-facing documentation updated
- [ ] CHANGELOG.md updated (for significant changes)

✅ **Git**
- [ ] Commit messages follow conventional format
- [ ] Branch is up to date with main
- [ ] Single feature/fix per PR
- [ ] No merge commits (use rebase)

## Code Style Guide

### Go Code Standards

**File Organization:**
```go
// Package comment
package cluster

// Standard library imports
import (
    "context"
    "fmt"
    "os"
)

// Third-party imports
import (
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
)

// Local imports
import (
    "github.com/flamingo-stack/openframe-cli/internal/shared"
)
```

**Function Documentation:**
```go
// CreateCluster creates a new K3d cluster with the specified configuration.
// It validates the cluster name, checks for existing clusters, and delegates
// to the K3d provider for the actual cluster creation.
//
// Parameters:
//   - name: The cluster name (must be valid DNS name)
//   - config: Cluster configuration including ports, nodes, etc.
//
// Returns an error if validation fails or cluster creation fails.
func (s *Service) CreateCluster(name string, config ClusterConfig) error {
    // Implementation
}
```

**Error Handling:**
```go
// Good: Wrap errors with context
if err := s.provider.CreateCluster(name, config); err != nil {
    return fmt.Errorf("failed to create cluster %q: %w", name, err)
}

// Good: Use typed errors for specific cases
var ErrClusterNotFound = errors.New("cluster not found")

func (s *Service) GetCluster(name string) (*Cluster, error) {
    cluster, err := s.provider.GetCluster(name)
    if err != nil {
        if isNotFoundError(err) {
            return nil, ErrClusterNotFound
        }
        return nil, fmt.Errorf("failed to get cluster: %w", err)
    }
    return cluster, nil
}
```

**Interface Design:**
```go
// Interfaces should be small and focused
type ClusterProvider interface {
    CreateCluster(name string, config ClusterConfig) error
    DeleteCluster(name string) error
    GetCluster(name string) (*Cluster, error)
}

// Use dependency injection
type Service struct {
    provider ClusterProvider
    logger   Logger
}

func NewService(provider ClusterProvider, logger Logger) *Service {
    return &Service{
        provider: provider,
        logger:   logger,
    }
}
```

### Naming Conventions

| Element | Convention | Examples |
|---------|------------|----------|
| **Packages** | Short, lowercase, no underscores | `cluster`, `chart`, `bootstrap` |
| **Files** | Lowercase with underscores | `service.go`, `cluster_test.go` |
| **Types** | PascalCase | `ClusterService`, `InstallRequest` |
| **Functions** | PascalCase (exported), camelCase (private) | `CreateCluster()`, `validateName()` |
| **Variables** | camelCase | `clusterName`, `deploymentMode` |
| **Constants** | PascalCase or UPPER_CASE | `DefaultTimeout`, `MAX_RETRIES` |

### Project Structure Conventions

```go
// Service layer structure
type Service struct {
    // Dependencies (interfaces preferred)
    provider    Provider
    validator   Validator
    logger      Logger
    
    // Configuration
    config      Config
    
    // State (if needed, prefer stateless)
    // Use sparingly and document why needed
}

// Constructor pattern
func NewService(deps Dependencies) *Service {
    return &Service{
        provider:  deps.Provider,
        validator: deps.Validator,
        logger:    deps.Logger,
        config:    deps.Config,
    }
}

// Main business method
func (s *Service) Execute(ctx context.Context, req Request) (*Response, error) {
    // 1. Validate input
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // 2. Execute business logic
    result, err := s.provider.DoSomething(ctx, req.ToProviderRequest())
    if err != nil {
        return nil, fmt.Errorf("operation failed: %w", err)
    }
    
    // 3. Return response
    return &Response{Data: result}, nil
}
```

## Testing Standards

### Test Organization

```go
// Test file structure
func TestServiceName_MethodName(t *testing.T) {
    // Arrange - Set up test data and mocks
    
    // Act - Execute the code being tested
    
    // Assert - Verify the results
}

func TestServiceName_MethodName_ErrorCase(t *testing.T) {
    // Test specific error scenarios
}
```

### Test Naming

```go
// Good test names - describe what's being tested
func TestClusterService_Create(t *testing.T)
func TestClusterService_Create_InvalidName(t *testing.T)
func TestClusterService_Create_ClusterExists(t *testing.T)
func TestClusterService_Delete_ClusterNotFound(t *testing.T)

// Table-driven tests for multiple scenarios
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "my-cluster", false},
        {"empty name", "", true},
        {"invalid chars", "cluster_name", true},
    }
    // ...
}
```

### Mock Usage

```go
// Use interfaces for dependencies
type MockProvider struct {
    mock.Mock
}

func (m *MockProvider) CreateCluster(name string, config ClusterConfig) error {
    args := m.Called(name, config)
    return args.Error(0)
}

// In tests
func TestService_Create(t *testing.T) {
    // Setup mock
    mockProvider := &MockProvider{}
    service := NewService(mockProvider)
    
    // Set expectations
    mockProvider.On("CreateCluster", "test", mock.AnythingOfType("ClusterConfig")).
        Return(nil)
    
    // Execute and verify
    err := service.Create("test")
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

## Documentation Standards

### Code Documentation

```go
// Package documentation
// Package cluster provides functionality for managing K3d Kubernetes clusters.
// It offers a high-level service interface for cluster lifecycle operations
// including creation, deletion, status checking, and cleanup.
package cluster

// Type documentation
// ClusterConfig holds configuration options for cluster creation.
// All fields are optional and will use sensible defaults if not specified.
type ClusterConfig struct {
    // Name is the cluster name (required)
    Name string `yaml:"name"`
    
    // Nodes specifies the number of worker nodes (default: 1)
    Nodes int `yaml:"nodes,omitempty"`
    
    // Ports maps host ports to cluster ports
    Ports []PortMapping `yaml:"ports,omitempty"`
}
```

### User Documentation

- **Clear headings**: Use descriptive headings and proper hierarchy
- **Code examples**: Include working, copy-pastable examples
- **Prerequisites**: Always list what's needed before starting
- **Expected output**: Show users what they should see
- **Troubleshooting**: Include common issues and solutions

### Documentation Examples

```markdown
# Good documentation structure

## Overview
Brief description of what this does and why it's useful.

## Prerequisites
- Docker installed and running
- kubectl configured

## Quick Start
```bash
# Simple example that works immediately
openframe bootstrap my-cluster
```

## Usage Examples

### Basic Usage
```bash
# Most common use case
openframe cluster create production
```

### Advanced Options
```bash
# More complex example with explanation
openframe cluster create staging \
  --nodes 3 \
  --port 8080:80 \
  --verbose
```

## Troubleshooting

### Port Already in Use
If you see "port 6443 already in use":

```bash
# Check what's using the port
sudo lsof -i :6443

# Kill the process or use different port
openframe cluster create --api-port 6444
```
```

## Review Process

### Review Checklist

**For Authors:**
- [ ] PR description clearly explains the change and why it's needed
- [ ] All tests pass locally
- [ ] Code follows project conventions
- [ ] Documentation updated
- [ ] Commit messages are clear and follow conventional format

**For Reviewers:**
- [ ] Code is clear and maintainable
- [ ] Tests adequately cover the changes
- [ ] Documentation is accurate and helpful
- [ ] No security issues or performance problems
- [ ] Changes align with project goals

### Review Guidelines

**For Authors:**
1. **Respond promptly** to review feedback
2. **Address all comments** or explain why not
3. **Update the PR** based on feedback
4. **Request re-review** when ready
5. **Be open to feedback** and willing to make changes

**For Reviewers:**
1. **Be constructive** - suggest improvements, don't just point out problems
2. **Explain the "why"** - help the author understand the reasoning
3. **Distinguish between** must-fix issues and suggestions
4. **Review promptly** - don't block contributors unnecessarily
5. **Approve when ready** - don't hold up good changes for perfection

### Review Comments Examples

**Good review comments:**
```markdown
# Constructive feedback
Consider using a switch statement here instead of multiple if-else 
statements for better readability.

# Suggestion with explanation
We should validate the input parameters before processing to fail fast
and provide better error messages to users.

# Requesting clarification
Could you add a comment explaining why we need to retry here? 
It's not immediately obvious from the code.
```

**Review approval:**
```markdown
Looks great! The new validation logic makes the error messages much 
clearer for users. Tests cover all the edge cases well.
```

## Release Process

### Version Numbering

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **Major**: Breaking changes
- **Minor**: New features (backward compatible)
- **Patch**: Bug fixes (backward compatible)

### Changelog

Update `CHANGELOG.md` for significant changes:

```markdown
## [1.2.0] - 2024-01-15

### Added
- Cluster template support for predefined configurations
- New `--dry-run` flag for bootstrap command

### Fixed
- Bootstrap timeout handling for slow networks
- Port conflict detection and resolution

### Changed
- Improved error messages for cluster creation failures

### Deprecated
- `--legacy-mode` flag (will be removed in v2.0.0)
```

## Getting Help

### Resources

| Resource | Purpose | Where to Use |
|----------|---------|--------------|
| **GitHub Issues** | Bug reports, feature requests | Questions about functionality |
| **GitHub Discussions** | General questions, ideas | Architecture discussions, usage help |
| **Pull Request Reviews** | Code feedback | Technical implementation questions |
| **Documentation** | Usage guides, API reference | How-to questions |

### Communication Guidelines

1. **Be respectful** and professional in all interactions
2. **Search first** - check if your question has been asked before
3. **Provide context** - include relevant information and examples
4. **Use appropriate channels** - bugs go to Issues, discussions to Discussions
5. **Follow up** - let us know if suggested solutions work

### Getting Support

```markdown
# Good issue template
**Describe the bug**
Bootstrap command fails with timeout error on slow networks.

**To Reproduce**
1. Run `openframe bootstrap test-cluster` on network with <10Mbps
2. Command times out after 5 minutes
3. Cluster is partially created but ArgoCD installation fails

**Expected behavior**
Command should wait longer or provide progress feedback.

**Environment:**
- OS: macOS 13.4
- OpenFrame CLI version: v1.1.0
- Docker version: 20.10.21
- K3d version: 5.4.6

**Additional context**
This happens consistently on networks slower than 10Mbps.
```

## Recognition

Contributors are recognized in:
- **Contributors list** in README.md
- **Release notes** for significant contributions
- **GitHub contributors** page
- **Special thanks** in major releases

Thank you for contributing to OpenFrame CLI! Your contributions help make the project better for everyone. 🎉

---

> 💡 **Remember**: The best contributions are those that help other users. Whether it's fixing a bug, adding a feature, or improving documentation, every contribution makes OpenFrame CLI more useful and accessible.