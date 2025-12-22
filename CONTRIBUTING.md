# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## 🚀 Quick Start for Contributors

### Prerequisites

- **Go 1.21+** - Programming language
- **Docker** - For running K3d clusters
- **Git** - Version control
- **Make** - Build automation (optional but recommended)

### Development Setup

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/openframe-cli.git
   cd openframe-cli
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   ```

3. **Build and Test**
   ```bash
   # Build the CLI
   go build -o openframe .
   
   # Run tests
   go test ./...
   
   # Run with verbose output
   go test -v ./...
   ```

4. **Run Locally**
   ```bash
   # Test your changes
   ./openframe --help
   ./openframe cluster create --help
   ```

## 🏗 Project Structure

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and global flags
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   ├── bootstrap/         # Bootstrap workflow commands
│   └── dev/               # Development workflow commands
├── internal/              # Internal packages (not public API)
│   ├── cluster/           # Cluster management business logic
│   ├── chart/             # Chart installation logic
│   ├── bootstrap/         # Bootstrap orchestration
│   ├── dev/               # Development tools integration
│   └── shared/            # Shared utilities
│       ├── executor/      # Command execution abstraction
│       ├── ui/            # Terminal UI components
│       └── prerequisites/ # Tool validation and installation
├── docs/                  # Documentation
├── scripts/               # Build and automation scripts
└── test/                  # Test utilities and fixtures
```

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/cluster/...

# Run integration tests (requires Docker)
go test -tags=integration ./...
```

### Test Organization

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test interactions with external tools (K3d, Helm, etc.)
- **Mock Tests**: Use mocks for external dependencies when appropriate

### Writing Tests

```go
func TestClusterCreate(t *testing.T) {
    // Arrange
    mockExecutor := &mocks.CommandExecutor{}
    clusterService := cluster.NewService(mockExecutor)
    
    // Act
    err := clusterService.Create("test-cluster")
    
    // Assert
    assert.NoError(t, err)
    mockExecutor.AssertExpectations(t)
}
```

## 📝 Code Style

### Go Conventions

- Follow standard Go formatting (`go fmt`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and small
- Use interfaces for testability

### CLI Design Principles

1. **Interactive First**: Provide helpful prompts and guidance
2. **Error Handling**: Clear, actionable error messages
3. **Progress Feedback**: Show progress for long-running operations
4. **Validation**: Validate inputs early and provide helpful suggestions
5. **Consistency**: Consistent command patterns and flag usage

### Example Command Structure

```go
func NewCreateCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create [cluster-name]",
        Short: "Create a new K3d cluster",
        Long:  "Create a new K3d cluster with interactive configuration",
        Args:  cobra.MaximumNArgs(1),
        RunE:  runCreate,
    }
    
    cmd.Flags().String("config", "", "Path to cluster configuration file")
    cmd.Flags().Bool("interactive", true, "Use interactive mode")
    
    return cmd
}
```

## 🔄 Development Workflow

### Branch Strategy

- `main` - Stable release branch
- `develop` - Integration branch for new features
- `feature/feature-name` - Feature development branches
- `fix/bug-description` - Bug fix branches

### Pull Request Process

1. **Create Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Changes**
   - Write code following our style guidelines
   - Add tests for new functionality
   - Update documentation as needed

3. **Test Thoroughly**
   ```bash
   go test ./...
   go build -o openframe .
   ./openframe cluster create  # Test manually
   ```

4. **Commit Changes**
   ```bash
   git add .
   git commit -m "feat: add cluster configuration validation"
   ```

5. **Push and Create PR**
   ```bash
   git push origin feature/your-feature-name
   ```
   Then create a pull request on GitHub.

### Commit Message Format

Use conventional commits for clear history:

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `test:` - Test additions or changes
- `refactor:` - Code refactoring
- `style:` - Code style changes
- `ci:` - CI/CD changes

Examples:
```
feat: add cluster configuration validation
fix: handle missing Docker daemon gracefully
docs: update installation instructions
test: add integration tests for bootstrap command
```

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment Information**
   - OS and version
   - Go version
   - Docker version
   - OpenFrame CLI version

2. **Steps to Reproduce**
   ```bash
   openframe cluster create
   # Error occurs here
   ```

3. **Expected vs Actual Behavior**
   - What you expected to happen
   - What actually happened

4. **Logs and Error Messages**
   ```bash
   # Run with debug logging
   openframe --verbose cluster create
   ```

## 💡 Feature Requests

For new features:

1. **Use Case**: Describe the problem you're trying to solve
2. **Proposed Solution**: Your ideas for implementation
3. **Alternatives**: Other approaches you've considered
4. **Impact**: Who would benefit from this feature

## 📚 Documentation

### Updating Documentation

- Update relevant `*.md` files in the `docs/` directory
- Update command help text in the code
- Add examples for new features
- Update README.md if the change affects basic usage

### Documentation Guidelines

- Use clear, concise language
- Include code examples
- Provide context for why something works a certain way
- Test all documented commands and examples

## 🎯 Areas for Contribution

We welcome contributions in these areas:

### High Priority
- **Windows Support**: Improve Windows compatibility
- **Error Handling**: Better error messages and recovery
- **Performance**: Optimize cluster creation and chart installation
- **Testing**: Increase test coverage, especially integration tests

### Medium Priority
- **Cloud Providers**: Support for EKS, GKE, AKS
- **CI/CD Integration**: GitHub Actions, GitLab CI examples
- **Monitoring**: Built-in cluster monitoring and alerting
- **Security**: Security scanning and hardening features

### Always Welcome
- **Bug Fixes**: Any bug fixes with tests
- **Documentation**: Improvements to docs and examples
- **User Experience**: UI/UX improvements for better usability
- **Code Quality**: Refactoring, performance improvements

## 🤝 Community

- **Discussions**: Use GitHub Discussions for questions and ideas
- **Issues**: Use GitHub Issues for bugs and feature requests
- **Code Review**: All changes go through code review process
- **Respectful Communication**: Follow our Code of Conduct

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the Flamingo AI Unified License v1.0.

---

Thank you for contributing to OpenFrame CLI! 🚀