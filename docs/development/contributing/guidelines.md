# Contributing Guidelines

Welcome to the OpenFrame CLI development community! This guide covers everything you need to know to contribute effectively to the project.

## Quick Start for Contributors

```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# 3. Set up development environment  
make dev-setup

# 4. Create feature branch
git checkout -b feature/your-feature-name

# 5. Make changes and test
make test

# 6. Submit pull request
git push origin feature/your-feature-name
```

## Code Style and Conventions

### Go Code Style

We follow standard Go conventions with some additional requirements:

#### Formatting

```bash
# All code must be formatted with gofmt/goimports
make fmt

# Use goimports (preferred over gofmt)
goimports -w .

# Verify formatting
make lint
```

#### Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| **Packages** | lowercase, single word | `cluster`, `bootstrap` |
| **Types** | PascalCase | `ClusterConfig`, `ServiceInterface` |
| **Functions** | PascalCase (exported), camelCase (internal) | `NewService()`, `validateInput()` |
| **Constants** | PascalCase or UPPER_SNAKE_CASE | `DefaultTimeout`, `MAX_RETRIES` |
| **Variables** | camelCase | `clusterName`, `serviceConfig` |

#### Code Organization

```go
// Package declaration and imports
package cluster

import (
    // Standard library first
    "context"
    "fmt"
    "time"
    
    // Third-party packages
    "github.com/spf13/cobra"
    
    // Local packages
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

// Constants and variables
const (
    DefaultClusterCPU    = 2
    DefaultClusterMemory = 4
)

// Interface definitions
type ServiceInterface interface {
    Create(config ClusterConfig) error
    Delete(name string) error
}

// Type definitions
type ClusterConfig struct {
    Name   string `yaml:"name" json:"name"`
    CPU    int    `yaml:"cpu" json:"cpu"`
    Memory int    `yaml:"memory" json:"memory"`
}

// Constructor functions
func NewService() *Service {
    return &Service{
        // ...
    }
}

// Methods (receiver methods grouped by type)
func (s *Service) Create(config ClusterConfig) error {
    // Implementation
}
```

#### Error Handling

```go
// Use wrapped errors for context
func (s *Service) Create(config ClusterConfig) error {
    if err := s.validate(config); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    if err := s.provider.Create(config); err != nil {
        return fmt.Errorf("failed to create cluster %s: %w", config.Name, err)
    }
    
    return nil
}

// Define package-specific error variables
var (
    ErrClusterNotFound = errors.New("cluster not found")
    ErrInvalidName     = errors.New("invalid cluster name")
)

// Use error variables for known conditions
func (s *Service) GetCluster(name string) (*Cluster, error) {
    cluster, err := s.provider.Get(name)
    if err != nil {
        if isNotFoundError(err) {
            return nil, ErrClusterNotFound
        }
        return nil, fmt.Errorf("failed to get cluster: %w", err)
    }
    return cluster, nil
}
```

#### Documentation

```go
// Package documentation
// Package cluster provides Kubernetes cluster management functionality.
// It supports creating, deleting, and managing local K3d clusters for
// OpenFrame development environments.
package cluster

// Interface documentation
// ServiceInterface defines the contract for cluster management operations.
// All implementations must provide cluster lifecycle management capabilities.
type ServiceInterface interface {
    // Create creates a new cluster with the specified configuration.
    // It returns an error if the cluster name already exists or if
    // the underlying provider fails.
    Create(config ClusterConfig) error
    
    // Delete removes an existing cluster by name.
    // It returns ErrClusterNotFound if the cluster doesn't exist.
    Delete(name string) error
}

// Complex function documentation
// ValidateClusterName validates that a cluster name meets requirements.
//
// Rules:
// - Must not be empty
// - Must be 1-63 characters long
// - Must contain only lowercase letters, numbers, and hyphens
// - Must start and end with alphanumeric character
//
// Returns an error describing the first validation failure found.
func ValidateClusterName(name string) error {
    // Implementation
}
```

