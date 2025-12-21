# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to the [Flamingo Community Code of Conduct](https://www.flamingo.run/code-of-conduct). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** - [Installation guide](https://golang.org/doc/install)
- **Docker** - For K3d cluster testing
- **Git** - For version control
- **kubectl** - Kubernetes command-line tool
- **k3d** - Lightweight Kubernetes distribution
- **helm** - Kubernetes package manager

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:

```bash
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli
```

3. Add the upstream remote:

```bash
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
```

## Development Environment

### Initial Setup

1. **Install dependencies:**

```bash
go mod download
```

2. **Build the CLI:**

```bash
go build -o openframe .
```

3. **Run tests to verify setup:**

```bash
go test ./...
```

### Development Tools

We use several tools to maintain code quality:

- **golangci-lint** - Code linting
- **gofmt** - Code formatting  
- **go mod tidy** - Dependency management
- **testify** - Testing framework

Install development tools:

```bash
# Install golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

## Development Workflow

### Branch Strategy

- **main** - Production-ready code
- **feature/*** - New features
- **fix/*** - Bug fixes
- **docs/*** - Documentation updates

### Making Changes

1. **Create a feature branch:**

```bash
git checkout -b feature/your-feature-name
```

2. **Make your changes** following our coding standards

3. **Test your changes:**

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/cluster/...

# Run with verbose output
go test -v ./...
```

4. **Lint your code:**

```bash
golangci-lint run
```

5. **Format your code:**

```bash
go fmt ./...
go mod tidy
```

## Coding Standards

### Go Style Guidelines

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Maintain consistent naming conventions

### Project Structure

```
openframe-cli/
├── cmd/                    # CLI commands
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart management commands
│   └── dev/               # Development tools commands
├── internal/              # Internal packages
│   ├── cluster/           # Cluster management logic
│   ├── chart/             # Chart management logic
│   ├── dev/               # Development tools logic
│   └── shared/            # Shared utilities
└── docs/                  # Documentation
```

### Code Conventions

- **Package naming:** Use lowercase, single-word package names
- **Function naming:** Use camelCase with descriptive names
- **Error handling:** Always check and handle errors appropriately
- **Comments:** Use godoc-style comments for exported functions
- **Testing:** Write tests for all public functions

### Error Handling

```go
// Good: Specific error messages with context
if err != nil {
    return fmt.Errorf("failed to create cluster %s: %w", name, err)
}

// Good: Custom error types for different scenarios
type ClusterNotFoundError struct {
    Name string
}

func (e *ClusterNotFoundError) Error() string {
    return fmt.Sprintf("cluster %s not found", e.Name)
}
```

### Interface Design

- Keep interfaces small and focused
- Use dependency injection for testability
- Mock external dependencies in tests

```go
// Good: Small, focused interface
type ClusterManager interface {
    CreateCluster(config ClusterConfig) error
    DeleteCluster(name string) error
}

// Good: Dependency injection
type Service struct {
    clusterManager ClusterManager
    executor       CommandExecutor
}
```

## Testing

### Test Organization

- **Unit tests:** Test individual functions and methods
- **Integration tests:** Test component interactions
- **End-to-end tests:** Test complete workflows

### Writing Tests

```go
func TestClusterService_CreateCluster(t *testing.T) {
    // Arrange
    mockManager := &mocks.ClusterManager{}
    service := NewClusterService(mockManager)
    
    config := ClusterConfig{
        Name: "test-cluster",
        Nodes: 3,
    }
    
    mockManager.On("CreateCluster", config).Return(nil)
    
    // Act
    err := service.CreateCluster(config)
    
    // Assert
    assert.NoError(t, err)
    mockManager.AssertExpectations(t)
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test -run TestClusterService_CreateCluster ./internal/cluster

# Run tests with race detection
go test -race ./...
```

## Pull Request Process

### Before Submitting

1. **Sync with upstream:**

```bash
git fetch upstream
git rebase upstream/main
```

2. **Ensure all tests pass:**

```bash
go test ./...
golangci-lint run
```

3. **Update documentation** if needed

4. **Commit with descriptive messages:**

```bash
git commit -m "feat: add cluster validation before creation

- Add validation for cluster name format
- Check for existing clusters with same name
- Return descriptive error messages

Fixes #123"
```

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Formatting changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### PR Template

When submitting a PR, include:

- **Description** of changes
- **Motivation** and context
- **Testing** performed
- **Breaking changes** (if any)
- **Related issues** (if any)

## Issue Guidelines

### Bug Reports

Include:

- **Steps to reproduce**
- **Expected vs actual behavior**
- **Environment details** (OS, Go version, CLI version)
- **Error messages** and logs

### Feature Requests

Include:

- **Use case** and motivation
- **Proposed solution** (if any)
- **Alternative approaches** considered
- **Impact** on existing functionality

## Release Process

Releases are managed by maintainers and follow semantic versioning:

- **Patch** (1.0.1) - Bug fixes
- **Minor** (1.1.0) - New features (backwards compatible)
- **Major** (2.0.0) - Breaking changes

### Release Checklist

1. Update version numbers
2. Update CHANGELOG.md
3. Create release tag
4. Build and publish binaries
5. Update documentation

## Getting Help

- **Documentation:** Check the [docs](./docs/README.md) directory
- **Issues:** Search existing [GitHub issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **Discussions:** Use [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)
- **Community:** Join the [Flamingo Discord](https://discord.gg/flamingo)

## Recognition

Contributors will be recognized in:

- GitHub contributor list
- Release notes
- Annual contributor highlights

Thank you for contributing to OpenFrame CLI! 🎉