# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Standards](#code-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Community Guidelines](#community-guidelines)

## Getting Started

### Prerequisites

Before contributing, ensure you have the following tools installed:

- **Go 1.21+** - Programming language runtime
- **Docker** - For running K3d clusters during testing
- **Git** - Version control system
- **Make** - Build automation tool

Optional but recommended:
- **K3d** - For testing cluster operations
- **kubectl** - For Kubernetes cluster interaction
- **Helm** - For testing chart installations

### First Contribution

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/openframe-cli.git
   cd openframe-cli
   ```
3. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```
4. **Make your changes** following our guidelines
5. **Test your changes** thoroughly
6. **Submit a pull request** with a clear description

## Development Setup

### Local Environment

1. **Clone the repository**:
   ```bash
   git clone https://github.com/flamingo-stack/openframe-cli.git
   cd openframe-cli
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build the CLI**:
   ```bash
   go build -o openframe .
   ```

4. **Run tests**:
   ```bash
   go test ./...
   ```

### Project Structure

```
openframe-cli/
├── cmd/                    # Command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management commands
│   ├── chart/              # Chart installation commands
│   └── dev/                # Development tool commands
├── internal/               # Internal packages
│   ├── cluster/            # Cluster management logic
│   ├── chart/              # Chart installation logic
│   ├── dev/                # Development tool integration
│   ├── bootstrap/          # Bootstrap orchestration
│   └── shared/             # Shared utilities and UI
├── docs/                   # Documentation
└── tests/                  # Test files
```

## Development Workflow

### Branch Naming

Use descriptive branch names with prefixes:

- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

Examples:
- `feature/add-cluster-metrics`
- `fix/helm-installation-timeout`
- `docs/update-getting-started`

### Commit Messages

Follow conventional commit format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Code formatting
- `refactor` - Code refactoring
- `test` - Test changes
- `chore` - Maintenance tasks

Examples:
```
feat(cluster): add cluster health monitoring

fix(chart): resolve ArgoCD installation timeout issue

docs(readme): update installation instructions
```

### Testing Strategy

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test complete workflows
4. **Manual Testing** - Test CLI commands interactively

## Code Standards

### Go Coding Standards

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for code formatting
- Run `golint` and address linting issues
- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for exported functions and types

### Code Organization

- **Commands** (`cmd/`) - CLI command definitions only
- **Business Logic** (`internal/*/services/`) - Core functionality
- **Models** (`internal/*/models/`) - Data structures and validation
- **UI Components** (`internal/shared/ui/`) - Reusable UI elements
- **Utilities** (`internal/shared/`) - Common helper functions

### Error Handling

- Use structured error handling with context
- Provide clear, actionable error messages
- Include troubleshooting hints where appropriate
- Log errors appropriately for debugging

Example:
```go
if err := cluster.Create(name); err != nil {
    return fmt.Errorf("failed to create cluster %s: %w\n\nTroubleshooting:\n- Ensure Docker is running\n- Check available disk space", name, err)
}
```

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cluster/...

# Run tests with verbose output
go test -v ./...
```

### Writing Tests

1. **Test File Naming** - `*_test.go`
2. **Function Naming** - `TestFunctionName`
3. **Table Tests** - Use for multiple test cases
4. **Mocking** - Mock external dependencies
5. **Test Coverage** - Aim for >80% coverage

Example test structure:
```go
func TestClusterCreate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    error
        wantErr bool
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Integration Testing

For testing commands that interact with external tools:

1. **Use Test Clusters** - Create temporary K3d clusters
2. **Clean Up Resources** - Always clean up test resources
3. **Mock External Calls** - Mock calls to external APIs when possible
4. **Environment Variables** - Use environment variables for test configuration

## Pull Request Process

### Before Submitting

1. **Update documentation** if your changes affect user-facing functionality
2. **Add or update tests** for new features or bug fixes
3. **Run the test suite** and ensure all tests pass
4. **Check code formatting** with `gofmt`
5. **Update CHANGELOG** if applicable

### Pull Request Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Code refactoring

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Documentation
- [ ] README updated (if applicable)
- [ ] Documentation updated (if applicable)
- [ ] Comments added to code

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Tests added/updated
- [ ] Documentation updated
```

### Review Process

1. **Automated Checks** - CI/CD pipeline runs tests and linting
2. **Code Review** - Team members review your changes
3. **Address Feedback** - Make requested changes
4. **Final Approval** - Maintainer approves and merges

## Issue Reporting

### Bug Reports

Use the bug report template and include:

- **Description** - Clear description of the issue
- **Steps to Reproduce** - Detailed reproduction steps
- **Expected Behavior** - What should happen
- **Actual Behavior** - What actually happens
- **Environment** - OS, Go version, Docker version
- **Logs** - Relevant error messages or logs

### Feature Requests

Use the feature request template and include:

- **Problem Statement** - What problem does this solve?
- **Proposed Solution** - Detailed description of the feature
- **Alternatives** - Other solutions you've considered
- **Use Cases** - How would this feature be used?

### Security Issues

For security vulnerabilities:

1. **Do not** create a public issue
2. **Email** security@flamingo.run with details
3. **Include** steps to reproduce
4. **Wait** for response before public disclosure

## Community Guidelines

### Code of Conduct

We are committed to providing a welcoming and inclusive environment. Please:

- **Be respectful** of differing viewpoints and experiences
- **Use inclusive language** in all communications
- **Accept constructive feedback** gracefully
- **Focus on collaboration** and community building

### Communication

- **GitHub Issues** - Bug reports and feature requests
- **Pull Request Comments** - Code review discussions
- **Discussions** - General questions and community chat

### Getting Help

- **Documentation** - Check the [docs](./docs/README.md) first
- **GitHub Issues** - Search existing issues
- **Discussions** - Ask questions in GitHub Discussions
- **Discord** - Join our [community Discord](https://discord.gg/flamingo)

## Development Resources

### Useful Commands

```bash
# Build for all platforms
make build-all

# Run linting
make lint

# Run tests with coverage
make test-coverage

# Generate documentation
make docs

# Clean build artifacts
make clean
```

### External Documentation

- [Cobra CLI Framework](https://cobra.dev/)
- [K3d Documentation](https://k3d.io/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/)

---

Thank you for contributing to OpenFrame CLI! Your efforts help make this tool better for everyone in the community.

For questions about contributing, please open a Discussion or reach out to the maintainers.