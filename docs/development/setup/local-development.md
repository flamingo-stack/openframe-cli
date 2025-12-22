# Local Development Guide

This guide covers cloning the OpenFrame CLI repository, building the project locally, setting up hot reload for development, and configuring debugging tools.

## Repository Setup

### Clone the Repository

```bash
# Clone the main repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Set up upstream remote for contributors
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
git fetch upstream
```

### Fork Setup (for Contributors)

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
git fetch upstream

# Create a development branch
git checkout -b feature/my-new-feature
```

## Project Structure Overview

Understanding the codebase organization:

```text
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── bootstrap/         # Bootstrap command implementation
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   ├── dev/               # Development tools commands
│   └── root.go            # Root command and CLI setup
├── internal/              # Private application logic
│   ├── bootstrap/         # Bootstrap service logic
│   ├── cluster/           # Cluster management services
│   ├── chart/             # Chart installation services
│   ├── dev/               # Development tools services
│   └── shared/            # Shared utilities and infrastructure
├── tests/                 # Test suites and utilities
├── scripts/               # Build and development scripts
├── deployments/           # Deployment configurations
├── Makefile               # Build automation
└── go.mod                 # Go module definition
```

## Building the Project

### Initial Setup

```bash
# Install dependencies
go mod download

# Generate code (mocks, etc.)
make generate

# Build the CLI
make build
```

The binary will be created at `./bin/openframe`.

### Build Targets

| Command | Purpose |
|---------|---------|
| `make build` | Build for current platform |
| `make build-all` | Build for all supported platforms |
| `make build-dev` | Build with development flags |
| `make clean` | Clean build artifacts |
| `make install` | Build and install to `$GOPATH/bin` |

### Verify Build

```bash
# Test the built binary
./bin/openframe --version
./bin/openframe --help

# Run a simple command
./bin/openframe cluster list
```

## Development Workflow

### Hot Reload Setup

For rapid development, use Go's built-in tools for automatic rebuilding:

#### Option 1: Using `go run`
```bash
# Run directly from source
go run main.go cluster create --name dev-test

# With environment variables
OPENFRAME_LOG_LEVEL=debug go run main.go bootstrap --verbose
```

#### Option 2: Using Air (Recommended)
Air provides automatic rebuilding when files change:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["cluster", "list"]
  bin = "tmp/openframe"
  cmd = "go build -o tmp/openframe ."
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

# Start hot reload
air
```

#### Option 3: Using Makefile Watch Target
```bash
# Start file watcher with make
make watch

# This runs:
# fswatch -o . -e ".*\.git.*" | xargs -n1 -I{} make build
```

### Development Commands

#### Running Tests During Development
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test packages
go test ./internal/cluster/...

# Run tests with verbose output
go test -v ./...

# Run tests continuously
make test-watch
```

#### Code Quality Checks
```bash
# Run linters
make lint

# Format code
make fmt

# Check for security issues
make security

# Generate missing code
make generate
```

### Debugging Configuration

#### VS Code Debugging

Create `.vscode/launch.json`:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--mode", "oss-tenant", "--verbose"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug",
        "OPENFRAME_DEV_MODE": "true"
      },
      "console": "integratedTerminal",
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug Cluster Create",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "--name", "debug-cluster"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.run", "TestClusterCreate"],
      "showLog": true
    }
  ]
}
```

#### Command Line Debugging
```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o bin/openframe-debug main.go

# Run with delve debugger
dlv exec ./bin/openframe-debug -- cluster create --name debug

# Debug a specific test
dlv test ./internal/cluster -- -test.run TestClusterService
```

#### GoLand/IntelliJ Debugging
1. Open Run/Debug Configurations
2. Add new "Go Application" configuration
3. Set program arguments and environment variables
4. Set breakpoints and run

## Local Testing Setup

### Test Cluster Management

Create test clusters for development:

```bash
# Create dedicated test cluster
k3d cluster create openframe-dev \
  --api-port 6445 \
  --port "8080:80@loadbalancer" \
  --agents 1

# Use test cluster
kubectl config use-context k3d-openframe-dev

# Clean up when done
k3d cluster delete openframe-dev
```

### Mock Development

Enable mocks for faster development:

