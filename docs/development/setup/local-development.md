# Local Development Guide

This guide walks you through cloning, building, and running OpenFrame CLI locally for development. Follow these steps to get a complete development workflow set up on your machine.

## Repository Setup

### Clone the Repository
```bash
# Clone the main repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Or clone your fork for contributions
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote (if you forked)
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
```

### Repository Structure
```text
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management commands  
│   ├── chart/              # Chart management commands
│   ├── dev/                # Development tools commands
│   └── root.go             # Root command setup
├── internal/               # Internal packages
│   ├── bootstrap/          # Bootstrap service logic
│   ├── cluster/            # Cluster management logic
│   ├── chart/              # Chart management logic
│   ├── dev/                # Development tools logic
│   └── shared/             # Shared utilities
├── tests/                  # Test files
│   ├── integration/        # Integration tests
│   ├── mocks/              # Test mocks
│   └── testutil/           # Test utilities
├── docs/                   # Documentation
├── scripts/                # Build and utility scripts
├── main.go                 # Application entry point
├── Makefile                # Build automation
├── go.mod                  # Go module definition
└── go.sum                  # Go module checksums
```

## Build and Development Workflow

### Initial Setup
```bash
# Verify Go version (1.21+)
go version

# Download dependencies
go mod download

# Verify dependencies
go mod verify
go mod tidy
```

### Building the CLI

#### Quick Build
```bash
# Build for current platform
go build -o openframe main.go

# Test the build
./openframe --version
./openframe --help
```

#### Using Makefile
```bash
# View available targets
make help

# Build optimized binary
make build

# Build with race detection (for testing)
make build-race

# Cross-compile for different platforms
make build-linux
make build-darwin
make build-windows
```

#### Development Build with Debug Info
```bash
# Build with debug symbols and no optimizations
go build -gcflags="all=-N -l" -o openframe-dev main.go

# Or use debug make target
make build-debug
```

### Running Locally

#### Direct Execution
```bash
# Run without building (slower)
go run main.go --help
go run main.go bootstrap --help

# Run with arguments
go run main.go cluster create test-cluster
go run main.go cluster list
```

#### Using Built Binary
```bash
# Build once, run multiple times (faster)
make build
./openframe bootstrap
./openframe cluster status
./openframe dev --help
```

### Live Reload Development

#### Using Air (Recommended)
```bash
# Install air if not already installed
go install github.com/cosmtrek/air@latest

# Run with live reload
air

# Air will automatically rebuild and restart when files change
```

#### Manual Watch Script
```bash
# Create a simple watch script
cat > watch.sh << 'EOF'
#!/bin/bash
while inotifywait -e modify -r .; do
    echo "Files changed, rebuilding..."
    make build && echo "Build complete!"
done
EOF

chmod +x watch.sh
./watch.sh
```

## Development Tasks

### Testing

#### Run All Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...

# Run specific test package
go test ./internal/cluster/...
```

#### Test with Different Verbosity
```bash
# Verbose output
go test -v ./...

# Show test coverage per function
go test -coverprofile=coverage.out ./...
go tool cover -func coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

#### Integration Tests
```bash
# Run integration tests (requires Docker)
export OPENFRAME_INTEGRATION_TESTS=true
go test -tags=integration ./tests/integration/...

# Run specific integration test
go test -tags=integration ./tests/integration/common/
```

### Code Quality

#### Linting
```bash
# Run all linters
make lint

# Run specific linters
golangci-lint run --enable=gosec
golangci-lint run --enable=gocyclo

# Fix auto-fixable issues
golangci-lint run --fix
```

#### Code Formatting
```bash
# Format all code
make fmt

# Check formatting without fixing
gofumpt -l .

# Format specific files
gofumpt -w ./cmd/cluster/
```

#### Security Scanning
```bash
# Run security checks
gosec ./...

# Check for vulnerabilities in dependencies
go list -json -m all | nancy sleuth
```

### Debugging

#### VS Code Debugging
1. Set breakpoints in VS Code
2. Press F5 or use "Run and Debug" panel
3. Choose "Debug CLI" configuration
4. Modify args in `.vscode/launch.json` as needed

#### Command Line Debugging
```bash
# Build with debug info
go build -gcflags="all=-N -l" -o openframe-debug main.go

# Run with delve debugger
dlv exec ./openframe-debug -- bootstrap --verbose
```

#### Verbose Logging
```bash
# Enable debug logging
export OPENFRAME_LOG_LEVEL=debug
./openframe bootstrap --verbose

# Trace specific operations
export OPENFRAME_TRACE=cluster,chart
./openframe bootstrap test-cluster
```

