# Development Environment Setup

This guide helps you configure an optimal development environment for working on OpenFrame CLI. Follow these steps to set up your IDE, tools, and development workflow.

## IDE Recommendations and Setup

### Visual Studio Code (Recommended)

VS Code provides excellent Go support and integrates well with the OpenFrame CLI development workflow.

#### Required Extensions

Install these essential extensions:

```bash
# Install via command line
code --install-extension golang.go
code --install-extension ms-vscode.vscode-yaml  
code --install-extension redhat.vscode-xml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension ms-vscode.makefile-tools
```

| Extension | Purpose | Key Features |
|-----------|---------|--------------|
| **Go** | Go language support | IntelliSense, debugging, testing |
| **YAML** | YAML editing | Kubernetes manifest editing |
| **Kubernetes** | K8s integration | Cluster management, manifest validation |
| **Makefile Tools** | Makefile support | Build task integration |

#### Recommended Extensions

```bash
# Additional helpful extensions
code --install-extension streetsidesoftware.code-spell-checker
code --install-extension davidanson.vscode-markdownlint
code --install-extension github.copilot                    # If you have access
```

#### VS Code Configuration

Create `.vscode/settings.json` in the project root:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s",
  "go.coverOnSingleTest": true,
  "go.coverageDisplayStyle": "file",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "yaml.schemas": {
    "kubernetes": "*.yaml",
    "https://json.schemastore.org/github-workflow": ".github/workflows/*.yml"
  },
  "files.exclude": {
    "bin/": true,
    "dist/": true,
    "*.exe": true
  }
}
```

#### VS Code Tasks Configuration

Create `.vscode/tasks.json` for integrated build tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      },
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
      "group": "test"
    }
  ]
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
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "debug-cluster"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      },
      "showLog": true
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch", 
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    }
  ]
}
```

### GoLand/IntelliJ IDEA

For JetBrains IDE users:

#### Configuration Steps
1. **Install Go Plugin** (if not already installed)
2. **Set GOPATH and GOROOT** in Settings → Go → GOPATH
3. **Enable Go Modules** in Settings → Go → Go Modules  
4. **Configure Code Style** using `gofmt` settings

#### Useful Plugins
- **Kubernetes**: YAML support and cluster integration
- **Makefile Language**: Build script support  
- **GitToolBox**: Enhanced Git integration

## Required Development Tools

### Go Development Environment

```bash
# Install Go (if not already installed)
# macOS
brew install go

# Linux
wget https://golang.org/dl/go1.19.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.19.linux-amd64.tar.gz

# Verify installation
go version  # Should show Go 1.19+

# Set up Go workspace (add to ~/.bashrc or ~/.zshrc)
export GOPATH=$HOME/go
export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
```

### Code Quality Tools

Install essential Go development tools:

```bash
# golangci-lint for comprehensive linting
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1

# Additional Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/air-verse/air@latest  # Hot reload for development
go install github.com/securecodewarrior/sast-scan@latest  # Security scanning
```

### Build and Automation Tools

```bash
# Make (usually pre-installed)
make --version

# If not installed:
# macOS: xcode-select --install
# Linux: sudo apt-get install build-essential

# Docker for integration testing
docker --version

# K3d for cluster testing  
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

## Environment Variables for Development

### Required Variables

Set these environment variables for development:

```bash
# Add to ~/.bashrc, ~/.zshrc, or your shell config
export OPENFRAME_LOG_LEVEL=debug          # Enable debug logging
export OPENFRAME_DEV_MODE=true            # Enable development features
export KUBECONFIG=$HOME/.kube/config      # Kubernetes configuration
export GO111MODULE=on                     # Enable Go modules (if needed)
```

### Optional Variables

```bash
# Performance and debugging
export GOGC=100                           # Go garbage collection tuning
export GODEBUG=gctrace=1                  # GC trace logging (for performance analysis)

