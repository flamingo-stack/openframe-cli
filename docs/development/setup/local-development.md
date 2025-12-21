# Local Development Guide

Learn how to clone, build, and run OpenFrame CLI locally with hot reloading, debugging, and efficient development workflows.

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# 2. Build and test
make build test

# 3. Install locally
make install

# 4. Verify installation
openframe --version
```

## Repository Structure

Understanding the codebase layout:

```text
openframe-cli/
├── cmd/                    # Command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management commands  
│   ├── chart/              # Chart management commands
│   └── dev/                # Development commands
├── internal/               # Internal packages
│   ├── bootstrap/          # Bootstrap service logic
│   ├── cluster/            # Cluster service logic
│   ├── chart/              # Chart service logic
│   ├── dev/                # Development service logic
│   └── shared/             # Shared utilities
├── pkg/                    # Public packages
├── test/                   # Integration tests
├── docs/                   # Documentation
├── scripts/                # Build and utility scripts
├── Makefile               # Build automation
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
└── main.go                # CLI entry point
```

## Development Workflow

### Building the CLI

```bash
# Development build (faster, with debug info)
make build-dev

# Production build (optimized)
make build

# Build for all platforms
make build-all

# Cross-platform build
GOOS=darwin GOARCH=amd64 go build -o bin/openframe-darwin-amd64 .
GOOS=linux GOARCH=arm64 go build -o bin/openframe-linux-arm64 .
```

### Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race

# Run specific test package
go test ./cmd/bootstrap -v

# Run specific test function
go test ./cmd/bootstrap -run TestBootstrapCmd -v

# Run integration tests (requires Docker)
make test-integration

# Watch tests (requires entr tool)
make test-watch
```

### Code Quality

```bash
# Run all linters
make lint

# Format code
make fmt

# Check for security issues
make security

# Full quality check
make quality  # Runs fmt, lint, test, security
```

## Hot Reloading Development

### Option 1: Air (Recommended)

Install and use Air for automatic rebuilding:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create Air configuration
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["bootstrap", "--verbose"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "node_modules"]
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

[screen]
  clear_on_rebuild = false
EOF

# Start development with hot reloading
air
```

### Option 2: Entr

Use entr for file watching and rebuilding:

```bash
# Install entr (macOS)
brew install entr

# Install entr (Ubuntu/Debian)
sudo apt-get install entr

# Watch Go files and rebuild on changes
find . -name "*.go" | entr -r sh -c 'make build && echo "Build complete - $(date)"'

# Watch and run tests
find . -name "*.go" | entr -r make test
```

### Option 3: Custom Watch Script

Create a custom development script:

```bash
# Create scripts/dev-watch.sh
cat > scripts/dev-watch.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting OpenFrame CLI development watch..."
echo "Watching for changes in *.go files..."

# Initial build
make build

# Watch for changes
fswatch -o . -e ".*" -i "\\.go$" | while read f; do
    echo "File changed, rebuilding..."
    if make build; then
        echo "✅ Build successful - $(date)"
        echo "Run: ./bin/openframe bootstrap --verbose"
    else
        echo "❌ Build failed - $(date)" 
    fi
done
EOF

chmod +x scripts/dev-watch.sh
./scripts/dev-watch.sh
```

## Debugging

### VS Code Debugging

Debug configuration is already set up in `.vscode/launch.json`:

1. Set breakpoints in your code
2. Press `F5` or go to Run and Debug
3. Select "Debug OpenFrame CLI" 
4. The debugger will start with `bootstrap --verbose`

### Command-Line Debugging with Delve

```bash
# Install Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the CLI
dlv debug . -- bootstrap --verbose

# Debug specific test
dlv test ./cmd/bootstrap -- -test.run TestBootstrapCmd

