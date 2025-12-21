# Development Environment Setup

This guide walks you through setting up a complete development environment for OpenFrame CLI, including IDE configuration, required tools, and recommended extensions.

## Development Prerequisites

### Required Software

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| **Go** | 1.21+ | Primary programming language | [golang.org](https://golang.org/doc/install) |
| **Git** | 2.30+ | Version control | [git-scm.com](https://git-scm.com/) |
| **Make** | 3.81+ | Build automation | Usually pre-installed on Linux/macOS |
| **Docker** | 24.0+ | Container runtime for testing | [docker.com](https://docs.docker.com/get-docker/) |

### Optional but Recommended

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| **golangci-lint** | 1.54+ | Code linting and quality checks | [golangci-lint.run](https://golangci-lint.run/usage/install/) |
| **gofumpt** | latest | Enhanced code formatting | `go install mvdan.cc/gofumpt@latest` |
| **gotestsum** | latest | Enhanced test output | `go install gotest.tools/gotestsum@latest` |
| **air** | latest | Live reload for development | `go install github.com/cosmtrek/air@latest` |

## IDE Setup

### VS Code (Recommended)

VS Code provides excellent Go support with the official Go extension.

#### Essential Extensions
```bash
# Install via VS Code Extensions Marketplace
code --install-extension golang.Go
code --install-extension ms-vscode.vscode-json
code --install-extension redhat.vscode-yaml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
```

#### Recommended Extensions
```bash
# Additional productivity extensions
code --install-extension ms-vscode.test-adapter-converter
code --install-extension github.copilot
code --install-extension eamodio.gitlens
code --install-extension ms-vscode.makefile-tools
```

#### VS Code Configuration

Create `.vscode/settings.json` in your project root:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ],
  "go.formatTool": "gofumpt",
  "go.useLanguageServer": true,
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s",
  "go.coverOnSave": true,
  "go.coverageDisplayStyle": "gutter",
  "go.toolsManagement.checkForUpdates": "local",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml"
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
      "args": ["--help"],
      "console": "integratedTerminal"
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose"],
      "console": "integratedTerminal"
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "console": "integratedTerminal"
    }
  ]
}
```

### GoLand/IntelliJ IDEA

Professional IDE option with advanced Go support.

#### Configuration
1. **Go SDK**: Configure Go SDK path in `Settings > Go > GOROOT`
2. **Linter**: Enable golangci-lint in `Settings > Tools > Go Linter`
3. **Formatter**: Set gofumpt as formatter in `Settings > Tools > File Watchers`
4. **Live Templates**: Import Go-specific code templates

#### Live Templates
```go
// Custom live template for CLI commands
func Get$NAME$Cmd() *cobra.Command {
    return &cobra.Command{
        Use:   "$name$",
        Short: "$description$",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Implementation
            return nil
        },
    }
}
```

### Vim/Neovim

Lightweight editor with excellent Go support via vim-go.

#### Plugin Setup (using vim-plug)
```vim
" .vimrc or init.vim
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'preservim/nerdtree'
Plug 'junegunn/fzf.vim'

" Go-specific settings
let g:go_fmt_command = "gofumpt"
let g:go_auto_type_info = 1
let g:go_def_mapping_enabled = 0
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
```

## Environment Variables

Set up essential environment variables for OpenFrame development:

### Go Environment
```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile

# Go workspace and module proxy
export GOPATH=$HOME/go
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Development-specific settings
export CGO_ENABLED=0  # For static binaries
export GOOS=linux     # Target OS (adjust as needed)
export GOARCH=amd64   # Target architecture

# OpenFrame development
export OPENFRAME_DEV=true
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=$HOME/.openframe
```

### Testing Environment
```bash
# Test-specific settings
export OPENFRAME_TEST_CLUSTER=openframe-test
export OPENFRAME_TEST_TIMEOUT=300s
export OPENFRAME_INTEGRATION_TESTS=false  # Enable for CI
```

### Docker and Kubernetes
```bash
# Docker settings
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config
export KUBE_EDITOR=vim  # Or your preferred editor
```

## Shell Configuration

### Bash Completion
```bash
# Add to ~/.bashrc
if command -v openframe >/dev/null 2>&1; then
    source <(openframe completion bash)
fi

# Kubernetes completion
if command -v kubectl >/dev/null 2>&1; then
    source <(kubectl completion bash)
fi
```

### Zsh Completion
```zsh
# Add to ~/.zshrc
if command -v openframe >/dev/null 2>&1; then
    source <(openframe completion zsh)
fi

# Oh My Zsh plugins (if using)
plugins=(git golang kubectl docker)
```

### Useful Aliases
```bash
# Development aliases
alias of='openframe'
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofch='openframe chart'
alias ofd='openframe dev'

# Go development
alias gor='go run .'
alias got='go test ./...'
alias gob='go build .'
alias goi='go install .'
alias gom='go mod tidy'

# Kubernetes shortcuts
alias k='kubectl'
alias kg='kubectl get'
alias kd='kubectl describe'
alias kl='kubectl logs'
alias ke='kubectl edit'
```

## Development Tools Configuration

### golangci-lint Configuration

Create `.golangci.yml` in project root:

```yaml
run:
  timeout: 5m
  modules-download-mode: readonly
  
linters-settings:
  govet:
    check-shadowing: true
  golint:
    min-confidence: 0
  gocyclo:
    min-complexity: 15
  maligned:
    suggest-new: true
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 2

linters:
  enable-all: true
  disable:
    - exhaustivestruct
    - maligned
    - interfacer
    - golint
    - scopelint
    - gofumpt  # We use gofumpt directly

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - dupl
        - gomnd
```

### Air Configuration (Live Reload)

Create `.air.toml`:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["--help"]
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ."
  delay = 1000
  exclude_dir = ["tmp", "vendor", "testdata"]
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

## Verification

### Validate Development Setup
```bash
# Check Go installation
go version
go env GOPATH
go env GOMOD

# Verify development tools
golangci-lint version
gofumpt --version
make --version

# Test Docker and Kubernetes tools
docker version
kubectl version --client
k3d version
helm version
```

### Build and Test
```bash
# Clone and build OpenFrame CLI
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the project
make build

# Run tests
make test

# Check code quality
make lint

# Format code
make fmt
```

### IDE Integration Test
```bash
# Test VS Code Go integration
code .
# Open any .go file and verify:
# - Syntax highlighting works
# - Go to definition works (Ctrl/Cmd + click)
# - Auto-completion works
# - Linting shows in Problems panel
```

## Troubleshooting

### Common Issues

**Go modules not working**:
```bash
# Clear module cache
go clean -modcache
go mod download
```

**golangci-lint errors**:
```bash
# Update linter
golangci-lint --version
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**VS Code Go extension issues**:
```bash
# Restart Go language server
# Ctrl/Cmd + Shift + P > "Go: Restart Language Server"

# Check Go tools
go install -a golang.org/x/tools/gopls@latest
```

**Docker permission issues**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
# Log out and back in

# macOS: Ensure Docker Desktop is running
open /Applications/Docker.app
```

### Performance Optimization

**Improve build speed**:
```bash
# Enable Go module proxy
go env -w GOPROXY=https://proxy.golang.org,direct

# Use build cache
go env -w GOCACHE=$HOME/.cache/go-build

# Parallel builds
export GOMAXPROCS=$(nproc)
```

---

## Next Steps

Environment ready? Continue with:
- **[Local Development Guide](local-development.md)** - Clone, build, and run OpenFrame CLI
- **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure
- **[Testing Guide](../testing/overview.md)** - Learn the testing approach and run tests