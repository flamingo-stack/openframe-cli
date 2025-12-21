# Development Environment Setup

This guide will help you set up a complete development environment for contributing to OpenFrame CLI. Whether you're fixing bugs, adding features, or improving documentation, this setup will ensure you have all the tools and configurations needed for productive development.

## Prerequisites

Before setting up your development environment, ensure you meet these requirements:

### System Requirements

| Requirement | Minimum | Recommended | Notes |
|-------------|---------|-------------|--------|
| **OS** | Linux, macOS, Windows (WSL2) | macOS/Linux | Windows requires WSL2 |
| **RAM** | 8GB | 16GB+ | For running local clusters |
| **CPU** | 2 cores | 4+ cores | Faster builds and tests |
| **Disk** | 20GB free | 50GB+ | For Go cache, Docker images |

### Required Software

Make sure you have these tools installed. If not, see the [Prerequisites Guide](../../getting-started/prerequisites.md):

- **Docker** (20.10+)
- **kubectl** (1.24+)
- **Helm** (3.8+)
- **K3d** (5.4+)
- **Git** (2.30+)

## Go Development Environment

### Install Go

**macOS:**
```bash
# Using Homebrew
brew install go

# Or download from https://golang.org/dl/
```

**Linux:**
```bash
# Download and install Go 1.19+
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH in ~/.bashrc or ~/.zshrc
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

**Windows (WSL2):**
```bash
# Same as Linux within WSL2 environment
```

### Verify Go Installation
```bash
go version
go env GOPATH
```

Expected output:
```
go version go1.21.0 linux/amd64
/home/username/go
```

### Configure Go Environment
```bash
# Set up Go workspace
mkdir -p $GOPATH/src/github.com/flamingo-stack

# Configure Go modules (already default in Go 1.16+)
go env -w GO111MODULE=on

# Enable Go module proxy (for faster downloads)
go env -w GOPROXY=https://proxy.golang.org,direct

# Configure private modules if working with private repos
# go env -w GOPRIVATE=github.com/flamingo-stack/*
```

## IDE and Editor Setup

### VS Code (Recommended)

**Install VS Code:**
```bash
# macOS
brew install --cask visual-studio-code

# Linux (Ubuntu/Debian)
sudo snap install code --classic

# Or download from https://code.visualstudio.com/
```

**Essential Extensions:**
```bash
# Install via command line
code --install-extension golang.go
code --install-extension ms-vscode.vscode-go
code --install-extension redhat.vscode-yaml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension eamodio.gitlens
code --install-extension esbenp.prettier-vscode
```

**VS Code Configuration:**

Create `.vscode/settings.json` in your workspace:
```json
{
    "go.toolsManagement.autoUpdate": true,
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.formatTool": "gofmt",
    "go.testFlags": ["-v", "-count=1"],
    "go.testTimeout": "30s",
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter"
    },
    "yaml.schemas": {
        "kubernetes": "*.yaml"
    },
    "files.associations": {
        "*.yaml": "yaml",
        "*.yml": "yaml"
    }
}
```

**VS Code Workspace Configuration:**

Create `.vscode/launch.json` for debugging:
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Bootstrap",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "args": ["bootstrap", "--verbose", "debug-cluster"],
            "env": {
                "OPENFRAME_LOG_LEVEL": "debug"
            }
        },
        {
            "name": "Debug Cluster Create",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}",
            "args": ["cluster", "create", "test-cluster", "--verbose"]
        }
    ]
}
```

### Alternative IDEs

**GoLand (JetBrains):**
```bash
# Download from https://www.jetbrains.com/go/
# Built-in Go support, excellent debugging, refactoring tools
```

**Vim/Neovim with vim-go:**
```bash
# Install vim-go plugin
# Add to ~/.vimrc:
# Plug 'fatih/vim-go'
```

## Development Tools

### Go Development Tools

Install essential Go development tools:

```bash
# Go language server
go install golang.org/x/tools/gopls@latest

# Linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Security checker
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Test coverage tool
go install golang.org/x/tools/cmd/cover@latest

# Dependency visualization
go install golang.org/x/exp/cmd/modgraphviz@latest

# Code generation
go install golang.org/x/tools/cmd/stringer@latest

# Delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest

# Mock generator (for testing)
go install github.com/golang/mock/mockgen@latest
```

### Build Tools

**Make (for build automation):**
```bash
# macOS
brew install make

# Linux
sudo apt-get install build-essential

# Windows (WSL2)
sudo apt-get install build-essential
```

**Air (for hot reloading during development):**
```bash
go install github.com/cosmtrek/air@latest
```

Create `.air.toml` for hot reloading configuration:
```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
args_bin = ["bootstrap", "--verbose"]
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

### Testing Tools

**Test Coverage Visualization:**
```bash
# Install gocov and gocov-html
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html@latest

# Generate HTML coverage report
go test -coverprofile=coverage.out ./...
gocov convert coverage.out | gocov-html > coverage.html
```

**Test Runners:**
```bash
# Install gotestsum for better test output
go install gotest.tools/gotestsum@latest

# Run tests with better formatting
gotestsum --format testname
```

## Project Configuration

### Git Configuration

Configure Git for the project:

```bash
# Set up Git user (if not already done)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Configure GPG signing (recommended for contributors)
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID

