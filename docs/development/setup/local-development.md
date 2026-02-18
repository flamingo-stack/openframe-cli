# Local Development Guide

This guide walks you through cloning, building, and running OpenFrame CLI locally for development. You'll learn the development workflow, debugging techniques, and testing strategies.

## Repository Setup

### Clone the Repository

```bash
# Clone the main repository
git clone https://github.com/flamingo-stack/openframe-oss-tenant.git
cd openframe-oss-tenant

# Verify repository structure
ls -la
```

**Expected structure:**
```text
openframe-oss-tenant/
├── cmd/              # CLI commands and entry points
├── internal/         # Internal packages and business logic
├── tests/            # Test files and utilities
├── go.mod           # Go module definition
├── go.sum           # Go module checksums
├── main.go          # Application entry point
└── README.md        # Project documentation
```

### Initialize Development Environment

```bash
# Verify Go module
go mod tidy

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### Set Up Git Configuration

```bash
# Configure Git for this repository
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Set up upstream remote (if you forked)
git remote add upstream https://github.com/flamingo-stack/openframe-oss-tenant.git
git fetch upstream
```

## Building the CLI

### Development Build

```bash
# Basic build
go build -o openframe .

# Build with version information
go build -ldflags "-X cmd.version=dev -X cmd.commit=$(git rev-parse HEAD) -X cmd.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o openframe .

# Verify build
./openframe --version
```

### Development Build Script

Create a `build-dev.sh` script for convenient development builds:

```bash
#!/bin/bash
# build-dev.sh

set -e

echo "🔨 Building OpenFrame CLI for development..."

# Get version info
VERSION=${VERSION:-"dev"}
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Build with version info
go build -ldflags "\
  -X cmd.version=$VERSION \
  -X cmd.commit=$COMMIT \
  -X cmd.date=$DATE" \
  -o openframe .

echo "✅ Build complete: ./openframe"
echo "📋 Version: $VERSION ($COMMIT)"

# Test the build
./openframe --version
```

Make it executable and use it:

```bash
chmod +x build-dev.sh
./build-dev.sh
```

### Cross-Platform Builds

Build for different platforms during development:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o openframe-linux-amd64 .

# macOS AMD64  
GOOS=darwin GOARCH=amd64 go build -o openframe-darwin-amd64 .

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o openframe-darwin-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o openframe-windows-amd64.exe .
```

## Running Locally

### Basic Usage

```bash
# Show help
./openframe --help

# Show version
./openframe --version

# Run with verbose output
./openframe --verbose cluster status
```

### Development Mode

Enable development mode for enhanced debugging:

```bash
# Set development environment variables
export OPENFRAME_DEV=1
export OPENFRAME_LOG_LEVEL=debug

# Run with development settings
./openframe bootstrap --dry-run
```

### Testing Commands Safely

Use dry-run and local modes for safe testing:

```bash
# Test bootstrap without creating resources
./openframe bootstrap --dry-run

# Test cluster creation with custom name
./openframe cluster create test-cluster --dry-run

# Test chart installation in local mode
./openframe chart install --mode=local --dry-run
```

## Development Workflow

### File Watching and Auto-Rebuild

#### Using `entr` (recommended)

Install `entr` for file watching:

```bash
# Linux (Ubuntu/Debian)
sudo apt install entr

# macOS
brew install entr

# Usage: Auto-rebuild on file changes
find . -name "*.go" | entr -r sh -c 'go build -o openframe . && echo "✅ Rebuild complete"'
```

#### Using `air` (Go-specific)

Install and configure `air` for Go development:

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Initialize air config
air init
```

Edit `.air.toml`:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
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
```

Run with air:

```bash
air
```

### Live Development Testing

Create a development testing script:

```bash
#!/bin/bash
# dev-test.sh

echo "🧪 Development Testing Script"
echo "============================="

# Build latest version
./build-dev.sh

echo ""
echo "🔍 Running basic tests..."

# Test version
echo "Version check:"
./openframe --version
echo ""

# Test help
echo "Help check:"
./openframe --help | head -10
echo ""

# Test prerequisite checking
echo "Prerequisites check:"
./openframe bootstrap --dry-run --verbose | head -20
echo ""

echo "✅ Development tests complete!"
```

## Debugging Techniques

### Using Delve Debugger

#### Command Line Debugging

```bash
# Install Delve if not already installed
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the main application
dlv debug . -- bootstrap --dry-run

# Debug with arguments
dlv debug . -- cluster create test --verbose

# Debug tests
dlv test ./internal/cluster
```

#### VS Code Debugging

Use the debug configuration from the environment setup:

1. Set breakpoints in VS Code
2. Press `F5` or use the debug panel
3. Select "Debug OpenFrame CLI" configuration
4. The debugger will start with your arguments

#### Debugging Specific Commands

Create focused debug configurations in `.vscode/launch.json`:

```json
{
  "name": "Debug Cluster Create",
  "type": "go",
  "request": "launch",
  "mode": "debug",
  "program": "${workspaceFolder}/main.go",
  "args": ["cluster", "create", "debug-cluster", "--verbose"],
  "cwd": "${workspaceFolder}",
  "env": {
    "OPENFRAME_DEV": "1",
    "OPENFRAME_LOG_LEVEL": "debug"
  },
  "console": "integratedTerminal"
}
```

### Logging and Tracing

#### Enhanced Logging

Add debug logging to your code:

```go
import (
    "log/slog"
    "os"
)

// Enable structured logging
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Use throughout the application
logger.Debug("Processing cluster creation", "name", clusterName, "config", config)
logger.Info("Cluster created successfully", "name", clusterName)
logger.Error("Failed to create cluster", "error", err)
```

