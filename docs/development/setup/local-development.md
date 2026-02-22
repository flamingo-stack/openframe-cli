# Local Development Guide

This guide covers cloning, building, running, and debugging OpenFrame CLI in your local development environment. Follow these steps to get the code running and start contributing.

## Prerequisites

Before starting local development, ensure you have completed:
- **[Prerequisites](../../getting-started/prerequisites.md)** - System requirements and dependencies
- **[Environment Setup](environment.md)** - IDE, tools, and development configuration

## Clone the Repository

### Fork and Clone (Recommended for Contributors)

1. **Fork the repository** on GitHub:
   - Go to https://github.com/flamingo-stack/openframe-cli
   - Click "Fork" in the top-right corner
   - Choose your GitHub account

2. **Clone your fork**:
   ```bash
   # Clone your fork
   git clone https://github.com/YOUR-USERNAME/openframe-cli.git
   cd openframe-cli
   
   # Add upstream remote for syncing
   git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
   
   # Verify remotes
   git remote -v
   ```

### Direct Clone (Read-only)

For read-only access or testing:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

## Project Structure Overview

Familiarize yourself with the codebase structure:

```text
openframe-cli/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency checksums
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and version info
│   ├── bootstrap/         # Complete environment bootstrap
│   ├── cluster/           # Kubernetes cluster management
│   ├── chart/             # Helm chart and ArgoCD operations
│   └── dev/               # Development workflow tools
├── internal/              # Private application code
│   ├── bootstrap/         # Bootstrap orchestration service
│   ├── cluster/           # Cluster lifecycle management
│   ├── chart/             # Chart installation and ArgoCD integration
│   ├── dev/               # Development tools (intercept, scaffold)
│   └── shared/            # Common utilities and adapters
├── tests/                 # Test suites and utilities
│   ├── integration/       # Integration tests
│   ├── mocks/             # Test mocks and fixtures
│   └── testutil/          # Test helper functions
├── docs/                  # Documentation
├── examples/              # Usage examples and samples
└── scripts/               # Build and utility scripts
```

## Build and Run

### Build the Binary

```bash
# Build for your current platform
go build -o openframe main.go

# Build with version information
VERSION=$(git describe --tags --always --dirty)
COMMIT=$(git rev-parse --short HEAD)
DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

go build -ldflags "-X main.version=$VERSION -X main.commit=$COMMIT -X main.date=$DATE" -o openframe main.go

# Test the build
./openframe --version
```

### Run During Development

For rapid development cycles, run directly with Go:

```bash
# Run with go run (rebuilds automatically)
go run main.go --help

# Run specific commands
go run main.go bootstrap --help
go run main.go cluster status
go run main.go chart list
```

### Hot Reload with Air

Install and use Air for automatic rebuilds:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
args_bin = []
bin = "./tmp/main"
cmd = "go build -o ./tmp/main main.go"
delay = 1000
exclude_dir = ["assets", "tmp", "vendor", "testdata", "docs"]
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

# Run with hot reload
air
```

Now changes to Go files will automatically trigger rebuilds.

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
# or xdg-open coverage.html  # Linux
```

### Integration Tests

Integration tests require a running Kubernetes cluster:

```bash
# Start a test cluster
k3d cluster create openframe-test

# Run integration tests
go test -tags=integration ./tests/integration/...

# Clean up
k3d cluster delete openframe-test
```

### Test Specific Packages

```bash
# Test specific packages
go test ./internal/cluster/...
go test ./internal/chart/...
go test ./cmd/bootstrap/...

# Test with timeout
go test -timeout=30s ./internal/bootstrap/...

# Run specific tests
go test -run TestClusterCreate ./internal/cluster/...
go test -run TestBootstrapService ./internal/bootstrap/...
```

## Development Workflow

### Create a Feature Branch

```bash
# Sync with upstream (if using fork)
git fetch upstream
git checkout main
git merge upstream/main

# Create feature branch
git checkout -b feature/your-feature-name

# Or for bug fixes
git checkout -b fix/issue-description
```

### Make Changes

1. **Write Code**: Implement your feature or fix
2. **Write Tests**: Add or update tests for your changes
3. **Run Tests**: Ensure all tests pass
4. **Format Code**: Use `gofmt` and `goimports`
5. **Lint Code**: Run `golangci-lint`

```bash
# Format and organize imports
gofmt -w .
goimports -w .

# Run linter
golangci-lint run

# Run all tests
go test ./...
```

### Commit Changes

Follow conventional commit format:

```bash
# Stage changes
git add .

# Commit with descriptive message
git commit -m "feat(cluster): add support for custom node labels"
git commit -m "fix(bootstrap): handle timeout errors gracefully"
git commit -m "docs(readme): update installation instructions"

# Push to your fork
git push origin feature/your-feature-name
```

## Debugging

### VS Code Debugging

Use the launch configurations from [Environment Setup](environment.md):

1. **Set breakpoints** in your code
2. **Press F5** or go to Run → Start Debugging
3. **Choose configuration**:
   - "Launch OpenFrame CLI" - Debug with `--help`
   - "Debug Bootstrap Command" - Debug bootstrap process
   - "Debug Cluster Status" - Debug cluster operations

### Command-line Debugging with Delve

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug main.go -- bootstrap --verbose

# Debug tests
dlv test ./internal/bootstrap/

