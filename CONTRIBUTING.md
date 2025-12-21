# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide will help you get started with setting up your development environment and understanding our contribution workflow.

## 🚀 Quick Start for Contributors

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/openframe-cli.git
   cd openframe-cli
   ```

2. **Set Up Development Environment**
   ```bash
   # Install dependencies
   make setup
   
   # Build the project
   make build
   
   # Run tests
   make test
   ```

3. **Make Your Changes**
   ```bash
   # Create a feature branch
   git checkout -b feature/your-feature-name
   
   # Make your changes and test
   make test
   make lint
   
   # Commit your changes
   git commit -m "Add your descriptive commit message"
   ```

4. **Submit a Pull Request**
   ```bash
   git push origin feature/your-feature-name
   # Then open a PR on GitHub
   ```

## 📋 Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** - Primary development language
- **Docker** - For running integration tests and K3d clusters
- **Make** - For build automation
- **Git** - For version control

### Optional Development Tools
- **kubectl** - For testing Kubernetes interactions
- **k3d** - For local cluster testing
- **golangci-lint** - For code linting (or use `make lint`)

## 🛠️ Development Environment Setup

### 1. Initial Setup

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Install development dependencies
make setup

# Verify setup
make verify
```

### 2. Building the Project

```bash
# Build for your platform
make build

# Build for all platforms
make build-all

# Install to your PATH
make install
```

### 3. Running Tests

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests (requires Docker)
make test-integration

# Run with coverage
make test-coverage
```

### 4. Code Quality

```bash
# Run linting
make lint

# Format code
make format

# Run full validation (lint + test + build)
make validate
```

## 📁 Project Structure

```
openframe-cli/
├── cmd/                    # Command definitions
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart management commands
│   └── dev/               # Development tool commands
├── internal/              # Internal packages
│   ├── cluster/           # Cluster service logic
│   ├── chart/             # Chart service logic
│   ├── dev/               # Development service logic
│   └── shared/            # Shared utilities
│       ├── ui/            # User interface components
│       ├── errors/        # Error handling
│       └── models/        # Data models
├── pkg/                   # Public packages
├── test/                  # Test files and fixtures
├── scripts/               # Build and development scripts
├── docs/                  # Documentation
└── Makefile              # Build automation
```

## 🎯 Types of Contributions

We welcome several types of contributions:

### 🐛 Bug Reports
- Use GitHub Issues with the "bug" label
- Provide clear reproduction steps
- Include system information and CLI version

### ✨ Feature Requests
- Use GitHub Issues with the "enhancement" label
- Describe the use case and expected behavior
- Discuss implementation approach if you have ideas

### 📖 Documentation
- Improve existing documentation
- Add examples and tutorials
- Fix typos and clarify confusing sections

### 💻 Code Contributions
- Bug fixes
- New features
- Performance improvements
- Code refactoring

## 📝 Contribution Guidelines

### Git Workflow

1. **Fork** the repository to your GitHub account
2. **Clone** your fork locally
3. **Create** a feature branch from `main`
4. **Make** your changes with clear, descriptive commits
5. **Test** your changes thoroughly
6. **Push** to your fork and **submit** a pull request

### Commit Message Format

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(cluster): add support for custom cluster configurations
fix(bootstrap): resolve ArgoCD installation timeout issue
docs(readme): update quick start installation instructions
```

### Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `golangci-lint` for linting
- Add tests for new functionality
- Update documentation for user-facing changes

### Testing Requirements

- **Unit Tests**: Required for all new functions and methods
- **Integration Tests**: Required for command-line interface changes
- **End-to-End Tests**: Encouraged for major features
- **Test Coverage**: Aim for >80% coverage on new code

#### Writing Tests

```bash
# Run specific test package
go test ./internal/cluster/...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### Documentation Standards

- Update relevant documentation for user-facing changes
- Include code examples in docstrings
- Update CLI help text for command changes
- Add or update architectural diagrams for structural changes

## 🔄 Pull Request Process

### Before Submitting

1. ✅ **Tests pass**: `make test`
2. ✅ **Linting passes**: `make lint`
3. ✅ **Build succeeds**: `make build`
4. ✅ **Documentation updated**: For user-facing changes
5. ✅ **Changelog entry**: For significant changes

### Pull Request Template

When opening a PR, please include:

- **Description** of changes and motivation
- **Type of change** (bug fix, feature, docs, etc.)
- **Testing** performed and instructions for reviewers
- **Breaking changes** if any
- **Related issues** that this PR addresses

### Review Process

1. **Automated Checks**: CI/CD will run tests and linting
2. **Maintainer Review**: Core maintainers will review the code
3. **Feedback**: Address any requested changes
4. **Approval**: PR will be approved when ready
5. **Merge**: Maintainers will merge the PR

## 🧪 Testing Guidelines

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    // Arrange
    input := "test-input"
    expected := "expected-output"
    
    // Act
    result := FunctionToTest(input)
    
    // Assert
    assert.Equal(t, expected, result)
}
```

### Integration Test Example

```go
func TestClusterCreate(t *testing.T) {
    // Skip if Docker not available
    if !isDockerAvailable() {
        t.Skip("Docker not available")
    }
    
    // Test cluster creation
    clusterName := "test-cluster"
    defer cleanupCluster(clusterName)
    
    err := CreateCluster(clusterName)
    assert.NoError(t, err)
    
    // Verify cluster exists
    exists := ClusterExists(clusterName)
    assert.True(t, exists)
}
```

## 🏷️ Release Process

Releases are managed by maintainers, but contributors should be aware of the process:

1. **Version Bump**: Update version in relevant files
2. **Changelog**: Update CHANGELOG.md with release notes
3. **Tag**: Create and push a Git tag
4. **Release**: GitHub Actions builds and publishes binaries
5. **Documentation**: Update installation documentation

## 🤝 Community Guidelines

- **Be Respectful**: Treat everyone with respect and kindness
- **Be Patient**: Maintainers are volunteers with limited time
- **Be Constructive**: Provide helpful feedback and suggestions
- **Ask Questions**: Don't hesitate to ask for clarification or help

## 📞 Getting Help

- **Documentation**: Check the [docs](./docs/README.md) first
- **Issues**: Search existing GitHub Issues
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Community**: Join our community channels (links in README)

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the same license as the project (Flamingo AI Unified License v1.0).

---

Thank you for contributing to OpenFrame CLI! Your efforts help make the project better for everyone. 🎉