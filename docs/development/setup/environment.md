# Development Environment Setup

This guide will help you set up a productive development environment for contributing to OpenFrame CLI. We'll cover IDE configuration, essential tools, development-specific environment variables, and recommended extensions.

## 🎯 Overview

A well-configured development environment will give you:
- 🚀 **Fast feedback loops** with hot reload and debugging
- 🔍 **Powerful debugging** with breakpoints and variable inspection  
- 📝 **Intelligent code completion** and error detection
- 🧪 **Integrated testing** with coverage reporting
- 🔧 **Seamless Git workflows** with built-in version control

## 📋 Prerequisites

Before setting up your development environment, ensure you have the basic prerequisites from the [Getting Started Prerequisites](../../getting-started/prerequisites.md):

- ✅ Go 1.21+
- ✅ Docker with daemon running
- ✅ kubectl configured
- ✅ Git configured

## 🛠️ IDE Recommendations & Setup

### VS Code (Recommended)

VS Code provides excellent Go support with rich extensions. Here's how to set it up:

#### Essential Extensions

Install these extensions for the best OpenFrame CLI development experience:

```bash
# Install VS Code extensions via command line
code --install-extension golang.Go
code --install-extension ms-vscode.vscode-yaml  
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension redhat.vscode-xml
code --install-extension GitHub.copilot
code --install-extension eamodio.gitlens
code --install-extension ms-vscode.test-adapter-converter
```

| Extension | Purpose | Key Features |
|-----------|---------|-------------|
| **Go** | Go language support | IntelliSense, debugging, testing |
| **YAML** | YAML/Kubernetes manifests | Validation, formatting |
| **Kubernetes** | K8s resource management | Cluster explorer, manifest editing |
| **GitLens** | Enhanced Git integration | Blame, history, authorship |
| **GitHub Copilot** | AI pair programming | Code completion, suggestions |

#### VS Code Configuration

Create a workspace-specific configuration:

<details>
<summary>📁 <strong>.vscode/settings.json</strong></summary>

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "10m",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "highlight",
    "coveredHighlightColor": "rgba(64,128,64,0.2)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.2)"
  },
  "go.useLanguageServer": true,
  "go.toolsManagement.checkForUpdates": "local",
  "files.exclude": {
    "**/vendor": true,
    "**/node_modules": true,
    "**/.git": true,
    "**/bin": true,
    "**/dist": true
  },
  "yaml.schemas": {
    "https://raw.githubusercontent.com/instrumenta/kubernetes-json-schema/master/v1.18.0-standalone-strict/all.json": [
      "*.k8s.yaml",
      "k8s/*.yaml",
      "examples/*.yaml"
    ]
  },
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

</details>

<details>
<summary>📁 <strong>.vscode/launch.json</strong></summary>

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["bootstrap", "test-cluster", "--verbose"],
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
      "args": ["cluster", "create", "debug-cluster", "--skip-wizard"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "env": {
        "OPENFRAME_TEST_MODE": "true"
      }
    }
  ]
}
```

</details>

#### Tasks Configuration

<details>
<summary>📁 <strong>.vscode/tasks.json</strong></summary>

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build OpenFrame CLI",
      "type": "shell",
      "command": "go",
      "args": ["build", "-o", "bin/openframe"],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "problemMatcher": "$go"
    },
    {
      "label": "Run Tests",
      "type": "shell", 
      "command": "go",
      "args": ["test", "-v", "./..."],
      "group": "test",
      "problemMatcher": "$go"
    },
    {
      "label": "Run Tests with Coverage",
      "type": "shell",
      "command": "go",
      "args": ["test", "-v", "-coverprofile=coverage.out", "./..."],
      "group": "test"
    },
    {
      "label": "Lint Code",
      "type": "shell",
      "command": "golangci-lint",
      "args": ["run"],
      "group": "test"
    }
  ]
}
```

</details>

### GoLand/IntelliJ IDEA

For JetBrains IDEs:

#### Essential Plugins
- **Go** (bundled) - Core Go support
- **Kubernetes** - Kubernetes manifest support
- **Docker** - Container management
- **Git Flow Integration** - Git workflow support

#### Configuration
1. **File → Settings → Go → GOPATH** - Ensure Go modules mode is enabled
2. **File → Settings → Tools → File Watchers** - Add gofmt and goimports
3. **Run → Edit Configurations** - Set up debug configurations similar to VS Code

### Vim/Neovim

For terminal-based development:

#### Essential Plugins (via vim-plug)
```vim
" Add to your .vimrc/.init.vim
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'preservim/nerdtree'
Plug 'tpope/vim-fugitive'
Plug 'airblade/vim-gitgutter'
```

## 🔧 Development Tools Installation

### Go Development Tools

Install essential Go tools for development:

```bash
# Go language server and tools
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/godoc@latest

# Linting and formatting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install mvdan.cc/gofumpt@latest

# Testing tools
go install github.com/rakyll/gotest@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest

# Debugging
go install github.com/go-delve/delve/cmd/dlv@latest

# Dependency management
go install github.com/psampaz/go-mod-outdated@latest
go install github.com/oligot/go-mod-upgrade@latest
```

### Additional Development Tools

```bash
# JSON/YAML processing
brew install jq yq    # macOS
apt install jq yq     # Ubuntu/Debian

# HTTP testing
brew install httpie   # macOS
apt install httpie    # Ubuntu/Debian

# File watching (for auto-rebuild)
go install github.com/cosmtrek/air@latest

# Benchmarking
go install golang.org/x/perf/cmd/benchstat@latest

# Code generation
go install github.com/vektra/mockery/v2@latest
```

### Kubernetes Development Tools

```bash
# Kubernetes tools
kubectl krew install ctx     # Context switching
kubectl krew install ns      # Namespace switching  
kubectl krew install tree    # Resource tree view
kubectl krew install tail    # Multi-pod log tailing

# Helm tools
helm plugin install https://github.com/databus23/helm-diff

# K3d (if not already installed)
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

## 🌍 Environment Variables

### Development-Specific Variables

Create a `.env.development` file in your project root:

```bash
# OpenFrame Development Configuration
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_DEV_MODE=true
export OPENFRAME_CONFIG_PATH=$HOME/.openframe-dev

# Go Configuration
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Testing Configuration  
export OPENFRAME_TEST_MODE=true
export OPENFRAME_TEST_CLUSTER_PREFIX=test-
export OPENFRAME_TEST_TIMEOUT=10m

# Build Configuration
export BUILD_VERSION=dev-$(git rev-parse --short HEAD)
export BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')

# Docker Configuration (if using custom registry)
export DOCKER_REGISTRY=localhost:5000
export DOCKER_BUILDKIT=1

# Kubernetes Development
export KUBECONFIG=$HOME/.kube/config:$HOME/.kube/dev-config
export KUBECTL_EXTERNAL_DIFF="code --wait --diff"
```

### Load Environment Variables

Add to your shell profile (`~/.bashrc`, `~/.zshrc`):

```bash
# Auto-load development environment
if [ -f "$PWD/.env.development" ]; then
    set -a
    source "$PWD/.env.development"
    set +a
fi

# Development aliases
alias ofd='go run main.go'  # Run OpenFrame in development
alias oftest='go test -v ./...'  # Run tests
alias ofbuild='go build -o bin/openframe-dev'  # Development build
alias oflint='golangci-lint run'  # Run linter
alias ofcover='go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out'
```

## 🚀 Development Workflow Setup

### Hot Reload with Air

Set up automatic rebuilds during development:

<details>
<summary>📁 <strong>.air.toml</strong></summary>

```toml
# Configuration for github.com/cosmtrek/air

root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["bootstrap", "dev-cluster", "--verbose"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "bin", "docs"]
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

</details>

Usage:
```bash
# Start development server with hot reload
air

# Or run specific commands
air -- cluster create test-cluster
```

### Git Configuration

Configure Git for optimal development workflow:

```bash
# Set up Git hooks
git config core.hooksPath .githooks

# Configure useful aliases
git config alias.co checkout
git config alias.br branch  
git config alias.ci commit
git config alias.st status
git config alias.unstage 'reset HEAD --'
git config alias.last 'log -1 HEAD'
git config alias.visual '!gitk'

# Configure merge/diff tools
git config merge.tool vscode
git config mergetool.vscode.cmd 'code --wait $MERGED'
git config diff.tool vscode  
git config difftool.vscode.cmd 'code --wait --diff $LOCAL $REMOTE'

# Configure commit template
git config commit.template .gitmessage
```

### Makefile for Common Tasks

Create a `Makefile` for development commands:

<details>
<summary>📁 <strong>Makefile</strong></summary>

```makefile
# OpenFrame CLI Development Makefile

.PHONY: help build test lint clean install dev

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=openframe
BUILD_DIR=bin
VERSION ?= dev-$(shell git rev-parse --short HEAD)
BUILD_DATE = $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.BuildDate=$(BUILD_DATE)"

## help: Show this help message
help:
	@echo "Available commands:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)

## test: Run tests
test:
	@echo "Running tests..."
	@go test -race -coverprofile=coverage.out ./...

## test-integration: Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -tags=integration -timeout=30m ./test/...

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out

## install: Install development tools
install:
	@echo "Installing development tools..."
	@go install golang.org/x/tools/gopls@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest

## dev: Start development server with hot reload
dev:
	@echo "Starting development server..."
	@air

## fmt: Format code
fmt:
	@echo "Formatting code..."
	@gofumpt -w .
	@goimports -w .

## mod-tidy: Tidy go modules
mod-tidy:
	@echo "Tidying go modules..."
	@go mod tidy

## coverage: Generate coverage report
coverage: test
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
```

</details>

## 🎨 Code Style Configuration

### EditorConfig

<details>
<summary>📁 <strong>.editorconfig</strong></summary>

```ini
# EditorConfig is awesome: https://EditorConfig.org

root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yml,yaml,json}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false

[Makefile]
indent_style = tab
```

</details>

### GolangCI-Lint Configuration

<details>
<summary>📁 <strong>.golangci.yml</strong></summary>

```yaml
linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unused
    - gocyclo
    - gofmt
    - goimports
    - golint
    - gosec
    - misspell
    - unconvert
    - unparam

linters-settings:
  gocyclo:
    min-complexity: 15
  golint:
    min-confidence: 0.8
  misspell:
    locale: US

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - unparam
```

</details>

## ✅ Verification

Verify your development environment is properly set up:

```bash
# Check Go installation and tools
go version
gopls version
golangci-lint version

# Verify project builds
make build
./bin/openframe --version

# Run tests
make test

# Check linting
make lint

# Test hot reload (in separate terminal)
make dev
```

## 🔧 Troubleshooting

### Common Issues

**Problem**: `gopls` not working in VS Code
```bash
# Reinstall Go tools
go install golang.org/x/tools/gopls@latest
# Restart VS Code
```

**Problem**: Tests failing with permission errors  
```bash
# Ensure Docker daemon is running and accessible
docker ps
sudo usermod -aG docker $USER  # Add user to docker group
newgrp docker                  # Refresh group membership
```

**Problem**: Hot reload not working
```bash
# Check air configuration
air -v
# Ensure .air.toml exists and is valid
```

**Problem**: Import resolution issues
```bash
# Clean module cache
go clean -modcache
go mod download
```

## 📚 Next Steps

Now that your development environment is configured:

1. **[Local Development Guide](local-development.md)** - Clone, build, and run the project
2. **[Architecture Overview](../architecture/overview.md)** - Understand the system design
3. **[Testing Overview](../testing/overview.md)** - Learn the testing strategy
4. **[Contributing Guidelines](../contributing/guidelines.md)** - Submit your first PR

---

> **💡 Pro Tip**: Bookmark this page and run through the verification steps whenever you update your development tools. A well-maintained development environment significantly improves productivity and reduces debugging time.