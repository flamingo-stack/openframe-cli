# Local Development Guide

This guide covers everything you need to know about developing OpenFrame CLI locally - from cloning the repository to running tests, debugging issues, and contributing changes.

## Quick Start

### Clone and Build
```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Install dependencies and build
make deps
make build

# Verify installation
./bin/openframe --version

# Run basic functionality test
./bin/openframe --help
```

### Run from Source
```bash
# Run directly with go run
go run . cluster --help

# Run with debug logging
OPENFRAME_DEBUG=true go run . --verbose cluster create test-cluster

# Build and run specific version
go build -o openframe .
./openframe cluster list
```

## Repository Structure

Understanding the codebase layout will help you navigate and contribute effectively:

```
openframe-cli/
├── cmd/                          # CLI command definitions
│   ├── root.go                   # Root command and CLI entry point
│   ├── cluster/                  # Cluster management commands
│   │   ├── cluster.go           # Cluster command group
│   │   ├── create.go            # Create cluster subcommand
│   │   ├── delete.go            # Delete cluster subcommand
│   │   └── ...
│   ├── chart/                    # Chart management commands
│   ├── dev/                      # Development tool commands
│   └── bootstrap/                # Bootstrap command
├── internal/                     # Internal packages (not importable)
│   ├── shared/                   # Shared utilities and components
│   │   ├── executor/            # Command execution abstraction
│   │   ├── ui/                  # User interface components
│   │   ├── errors/              # Error handling utilities
│   │   └── config/              # Configuration management
│   ├── cluster/                  # Cluster management business logic
│   │   ├── models/              # Data models and types
│   │   ├── services/            # Core business logic
│   │   ├── providers/           # External tool integrations
│   │   ├── ui/                  # Cluster-specific UI components
│   │   └── prerequisites/       # Prerequisite checking/installation
│   ├── chart/                    # Chart management business logic
│   ├── dev/                      # Development tools business logic
│   └── bootstrap/                # Bootstrap service logic
├── tests/                        # Test files and utilities
│   ├── integration/             # Integration tests
│   ├── mocks/                   # Mock implementations for testing
│   └── testutil/                # Test utilities and helpers
├── scripts/                      # Build and development scripts
├── docs/                        # Documentation
├── examples/                    # Usage examples and templates
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
└── main.go                      # Application entry point
```

### Key Architectural Patterns

| Pattern | Location | Purpose |
|---------|----------|---------|
| **Command Pattern** | `cmd/` | CLI command structure using Cobra |
| **Service Layer** | `internal/*/services/` | Business logic separation |
| **Provider Pattern** | `internal/*/providers/` | External tool integration |
| **Repository Pattern** | `internal/*/models/` | Data access abstraction |
| **Factory Pattern** | Throughout | Object creation and dependency injection |

## Development Workflow

### 1. Set Up Development Environment

```bash
# Set development environment variables
export OPENFRAME_DEBUG=true
export OPENFRAME_LOG_LEVEL=trace

# Create development cluster for testing
k3d cluster create openframe-dev --api-port 6443 --port "8080:80@loadbalancer"

# Set kubectl context
kubectl config use-context k3d-openframe-dev
```

### 2. Running and Testing Changes

#### Hot Reload Development
```bash
# Use air for hot reloading (install: go install github.com/cosmtrek/air@latest)
air

# Or use go run directly
go run . --verbose cluster create test-cluster --dry-run
```

#### Manual Testing
```bash
# Test cluster operations
go run . cluster create dev-test --nodes 1
go run . cluster list
go run . cluster status dev-test
go run . cluster delete dev-test

# Test chart operations
go run . chart install --cluster dev-test

# Test dev tools
go run . dev intercept --help
```

### 3. Running Tests

#### Unit Tests
```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test ./internal/cluster/...

# Run specific test
go test -run TestClusterCreate ./internal/cluster/services/

# Run tests with race detection
go test -race ./...
```

#### Integration Tests
```bash
# Run integration tests (requires Docker)
make integration-test

# Run specific integration test
go test -tags=integration ./tests/integration/cluster/

# Run integration tests with custom cluster
INTEGRATION_CLUSTER=my-test-cluster make integration-test
```

#### Test with Different Configurations
```bash
# Test with different Go versions
go test ./...
GOOS=windows GOARCH=amd64 go test ./...
GOOS=darwin GOARCH=amd64 go test ./...

# Test with different build tags
go test -tags=debug ./...
go test -tags=integration ./...
```

## Debugging

### Local Debugging

#### Using VS Code
1. Set breakpoints by clicking in the gutter
2. Press F5 or use "Run and Debug" panel  
3. Select "Debug OpenFrame" configuration
4. Use debug console for interactive debugging

#### Using Delve (Command Line)
```bash
# Debug main application
dlv debug . -- cluster create debug-cluster --verbose

# Set breakpoints
(dlv) break internal/cluster/services.(*ClusterService).Create
(dlv) continue

# Inspect variables
(dlv) locals
(dlv) print clusterConfig

# Step through code
(dlv) next    # Step over
(dlv) step    # Step into
(dlv) finish  # Step out
```

