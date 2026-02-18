# Contributing Guidelines

Welcome to the OpenFrame CLI project! We appreciate your interest in contributing to this open-source tool that helps manage Kubernetes clusters and development workflows. This guide will help you understand our development process, coding standards, and contribution workflow.

## Getting Started

### Before You Contribute

1. **Join Our Community**: Connect with us on [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
2. **Read the Documentation**: Familiarize yourself with the project architecture and setup
3. **Check Open Issues**: Look for issues labeled `good first issue` or `help wanted`
4. **Discuss Large Changes**: For significant features or changes, discuss with maintainers first

### Development Setup

Ensure you have completed the development environment setup:

1. **[Environment Setup](../setup/environment.md)** - IDE and development tools
2. **[Local Development](../setup/local-development.md)** - Clone and build the project
3. **[Architecture Overview](../architecture/README.md)** - Understand the system design

## Contribution Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR-USERNAME/openframe-oss-tenant.git
cd openframe-oss-tenant

# Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-oss-tenant.git
git fetch upstream
```

### 2. Create a Feature Branch

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description

# Or for documentation
git checkout -b docs/documentation-update
```

### 3. Make Your Changes

- Follow the coding standards outlined below
- Write or update tests for your changes
- Update documentation if needed
- Ensure your changes work across platforms (Linux, macOS, Windows/WSL2)

### 4. Test Your Changes

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration

# Check code formatting and linting
make lint
make format

# Build and test locally
make build
./openframe --version
```

### 5. Commit Your Changes

Follow our commit message conventions:

```bash
# Stage your changes
git add .

# Commit with descriptive message
git commit -m "feat: add cluster auto-scaling support

- Implement horizontal pod autoscaling for services
- Add configuration options for scaling thresholds
- Update documentation with scaling examples

Fixes #123"
```

### 6. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Fill out the PR template with details about your changes
```

## Code Style and Standards

### Go Code Style

We follow standard Go conventions with some additional guidelines:

#### Formatting

```go
// Use gofmt and goimports for formatting
// Run these commands or configure your IDE to run them automatically
gofmt -s -w .
goimports -w .
```

#### Naming Conventions

```go
// Exported types use PascalCase
type ClusterManager struct {
    name string
}

// Exported functions use PascalCase
func CreateCluster(config *Config) error {
    return nil
}

// Private functions use camelCase
func validateClusterName(name string) error {
    return nil
}

// Constants use PascalCase or UPPER_CASE for module-level constants
const DefaultTimeout = 30 * time.Second
const MAX_RETRY_ATTEMPTS = 3

// Interface names should be descriptive
type ClusterProvider interface {
    CreateCluster(config *Config) (*Result, error)
    DeleteCluster(name string) error
}
```

#### Error Handling

```go
// Wrap errors with context
func createCluster(name string) error {
    if err := validateName(name); err != nil {
        return fmt.Errorf("invalid cluster name: %w", err)
    }
    
    if err := executeCommand(name); err != nil {
        return fmt.Errorf("failed to create cluster %s: %w", name, err)
    }
    
    return nil
}

// Use custom error types for domain-specific errors
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

#### Documentation

```go
// Package documentation
// Package cluster provides Kubernetes cluster management functionality.
//
// This package includes tools for creating, managing, and monitoring
// K3D clusters with integrated ArgoCD support.
package cluster

// Function documentation with examples
// CreateCluster creates a new K3D cluster with the specified configuration.
//
// The function performs prerequisite checks, creates the cluster,
// and configures it for OpenFrame development.
//
// Example:
//   config := &ClusterConfig{
//       Name: "dev-cluster",
//       Registry: "local",
//   }
//   result, err := CreateCluster(ctx, config)
//   if err != nil {
//       return fmt.Errorf("cluster creation failed: %w", err)
//   }
func CreateCluster(ctx context.Context, config *ClusterConfig) (*ClusterResult, error) {
    // Implementation
}

// Struct documentation
// ClusterConfig represents the configuration for cluster creation.
type ClusterConfig struct {
    // Name is the cluster identifier (required)
    Name string `json:"name"`
    
    // Registry specifies the container registry type
    // Valid values: "local", "ghcr", "docker"
    Registry string `json:"registry"`
    
    // Ports defines port mappings for the cluster
    Ports []PortMapping `json:"ports,omitempty"`
}
```

### Project Structure

Follow these guidelines for organizing code:

```text
internal/
├── bootstrap/          # Bootstrap orchestration
│   ├── service.go     # Main service implementation
│   └── service_test.go # Unit tests
├── cluster/           # Cluster management domain
│   ├── models/        # Data structures
│   ├── providers/     # External integrations
│   ├── service.go     # Domain service
│   └── ui/           # User interface components
└── shared/           # Shared utilities
    ├── executor/     # Command execution
    ├── ui/          # Common UI components
    └── errors/      # Error handling utilities
```

#### File Organization Rules

1. **Service Pattern**: Each domain has a main service file
2. **Interface Separation**: Define interfaces near their usage
3. **Provider Pattern**: External integrations in `providers/` subdirectory
4. **UI Components**: User interface logic in `ui/` subdirectory
5. **Test Colocation**: Unit tests alongside source files

### Testing Standards

#### Test Structure

```go
func TestServiceMethod(t *testing.T) {
    tests := []struct {
        name           string
        input          InputType
        mockSetup      func(*mocks.Dependency)
        expectedResult ExpectedType
        expectedError  string
    }{
        {
            name: "successful operation",
            input: InputType{
                Field: "value",
            },
            mockSetup: func(mock *mocks.Dependency) {
                mock.On("Method", mock.Anything).Return(nil)
            },
            expectedResult: ExpectedType{},
            expectedError:  "",
        },
        {
            name: "error condition",
            input: InputType{
                Field: "invalid",
            },
            mockSetup: func(mock *mocks.Dependency) {
                mock.On("Method", mock.Anything).Return(errors.New("mock error"))
            },
            expectedResult: ExpectedType{},
            expectedError:  "mock error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockDep := &mocks.Dependency{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockDep)
            }
            
            service := NewService(mockDep)

            // Act
            result, err := service.Method(tt.input)

            // Assert
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expectedResult, result)
            }
            
            mockDep.AssertExpectations(t)
        })
    }
}
```

#### Test Requirements

1. **Test Coverage**: Maintain >80% code coverage
2. **Table-Driven Tests**: Use for multiple test cases
3. **Mocking**: Mock external dependencies
4. **Integration Tests**: Test with real dependencies when appropriate
5. **Error Testing**: Test both success and failure paths

## Commit Message Conventions

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

### Format

```text
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat: add cluster auto-scaling` |
| `fix` | Bug fix | `fix: resolve cluster creation timeout` |
| `docs` | Documentation changes | `docs: update installation guide` |
| `style` | Code style changes | `style: format cluster service` |
| `refactor` | Code refactoring | `refactor: extract chart validation` |
| `test` | Test additions or changes | `test: add cluster lifecycle tests` |
| `chore` | Build or tool changes | `chore: update dependencies` |

### Examples

```bash
# Feature addition
git commit -m "feat(cluster): add support for custom node labels

