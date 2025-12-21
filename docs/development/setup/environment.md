# Development Environment Setup

Set up a productive development environment for OpenFrame CLI with the right tools, IDE configuration, and extensions.

## IDE Recommendations

### Primary: Visual Studio Code

**Recommended for**: All developers, especially those new to Go

**Advantages:**
- Excellent Go extension ecosystem
- Integrated terminal and debugging
- Great Git integration
- Strong Kubernetes extension support

#### VS Code Setup

```bash
# Install VS Code (if not already installed)
# macOS
brew install --cask visual-studio-code

# Ubuntu/Debian  
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" | sudo tee /etc/apt/sources.list.d/vscode.list
sudo apt update && sudo apt install code
```

### Alternative: GoLand

**Recommended for**: Experienced developers who prefer JetBrains IDEs

**Advantages:**
- Superior refactoring tools
- Advanced debugging capabilities  
- Excellent code analysis
- Built-in testing tools

### Alternative: Vim/Neovim

**Recommended for**: Developers who prefer terminal-based editing

**Advantages:**
- Lightning fast editing
- Highly customizable
- Efficient keyboard-driven workflow
- Works great over SSH

## Required Development Tools

### Core Tools

| Tool | Version | Installation | Purpose |
|------|---------|--------------|---------|
| **Go** | 1.21+ | [golang.org](https://golang.org/dl/) | Primary language |
| **Git** | 2.25+ | System package manager | Version control |
| **Make** | 4.0+ | System package manager | Build automation |
| **Docker** | 20.0+ | [docker.com](https://docs.docker.com/get-docker/) | Container runtime |

#### Go Installation

```bash
# Install Go (replace with latest version)
# macOS
brew install go

# Ubuntu/Debian
sudo snap install go --classic

# CentOS/RHEL
sudo dnf install golang

# Verify installation
go version
```

#### Configure Go Environment

```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
export GOPATH=$HOME/go
export GOROOT=/usr/local/go  # Adjust based on your installation
export PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# Create Go workspace
mkdir -p $GOPATH/{bin,src,pkg}

# Verify configuration
go env GOPATH
go env GOROOT
```

### Go Development Tools

Install essential Go development tools:

```bash
# Go language server and tools
go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/godoc@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Debugging
go install github.com/go-delve/delve/cmd/dlv@latest

# Testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/mfridman/tparse@latest

# Development utilities
go install github.com/cosmtrek/air@latest  # Hot reloading
go install github.com/mikefarah/yq/v4@latest  # YAML processing
```

### Kubernetes Tools

```bash
# kubectl (if not already installed via bootstrap)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# k3d (for local clusters)
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Helm 
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# ArgoCD CLI
curl -sSL -o argocd-linux-amd64 https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
sudo install -m 555 argocd-linux-amd64 /usr/local/bin/argocd
```

## VS Code Extension Configuration

### Essential Extensions

Install these extensions for optimal Go development:

```bash
# Install via VS Code command palette (Ctrl+Shift+P)
# Or install via command line
code --install-extension golang.go
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension ms-vscode.vscode-yaml
code --install-extension redhat.vscode-yaml
code --install-extension bradlc.vscode-tailwindcss
code --install-extension esbenp.prettier-vscode
code --install-extension streetsidesoftware.code-spell-checker
```

### Extension Details

| Extension | Purpose | Configuration |
|-----------|---------|---------------|
| **Go** | Go language support, debugging, testing | Auto-configured |
| **Kubernetes** | YAML editing, cluster management | Connect to local clusters |
| **YAML** | YAML syntax highlighting and validation | Essential for K8s manifests |
| **GitLens** | Enhanced Git integration | Useful for code history |
| **REST Client** | API testing directly in VS Code | Test OpenFrame APIs |

### VS Code Configuration

Create `.vscode/settings.json` in your workspace:

```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.vetOnSave": "package",
  "go.buildOnSave": "package",
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "10m",
  "go.coverOnSave": false,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.exclude": {
    "**/bin": true,
    "**/dist": true,
    "**/.git": true
  },
  "yaml.schemas": {
    "kubernetes://schema/v1@deployment": "k8s-*.yaml",
    "kubernetes://schema/v1@service": "k8s-*.yaml"
  }
}
```

### Launch Configuration

Create `.vscode/launch.json` for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug OpenFrame CLI",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose"],
      "env": {
        "OPENFRAME_CONFIG_DIR": "${workspaceFolder}/.openframe-dev"
      },
      "console": "integratedTerminal"
    },
    {
      "name": "Test Package",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.v"]
    }
  ]
}
```

### Task Configuration

Create `.vscode/tasks.json` for build tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "type": "shell",
      "command": "make",
      "args": ["build"],
      "group": {
        "kind": "build",
        "isDefault": true
      },
      "presentation": {
        "panel": "shared",
        "clear": true
      }
    },
    {
      "label": "test",
      "type": "shell", 
      "command": "make",
      "args": ["test"],
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "lint",
      "type": "shell",
      "command": "make", 
      "args": ["lint"],
      "group": "build"
    }
  ]
}
```

