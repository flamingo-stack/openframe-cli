# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! We welcome contributions from the community and are excited to work with you.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## 📜 Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/). By participating, you are expected to uphold this code. Please report unacceptable behavior to the maintainers.

**Our Standards:**
- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards community members
- Prioritize the collective good of the project

## 🚀 Getting Started

### Prerequisites

Before contributing, ensure you have the following tools installed:

- **Go 1.21+** - Programming language
- **Docker** - Container runtime
- **kubectl** - Kubernetes CLI
- **Helm** - Package manager for Kubernetes
- **K3d** - Lightweight Kubernetes for development
- **Git** - Version control

### Setting Up Your Development Environment

1. **Fork the Repository**
   ```bash
   # Fork https://github.com/flamingo-stack/openframe-cli to your GitHub account
   # Clone your fork
   git clone https://github.com/YOUR_USERNAME/openframe-cli.git
   cd openframe-cli
   ```

2. **Configure Remotes**
   ```bash
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   git remote -v
   ```

3. **Set Up Development Environment**
   ```bash
   # Install development dependencies
   make setup
   
   # Verify your setup
   make test
   make build
   ```

4. **Verify Installation**
   ```bash
   ./openframe --version
   ./openframe --help
   ```

For detailed setup instructions, see [Development Environment Setup](docs/development/setup/environment.md).

## 🔄 Development Workflow

### Creating a Feature Branch

```bash
# Sync with upstream
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### Making Changes

1. **Follow the Project Structure**
   ```
   cmd/           # CLI command implementations
   internal/      # Private application code
   docs/          # Documentation
   examples/      # Usage examples
   scripts/       # Build and development scripts
   ```

2. **Write Clean Code**
   - Follow Go conventions and idioms
   - Add comments for public functions
   - Keep functions focused and small
   - Use meaningful variable names

3. **Update Documentation**
   - Update relevant documentation in `docs/`
   - Add examples for new features
   - Update command help text

4. **Add Tests**
   - Write unit tests for new functionality
   - Add integration tests for CLI commands
   - Ensure existing tests still pass

### Testing Your Changes

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration

# Check code coverage
make coverage

# Lint your code
make lint

# Format code
make fmt
```

### Building and Testing Locally

```bash
# Build the binary
make build

# Test your changes manually
./openframe bootstrap test-cluster --verbose
./openframe cluster list
./openframe cluster delete test-cluster
```

## 🔁 Pull Request Process

### Before Submitting

1. **Ensure Quality**
   - [ ] All tests pass (`make test`)
   - [ ] Code is properly formatted (`make fmt`)
   - [ ] No linting errors (`make lint`)
   - [ ] Documentation is updated
   - [ ] Examples work as expected

2. **Commit Message Guidelines**
   ```
   type(scope): brief description
   
   Longer explanation if needed
   
   Fixes #123
   ```
   
   **Types:**
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation changes
   - `test`: Adding or updating tests
   - `refactor`: Code refactoring
   - `chore`: Maintenance tasks

   **Examples:**
   ```
   feat(cluster): add support for custom node labels
   fix(bootstrap): handle edge case in ArgoCD installation
   docs(cli): update help text for dev commands
   ```

### Submitting Your PR

1. **Push Changes**
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request**
   - Go to GitHub and create a PR from your branch
   - Use the PR template to provide details
   - Link related issues with `Fixes #123` or `Closes #456`
   - Add screenshots/demos for UI changes