### Debugging Specific Components

#### Cluster Operations
```bash
# Debug cluster creation with extra logging
OPENFRAME_DEBUG=true go run . cluster create debug-cluster \
  --nodes 1 \
  --api-port 6444 \
  --verbose

# Debug with k3d logs
k3d cluster create debug-cluster --verbose
k3d cluster logs debug-cluster
```

#### Chart Installation
```bash
# Debug ArgoCD installation
OPENFRAME_DEBUG=true go run . chart install \
  --cluster debug-cluster \
  --verbose

# Check ArgoCD status
kubectl get pods -n argocd
kubectl logs -n argocd deployment/argocd-server
```

#### Development Tools
```bash
# Debug Telepresence intercepts
OPENFRAME_DEBUG=true go run . dev intercept my-service \
  --port 3000 \
  --verbose

# Debug Skaffold integration
OPENFRAME_DEBUG=true go run . dev scaffold \
  --cluster debug-cluster \
  --verbose
```

### Common Debug Scenarios

#### External Command Execution
```bash
# Enable command execution debugging
export OPENFRAME_EXECUTOR_DEBUG=true

# This will log all external commands executed
go run . cluster create test-cluster
```

#### Prerequisites Installation  
```bash
# Debug prerequisite checking and installation
export OPENFRAME_PREREQ_DEBUG=true

go run . cluster create test-cluster
```

#### UI and Prompts
```bash
# Debug UI components and user prompts
export OPENFRAME_UI_DEBUG=true

go run . bootstrap  # Interactive mode will show debug info
```

## Development Best Practices

### Code Organization

#### Adding New Commands
1. **Create command file**: `cmd/{group}/{command}.go`
2. **Implement Cobra command**: Use consistent patterns from existing commands
3. **Add business logic**: Create or extend services in `internal/{group}/services/`
4. **Handle errors properly**: Use structured errors from `internal/shared/errors/`
5. **Add tests**: Both unit tests for services and integration tests for commands

#### Example: Adding a New Subcommand
```go
// cmd/cluster/restart.go
package cluster

import (
    "github.com/spf13/cobra"
    "github.com/flamingo-stack/openframe-cli/internal/cluster/services"
    "github.com/flamingo-stack/openframe-cli/internal/shared/ui"
)

func getRestartCmd() *cobra.Command {
    var restartCmd = &cobra.Command{
        Use:   "restart [cluster-name]",
        Short: "Restart a K3d cluster",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            clusterName := args[0]
            
            // Show progress
            ui.ShowLogo()
            ui.Info("Restarting cluster: %s", clusterName)
            
            // Execute business logic
            clusterService := services.NewClusterService()
            return clusterService.Restart(clusterName)
        },
    }
    
    return restartCmd
}
```

### Error Handling

#### Use Structured Errors
```go
// Good: Use structured errors
import "github.com/flamingo-stack/openframe-cli/internal/shared/errors"

func (s *ClusterService) Create(config *models.ClusterConfig) error {
    if !s.validator.IsValid(config) {
        return errors.NewValidationError("invalid cluster configuration", 
            errors.WithField("cluster", config.Name),
            errors.WithSuggestion("Check cluster name and node count"))
    }
    
    if err := s.provider.CreateCluster(config); err != nil {
        return errors.NewClusterError("failed to create cluster", err,
            errors.WithCluster(config.Name),
            errors.WithRetryable(true))
    }
    
    return nil
}
```

#### Provide Actionable Error Messages
```go
// Good: Actionable error message with context
return errors.NewPrerequisiteError("Docker is not running",
    errors.WithSuggestion("Start Docker Desktop or run 'sudo systemctl start docker'"),
    errors.WithDocLink("https://docs.openframe.dev/troubleshooting/docker"))

// Bad: Generic error
return fmt.Errorf("docker error")
```

### Testing Patterns

#### Unit Testing Services
```go
// internal/cluster/services/cluster_service_test.go
func TestClusterService_Create(t *testing.T) {
    // Use dependency injection for testing
    mockProvider := &mocks.ClusterProvider{}
    mockUI := &mocks.UIProvider{}
    
    service := &ClusterService{
        provider: mockProvider,
        ui:       mockUI,
    }
    
    // Setup expectations
    mockProvider.On("CreateCluster", mock.AnythingOfType("*models.ClusterConfig")).
        Return(nil)
    
    // Execute test
    config := &models.ClusterConfig{Name: "test-cluster", Nodes: 1}
    err := service.Create(config)
    
    // Assert results
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

#### Integration Testing Commands
```go
// tests/integration/cluster/cluster_test.go
//go:build integration

