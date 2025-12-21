# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## 🎯 Ways to Contribute

We welcome various types of contributions:

- **🐛 Bug Reports** - Help us identify and fix issues
- **✨ Feature Requests** - Suggest new functionality
- **📝 Documentation** - Improve guides and references
- **💻 Code Contributions** - Implement features and fixes
- **🧪 Testing** - Add test coverage and quality assurance
- **💬 Community Support** - Help other users in discussions

## 🚀 Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.19+** installed
- **Docker** Desktop or Engine running
- **kubectl** command-line tool
- **Git** for version control
- **Make** for build automation

### Setting Up Your Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/your-username/openframe-cli.git
   cd openframe-cli
   ```

2. **Set up development tools:**
   ```bash
   make dev-setup
   ```

3. **Build and test:**
   ```bash
   make build
   make test
   ```

4. **Verify installation:**
   ```bash
   ./openframe --help
   ```

For detailed setup instructions, see our [Development Environment Guide](./docs/development/setup/environment.md).

## 📋 Development Workflow

### 1. Planning Your Contribution

- **For bugs:** Create an issue describing the problem
- **For features:** Open a discussion or issue to validate the idea
- **Check existing issues:** Avoid duplicate work

### 2. Creating a Feature Branch

```bash
# Create and switch to a new branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 3. Making Changes

- **Follow the existing code structure and patterns**
- **Write tests** for new functionality
- **Update documentation** as needed
- **Keep commits atomic** and well-described

### 4. Testing Your Changes

```bash
# Run all tests
make test

# Run specific test package
go test ./internal/cluster/...

# Run with coverage
make test-coverage

# Test the CLI manually
./openframe bootstrap test-cluster
```

### 5. Submitting Your Contribution

```bash
# Push your branch
git push origin feature/your-feature-name

# Create a Pull Request on GitHub
# Include a clear description of changes
```

## 📏 Code Standards

### Go Code Style

We follow standard Go conventions:

- **Use `gofmt`** for consistent formatting
- **Follow effective Go practices** 
- **Use descriptive variable and function names**
- **Add comments for exported functions and complex logic**
- **Handle errors appropriately**

### Project Structure

```text
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management
│   ├── chart/              # Chart management  
│   └── dev/                # Development tools
├── internal/               # Internal packages
│   ├── [module]/models/    # Data structures
│   ├── [module]/services/  # Business logic
│   ├── [module]/ui/        # User interfaces
│   └── shared/             # Shared utilities
```

### Naming Conventions

- **Commands:** Use verb-noun pattern (`cluster create`, `chart install`)
- **Files:** Use kebab-case (`cluster-config.go`)
- **Functions:** Use camelCase (`CreateCluster`, `validateConfig`)
- **Constants:** Use SCREAMING_SNAKE_CASE (`DEFAULT_TIMEOUT`)

## 🧪 Testing Guidelines

### Test Coverage

- **Unit tests:** For business logic and utilities
- **Integration tests:** For command workflows
- **End-to-end tests:** For complete user scenarios

### Writing Tests

```go
func TestCreateCluster(t *testing.T) {
    // Arrange
    config := &ClusterConfig{Name: "test-cluster"}
    
    // Act
    result, err := CreateCluster(config)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "test-cluster", result.Name)
}
```

### Test Commands

```bash
# Run all tests
make test

# Run with verbose output
make test-verbose

# Generate coverage report
make test-coverage

# Run integration tests
make test-integration
```

## 📝 Documentation Standards

### Code Documentation

- **Add godoc comments** for exported functions
- **Include usage examples** in complex functions
- **Document parameters and return values**

### User Documentation

- **Keep instructions clear and actionable**
- **Include code examples with expected output**
- **Test all documented commands**
- **Update relevant sections** when adding features

### Documentation Structure

- **Getting Started:** User-facing guides
- **Development:** Contributor documentation  
- **Reference:** Technical specifications
- **Architecture:** System design documentation

## 🔍 Pull Request Guidelines

### PR Title Format