## Environment Variables

### Development Configuration

Set up environment variables for development:

```bash
# Add to ~/.bashrc, ~/.zshrc, etc.

# Go development
export GOPATH=$HOME/go
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org

# OpenFrame CLI development
export OPENFRAME_DEV=true
export OPENFRAME_CONFIG_DIR=$HOME/.openframe-dev
export OPENFRAME_LOG_LEVEL=debug

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config:$HOME/.kube/dev-config

# Container development
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Development tools
export EDITOR=code  # or vim, nano, etc.
export BROWSER=firefox  # or chrome, safari, etc.
```

### Development Aliases

Add helpful aliases for efficient development:

```bash
# Add to ~/.bashrc, ~/.zshrc, etc.

# OpenFrame CLI aliases
alias of='openframe'
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofch='openframe chart'
alias ofd='openframe dev'

# Go development aliases
alias gob='go build'
alias got='go test'
alias gor='go run'
alias gom='go mod'
alias gof='go fmt'

# Kubernetes aliases
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kgd='kubectl get deployments'
alias kaf='kubectl apply -f'
alias kdf='kubectl delete -f'

# Docker aliases
alias d='docker'
alias dc='docker-compose'
alias dps='docker ps'
alias di='docker images'

# Development workflow
alias dev-setup='make clean build test lint'
alias dev-test='make test-watch'
alias dev-build='make build-dev'
```

## Shell Configuration

### Bash Completion

Enable command completion for better productivity:

```bash
# Install bash completion for kubectl
kubectl completion bash | sudo tee /etc/bash_completion.d/kubectl

# Install completion for other tools
helm completion bash | sudo tee /etc/bash_completion.d/helm
k3d completion bash | sudo tee /etc/bash_completion.d/k3d

# Add to ~/.bashrc
if [ -f /etc/bash_completion ]; then
  . /etc/bash_completion
fi
```

### Zsh Configuration (Oh My Zsh)

If using Zsh with Oh My Zsh:

```bash
# Install Oh My Zsh plugins
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
git clone https://github.com/zsh-users/zsh-syntax-highlighting ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting

# Edit ~/.zshrc
plugins=(
  git
  golang
  kubectl
  docker
  docker-compose
  zsh-autosuggestions
  zsh-syntax-highlighting
)

# Enable kubectl completion
source <(kubectl completion zsh)
```

## Git Configuration

### Git Hooks for Development

Set up Git hooks to maintain code quality:

```bash
# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Format Go code
echo "Formatting Go code..."
gofmt -w -s .

# Run linting
echo "Running linters..."
golangci-lint run

# Run tests
echo "Running tests..."
go test ./... -short

echo "Pre-commit checks passed!"
EOF

chmod +x .git/hooks/pre-commit
```

### Git Configuration

```bash
# Configure Git for development
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Useful Git aliases
git config alias.st status
git config alias.co checkout
git config alias.br branch
git config alias.ci commit
git config alias.unstage 'reset HEAD --'
git config alias.last 'log -1 HEAD'
git config alias.visual '!gitk'

# Better diff and merge tools
git config merge.tool vimdiff
git config diff.tool vimdiff
```

## Testing Your Setup

Verify everything is working correctly:

```bash
# Test Go installation
go version

# Test development tools
golangci-lint --version
dlv version
kubectl version --client

# Test VS Code (if using)
code --version

# Test Docker
docker --version
docker run hello-world

# Test Kubernetes tools  
k3d --version
helm version
```

**Expected Output:**
All commands should show version information without errors.

## Troubleshooting

### Common Issues

#### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Verify module setup
go mod verify
go mod tidy
```

#### PATH Issues
```bash
# Check current PATH
echo $PATH

# Verify Go binary locations
which go
which golangci-lint
```

#### Permission Issues
```bash
# Fix Go workspace permissions
sudo chown -R $USER:$USER $GOPATH

# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

## Next Steps

With your development environment set up:

1. **[Local Development Guide](./local-development.md)** - Clone and run OpenFrame CLI
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure  
3. **[Testing Overview](../testing/overview.md)** - Learn the testing approach
4. **[Contributing Guidelines](../contributing/guidelines.md)** - Start contributing

## IDE-Specific Guides

### GoLand Setup

```bash
# Install GoLand plugins
# Go to Preferences > Plugins and install:
# - Kubernetes
# - YAML/Ansible support
# - .env files support
```

### Vim/Neovim Setup

```bash
# Install vim-go plugin
# Add to ~/.vimrc or ~/.config/nvim/init.vim
call plug#begin()
Plug 'fatih/vim-go'
Plug 'tpope/vim-fugitive'
call plug#end()

# Configure vim-go
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
```

---

**Environment ready!** 🎉 Next: **[Local Development Guide](./local-development.md)** to get the code running locally.