### Hot Reload Development Setup

#### Terminal Setup for Efficient Development
```bash
# Terminal 1: Watch and build
air

# Terminal 2: Test cluster operations
./openframe cluster create dev-test
./openframe cluster status

# Terminal 3: Monitor logs
tail -f ~/.openframe/logs/openframe.log

# Terminal 4: Watch tests
gotestsum --watch ./...
```

#### Development Configuration

Create a development config file `~/.openframe/dev-config.yaml`:
```yaml
development:
  auto_cleanup: true
  verbose_logging: true
  cluster_prefix: "dev-"
  chart_timeout: "300s"
  
testing:
  integration_enabled: true
  test_cluster_name: "openframe-test"
  cleanup_after_tests: true
```

## Working with Dependencies

### Go Module Management
```bash
# Add new dependency
go get github.com/example/package

# Upgrade dependency
go get -u github.com/example/package

# Upgrade all dependencies
go get -u ./...

# Remove unused dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Dependency Verification
```bash
# Verify all dependencies
go mod verify

# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Audit dependencies
go list -m -versions all
```

## Local Testing Workflows

### Manual Testing
```bash
# Create test cluster
./openframe cluster create test-$(date +%s)

# Test bootstrap flow
./openframe bootstrap test-bootstrap --verbose

# Test development tools
./openframe dev intercept --list

# Cleanup
./openframe cluster list
./openframe cluster delete test-cluster
```

### Automated Testing Scripts

Create `scripts/test-all.sh`:
```bash
#!/bin/bash
set -e

echo "🧪 Running OpenFrame CLI test suite..."

# Build
echo "🔨 Building..."
make build

# Unit tests
echo "🔬 Unit tests..."
make test-coverage

# Linting
echo "📋 Linting..."
make lint

# Integration tests (optional)
if [ "${OPENFRAME_INTEGRATION_TESTS}" = "true" ]; then
    echo "🔗 Integration tests..."
    go test -tags=integration ./tests/integration/...
fi

# Manual smoke test
echo "💨 Smoke test..."
./openframe --version
./openframe --help

echo "✅ All tests passed!"
```

Make it executable and run:
```bash
chmod +x scripts/test-all.sh
./scripts/test-all.sh
```

## Performance Profiling

### CPU Profiling
```bash
# Build with profiling
go build -o openframe-profile main.go

# Run with CPU profiling
./openframe-profile bootstrap test-profile --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling
```bash
# Run with memory profiling
./openframe bootstrap test-memory --memprofile=mem.prof

# Analyze memory usage
go tool pprof mem.prof
```

### Benchmark Testing
```bash
# Run benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkClusterCreate ./internal/cluster/
```

## Troubleshooting

### Common Development Issues

**Build Failures**:
```bash
# Clean build cache
go clean -cache
go clean -modcache

# Rebuild dependencies
go mod download
go mod tidy
```

**Import Path Issues**:
```bash
# Verify module path
grep module go.mod

# Fix import paths in IDE
# VS Code: Ctrl+Shift+P > "Go: Restart Language Server"
```

**Test Failures**:
```bash
# Run specific failing test
go test -v ./internal/cluster/ -run TestClusterCreate

# Run with race detection
go test -race ./internal/cluster/

# Clean test cache
go clean -testcache
```

**Docker/K3D Issues**:
```bash
# Ensure Docker is running
docker info

# Clean up test clusters
k3d cluster list
k3d cluster delete --all

# Reset Docker if needed
docker system prune -a
```

### Development Environment Issues

**VS Code not working**:
```bash
# Restart Go language server
# Command Palette: "Go: Restart Language Server"

# Reinstall Go tools
# Command Palette: "Go: Install/Update Tools"
```

**Live reload not working**:
```bash
# Check air configuration
air -v

# Verify file watchers
ulimit -n  # Should be > 1024

# Manual restart
pkill air && air
```

## Next Steps

Now that you have local development working:

1. **[Understand the Architecture](../architecture/overview.md)** - Learn how components interact
2. **[Run Tests](../testing/overview.md)** - Understand the testing strategy  
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the contribution process
4. **Make Your First Change** - Pick up a "good first issue" and start contributing!

### Recommended First Development Tasks

1. **Fix a typo or improve documentation**
2. **Add a test case for existing functionality**
3. **Implement a small feature or enhancement**
4. **Improve error messages or help text**
5. **Add validation to command inputs**

---

Happy coding! 🚀 The OpenFrame CLI development environment is now ready for productive development work.