# Development Environment Setup

This guide will help you set up a complete development environment for working with OpenFrame CLI. We'll configure your IDE, install required tools, and set up useful extensions for an optimal development experience.

## IDE Recommendations and Setup

### Visual Studio Code (Recommended)

VS Code provides excellent Go support and integrates well with Kubernetes development:

#### Installation
```bash
# Ubuntu/Debian
sudo snap install --classic code

# macOS
brew install --cask visual-studio-code

# Or download from https://code.visualstudio.com/
```

#### Essential Extensions

Install these extensions for optimal Go and Kubernetes development:

| Extension | Purpose | Install Command |
|-----------|---------|-----------------|
| **Go** | Go language support, debugging, testing | `code --install-extension golang.go` |
| **Kubernetes** | K8s resource management and YAML support | `code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools` |
| **YAML** | YAML syntax highlighting and validation | `code --install-extension redhat.vscode-yaml` |
| **GitLens** | Enhanced Git integration and history | `code --install-extension eamodio.gitlens` |
| **Thunder Client** | API testing (useful for testing endpoints) | `code --install-extension rangav.vscode-thunder-client` |
| **Error Lens** | Inline error highlighting | `code --install-extension usernamehw.errorlens` |

#### Quick Extension Install
```bash
code --install-extension golang.go
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools  
code --install-extension redhat.vscode-yaml
code --install-extension eamodio.gitlens
code --install-extension rangav.vscode-thunder-client
code --install-extension usernamehw.errorlens
```

#### VS Code Settings for Go Development

Create `.vscode/settings.json` in your project root:

```json
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.testFlags": ["-v", "-race"],
    "go.testTimeout": "10s",
    "go.coverOnSave": false,
    "go.coverOnSingleTest": true,
    "editor.formatOnSave": true,
    "files.eol": "\n",
    "yaml.schemas": {
        "https://json.schemastore.org/kustomization": "kustomization.yaml",
        "kubernetes": "*.k8s.yaml"
    },
    "kubernetes.vs-kubernetes.outputFormat": "yaml"
}
```

### Alternative IDEs

#### GoLand (JetBrains)
```bash
# Professional IDE for Go development
# Download from: https://www.jetbrains.com/go/
# Excellent for complex refactoring and debugging
```

#### Vim/Neovim with vim-go
```bash
# For terminal-based development
# Install vim-go: https://github.com/fatih/vim-go
```

## Required Development Tools

### Go Development Environment

#### Install Go 1.21+
```bash
# Linux/macOS via official installer
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# macOS via Homebrew
brew install go

# Verify installation
go version
```

#### Configure Go Environment
```bash
# Add to your shell profile (.bashrc, .zshrc)
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH

# Reload shell
source ~/.bashrc  # or ~/.zshrc
```

#### Install Go Development Tools
```bash
# Essential Go tools for development
go install -a std
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Verify tools installation
goimports -h
golangci-lint --version
dlv version
```

### Container and Kubernetes Tools

These tools are required for OpenFrame CLI development:

#### Docker
```bash
# Ubuntu/Debian
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# macOS
brew install --cask docker

# Start Docker and verify
docker --version
docker run hello-world
```

#### kubectl
```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# macOS
brew install kubectl

# Verify installation
kubectl version --client
```

#### K3d
```bash
# Install K3d for local Kubernetes clusters
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Verify installation
k3d version
```

#### Helm
```bash
# Install Helm for package management
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# macOS alternative
brew install helm

# Verify installation
helm version
```

## Optional Development Tools

### Development Workflow Tools

#### Telepresence (for traffic interception)
```bash
# Linux
sudo curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.16.1/telepresence-linux-amd64 -o /usr/local/bin/telepresence
sudo chmod a+x /usr/local/bin/telepresence

# macOS
brew install datawire/blackbird/telepresence-oss

# Verify installation
telepresence version
```

#### Skaffold (for live development)
```bash
# Linux
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/

# macOS
brew install skaffold

# Verify installation
skaffold version
```

#### Make (for build automation)
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# macOS (usually pre-installed)
xcode-select --install
# Or via Homebrew
brew install make

# Verify installation
make --version
```

## Environment Variables for Development

Set up these environment variables for optimal development experience:

```bash
# Add to your shell profile (.bashrc, .zshrc)

# Go development
export GOPATH=$HOME/go
export GOROOT=/usr/local/go
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config
export K3D_FIX_DNS=1  # Fix DNS issues in K3d

# OpenFrame CLI development
export OPENFRAME_DEV=true
export OPENFRAME_LOG_LEVEL=debug

