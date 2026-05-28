# Contributing to OpenFrame CLI

Welcome to the OpenFrame CLI community! We're excited to have you contribute to our mission of simplifying Kubernetes cluster management and GitOps workflows for MSP environments.

## 🚀 Quick Start for Contributors

### Prerequisites for Development

- **Go 1.21+** - [Install Go](https://golang.org/doc/install)
- **Docker 20.10+** - [Get Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)
- **Make** - For build automation

**Hardware Requirements:**
- Minimum: 24GB RAM, 6 CPU cores, 50GB disk
- Recommended: 32GB RAM, 12 CPU cores, 100GB disk

### Development Setup

```bash
# 1. Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# 2. Install dependencies
go mod download
go mod tidy

# 3. Build the CLI
go build -o openframe ./main.go

# 4. Run tests
go test ./...

# 5. Verify your setup
./openframe --help
```

## 🏗️ Development Workflow

### Setting Up Your Development Environment

```bash
# Set up development environment variables
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug

# Configure Git (if not already done)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test locally**:
   ```bash
   # Run unit tests
   go test ./...

   # Run with race detection
   go test -race ./...

   # Test the CLI functionality
   ./openframe bootstrap test-cluster --verbose
   ```

4. **Lint your code**:
   ```bash
   golangci-lint run
   ```

5. **Commit and push**:
   ```bash
   git add .
   git commit -m "feat(component): add new feature

   - Detailed description of the change
   - Why this change is needed
   - Any breaking changes

   Closes #123"
   git push origin feature/your-feature-name
   ```

## 📝 Contribution Guidelines

### Code Standards

- **Test Coverage**: Minimum 80% for new code
- **Linting**: All code must pass `golangci-lint`
- **Documentation**: Public APIs must have Go doc comments
- **Error Handling**: All errors must be properly handled and meaningful
- **Logging**: Use structured logging with appropriate levels

### Commit Message Format

Follow the conventional commit format:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix  
- `docs` - Documentation changes
- `style` - Formatting changes
- `refactor` - Code refactoring
- `test` - Adding tests
- `chore` - Maintenance tasks

**Examples:**
```
feat(cluster): add multi-node support

- Add node count configuration to cluster creation
- Update cluster status to show all nodes
- Add validation for node resource requirements

Closes #123
```

### Pull Request Process

1. **Ensure all tests pass** and code follows standards
2. **Update documentation** if needed
3. **Add/update tests** for new functionality
4. **Create detailed PR description** explaining:
   - What changes were made
   - Why they were necessary
   - How to test the changes
   - Any breaking changes
5. **Request review** from maintainers
6. **Address feedback** promptly

### PR Review Checklist

Before submitting, ensure your PR:

- [ ] **Builds successfully** on all supported platforms
- [ ] **Tests pass** with `go test ./...`
- [ ] **Linting passes** with `golangci-lint run`
- [ ] **Documentation** is updated (if applicable)
- [ ] **No breaking changes** (or clearly documented)
- [ ] **Commit messages** follow conventional format
- [ ] **PR description** explains the change and rationale

## 🧪 Testing

### Running Tests

```bash
# Unit tests
go test ./...

# Tests with verbose output
go test -v ./...

# Tests with race detection
go test -race ./...

# Tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Testing

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./...

# Test specific components
go test ./internal/bootstrap/...
go test ./internal/cluster/...
```

### Writing Tests

- **Unit tests** for all business logic in `internal/` packages
- **Integration tests** for end-to-end command workflows
- **Mock external dependencies** (Docker, K3d, kubectl)
- **Test error conditions** and edge cases
- **Use table-driven tests** for multiple scenarios

Example test structure:
```go
func TestClusterCreate(t *testing.T) {
    tests := []struct {
        name    string
        config  ClusterConfig
        want    error
    }{
        {
            name: "valid config",
            config: ClusterConfig{Name: "test", Nodes: 1},
            want: nil,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CreateCluster(tt.config)
            if got != tt.want {
                t.Errorf("CreateCluster() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## 🏛️ Architecture Guidelines

### Project Structure

```text
openframe-cli/
├── cmd/                    # CLI command implementations
│   ├── bootstrap/         # Complete environment setup
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Helm chart management
│   └── dev/               # Development workflow tools
├── internal/              # Internal packages
│   ├── bootstrap/         # Bootstrap service implementation
│   ├── cluster/           # Cluster service implementation
│   ├── chart/             # Chart service implementation
│   └── utils/             # Utility functions
├── pkg/                   # Public packages
└── main.go                # CLI entry point
```

### Design Principles

- **Modularity**: Each command is self-contained
- **User Experience**: Interactive wizards and helpful error messages
- **Safety**: Confirmation prompts and safe defaults
- **Extensibility**: Plugin-friendly architecture
- **GitOps Native**: Built-in ArgoCD integration

### Code Organization

- Keep command implementations in `cmd/` directories **simple**
- Put business logic in `internal/` packages
- Use `pkg/` for reusable public packages
- Write **testable code** with dependency injection
- Use interfaces for external dependencies

## 🐛 Reporting Issues

### Bug Reports

When reporting bugs, include:

1. **OpenFrame CLI version**: `openframe --version`
2. **Operating system** and version
3. **Go version**: `go version`
4. **Docker version**: `docker --version`
5. **Steps to reproduce** the issue
6. **Expected vs actual behavior**
7. **Log output** with `--verbose` flag
8. **Error messages** (full stack trace if available)

### Feature Requests

For feature requests, describe:

1. **Use case** - What problem does this solve?
2. **Proposed solution** - How should it work?
3. **Alternatives considered** - What else did you consider?
4. **Additional context** - Mockups, examples, etc.

## 🔐 Security

### Security Policy

- Report security vulnerabilities privately to our security team
- Follow OWASP guidelines for secure coding
- Never commit secrets, API keys, or credentials
- Use secure defaults for all configurations
- Validate all user inputs

### Security Considerations

- **Input validation** for all user-provided data
- **Secure file handling** for configuration files
- **Safe subprocess execution** for external tools
- **Proper error handling** that doesn't leak sensitive information

## 🌍 Community

### Communication Channels

- **OpenMSP Slack**: [Join our community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and ideas

### Community Guidelines

- Be respectful and inclusive
- Help others learn and contribute
- Share knowledge and best practices
- Provide constructive feedback
- Follow our code of conduct

### Getting Help

- Check existing documentation first
- Search issues and discussions
- Ask questions in our Slack community
- Provide context and details when asking for help

## 📚 Additional Resources

### External Documentation
- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://cobra.dev/)
- [K3d Documentation](https://k3d.io/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

### Development Guides
- [Local Development Setup](./docs/development/setup/local-development.md)
- [Environment Configuration](./docs/development/setup/environment.md)
- [Architecture Overview](./docs/development/architecture/README.md)
- [Security Guidelines](./docs/development/security/README.md)

## 🎉 Recognition

We appreciate all contributions! Contributors will be:

- **Listed** in our README and release notes
- **Invited** to special community events
- **Given credit** in our documentation
- **Welcomed** to our contributor program

Thank you for contributing to OpenFrame CLI and helping make Kubernetes development easier for everyone!

---

> 💡 **Questions?** Join our [OpenMSP Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) - we're here to help!