# Development workflow
export OPENFRAME_CONFIG_DIR=$HOME/.config/openframe  # Custom config directory
export DOCKER_HOST=unix:///var/run/docker.sock       # Docker daemon (if non-standard)
```

### Environment Variables File

Create a `.env` file in the project root for local development:

```bash
# .env file (add to .gitignore)
OPENFRAME_LOG_LEVEL=debug
OPENFRAME_DEV_MODE=true
KUBECONFIG=/home/dev/.kube/config
DOCKER_HOST=unix:///var/run/docker.sock
```

Load with:
```bash
# Source environment file
set -a && source .env && set +a
```

## Editor Extensions and Plugins

### Essential Language Support

| Language/Format | VS Code Extension | Purpose |
|-----------------|-------------------|---------|
| **Go** | `golang.go` | Language server, debugging, testing |
| **YAML** | `redhat.vscode-yaml` | Kubernetes manifests, CI/CD files |
| **Markdown** | `yzhang.markdown-all-in-one` | Documentation editing |
| **JSON** | Built-in | Configuration files |

### Development Workflow Extensions

| Extension | Purpose | Benefit |
|-----------|---------|---------|
| **GitLens** | Advanced Git integration | Blame, history, comparisons |
| **Thunder Client** | API testing | Test ArgoCD APIs, webhooks |
| **Error Lens** | Inline error display | Quick issue identification |
| **Todo Tree** | TODO/FIXME highlighting | Track development tasks |

### Kubernetes Development

```bash
# Install kubectl for VS Code integration
# VS Code extension: ms-kubernetes-tools.vscode-kubernetes-tools

# Helm extension for chart development
# Extension: ms-kubernetes-tools.vscode-helm

# YAML schema validation
# Extension: redhat.vscode-yaml with Kubernetes schemas
```

## Development Workflow Integration

### Git Configuration

```bash
# Set up Git for the project
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Enable helpful Git features
git config --global pull.rebase true
git config --global init.defaultBranch main
git config --global core.editor "code --wait"  # Use VS Code as editor
```

### Pre-commit Hooks (Optional)

Set up pre-commit hooks for code quality:

```bash
# Install pre-commit (optional but recommended)
pip install pre-commit

# Create .pre-commit-config.yaml
cat > .pre-commit-config.yaml << 'EOF'
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
  
  - repo: https://github.com/securecodewarrior/sast-scan
    rev: master
    hooks:
      - id: secrets-scan
EOF

# Install hooks
pre-commit install
```

### Shell Aliases for Development

Add these aliases to your shell configuration:

```bash
# Add to ~/.bashrc or ~/.zshrc

# OpenFrame development aliases
alias of-build="make build"
alias of-test="make test" 
alias of-lint="make lint"
alias of-run="./bin/openframe"

# Kubernetes development aliases  
alias k="kubectl"
alias kgp="kubectl get pods"
alias kgs="kubectl get services"
alias kdp="kubectl describe pod"

# Git aliases
alias gs="git status"
alias gc="git commit"
alias gp="git pull"
alias gb="git branch"
```

## Verification and Testing

### Verify Development Environment

Run these commands to verify your setup:

```bash
# Check Go installation
go version
go env GOPATH
go env GOROOT

# Check development tools
golangci-lint version
make --version
docker --version
k3d version

# Check environment variables
echo $OPENFRAME_LOG_LEVEL
echo $KUBECONFIG

# Test project build
cd openframe-cli
make build
./bin/openframe --version
```

### Test IDE Integration

```bash
# Open project in VS Code
code .

# Verify Go extension is working:
# 1. Open any .go file
# 2. Check syntax highlighting
# 3. Hover over Go symbols (should show documentation)
# 4. Try "Go to Definition" (F12)

# Test debugging:
# 1. Set a breakpoint in main.go
# 2. Run debug configuration
# 3. Verify debugger stops at breakpoint
```

## Troubleshooting Common Issues

### Go Module Issues

```bash
# Clean Go module cache
go clean -modcache

# Re-download dependencies
go mod download

# Update go.sum
go mod tidy
```

### VS Code Go Extension Issues

```bash
# Restart Go language server
# Command Palette (Ctrl+Shift+P) → "Go: Restart Language Server"

# Check Go tools installation
# Command Palette → "Go: Install/Update Tools"

# Clear VS Code workspace cache
rm -rf .vscode/settings.json
# Then reconfigure
```

### Environment Variable Issues

```bash
# Check all environment variables
printenv | grep -E "(GO|OPENFRAME|KUBE|DOCKER)"

# Reset environment
unset GOPATH GOROOT
# Then follow installation steps again
```

## Next Steps

After setting up your development environment:

1. **[Continue to Local Development Guide](./local-development.md)** to clone and build the project
2. **[Review Architecture Overview](../architecture/overview.md)** to understand the codebase
3. **[Read Contributing Guidelines](../contributing/guidelines.md)** before making changes

---

Your development environment is now optimized for OpenFrame CLI development! 🚀