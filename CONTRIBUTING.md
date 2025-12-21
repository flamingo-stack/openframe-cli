# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! We welcome contributions from the community and are excited to work with you.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Environment](#development-environment)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Documentation](#documentation)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. We are committed to providing a welcoming and inspiring community for all.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/openframe-cli.git
   cd openframe-cli
   ```
3. **Set up the upstream remote**:
   ```bash
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   ```

## Development Environment

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Docker** - [Install Docker](https://docs.docker.com/get-docker/)
- **K3d** - [Install K3d](https://k3d.io/v5.6.0/#installation)
- **Helm** - [Install Helm](https://helm.sh/docs/intro/install/)
- **Git** - [Install Git](https://git-scm.com/downloads)

### Setup

1. **Install dependencies**:
   ```bash
   go mod download
   ```

2. **Build the project**:
   ```bash
   go build -o openframe .
   ```

3. **Run tests**:
   ```bash
   go test ./...
   ```

4. **Verify installation**:
   ```bash
   ./openframe --help
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit them with clear messages:
   ```bash
   git add .
   git commit -m "Add feature: description of your changes"
   ```

3. **Keep your fork updated**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

## Making Changes

### Project Structure

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/         # Bootstrap commands
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart management commands
│   └── dev/               # Development tool commands
├── internal/              # Internal packages
│   ├── bootstrap/         # Bootstrap service logic
│   ├── cluster/           # Cluster management logic
│   ├── chart/             # Chart management logic
│   ├── dev/               # Development tools logic
│   └── shared/            # Shared utilities
├── docs/                  # Documentation
└── tests/                 # Test files
```

### Adding New Commands

1. **Create command file** in appropriate `cmd/` subdirectory
2. **Implement service logic** in corresponding `internal/` package
3. **Add tests** for both command and service layers
4. **Update documentation** in `docs/` directory

### Adding New Features

1. **Check existing issues** to avoid duplicate work
2. **Open an issue** to discuss the feature before implementing
3. **Follow the existing patterns** in the codebase
4. **Add comprehensive tests** for new functionality
5. **Update documentation** as needed

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/cluster/...

# Run tests with verbose output
go test -v ./...
```

### Test Structure

- **Unit tests** - Test individual functions and methods
- **Integration tests** - Test component interactions
- **End-to-end tests** - Test complete workflows

### Writing Tests

1. **Follow Go testing conventions**
2. **Use table-driven tests** when appropriate
3. **Mock external dependencies** (Docker, K3d, etc.)
4. **Test both success and error cases**
5. **Aim for high test coverage** (>80%)

Example test structure:
```go
func TestClusterCreate(t *testing.T) {
    tests := []struct {
        name     string
        input    ClusterConfig
        expected error
    }{
        {
            name: "valid cluster creation",
            input: ClusterConfig{Name: "test"},
            expected: nil,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Ensure tests pass**:
   ```bash
   go test ./...
   go vet ./...
   ```

2. **Update documentation** if needed

3. **Create a pull request** with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference to related issues
   - Screenshots for UI changes (if applicable)

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
```

### Review Process

1. **Automated checks** must pass (CI/CD)
2. **Code review** by maintainers
3. **Address feedback** and update as needed
4. **Approval and merge** by maintainers

## Code Style

### Go Style Guidelines

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format code
- Use `go vet` to check for issues
- Follow [Effective Go](https://golang.org/doc/effective_go.html) principles

### Specific Conventions

1. **Package naming** - Use lowercase, single word package names
2. **Function naming** - Use camelCase, start with capital for exported functions
3. **Error handling** - Always check and handle errors appropriately
4. **Comments** - Add comments for exported functions and complex logic
5. **Imports** - Group standard, third-party, and local imports

### Code Formatting

```bash
# Format code
go fmt ./...

# Vet code
go vet ./...

# Run linter (if available)
golangci-lint run
```

## Documentation

### Types of Documentation

1. **Code comments** - Inline documentation for functions and complex logic
2. **README updates** - Project-level documentation changes
3. **User guides** - Step-by-step instructions in `docs/getting-started/`
4. **Developer guides** - Technical documentation in `docs/development/`
5. **Architecture docs** - System design in `docs/reference/`

### Documentation Standards

- Use clear, concise language
- Provide examples where helpful
- Keep documentation up-to-date with code changes
- Follow markdown formatting guidelines
- Include diagrams for complex concepts

### Updating Documentation

When making changes that affect user-facing functionality:

1. **Update command help text** in CLI code
2. **Update relevant markdown files** in `docs/`
3. **Add examples** for new features
4. **Update architecture diagrams** if needed

## Getting Help

- **Documentation** - Check `docs/` directory first
- **Issues** - Search existing issues or create a new one
- **Discussions** - Use GitHub Discussions for questions
- **Discord** - Join the Flamingo community Discord (link in main README)

## Recognition

Contributors will be recognized in:
- Project README contributors section
- Release notes for significant contributions
- Special recognition for first-time contributors

Thank you for contributing to OpenFrame CLI! Your efforts help make this tool better for the entire community.

---

*This contributing guide is inspired by open source best practices and tailored for the OpenFrame CLI project.*