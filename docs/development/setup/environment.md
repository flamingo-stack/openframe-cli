# Development Environment Setup

This guide will help you set up a productive development environment for contributing to OpenFrame CLI. We'll cover IDE configuration, essential tools, debugging setup, and development workflows.

## Development Prerequisites

### Required Tools

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| **Go** | 1.21+ | Core language | [Download Go](https://golang.org/dl/) |
| **Git** | 2.30+ | Version control | [Install Git](https://git-scm.com/downloads) |
| **Docker** | 20.0+ | Container runtime | [Docker Desktop](https://docker.com/products/docker-desktop) |
| **Make** | Any | Build automation | Usually pre-installed (Linux/macOS) |
| **kubectl** | 1.25+ | Kubernetes CLI | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |

### Recommended Tools

| Tool | Purpose | Why Recommended |
|------|---------|----------------|
| **golangci-lint** | Code linting | Ensures code quality and consistency |
| **k3d** | Local K8s clusters | Testing cluster operations locally |
| **jq** | JSON processing | Debugging API responses |
| **direnv** | Environment management | Auto-load project environment variables |
| **pre-commit** | Git hooks | Automated code quality checks |

## IDE Configuration

### VS Code (Recommended)

VS Code provides excellent Go support and integrates well with the development workflow.

#### Essential Extensions
```json
{
  "recommendations": [
    "golang.go",                    // Official Go extension
    "ms-vscode.vscode-json",       // JSON support  
    "redhat.vscode-yaml",          // YAML support
    "ms-kubernetes-tools.vscode-kubernetes-tools", // K8s support
    "github.copilot",              // AI assistance (optional)
    "streetsidesoftware.code-spell-checker" // Spell checking
  ]
}
```

#### VS Code Settings
Create `.vscode/settings.json` in your project:
```json
{
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "workspace",
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "300s",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,64,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.5)"
  },
  "files.exclude": {
    "**/vendor/**": true,
    "**/bin/**": true
  },
  "yaml.schemas": {
    "https://json.schemastore.org/kustomization": "kustomization.yaml"
  }
}
```

#### VS Code Tasks
Create `.vscode/tasks.json`:
```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build OpenFrame",
      "type": "shell",
      "command": "make",
      "args": ["build"],
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
      "label": "Run Tests",
      "type": "shell",
      "command": "make",
      "args": ["test"],
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
      "group": "test"
    }
  ]
}
```

#### VS Code Launch Configuration
Create `.vscode/launch.json` for debugging:
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug OpenFrame",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "--help"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      }
    },
    {
      "name": "Debug Cluster Create",
      "type": "go",
      "request": "launch",
      "mode": "debug", 
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "debug-cluster", "--nodes", "1"],
      "env": {
        "OPENFRAME_DEBUG": "true"
      }
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/cluster",
      "env": {
        "OPENFRAME_TEST_DEBUG": "true"
      }
    }
  ]
}
```

### GoLand/IntelliJ IDEA

For JetBrains IDE users:

#### Essential Plugins
- **Go plugin** (usually pre-installed)
- **Docker plugin**
- **Kubernetes plugin** 
- **YAML/Ansible Support**
- **Makefile Language**

#### GoLand Configuration
1. **File → Settings → Go → Go Modules**
   - Enable "Enable Go modules integration"
   - Set "Proxy" to "direct" for faster builds

2. **File → Settings → Tools → Go Tools**
   - Set "goimports" as formatter
   - Enable "On save" formatting
   - Set golangci-lint as linter

3. **Run Configurations**
   - Create "Go Build" configuration for `main.go`
   - Add environment variables for debugging
   - Set working directory to project root

### Vim/Neovim

For terminal-based development:

#### Essential Plugins (vim-plug)
```vim
" Language support
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}

