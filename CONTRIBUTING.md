# Contributing to OpenFrame CLI

Thank you for your interest in contributing to OpenFrame CLI! This guide will help you get started with contributing to our Go-based command-line tool for Kubernetes cluster management and MSP development workflows.

## 🌟 Welcome Contributors

OpenFrame CLI is built by the community, for the community. Whether you're fixing bugs, adding features, improving documentation, or helping with testing, every contribution makes a difference.

## 📋 Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contribution Workflow](#contribution-workflow)
- [Code Standards](#code-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have:

- **System Requirements**:
  - Minimum: 24GB RAM, 6 CPU cores, 50GB disk
  - Recommended: 32GB RAM, 12 CPU cores, 100GB disk
- **Development Tools**:
  - Go 1.21+ installed and configured
  - Docker installed and running
  - Git for version control
  - Your favorite IDE or editor

### Quick Contributor Setup

```bash
# 1. Fork the repository on GitHub
# 2. Clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# 3. Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git

# 4. Install dependencies
go mod download
go mod verify

# 5. Build from source
go build -o bin/openframe main.go

# 6. Verify your setup
./bin/openframe --version
go test ./... -short
```

## 🔧 Development Setup

### Local Development Environment

Follow our comprehensive [Local Development Guide](./docs/development/setup/local-development.md) for detailed setup instructions.

**Quick Setup:**

```bash
# Install build dependencies
make install-deps

# Development build
make build

# Run tests
make test

# Start development with hot reload (optional)
make dev
```

### Repository Structure

Understanding the codebase:

```text
openframe-cli/
├── cmd/                    # CLI command definitions (Cobra)
│   ├── bootstrap/         # Bootstrap command
│   ├── cluster/           # Cluster management
│   ├── chart/             # Chart installation
│   ├── dev/               # Development tools
│   └── root.go            # Root command setup
├── internal/              # Internal packages
│   ├── bootstrap/         # Bootstrap service logic
│   ├── cluster/           # K3D cluster management
│   ├── chart/             # ArgoCD/Helm services
│   ├── dev/               # Development workflows
│   └── shared/            # Shared utilities
├── tests/                 # Test suites
│   ├── integration/       # End-to-end tests
│   ├── mocks/            # Generated mocks
│   └── testutil/         # Test utilities
├── docs/                 # Documentation
└── main.go               # Application entry point
```

## 🔄 Contribution Workflow

### 1. Find or Create an Issue

- **Existing Issues**: Browse our Slack community for open discussions
- **New Features**: Propose ideas in our OpenMSP Slack community first
- **Bug Reports**: Join our Slack to report and discuss issues

**🔗 Join our community**: https://www.openmsp.ai/

### 2. Fork and Branch

```bash
# Fork the repository on GitHub
# Clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### 3. Development Process

```bash
# Make your changes
# ... edit code ...

# Test your changes
go test ./...
go test -race ./...

# Build and test manually
go build -o bin/openframe main.go
./bin/openframe --version

# Run integration tests (if applicable)
export OPENFRAME_INTEGRATION_TESTS="true"
go test ./tests/integration/... -timeout=10m
```

### 4. Commit Guidelines

We follow conventional commit standards:

```bash
# Feature commits
git commit -m "feat: add cluster auto-scaling support"

# Bug fix commits
git commit -m "fix: resolve K3D networking issue"

# Documentation commits
git commit -m "docs: update installation guide"

# Test commits
git commit -m "test: add integration tests for bootstrap command"
```

**Commit Types:**
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `test`: Test additions or modifications
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

### 5. Submit Pull Request

```bash
# Push to your fork
git push origin feature/your-feature-name

# Create Pull Request on GitHub with:
# - Clear title and description
# - Reference to related issues/discussions
# - Screenshots/videos if applicable
# - Test results and verification steps
```

## 📏 Code Standards

### Go Code Guidelines

**1. Follow Go Best Practices:**
```go
// Use meaningful names
func CreateClusterWithConfig(name string, config *ClusterConfig) error {
    // Implementation
}

// Handle errors properly
if err := cluster.Create(); err != nil {
    return fmt.Errorf("failed to create cluster: %w", err)
}

// Use context for cancellation
func (s *Service) CreateCluster(ctx context.Context, name string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue with creation
    }
}
```

**2. Package Organization:**
```go
// Package declaration with clear purpose
// Package cluster provides Kubernetes cluster management functionality.
package cluster

// Imports organized: std lib, external, internal
import (
    "context"
    "fmt"
    
    "github.com/spf13/cobra"
    
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)
```

**3. Error Handling:**
```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to validate cluster config: %w", err)
}

// Use custom error types for specific cases
type ClusterNotFoundError struct {
    Name string
}

func (e *ClusterNotFoundError) Error() string {
    return fmt.Sprintf("cluster %q not found", e.Name)
}
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Organize imports
goimports -w .

# Run linter
golangci-lint run

# Check for common issues
go vet ./...
```

### CLI Command Standards

**Command Structure:**
```go
func NewCreateCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create [NAME]",
        Short: "Create a new OpenFrame cluster",
        Long: `Create a new OpenFrame cluster with the specified configuration.
        
This command will create a K3D cluster with networking and certificates
configured for OpenFrame development.`,
        Args: cobra.ExactArgs(1),
        RunE: runCreate,
    }
    
    // Add flags with clear descriptions
    cmd.Flags().StringSlice("nodes", []string{}, "Number of worker nodes")
    cmd.Flags().String("version", "latest", "Kubernetes version")
    
    return cmd
}
```

**Interactive UI Standards:**
```go
// Use consistent UI components
import "github.com/flamingo-stack/openframe-cli/internal/shared/ui"

