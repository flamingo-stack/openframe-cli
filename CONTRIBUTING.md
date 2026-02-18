# Contributing to OpenFrame CLI

Welcome to the OpenFrame CLI project! We appreciate your interest in contributing to this open-source tool that helps manage Kubernetes clusters and development workflows. This guide will help you understand our development process, coding standards, and contribution workflow.

## 🚀 Getting Started

### Before You Contribute

1. **Join Our Community**: Connect with us on [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
2. **Read the Documentation**: Familiarize yourself with the [project documentation](./docs/README.md)
3. **Check Open Issues**: Look for issues labeled `good first issue` or `help wanted`
4. **Discuss Large Changes**: For significant features or changes, discuss with maintainers first

### System Requirements

- **Hardware**: Minimum 24GB RAM, 6 CPU cores, 50GB disk
- **Software**: Go 1.19+, Docker 20.10+, Git 2.30+
- **OS**: Linux, macOS, or Windows with WSL2

## 🔄 Contribution Workflow

### 1. Fork and Clone

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR-USERNAME/openframe-oss-tenant.git
cd openframe-oss-tenant

# Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-oss-tenant.git
git fetch upstream
```

### 2. Set Up Development Environment

```bash
# Build the CLI
go build -o openframe .

# Verify installation
./openframe --version

# Run tests to ensure environment is working
make test
```

### 3. Create a Feature Branch

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description

# Or for documentation
git checkout -b docs/documentation-update
```

### 4. Make Your Changes

- Follow the coding standards outlined below
- Write or update tests for your changes
- Update documentation if needed
- Ensure your changes work across platforms (Linux, macOS, Windows/WSL2)

### 5. Test Your Changes

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

### 6. Commit Your Changes

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

### 7. Push and Create Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create pull request on GitHub
# Fill out the PR template with details about your changes
```

## 📝 Code Style and Standards

### Go Code Style

We follow standard Go conventions with additional guidelines:

#### Formatting

```bash
# Use gofmt and goimports for formatting
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

// Constants use PascalCase or UPPER_CASE
const DefaultTimeout = 30 * time.Second
const MAX_RETRY_ATTEMPTS = 3
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
```

#### Documentation

```go
// Package documentation
// Package cluster provides Kubernetes cluster management functionality.
package cluster

// Function documentation with examples
// CreateCluster creates a new K3D cluster with the specified configuration.
//
// Example:
//   config := &ClusterConfig{Name: "dev-cluster"}
//   result, err := CreateCluster(ctx, config)
//   if err != nil {
//       return fmt.Errorf("cluster creation failed: %w", err)
//   }
func CreateCluster(ctx context.Context, config *ClusterConfig) (*ClusterResult, error) {
    // Implementation
}
```

### Project Structure

```text
internal/
├── bootstrap/          # Bootstrap orchestration
├── cluster/           # Cluster management domain
├── chart/            # Chart installation services
├── dev/              # Development tools
└── shared/           # Shared utilities
    ├── executor/     # Command execution
    ├── ui/          # Common UI components
    └── errors/      # Error handling utilities
```

### Testing Standards

```go
func TestServiceMethod(t *testing.T) {
    tests := []struct {
        name           string
        input          InputType
        expectedResult ExpectedType
        expectedError  string
    }{
        {
            name: "successful operation",
            input: InputType{Field: "value"},
            expectedResult: ExpectedType{},
            expectedError: "",
        },
        {
            name: "error condition",
            input: InputType{Field: "invalid"},
            expectedResult: ExpectedType{},
            expectedError: "validation error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            service := NewService()

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

## 💬 Commit Message Conventions

We follow [Conventional Commits](https://www.conventionalcommits.org/):

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
```

## 🔍 Pull Request Process

### PR Template

```markdown
## Description
Brief description of the changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
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
```

### Review Process

1. **Automated Checks**: CI/CD pipeline runs tests and linting
2. **Maintainer Review**: Core maintainers review code quality and design
3. **Community Review**: Community members may provide feedback
4. **Approval**: At least one maintainer approval required
5. **Merge**: Maintainers handle merging approved PRs

### Branch Naming

```bash
# Features
feature/cluster-auto-scaling
feature/add-helm-chart-validation

# Bug fixes
fix/cluster-creation-timeout
fix/argocd-sync-issue

# Documentation
docs/update-installation-guide
docs/add-troubleshooting-section
```

## 🐛 Issue Reports and Feature Requests

### Creating Issues

1. **Search existing issues**: Avoid duplicates
2. **Use templates**: Fill out the provided issue templates
3. **Be specific**: Provide detailed descriptions and steps to reproduce
4. **Include context**: Operating system, version, environment details

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

## Additional Context
Add any other context about the problem here.
```

## 🏷️ Versioning and Releases

We use [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

## 👥 Community Guidelines

### Code of Conduct

1. **Be respectful**: Treat all community members with respect
2. **Be inclusive**: Welcome newcomers and help them get started
3. **Be collaborative**: Work together toward common goals
4. **Be constructive**: Provide helpful feedback and suggestions

### Communication Channels

- **Slack**: [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) for discussions and support
- **GitHub Issues**: For bug reports and feature requests
- **Pull Requests**: For code contributions and reviews

> **Note**: We don't use GitHub Discussions. All community interaction happens in Slack.

## 🆘 Getting Help

1. **Check Documentation**: Start with the [documentation](./docs/README.md)
2. **Search Issues**: Look for existing similar issues
3. **Ask in Slack**: Join our Slack community for help
4. **Create an Issue**: If you find a bug or need a feature

## 🙏 Recognition

Contributors will be:
- Listed in release notes for their contributions
- Recognized in the project README
- Invited to join contributor Slack channels
- Considered for maintainer roles based on consistent contributions

## 📞 Questions?

If you have questions about contributing:
1. Join our [Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
2. Ask in the `#openframe-cli` channel
3. Tag maintainers for specific questions

## 📚 External Repository

The OpenFrame CLI main codebase is maintained in a separate repository:
- **Repository**: [flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **Development Setup**: Follow the setup instructions in the external repository

Thank you for contributing to OpenFrame CLI! Your contributions help make Kubernetes cluster management more accessible and efficient for the entire community.