### Command Structure

All commands follow a consistent structure:

```go
// cmd/example/example.go
package example

import (
    "github.com/spf13/cobra"
    "github.com/flamingo-stack/openframe-cli/internal/example"
)

// GetExampleCmd returns the example command
func GetExampleCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "example [args]",
        Short: "Brief description (< 50 chars)",
        Long: `Detailed description with examples.

This command does something specific and provides
examples of how to use it effectively.

Examples:
  openframe example                          # Basic usage
  openframe example --flag value            # With options
  openframe example arg1 arg2               # With arguments`,
        Args: cobra.MaximumNArgs(2),
        RunE: func(cmd *cobra.Command, args []string) error {
            return example.NewService().Execute(cmd, args)
        },
    }

    // Add flags with clear descriptions
    cmd.Flags().String("deployment-mode", "", "Deployment mode: oss-tenant, saas-tenant, saas-shared")
    cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
    cmd.Flags().Bool("non-interactive", false, "Skip interactive prompts")

    return cmd
}
```

## Branch Naming and PR Process

### Branch Naming

| Type | Format | Example |
|------|--------|---------|
| **Feature** | `feature/description` | `feature/add-cluster-templates` |
| **Bug Fix** | `fix/description` | `fix/cluster-deletion-timeout` |
| **Documentation** | `docs/description` | `docs/update-installation-guide` |
| **Refactoring** | `refactor/description` | `refactor/extract-ui-components` |
| **Performance** | `perf/description` | `perf/optimize-cluster-creation` |

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
# Format
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]

# Examples
feat(cluster): add cluster template support

feat: add support for custom cluster templates
BREAKING CHANGE: default cluster configuration format changed

fix(bootstrap): resolve ArgoCD installation timeout
Closes #123

docs: update installation prerequisites

test: add integration tests for cluster lifecycle

refactor(ui): extract spinner component for reuse

perf(bootstrap): parallelize cluster creation and chart installation
```

#### Commit Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat: add cluster templates` |
| `fix` | Bug fix | `fix: resolve timeout issue` |
| `docs` | Documentation | `docs: update README` |
| `style` | Code style (no logic change) | `style: format code` |
| `refactor` | Code refactoring | `refactor: extract helper` |
| `perf` | Performance improvement | `perf: optimize startup` |
| `test` | Adding/updating tests | `test: add unit tests` |
| `chore` | Maintenance tasks | `chore: update dependencies` |

### Pull Request Process

#### 1. Pre-PR Checklist

```bash
# Before creating PR, ensure:
make quality           # Passes all quality checks
make test             # All tests pass
make test-integration # Integration tests pass (if applicable)
```

- [ ] Code follows style guidelines
- [ ] Tests added for new features
- [ ] Documentation updated
- [ ] Commits follow conventional format
- [ ] No merge commits (rebase if needed)

#### 2. PR Template

Use this template for your pull request:

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review of code completed
- [ ] Comments added to complex areas
- [ ] Documentation updated
- [ ] Tests added and passing

## Related Issues
Closes #123
Related to #456
```

#### 3. Review Process

1. **Automated Checks**: CI must pass
2. **Code Review**: At least one maintainer approval
3. **Testing**: All tests must pass
4. **Documentation**: Relevant docs updated

#### 4. Merge Requirements

- [ ] All CI checks pass
- [ ] At least 1 approving review from maintainer
- [ ] No merge conflicts
- [ ] Branch is up to date with main
- [ ] All conversations resolved

## Code Review Standards

### What We Look For

#### Code Quality
- **Readability**: Code is clear and self-documenting
- **Simplicity**: Solutions are as simple as possible  
- **Performance**: No obvious performance issues
- **Error Handling**: Comprehensive error handling with context

#### Testing
- **Coverage**: New code has appropriate test coverage
- **Test Quality**: Tests are meaningful and not brittle
- **Edge Cases**: Error conditions and edge cases tested

#### Design
- **Interface Design**: Clean, focused interfaces
- **Separation of Concerns**: Proper layering and abstraction
- **Extensibility**: Code can be extended without major changes

### Review Etiquette

#### For Authors
- **Small PRs**: Keep changes focused and reviewable
- **Context**: Provide clear description and motivation  
- **Responsive**: Address feedback promptly and thoughtfully
- **Testing**: Test your changes thoroughly

#### For Reviewers
- **Constructive**: Provide helpful, specific feedback
- **Timely**: Review within 1-2 business days
- **Thorough**: Check logic, style, tests, and documentation
- **Respectful**: Be kind and assume positive intent

### Review Comments

#### Good Review Comments
```markdown
// Good - Specific and actionable
Consider using a context.WithTimeout here to prevent hanging 
if the external service is unresponsive.

