# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide will help you get started with contributing to our modern Kubernetes cluster management tool.

## 🚀 Quick Start for Contributors

1. **Fork and Clone**
   ```bash
   git clone https://github.com/your-username/openframe-cli.git
   cd openframe-cli
   ```

2. **Set Up Development Environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Build the project
   make build
   
   # Run tests
   make test
   ```

3. **Make Your Changes**
   ```bash
   # Create a feature branch
   git checkout -b feature/your-feature-name
   
   # Make your changes and commit
   git commit -m "feat: add your feature description"
   ```

4. **Submit a Pull Request**
   ```bash
   # Push your branch
   git push origin feature/your-feature-name
   
   # Open a PR on GitHub
   ```

## 📋 Development Setup

### Prerequisites

Ensure you have the following installed:

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Docker** - Required for k3d cluster testing
- **Git** - For version control
- **Make** - For build automation (Windows users: use WSL or Git Bash)

### Environment Setup

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Install dependencies
go mod download

# Build the project
make build

# Verify the build
./bin/openframe --version

# Run tests
make test

# Run linting
make lint
```

### Development Workflow

```bash
# Run from source during development
go run . cluster create --help

# Build and test changes
make build && ./bin/openframe cluster list

# Run with verbose logging for debugging
go run . --verbose cluster create my-test-cluster
```

## 🏗️ Project Structure

Understanding the codebase structure will help you navigate and contribute effectively:

```
openframe-cli/
├── cmd/                    # Command definitions (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   └── dev/               # Development workflow commands
├── internal/              # Internal packages
│   ├── cluster/           # Cluster management logic
│   ├── chart/             # Chart and ArgoCD logic
│   ├── dev/               # Development tools logic
│   ├── bootstrap/         # Combined cluster + chart setup
│   └── shared/            # Shared utilities
│       ├── executor/      # Command execution abstraction
│       ├── ui/            # Terminal UI components
│       ├── config/        # Configuration management
│       └── errors/        # Error handling
├── docs/                  # Documentation
├── scripts/               # Build and release scripts
└── tests/                 # Test files
```

### Key Components

- **Commands** (`cmd/`): CLI interface using Cobra framework
- **Services** (`internal/*/services/`): Core business logic
- **Providers** (`internal/*/providers/`): External tool integrations (k3d, kubectl, etc.)
- **UI** (`internal/*/ui/`): Interactive prompts and progress displays
- **Models** (`internal/*/models/`): Data structures and types

## 🎯 How to Contribute

### Types of Contributions

We welcome various types of contributions:

1. **🐛 Bug Fixes**
   - Fix reported issues
   - Improve error handling
   - Enhance edge case handling

2. **✨ New Features**
   - Add new CLI commands
   - Integrate additional tools
   - Improve user experience

3. **📚 Documentation**
   - Improve README and guides
   - Add code comments
   - Create examples and tutorials

4. **🧪 Testing**
   - Add unit tests
   - Create integration tests
   - Improve test coverage

5. **🔧 Infrastructure**
   - Improve build processes
   - Enhance CI/CD pipelines
   - Optimize performance

### Finding Work

- **Good First Issues**: Look for issues labeled `good first issue`
- **Help Wanted**: Check issues labeled `help wanted`
- **Documentation**: Improve docs or add missing documentation
- **Testing**: Add tests for untested code paths
- **Bug Reports**: Fix reported bugs in GitHub Issues

### Reporting Issues

When reporting bugs:

1. **Search existing issues** to avoid duplicates
2. **Use the bug report template** provided
3. **Include system information**: OS, Go version, Docker version
4. **Provide reproduction steps** with commands and expected vs actual behavior
5. **Include logs** with `--verbose` flag output

## 🔧 Development Guidelines

### Code Style

We follow Go best practices and conventions:

- **Go formatting**: Use `gofmt` and `goimports`
- **Linting**: Pass `golangci-lint` checks
- **Naming**: Use descriptive names following Go conventions
- **Error handling**: Use structured errors from `internal/shared/errors/`
- **Documentation**: Add comments for exported functions and types

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

[optional body]

[optional footer]
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding tests
- `refactor`: Code refactoring
- `chore`: Build/tool changes