- Allow users to specify custom labels for cluster nodes
- Update configuration validation to accept label format
- Add tests for label validation and application

Closes #456"

# Bug fix
git commit -m "fix(chart): resolve ArgoCD installation timeout

The ArgoCD installation was timing out due to insufficient
wait time for the operator to become ready. Increased the
timeout and improved status checking logic.

Fixes #789"

# Documentation update
git commit -m "docs: add troubleshooting section to setup guide

- Document common installation issues
- Add solutions for WSL2 integration problems
- Include debugging commands and techniques"
```

### Commit Guidelines

1. **Keep commits atomic**: One logical change per commit
2. **Write clear descriptions**: Explain what and why, not just what
3. **Reference issues**: Use `Closes #123` or `Fixes #123`
4. **Limit line length**: Keep subject line under 72 characters
5. **Use imperative mood**: "Add feature" not "Added feature"

## Pull Request Process

### PR Template

When you create a pull request, use this template:

```markdown
## Description
Brief description of the changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring

## Testing
Describe the tests you ran and how to reproduce them:
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing performed

## Checklist
- [ ] My code follows the project's coding standards
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes

## Screenshots (if applicable)
Add screenshots to help explain your changes.

## Additional Notes
Any additional information or context about the changes.
```

### PR Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Maintainer Review**: Core maintainers review code quality and design
3. **Community Review**: Community members may provide feedback
4. **Testing**: Reviewers may test changes locally
5. **Approval**: At least one maintainer approval required
6. **Merge**: Maintainers handle merging approved PRs

### PR Guidelines

1. **Keep PRs focused**: One feature or fix per PR
2. **Write good descriptions**: Explain the problem and solution
3. **Update documentation**: Include relevant documentation updates
4. **Add tests**: Ensure new code is tested
5. **Respond to feedback**: Address reviewer comments promptly
6. **Keep up to date**: Rebase on main branch if needed

## Branch Naming Conventions

Use descriptive branch names that indicate the type of work:

```bash
# Features
feature/cluster-auto-scaling
feature/add-helm-chart-validation
feature/windows-wsl2-support

# Bug fixes
fix/cluster-creation-timeout
fix/argocd-sync-issue
fix/memory-leak-in-provider

# Documentation
docs/update-installation-guide
docs/add-troubleshooting-section
docs/improve-api-documentation

# Refactoring
refactor/extract-validation-logic
refactor/improve-error-handling
refactor/simplify-config-management
```