// Good - Explains reasoning
This error message could be more helpful to users. Consider 
adding the actual cluster name and suggested remediation steps.

// Good - Positive feedback
Nice use of the strategy pattern here! This makes it easy to 
add new deployment modes.
```

#### Comments to Avoid
```markdown
// Avoid - Too vague
This doesn't look right.

// Avoid - Not constructive  
This is wrong.

// Avoid - Style only (use automated tools)
Missing space after if.
```

## Testing Requirements

### Test Coverage Standards

| Component | Minimum Coverage | Quality Gate |
|-----------|------------------|--------------|
| **New Features** | 80% | Required |
| **Bug Fixes** | Reproduce bug + fix verification | Required |
| **Refactoring** | Maintain existing coverage | Required |

### Test Types Required

#### Unit Tests
```go
// Always required for new functions
func TestNewFunction(t *testing.T) {
    // Test happy path
    // Test error conditions  
    // Test edge cases
}
```

#### Integration Tests
```go
// Required for external integrations
func TestClusterProvider_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    // Test with real dependencies
}
```

#### E2E Tests
```go
// Required for user-facing workflows
func TestBootstrapWorkflow_E2E(t *testing.T) {
    // Test complete user scenario
}
```

### Test Quality Standards

- **Isolated**: Tests don't depend on each other
- **Deterministic**: Tests produce consistent results
- **Fast**: Unit tests complete quickly (< 1s each)
- **Clear**: Test names describe what is being tested
- **Comprehensive**: Cover happy path, errors, and edge cases

## Documentation Standards

### Code Documentation

```go
// Package-level documentation
// Package cluster provides Kubernetes cluster lifecycle management
// for OpenFrame development environments using K3d.
//
// The package supports creating, deleting, listing, and managing
// local clusters with integrated networking and storage configuration.
package cluster

// Type documentation with usage examples
// ClusterConfig defines the configuration for creating a new cluster.
//
// Example:
//   config := ClusterConfig{
//       Name:   "my-cluster", 
//       CPU:    2,
//       Memory: 4,
//   }
//   service.Create(config)
type ClusterConfig struct {
    // Name is the cluster identifier (must be unique)
    Name string `yaml:"name"`
    
    // CPU allocation in cores (minimum: 1, maximum: 16)
    CPU int `yaml:"cpu"`
    
    // Memory allocation in GB (minimum: 1, maximum: 64)  
    Memory int `yaml:"memory"`
}
```

### User Documentation

All user-facing features require documentation:

- **Command help text**: Clear usage and examples
- **README updates**: For significant features
- **Tutorial updates**: For workflow changes
- **API documentation**: For public interfaces

## Performance Guidelines

### Performance Standards

- **Startup Time**: CLI commands start < 500ms
- **Bootstrap Time**: Complete environment setup < 5 minutes
- **Resource Usage**: Reasonable memory and CPU usage
- **Network Efficiency**: Minimize external requests

### Performance Testing

```go
func BenchmarkClusterCreate(b *testing.B) {
    config := getTestConfig()
    service := NewService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.Create(config)
        service.Delete(config.Name)
    }
}
```

### Optimization Guidelines

- **Concurrent Operations**: Use goroutines for independent tasks
- **Caching**: Cache expensive operations when appropriate
- **Lazy Loading**: Load resources only when needed
- **Progress Feedback**: Show progress for long operations

## Security Guidelines

### Security Requirements

- **Input Validation**: Validate all user inputs
- **Command Injection**: Use exec safely with validated inputs
- **File Permissions**: Set appropriate file permissions
- **Secrets**: Never log or expose secrets

### Security Implementation

```go
// Input validation
func ValidateClusterName(name string) error {
    if !regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name) {
        return ErrInvalidClusterName
    }
    return nil
}