**Examples**:
```
feat(cluster): add support for custom node labels
fix(chart): resolve ArgoCD installation timeout issue
docs(readme): update installation instructions
test(cluster): add unit tests for k3d provider
```

### Testing Requirements

All contributions must include appropriate tests:

1. **Unit Tests**
   ```bash
   # Run unit tests
   go test ./...
   
   # Run tests with coverage
   go test -cover ./...
   
   # Run tests for specific package
   go test ./internal/cluster/services/
   ```

2. **Integration Tests**
   ```bash
   # Run integration tests (requires Docker)
   make integration-test
   ```

3. **Test Guidelines**
   - Test both success and error paths
   - Mock external dependencies
   - Use table-driven tests for multiple scenarios
   - Keep tests fast and focused
   - Aim for >80% code coverage on new code

### Pull Request Process

1. **Before Starting**
   - Check if there's an existing issue
   - Comment on the issue to claim it
   - Discuss approach for large changes

2. **Development**
   - Create a feature branch from `main`
   - Make small, focused commits
   - Write tests for new functionality
   - Update documentation as needed

3. **Before Submitting**
   - Run all tests: `make test`
   - Run linting: `make lint`
   - Ensure builds work: `make build`
   - Update documentation if needed

4. **Pull Request**
   - Use the PR template provided
   - Write a clear title and description
   - Link to related issues
   - Add screenshots for UI changes
   - Mark as draft if work in progress

5. **Review Process**
   - Address feedback promptly
   - Keep discussions focused and respectful
   - Update tests based on review feedback
   - Rebase if requested to keep history clean

### Code Review Guidelines

**For Reviewers**:
- Be constructive and respectful
- Focus on code quality and maintainability
- Check for proper testing
- Verify documentation updates
- Test functionality locally when needed

**For Contributors**:
- Respond to feedback promptly
- Ask questions if feedback is unclear
- Make requested changes or discuss alternatives
- Be open to learning and improving

## 🧪 Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests (requires Docker)
make integration-test

# Run specific test
go test ./internal/cluster/services/ -v

# Run tests with race detection
go test -race ./...
```

### Test Structure

```
tests/
├── unit/              # Unit tests (fast, isolated)
├── integration/       # Integration tests (slower, real dependencies)
└── fixtures/          # Test data and fixtures
```

### Writing Tests

```go
func TestClusterService_CreateCluster(t *testing.T) {
    tests := []struct {
        name    string
        config  ClusterConfig
        want    error
        setup   func(*testing.T)
    }{
        {
            name: "successful cluster creation",
            config: ClusterConfig{
                Name:  "test-cluster",
                Nodes: 3,
            },
            want: nil,
            setup: func(t *testing.T) {
                // Setup mocks or test environment
            },
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup(t)
            }
            // Test implementation
        })
    }
}
```

## 📚 Documentation Standards

### Documentation Requirements

- **README**: Keep main README up to date
- **Code Comments**: Add comments for exported functions
- **API Documentation**: Document public interfaces
- **User Guides**: Update user-facing documentation
- **Developer Docs**: Update architecture docs for significant changes

### Documentation Format

- Use Markdown for all documentation
- Include code examples with syntax highlighting
- Add diagrams using Mermaid for complex workflows
- Keep language clear and concise
- Test all code examples

## 🔄 Release Process

### Versioning

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Workflow

Releases are automated through GitHub Actions when tags are pushed:

```bash
# Create and push a tag
git tag v0.4.0
git push origin v0.4.0
```

This triggers:
- Cross-platform builds
- GitHub release creation
- Asset uploads
- Documentation updates

## 🌟 Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes
- Project documentation
- Community showcases

## 📞 Getting Help

### Development Support

- **GitHub Discussions**: Ask questions about development
- **Issues**: Report bugs or request features
- **Discord/Slack**: Real-time chat with maintainers
- **Email**: Contact maintainers directly for sensitive issues

### Resources

- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://cobra.dev/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [K3d Documentation](https://k3d.io/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the Flamingo AI Unified License v1.0.

---

Thank you for contributing to OpenFrame CLI! Together, we're making Kubernetes development more accessible and productive for everyone. 🚀