Use clear, descriptive titles:
- `feat: add cluster status command`
- `fix: resolve k3d cleanup issue`
- `docs: update quick start guide`
- `test: add cluster creation tests`

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Test improvement
- [ ] Refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project conventions
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] No breaking changes (or marked as such)
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **Code review** by maintainers
3. **Testing** verification if needed
4. **Approval** and merge by maintainers

## 🏗️ Architecture Guidelines

### Adding New Commands

1. **Create command in appropriate `cmd/` directory**
2. **Add business logic to `internal/[module]/services/`**
3. **Define data models in `internal/[module]/models/`**
4. **Create UI components in `internal/[module]/ui/`**

### Code Organization

- **Separation of concerns:** Keep UI, business logic, and models separate
- **Dependency injection:** Use interfaces for testability
- **Error handling:** Consistent error types and messages
- **Configuration:** Centralized configuration management

### External Dependencies

- **Minimize dependencies:** Only add if truly necessary
- **Use standard library** when possible
- **Pin versions** in `go.mod`
- **Document new dependencies** in PR description

## 🚨 Issue and Bug Reporting

### Bug Reports

Include the following information:

```markdown
## Bug Description
Clear description of the issue

## Steps to Reproduce
1. Run command: `openframe cluster create test`
2. Observe error message
3. Check cluster status

## Expected Behavior
What should have happened

## Actual Behavior
What actually happened

## Environment
- OS: [e.g., macOS 12.0, Ubuntu 20.04]
- OpenFrame CLI version: [e.g., v1.0.0]
- Docker version: [e.g., 20.10.8]
- kubectl version: [e.g., v1.21.0]

## Additional Context
Any other relevant information
```

### Feature Requests

```markdown
## Feature Description
Clear description of the proposed feature

## Use Case
Why is this feature needed?

## Proposed Solution
How do you envision this working?

## Alternatives Considered
Other approaches you've thought about

## Additional Context
Any other relevant information
```

## 📋 Development Commands

### Build Commands
```bash
make build          # Build the CLI binary
make build-all      # Build for all platforms
make clean          # Clean build artifacts
```

### Testing Commands
```bash
make test           # Run all tests
make test-unit      # Run unit tests only
make test-integration # Run integration tests
make test-coverage  # Generate coverage report
```

### Quality Commands
```bash
make lint           # Run Go linters
make fmt            # Format Go code
make vet            # Run Go vet
make check          # Run all quality checks
```

### Development Commands
```bash
make dev-setup      # Set up development environment
make dev-build      # Build for development
make dev-test       # Run tests in development mode
```

## 🎓 Learning Resources

### Go Development
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Go Testing](https://golang.org/pkg/testing/)

### Kubernetes Development
- [Kubernetes Client Library](https://github.com/kubernetes/client-go)
- [Kubectl Source Code](https://github.com/kubernetes/kubectl)
- [Helm SDK](https://helm.sh/docs/topics/advanced/#go-sdk)

### CLI Development
- [Cobra CLI Framework](https://cobra.dev/)
- [CLI Design Guidelines](https://clig.dev/)

## 🤝 Community Guidelines

### Code of Conduct

- **Be respectful** and inclusive
- **Help others** learn and contribute
- **Focus on constructive feedback**
- **Assume good intentions**

### Communication

- **GitHub Issues:** Bug reports and feature requests
- **GitHub Discussions:** Questions and general discussion
- **Pull Requests:** Code review and collaboration
- **Documentation:** Clear and helpful guides

## 📞 Getting Help

### For Contributors

- **Review existing documentation** in the `docs/` directory
- **Check GitHub issues** for similar questions
- **Ask in GitHub discussions** for general questions
- **Create detailed issues** for specific problems

### For Maintainers

- **Respond to issues and PRs** in a timely manner
- **Provide constructive feedback** on contributions
- **Maintain project documentation** and guidelines
- **Help onboard new contributors**

## 🎉 Recognition

We appreciate all contributions! Contributors are recognized in:

- **Release notes** for significant contributions
- **README contributors section** for ongoing contributors
- **GitHub contributors graph** for all code contributors

---

Thank you for contributing to OpenFrame CLI! Your efforts help make Kubernetes development more accessible and enjoyable for developers everywhere. 🚀