# Local Development Guide

This guide walks you through setting up OpenFrame CLI for local development, including building from source, running with hot reload, and debugging configurations.

## Prerequisites

Before starting local development, ensure you have:

1. ✅ **Completed** [Environment Setup](./environment.md) - Development tools configured
2. ✅ **Verified** [Prerequisites](../../getting-started/prerequisites.md) - Docker, kubectl, etc. installed
3. ✅ **Access** to the OpenFrame CLI repository

## Repository Setup

### 1. Clone and Initialize

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Set up development environment
make dev-setup

# Verify the setup
make test
```

### 2. Project Structure Overview

```text
openframe-cli/
├── cmd/                    # Command definitions
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/            # Cluster management commands  
│   ├── chart/              # Chart management commands
│   └── dev/                # Development tools
├── internal/               # Internal packages
│   ├── bootstrap/          # Bootstrap business logic
│   ├── cluster/            # Cluster management
│   │   ├── models/         # Data structures
│   │   ├── services/       # Business logic
│   │   ├── ui/             # User interface
│   │   └── utils/          # Utilities
│   ├── chart/              # Chart management logic
│   ├── dev/                # Development tools logic
│   └── shared/             # Shared components
├── tests/                  # Test files
├── scripts/                # Development scripts
├── Makefile               # Build automation
├── go.mod                 # Go module definition
└── main.go                # Application entry point
```

### 3. Understanding the Module Structure

```bash
# View module dependencies
go mod graph

# Check module information
go list -m all

# Verify module consistency
go mod tidy
go mod verify
```

## Building from Source

### 1. Basic Build

```bash
# Build binary
make build

# Or use go directly
go build -o bin/openframe .

# Verify the build
./bin/openframe --version
./bin/openframe --help
```

### 2. Development Build with Debug Symbols

```bash
# Build with debug information
go build -gcflags="all=-N -l" -o bin/openframe-debug .

# Build with race detection (for testing)
go build -race -o bin/openframe-race .
```

### 3. Cross-Platform Builds

```bash
# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/openframe-linux .
GOOS=darwin GOARCH=amd64 go build -o bin/openframe-darwin .
GOOS=windows GOARCH=amd64 go build -o bin/openframe-windows.exe .

# Or use the Makefile target (if available)
make build-all
```

## Running Locally

### 1. Basic Local Execution

```bash
# Run directly with go
go run . --help

# Run specific commands
go run . cluster --help
go run . bootstrap --help

# Run with arguments
go run . cluster create test-cluster --skip-wizard
```

### 2. Using Built Binary

```bash
# Build and run
make build
./bin/openframe cluster create dev-cluster

# Test bootstrap functionality
./bin/openframe bootstrap test-env --deployment-mode=oss-tenant --verbose
```

### 3. Environment-Specific Configuration

```bash
# Set development environment variables
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=./dev-config

# Run with development settings
./bin/openframe cluster create dev-cluster --verbose
```

## Hot Reload Development

### 1. Using Air (Recommended)

Install and configure Air for automatic rebuilds:

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Create Air configuration (if not exists)
cat > .air.toml <<'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["cluster", "status"]
  bin = "./tmp/openframe"
  cmd = "go build -o ./tmp/openframe ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "bin"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
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
  clean_on_exit = true
EOF

# Start development with hot reload
air
```

### 2. Manual Watch Script

Create a simple watch script if Air is not available:

```bash
#!/bin/bash
# scripts/watch.sh

echo "Starting OpenFrame development watch..."

# Function to build and run
build_and_run() {
    echo "Building..."
    if go build -o tmp/openframe .; then
        echo "Build successful. Running with args: $@"
        ./tmp/openframe "$@"
    else
        echo "Build failed."
    fi
}

# Watch for changes
while true; do
    inotifywait -e modify,create,delete -r . --exclude='(tmp/|\.git/|bin/)' -q
    echo "Changes detected, rebuilding..."
    build_and_run "$@"
done
```

```bash
# Make executable and run
chmod +x scripts/watch.sh
./scripts/watch.sh cluster status
```

## Debugging Configuration

### 1. VS Code Debugging

Ensure you have `.vscode/launch.json` configured:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Bootstrap",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["bootstrap", "debug-cluster", "--verbose", "--deployment-mode=oss-tenant"],
            "console": "integratedTerminal",
            "env": {
                "OPENFRAME_DEV_MODE": "true",
                "OPENFRAME_LOG_LEVEL": "debug"
            }
        },
        {
            "name": "Debug Cluster Create",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["cluster", "create", "debug-cluster", "--skip-wizard", "--verbose"],
            "console": "integratedTerminal"
        },
        {
            "name": "Debug Current File Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${file}",
            "console": "integratedTerminal"
        }
    ]
}
```

### 2. Command Line Debugging with Delve

```bash
# Install delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug bootstrap command
dlv debug -- bootstrap test-cluster --verbose

# Debug with breakpoint at main
dlv debug --build-flags="-tags=debug" -- cluster create test

# Attach to running process (if needed)
dlv attach <pid>
```

### 3. Debug Session Example

```bash
# Start debug session
dlv debug -- cluster create debug-cluster --verbose

