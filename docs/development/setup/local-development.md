# Local Development Guide

This guide covers cloning, building, running, and debugging OpenFrame CLI locally for development purposes.

[![OpenFrame v0.3.7 - Enhanced Developer Experience](https://img.youtube.com/vi/O8hbBO5Mym8/maxresdefault.jpg)](https://www.youtube.com/watch?v=O8hbBO5Mym8)

## Quick Start for Developers

```bash
# 1. Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# 2. Install dependencies
go mod tidy

# 3. Build the CLI
go build -o openframe ./main.go

# 4. Run locally
./openframe --help
```

## Repository Setup

### Clone and Initialize

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Set up Git remotes (if forking)
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
git remote -v

# Install dependencies
go mod download
go mod tidy
```

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
├── docs/                  # Documentation
├── scripts/               # Build and deployment scripts
├── .github/               # GitHub workflows and templates
├── main.go                # CLI entry point
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build automation
└── README.md             # Project overview
```

## Building the Project

### Standard Build

```bash
# Build for current platform
go build -o openframe ./main.go

# Build with version info
go build -ldflags "-X main.version=$(git describe --tags --always)" -o openframe ./main.go

# Build for production (optimized)
CGO_ENABLED=0 go build -ldflags "-s -w" -o openframe ./main.go
```

### Cross-Platform Builds

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o openframe-linux-amd64 ./main.go

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o openframe-darwin-amd64 ./main.go
GOOS=darwin GOARCH=arm64 go build -o openframe-darwin-arm64 ./main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o openframe-windows-amd64.exe ./main.go
```

### Using Make

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Run tests
make test

# Run linting
make lint
```

## Running Locally

### Basic Execution

```bash
# Run the built binary
./openframe --help
./openframe --version

# Test bootstrap command (dry-run)
./openframe bootstrap --help

# Test cluster commands
./openframe cluster --help
```

### Development Mode

Set development environment variables:

```bash
# Enable development mode
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug

# Run with development settings
./openframe bootstrap --verbose
```

### Running Without Building

```bash
# Run directly with go run
go run ./main.go --help
go run ./main.go bootstrap --verbose

# Run with development flags
OPENFRAME_DEV_MODE=true go run ./main.go cluster status
```

## Hot Reload Development

### Using Air

Install and configure Air for live reloading:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create .air.toml (see environment setup guide)

# Start live reloading
air

# Air will automatically rebuild and restart when files change
```

### Manual Hot Reload

Create a simple script for quick iteration:

```bash
#!/bin/bash
# Save as scripts/dev-watch.sh
while inotifywait -e modify -r ./cmd ./internal ./pkg; do
    clear
    echo "Rebuilding..."
    go build -o openframe ./main.go
    echo "Ready for testing!"
done
```

## Testing During Development

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./...

# Run specific test packages
go test ./internal/bootstrap/...
go test ./internal/cluster/...

# Run specific tests
go test -run TestBootstrapCommand ./cmd/bootstrap/
```

### Manual Testing

Create test scripts for common scenarios:

```bash
#!/bin/bash
# scripts/test-bootstrap.sh
set -e

echo "Testing bootstrap command..."
./openframe bootstrap test-cluster --deployment-mode=oss-tenant --non-interactive

echo "Testing cluster status..."
./openframe cluster status

echo "Cleaning up..."
./openframe cluster delete test-cluster
```

## Debug Configuration

### Command Line Debugging

```bash
# Debug with delve
dlv debug ./main.go -- bootstrap --verbose

# Debug with specific breakpoints
dlv debug ./main.go
(dlv) break main.main
(dlv) break cmd/bootstrap.(*Service).Execute
(dlv) continue
```

### Logging and Observability

```bash
# Enable debug logging
export OPENFRAME_LOG_LEVEL=debug
./openframe bootstrap --verbose

# Custom logging configuration
export OPENFRAME_LOG_FORMAT=json
export OPENFRAME_LOG_OUTPUT=/tmp/openframe.log
```

### Profiling During Development

```bash
# CPU profiling
go build -o openframe ./main.go
./openframe bootstrap --cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling
go build -gcflags="-m" -o openframe ./main.go
./openframe bootstrap --memprofile=mem.prof
go tool pprof mem.prof
```

## Local Development Workflows

### Feature Development

```bash
# 1. Create feature branch
git checkout -b feature/new-deployment-mode

# 2. Make changes and test locally
go run ./main.go bootstrap --deployment-mode=new-mode

# 3. Run tests
go test ./...

# 4. Lint code
golangci-lint run

# 5. Commit and push
git add .
git commit -m "Add new deployment mode"
git push origin feature/new-deployment-mode
```

### Bug Fixing

```bash
# 1. Reproduce the bug locally
./openframe bootstrap --verbose 2>&1 | tee debug.log

# 2. Add debug logging
export OPENFRAME_LOG_LEVEL=debug

# 3. Use debugger to investigate
dlv debug ./main.go -- bootstrap

# 4. Write failing test
go test -run TestBugRepro ./internal/bootstrap/

# 5. Fix and verify
go test ./internal/bootstrap/
```

### Performance Investigation

```bash
# 1. Benchmark current performance
go test -bench=. -benchmem ./...

# 2. Profile specific operations
go test -cpuprofile=cpu.prof -bench=BenchmarkBootstrap ./internal/bootstrap/

# 3. Analyze results
go tool pprof cpu.prof
```

## Local Environment Configuration

### Configuration Files

Create local configuration for development:

```yaml
# ~/.openframe/config.yaml
development:
  log_level: debug
  cluster_prefix: dev-
  auto_cleanup: true
  timeout: 30m

clusters:
  default_mode: oss-tenant
  auto_install_charts: true
  
charts:
  argocd_version: "5.46.7"
  timeout: "10m"
```

### Environment Variables

```bash
# Development environment setup
export OPENFRAME_CONFIG_FILE=$HOME/.openframe/config.yaml
export OPENFRAME_DEV_MODE=true
export OPENFRAME_CLUSTER_PREFIX=dev-$(whoami)-
export OPENFRAME_AUTO_CLEANUP=true

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config
export K3D_FIX_DNS=1

# Docker development
export DOCKER_BUILDKIT=1
export DOCKER_CLI_EXPERIMENTAL=enabled
```

## Troubleshooting Local Development

### Common Build Issues

```bash
# Go module issues
go clean -modcache
go mod download
go mod tidy

# Build cache issues
go clean -cache
go build -a ./...

# Dependency conflicts
go mod graph | grep conflicting-package
go mod why problematic-dependency
```

### Runtime Issues

```bash
# Docker daemon not running
sudo systemctl start docker

# K3d clusters conflicting
k3d cluster list
k3d cluster delete --all

# Port conflicts
sudo netstat -tulpn | grep :8080
sudo lsof -i :8080
```

### Performance Issues

```bash
# Memory usage monitoring
go build -race -o openframe ./main.go
./openframe bootstrap --memprofile=mem.prof

# CPU usage monitoring  
go build -o openframe ./main.go
./openframe bootstrap --cpuprofile=cpu.prof
```

## Development Best Practices

### Code Organization

- Keep command implementations in `cmd/` directories simple
- Put business logic in `internal/` packages
- Use `pkg/` for reusable public packages
- Write testable code with dependency injection

### Testing Strategy

- Write unit tests for all business logic
- Use integration tests for end-to-end workflows
- Mock external dependencies (Docker, K3d, kubectl)
- Test error conditions and edge cases

### Version Control

```bash
# Commit message format
git commit -m "feat(cluster): add multi-node support

- Add node count configuration to cluster creation
- Update cluster status to show all nodes
- Add validation for node resource requirements

Closes #123"

# Keep commits atomic
git add cmd/cluster/create.go
git commit -m "feat(cluster): add node count parameter"

git add internal/cluster/service.go
git commit -m "feat(cluster): implement multi-node creation"
```

## Next Steps

With local development set up:

1. **[Architecture Overview](../architecture/README.md)** - Understand the system design
2. **[Testing Guide](../testing/README.md)** - Learn the testing approach
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Follow contribution standards

> 🚀 **Pro Tip**: Start by exploring existing commands to understand the patterns, then try modifying a simple command before building new features from scratch.