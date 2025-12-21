# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide will help you get started with contributing to the project.

## 🤝 Code of Conduct

We are committed to fostering a welcoming and inclusive community. Please review and follow our code of conduct in all interactions.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Respecting differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- Harassment, trolling, or discriminatory language
- Personal attacks or inflammatory comments
- Publishing others' private information without consent
- Any conduct that could reasonably be considered inappropriate

## 🚀 Getting Started

### Prerequisites

Before contributing, ensure you have:

- **Go 1.21+** installed and configured
- **Docker** running for local K3d clusters  
- **kubectl** configured for Kubernetes access
- **Make** for build automation
- **Git** for version control

### Development Environment Setup

1. **Fork and clone the repository:**
```bash
git clone https://github.com/YOUR-USERNAME/openframe-cli.git
cd openframe-cli
```

2. **Install dependencies:**
```bash
go mod download
```

3. **Verify your setup:**
```bash
# Run tests to ensure everything works
make test

# Build the CLI
make build

# Verify the binary works
./bin/openframe --help
```

4. **Set up pre-commit hooks (recommended):**
```bash
# Install pre-commit tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linting
make lint
```

## 📋 Development Workflow

### 1. Create a Branch

```bash
# Create and checkout a feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 2. Make Changes

- Follow the existing code patterns and architecture
- Write clear, concise commit messages
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
make test

# Run specific tests
go test ./cmd/cluster/...

# Test with a real cluster (optional)
./bin/openframe bootstrap test-cluster
./bin/openframe cluster delete test-cluster
```

### 4. Submit a Pull Request

- Push your branch to your fork
- Open a Pull Request with a clear title and description
- Link any related issues
- Ensure all CI checks pass

## 🏗️ Project Structure

Understanding the codebase organization:

```text
openframe-cli/
├── cmd/                    # Command definitions (Cobra)
│   ├── bootstrap/         # Bootstrap command implementation
│   ├── cluster/           # Cluster management commands
│   ├── chart/            # Chart installation commands
│   └── dev/              # Development workflow commands
├── internal/             # Internal packages
│   ├── bootstrap/        # Bootstrap business logic
│   ├── cluster/          # Cluster services and models
│   │   ├── services/     # Core cluster operations
│   │   ├── ui/           # User interface components
│   │   └── prerequisites/ # Dependency validation
│   ├── chart/           # Chart installation services
│   ├── dev/             # Development tool integrations
│   └── shared/          # Common utilities and UI components
├── docs/                # Documentation
├── tests/               # Test files and fixtures  
└── Makefile            # Build automation
```

### Key Principles

- **Modular Design**: Each command group (`cluster`, `chart`, `dev`) is self-contained
- **Service Layer**: Business logic is separated from command interfaces
- **UI Components**: Reusable components for consistent user experience  
- **Error Handling**: Comprehensive error handling with helpful messages

## 📝 Coding Standards

### Go Code Style

**Follow standard Go conventions:**

```go
// ✅ Good: Clear function naming and documentation
// CreateCluster creates a new K3d cluster with the specified configuration
func CreateCluster(ctx context.Context, config ClusterConfig) error {
    if err := validateConfig(config); err != nil {
        return fmt.Errorf("invalid cluster configuration: %w", err)
    }
    // ... implementation
}

// ✅ Good: Proper error wrapping
if err := k3d.CreateCluster(config); err != nil {
    return fmt.Errorf("failed to create K3d cluster: %w", err)
}
```

**Use consistent patterns:**

```go
// ✅ Good: Consistent service pattern
type ClusterService struct {
    k3dClient K3dClient
    ui        UIService
    logger    Logger
}

func (s *ClusterService) Create(ctx context.Context, config ClusterConfig) error {
    // Service implementation
}
```

### Command Structure

**Follow Cobra conventions:**

```go
// ✅ Good: Clear command definition
var createCmd = &cobra.Command{
    Use:   "create [cluster-name]",
    Short: "Create a new K3d cluster",
    Long:  `Create a new K3d cluster with interactive configuration...`,
    Args:  cobra.MaximumNArgs(1),
    RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
    // Command implementation
}
```

### Error Messages

**Provide helpful, actionable error messages:**

```go
// ✅ Good: Helpful error with solution
if !docker.IsRunning() {
    return fmt.Errorf("Docker is not running. Please start Docker and try again.\n" +
        "💡 Tip: Run 'docker info' to check Docker status")
}

// ✅ Good: Context-aware error
if err := validateClusterName(name); err != nil {
    return fmt.Errorf("invalid cluster name %q: %w\n" +
        "💡 Cluster names must be lowercase alphanumeric with hyphens", name, err)
}
```

### Testing

**Write comprehensive tests:**

```go
// ✅ Good: Table-driven tests
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        wantErr  bool
    }{
        {"valid name", "my-cluster", false},
        {"uppercase invalid", "My-Cluster", true},
        {"empty invalid", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateClusterName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateClusterName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🧪 Testing Guidelines

### Test Structure

```text
tests/
├── unit/              # Unit tests for individual components
├── integration/       # Integration tests with real dependencies  
├── e2e/              # End-to-end tests with full workflows
└── fixtures/         # Test data and configuration files
```

### Running Tests

```bash
# Unit tests (fast)
make test-unit