// Safe command execution
func (e *Executor) Run(cmd string, args ...string) error {
    // Validate command against whitelist
    if !isAllowedCommand(cmd) {
        return ErrCommandNotAllowed  
    }
    
    // Execute safely
    return exec.Command(cmd, args...).Run()
}

// File operations
func WriteConfigFile(path string, data []byte) error {
    return ioutil.WriteFile(path, data, 0600) // Secure permissions
}
```

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes (`2.0.0`)
- **MINOR**: New features (`1.1.0`)  
- **PATCH**: Bug fixes (`1.0.1`)

### Release Checklist

- [ ] All tests passing on main branch
- [ ] CHANGELOG.md updated
- [ ] Version bumped in appropriate files
- [ ] Release notes prepared
- [ ] Documentation updated
- [ ] Breaking changes documented

## Getting Help

### Documentation
- **Architecture**: [Architecture Overview](../architecture/overview.md)
- **Development**: [Local Development Guide](../setup/local-development.md)
- **Testing**: [Testing Overview](../testing/overview.md)

### Communication Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: Questions and community support
- **Pull Requests**: Code review and collaboration

### Maintainers
Current maintainers for different areas:

| Area | Maintainer | Expertise |
|------|------------|-----------|
| **Core CLI** | TBD | Command structure, interfaces |
| **Cluster Management** | TBD | K3d integration, Kubernetes |
| **Chart Management** | TBD | Helm, ArgoCD, GitOps |
| **Development Tools** | TBD | Telepresence, Skaffold |
| **Documentation** | TBD | Tutorials, API docs |

## Recognition and Credits

### Contributors

We recognize contributors in:
- `CONTRIBUTORS.md` file
- Release notes
- GitHub contributor insights
- Special recognition for significant contributions

### Types of Contributions

All contributions are valued:
- 🐛 **Bug Reports**: Help us find and fix issues
- 💡 **Feature Requests**: Suggest improvements
- 📝 **Documentation**: Improve user experience
- 🧪 **Testing**: Add test coverage
- 💻 **Code**: Implement features and fixes
- 🎨 **UX/UI**: Improve user interface
- 🔍 **Code Reviews**: Help maintain quality

## Quick Reference

### Essential Commands

| Task | Command |
|------|---------|
| Setup development | `make dev-setup` |
| Run all tests | `make test` |
| Code formatting | `make fmt` |
| Code linting | `make lint` |
| Full quality check | `make quality` |
| Build CLI | `make build` |
| Integration tests | `make test-integration` |

### Workflow Summary

1. **Fork & Clone** repository
2. **Create branch** following naming conventions
3. **Make changes** following code style
4. **Add tests** for new functionality
5. **Update docs** for user-facing changes
6. **Commit** using conventional format
7. **Push & Create PR** with proper description
8. **Address review** feedback
9. **Merge** after approval

---

**Welcome to the OpenFrame CLI community!** 🎉 

Your contributions help make OpenFrame CLI better for everyone. Whether you're fixing a small bug or adding a major feature, we appreciate your effort and expertise.

**Ready to contribute?** Start by exploring the codebase and picking up a ["good first issue"](https://github.com/flamingo-stack/openframe-cli/labels/good%20first%20issue) to get familiar with the project!