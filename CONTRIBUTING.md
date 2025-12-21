# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! We welcome contributions from the community and are excited to collaborate with you.

## 📋 Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Environment](#development-environment)
4. [Development Workflow](#development-workflow)
5. [Contribution Guidelines](#contribution-guidelines)
6. [Code Style and Standards](#code-style-and-standards)
7. [Testing](#testing)
8. [Documentation](#documentation)
9. [Pull Request Process](#pull-request-process)
10. [Community and Support](#community-and-support)

## 🤝 Code of Conduct

This project adheres to the Flamingo Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

### Our Standards

- **Be Respectful**: Treat everyone with respect and kindness
- **Be Collaborative**: Work together constructively and help each other
- **Be Inclusive**: Welcome people from all backgrounds and experience levels
- **Be Professional**: Keep discussions focused and productive

## 🚀 Getting Started

### Prerequisites

Before contributing, ensure you have the following installed:

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.21+ | Primary development language |
| **Docker** | 20.10+ | Container runtime for testing |
| **kubectl** | 1.20+ | Kubernetes CLI for testing |
| **K3d** | 5.0+ | Local Kubernetes for testing |
| **Make** | 3.81+ | Build automation |
| **Git** | 2.30+ | Version control |

### Quick Setup

```bash
# Fork the repository on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Set up the original repository as upstream
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git

# Install dependencies
go mod download

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Verify setup
make check
```

## 🛠 Development Environment

### Repository Structure

```
openframe-cli/
├── cmd/                    # CLI commands and subcommands
│   ├── bootstrap/         # Bootstrap command implementation
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart and ArgoCD commands
│   └── dev/               # Development workflow commands
├── internal/              # Internal packages (not for external import)
│   ├── bootstrap/         # Bootstrap service logic
│   ├── cluster/           # Cluster management services
│   ├── chart/             # Helm and ArgoCD integration
│   ├── dev/               # Development tool integration
│   └── shared/            # Shared utilities and components
├── docs/                  # Documentation
├── scripts/               # Build and automation scripts
└── main.go               # Application entry point
```

### Build and Test Commands

```bash
# Build the project
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Run linter
make lint

# Format code
make fmt

# Run all checks (test + lint + fmt)
make check

# Clean build artifacts
make clean
```

### Environment Configuration

Create a `.env` file for local development:

```bash
# Copy example configuration
cp .env.example .env

# Edit as needed
vim .env
```

Example `.env` settings:
```bash
LOG_LEVEL=debug
DEV_MODE=true
SKIP_PREREQUISITES=false
TEST_CLUSTER_PREFIX=contrib-test-
CLEANUP_ON_EXIT=true
```

## 🔄 Development Workflow

### Branch Strategy

We use a **feature branch workflow**:

```bash
# Start from main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description

# Or for documentation
git checkout -b docs/update-readme
```

### Making Changes

1. **Write code** following our style guidelines
2. **Add tests** for new functionality
3. **Update documentation** as needed
4. **Test locally** to ensure everything works
5. **Commit changes** with descriptive messages

### Commit Message Guidelines

We follow [Conventional Commits](https://conventionalcommits.org/):

```bash
# Format: type(scope): description
git commit -m "feat(cluster): add cluster backup command"
git commit -m "fix(chart): resolve ArgoCD installation timeout"
git commit -m "docs(readme): update installation instructions"
git commit -m "test(cluster): add integration tests for creation"
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or fixing tests
- `refactor`: Code refactoring
- `style`: Code style changes
- `chore`: Build process or auxiliary tool changes

### Testing Your Changes

```bash
# Run unit tests
make test

# Run integration tests (requires Docker)
make test-integration

# Test manually with a real cluster
openframe bootstrap test-contrib --verbose

# Cleanup after testing
openframe cluster delete test-contrib
openframe cluster cleanup
```

## 📝 Contribution Guidelines

### Types of Contributions

We welcome various types of contributions:

#### 🐛 Bug Reports

When reporting bugs, please include:
- **Description**: Clear description of the issue
- **Steps to reproduce**: Exact commands and steps
- **Expected behavior**: What should have happened
- **Actual behavior**: What actually happened
- **Environment**: OS, Docker version, OpenFrame CLI version
- **Logs**: Relevant log output (use `--verbose` flag)

#### ✨ Feature Requests

For feature requests, please provide:
- **Use case**: Why is this feature needed?
- **Description**: Detailed explanation of the feature
- **Examples**: How would it be used?
- **Alternatives**: What alternatives have you considered?

#### 🔧 Code Contributions

Areas where we especially welcome contributions:
- Bug fixes and stability improvements
- New cluster management features
- Developer workflow enhancements
- Documentation improvements
- Test coverage improvements
- Performance optimizations

#### 📚 Documentation

Help improve our documentation by:
- Fixing typos and grammar
- Adding examples and use cases
- Improving clarity and organization
- Adding translations
- Writing tutorials and guides

### Contribution Process

1. **Check existing issues** to avoid duplicates
2. **Open an issue** to discuss large changes before implementing
3. **Fork the repository** and create a feature branch
4. **Implement your changes** following our guidelines
5. **Add tests** and ensure all tests pass
6. **Update documentation** as needed
7. **Submit a pull request** with clear description

## 🎨 Code Style and Standards

### Go Style Guidelines

We follow standard Go conventions plus additional rules:

#### Code Formatting

```bash
# Format code automatically
make fmt

# Check formatting
make fmt-check
```

#### Naming Conventions

```go
// ✅ Good - descriptive names
func CreateClusterWithValidation(config ClusterConfig) error
func ValidateClusterConfiguration(config ClusterConfig) error

// ❌ Bad - abbreviated names
func CreateClstr(cfg ClusterConfig) error
func ValidateClstrCfg(cfg ClusterConfig) error

// ✅ Good - interface naming
type ClusterService interface {
    CreateCluster(ctx context.Context, config ClusterConfig) error
}

// ✅ Good - constants
const (
    DefaultClusterName     = "openframe-dev"
    DefaultNodeCount       = 3
    MaxClusterNameLength   = 63
)
```

#### Error Handling

```go
// ✅ Good - wrap errors with context
func CreateCluster(name string) error {
    if err := validateClusterName(name); err != nil {
        return fmt.Errorf("invalid cluster name %q: %w", name, err)
    }
    
    if err := k3dClient.Create(name); err != nil {
        return fmt.Errorf("failed to create cluster %q: %w", name, err)
    }
    
    return nil
}

// ✅ Good - sentinel errors
var (
    ErrClusterNotFound = errors.New("cluster not found")
    ErrClusterExists   = errors.New("cluster already exists")
)
```

#### Package Organization

```go
package cluster

import (
    // Standard library first
    "context"
    "fmt"
    "strings"
    
    // Third-party libraries
    "github.com/spf13/cobra"
    "k8s.io/client-go/kubernetes"
    
    // Internal packages last
    "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
    "github.com/flamingo-stack/openframe-cli/internal/shared/errors"
)
```

#### Documentation

```go
// ClusterService provides cluster lifecycle management operations.
// It handles K3d cluster creation, deletion, and status monitoring
// with integrated prerequisite checking and error recovery.
type ClusterService interface {
    // CreateCluster creates a new K3d cluster with the given configuration.
    // It validates the configuration, checks prerequisites, and sets up
    // the cluster with appropriate networking and storage settings.
    //
    // Returns an error if the cluster already exists, prerequisites are
    // not met, or the cluster creation fails.
    CreateCluster(ctx context.Context, config ClusterConfig) error
}
```

### Command Guidelines

#### Help Text

```go
var createCmd = &cobra.Command{
    Use:   "create [NAME]",
    Short: "Create a new Kubernetes cluster",
    Long: `Create a new Kubernetes cluster with quick defaults or interactive configuration.

By default, shows a selection menu where you can choose:
1. Quick start with defaults (press Enter) - creates cluster with default settings
2. Interactive configuration wizard - step-by-step cluster customization

Examples:
  openframe cluster create                    # Show creation mode selection
  openframe cluster create my-cluster        # Show selection with custom name
  openframe cluster create --skip-wizard     # Direct creation with defaults`,
    Args: cobra.MaximumNArgs(1),
    RunE: runCreateCluster,
}
```

#### Flag Definition

```go
// Add flags with clear descriptions and sensible defaults
cmd.Flags().StringP("deployment-mode", "d", "oss-tenant", 
    "Deployment mode (oss-tenant, saas-tenant, saas-shared)")
cmd.Flags().IntP("nodes", "n", 3, 
    "Number of worker nodes to create")
cmd.Flags().BoolP("skip-wizard", "s", false, 
    "Skip interactive wizard and use defaults")
cmd.Flags().BoolP("verbose", "v", false, 
    "Enable verbose output for debugging")
```

## 🧪 Testing

### Test Structure

We use table-driven tests for comprehensive coverage:

```go
func TestClusterService_CreateCluster(t *testing.T) {
    tests := []struct {
        name           string
        config         ClusterConfig
        existingCluster bool
        wantErr        bool
        expectedError  error
    }{
        {
            name: "valid cluster creation",
            config: ClusterConfig{
                Name:      "test-cluster",
                NodeCount: 3,
            },
            existingCluster: false,
            wantErr:        false,
        },
        {
            name: "cluster already exists",
            config: ClusterConfig{
                Name: "existing-cluster",
            },
            existingCluster: true,
            wantErr:        true,
            expectedError:  ErrClusterExists,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := setupTestService(t)
            
            if tt.existingCluster {
                setupExistingCluster(t, tt.config.Name)
            }
            
            err := service.CreateCluster(context.Background(), tt.config)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.expectedError != nil {
                    assert.True(t, errors.Is(err, tt.expectedError))
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Running Tests

```bash
# Unit tests only
go test ./...

# With coverage
go test -cover ./...

# Integration tests (requires Docker)
make test-integration

# Specific package
go test ./cmd/cluster/...

# Specific test
go test ./cmd/cluster/ -run TestCreateCluster

# Verbose output
go test -v ./...
```

### Test Categories

1. **Unit Tests**: Fast, isolated, no external dependencies
2. **Integration Tests**: Test with real K3d clusters and Docker
3. **End-to-End Tests**: Full workflow testing with all components

## 📖 Documentation

### Documentation Types

1. **README.md**: Project overview and quick start
2. **Inline Documentation**: Hidden `.*.md` files in packages
3. **API Documentation**: Generated from Go comments
4. **Tutorials**: Step-by-step guides in `docs/tutorials/`
5. **Architecture Docs**: Technical overviews in `docs/dev/`

### Documentation Guidelines

#### Markdown Style

```markdown
# Main Title (H1) - Only one per document

## Section Headers (H2)

### Subsections (H3)

- Use bullet points for lists
- **Bold** for emphasis
- `Code snippets` in backticks

```bash
# Code blocks with syntax highlighting
openframe cluster create example
```

| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data     | More     | Info     |
```

#### Code Examples

Include complete, runnable examples:

```bash
# ✅ Good - complete example with context
# Create a development cluster with custom configuration
openframe cluster create dev-env \
  --nodes 3 \
  --deployment-mode oss-tenant \
  --verbose

# Check that it was created successfully
openframe cluster list
```

```bash
# ❌ Bad - incomplete or unclear
openframe create dev
```

## 🔄 Pull Request Process

### Before Submitting

Complete this checklist before creating a pull request:

- [ ] **Code compiles** without warnings
- [ ] **All tests pass**: `make test`
- [ ] **Linter passes**: `make lint`
- [ ] **Code is formatted**: `make fmt`
- [ ] **Documentation updated** for any user-facing changes
- [ ] **Manual testing completed** on relevant platforms
- [ ] **No breaking changes** (or properly documented if necessary)

### Pull Request Template

When creating a pull request, include:

```markdown
## Description
Brief description of changes and their purpose.

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## How Has This Been Tested?
Describe the tests that you ran to verify your changes.

## Related Issues
Closes #123
Fixes #456

## Screenshots (if applicable)
Add screenshots for UI changes.

## Additional Notes
Any additional information that reviewers should know.
```

### Review Process

1. **Automated Checks**: CI runs tests and linting
2. **Code Review**: Maintainers and community review code
3. **Testing**: Manual testing on different platforms
4. **Documentation**: Ensure docs are updated
5. **Approval**: At least one maintainer approval required
6. **Merge**: Squash and merge to main branch

### Review Criteria

Reviewers will check for:

- **Functionality**: Does it work as intended?
- **Code Quality**: Is it well-written and maintainable?
- **Testing**: Are tests adequate and passing?
- **Documentation**: Is it properly documented?
- **Performance**: Any performance implications?
- **Security**: Are there security considerations?
- **Backwards Compatibility**: Breaking changes minimized?

## 🐛 Debugging and Troubleshooting

### Local Development Debugging

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o openframe-debug .

# Run with verbose output
./openframe-debug cluster create debug-test --verbose

# Use delve debugger
dlv exec openframe-debug -- cluster create debug-test
```

### Common Development Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| `package not found` | Incorrect import path | Check `go.mod` and fix imports |
| `tests timeout` | Long operations without context | Add `context.Context` with timeouts |
| `lint errors` | Code style violations | Run `make fmt lint` to fix |
| `Docker permission denied` | User not in docker group | `sudo usermod -aG docker $USER` |
| `K3d cluster creation fails` | Port conflicts | Change default ports or stop conflicting services |

### Debug Logging

Add debug logging to your code:

```go
func (s *clusterService) CreateCluster(config ClusterConfig) error {
    s.logger.Debug("Creating cluster",
        "name", config.Name,
        "nodeCount", config.NodeCount,
        "k8sVersion", config.K8sVersion)
    
    // Implementation...
    
    s.logger.Info("Cluster created successfully", "name", config.Name)
    return nil
}
```

## 🌟 Recognition

### Contributors

We recognize contributors through:

- **GitHub Contributors Graph**: Automatic recognition
- **Release Notes**: Major contributors mentioned
- **Documentation**: Contributors credited in relevant docs
- **Community Shout-outs**: Recognition in community channels

### Contribution Levels

- **First-time Contributors**: Special welcome and guidance
- **Regular Contributors**: Increased review privileges
- **Core Contributors**: Maintainer status consideration
- **Maintainers**: Full project access and responsibility

## 📞 Community and Support

### Getting Help

- **GitHub Issues**: Report bugs and request features
- **GitHub Discussions**: Ask questions and share ideas
- **Code Reviews**: Learn from feedback on pull requests
- **Documentation**: Check existing docs first

### Communication Channels

- **Issues**: For bug reports and feature requests
- **Pull Requests**: For code contributions and technical discussions
- **Discussions**: For questions, ideas, and community support

### Response Times

- **Issues**: We aim to respond within 48 hours
- **Pull Requests**: Initial review within 72 hours
- **Security Issues**: Immediate attention (see security policy)

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the same license as the project (Flamingo AI Unified License v1.0).

## 🙏 Thank You

Thank you for contributing to OpenFrame CLI! Your contributions help make this tool better for the entire community. Whether you're fixing a small typo or adding a major feature, every contribution is valuable and appreciated.

---

**Happy Contributing!** 🚀

If you have any questions about contributing, please don't hesitate to ask through GitHub issues or discussions.