# Docker optimization
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Development convenience
export EDITOR=code  # or vim, nano, etc.
```

### Environment Validation Script

Create `scripts/validate-env.sh` to verify your setup:

```bash
#!/bin/bash

echo "🔍 Validating OpenFrame CLI Development Environment..."
echo

# Check Go
if command -v go &> /dev/null; then
    go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | head -1)
    echo "✅ Go $go_version found"
else
    echo "❌ Go not found - install Go 1.21+"
fi

# Check Docker
if command -v docker &> /dev/null && docker info &> /dev/null; then
    docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo "✅ Docker $docker_version running"
else
    echo "❌ Docker not found or not running"
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    kubectl_version=$(kubectl version --client --short 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+')
    echo "✅ kubectl $kubectl_version found"
else
    echo "❌ kubectl not found"
fi

# Check K3d
if command -v k3d &> /dev/null; then
    k3d_version=$(k3d version | grep k3d | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
    echo "✅ K3d $k3d_version found"
else
    echo "❌ K3d not found"
fi

# Check Helm
if command -v helm &> /dev/null; then
    helm_version=$(helm version --short | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
    echo "✅ Helm $helm_version found"
else
    echo "❌ Helm not found"
fi

# Check Go tools
echo
echo "🔧 Checking Go development tools..."

tools=("goimports" "golangci-lint" "dlv")
for tool in "${tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool found"
    else
        echo "⚠️  $tool not found - install with: go install <package>"
    fi
done

echo
echo "🎯 Environment validation complete!"
```

Make it executable and run:
```bash
chmod +x scripts/validate-env.sh
./scripts/validate-env.sh
```

## IDE Configuration for OpenFrame CLI

### VS Code Workspace Settings

Create `.vscode/tasks.json` for build tasks:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "build",
            "type": "shell",
            "command": "go",
            "args": ["build", "-o", "bin/openframe", "./cmd/openframe"],
            "group": {
                "kind": "build",
                "isDefault": true
            },
            "presentation": {
                "echo": true,
                "reveal": "silent",
                "focus": false,
                "panel": "shared"
            }
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
            }
        },
        {
            "label": "lint",
            "type": "shell",
            "command": "golangci-lint",
            "args": ["run"],
            "group": "build"
        }
    ]
}
```

### VS Code Launch Configuration

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
            "program": "${workspaceFolder}/cmd/openframe",
            "args": ["--help"],
            "env": {
                "OPENFRAME_DEV": "true",
                "OPENFRAME_LOG_LEVEL": "debug"
            },
            "console": "integratedTerminal"
        },
        {
            "name": "Debug Bootstrap Command",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/openframe",
            "args": ["bootstrap", "test-cluster", "--verbose"],
            "env": {
                "OPENFRAME_DEV": "true"
            },
            "console": "integratedTerminal"
        }
    ]
}
```

## Shell Enhancements for Development

### Useful Aliases

Add these to your `.bashrc` or `.zshrc`:

```bash
# OpenFrame development aliases
alias of='openframe'
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofch='openframe chart'
alias ofd='openframe dev'

# Go development aliases  
alias gb='go build'
alias gt='go test'
alias gr='go run'
alias gm='go mod'

# Kubernetes aliases
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
alias kgn='kubectl get nodes'
alias kdp='kubectl describe pod'

# Docker aliases
alias d='docker'
alias dc='docker compose'
alias dps='docker ps'
alias di='docker images'
```

### Auto-completion

Enable shell completion for tools:

```bash
# Add to your shell profile

# kubectl completion
source <(kubectl completion bash)  # or zsh

# helm completion  
source <(helm completion bash)     # or zsh

# docker completion (if available)
source <(docker completion bash)   # or zsh
```

## Next Steps

Once your environment is set up:

1. **[Local Development Guide](local-development.md)** - Clone and run OpenFrame CLI locally
2. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn our development process
3. **[Testing Overview](../testing/overview.md)** - Understand our testing approach

## Troubleshooting Environment Issues

### Go Module Issues
```bash
# Clean Go module cache
go clean -modcache

# Verify module status
go mod verify
go mod tidy
```

### Docker Permission Issues
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### K3d DNS Issues
```bash
# Set environment variable
export K3D_FIX_DNS=1

# Or use custom DNS
k3d cluster create --k3s-arg '--kubelet-arg=cluster-dns=1.1.1.1@server:0'
```

### IDE Performance Issues
```bash
# Increase VS Code memory limit
code --max-memory=8192

# Disable unused extensions
code --disable-extension <extension-id>
```

---

Your development environment is now ready! Proceed to the **[Local Development Guide](local-development.md)** to start working with the OpenFrame CLI codebase.