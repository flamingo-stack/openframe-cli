# Local Development Guide

This guide walks you through cloning, building, and running OpenFrame CLI locally for development and testing purposes.

## Prerequisites

Ensure you have completed the [Environment Setup](environment.md) guide before proceeding:
- Go 1.21+ installed and configured
- Development tools (goimports, golangci-lint, etc.)
- IDE configured with Go support
- Docker and Kubernetes tools for testing

## Repository Setup

### Clone the Repository

```bash
# Clone via HTTPS
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Or clone via SSH (if you have SSH keys configured)
git clone git@github.com:flamingo-stack/openframe-cli.git
cd openframe-cli

# Verify the clone
ls -la
git status
```

### Understand the Repository Structure

```bash
# Explore the project structure
tree -L 2
# OR
find . -type d -maxdepth 2 | grep -v '.git' | sort
```

**Key directories:**
```text
openframe-cli/
├── cmd/                    # CLI command implementations
├── internal/               # Internal packages and business logic
├── tests/                  # Test suites and utilities
├── docs/                   # Documentation
├── scripts/                # Build and utility scripts
├── deployments/           # Deployment configurations
├── .github/               # GitHub workflows and templates
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── Makefile              # Build automation
└── main.go               # Application entry point
```

### Initialize Development Environment

```bash
# Download Go dependencies
go mod download

# Install development tools
make dev-deps

# Verify setup
go mod verify
go mod tidy
```

## Building the Project

### Standard Build

```bash
# Build the CLI binary
make build

# Or manually
go build -o bin/openframe main.go

# Verify the build
./bin/openframe --version
./bin/openframe --help
```

### Development Build

```bash
# Build with debug information and race detection
go build -race -gcflags="all=-N -l" -o bin/openframe-debug main.go

# Build for all platforms (cross-compilation)
make build-all

# View available build targets
make help
```

### Custom Build Flags

```bash
# Build with version information
VERSION=$(git describe --tags --always)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags="-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o bin/openframe main.go
```

## Running Locally

### Basic Usage

```bash
# Run the locally built CLI
./bin/openframe --help
./bin/openframe --version

# Test basic commands
./bin/openframe cluster --help
./bin/openframe bootstrap --help
```

### Development Mode

```bash
# Set development environment variables
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug

# Run with verbose logging
./bin/openframe bootstrap --deployment-mode=oss-tenant --verbose

# Run without building (go run)
go run main.go cluster list
```

### Testing Different Commands

```bash
# Test cluster operations
./bin/openframe cluster create test-cluster
./bin/openframe cluster list
./bin/openframe cluster status test-cluster

# Test bootstrap functionality
./bin/openframe bootstrap test-env --deployment-mode=oss-tenant

# Test chart operations
./bin/openframe chart install --help
```

## Hot Reload Development

### Using air for Hot Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["--deployment-mode=oss-tenant", "--verbose"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  kill_delay = "0s"
  log = "build-errors.log"
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = false
EOF

# Start hot reload development
air
```

### Using Watch Mode

```bash
# Simple file watcher script
cat > scripts/watch-dev.sh << 'EOF'
#!/bin/bash

echo "Starting OpenFrame CLI development watch mode..."

# Function to rebuild and restart
rebuild_and_restart() {
    echo "🔄 Changes detected, rebuilding..."
    
    # Kill existing process
    pkill -f "bin/openframe" 2>/dev/null
    
    # Rebuild
    if make build; then
        echo "✅ Build successful"
        # Run in background for testing
        # ./bin/openframe bootstrap --deployment-mode=oss-tenant &
    else
        echo "❌ Build failed"
    fi
}

# Watch for changes
fswatch -o . --exclude='.git' --exclude='bin/' --exclude='tmp/' | while read f; do
    rebuild_and_restart
done
EOF

chmod +x scripts/watch-dev.sh
./scripts/watch-dev.sh
```

## Debug Configuration

### VS Code Debugging

The [Environment Setup](environment.md) guide includes VS Code debug configurations. To use them:

1. Open project in VS Code: `code .`
2. Set breakpoints in the code
3. Press `F5` or use Debug menu
4. Choose "Debug OpenFrame CLI" configuration

### Command Line Debugging

```bash
# Debug with delve
dlv debug main.go -- bootstrap --deployment-mode=oss-tenant

# In dlv debugger:
# (dlv) break main.main
# (dlv) continue
# (dlv) print variableName
# (dlv) step
# (dlv) next
```

### Debug Specific Components

```bash
# Debug cluster service
dlv debug main.go -- cluster create debug-cluster --verbose

# Debug chart installation
dlv debug main.go -- chart install --deployment-mode=oss-tenant --dry-run

# Debug with environment variables
OPENFRAME_LOG_LEVEL=trace dlv debug main.go -- bootstrap test
```

## Testing Locally

### Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run tests for specific package
go test -v ./internal/cluster/...

# Run tests with race detection
go test -race ./...

# Run specific test
go test -v -run TestClusterService_CreateCluster ./internal/cluster/
```

