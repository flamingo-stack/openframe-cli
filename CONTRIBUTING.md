# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This document provides guidelines and information for contributors.

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. We are committed to providing a welcoming and inclusive environment for all contributors.

## Getting Started

### Development Prerequisites

Before you begin development, ensure you have the following tools installed:

| Tool | Purpose | Installation |
|------|---------|--------------|
| **Go 1.21+** | Primary development language | [Install Go](https://golang.org/dl/) |
| **Docker** | Container runtime for testing | [Install Docker](https://docs.docker.com/get-docker/) |
| **kubectl** | Kubernetes CLI | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| **k3d** | Local Kubernetes clusters | `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh \| bash` |
| **Helm** | Package manager | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` |
| **Make** | Build automation | Usually pre-installed on Linux/macOS |

### Setting Up Development Environment

1. **Fork and Clone the Repository**

```bash
# Fork the repository on GitHub, then clone your fork
git clone https://github.com/YOUR-USERNAME/openframe-cli.git
cd openframe-cli

# Add the original repository as upstream
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
```

2. **Install Dependencies**

```bash
# Download Go modules
go mod download

# Install development tools
make install-tools
```

3. **Verify Your Setup**

```bash
# Build the project
make build

# Run tests
make test

# Run the CLI locally
./bin/openframe --help
```

## Development Workflow

### Branch Naming

Use descriptive branch names that follow this pattern:

- `feature/description` - New features
- `fix/description` - Bug fixes  
- `docs/description` - Documentation changes
- `refactor/description` - Code refactoring
- `test/description` - Test improvements

Examples:
- `feature/add-cluster-logs-command`
- `fix/chart-installation-timeout`
- `docs/update-getting-started-guide`

### Making Changes

1. **Create a Feature Branch**

```bash
git checkout main
git pull upstream main
git checkout -b feature/your-feature-name
```

2. **Make Your Changes**

Follow these guidelines when making changes:

- Write clear, descriptive commit messages
- Keep commits focused and atomic
- Add tests for new functionality
- Update documentation as needed
- Follow Go coding conventions
- Run `make lint` to check code style

3. **Test Your Changes**

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Test manually with local builds
make build
./bin/openframe cluster create test-cluster
```

4. **Commit and Push**

```bash
# Add your changes
git add .

# Commit with a descriptive message
git commit -m "feat: add cluster logs command for debugging

- Add 'openframe cluster logs' command
- Support filtering by service name
- Include timestamps in output
- Add unit tests and integration tests"

# Push to your fork
git push origin feature/your-feature-name
```

## Project Structure

Understanding the codebase structure will help you navigate and contribute effectively:

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/          # Cluster management commands
│   ├── chart/           # Chart installation commands
│   └── dev/            # Development tools commands
├── internal/              # Private application code
│   ├── bootstrap/        # Bootstrap orchestration logic
│   ├── cluster/         # Cluster management services
│   ├── chart/          # Chart installation services
│   ├── dev/           # Development tools implementation
│   └── shared/       # Common utilities and models
├── docs/               # Documentation
│   ├── tutorials/     # User and developer guides
│   └── dev/          # Architecture documentation
├── test/              # Integration tests
├── scripts/          # Build and automation scripts
├── Makefile         # Build automation
└── README.md       # Project overview
```

### Key Design Principles

- **Separation of Concerns**: Commands in `cmd/` handle user input, business logic in `internal/`
- **Testability**: All business logic is unit testable with clear interfaces
- **User Experience**: Consistent CLI patterns, helpful error messages, interactive prompts
- **Reliability**: Robust error handling and recovery mechanisms

## Testing Guidelines

### Running Tests

```bash
# Run unit tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires Docker)
make test-integration

# Run specific test packages
go test ./internal/cluster/...
```

### Writing Tests

- **Unit Tests**: Test individual functions and methods in isolation
- **Integration Tests**: Test complete workflows end-to-end
- **Table-Driven Tests**: Use for testing multiple scenarios

Example unit test:

```go
func TestClusterValidation(t *testing.T) {
    tests := []struct {
        name        string
        clusterName string
        wantErr     bool
    }{
        {"valid name", "my-cluster", false},
        {"empty name", "", true},
        {"invalid chars", "my_cluster!", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateClusterName(tt.clusterName)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateClusterName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing Best Practices

- Mock external dependencies (Docker, Kubernetes APIs)
- Use table-driven tests for comprehensive coverage
- Test both success and error scenarios
- Include integration tests for critical workflows
- Maintain test data in `test/fixtures/`

## Submitting Changes

### Pull Request Process

1. **Ensure Your Branch is Up to Date**

```bash
git checkout main
git pull upstream main
git checkout feature/your-feature-name
git rebase main
```

2. **Create a Pull Request**

- Push your branch to your fork
- Open a Pull Request against the `main` branch
- Use the provided PR template
- Include a clear description of changes
- Link any related issues

3. **PR Review Process**

- All PRs require at least one approval
- Address review feedback promptly
- Keep PRs focused and reasonably sized
- Ensure CI checks pass before merging

### PR Title Format

Use conventional commit format for PR titles:

- `feat: description` - New features
- `fix: description` - Bug fixes
- `docs: description` - Documentation
- `refactor: description` - Code refactoring
- `test: description` - Test improvements
- `chore: description` - Maintenance tasks

## Documentation

### Updating Documentation

When making changes, update relevant documentation:

- **Code Comments**: Document complex logic and public APIs
- **User Guides**: Update tutorials for user-facing changes
- **Architecture Docs**: Update technical documentation for structural changes
- **README**: Update feature lists and examples as needed

### Documentation Standards

- Use clear, concise language
- Include code examples where helpful
- Follow markdown best practices
- Test all command examples before submitting

## Release Process

### Versioning

OpenFrame CLI follows [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### Release Workflow

Releases are automated through GitHub Actions when tags are created:

1. **Prepare Release**
   - Update version in relevant files
   - Update CHANGELOG.md
   - Create release PR

2. **Create Release Tag**
   ```bash
   git tag v1.2.3
   git push upstream v1.2.3
   ```

3. **Automated Release**
   - GitHub Actions builds binaries for multiple platforms
   - Creates GitHub release with assets
   - Updates package repositories

## Getting Help

### Development Questions

- **Architecture Questions**: Check `docs/dev/overview.md` or ask in issues
- **Implementation Help**: Review existing code patterns in similar features
- **Testing Guidance**: Look at existing tests for patterns and examples

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests, questions
- **Pull Requests**: Code review and discussion
- **GitHub Discussions**: General questions and community chat

### Debugging Tips

```bash
# Enable verbose logging
openframe cluster create --verbose

# Use dry-run to see what would happen
openframe cluster create --dry-run

# Check detailed status
openframe cluster status --detailed

# Access cluster directly
kubectl get pods --all-namespaces
```

## Recognition

Contributors will be recognized in:

- GitHub contributors list
- Release notes for significant contributions
- Special recognition for ongoing contributions

## Questions?

Don't hesitate to ask questions! Whether you're new to Go, Kubernetes, or open source development, we're here to help. The best way to get started is to:

1. Read through this guide
2. Look at existing issues labeled `good first issue`
3. Set up your development environment
4. Make a small contribution to get familiar with the workflow

Thank you for contributing to OpenFrame CLI! 🚀