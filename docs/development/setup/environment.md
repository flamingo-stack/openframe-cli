# Development Environment Setup

This guide helps you configure a productive development environment for contributing to OpenFrame CLI. We'll set up your IDE, development tools, and recommended extensions for Go development.

## IDE Recommendations and Setup

### Visual Studio Code (Recommended)

VS Code provides excellent Go support and integrates well with the OpenFrame CLI development workflow.

#### Essential Extensions

Install these extensions for the best development experience:

```bash
# Install via command line
code --install-extension golang.go
code --install-extension ms-vscode.vscode-yaml
code --install-extension redhat.vscode-xml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension ms-azuretools.vscode-docker
code --install-extension eamodio.gitlens
code --install-extension github.copilot  # Optional: AI assistance
```

| Extension | Purpose | Benefits |
|-----------|---------|----------|
| **Go** | Go language support | Syntax highlighting, IntelliSense, debugging, testing |
| **YAML** | YAML file support | Kubernetes manifests, Helm charts validation |
| **Kubernetes** | Kubernetes integration | Cluster management, resource inspection |
| **Docker** | Container support | Dockerfile editing, container management |
| **GitLens** | Git integration | Blame annotations, commit history |

#### VS Code Configuration

Create `.vscode/settings.json` in your OpenFrame CLI project:

```json
{
  "go.buildTags": "integration",
  "go.testTags": "unit,integration",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.useLanguageServer": true,
  "go.languageServerFlags": ["-rpc.trace"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "yaml.schemas": {
    "https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json": "*.k8s.yaml"
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml"
  }
}
```

#### Debug Configuration

Create `.vscode/launch.json` for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug OpenFrame CLI",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "test-cluster", "--verbose"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      },
      "console": "integratedTerminal"
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "debug-cluster", "--deployment-mode=oss-tenant"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      },
      "console": "integratedTerminal"
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.v"],
      "console": "integratedTerminal"
    }
  ]
}
```

### GoLand/IntelliJ IDEA

For JetBrains IDE users:

#### Required Plugins
- Go plugin (built-in)
- Kubernetes plugin
- Docker plugin
- YAML/Ansible Support

#### Configuration
1. **Go Module Support**: Enable Go modules in Settings → Go → Go Modules
2. **Code Style**: Import the Go code style settings
3. **File Watchers**: Set up automatic formatting on save
4. **Run Configurations**: Create configurations for common commands

### Vim/Neovim

For command-line editors:

#### Essential Plugins
```vim
" .vimrc / init.vim
Plug 'fatih/vim-go'
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'vim-airline/vim-airline'
Plug 'tpope/vim-fugitive'
Plug 'airblade/vim-gitgutter'
```

#### Go Configuration
```vim
" Go settings
let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_function_calls = 1
let g:go_auto_type_info = 1
let g:go_fmt_command = "goimports"
```

## Required Development Tools

### Core Tools Installation

#### macOS
```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install development tools
brew install go
brew install golangci-lint
brew install docker
brew install kubectl
brew install k3d
brew install helm

# Optional: Install pre-commit hooks
brew install pre-commit
```

#### Ubuntu/Debian
```bash
# Go installation
curl -L https://go.dev/dl/go1.21.0.linux-amd64.tar.gz | sudo tar -xzf - -C /usr/local
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Development tools
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Kubernetes tools
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# K3d and Helm
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update && sudo apt-get install helm
```

#### Windows (WSL2)
```powershell
# Install WSL2 and Ubuntu
wsl --install Ubuntu-22.04

# In WSL2, follow Ubuntu instructions above
```

### Tool Versions

Ensure you have compatible versions:

| Tool | Minimum Version | Recommended | Check Command |
|------|----------------|-------------|---------------|
| **Go** | 1.19 | 1.21+ | `go version` |
| **golangci-lint** | 1.50 | 1.54+ | `golangci-lint --version` |
| **Docker** | 20.10 | 24.0+ | `docker --version` |
| **kubectl** | 1.25 | 1.28+ | `kubectl version --client` |
| **K3d** | 5.4 | 5.6+ | `k3d --version` |
| **Helm** | 3.8 | 3.12+ | `helm version` |

## Environment Variables

Configure these environment variables for development:

### Essential Variables

```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# OpenFrame CLI development
export OPENFRAME_DEBUG=true
export OPENFRAME_CONFIG_DIR=$HOME/.openframe
export OPENFRAME_LOG_LEVEL=debug

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config
export KUBERNETES_MASTER=http://localhost:8080  # For testing

# Docker configuration
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

### Optional Variables

```bash
# Development convenience
export EDITOR=code  # or vim, nano, etc.
export OPENFRAME_DEFAULT_CLUSTER_NAME=dev-cluster
export OPENFRAME_AUTO_CLEANUP=true

# Testing configuration
export OPENFRAME_TEST_TIMEOUT=300s
export OPENFRAME_INTEGRATION_TESTS=true
```

## Pre-commit Hooks

Set up pre-commit hooks to catch issues early:

### Install pre-commit