# Integration tests (requires Docker)
make test-integration  

# All tests
make test

# Test with coverage
make test-coverage
```

### Writing Tests

**Test different scenarios:**

```go
func TestBootstrapCommand(t *testing.T) {
    tests := []struct {
        name           string
        args           []string
        flags          map[string]string
        wantErr        bool
        expectedCalls  []string
    }{
        {
            name: "successful bootstrap",
            args: []string{"test-cluster"},
            expectedCalls: []string{"CreateCluster", "InstallCharts"},
        },
        {
            name:    "missing cluster name",
            args:    []string{},
            wantErr: true,
        },
    }
    // ... test implementation
}
```

## 📖 Documentation

### Code Documentation

**Document public functions and types:**

```go
// ClusterConfig defines the configuration for creating a K3d cluster.
// It includes network settings, resource limits, and integration options.
type ClusterConfig struct {
    // Name is the cluster identifier (must be DNS-compatible)
    Name string `json:"name"`
    
    // Ports defines port mappings between host and cluster
    Ports []PortMapping `json:"ports,omitempty"`
    
    // RegistryPort configures the local container registry port
    RegistryPort int `json:"registryPort,omitempty"`
}

// CreateCluster creates a new K3d cluster with the specified configuration.
// It performs prerequisite checks, validates configuration, and creates
// the cluster with proper networking and registry setup.
func CreateCluster(ctx context.Context, config ClusterConfig) error {
    // ... implementation
}
```

### Updating Documentation

When making changes that affect user experience:

1. Update relevant markdown files in `docs/`
2. Add examples for new commands or flags
3. Update the main README if adding major features
4. Include any architectural changes in `docs/reference/architecture/`

## 🔍 Pull Request Process

### Before Submitting

**Checklist:**
- [ ] Code follows project conventions and style
- [ ] Tests pass locally (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] Documentation updated if needed
- [ ] Commit messages are clear and descriptive
- [ ] No sensitive information (keys, passwords) in commits

### Pull Request Template

**Title Format:**
- `feat: add cluster cleanup command`
- `fix: handle k3d registry creation error`
- `docs: update bootstrap command examples`

**Description should include:**
- Summary of changes
- Motivation and context  
- Testing performed
- Screenshots/examples if UI changes
- Related issues (fixes #123)

### Review Process

1. **Automated Checks**: All CI checks must pass
2. **Code Review**: At least one maintainer approval required
3. **Testing**: Reviewers may test changes locally
4. **Documentation**: Verify documentation is accurate and complete

## 🏷️ Commit Guidelines

### Commit Message Format

```text
<type>: <description>

[optional body]

[optional footer]
```

### Types

- **feat**: New feature
- **fix**: Bug fix  
- **docs**: Documentation changes
- **style**: Code style changes (formatting, etc.)
- **refactor**: Code refactoring
- **test**: Adding or updating tests
- **chore**: Maintenance tasks

### Examples

```text
feat: add cluster cleanup command

Add new 'cleanup' subcommand to remove unused cluster resources
and free disk space. Includes interactive confirmation and 
dry-run mode.

Closes #45

fix: handle k3d registry creation error

Properly handle cases where K3d registry creation fails due to
port conflicts. Now suggests alternative ports and retries.

docs: update bootstrap command examples

Add examples for CI/CD usage and non-interactive mode to the
bootstrap command documentation.
```

## 🐛 Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **OpenFrame CLI version**: `openframe --version`
- **Operating system**: OS and version
- **Go version**: `go version`  
- **Docker version**: `docker --version`
- **Steps to reproduce**: Exact commands that trigger the issue
- **Expected behavior**: What should have happened
- **Actual behavior**: What actually happened  
- **Logs**: Any error messages or relevant log output
- **Configuration**: Any relevant configuration files

### Feature Requests

For feature requests, please describe:

- **Use case**: What problem does this solve?
- **Proposed solution**: How should this feature work?
- **Alternatives considered**: Other approaches you've considered
- **Additional context**: Any other relevant information

## 💡 Getting Help

### Questions and Discussions

- **GitHub Discussions**: For questions about usage or development
- **GitHub Issues**: For bug reports and feature requests
- **Code Comments**: For specific implementation questions

### Resources

- **[Development Documentation](./docs/development/README.md)** - Comprehensive development guides
- **[Architecture Overview](./docs/reference/architecture/overview.md)** - Technical architecture details
- **[Go Documentation](https://golang.org/doc/)** - Go language reference
- **[Cobra Guide](https://github.com/spf13/cobra)** - CLI framework documentation

## 🎉 Recognition

Contributors are recognized in several ways:

- **Contributor List**: All contributors are listed in the project documentation
- **Release Notes**: Significant contributions are highlighted in release notes  
- **Community Appreciation**: Regular shout-outs for helpful contributions

We appreciate all forms of contribution, from code and documentation to bug reports and feature suggestions!

---

Thank you for contributing to OpenFrame CLI! Together, we're building better tools for Kubernetes development and deployment. 🚀