### Integration Tests

```bash
# Run integration tests (requires Docker)
make test-integration

# Run with verbose output
make test-integration ARGS="-v"

# Run specific integration test
go test -v ./tests/integration/cluster/

# Run integration tests with custom cluster
CLUSTER_NAME=integration-test make test-integration
```

### Manual Testing

```bash
# Create test cluster for manual testing
./bin/openframe cluster create manual-test --verbose

# Test different deployment modes
./bin/openframe bootstrap test-oss --deployment-mode=oss-tenant
./bin/openframe bootstrap test-saas --deployment-mode=saas-tenant

# Test error conditions
./bin/openframe cluster create "" # Should show validation error
./bin/openframe bootstrap --deployment-mode=invalid # Should show error
```

## Development Workflow

### Daily Development Flow

```bash
# 1. Start development session
cd openframe-cli
git status
git pull origin main

# 2. Create feature branch
git checkout -b feature/my-new-feature

# 3. Make changes and test iteratively
# Edit code...
make build
./bin/openframe <test-command>

# 4. Run tests
make test
make lint

# 5. Commit changes
git add .
git commit -m "feat: add new feature"

# 6. Push and create PR
git push origin feature/my-new-feature
```

### Code Quality Checks

```bash
# Format code
make fmt

# Run linter
make lint

# Fix common issues
goimports -w .
go mod tidy

# Run full quality check
make dev  # Runs fmt, vet, lint, test, build
```

### Dependency Management

```bash
# Add new dependency
go get github.com/some/package@v1.2.3

# Update dependencies
go get -u ./...

# Clean up dependencies
go mod tidy

# Verify dependencies
go mod verify

# View dependency graph
go mod graph | head -20
```

## Local Testing with Real Clusters

### K3d Testing

```bash
# Create K3d cluster for testing
k3d cluster create openframe-dev --port "8080:80@loadbalancer"

# Test against real cluster
export KUBECONFIG=$(k3d kubeconfig write openframe-dev)
./bin/openframe chart install --deployment-mode=oss-tenant

# Cleanup
k3d cluster delete openframe-dev
```

### Docker Desktop Testing

```bash
# Use Docker Desktop Kubernetes
kubectl config use-context docker-desktop

# Test OpenFrame CLI
./bin/openframe bootstrap docker-test --deployment-mode=oss-tenant

# Cleanup
kubectl delete namespace openframe
```

## Troubleshooting Local Development

### Common Build Issues

```bash
# Clear Go module cache
go clean -modcache

# Rebuild everything
make clean && make build

# Check for conflicting versions
go version
which go
```

### Runtime Issues

```bash
# Check environment
env | grep OPENFRAME
env | grep KUBECONFIG

# Test with clean environment
env -i PATH=$PATH HOME=$HOME go run main.go --version

# Debug with verbose logging
OPENFRAME_LOG_LEVEL=trace ./bin/openframe bootstrap test --verbose
```

### Permission Issues

```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Fix file permissions
chmod +x bin/openframe
chmod +x scripts/*.sh
```

### IDE Issues

```bash
# Restart Go language server in VS Code
# Cmd/Ctrl + Shift + P -> "Go: Restart Language Server"

# Clear VS Code workspace
rm -rf .vscode/settings.json
# Recreate from environment setup guide

# Verify Go tools
go env GOROOT
go env GOPATH
which goimports
```

## Performance Optimization

### Build Performance

```bash
# Use build cache
export GOCACHE=$(go env GOCACHE)

# Parallel builds
make build -j$(nproc)

# Use go build cache
go env GOCACHE  # Should show cache directory
```

### Runtime Performance

```bash
# Profile CPU usage
./bin/openframe bootstrap test --cpuprofile=cpu.prof

# Profile memory usage
./bin/openframe bootstrap test --memprofile=mem.prof

# Analyze profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

## Next Steps

With your local development environment running:

1. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure
2. **[Testing Overview](../testing/overview.md)** - Learn the testing strategy
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Submit your first contribution

## Development Resources

### Useful Commands Reference

| Task | Command |
|------|---------|
| **Build** | `make build` |
| **Test** | `make test` |
| **Lint** | `make lint` |
| **Format** | `make fmt` |
| **All Quality Checks** | `make dev` |
| **Clean** | `make clean` |
| **Debug** | `dlv debug main.go -- <args>` |
| **Hot Reload** | `air` |
| **Integration Tests** | `make test-integration` |

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `OPENFRAME_DEV_MODE` | Enable development features | `true` |
| `OPENFRAME_LOG_LEVEL` | Set logging verbosity | `debug`, `trace` |
| `KUBECONFIG` | Kubernetes config file | `~/.kube/config` |
| `DOCKER_HOST` | Docker daemon socket | `unix:///var/run/docker.sock` |

---

**Development Environment Ready!** You can now efficiently develop, test, and debug OpenFrame CLI locally. The hot reload setup will help you iterate quickly during development.