# Debug with breakpoint
dlv debug . -- bootstrap --verbose
(dlv) b main.main
(dlv) c
```

### Debugging Commands

```bash
# Common Delve commands
(dlv) help           # Show all commands
(dlv) b main.go:10   # Set breakpoint at line 10
(dlv) c              # Continue execution
(dlv) n              # Next line
(dlv) s              # Step into function
(dlv) p variable     # Print variable value
(dlv) locals         # Show local variables
(dlv) goroutines     # List goroutines
(dlv) quit           # Exit debugger
```

## Development Environment

### Environment Variables

Set up development-specific configuration:

```bash
# Create development environment file
cat > .env.development << 'EOF'
# OpenFrame CLI Development Configuration
OPENFRAME_DEV=true
OPENFRAME_CONFIG_DIR=./.openframe-dev
OPENFRAME_LOG_LEVEL=debug
OPENFRAME_VERBOSE=true

# Development cluster settings  
OPENFRAME_DEFAULT_CLUSTER=dev-cluster
OPENFRAME_CLUSTER_CPU=2
OPENFRAME_CLUSTER_MEMORY=4

# Development timeouts (longer for debugging)
OPENFRAME_TIMEOUT=300s
OPENFRAME_RETRY_ATTEMPTS=3

# Disable telemetry during development
OPENFRAME_DISABLE_TELEMETRY=true
EOF

# Load in development session
source .env.development
```

### Development Configuration

Create a development-specific config:

```bash
# Create development config directory
mkdir -p .openframe-dev

# Create development configuration
cat > .openframe-dev/config.yaml << 'EOF'
development: true
logLevel: debug
verbose: true
cluster:
  provider: k3d
  defaultName: dev-cluster
  resources:
    cpu: 2
    memory: 4Gi
chart:
  defaultMode: oss-tenant
  timeout: 10m
dev:
  hotReload: true
  debugMode: true
EOF
```

## Local Cluster Development

### Development Cluster Setup

Create a dedicated development cluster:

```bash
# Create development cluster with specific configuration
k3d cluster create openframe-dev \
  --agents 1 \
  --port "8080:80@loadbalancer" \
  --port "8443:443@loadbalancer" \
  --port "6443:6443@server:0" \
  --k3s-arg "--disable=traefik@server:0" \
  --volume "$(pwd):/workspace@all"

# Verify cluster
kubectl cluster-info
kubectl get nodes
```

### Development with Real Clusters

Test against real clusters safely:

```bash
# Create isolated namespace for development
kubectl create namespace openframe-dev-$(whoami)

# Set as default namespace for development
kubectl config set-context --current --namespace=openframe-dev-$(whoami)

# Run CLI against development namespace
openframe chart install --namespace=openframe-dev-$(whoami)
```

## Testing Local Changes

### Unit Testing

```bash
# Run tests for specific component
go test ./internal/bootstrap -v
go test ./cmd/cluster -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run table tests with verbose output
go test ./internal/cluster -v -run TestClusterService
```

### Integration Testing

```bash
# Run integration tests (requires Docker)
make test-integration

# Run specific integration test
go test ./test -run TestBootstrapIntegration -v

# Run integration tests with cleanup
make test-integration-clean
```

### Manual Testing

```bash
# Test bootstrap command
./bin/openframe bootstrap test-cluster --deployment-mode=oss-tenant --non-interactive --verbose

# Test cluster commands
./bin/openframe cluster list
./bin/openframe cluster status test-cluster
./bin/openframe cluster delete test-cluster

# Test chart commands  
./bin/openframe chart install --mode=oss-tenant --verbose

# Test development commands
./bin/openframe dev intercept test-service --port=8080
```

## Development Helpers

### Makefile Targets

```bash
# View all available targets
make help

# Common development targets
make dev           # Build and run in development mode
make dev-install   # Install development build locally  
make dev-clean     # Clean development artifacts
make dev-reset     # Reset development environment

# Testing targets
make test-unit     # Run only unit tests
make test-e2e      # Run end-to-end tests
make test-bench    # Run benchmark tests

# Quality targets
make fmt           # Format all Go code
make lint-fix      # Fix linting issues automatically
make vet           # Run go vet
make ineffassign   # Find ineffectual assignments
```

### Custom Scripts

Create helper scripts for common tasks:

```bash
# scripts/dev-reset.sh - Reset development environment
cat > scripts/dev-reset.sh << 'EOF'
#!/bin/bash
echo "Resetting OpenFrame development environment..."

