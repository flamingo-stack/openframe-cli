# Development Environment Setup

This guide will help you set up a complete development environment for contributing to OpenFrame CLI. Follow these steps to get your tools configured and ready for development.

## Required Development Tools

### Core Development Stack

| Tool | Version | Purpose |
|------|---------|---------|
| **Go** | 1.21+ | Primary development language |
| **Git** | 2.30+ | Version control |
| **Make** | 4.0+ | Build automation |
| **Docker** | 20.10+ | Container runtime for testing |
| **kubectl** | 1.24+ | Kubernetes CLI |
| **K3d** | 5.4+ | Local Kubernetes clusters |

### IDE Recommendations

#### Visual Studio Code (Recommended)
VS Code provides excellent Go support and integrates well with our development workflow.

**Required Extensions:**
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-go",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "redhat.vscode-yaml",
    "davidanson.vscode-markdownlint"
  ]
}
```

**VS Code Settings for OpenFrame:**
```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v", "-race"],
  "go.buildTags": "integration",
  "yaml.schemas": {
    "https://json.schemastore.org/kustomization": "kustomization.yaml"
  }
}
```

#### GoLand / IntelliJ IDEA
Professional IDE with advanced refactoring and debugging features.

**Recommended Plugins:**
- Kubernetes
- Docker
- Helm
- Makefile Language

#### Vim/Neovim
For developers preferring terminal-based editors.

**Essential Plugins:**
- `vim-go` or `govim`
- `ale` for linting
- `fzf.vim` for file navigation

## Go Development Setup

### Install Go

#### Using Official Installer
```bash
# Download and install Go 1.21+
wget https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
```

#### Using Package Managers

**macOS (Homebrew):**
```bash
brew install go
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install golang-go
```

**Windows (Chocolatey):**
```bash
choco install golang
```

### Configure Go Environment

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
export GO111MODULE=on
export CGO_ENABLED=1
```

Verify your Go installation:
```bash
go version  # Should show Go 1.21+
go env GOPATH
go env GOROOT
```

## Development Tools Installation

### golangci-lint (Code Linting)
```bash
# Install latest version
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest
```

### go-mock (Test Mocking)
```bash
go install github.com/golang/mock/mockgen@latest
```

### goimports (Code Formatting)
```bash
go install golang.org/x/tools/cmd/goimports@latest
```

### Wire (Dependency Injection)
```bash
go install github.com/google/wire/cmd/wire@latest
```

### Testing Tools
```bash
# Test coverage tools
go install golang.org/x/tools/cmd/cover@latest

# Benchmarking and profiling
go install golang.org/x/perf/cmd/benchstat@latest
```

## Container and Kubernetes Tools

### Docker Desktop Installation

#### macOS
```bash
# Using Homebrew
brew install --cask docker

# Or download from https://www.docker.com/products/docker-desktop
```

#### Windows
Download Docker Desktop from [docker.com](https://www.docker.com/products/docker-desktop) and install.

#### Linux
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

### kubectl Installation
```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Windows (PowerShell)
choco install kubernetes-cli
```

### K3d Installation
```bash
# All platforms
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Or using package managers
brew install k3d  # macOS
choco install k3d  # Windows
```

### Helm Installation
```bash
# Using script
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Or package managers
brew install helm  # macOS
choco install kubernetes-helm  # Windows
```

## Editor Configuration

### VS Code Configuration

Create `.vscode/settings.json` in the project root:
```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.gopath": "${workspaceFolder}",
  "go.goroot": "/usr/local/go",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--enable-all",
    "--disable=wsl,nlreturn,exhaustivestruct,gofumpt"
  ],
  "go.buildTags": "integration",
  "go.testFlags": ["-v", "-race", "-count=1"],
  "go.testTimeout": "30s",
  "files.exclude": {
    "**/vendor/**": true,
    "**/node_modules/**": true,
    "**/.git/**": true
  }
}
```

Create `.vscode/launch.json` for debugging:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug CLI",
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
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.v", "-test.run", "TestName"]
    }
  ]
}
```

### Git Configuration

Configure Git for OpenFrame development:
```bash
# Set your identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Helpful aliases
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status

# Line ending configuration
git config --global core.autocrlf input  # Linux/macOS
git config --global core.autocrlf true   # Windows
```

## Environment Variables

Set these environment variables for development:

### Required Variables
```bash
# Go configuration
export GO111MODULE=on
export CGO_ENABLED=1

# OpenFrame CLI configuration
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=$HOME/.openframe
export KUBECONFIG=$HOME/.kube/config

# Development flags
export OPENFRAME_DEV_MODE=true
export OPENFRAME_MOCK_PROVIDERS=false
```

### Optional Variables
```bash
# Docker configuration
export DOCKER_HOST=unix:///var/run/docker.sock

# Testing configuration
export TEST_CLUSTER_NAME=openframe-test
export TEST_TIMEOUT=300s

# Build configuration
export CGO_ENABLED=1
export GOOS=linux
export GOARCH=amd64
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.) to persist them.

## Verification

Verify your development environment is properly configured:

### Go Environment
```bash
# Check Go installation
go version
go env GOPATH
go env GOPROXY

# Verify development tools
golangci-lint version
mockgen -version
goimports -h
```

### Container Tools
```bash
# Check Docker
docker version
docker ps

# Check Kubernetes tools
kubectl version --client
k3d version
helm version
```

### OpenFrame CLI Build Test
```bash
# Clone the repository (if not done already)
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the CLI
make build

# Run basic test
./bin/openframe --version
```

Expected output:
```text
OpenFrame CLI - Kubernetes cluster bootstrapping and development tools
Version: dev (none) built on unknown
```

## IDE-Specific Setup

### VS Code Tasks

Create `.vscode/tasks.json` for common development tasks:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "type": "shell",
      "command": "make build",
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    },
    {
      "label": "test",
      "type": "shell",
      "command": "make test",
      "group": "test",
      "presentation": {
        "echo": true,
        "reveal": "always",
        "focus": false,
        "panel": "shared"
      }
    },
    {
      "label": "lint",
      "type": "shell",
      "command": "make lint",
      "group": "build"
    }
  ]
}
```

## Troubleshooting Common Issues

### Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Verify module
go mod tidy
go mod verify
```

### Docker Permission Issues (Linux)
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again
```

### kubectl Context Issues
```bash
# Reset kubectl config
kubectl config unset current-context
kubectl config get-contexts
```

### IDE Performance Issues
```bash
# Exclude large directories from indexing
# Add to .gitignore or IDE settings:
vendor/
.git/
node_modules/
*.log
```

## Next Steps

With your development environment configured:

1. **[Clone and build locally](local-development.md)** - Get the code running
2. **[Understand the architecture](../architecture/overview.md)** - Learn the system design
3. **[Run the test suite](../testing/overview.md)** - Verify everything works
4. **[Read contributing guidelines](../contributing/guidelines.md)** - Follow our standards

> **💡 Pro Tip**: Bookmark the [Makefile targets reference](../reference/makefile.md) for quick access to common development commands.