# Set up useful Git aliases
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.unstage 'reset HEAD --'
git config --global alias.last 'log -1 HEAD'
git config --global alias.visual '!gitk'
```

### Pre-commit Hooks

Set up pre-commit hooks to ensure code quality:

**Install pre-commit:**
```bash
# macOS
brew install pre-commit

# Linux
pip install pre-commit

# Or using Go
go install github.com/pre-commit/pre-commit@latest
```

**Create `.pre-commit-config.yaml`:**
```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
  
  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        language: system
        args: [-w]
        files: \.go$
      
      - id: go-vet
        name: go vet
        entry: go vet
        language: system
        files: \.go$
        pass_filenames: false
      
      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run
        language: system
        files: \.go$
        pass_filenames: false
```

**Install hooks:**
```bash
pre-commit install
```

### Environment Variables

Create a `.env` file for development:

```bash
# Development environment variables
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_DEFAULT_CLUSTER=dev-cluster
export OPENFRAME_DEPLOYMENT_MODE=oss-tenant

# Development flags
export OPENFRAME_DEV_MODE=true
export OPENFRAME_SKIP_PREREQUISITES=false

# Tool paths (if needed)
export K3D_PATH=/usr/local/bin/k3d
export HELM_PATH=/usr/local/bin/helm
export KUBECTL_PATH=/usr/local/bin/kubectl

# CI/CD settings for local testing
export CI=false
export GITHUB_ACTIONS=false
```

Load environment variables:
```bash
# Add to ~/.bashrc or ~/.zshrc
source /path/to/openframe-cli/.env

# Or use direnv for automatic loading
echo 'source .env' > .envrc
direnv allow
```

## Development Workflow Setup

### Makefile

Create a `Makefile` for common development tasks:

```makefile
.PHONY: build test lint clean install deps dev

# Variables
BINARY_NAME=openframe
VERSION=dev
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

# Default target
all: clean deps lint test build

# Build the binary
build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint code
lint:
	golangci-lint run

# Security check
security:
	gosec ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Install the binary locally
install: build
	sudo mv $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Development server with hot reload
dev:
	air

# Generate mocks for testing
mocks:
	mockgen -source=internal/cluster/interfaces.go -destination=internal/cluster/mocks/mocks.go

# Format code
fmt:
	gofmt -w $(GO_FILES)

# Update dependencies
update:
	go get -u ./...
	go mod tidy

# Run integration tests
integration-test:
	go test -tags=integration -v ./tests/...
```

### Shell Aliases

Add these to your shell profile for faster development:

```bash
# OpenFrame CLI development aliases
alias ofdev="cd $GOPATH/src/github.com/flamingo-stack/openframe-cli"
alias ofbuild="make build"
alias oftest="make test"
alias oflint="make lint"
alias ofrun="./bin/openframe"

# Go development aliases
alias gob="go build"
alias got="go test"
alias gom="go mod"
alias gof="go fmt"
alias gor="go run"

# Git workflow aliases
alias gcm="git checkout main"
alias gco="git checkout"
alias gcb="git checkout -b"
alias gst="git status"
alias gad="git add"
alias gcmt="git commit -m"
alias gps="git push"
alias gpl="git pull"
```

## Verification

### Environment Check Script

Create a script to verify your development environment:

```bash
#!/bin/bash

echo "🔍 Checking OpenFrame CLI Development Environment..."
echo

errors=0

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
    if [[ $(go version | grep -o 'go[0-9]*\.[0-9]*' | sed 's/go//') < "1.19" ]]; then
        echo "❌ Go version should be 1.19 or higher"
        errors=$((errors + 1))
    fi
else
    echo "❌ Go not found"
    errors=$((errors + 1))
fi

# Check development tools
tools=("docker" "kubectl" "helm" "k3d" "git" "make")
for tool in "${tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool: $(command -v $tool)"
    else
        echo "❌ $tool not found"
        errors=$((errors + 1))
    fi
done

# Check Go tools
go_tools=("golangci-lint" "dlv" "gotestsum")
for tool in "${go_tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool: $(command -v $tool)"
    else
        echo "⚠️  $tool not found (optional but recommended)"
    fi
done

# Check IDE
if command -v code &> /dev/null; then
    echo "✅ VS Code: $(command -v code)"
else
    echo "⚠️  VS Code not found (recommended IDE)"
fi

echo
if [ $errors -eq 0 ]; then
    echo "🎉 Development environment is ready!"
    echo "Next steps:"
    echo "1. Clone the repository: git clone https://github.com/flamingo-stack/openframe-cli.git"
    echo "2. Navigate to project: cd openframe-cli"
    echo "3. Run setup: make deps"
    echo "4. Build project: make build"
    echo "5. Run tests: make test"
else
    echo "❌ Found $errors issues. Please resolve them before continuing."
    exit 1
fi
```

Save as `check-dev-env.sh` and run:

```bash
chmod +x check-dev-env.sh
./check-dev-env.sh
```

## Next Steps

Once your development environment is set up:

1. **[Local Development Guide](local-development.md)** - Clone the repo and start developing
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure
3. **[Testing Guide](../testing/overview.md)** - Learn about testing practices
4. **[Contributing Guidelines](../contributing/guidelines.md)** - Review contribution workflow

---

> 💡 **Pro Tip**: Set up your environment once properly, and it will serve you well throughout your OpenFrame CLI development journey. The time invested in proper tooling pays dividends in productivity and code quality.