# Clean build artifacts
make clean

# Delete development clusters
k3d cluster delete openframe-dev 2>/dev/null || true

# Clean development config
rm -rf .openframe-dev

# Rebuild
make build

echo "✅ Development environment reset complete"
EOF

# scripts/quick-test.sh - Quick test specific components  
cat > scripts/quick-test.sh << 'EOF'
#!/bin/bash
component=${1:-"all"}

case $component in
  "bootstrap")
    go test ./cmd/bootstrap ./internal/bootstrap -v
    ;;
  "cluster") 
    go test ./cmd/cluster ./internal/cluster -v
    ;;
  "chart")
    go test ./cmd/chart ./internal/chart -v
    ;;
  "all")
    make test
    ;;
  *)
    echo "Usage: $0 [bootstrap|cluster|chart|all]"
    exit 1
    ;;
esac
EOF

chmod +x scripts/*.sh
```

## Performance Optimization

### Build Optimization

```bash
# Fast development builds (no optimization)
go build -gcflags="all=-N -l" -o bin/openframe-dev .

# Optimized builds
go build -ldflags="-w -s" -o bin/openframe .

# Profile-guided optimization (Go 1.21+)
go build -pgo=auto -o bin/openframe .
```

### Testing Optimization  

```bash
# Parallel testing
go test ./... -parallel 4

# Short tests only (skip long-running tests)
go test ./... -short

# Test with build cache
go test -count=1 ./...  # Disable cache
go test ./...           # Use cache
```

### Memory and CPU Profiling

```bash
# CPU profiling
go test ./internal/cluster -cpuprofile=cpu.prof
go tool pprof cpu.prof

# Memory profiling  
go test ./internal/cluster -memprofile=mem.prof
go tool pprof mem.prof

# Benchmark profiling
go test ./internal/cluster -bench=. -cpuprofile=bench.prof
```

## Troubleshooting Development Issues

### Build Issues

```bash
# Clear module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod vendor  # If using vendor

# Verify dependencies
go mod verify
```

### Test Issues

```bash
# Clean test cache
go clean -testcache

# Run tests with detailed output
go test ./... -v -x

# Debug failing tests
go test ./cmd/bootstrap -v -run TestSpecificFunction
```

### Runtime Issues

```bash
# Enable debug logging
export OPENFRAME_LOG_LEVEL=debug
./bin/openframe bootstrap --verbose

# Check configuration
./bin/openframe --help
./bin/openframe cluster --help

# Verify prerequisites
which docker kubectl k3d helm
```

## Contributing Workflow

### Development Branch Strategy

```bash
# Create feature branch
git checkout -b feature/my-new-feature

# Make changes and test
make dev-test

# Commit with conventional format
git commit -m "feat: add new cluster management feature"

# Push and create PR
git push origin feature/my-new-feature
```

### Pre-commit Checks

```bash
# Run full quality check before committing
make quality

# Check specific areas
make lint        # Linting
make test       # Testing  
make security   # Security analysis
make fmt        # Code formatting
```

## Next Steps

With local development set up:

1. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase design
2. **[Testing Overview](../testing/overview.md)** - Learn testing patterns and practices  
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Contribute your improvements
4. Explore the codebase starting with `main.go` and `cmd/` packages

## Quick Reference

### Development Commands

| Task | Command |
|------|---------|
| Build development binary | `make build-dev` |
| Run with hot reload | `air` |
| Run all tests | `make test` |
| Debug with Delve | `dlv debug . -- bootstrap --verbose` |
| Integration test | `make test-integration` |
| Format and lint | `make quality` |
| Reset dev environment | `./scripts/dev-reset.sh` |

---

**Happy developing!** 🚀 You're now ready to contribute to OpenFrame CLI. Next: **[Architecture Overview](../architecture/overview.md)** to understand the codebase structure.