" File navigation
Plug 'junegunn/fzf', { 'do': { -> fzf#install() } }
Plug 'junegunn/fzf.vim'

" Git integration  
Plug 'tpope/vim-fugitive'
Plug 'airblade/vim-gitgutter'

" Status line
Plug 'vim-airline/vim-airline'
Plug 'vim-airline/vim-airline-themes'
```

#### Vim Configuration
Add to your `.vimrc` or `init.vim`:
```vim
" Go configuration
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_function_calls = 1
let g:go_highlight_operators = 1
let g:go_highlight_extra_types = 1

" Key mappings for Go
au FileType go nmap <leader>r <Plug>(go-run)
au FileType go nmap <leader>b <Plug>(go-build)
au FileType go nmap <leader>t <Plug>(go-test)
au FileType go nmap <leader>c <Plug>(go-coverage)

" CoC configuration for autocompletion
inoremap <silent><expr> <TAB>
  \ pumvisible() ? "\<C-n>" :
  \ <SID>check_back_space() ? "\<TAB>" :
  \ coc#refresh()
```

## Environment Setup

### Go Environment
```bash
# Check Go installation
go version  # Should be 1.21+

# Configure Go environment
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Enable Go modules (should be default in Go 1.16+)
export GO111MODULE=on

# Add to ~/.bashrc or ~/.zshrc for persistence
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
```

### Install Development Tools
```bash
# Install golangci-lint for code quality
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2

# Install additional Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest  # Debugger
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install k3d for local testing
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Install pre-commit for git hooks
pip install pre-commit  # or brew install pre-commit
```

### Environment Variables
Create a `.env` file in your project root:
```bash
# Development configuration
OPENFRAME_DEBUG=true
OPENFRAME_LOG_LEVEL=debug

# Test configuration  
OPENFRAME_TEST_TIMEOUT=300s
OPENFRAME_TEST_PARALLEL=4

# Docker configuration
DOCKER_HOST=unix:///var/run/docker.sock

# K8s configuration
KUBECONFIG=$HOME/.kube/config
```

### direnv Configuration (Optional)
If using direnv, create `.envrc`:
```bash
#!/bin/bash

# Load environment variables
dotenv_if_exists .env

# Add project bin to PATH
PATH_add ./bin

# Go configuration
export GO111MODULE=on
export CGO_ENABLED=0

# Development helpers
export OPENFRAME_LOCAL_DEV=true

# Source completion if available
if [[ -f ./scripts/completion.bash ]]; then
    source ./scripts/completion.bash
fi
```

Then run:
```bash
direnv allow .
```

## Development Workflow Setup

### Git Configuration
```bash
# Configure Git for the project
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Set up useful aliases
git config alias.co checkout
git config alias.br branch
git config alias.ci commit
git config alias.st status
git config alias.unstage 'reset HEAD --'
git config alias.last 'log -1 HEAD'
git config alias.visual '!gitk'

# Configure line endings
git config core.autocrlf input  # Unix/Mac
# git config core.autocrlf true   # Windows
```

### Pre-commit Hooks
Set up automated code quality checks:

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
        entry: gofmt -l -s
        language: golang
        types: [go]
        
      - id: go-imports
        name: go imports  
        entry: goimports
        language: golang
        types: [go]
        args: [-local, github.com/flamingo-stack/openframe-cli]
        
      - id: golangci-lint
        name: golangci-lint
        entry: golangci-lint run --fix
        language: golang
        types: [go]
        pass_filenames: false
```

Install the hooks:
```bash
pre-commit install
```

### Makefile Configuration
Ensure you understand the project Makefile:
```bash
# View available targets
make help

# Common development commands
make build      # Build the binary
make test       # Run tests
make lint       # Run linters
make clean      # Clean build artifacts
make install    # Install to GOPATH/bin
```

## Debugging Setup

### VS Code Debugging
Use the launch configurations provided above. Key debugging features:
- **Breakpoints**: Click in the gutter or press F9
- **Step Through**: F10 (step over), F11 (step into)
- **Variable Inspection**: Hover over variables or use the Variables pane
- **Watch Expressions**: Add variables to watch their values

### Command Line Debugging with Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the main program
dlv debug . -- cluster create test-cluster

# Debug a specific test
dlv test ./internal/cluster -- -test.run TestClusterCreate

# Debug with breakpoints
dlv debug . -- cluster create test-cluster
(dlv) break main.main
(dlv) continue
```

### Logging and Tracing
Enable detailed logging during development:
```bash
# Set debug environment variables
export OPENFRAME_DEBUG=true
export OPENFRAME_LOG_LEVEL=trace

# Run with verbose output
go run . --verbose cluster create debug-cluster

# Enable Go race detector
go run -race . cluster create test-cluster
```

## Performance Profiling

### CPU Profiling
```bash
# Add profiling to your test
go test -cpuprofile cpu.prof -bench . ./internal/cluster

# Analyze profile
go tool pprof cpu.prof
(pprof) top
(pprof) web  # Opens browser visualization
```

### Memory Profiling
```bash
# Memory profiling
go test -memprofile mem.prof -bench . ./internal/cluster

# Analyze memory usage
go tool pprof mem.prof
(pprof) top
(pprof) list <function_name>
```

## Troubleshooting Development Environment

### Common Issues

#### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Reinitialize modules
rm go.sum
go mod tidy
go mod download
```

#### Docker Issues
```bash
# Check Docker is running
docker info

# Reset Docker if needed
docker system prune -a
```

#### IDE Issues
```bash
# Reload VS Code Go extension
# Ctrl+Shift+P -> "Go: Restart Language Server"

# Clear Go tool cache
go clean -cache
```

### Development Environment Verification

Run this script to verify your setup:
```bash
#!/bin/bash
echo "🔍 Verifying development environment..."

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
else
    echo "❌ Go: Not installed"
fi

# Check Docker
if command -v docker &> /dev/null && docker info &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
else
    echo "❌ Docker: Not running"
fi

# Check golangci-lint
if command -v golangci-lint &> /dev/null; then
    echo "✅ golangci-lint: $(golangci-lint --version)"
else
    echo "❌ golangci-lint: Not installed"
fi

# Check Make
if command -v make &> /dev/null; then
    echo "✅ Make: Available"
else
    echo "❌ Make: Not installed"
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    echo "✅ kubectl: $(kubectl version --client --short)"
else
    echo "❌ kubectl: Not installed"
fi

# Check k3d
if command -v k3d &> /dev/null; then
    echo "✅ k3d: $(k3d version | head -1)"
else
    echo "❌ k3d: Not installed"
fi

echo ""
echo "🚀 Ready to start development? Run 'make help' to see available commands."
```

---

**Next**: Ready to start coding? Check out [Local Development](local-development.md) to clone and build the project, or dive into the [Architecture Overview](../architecture/overview.md) to understand the codebase structure.