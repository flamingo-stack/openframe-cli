# Development Environment Setup

This guide will help you set up a complete development environment for OpenFrame CLI. Follow these steps to ensure you have all the necessary tools and configurations for productive development.

## IDE Recommendations

### Primary: Visual Studio Code

**Why VS Code?**
- Excellent Go extension with debugging support
- Integrated terminal for CLI testing
- Rich extension ecosystem
- Cross-platform consistency

**Installation:**
1. Download from [code.visualstudio.com](https://code.visualstudio.com/)
2. Install for your platform (Linux, macOS, Windows)

**Essential Extensions:**

| Extension | Purpose | Installation |
|-----------|---------|--------------|
| **Go** (Google) | Go language support, debugging, testing | `ext install golang.Go` |
| **Kubernetes** (Microsoft) | YAML support, cluster integration | `ext install ms-kubernetes-tools.vscode-kubernetes-tools` |
| **YAML** (Red Hat) | YAML syntax and validation | `ext install redhat.vscode-yaml` |
| **Git Lens** (GitKraken) | Enhanced Git integration | `ext install eamodio.gitlens` |
| **Docker** (Microsoft) | Container support | `ext install ms-azuretools.vscode-docker` |

**Install all extensions:**
```bash
code --install-extension golang.Go
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools  
code --install-extension redhat.vscode-yaml
code --install-extension eamodio.gitlens
code --install-extension ms-azuretools.vscode-docker
```

### Alternative: GoLand

**Why GoLand?**
- Professional Go IDE with advanced refactoring
- Excellent debugging and profiling tools
- Built-in database and REST client tools

**Installation:**
1. Download from [jetbrains.com/go](https://www.jetbrains.com/go/)
2. 30-day free trial, then requires license

### Alternative: Neovim/Vim

**Why Neovim?**
- Lightweight and fast
- Highly customizable
- Excellent terminal integration

**Setup with Go support:**
```bash
# Install Neovim
# Ubuntu/Debian
sudo apt install neovim

# macOS
brew install neovim

# Install vim-plug
curl -fLo ~/.local/share/nvim/site/autoload/plug.vim --create-dirs \
    https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim

# Add to ~/.config/nvim/init.vim
call plug#begin()
Plug 'fatih/vim-go', { 'do': ':GoInstallBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}
call plug#end()
```

## Required Development Tools

### Go Development Environment

**Go 1.19+ Installation:**

**Linux (Ubuntu/Debian):**
```bash
# Remove old Go versions
sudo rm -rf /usr/local/go

# Download and install latest Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to ~/.bashrc or ~/.zshrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc
```

**macOS:**
```bash
# Using Homebrew (recommended)
brew install go

# Or download from https://go.dev/dl/
# Add to ~/.zshrc
echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.zshrc
source ~/.zshrc
```

**Windows (WSL2):**
```bash
# Install inside WSL2
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to ~/.bashrc
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc
```

**Verify Installation:**
```bash
go version
# Expected: go version go1.21.5 linux/amd64

go env GOPATH
# Expected: /home/username/go
```

### Essential Go Tools

Install development tools that enhance the Go experience:

```bash
# Go language server
go install golang.org/x/tools/gopls@latest

# Code formatting and imports
go install golang.org/x/tools/cmd/goimports@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Debugging
go install github.com/go-delve/delve/cmd/dlv@latest

# Testing utilities
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install gotest.tools/gotestsum@latest

# Documentation
go install golang.org/x/tools/cmd/godoc@latest
```

### Container and Kubernetes Tools

**Docker:**
Follow platform-specific installation:
- **Linux**: [Docker Engine](https://docs.docker.com/engine/install/)
- **macOS**: [Docker Desktop](https://docs.docker.com/desktop/mac/install/)
- **Windows**: [Docker Desktop with WSL2](https://docs.docker.com/desktop/windows/install/)

**Verify Docker:**
```bash
docker --version
docker ps
```

**kubectl:**
```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# macOS
brew install kubectl

# Windows (WSL2)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

**K3D (for testing):**
```bash
# Install K3D
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

### Version Control

**Git Configuration:**
```bash
# Set identity
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Set editor
git config --global core.editor "code --wait"

# Enable helpful defaults
git config --global init.defaultBranch main
git config --global pull.rebase false
git config --global push.default simple
```

## Environment Variables

### Go Environment

Add these to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Go environment
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# Go development settings
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
export CGO_ENABLED=1

# Enable Go modules (default in Go 1.16+)
export GO111MODULE=on
```

### Kubernetes Development

```bash
# Kubernetes configuration
export KUBECONFIG=$HOME/.kube/config

# K3D specific settings
export K3D_FIX_CGROUPV2=1  # For newer Linux systems
```

### OpenFrame CLI Development

```bash
# OpenFrame development
export OPENFRAME_DEV=1
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=$HOME/.openframe

# Optional: Custom paths for external tools
export K3D_BINARY_PATH=/usr/local/bin/k3d
export HELM_BINARY_PATH=/usr/local/bin/helm
export KUBECTL_BINARY_PATH=/usr/local/bin/kubectl
```

### Platform-Specific Variables

**WSL2 (Windows):**
```bash
# WSL2 integration
export WSL_DISTRO_NAME=$(cat /proc/version | grep -oP 'WSL\d+')
export DOCKER_HOST=unix:///var/run/docker.sock

# Fix for systemd issues
export XDG_RUNTIME_DIR=/tmp/runtime-$USER
mkdir -p $XDG_RUNTIME_DIR
chmod 0700 $XDG_RUNTIME_DIR
```

## IDE Configuration

### VS Code Configuration

Create `.vscode/settings.json` in your OpenFrame CLI workspace:

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
  "go.testTimeout": "60s",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,128,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.25)"
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml",
    "Dockerfile*": "dockerfile",
    "*.go": "go"
  },
  "yaml.schemas": {
    "kubernetes": "**/*.k8s.yaml"
  }
}
```

### Debug Configuration

Create `.vscode/launch.json`:

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
      "args": ["--help"],
      "cwd": "${workspaceFolder}",
      "env": {
        "OPENFRAME_DEV": "1",
        "OPENFRAME_LOG_LEVEL": "debug"
      },
      "console": "integratedTerminal"
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--dry-run"],
      "cwd": "${workspaceFolder}",
      "env": {
        "OPENFRAME_DEV": "1",
        "OPENFRAME_LOG_LEVEL": "debug"
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
      "cwd": "${workspaceFolder}"
    }
  ]
}
```

## Development Utilities

### Shell Aliases

Add helpful aliases to your shell profile:

```bash
# OpenFrame development aliases
alias of='./openframe'
alias ofb='go build -o openframe . && ./openframe'
alias ofd='go build -o openframe . && ./openframe --verbose'
alias oft='go test ./... -v -race'
alias ofc='golangci-lint run'

# Kubernetes aliases  
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get svc'
alias kga='kubectl get all'
alias kaf='kubectl apply -f'
alias kdel='kubectl delete'

# Development workflow
alias gb='go build'
alias gt='go test ./...'
alias gf='gofmt -s -w .'
alias gi='goimports -w .'
alias gl='golangci-lint run'
```

### Git Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
echo "Running pre-commit checks..."

# Format code
goimports -w .
gofmt -s -w .

# Run linter
golangci-lint run
if [ $? -ne 0 ]; then
    echo "Linting failed. Please fix the issues before committing."
    exit 1
fi

# Run tests
go test ./... -race -short
if [ $? -ne 0 ]; then
    echo "Tests failed. Please fix the issues before committing."
    exit 1
fi

echo "Pre-commit checks passed!"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Verification Script

Create a script to verify your environment setup:

```bash
#!/bin/bash
# save as check-dev-env.sh

echo "🔍 OpenFrame CLI Development Environment Check"
echo "============================================="

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
    echo "   GOPATH: $GOPATH"
    echo "   GOROOT: $(go env GOROOT)"
else
    echo "❌ Go: Not found"
fi

# Check Go tools
echo "🔧 Go Tools:"
for tool in gopls goimports golangci-lint dlv; do
    if command -v $tool &> /dev/null; then
        echo "   ✅ $tool"
    else
        echo "   ❌ $tool: Not found"
    fi
done

# Check Docker
if command -v docker &> /dev/null && docker ps &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
else
    echo "❌ Docker: Not found or not running"
fi

# Check Kubernetes tools
for tool in kubectl k3d helm; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool: $(command $tool version --client 2>/dev/null | head -1)"
    else
        echo "❌ $tool: Not found"
    fi
done

# Check IDE
if command -v code &> /dev/null; then
    echo "✅ VS Code: $(code --version | head -1)"
else
    echo "ℹ️  VS Code: Not found (optional)"
fi

echo ""
echo "Environment check complete!"
```

Run the verification:
```bash
chmod +x check-dev-env.sh
./check-dev-env.sh
```

## Next Steps

After completing the environment setup:

1. **[Local Development Guide](local-development.md)** - Clone and build the project
2. **[Architecture Overview](../architecture/README.md)** - Understand the codebase structure
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the development workflow

## Troubleshooting

### Common Go Issues

**GOPROXY errors:**
```bash
export GOPROXY=direct
go clean -modcache
go mod download
```

**Module verification failures:**
```bash
export GOSUMDB=off  # Temporary workaround
go mod tidy
export GOSUMDB=sum.golang.org  # Re-enable
```

### Docker Permission Issues (Linux)

```bash
sudo usermod -aG docker $USER
newgrp docker
```

### WSL2 Integration Issues

```bash
# Reset Docker WSL integration
wsl --shutdown
# Restart Docker Desktop
# Re-enable WSL integration in Docker settings
```

Your development environment is now ready for OpenFrame CLI development! The next step is to clone the repository and start building.