```bash
# Set environment for mock providers
export OPENFRAME_MOCK_PROVIDERS=true
export OPENFRAME_DEV_MODE=true

# Run with mocks
go run main.go cluster create --name mock-cluster
```

### Integration Test Setup

```bash
# Create integration test environment
make test-env-setup

# Run integration tests
make test-integration

# Cleanup test environment
make test-env-cleanup
```

## Development Scripts

### Useful Scripts

Create these helper scripts in your local environment:

#### `scripts/dev-cluster.sh`
```bash
#!/bin/bash
set -e

CLUSTER_NAME=${1:-openframe-dev}

echo "🚀 Creating development cluster: $CLUSTER_NAME"

k3d cluster create $CLUSTER_NAME \
  --api-port 6445 \
  --port "8080:80@loadbalancer" \
  --port "8443:443@loadbalancer" \
  --agents 1 \
  --wait

echo "✅ Cluster $CLUSTER_NAME ready!"
echo "💡 Use: kubectl config use-context k3d-$CLUSTER_NAME"
```

#### `scripts/quick-test.sh`
```bash
#!/bin/bash
# Quick test and build script

set -e

echo "🧪 Running tests..."
make test

echo "🔨 Building binary..."
make build

echo "✅ Testing built binary..."
./bin/openframe --version

echo "🎉 Ready for development!"
```

#### `scripts/reset-dev.sh`
```bash
#!/bin/bash
# Reset development environment

set -e

echo "🗑️  Cleaning up..."
k3d cluster delete openframe-dev 2>/dev/null || true
make clean

echo "🔄 Resetting..."
make build
./scripts/dev-cluster.sh

echo "🎯 Development environment reset!"
```

## Common Development Tasks

### Adding a New Command

1. **Create command file** in `cmd/` directory:
```bash
mkdir -p cmd/mynewcommand
touch cmd/mynewcommand/mynewcommand.go
```

2. **Implement command structure**:
```go
package mynewcommand

import (
    "github.com/spf13/cobra"
)

func GetMyNewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "mynew",
        Short: "Description of my new command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Command logic here
            return nil
        },
    }
    
    return cmd
}
```

3. **Add to root command** in `cmd/root.go`:
```go
rootCmd.AddCommand(getMyNewCmd())
```

4. **Create service layer** in `internal/mynewcommand/`
5. **Add tests** in `tests/`

### Adding a New Provider

1. **Define interface** in `internal/shared/interfaces/`
2. **Implement provider** in appropriate service directory
3. **Add mock** using `mockgen`
4. **Write tests** with mocked dependencies

### Debugging Common Issues

#### Build Issues
```bash
# Clean and rebuild
make clean
go mod tidy
make build
```

#### Test Failures
```bash
# Run specific failing test with verbose output
go test -v -run TestSpecificFunction ./internal/package

# Check for race conditions
go test -race ./...
```

#### Import Issues
```bash
# Update dependencies
go get -u ./...
go mod tidy
```

## Performance Profiling

### CPU Profiling
```bash
# Build with profiling
go build -o bin/openframe-profile main.go

# Run with CPU profiling
./bin/openframe-profile cluster create --cpuprofile=cpu.prof

# Analyze profile
go tool pprof cpu.prof
```

### Memory Profiling
```bash
# Run with memory profiling
./bin/openframe-profile bootstrap --memprofile=mem.prof

# Analyze memory usage
go tool pprof mem.prof
```

## Git Workflow for Development

### Feature Development
```bash
# Create feature branch
git checkout -b feature/my-awesome-feature

# Make changes and commit regularly
git add .
git commit -m "feat: add awesome new feature"

# Push and create PR
git push origin feature/my-awesome-feature
```

### Keeping Branch Updated
```bash
# Fetch latest changes
git fetch upstream

# Rebase on main
git rebase upstream/main

# Force push if needed (after rebase)
git push --force-with-lease origin feature/my-awesome-feature
```

## What's Next?

Now that you have local development running:

1. **[Understand the architecture](../architecture/overview.md)** - Learn how components interact
2. **[Explore testing strategies](../testing/overview.md)** - Write effective tests  
3. **[Follow contributing guidelines](../contributing/guidelines.md)** - Submit quality contributions

> **💡 Pro Tip**: Use the `make help` command to see all available build targets and development commands.