```bash
# Install pre-commit
go install github.com/pre-commit/pre-commit@latest

# Or via package manager
brew install pre-commit  # macOS
apt install pre-commit   # Ubuntu
```

### Configuration

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict

  - repo: local
    hooks:
      - id: go-fmt
        name: go fmt
        entry: gofmt
        language: system
        args: [-w]
        files: \.go$
        
      - id: go-imports
        name: go imports
        entry: goimports
        language: system
        args: [-w]
        files: \.go$
        
      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint
        language: system
        args: [run]
        files: \.go$
        
      - id: go-test
        name: go test
        entry: go
        language: system
        args: [test, ./...]
        files: \.go$
        pass_filenames: false
```

### Install hooks

```bash
# Install the pre-commit hook
pre-commit install

# Test the hooks
pre-commit run --all-files
```

## IDE-Specific Configurations

### VS Code Tasks

Create `.vscode/tasks.json` for common development tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build OpenFrame CLI",
      "type": "shell",
      "command": "go",
      "args": ["build", "-o", "openframe", "./cmd"],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "silent",
        "focus": false,
        "panel": "shared"
      }
    },
    {
      "label": "Run Tests",
      "type": "shell",
      "command": "go",
      "args": ["test", "./..."],
      "group": "test",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    },
    {
      "label": "Lint Code",
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
    },
    {
      "label": "Format Code",
      "type": "shell",
      "command": "goimports",
      "args": ["-w", "."],
      "group": "build"
    }
  ]
}
```

### Makefile for Development

Create a `Makefile` for common tasks:

```makefile
.PHONY: build test lint format clean install

# Variables
BINARY_NAME=openframe
BUILD_DIR=bin

# Build binary
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

# Run tests
test:
	go test ./... -v

# Run integration tests
test-integration:
	go test ./... -tags=integration -v

# Lint code
lint:
	golangci-lint run

# Format code
format:
	gofmt -w .
	goimports -w .

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	go clean

# Install binary
install:
	go install ./cmd

# Development setup
dev-setup: install-tools
	pre-commit install

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

# Run development server
dev: build
	./$(BUILD_DIR)/$(BINARY_NAME) cluster create dev-cluster --verbose

# Quick validation
validate: format lint test
```

## Verification

### Environment Check Script

Create `scripts/check-dev-env.sh`:

```bash
#!/bin/bash
set -e

echo "🔍 Checking OpenFrame CLI Development Environment..."

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    echo "✅ Go: $GO_VERSION"
else
    echo "❌ Go not found"
    exit 1
fi

# Check golangci-lint
if command -v golangci-lint &> /dev/null; then
    LINT_VERSION=$(golangci-lint --version | cut -d' ' -f4)
    echo "✅ golangci-lint: $LINT_VERSION"
else
    echo "❌ golangci-lint not found"
fi

# Check Docker
if command -v docker &> /dev/null && docker info &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "✅ Docker: $DOCKER_VERSION"
else
    echo "❌ Docker not available"
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    KUBECTL_VERSION=$(kubectl version --client --short 2>/dev/null | cut -d' ' -f3)
    echo "✅ kubectl: $KUBECTL_VERSION"
else
    echo "❌ kubectl not found"
fi

# Check K3d
if command -v k3d &> /dev/null; then
    K3D_VERSION=$(k3d --version | head -1 | cut -d' ' -f3)
    echo "✅ K3d: $K3D_VERSION"
else
    echo "❌ K3d not found"
fi

# Check Helm
if command -v helm &> /dev/null; then
    HELM_VERSION=$(helm version --short | cut -d' ' -f1)
    echo "✅ Helm: $HELM_VERSION"
else
    echo "❌ Helm not found"
fi

# Check environment variables
echo ""
echo "📋 Environment Variables:"
echo "GOPATH: ${GOPATH:-Not set}"
echo "OPENFRAME_DEBUG: ${OPENFRAME_DEBUG:-Not set}"
echo "KUBECONFIG: ${KUBECONFIG:-Not set}"

echo ""
echo "🎉 Development environment check complete!"
```

### Run Environment Check

```bash
chmod +x scripts/check-dev-env.sh
./scripts/check-dev-env.sh
```

## Next Steps

With your development environment configured:

1. **[Local Development](./local-development.md)** - Clone, build, and run OpenFrame CLI locally
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the contribution workflow

## Troubleshooting

### Common Issues

#### Go Path Issues
```bash
# Check Go installation
go env GOPATH GOROOT

# Fix PATH if needed
export PATH=$PATH:/usr/local/go/bin:$(go env GOPATH)/bin
```

#### VS Code Go Extension Issues
```bash
# Reset Go tools
go clean -modcache
code --install-extension golang.go --force
```

#### Docker Permission Issues (Linux)
```bash
sudo usermod -aG docker $USER
newgrp docker
```

---

> **💡 Pro Tip**: Set up your environment incrementally. Start with the core tools (Go, Docker) and add development enhancements (linting, pre-commit hooks) as you become familiar with the codebase.

Your development environment is now ready! Continue to [Local Development](./local-development.md) to start working with the OpenFrame CLI codebase.