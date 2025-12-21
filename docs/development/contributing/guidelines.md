# Contributing Guidelines

Thank you for your interest in contributing to OpenFrame CLI! This guide outlines our development practices, code standards, and contribution workflow to help you make successful contributions.

## Code Style and Conventions

### Go Code Standards

OpenFrame CLI follows standard Go conventions with some additional guidelines:

#### Formatting and Imports

```go
// ✅ Good: Standard formatting
package cluster

import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

// ❌ Bad: Mixed import grouping  
import (
    "fmt"
    "github.com/spf13/cobra"
    "context"
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)
```

#### Function and Variable Naming

```go
// ✅ Good: Clear, descriptive names
func CreateClusterWithConfig(name string, config ClusterConfig) error {
    clusterExists, err := checkClusterExists(name)
    if err != nil {
        return fmt.Errorf("failed to check cluster existence: %w", err)
    }
    
    if clusterExists {
        return NewClusterExistsError(name)
    }
    
    return nil
}

// ❌ Bad: Abbreviated or unclear names
func CreateClstr(n string, cfg ClusterConfig) error {
    exists, e := chkClstr(n)
    if e != nil {
        return e
    }
    // ...
}
```

#### Error Handling

```go
// ✅ Good: Wrap errors with context
func (s *ClusterService) Create(name string) error {
    if err := s.validateName(name); err != nil {
        return fmt.Errorf("invalid cluster name %q: %w", name, err)
    }
    
    if err := s.k3dProvider.CreateCluster(name); err != nil {
        return fmt.Errorf("failed to create cluster %q: %w", name, err)
    }
    
    return nil
}

// ❌ Bad: Silent errors or poor context
func (s *ClusterService) Create(name string) error {
    s.validateName(name)  // Ignoring error
    
    err := s.k3dProvider.CreateCluster(name)
    if err != nil {
        return err  // No context added
    }
    
    return nil
}
```

#### Interface Design

```go
// ✅ Good: Small, focused interfaces
type ClusterProvider interface {
    CreateCluster(name string, config ClusterConfig) error
    DeleteCluster(name string) error  
    GetClusterStatus(name string) (*ClusterStatus, error)
}

type UIService interface {
    ShowSuccess(message string)
    ShowError(err error)
    PromptForInput(prompt string) (string, error)
}

// ❌ Bad: Large, multi-purpose interfaces
type ClusterManager interface {
    CreateCluster(name string, config ClusterConfig) error
    DeleteCluster(name string) error
    InstallArgoCD(clusterName string) error  // Different concern
    ShowProgress(message string)              // UI concern
    ValidateConfig(config Config) error       // Validation concern
}
```

### Documentation Standards

#### Function Documentation

```go
// CreateCluster creates a new K3d cluster with the specified configuration.
//
// The cluster name must be unique and follow DNS naming conventions.
// If a cluster with the same name already exists, an error is returned.
//
// Parameters:
//   - name: The cluster name (must be unique and valid DNS name)
//   - config: Cluster configuration including node count and port mappings
//
// Returns an error if:
//   - The cluster name is invalid
//   - A cluster with the same name already exists  
//   - Docker is not running or accessible
//   - K3d binary is not found
func (p *K3dProvider) CreateCluster(name string, config ClusterConfig) error {
    // Implementation
}
```

#### Package Documentation

```go
// Package cluster provides Kubernetes cluster management functionality.
//
// This package integrates with K3d to create and manage lightweight
// Kubernetes clusters for development. It supports cluster creation,
// deletion, and status monitoring.
//
// Example usage:
//
//     provider := cluster.NewK3dProvider()
//     config := cluster.ClusterConfig{Nodes: 1}
//     err := provider.CreateCluster("my-cluster", config)
//
package cluster
```

### File Organization

#### Directory Structure

```
internal/cluster/
├── service.go              # Main service implementation
├── service_test.go         # Service unit tests
├── k3d_provider.go         # K3d integration
├── k3d_provider_test.go    # Provider tests
├── models.go               # Data models
├── errors.go               # Custom errors
└── prerequisites.go        # Prerequisite checks
```

#### File Naming Conventions

| File Type | Pattern | Example |
|-----------|---------|---------|
| **Implementation** | `{component}.go` | `service.go`, `k3d_provider.go` |
| **Tests** | `{component}_test.go` | `service_test.go` |
| **Integration Tests** | `{component}_integration_test.go` | `k3d_provider_integration_test.go` |
| **Models** | `models.go` or `{domain}_models.go` | `models.go`, `cluster_models.go` |
| **Errors** | `errors.go` | `errors.go` |

## Branch Naming and PR Process

### Branch Naming Convention

Use descriptive branch names that indicate the type of change:

| Type | Pattern | Example |
|------|---------|---------|
| **Feature** | `feature/{description}` | `feature/cluster-status-command` |
| **Bug Fix** | `fix/{description}` | `fix/cluster-deletion-timeout` |
| **Documentation** | `docs/{description}` | `docs/update-setup-guide` |
| **Refactor** | `refactor/{description}` | `refactor/extract-ui-components` |
| **Test** | `test/{description}` | `test/add-integration-tests` |