func TestClusterCreateCommand(t *testing.T) {
    // Setup test environment
    testCluster := "integration-test-" + uuid.New().String()[:8]
    defer func() {
        // Cleanup
        exec.Command("k3d", "cluster", "delete", testCluster).Run()
    }()
    
    // Execute command
    cmd := exec.Command("go", "run", ".", "cluster", "create", testCluster)
    output, err := cmd.CombinedOutput()
    
    // Assert results  
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Cluster created successfully")
    
    // Verify cluster exists
    verifyCmd := exec.Command("k3d", "cluster", "list")
    verifyOutput, _ := verifyCmd.Output()
    assert.Contains(t, string(verifyOutput), testCluster)
}
```

## Build and Release

### Building for Different Platforms

```bash
# Build for current platform
make build

# Build for all supported platforms
make build-all

# Build for specific platform
GOOS=windows GOARCH=amd64 make build
GOOS=darwin GOARCH=arm64 make build  # Apple Silicon
GOOS=linux GOARCH=amd64 make build
```

### Local Installation
```bash
# Install to GOPATH/bin
make install

# Install to custom location
PREFIX=/usr/local make install

# Create distribution package
make dist
```

### Development Build Optimization

#### Fast Iteration Builds
```bash
# Disable CGO for faster builds
CGO_ENABLED=0 go build -o openframe .

# Build without optimization for faster compile
go build -gcflags="all=-N -l" -o openframe .

# Use build cache
go build -a -o openframe .
```

#### Debug Builds
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -ldflags="-X main.version=debug" -o openframe .

# Build with race detector
go build -race -o openframe .

# Build with memory sanitizer (Linux only)
go build -msan -o openframe .
```

## Performance Profiling

### CPU Profiling
```bash
# Add profiling to your code
import _ "net/http/pprof"
import "net/http"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Profile during execution
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze profile
(pprof) top10
(pprof) web
(pprof) list main.main
```

### Memory Profiling
```bash
# Profile memory allocation
go tool pprof http://localhost:6060/debug/pprof/heap

# Profile memory allocations over time
go tool pprof http://localhost:6060/debug/pprof/allocs
```

### Benchmark Testing
```bash
# Run benchmarks
go test -bench=. ./internal/cluster/services/

# Profile benchmarks
go test -bench=. -cpuprofile=cpu.prof ./internal/cluster/services/
go tool pprof cpu.prof
```

## Troubleshooting Development Issues

### Common Build Issues

#### Dependency Problems
```bash
# Clear module cache
go clean -modcache

# Update dependencies
go get -u ./...
go mod tidy

# Verify checksums
go mod verify
```

#### Import Path Issues
```bash
# Fix import paths
goimports -w .

# Check for circular dependencies
go list -f '{{join .Deps "\n"}}' . | sort | uniq
```

### Common Runtime Issues

#### Docker Connection Issues
```bash
# Check Docker socket
ls -la /var/run/docker.sock

# Test Docker connection
docker info

# Fix permissions (Linux)
sudo usermod -aG docker $USER
# Log out and back in
```

#### Kubernetes Context Issues
```bash
# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch context
kubectl config use-context k3d-openframe-dev
```

#### Port Conflicts
```bash
# Find processes using ports
lsof -i :8080
lsof -i :6443

# Kill process using port
kill -9 $(lsof -t -i:8080)
```

## Development Scripts and Automation

### Useful Development Scripts

#### Quick Test Script
```bash
#!/bin/bash
# scripts/quick-test.sh

echo "🧪 Running quick development tests..."

# Unit tests
echo "Running unit tests..."
go test -short ./...

# Lint
echo "Running linter..."
golangci-lint run

# Build test
echo "Testing build..."
go build -o /tmp/openframe .

echo "✅ Quick tests completed!"
```

#### Integration Test Script
```bash
#!/bin/bash
# scripts/integration-test.sh

set -e

CLUSTER_NAME="integration-test-$$"

echo "🔧 Setting up integration test environment..."

# Create test cluster
k3d cluster create $CLUSTER_NAME --wait

# Run integration tests
go test -tags=integration ./tests/integration/...

# Cleanup
k3d cluster delete $CLUSTER_NAME

echo "✅ Integration tests completed!"
```

#### Development Environment Reset
```bash
#!/bin/bash
# scripts/reset-dev-env.sh

echo "🔄 Resetting development environment..."

# Stop all k3d clusters
k3d cluster list -o json | jq -r '.[].name' | xargs -I {} k3d cluster delete {}

# Clean Go cache
go clean -cache -modcache

# Reset Docker
docker system prune -f

# Rebuild
make clean build

echo "✅ Development environment reset complete!"
```

---

**Next Steps:**
- **[Architecture Overview](../architecture/overview.md)** - Understand the system design
- **[Testing Overview](../testing/overview.md)** - Learn about the testing strategy  
- **[Contributing Guidelines](../contributing/guidelines.md)** - Submit your first contribution

> **💡 Pro Tip**: Set up aliases for common development commands to speed up your workflow:
> ```bash
> alias ofdev="go run . --verbose"
> alias oftest="go test ./... -v"
> alias ofbuild="make build && ./bin/openframe"
> ```