# In delve prompt:
(dlv) break main.main
(dlv) break internal/cluster/services.CreateCluster
(dlv) continue
(dlv) print args
(dlv) step
(dlv) exit
```

## Testing During Development

### 1. Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/cluster/...
go test ./cmd/bootstrap/...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 2. Integration Testing

```bash
# Run integration tests (if available)
go test -tags=integration ./tests/...

# Test specific functionality end-to-end
./bin/openframe cluster create integration-test --skip-wizard
kubectl get nodes
./bin/openframe cluster delete integration-test
```

### 3. Manual Testing Workflow

```bash
# 1. Build latest version
make build

# 2. Test cluster creation
./bin/openframe cluster create manual-test --verbose

# 3. Verify cluster
kubectl config current-context
kubectl get nodes
kubectl get pods -A

# 4. Test chart installation
./bin/openframe chart install manual-test

# 5. Verify ArgoCD
kubectl get pods -n argocd
curl -k https://localhost:8080

# 6. Clean up
./bin/openframe cluster delete manual-test
```

## Development Workflows

### 1. Feature Development Workflow

```bash
# 1. Create feature branch
git checkout -b feature/new-command

# 2. Set up hot reload for development
air -- cluster status

# 3. Make changes and test automatically
# (Air will rebuild and run tests on file changes)

# 4. Run comprehensive tests
make test
make lint

# 5. Test with real cluster
make build
./bin/openframe cluster create feature-test
# Test your feature...
./bin/openframe cluster delete feature-test

# 6. Commit changes
git add .
git commit -m "feat: add new cluster command"
```

### 2. Bug Fix Workflow

```bash
# 1. Reproduce the bug
./bin/openframe <command-that-fails>

# 2. Create test case
cat > internal/cluster/bug_test.go <<'EOF'
func TestBugFix(t *testing.T) {
    // Test case that reproduces the bug
}
EOF

# 3. Verify test fails
go test ./internal/cluster/ -run TestBugFix

# 4. Debug the issue
dlv debug -- <command-that-fails>

# 5. Fix and verify
go test ./internal/cluster/ -run TestBugFix
make test
```

### 3. Documentation Development

```bash
# 1. Start documentation server (if available)
make docs-serve

# 2. Edit documentation
code docs/development/setup/local-development.md

# 3. Test code examples
./bin/openframe cluster create doc-test
# Verify examples work...
./bin/openframe cluster delete doc-test
```

## Performance Profiling

### 1. CPU Profiling

```bash
# Build with profiling
go build -o bin/openframe-profile .

# Run with CPU profiling
./bin/openframe-profile bootstrap test-cluster --deployment-mode=oss-tenant &
PID=$!

# Capture profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Analyze profile
go tool pprof -top profile.pb.gz
go tool pprof -web profile.pb.gz
```

### 2. Memory Profiling

```bash
# Capture memory profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Or during specific operation
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Common Development Tasks

### 1. Adding a New Command

```bash
# 1. Create command file
touch cmd/cluster/new-command.go

# 2. Implement command structure
cat > cmd/cluster/new-command.go <<'EOF'
package cluster

import (
    "github.com/spf13/cobra"
)

func getNewCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "new",
        Short: "New cluster command",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
    return cmd
}
EOF

# 3. Register in main command
# Edit cmd/cluster/cluster.go to add the subcommand

# 4. Test the new command
go run . cluster new --help
```

### 2. Adding New Business Logic

```bash
# 1. Create service file
mkdir -p internal/cluster/services
touch internal/cluster/services/new-service.go

# 2. Write tests first
touch internal/cluster/services/new-service_test.go

# 3. Implement and test
go test ./internal/cluster/services/...
```

### 3. Updating Dependencies

```bash
# Update specific dependency
go get -u github.com/spf13/cobra

# Update all dependencies
go get -u ./...

# Tidy module
go mod tidy

# Verify everything still works
make test
make build
```

## Troubleshooting Development Issues

### 1. Build Issues

```bash
# Clean build cache
go clean -cache
go clean -modcache

# Rebuild dependencies
go mod download
go mod tidy

# Check for version conflicts
go list -m -versions github.com/spf13/cobra
```

### 2. Import Issues

```bash
# Fix imports
goimports -w .

# Check for unused imports
go mod tidy

# Verify module structure
go list -m all
```

### 3. Test Issues

```bash
# Run tests in verbose mode
go test -v ./...

# Run specific test
go test -run TestSpecificTest ./internal/cluster/

# Check test coverage
go test -cover ./...
```

### 4. Runtime Issues

```bash
# Enable debug logging
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_DEV_MODE=true

# Check cluster state
kubectl config current-context
kubectl get nodes
k3d cluster list

# Clean up test resources
k3d cluster delete --all
docker system prune -f
```

## Next Steps

Now that you have local development set up:

1. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure
2. **[Testing Overview](../testing/overview.md)** - Learn testing practices and patterns
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Follow contribution standards

## Development Best Practices

### Code Organization

- Keep commands in `cmd/` directory
- Put business logic in `internal/` packages
- Use interfaces for testability
- Follow Go naming conventions

### Testing Strategy

- Write tests before implementing features
- Use table-driven tests for multiple scenarios
- Mock external dependencies
- Test error conditions

### Git Workflow

- Create feature branches for new work
- Write descriptive commit messages
- Test thoroughly before pushing
- Use pull requests for code review

Happy coding! 🚀 Your local development environment is ready for building awesome features in OpenFrame CLI.