// Progress indication
ui.Info("Creating cluster", "name", clusterName)
spinner := ui.NewSpinner("Installing components...")
spinner.Start()
defer spinner.Stop()

// Success/error reporting
ui.Success("Cluster created successfully", "name", clusterName)
ui.Error("Failed to create cluster", "error", err)
```

## 🧪 Testing Guidelines

### Test Categories

**1. Unit Tests:**
```go
func TestCreateCluster(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    error
        wantErr bool
    }{
        {
            name:    "valid cluster name",
            input:   "test-cluster",
            want:    nil,
            wantErr: false,
        },
        {
            name:    "invalid cluster name",
            input:   "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateCluster(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("CreateCluster() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**2. Integration Tests:**
```go
func TestBootstrapIntegration(t *testing.T) {
    if !testing.Short() && os.Getenv("OPENFRAME_INTEGRATION_TESTS") == "true" {
        t.Skip("Skipping integration tests")
    }
    
    // Test full bootstrap workflow
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    err := bootstrap.Execute(ctx, &bootstrap.Config{
        ClusterName: "test-integration",
        NonInteractive: true,
    })
    
    require.NoError(t, err)
    
    // Cleanup
    defer cluster.Delete("test-integration")
}
```

### Running Tests

```bash
# Run unit tests
go test ./internal/... -v

# Run with race detection
go test -race ./...

# Run integration tests (requires Docker)
export OPENFRAME_INTEGRATION_TESTS="true"
go test ./tests/integration/... -timeout=15m

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Requirements

- **Unit test coverage**: Aim for >80% coverage on new code
- **Integration tests**: Add for new commands and workflows
- **Error scenarios**: Test error paths and edge cases
- **Mocking**: Use mocks for external dependencies

## 📚 Documentation

### Code Documentation

```go
// Package cluster provides Kubernetes cluster lifecycle management.
//
// This package handles creating, managing, and destroying K3D clusters
// for OpenFrame development environments.
package cluster

// ClusterService manages Kubernetes cluster operations.
type ClusterService struct {
    provider ClusterProvider
    ui       ui.Interface
}

// Create creates a new Kubernetes cluster with the specified configuration.
//
// The cluster will be configured with networking, certificates, and
// storage required for OpenFrame services.
//
// Example:
//   service := cluster.NewService(provider, ui)
//   err := service.Create(ctx, "my-cluster", config)
func (s *ClusterService) Create(ctx context.Context, name string, config *Config) error {
    // Implementation
}
```

### Documentation Updates

- Update relevant documentation in `docs/` directory
- Include examples and use cases
- Add troubleshooting information
- Update CLI help text and descriptions

## 🎯 Contribution Checklist

Before submitting your contribution:

### Code Quality
- [ ] **Code builds successfully**: `go build main.go`
- [ ] **Tests pass**: `go test ./...`
- [ ] **Linting passes**: `golangci-lint run`
- [ ] **Code formatted**: `go fmt ./...`
- [ ] **Dependencies updated**: `go mod tidy`

### Testing
- [ ] **Unit tests written** for new functionality
- [ ] **Integration tests added** (if applicable)
- [ ] **Manual testing completed** on target platforms
- [ ] **Error scenarios tested**

### Documentation
- [ ] **Code documented** with clear comments
- [ ] **CLI help updated** for new commands
- [ ] **User documentation updated** in `docs/`
- [ ] **README updated** (if needed)

### Review Preparation
- [ ] **Commit messages follow conventions**
- [ ] **Branch is up-to-date** with upstream main
- [ ] **PR description is comprehensive**
- [ ] **Related issues referenced**

## 🏗️ Architecture Contributions

### Adding New Commands

1. Create command in `cmd/` directory
2. Implement service logic in `internal/`
3. Add provider interfaces and implementations
4. Include comprehensive tests
5. Update documentation

### Service Layer Patterns

```go
// Service interface pattern
type Service interface {
    Execute(ctx context.Context, config *Config) error
}

// Provider interface pattern  
type Provider interface {
    Create(ctx context.Context, name string) error
    Delete(ctx context.Context, name string) error
}

// Implementation with dependency injection
func NewService(provider Provider, ui ui.Interface) *ServiceImpl {
    return &ServiceImpl{
        provider: provider,
        ui:       ui,
    }
}
```

## 🤝 Community

### Communication Channels

- **Primary Community**: OpenMSP Slack Community
  - Join: https://www.openmsp.ai/
  - Invite Link: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA
- **GitHub**: For code reviews and pull requests
- **Documentation**: In-repo docs for technical reference

### Getting Help

1. **Development Questions**: Ask in #openframe-dev channel
2. **Feature Discussions**: Use #feature-requests channel
3. **Bug Reports**: Report in #bug-reports channel
4. **General Support**: Use #general channel

### Code Review Process

1. **Automated Checks**: CI runs tests and linting
2. **Peer Review**: Community members review code
3. **Maintainer Review**: Core team provides final review
4. **Merge**: Approved changes are merged to main

## 🎉 Recognition

We appreciate all contributors! Contributors will be:

- Listed in project contributors
- Recognized in release notes
- Invited to contributor channels
- Eligible for contributor swag (when available)

## 📄 License

By contributing to OpenFrame CLI, you agree that your contributions will be licensed under the [Flamingo AI Unified License v1.0](LICENSE.md).

## 🚀 Next Steps

1. **Join the Community**: https://www.openmsp.ai/
2. **Set Up Development**: Follow the [Local Development Guide](./docs/development/setup/local-development.md)
3. **Pick Your First Issue**: Ask in Slack for good first contribution ideas
4. **Start Contributing**: Follow this guide and submit your first PR!

---

**Thank you for contributing to OpenFrame CLI!** 🙏

Your contributions help make MSP development more accessible and efficient for the entire community.

*For questions about this guide, reach out in our Slack community.*