# Debug with arguments
dlv debug main.go -- cluster create --name=test-cluster
```

### Debug Commands in Delve

```text
(dlv) break main.main          # Set breakpoint at main function
(dlv) break bootstrap.go:45    # Set breakpoint at line 45 in bootstrap.go
(dlv) continue                 # Continue execution
(dlv) next                     # Execute next line
(dlv) step                     # Step into function calls
(dlv) print variable_name      # Print variable value
(dlv) goroutines               # List all goroutines
(dlv) exit                     # Exit debugger
```

### Debugging with Print Statements

For quick debugging, add log statements:

```go
package main

import (
    "log"
    "os"
)

func debugFunction() {
    log.Printf("DEBUG: variable value: %+v", variable)
    
    // Pretty print structs
    log.Printf("DEBUG: struct: %#v", structVariable)
    
    // Print with file and line info
    log.Printf("DEBUG [%s:%d]: message", "filename.go", 123)
}

func init() {
    // Enable debug logging during development
    if os.Getenv("DEBUG") == "true" {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }
}
```

Run with debug logging:
```bash
DEBUG=true go run main.go bootstrap
```

## Testing Your Changes

### Manual Testing

Create test scenarios to verify your changes:

```bash
# Test bootstrap functionality
go run main.go bootstrap --mode=oss-tenant --non-interactive

# Test cluster operations
go run main.go cluster create test-cluster
go run main.go cluster status test-cluster
go run main.go cluster delete test-cluster

# Test chart operations
go run main.go chart install test-app --repo=https://charts.example.com
go run main.go chart list

# Test development tools
go run main.go dev scaffold my-service --template=microservice
go run main.go dev intercept my-service --port=3000:8080
```

### Integration Testing

Test with real Kubernetes clusters:

```bash
# Create test cluster
k3d cluster create openframe-test --agents 2

# Run your changes against the cluster
KUBECONFIG=$(k3d kubeconfig write openframe-test) go run main.go bootstrap

# Verify results
kubectl get pods --all-namespaces
kubectl get applications -n argocd

# Clean up
k3d cluster delete openframe-test
```

### Performance Testing

Monitor resource usage and performance:

```bash
# Build optimized binary
go build -ldflags="-s -w" -o openframe main.go

# Monitor memory usage
/usr/bin/time -v ./openframe bootstrap

# Profile CPU usage
CPUPROFILE=cpu.prof go run main.go bootstrap
go tool pprof cpu.prof

# Profile memory usage
MEMPROFILE=mem.prof go run main.go bootstrap
go tool pprof mem.prof
```

## Code Quality

### Automated Checks

Run all quality checks before committing:

```bash
#!/bin/bash
# quality-check.sh

echo "Running code quality checks..."

# Format code
echo "Formatting code..."
gofmt -l -w .
goimports -l -w .

# Vet code
echo "Vetting code..."
go vet ./...

# Run linter
echo "Running linter..."
golangci-lint run

# Run tests
echo "Running tests..."
go test -race -cover ./...

# Check for security issues
echo "Checking security..."
gosec ./...

# Check dependencies
echo "Checking dependencies..."
go mod tidy
go mod verify

echo "All checks passed!"
```

Make it executable and run:
```bash
chmod +x quality-check.sh
./quality-check.sh
```

### Manual Code Review

Before submitting changes, review:

1. **Code Structure**: Is the code well-organized and follows Go conventions?
2. **Error Handling**: Are errors properly handled and user-friendly?
3. **Documentation**: Are public functions and packages documented?
4. **Tests**: Are there adequate unit and integration tests?
5. **Performance**: Are there any obvious performance issues?

## Advanced Development

### Working with Dependencies

```bash
# Add a new dependency
go get github.com/new/dependency@latest

# Update dependencies
go get -u ./...

# Vendor dependencies (if needed)
go mod vendor

# Remove unused dependencies
go mod tidy
```

### Working with Build Tags

Use build tags for conditional compilation:

```go
// +build debug

package debug

func init() {
    // Debug-only initialization
}
```

```bash
# Build with debug tag
go build -tags debug -o openframe-debug main.go

# Run tests with integration tag
go test -tags integration ./...
```

### Cross-platform Development

Build for multiple platforms:

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o openframe-linux main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o openframe.exe main.go

# Build for macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o openframe-darwin-amd64 main.go

# Build for macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o openframe-darwin-arm64 main.go
```

## Troubleshooting

### Common Development Issues

#### Module Problems
```bash
# Clear module cache
go clean -modcache

# Reinitialize modules
rm go.sum
go mod tidy
```

#### Build Errors
```bash
# Clean build cache
go clean -cache

# Rebuild everything
go build -a main.go
```

#### Test Failures
```bash
# Run tests with verbose output
go test -v -race ./...

# Run specific failing test
go test -v -run TestSpecificFunction ./path/to/package
```

#### Kubernetes Context Issues
```bash
# Check current context
kubectl config current-context

# Switch to correct context
kubectl config use-context k3d-openframe-local

# Verify cluster connectivity
kubectl cluster-info
```

### Debug Environment Variables

Set these for debugging:

```bash
export GODEBUG="gctrace=1"          # GC tracing
export GOTRACEBACK="all"            # Full stack traces
export OPENFRAME_LOG_LEVEL="debug"  # Detailed logging
export KUBECONFIG="$HOME/.kube/config"
```

## Next Steps

Now that you have a working development environment:

1. **[Architecture Overview](../architecture/README.md)** - Understand the system design
2. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the contribution process
3. **[Testing Guide](../testing/README.md)** - Deep dive into testing strategies

## Getting Help

If you encounter issues:

- **Check existing issues**: Search GitHub issues for similar problems
- **Ask in Slack**: Join the [OpenMSP community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Review documentation**: Check other guides in this repository
- **Debug systematically**: Use logging and debugging tools to isolate issues

Happy coding! 🚀