# Development Environment Setup

Configure your development environment for optimal OpenFrame CLI development experience with the right tools, IDE settings, and extensions.

## 🎯 Development Environment Goals

- **Efficient Code Development** - Fast iteration and debugging
- **Code Quality** - Automated formatting, linting, and validation  
- **Testing Integration** - Seamless test execution and coverage
- **Debugging Support** - Step-through debugging and profiling
- **Git Workflow** - Streamlined version control operations

## 🛠️ Required Development Tools

### Go Development Environment

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Go** | 1.21+ | Core language runtime | [golang.org](https://golang.org/dl/) |
| **Git** | 2.30+ | Version control | System package manager |
| **Make** | 4.0+ | Build automation | System package manager |
| **Docker** | 20.10+ | Container operations | [Docker Desktop](https://docker.com) |

### Go Tools and Utilities

Install essential Go development tools:

```bash
# Linting and code quality
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing and mocking
go install github.com/vektra/mockery/v2@latest
go install gotest.tools/gotestsum@latest

# Build and release tools
go install github.com/goreleaser/goreleaser@latest

# Code generation
go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest

# Debugging and profiling
go install github.com/go-delve/delve/cmd/dlv@latest
```

Verify installation:

```bash
# Check Go installation and workspace
go version
go env GOROOT GOPATH

# Verify installed tools
golangci-lint --version
mockery --version
goreleaser --version
```

## 🔧 IDE Configuration

### Visual Studio Code (Recommended)

Install VS Code and essential extensions:

#### Required Extensions

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "ms-kubernetes-tools.vscode-kubernetes-tools",
    "ms-vscode.docker",
    "eamodio.gitlens",
    "github.copilot"
  ]
}
```

Install extensions:
```bash
# Install Go extension (includes many tools)
code --install-extension golang.go

# Install Kubernetes tools
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools

# Install Docker support
code --install-extension ms-vscode.docker

# Install YAML/JSON support
code --install-extension redhat.vscode-yaml
code --install-extension ms-vscode.vscode-json

# Install Git tools
code --install-extension eamodio.gitlens

# Optional: GitHub Copilot for AI assistance
code --install-extension github.copilot
```

#### VS Code Settings

Create `.vscode/settings.json` in your workspace:

```json
{
  "go.useLanguageServer": true,
  "go.languageServerExperimentalFeatures": {
    "diagnostics": true,
    "documentLink": true
  },
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "300s",
  "go.coverOnSave": true,
  "go.coverOnSingleTest": true,
  "go.buildOnSave": "workspace",
  "go.formatTool": "goimports",
  "go.generateTestsFlags": [
    "-all",
    "-exported"
  ],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml"
  },
  "yaml.schemas": {
    "kubernetes": [
      "k8s-*.yaml",
      "kustomization.yaml"
    ]
  }
}
```

#### Launch Configuration

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
        "OPENFRAME_DEV_MODE": "true",
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "args": ["-test.v", "-test.run", "TestBootstrap"]
    },
    {
      "name": "Debug Current File Tests",
      "type": "go", 
      "request": "launch",
      "mode": "test",
      "program": "${fileDirname}",
      "args": ["-test.v"]
    }
  ]
}
```

### GoLand/IntelliJ IDEA

For JetBrains IDEs, configure:

#### Essential Plugins

- **Go** (bundled)
- **Kubernetes** 
- **Docker**
- **YAML/Ansible Support**
- **GitToolBox**

#### Configuration Steps

1. **Import Project**: Open the `openframe-cli` directory
2. **Configure Go SDK**: Settings → Go → GOROOT (should auto-detect)
3. **Enable Go Modules**: Settings → Go → Go Modules → Enable
4. **Configure Linter**: Settings → Tools → Go → Golangci-lint
5. **Set up Run Configurations**: Add configurations for main.go and tests

### Vim/Neovim

For terminal-based development:

#### vim-go Configuration

```vim
" Essential vim-go settings
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
let g:go_highlight_fields = 1
let g:go_highlight_types = 1
let g:go_highlight_operators = 1
let g:go_highlight_build_constraints = 1
let g:go_metalinter_enabled = ['vet', 'golint', 'errcheck']
let g:go_metalinter_autosave = 1

" Key mappings
au FileType go nmap <leader>r <Plug>(go-run)
au FileType go nmap <leader>b <Plug>(go-build)
au FileType go nmap <leader>t <Plug>(go-test)
au FileType go nmap <leader>c <Plug>(go-coverage)
```

## 🐚 Shell Environment

### Environment Variables

Add to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
# OpenFrame CLI development
export OPENFRAME_DEV_MODE="true"
export OPENFRAME_LOG_LEVEL="debug"
export OPENFRAME_CONFIG_DIR="$HOME/.openframe/dev"

# Go development
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

