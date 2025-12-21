# Local Development Guide

This guide walks you through cloning the OpenFrame CLI repository, setting up your local development environment, building the project, and establishing an efficient development workflow.

## Repository Structure

Before we start, let's understand what you'll be working with:

```
openframe-cli/
├── cmd/                    # Command definitions (CLI interface)
│   ├── bootstrap/          # Bootstrap command
│   ├── cluster/           # Cluster management commands  
│   ├── chart/             # Chart installation commands
│   └── dev/               # Development workflow commands
├── internal/              # Internal packages (business logic)
│   ├── bootstrap/         # Bootstrap service implementation
│   ├── cluster/           # Cluster management services
│   ├── chart/             # Chart management services
│   ├── dev/               # Development workflow services
│   └── shared/            # Shared utilities and components
├── docs/                  # Documentation
├── scripts/               # Build and automation scripts
├── .github/               # GitHub workflows and templates
├── Makefile              # Build automation
├── go.mod                # Go module definition
└── main.go               # Application entry point (if present)
```

## Clone and Initial Setup

### 1. Fork and Clone the Repository

```bash
# Fork the repository on GitHub first, then clone your fork
git clone https://github.com/YOUR_USERNAME/openframe-cli.git
cd openframe-cli

# Add upstream remote for syncing
git remote add upstream https://github.com/flamingo-stack/openframe-cli.git
git remote -v
```

### 2. Initial Environment Setup

```bash
# Ensure you have the required Go version
go version  # Should be 1.23+

# Download dependencies
go mod download

# Verify dependencies
go mod verify
```

### 3. Install Development Dependencies

```bash
# Install development tools (if not already installed)
make dev-setup

# Or install manually:
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest
go install github.com/go-delve/delve/cmd/dlv@latest
```

## Building the Project

### Build Commands

The project uses a Makefile for consistent build operations:

```bash
# View available build targets
make help

# Build the binary
make build

# Build with development flags (includes debug info)
make dev-build

# Clean build artifacts
make clean
```

### Manual Build

You can also build manually with Go:

```bash
# Basic build
go build -o openframe .

# Build with version info
go build -ldflags="-X main.version=$(git describe --tags --always)" -o openframe .

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o openframe-linux .
GOOS=darwin GOARCH=amd64 go build -o openframe-macos .
GOOS=windows GOARCH=amd64 go build -o openframe.exe .
```

### Verify Your Build

```bash
# Test the built binary
./openframe --help
./openframe --version

# Test basic commands
./openframe cluster --help
./openframe bootstrap --help
```

Expected output:
```
OpenFrame CLI - Kubernetes cluster management and development workflows

Usage:
  openframe [command]

Available Commands:
  bootstrap   Bootstrap complete OpenFrame environment
  cluster     Manage Kubernetes clusters
  chart       Manage Helm charts and GitOps deployments  
  dev         Development workflow tools
  completion  Generate completion script
  help        Help about any command

Flags:
  -h, --help      help for openframe
  -v, --version   version for openframe
```

## Running Locally

### Development Mode

For development, you can run the CLI directly with Go:

```bash
# Run without building
go run . --help
go run . cluster --help
go run . bootstrap --help

# Run with debug flags
OPENFRAME_DEBUG=true go run . cluster status

# Run specific commands for testing
go run . cluster create test-cluster --dry-run
go run . bootstrap --help
```

### Hot Reload Development

For rapid development iteration, set up file watching:

#### Option 1: Using `entr` (Unix/Linux/macOS)

```bash
# Install entr
# macOS: brew install entr
# Linux: apt-get install entr

# Watch Go files and rebuild on changes
find . -name "*.go" | entr -r sh -c 'go build -o openframe-dev . && echo "Build complete"'
```

#### Option 2: Using `air` (Cross-platform)

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
cat > .air.toml << 'EOF'
root = "."
cmd = "go build -o openframe-dev ."
bin = "./openframe-dev"
include_ext = ["go", "yaml", "yml"]
exclude_dir = ["vendor", ".git", "docs"]
delay = 1000
stop_on_error = true
EOF

# Start hot reload development
air
```

### Debug Configuration

Set up debugging for your development environment:

#### VS Code Debugging

Create `.vscode/launch.json`:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug CLI",
      "type": "go",
      "request": "launch",
      "mode": "debug", 
      "program": "${workspaceFolder}",
      "args": ["--help"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      },
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug Bootstrap",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["bootstrap", "test-cluster", "--dry-run"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      }
    },
    {
      "name": "Debug Cluster Create",
      "type": "go", 
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["cluster", "create", "debug-cluster"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      }
    }
  ]
}
```

#### Command-line Debugging with Delve

```bash
# Debug the main application
dlv debug . -- --help

# Debug with arguments
dlv debug . -- cluster create test-cluster

# Debug a specific test
dlv test ./internal/cluster
```

## Development Workflow

### Typical Development Cycle

```bash
# 1. Sync with upstream
git fetch upstream
git checkout main  
git merge upstream/main

# 2. Create feature branch
git checkout -b feature/new-command

# 3. Make changes and test frequently
go run . cluster --help  # Test as you develop

# 4. Run tests
make test

# 5. Lint code  
make lint

# 6. Build and test binary
make build
./openframe cluster --help

# 7. Commit changes
git add .
git commit -m "feat: add new cluster command"

# 8. Push and create PR
git push origin feature/new-command
```

### Testing Your Changes

#### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Test specific packages
go test ./internal/cluster/...
go test ./cmd/bootstrap/...
```

#### Integration Testing
```bash
# Run integration tests (requires Docker)
make test-integration