3. **PR Template Checklist**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Unit tests added/updated
   - [ ] Integration tests pass
   - [ ] Manual testing completed
   
   ## Checklist
   - [ ] Code follows project standards
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] No breaking changes (or properly documented)
   ```

### Review Process

1. **Automated Checks**
   - CI/CD pipeline runs automatically
   - All tests must pass
   - Code coverage should not decrease significantly

2. **Manual Review**
   - At least one maintainer review required
   - Address feedback promptly and respectfully
   - Update PR based on review comments

3. **Approval and Merge**
   - Maintainer approval required
   - Squash and merge preferred
   - Delete feature branch after merge

## 📝 Coding Standards

### Go Code Style

1. **Follow Go Conventions**
   ```go
   // Good: Exported function with documentation
   // CreateCluster creates a new Kubernetes cluster with the specified configuration
   func CreateCluster(config ClusterConfig) error {
       if err := validateConfig(config); err != nil {
           return fmt.Errorf("invalid config: %w", err)
       }
       // Implementation...
   }
   ```

2. **Error Handling**
   ```go
   // Good: Wrap errors with context
   if err := someOperation(); err != nil {
       return fmt.Errorf("failed to perform operation: %w", err)
   }
   
   // Good: Check for specific error types
   if errors.Is(err, ErrClusterNotFound) {
       return handleMissingCluster()
   }
   ```

3. **Package Structure**
   ```go
   // internal/cluster/services/cluster.go
   package services
   
   import (
       "context"
       "fmt"
       
       "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
   )
   ```

### CLI Design Principles

1. **Consistent Command Structure**
   ```bash
   openframe <resource> <action> [name] [flags]
   
   # Examples:
   openframe cluster create my-cluster --verbose
   openframe chart install --deployment-mode=oss-tenant
   ```

2. **User-Friendly Output**
   ```go
   // Use shared UI components for consistency
   ui.PrintSuccess("✅ Cluster created successfully")
   ui.PrintError("❌ Failed to create cluster: %v", err)
   ui.PrintInfo("ℹ️  Checking prerequisites...")
   ```

3. **Interactive vs Non-Interactive**
   ```go
   // Support both modes
   if !nonInteractive {
       mode, err := ui.PromptSelect("Select deployment mode:", modes)
   } else {
       mode = defaultMode
   }
   ```

### Documentation Standards

1. **Code Documentation**
   ```go
   // ClusterService manages Kubernetes cluster lifecycle operations
   type ClusterService interface {
       // CreateCluster creates a new cluster with the specified configuration.
       // Returns an error if cluster creation fails or cluster already exists.
       CreateCluster(ctx context.Context, config ClusterConfig) error
   }
   ```

2. **Markdown Documentation**
   - Use clear headings and structure
   - Include code examples that work
   - Add troubleshooting sections
   - Keep examples up to date

## 🧪 Testing Guidelines

### Test Structure

```go
func TestCreateCluster(t *testing.T) {
    tests := []struct {
        name        string
        config      ClusterConfig
        expectError bool
        errorType   error
    }{
        {
            name:        "valid config creates cluster",
            config:      validClusterConfig(),
            expectError: false,
        },
        {
            name:        "invalid config returns error",
            config:      invalidClusterConfig(),
            expectError: true,
            errorType:   ErrInvalidConfig,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Categories

1. **Unit Tests** (`_test.go`)
   - Test individual functions and methods
   - Mock external dependencies
   - Fast execution (< 1 second each)

2. **Integration Tests** (`_integration_test.go`)
   - Test component interactions
   - Use real external tools when possible
   - Slower execution but more realistic

3. **CLI Tests**
   - Test complete command execution
   - Verify output formatting
   - Test error conditions

### Coverage Goals

- **Minimum**: 80% code coverage
- **Target**: 90% code coverage
- **Critical paths**: 100% coverage (cluster creation, bootstrap)

## 📚 Documentation

### Types of Documentation

1. **User Documentation** (`docs/getting-started/`)
   - Installation guides
   - Quick start tutorials
   - Command references

2. **Developer Documentation** (`docs/development/`)
   - Architecture overviews
   - Contributing guidelines
   - Setup instructions

3. **Reference Documentation** (`docs/reference/`)
   - API documentation
   - Configuration references
   - Troubleshooting guides

### Documentation Standards

- **Clear and Concise**: Write for your audience
- **Actionable**: Include working examples
- **Up-to-Date**: Update docs with code changes
- **Well-Structured**: Use consistent formatting

## 💬 Community

### Getting Help

1. **Documentation**: Check existing docs first
2. **GitHub Issues**: Search for similar issues
3. **Discord**: Join our development channel
4. **Maintainers**: Tag maintainers for urgent issues

### Communication Guidelines

- **Be Patient**: Maintainers are volunteers
- **Be Specific**: Provide detailed error messages and steps
- **Be Respectful**: Follow the code of conduct
- **Be Helpful**: Help others when you can

### Recognition

We appreciate all contributors! Contributors are recognized through:

- **GitHub**: Automatic contributor listing
- **Releases**: Contribution acknowledgments
- **Community**: Shout-outs in Discord and social media

## 🏷️ Issue Labels

| Label | Description |
|-------|-------------|
| `good first issue` | Good for newcomers |
| `bug` | Something isn't working |
| `enhancement` | New feature or request |
| `documentation` | Improvements to docs |
| `help wanted` | Extra attention is needed |
| `question` | Further information requested |
| `wontfix` | This will not be worked on |

## 📦 Release Process

1. **Version Planning**
   - Follow semantic versioning (SemVer)
   - Major: Breaking changes
   - Minor: New features (backward compatible)
   - Patch: Bug fixes

2. **Release Preparation**
   - Update changelog
   - Update version numbers
   - Test release candidates
   - Update documentation

3. **Release Distribution**
   - GitHub releases
   - Binary distributions
   - Documentation updates

## 🎉 Thank You

Thank you for contributing to OpenFrame CLI! Your involvement helps make this project better for everyone. Whether you're fixing bugs, adding features, improving documentation, or helping other users, your contributions are valued and appreciated.

For detailed development information, see our [Development Documentation](docs/development/README.md).

---

**Questions?** Feel free to reach out to the maintainers or ask in our Discord community. We're here to help!