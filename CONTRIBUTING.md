# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! We welcome contributions from developers of all experience levels. This guide will help you get started and ensure your contributions align with our project goals and standards.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Testing Requirements](#testing-requirements)
- [Code Style](#code-style)
- [Documentation](#documentation)
- [Release Process](#release-process)
- [Community](#community)

## Code of Conduct

This project follows the [Flamingo AI Code of Conduct](https://www.flamingo.run/code-of-conduct). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

### Our Pledge

- Be respectful and inclusive of all contributors
- Focus on constructive feedback and collaboration
- Help create a welcoming environment for newcomers
- Prioritize the project's success and user experience

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.23 or later** installed
- **Docker** running on your system
- **Git** configured with your GitHub account
- **Make** available for build automation
- Basic familiarity with Kubernetes concepts

### First-Time Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openframe-cli.git
   cd openframe-cli
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   ```
4. **Install dependencies** and build:
   ```bash
   make deps
   make build
   ```
5. **Verify your setup**:
   ```bash
   ./openframe --help
   make test
   ```

### Project Structure

```
openframe-cli/
├── cmd/                    # Command implementations
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/          # Cluster management commands
│   ├── chart/            # Chart management commands
│   └── dev/              # Development tools commands
├── internal/             # Internal packages
│   ├── cluster/          # Cluster services and logic
│   ├── chart/            # Chart services and logic
│   ├── dev/              # Development tool integrations
│   └── shared/           # Shared utilities and components
├── docs/                 # Documentation
├── tests/                # Test files
└── scripts/              # Build and automation scripts
```

## Development Setup

### IDE Configuration

**Visual Studio Code**:
```json
{
  "go.useLanguageServer": true,
  "go.formatTool": "gofumpt",
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true
}
```

**GoLand**: Enable Go modules and configure code style to match our standards.

### Required Tools

Install development tools:

```bash
# Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest

# Kubernetes tools (for testing)
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

### Build Commands

```bash
# Development commands
make build          # Build the CLI binary
make test           # Run all tests
make lint           # Run code linting
make fmt            # Format code
make clean          # Clean build artifacts

# Testing commands
make test-unit      # Run unit tests only
make test-integration # Run integration tests
make test-e2e       # Run end-to-end tests
make test-cover     # Run tests with coverage

# Release commands
make release-local  # Build release binaries locally
```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

| Type | Examples | Guidelines |
|------|----------|------------|
| **🐛 Bug Fixes** | Command errors, logic issues | Include reproduction steps and tests |
| **✨ Features** | New commands, integrations | Discuss in an issue first |
| **📚 Documentation** | Guides, examples, API docs | Clear, comprehensive, tested |
| **🧪 Testing** | Test coverage, test utilities | Follow testing patterns |
| **🔧 Refactoring** | Code organization, performance | Maintain backward compatibility |

### Before You Start

1. **Check existing issues** to avoid duplicate work
2. **Create or comment on an issue** to discuss your planned changes
3. **Get maintainer approval** for significant features or architectural changes
4. **Read relevant documentation** in the [docs/development](./docs/development/) section

### Development Workflow

#### 1. Create a Feature Branch

```bash
# Sync with upstream
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name
```

#### 2. Make Your Changes

- Write clear, focused commits
- Include tests for new functionality
- Update documentation as needed
- Follow our code style guidelines

#### 3. Test Your Changes

```bash
# Run the full test suite
make test

# Test specific components
go test ./cmd/cluster/...
go test ./internal/shared/...

# Manual testing
./openframe cluster create test-cluster
./openframe cluster delete test-cluster
```

#### 4. Commit Your Changes

Use conventional commit format:

```bash
git commit -m "feat(cluster): add support for custom node labels"
git commit -m "fix(bootstrap): handle missing Docker daemon error"
git commit -m "docs(README): update installation instructions"
```

#### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Pull Request Process

### PR Requirements

Before submitting a pull request, ensure:

- [ ] **Tests pass**: `make test` runs successfully
- [ ] **Code is formatted**: `make fmt` applied
- [ ] **Linting passes**: `make lint` shows no errors
- [ ] **Documentation updated**: If you changed user-facing behavior
- [ ] **Conventional commits**: Use proper commit message format
- [ ] **Issue linked**: Reference the relevant issue number

### PR Template

Use this template for your pull request description:

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests you ran and their results.

## Checklist
- [ ] Tests pass locally
- [ ] Code follows project style guidelines
- [ ] Documentation updated (if applicable)
- [ ] Commit messages follow conventional format

## Related Issues
Fixes #123
```

### Review Process

1. **Automated checks** must pass (tests, linting, build)
2. **Code review** by at least one maintainer
3. **Manual testing** for significant changes
4. **Documentation review** for user-facing changes
5. **Final approval** and merge by maintainer

### Addressing Review Feedback

- Respond to all review comments
- Make requested changes in new commits
- Use `git commit --fixup` for small fixes
- Ping reviewers when ready for re-review

## Issue Guidelines

### Bug Reports

Use the bug report template and include:

- **OpenFrame CLI version**: `openframe --version`
- **Operating system**: macOS, Linux, Windows version
- **Docker version**: `docker --version`
- **Go version**: `go version` (if building from source)
- **Reproduction steps**: Detailed steps to reproduce the issue
- **Expected behavior**: What should have happened
- **Actual behavior**: What actually happened
- **Logs/output**: Relevant command output or error messages

### Feature Requests

Use the feature request template and include:

- **Problem statement**: What problem does this solve?
- **Proposed solution**: How should this be implemented?
- **Alternatives considered**: Other approaches you've considered
- **Use cases**: Real-world scenarios where this would be helpful
- **Breaking changes**: Any potential backward compatibility issues

### Issue Labels

| Label | Purpose |
|-------|---------|
| `bug` | Something isn't working |
| `enhancement` | New feature or request |
| `documentation` | Documentation improvements |
| `good-first-issue` | Good for newcomers |
| `help-wanted` | Extra attention needed |
| `priority/high` | High priority issue |
| `status/in-progress` | Currently being worked on |

## Testing Requirements

### Test Structure

```
tests/
├── unit/           # Unit tests for individual functions
├── integration/    # Integration tests for component interactions
├── e2e/           # End-to-end tests for complete workflows
└── fixtures/      # Test data and configurations
```

### Writing Tests

#### Unit Tests

```go
func TestClusterCreate(t *testing.T) {
    tests := []struct {
        name     string
        input    ClusterConfig
        expected error
    }{
        {
            name:     "valid config",
            input:    validConfig,
            expected: nil,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateCluster(tt.input)
            assert.Equal(t, tt.expected, err)
        })
    }
}
```

#### Integration Tests

```go
func TestBootstrapWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Test complete bootstrap workflow
    // including cluster creation and chart installation
}
```

#### E2E Tests

```go
func TestFullWorkflow(t *testing.T) {
    if os.Getenv("E2E_TESTS") == "" {
        t.Skip("E2E tests disabled")
    }
    
    // Test complete user workflow from CLI
}
```

### Test Commands

```bash
# Run specific test types
make test-unit              # Fast unit tests only
make test-integration       # Integration tests (requires Docker)
make test-e2e E2E_TESTS=1   # End-to-end tests (slow)

# Coverage reporting
make test-cover             # Generate coverage report
make test-cover-html        # Generate HTML coverage report
```

### Test Requirements

- **Unit tests** for all public functions
- **Integration tests** for external tool interactions
- **E2E tests** for critical user workflows
- **Table-driven tests** for multiple input scenarios
- **Error case coverage** for all error paths

## Code Style

### Go Standards

Follow standard Go conventions and these project-specific guidelines:

#### Formatting

```bash
# Use gofumpt for consistent formatting
make fmt
```

#### Naming Conventions

```go
// Packages: lowercase, single word
package cluster

// Functions: PascalCase for exported, camelCase for unexported
func CreateCluster() {}
func validateConfig() {}

// Variables: camelCase, descriptive names
var clusterConfig ClusterConfig
var isValidConfig bool

// Constants: PascalCase for exported, camelCase for unexported
const DefaultNodeCount = 3
const maxRetries = 5
```

#### Error Handling

```go
// Use pkg/errors for error wrapping
import "github.com/pkg/errors"

func CreateCluster(config Config) error {
    if err := validateConfig(config); err != nil {
        return errors.Wrap(err, "invalid cluster configuration")
    }
    
    if err := createNodes(config); err != nil {
        return errors.Wrap(err, "failed to create cluster nodes")
    }
    
    return nil
}
```

#### Comments and Documentation

```go
// ClusterService manages Kubernetes cluster lifecycle operations.
// It provides methods for creating, deleting, and monitoring K3d clusters
// with integrated validation and error handling.
type ClusterService struct {
    k3dClient K3dClient
    docker    DockerClient
}

// CreateCluster creates a new K3d cluster with the specified configuration.
// It validates the configuration, ensures prerequisites are met, and
// creates the cluster nodes with proper networking setup.
func (s *ClusterService) CreateCluster(config ClusterConfig) error {
    // Implementation...
}
```

### CLI Design Principles

#### Command Structure

```go
// Use Cobra command structure
var createCmd = &cobra.Command{
    Use:   "create [name]",
    Short: "Create a new K3d cluster",
    Long: `Create a new K3d cluster with interactive configuration.
    
This command guides you through cluster creation with smart defaults
and validation. You can also use flags for non-interactive mode.`,
    Example: `  openframe cluster create
  openframe cluster create my-cluster --nodes 3
  openframe cluster create --non-interactive`,
    Args: cobra.MaximumNArgs(1),
    RunE: runCreateCluster,
}
```

#### User Experience

- **Clear output**: Use colors and formatting for readability
- **Progress indication**: Show progress for long operations
- **Helpful errors**: Provide actionable error messages
- **Interactive mode**: Guide users through complex operations
- **Non-interactive mode**: Support CI/CD automation

## Documentation

### Documentation Types

| Type | Location | Purpose |
|------|----------|---------|
| **User Guides** | `docs/getting-started/` | Help users get started |
| **Developer Docs** | `docs/development/` | Guide contributors |
| **API Reference** | `docs/reference/` | Technical specifications |
| **Architecture** | `docs/architecture/` | System design |

### Writing Guidelines

- **Clear and concise**: Write for your audience
- **Tested examples**: All code examples must work
- **Up-to-date**: Keep documentation synchronized with code
- **Accessible**: Use inclusive language and clear structure
- **Visual aids**: Include diagrams and screenshots where helpful

### Documentation Commands

```bash
# Generate documentation
make docs-generate      # Generate API docs
make docs-serve         # Serve docs locally
make docs-validate      # Validate all links and examples
```

## Release Process

### Version Management

We use semantic versioning (SemVer):

- **Major** (1.0.0): Breaking changes
- **Minor** (0.1.0): New features, backward compatible
- **Patch** (0.0.1): Bug fixes, backward compatible

### Release Workflow

1. **Create release branch**: `release/v1.2.0`
2. **Update version**: Update version in code and docs
3. **Update changelog**: Document all changes
4. **Test release candidate**: Run full test suite
5. **Create release PR**: Merge to main with approval
6. **Tag release**: Create Git tag with release notes
7. **Automated build**: GitHub Actions builds and publishes binaries

### Release Notes

Include in release notes:

- **New features** with examples
- **Bug fixes** with issue references  
- **Breaking changes** with migration guide
- **Deprecations** with timeline
- **Contributors** acknowledgment

## Community

### Communication Channels

- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General questions and ideas
- **Pull Requests**: Code review and collaboration
- **Documentation**: Comprehensive guides and references

### Getting Help

- **New contributors**: Look for `good-first-issue` labels
- **Questions**: Use GitHub Discussions
- **Bugs**: Create detailed issue reports
- **Features**: Discuss in issues before implementing

### Recognition

We value all contributions and maintain a [contributors file](./CONTRIBUTORS.md) recognizing:

- Code contributors
- Documentation writers  
- Issue reporters and triagers
- Community supporters

## Helpful Resources

### External Documentation

- **Go Documentation**: https://pkg.go.dev/
- **Cobra CLI Framework**: https://cobra.dev/
- **Kubernetes Client**: https://pkg.go.dev/k8s.io/client-go
- **Docker SDK**: https://docs.docker.com/engine/api/sdk/
- **K3d Documentation**: https://k3d.io/

### Project-Specific Resources

- **Architecture Overview**: [docs/development/architecture/overview.md](./docs/development/architecture/overview.md)
- **Development Setup**: [docs/development/setup/environment.md](./docs/development/setup/environment.md)
- **Testing Guide**: [docs/development/testing/overview.md](./docs/development/testing/overview.md)

---

Thank you for contributing to OpenFrame CLI! Your efforts help make Kubernetes development more accessible and enjoyable for developers everywhere. 🚀

If you have questions about contributing, please don't hesitate to open an issue or start a discussion. We're here to help and excited to work with you!