#### Trace External Commands

Log all external command executions:

```go
// In internal/shared/executor/executor.go
func (e *Executor) Execute(cmd *exec.Cmd) error {
    if os.Getenv("OPENFRAME_DEV") == "1" {
        fmt.Printf("🔧 Executing: %s %v\n", cmd.Path, cmd.Args)
    }
    
    output, err := cmd.CombinedOutput()
    
    if os.Getenv("OPENFRAME_DEV") == "1" {
        fmt.Printf("📤 Output: %s\n", string(output))
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
        }
    }
    
    return err
}
```

## Testing During Development

### Unit Tests

Run tests with various options:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cluster/...

# Run specific test function
go test -run TestClusterCreate ./internal/cluster
```

### Integration Tests

Run integration tests against real resources:

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./tests/integration/...

# Run with verbose output
go test -v -tags=integration ./tests/integration/...

# Run specific integration test
go test -tags=integration -run TestBootstrapIntegration ./tests/integration/
```

### Test Coverage Analysis

Generate and view test coverage:

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Open in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Benchmark Tests

Run performance benchmarks:

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkClusterCreate ./internal/cluster
```

## Development Tools Integration

### Makefile for Development

Create a `Makefile` for common development tasks:

```makefile
.PHONY: build test clean install lint format help

# Variables
BINARY_NAME=openframe
VERSION?=dev
COMMIT?=$(shell git rev-parse --short HEAD)
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X cmd.version=$(VERSION) -X cmd.commit=$(COMMIT) -X cmd.date=$(DATE)"

# Default target
help: ## Show this help message
	@echo "OpenFrame CLI Development"
	@echo "========================="
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

build: ## Build the CLI binary
	@echo "🔨 Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "✅ Build complete: ./$(BINARY_NAME)"

test: ## Run all tests
	@echo "🧪 Running tests..."
	@go test -v -race ./...

test-integration: ## Run integration tests
	@echo "🧪 Running integration tests..."
	@go test -v -tags=integration ./tests/integration/...

coverage: ## Generate test coverage report
	@echo "📊 Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

lint: ## Run linter
	@echo "🔍 Running linter..."
	@golangci-lint run

format: ## Format code
	@echo "🎨 Formatting code..."
	@gofmt -s -w .
	@goimports -w .

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage.out coverage.html
	@rm -rf tmp/

install: build ## Install the CLI binary
	@echo "📦 Installing $(BINARY_NAME)..."
	@sudo cp $(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"

dev: ## Build and run in development mode
	@$(MAKE) build
	@OPENFRAME_DEV=1 OPENFRAME_LOG_LEVEL=debug ./$(BINARY_NAME) $(ARGS)

release: ## Build release binaries for all platforms
	@echo "🚀 Building release binaries..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✅ Release binaries created in dist/"
```

Usage examples:

```bash
# Build
make build

# Test
make test

# Development mode with arguments
make dev ARGS="cluster create test-cluster --dry-run"

# Full development cycle
make clean build test lint
```

### VS Code Tasks

Create `.vscode/tasks.json` for integrated development tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "type": "shell",
      "command": "go",
      "args": ["build", "-o", "openframe", "."],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "test",
      "type": "shell",
      "command": "go",
      "args": ["test", "-v", "./..."],
      "group": "test",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      },
      "problemMatcher": ["$go"]
    },
    {
      "label": "lint",
      "type": "shell",
      "command": "golangci-lint",
      "args": ["run"],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    }
  ]
}
```

Use tasks with `Ctrl+Shift+P` → "Tasks: Run Task".

## Hot Reload Development

### Air Configuration for OpenFrame CLI

Advanced `.air.toml` configuration for CLI development:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["--help"]
  bin = "./tmp/openframe"
  cmd = "go build -o ./tmp/openframe ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "dist", ".git"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = "OPENFRAME_DEV=1 OPENFRAME_LOG_LEVEL=debug ./tmp/openframe"
  include_dir = []
  include_ext = ["go"]
  kill_delay = "0s"
  log = "tmp/build-errors.log"
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
  clean_on_exit = true

[screen]
  clear_on_rebuild = true
  keep_scroll = true
```

## Troubleshooting Development Issues

### Go Module Issues

```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download

# Tidy up dependencies
go mod tidy

# Verify module integrity
go mod verify
```

### Build Issues

```bash
# Clean and rebuild
rm -f openframe
go clean -cache
go build -v .

# Check for import cycles
go list -json . | jq '.Deps'

# Verbose build output
go build -x .
```

### Debug Configuration Issues

```bash
# Check environment variables
env | grep OPENFRAME

# Test with minimal config
unset OPENFRAME_DEV
unset OPENFRAME_LOG_LEVEL
./openframe --version
```

### External Tool Integration Issues

```bash
# Test external tool availability
which k3d kubectl helm docker

# Test Docker connectivity
docker ps
docker run hello-world

# Test Kubernetes connectivity
kubectl version --client
kubectl cluster-info
```

## Next Steps

After setting up local development:

1. **[Architecture Overview](../architecture/README.md)** - Understand the codebase structure
2. **[Testing Guide](../testing/README.md)** - Learn testing strategies and best practices  
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Understand the contribution workflow

## Development Best Practices

1. **Incremental Development**: Make small, testable changes
2. **Test-Driven Development**: Write tests before implementing features
3. **Debug Early**: Use debugger instead of print statements
4. **Version Everything**: Use meaningful git commit messages
5. **Document Decisions**: Comment complex logic and architectural choices

The local development setup provides a solid foundation for contributing to OpenFrame CLI. The tools and workflows described here will help you develop efficiently and maintain code quality throughout the development process.