### Pull Request Workflow

#### 1. Preparation

```bash
# Sync with main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/new-cluster-command

# Make your changes
# ... code changes ...

# Ensure code quality
make lint
make test
go mod tidy
```

#### 2. Commit Message Format

Follow conventional commit format:

```bash
# Format: <type>(<scope>): <description>
#
# Examples:
git commit -m "feat(cluster): add cluster status command"
git commit -m "fix(bootstrap): handle docker permission error"
git commit -m "docs(readme): update installation instructions"
git commit -m "test(cluster): add integration tests for k3d provider"
git commit -m "refactor(ui): extract common prompt components"
```

#### Commit Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(dev): add skaffold integration` |
| `fix` | Bug fix | `fix(cluster): resolve port conflict error` |
| `docs` | Documentation | `docs(api): add service documentation` |
| `style` | Code style (no logic change) | `style: fix linting issues` |
| `refactor` | Code refactoring | `refactor(shared): extract error types` |
| `test` | Test additions | `test(bootstrap): add unit tests` |
| `chore` | Maintenance tasks | `chore: update dependencies` |

#### 3. Pull Request Creation

**Title Format:**
```
<type>(<scope>): <description>

# Examples:
feat(cluster): add cluster status command
fix(bootstrap): handle docker permission error
```

**Description Template:**
```markdown
## Description
Brief summary of the changes and why they are needed.

## Changes Made
- [ ] Added new cluster status command
- [ ] Updated CLI help text
- [ ] Added unit tests
- [ ] Updated documentation

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass (if applicable)
- [ ] Manual testing completed

## Screenshots (if UI changes)
[Add screenshots if relevant]

## Related Issues
Closes #123
Fixes #456

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] No breaking changes (or marked as breaking)
```

### Review Process

#### Self-Review Checklist

Before requesting review, ensure:

- [ ] **Code Quality**
  - [ ] Follows Go conventions and project style
  - [ ] Functions are well-documented  
  - [ ] Error handling is comprehensive
  - [ ] No hardcoded values (use constants/config)

- [ ] **Testing**
  - [ ] Unit tests added for new functionality
  - [ ] Integration tests added if needed
  - [ ] All tests pass (`make test`)
  - [ ] Test coverage maintained or improved

- [ ] **Documentation**
  - [ ] Public functions are documented
  - [ ] README updated if needed
  - [ ] Examples provided for complex features

- [ ] **Performance**
  - [ ] No obvious performance regressions
  - [ ] Async operations use context properly
  - [ ] Resource cleanup implemented

#### Code Review Guidelines

**For Reviewers:**

- **Be Constructive**: Provide specific, actionable feedback
- **Focus on Impact**: Prioritize correctness, performance, and maintainability
- **Ask Questions**: If unclear about intent, ask rather than assume
- **Suggest Alternatives**: Offer better approaches when possible

**Review Checklist:**
- [ ] Code is clear and maintainable
- [ ] Error handling is appropriate
- [ ] Tests adequately cover the changes
- [ ] Documentation is sufficient
- [ ] No obvious security issues
- [ ] Performance impact is acceptable

**Example Review Comments:**

```markdown
# ✅ Good feedback
This function could be simplified by extracting the validation logic:

```go
// Consider extracting this:
func validateClusterConfig(config ClusterConfig) error {
    // validation logic here
}
```

# ✅ Good feedback
Consider using a constant for this magic number:

```go
const DefaultTimeoutSeconds = 300
```

# ❌ Poor feedback
"This is wrong"  // Not helpful

# ❌ Poor feedback  
"I don't like this approach"  // No alternative provided
```

## Commit Message Guidelines

### Format Structure

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Examples

#### Simple Changes
```bash
feat(cluster): add delete command

fix(ui): resolve spinner display issue

docs: update contributing guidelines
```

#### Complex Changes
```bash
feat(bootstrap): add deployment mode selection

Add interactive deployment mode selection to bootstrap command.
Users can now choose between oss-tenant, saas-tenant, and 
saas-shared modes during cluster setup.

- Add deployment mode prompts
- Update bootstrap service logic
- Add validation for mode selection
- Update help text and examples

Closes #45
```

#### Breaking Changes
```bash
feat(api)!: change cluster config structure

BREAKING CHANGE: ClusterConfig struct has been restructured.
The 'ports' field is now 'portMappings' and expects a different format.

Before:
  config.Ports = []string{"8080:80"}

After:
  config.PortMappings = []PortMapping{{Host: 8080, Container: 80}}

Migration guide available at docs/migration/v2.md
```

## Review Checklist

### For Authors

Before submitting a PR:

#### Code Quality
- [ ] Code follows project conventions
- [ ] Functions are focused and well-named
- [ ] Error handling includes proper context
- [ ] No TODO comments without issue references
- [ ] Imports are properly organized