## Code Review Guidelines

### For Contributors

When requesting a review:

1. **Self-review first**: Review your own changes before submitting
2. **Test thoroughly**: Ensure all tests pass and functionality works
3. **Document changes**: Update relevant documentation
4. **Small PRs**: Keep changes focused and reviewable
5. **Respond promptly**: Address feedback in a timely manner

### For Reviewers

When reviewing code:

1. **Be constructive**: Provide helpful, actionable feedback
2. **Focus on substance**: Prioritize functionality, security, and maintainability
3. **Ask questions**: If something isn't clear, ask for clarification
4. **Test locally**: For significant changes, test the functionality
5. **Approve explicitly**: Use GitHub's approval feature when satisfied

### Review Checklist

#### Functionality
- [ ] Does the code solve the stated problem?
- [ ] Are edge cases handled appropriately?
- [ ] Is error handling comprehensive?
- [ ] Are there any potential security issues?

#### Code Quality
- [ ] Is the code readable and well-organized?
- [ ] Are functions and variables named clearly?
- [ ] Is there appropriate commenting?
- [ ] Does the code follow project conventions?

#### Testing
- [ ] Are there adequate unit tests?
- [ ] Do integration tests cover the main workflows?
- [ ] Are error paths tested?
- [ ] Is test coverage maintained or improved?

#### Documentation
- [ ] Is public API documented?
- [ ] Are configuration changes documented?
- [ ] Is the README updated if needed?
- [ ] Are examples provided for new features?

## Issue and Bug Reports

### Creating Issues

When creating issues, please:

1. **Search existing issues**: Avoid duplicates
2. **Use templates**: Fill out the provided issue templates
3. **Be specific**: Provide detailed descriptions and steps to reproduce
4. **Include context**: Operating system, version, environment details
5. **Label appropriately**: Use relevant labels (bug, feature, documentation)

### Bug Report Template

```markdown
## Bug Description
A clear and concise description of what the bug is.

## Steps to Reproduce
1. Run command '...'
2. Configure option '...'
3. See error

## Expected Behavior
A clear description of what you expected to happen.

## Actual Behavior
What actually happened instead.

## Environment
- OS: [e.g., Ubuntu 20.04, macOS 12, Windows 11]
- OpenFrame CLI Version: [e.g., v1.2.3]
- Go Version: [e.g., 1.19.5]
- Docker Version: [e.g., 20.10.21]
- K3D Version: [e.g., 5.4.6]

## Additional Context
Add any other context about the problem here.
```

### Feature Request Template

```markdown
## Feature Description
A clear and concise description of what you want to happen.

## Motivation
Why is this feature needed? What problem does it solve?

## Detailed Design
If you have ideas about implementation, describe them here.

## Alternatives Considered
Describe alternative solutions or features you've considered.

## Additional Context
Add any other context or screenshots about the feature request here.
```

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Checklist

For maintainers preparing releases:

1. **Update Version**: Update version numbers in relevant files
2. **Update Changelog**: Document all changes since last release
3. **Test Thoroughly**: Run full test suite including E2E tests
4. **Build Artifacts**: Create binaries for all supported platforms
5. **Tag Release**: Create git tag with version number
6. **Publish Release**: Create GitHub release with artifacts
7. **Update Documentation**: Ensure docs reflect new version

## Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment:

1. **Be respectful**: Treat all community members with respect
2. **Be inclusive**: Welcome newcomers and help them get started
3. **Be collaborative**: Work together toward common goals
4. **Be constructive**: Provide helpful feedback and suggestions
5. **Be patient**: Remember that everyone is learning

### Communication Channels

- **Slack**: [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) for discussions and support
- **GitHub Issues**: For bug reports and feature requests
- **Pull Requests**: For code contributions and reviews

> **Note**: We don't use GitHub Discussions. All community interaction happens in Slack.

### Getting Help

If you need help:

1. **Check Documentation**: Start with the docs in this repository
2. **Search Issues**: Look for existing similar issues
3. **Ask in Slack**: Join our Slack community for help
4. **Create an Issue**: If you find a bug or need a feature

## Recognition

We appreciate all contributors! Contributors will be:

- Listed in release notes for their contributions
- Recognized in the project README
- Invited to join the contributor Slack channels
- Considered for maintainer roles based on consistent contributions

## Questions?

If you have questions about contributing:

1. Join our [Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
2. Ask in the `#openframe-cli` channel
3. Tag maintainers for specific questions

Thank you for contributing to OpenFrame CLI! Your contributions help make Kubernetes cluster management more accessible and efficient for the entire community.