# Kubernetes development
export KUBECONFIG="$HOME/.kube/config"
export K3D_FIX_DNS=1

# Docker development
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

### Shell Aliases

Add helpful aliases for development:

```bash
# OpenFrame CLI aliases
alias of="openframe"
alias ofb="openframe bootstrap"
alias ofc="openframe cluster"
alias ofd="openframe dev"

# Go development aliases
alias gob="go build"
alias gor="go run"
alias got="go test"
alias gotv="go test -v"
alias gotr="go test -race"
alias gocov="go test -cover"

# Kubernetes aliases
alias k="kubectl"
alias kgp="kubectl get pods"
alias kgs="kubectl get services"
alias kga="kubectl get all"
alias kdp="kubectl describe pod"

# Docker aliases
alias dk="docker"
alias dkc="docker-compose"
alias dki="docker images"
alias dkp="docker ps"
```

### Shell Completions

Enable command completion for better productivity:

```bash
# Add to your shell profile
source <(openframe completion bash)  # For bash
source <(openframe completion zsh)   # For zsh
source <(kubectl completion bash)    # Kubernetes completion
```

## 🧪 Testing Environment

### Test Dependencies

Ensure testing tools are available:

```bash
# Test runner with better output
go install gotest.tools/gotestsum@latest

# Mock generation
go install github.com/vektra/mockery/v2@latest

# Coverage tools
go install golang.org/x/tools/cmd/cover@latest
```

### Test Configuration

Create test-specific environment variables:

```bash
# Test environment
export OPENFRAME_TEST_MODE="true"
export OPENFRAME_TEST_CLUSTER="openframe-test"
export OPENFRAME_TEST_NAMESPACE="openframe-test-ns"

# Disable interactive prompts in tests
export OPENFRAME_NON_INTERACTIVE="true"
```

### IDE Test Integration

Configure your IDE for optimal testing:

**VS Code:**
- Use the Go extension's built-in test discovery
- Configure test flags in settings.json
- Use the Test Explorer for visual test management

**GoLand:**
- Use built-in Go test runner
- Configure test templates and live templates
- Enable test coverage highlighting

## 🔍 Debugging Setup

### Delve Debugger

Configure Delve for advanced debugging:

```bash
# Install latest Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug OpenFrame CLI
dlv debug main.go -- bootstrap --verbose

# Debug tests
dlv test ./internal/bootstrap -- -test.v
```

### Debug Environment Variables

Set up debugging environment:

```bash
# Enable detailed logging
export OPENFRAME_LOG_LEVEL="trace"
export OPENFRAME_DEBUG_COMMANDS="true"

# Enable pprof profiling
export OPENFRAME_ENABLE_PPROF="true"
export OPENFRAME_PPROF_PORT="6060"
```

## 📋 Pre-commit Setup

Install and configure pre-commit hooks:

```bash
# Install pre-commit (if not already installed)
pip install pre-commit

# Install hooks in repository
pre-commit install

# Run hooks manually
pre-commit run --all-files
```

### Git Hooks Configuration

Create `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.4.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-json
      
  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-vet-mod
      - id: go-mod-tidy
      - id: golangci-lint
```

## ✅ Environment Verification

### Verification Script

Create a script to verify your development environment:

```bash
#!/bin/bash
# verify-dev-env.sh

echo "🔍 Verifying OpenFrame CLI Development Environment"

# Check Go installation
echo "Checking Go..."
go version || { echo "❌ Go not installed"; exit 1; }
echo "✅ Go installed"

# Check required tools
echo "Checking development tools..."
golangci-lint --version >/dev/null || { echo "❌ golangci-lint not installed"; exit 1; }
mockery --version >/dev/null || { echo "❌ mockery not installed"; exit 1; }
docker --version >/dev/null || { echo "❌ Docker not installed"; exit 1; }
kubectl version --client >/dev/null || { echo "❌ kubectl not installed"; exit 1; }
echo "✅ Development tools installed"

# Check environment variables
echo "Checking environment..."
[[ -n "$GOPATH" ]] || { echo "❌ GOPATH not set"; exit 1; }
[[ -n "$KUBECONFIG" ]] || { echo "❌ KUBECONFIG not set"; exit 1; }
echo "✅ Environment configured"

# Test Go modules
echo "Testing Go modules..."
go mod tidy && go mod verify || { echo "❌ Go modules issue"; exit 1; }
echo "✅ Go modules working"

echo "🎉 Development environment ready!"
```

Make it executable and run:

```bash
chmod +x verify-dev-env.sh
./verify-dev-env.sh
```

## 🎯 Next Steps

With your development environment configured:

1. **[Local Development Setup](local-development.md)** - Clone and build OpenFrame CLI
2. **[Architecture Overview](../architecture/README.md)** - Understand the codebase structure
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the development workflow

Your development environment is now optimized for OpenFrame CLI development. Happy coding! 🚀