#### Testing
- [ ] New code has appropriate test coverage
- [ ] Tests are independent and deterministic
- [ ] Integration tests added for external dependencies
- [ ] Manual testing completed for UI changes

#### Documentation
- [ ] Public APIs are documented
- [ ] Complex logic has inline comments
- [ ] README updated if user-facing changes
- [ ] Examples added for new features

#### Performance
- [ ] No obvious performance regressions
- [ ] Large operations are cancellable (use context)
- [ ] Resources are properly cleaned up
- [ ] Caching is used appropriately

#### Security
- [ ] Input validation implemented
- [ ] No hardcoded secrets or credentials
- [ ] External commands properly escaped
- [ ] File permissions are appropriate

### For Reviewers

#### Functionality
- [ ] Code does what it claims to do
- [ ] Edge cases are handled appropriately
- [ ] Error conditions are handled gracefully
- [ ] User experience is intuitive

#### Design
- [ ] Code follows established patterns
- [ ] Dependencies are appropriate
- [ ] Interfaces are well-designed
- [ ] Separation of concerns is maintained

#### Testing
- [ ] Test coverage is adequate
- [ ] Tests verify the right behavior
- [ ] Test names clearly indicate intent
- [ ] Tests are maintainable

## Development Tools Integration

### Pre-commit Hooks

Set up pre-commit hooks to catch issues early:

```bash
# .git/hooks/pre-commit
#!/bin/bash

echo "Running pre-commit checks..."

# Format code
echo "Formatting code..."
gofmt -s -w .
goimports -w .

# Lint code  
echo "Linting..."
if ! golangci-lint run; then
    echo "❌ Linting failed"
    exit 1
fi

# Run tests
echo "Running tests..."
if ! go test -short ./...; then
    echo "❌ Tests failed"
    exit 1
fi

# Check go.mod
echo "Tidying go.mod..."
go mod tidy
if ! git diff --quiet go.mod go.sum; then
    echo "❌ go.mod/go.sum not tidy"
    exit 1
fi

echo "✅ Pre-commit checks passed"
```

### IDE Configuration

#### VS Code Settings

```json
{
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "go.testFlags": ["-v", "-count=1"],
  "go.buildFlags": ["-v"]
}
```

### Continuous Integration

Our CI pipeline runs on every PR:

```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Lint
      run: |
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
        golangci-lint run
    
    - name: Test
      run: go test -race -cover ./...
    
    - name: Build
      run: go build .
```

## Getting Help

### Where to Ask Questions

- **GitHub Discussions**: General questions and feature discussions
- **GitHub Issues**: Bug reports and specific problems
- **Code Review**: Implementation details and architectural questions

### Common Issues

#### Import Organization
```go
// ✅ Correct order: standard -> third-party -> local
import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    "github.com/stretchr/testify/assert"
    
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)
```

#### Error Handling
```go
// ✅ Good: Wrap errors with context
return fmt.Errorf("failed to create cluster %q: %w", name, err)

// ❌ Bad: Return raw errors
return err
```

#### Testing Patterns
```go
// ✅ Good: Table-driven tests for multiple scenarios
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "my-cluster", false},
        {"empty name", "", true},
        {"invalid chars", "my_cluster", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateClusterName(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Release Process

### Version Management

OpenFrame CLI follows semantic versioning:

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)  
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

For maintainers preparing releases:

- [ ] Update CHANGELOG.md
- [ ] Update version in main.go
- [ ] Create release notes
- [ ] Tag release (`git tag v1.2.3`)
- [ ] Push tags (`git push --tags`)
- [ ] GitHub Actions builds and publishes binaries

## Contributing Types

We welcome various types of contributions:

| Type | Description | Examples |
|------|-------------|----------|
| **Bug Fixes** | Fix existing functionality | Resolve command errors, fix edge cases |
| **Features** | Add new functionality | New commands, integrations |
| **Documentation** | Improve docs and examples | Guides, API docs, tutorials |
| **Testing** | Improve test coverage | Unit tests, integration tests |
| **Performance** | Optimize existing code | Reduce latency, memory usage |
| **Refactoring** | Improve code quality | Extract functions, organize code |

## Recognition

Contributors are recognized in:

- **Releases Notes**: Major contributors mentioned
- **Contributors Graph**: GitHub automatically tracks contributions  
- **Special Thanks**: Significant contributions get special recognition

Thank you for contributing to OpenFrame CLI! Your efforts help make Kubernetes development easier for everyone. 🚀

## Next Steps

Ready to contribute?

1. **Pick an Issue**: Browse [good first issues](https://github.com/flamingo-stack/openframe-cli/labels/good%20first%20issue)
2. **Set Up Development**: Follow [Local Development Guide](../setup/local-development.md)
3. **Make Your Changes**: Follow these guidelines
4. **Submit a PR**: Use our templates and checklist

Every contribution, no matter how small, is valuable and appreciated!