# Test with a real cluster
./openframe cluster create test-dev
./openframe cluster status test-dev
./openframe cluster delete test-dev
```

#### Manual Testing Workflow
```bash
# Test bootstrap flow
./openframe bootstrap test-cluster --deployment-mode=oss-tenant --non-interactive

# Test cluster operations
./openframe cluster list
./openframe cluster status test-cluster

# Test cleanup
./openframe cluster cleanup
./openframe cluster delete test-cluster
```

### Code Quality

#### Linting and Formatting

```bash
# Format code
go fmt ./...
gofumpt -l -w .
goimports -w .

# Run linters
golangci-lint run
staticcheck ./...

# Check for common mistakes
go vet ./...
```

#### Security Scanning
```bash
# Install gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...
```

## Watch Mode for Development

Set up efficient file watching for rapid development:

### Using Make with Watch

Create a development target in your Makefile:

```makefile
.PHONY: dev-watch
dev-watch: ## Watch files and rebuild on changes
	@echo "Watching for changes..."
	@while true; do \
		find . -name "*.go" -newer .last-build 2>/dev/null | head -1 | grep -q . && { \
			echo "Changes detected, rebuilding..."; \
			touch .last-build; \
			$(MAKE) build; \
		}; \
		sleep 1; \
	done

.PHONY: dev-test-watch  
dev-test-watch: ## Watch files and run tests on changes
	@echo "Watching for test changes..."
	@while true; do \
		find . -name "*.go" -newer .last-test 2>/dev/null | head -1 | grep -q . && { \
			echo "Running tests..."; \
			touch .last-test; \
			go test ./...; \
		}; \
		sleep 1; \
	done
```

### Development Scripts

Create useful development scripts:

#### `scripts/dev.sh` - Development Helper
```bash
#!/bin/bash
set -e

case "$1" in
  "build")
    echo "🔨 Building..."
    go build -o openframe-dev .
    echo "✅ Build complete"
    ;;
  "test")
    echo "🧪 Running tests..."
    go test ./...
    echo "✅ Tests complete"
    ;;
  "lint")
    echo "🔍 Linting..."
    golangci-lint run
    echo "✅ Lint complete" 
    ;;
  "clean")
    echo "🧹 Cleaning..."
    rm -f openframe-dev
    go clean -cache -modcache -i -r
    echo "✅ Clean complete"
    ;;
  "reset")
    echo "🔄 Resetting development environment..."
    k3d cluster delete --all
    docker system prune -f
    echo "✅ Reset complete"
    ;;
  *)
    echo "Usage: $0 {build|test|lint|clean|reset}"
    exit 1
    ;;
esac
```

Make it executable:
```bash
chmod +x scripts/dev.sh
```

## Debugging Common Issues

### Build Issues

#### Module Issues
```bash
# Clean and reset modules
go clean -modcache
rm go.sum
go mod download
go mod tidy
```

#### Version Issues  
```bash
# Check Go version
go version  # Must be 1.23+

# Update Go if needed
sudo rm -rf /usr/local/go
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz
```

### Runtime Issues

#### Docker Issues
```bash
# Check Docker daemon
docker ps

# Restart Docker service
sudo systemctl restart docker  # Linux
# Or restart Docker Desktop on macOS/Windows
```

#### Permission Issues
```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker

# Fix kubectl permissions
chmod 600 ~/.kube/config
```

### Development Environment Issues

#### Go Tools Not Found
```bash
# Check GOPATH
go env GOPATH

# Add to PATH
export PATH=$PATH:$(go env GOPATH)/bin
echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc
```

#### IDE Issues
```bash
# Reset VS Code Go extension
# Command Palette -> Go: Install/Update Tools
# Select all and install

# Restart language server
# Command Palette -> Go: Restart Language Server
```

## Performance Optimization

### Build Performance
```bash
# Use build cache
export GOCACHE=$(go env GOCACHE)

# Parallel builds
go build -p 4 .

# Skip tests during development builds
go build -tags dev .
```

### Development Performance
```bash
# Use go install for faster iteration
go install .  # Installs to GOPATH/bin

# Use go run with caching
go run -gcflags="-N -l" .  # Disable optimizations for debugging
```

## Next Steps

Now that you have local development set up:

1. **Explore the Code** - Start with [Architecture Overview](../architecture/overview.md)
2. **Run Tests** - Learn about [Testing Strategies](../testing/overview.md)  
3. **Make Changes** - Follow [Contributing Guidelines](../contributing/guidelines.md)
4. **Submit PRs** - Review the contribution workflow

## Development Tips

### Useful Aliases for Development
```bash
# Add these to your shell profile
alias ofd='./openframe-dev'        # Run development build
alias ofb='go run . bootstrap'      # Quick bootstrap
alias ofc='go run . cluster'        # Quick cluster commands
alias oft='go test ./...'           # Run tests
alias ofl='golangci-lint run'       # Lint code
```

### Quick Commands Reference
```bash
# Development workflow
make build                   # Build binary
make test                   # Run tests
make lint                   # Lint code
go run . <command>          # Run without building
./openframe-dev <command>   # Run development build

# Testing
go test -v ./internal/cluster/...    # Test specific package
go test -run TestBootstrap ./...     # Run specific test
go test -cover ./...                 # Test with coverage

# Debugging  
dlv debug . -- cluster create test  # Debug with delve
OPENFRAME_DEBUG=true go run . <cmd>  # Debug mode
```

You're now ready for productive OpenFrame CLI development! 🚀

The next step is understanding the [architecture](../architecture/overview.md) to know how everything fits together.