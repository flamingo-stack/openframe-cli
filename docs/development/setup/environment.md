# Development Environment Setup

This guide walks you through setting up an optimal development environment for contributing to OpenFrame CLI.

## Prerequisites

Ensure you've completed the [Prerequisites guide](../../getting-started/prerequisites.md) before proceeding with development setup.

## IDE and Editor Setup

### VS Code (Recommended)

VS Code provides excellent Go support with these essential extensions:

#### Required Extensions
```bash
# Install via VS Code Extensions marketplace or CLI
code --install-extension golang.go
code --install-extension ms-vscode.vscode-json
code --install-extension redhat.vscode-yaml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
```

#### VS Code Settings
Create `.vscode/settings.json` in your project:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.useLanguageServer": true,
  "go.testFlags": ["-v", "-race"],
  "go.buildTags": "integration",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### VS Code Tasks
Create `.vscode/tasks.json` for common development tasks:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Build OpenFrame",
      "type": "shell",
      "command": "go",
      "args": ["build", "-o", "openframe", "./main.go"],
      "group": "build",
      "presentation": {
        "reveal": "always",
        "panel": "new"
      }
    },
    {
      "label": "Run Tests",
      "type": "shell",
      "command": "go",
      "args": ["test", "./..."],
      "group": "test",
      "presentation": {
        "reveal": "always",
        "panel": "new"
      }
    },
    {
      "label": "Lint Code",
      "type": "shell",
      "command": "golangci-lint",
      "args": ["run"],
      "group": "build",
      "presentation": {
        "reveal": "always",
        "panel": "new"
      }
    }
  ]
}
```

### GoLand (Alternative)

GoLand provides built-in Go support. Configure these settings:

1. **Go Modules**: Enable `Go → Go Modules → Enable Go modules integration`
2. **Code Style**: Set to `gofmt` formatting
3. **Inspections**: Enable all Go-related inspections
4. **Run Configurations**: Set up configurations for tests and builds

## Required Development Tools

### Go Environment

```bash
# Verify Go version (1.21+ required)
go version

# Configure Go environment
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH
export GO111MODULE=on

# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/cosmtrek/air@latest
go install golang.org/x/tools/cmd/godoc@latest
```

### Linting and Code Quality

#### golangci-lint Configuration

Create `.golangci.yml` in project root:

```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - errcheck
    - gofmt
    - goimports
    - golint
    - gosec
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - typecheck
    - unconvert
    - unused
    - varcheck
    - deadcode

linters-settings:
  golint:
    min-confidence: 0.8
  gosec:
    excludes:
      - G104 # Audit errors not checked
      - G204 # Subprocess launched with variable
  errcheck:
    check-type-assertions: true
    check-blank: true

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - errcheck
```

### Git Configuration

```bash
# Configure Git for development
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Set up Git hooks (optional)
git config core.hooksPath .githooks

# Configure Git to handle Go modules
git config --global url."git@github.com:".insteadOf "https://github.com/"
```

### Make and Build Tools

Install build automation tools:

```bash
# Install Make (if not already installed)
# Linux
sudo apt-get install make

# macOS
brew install make

# Verify installation
make --version
```

## Environment Variables

### Development Environment Variables

Create a `.env` file or add to your shell profile:

```bash
# OpenFrame Development
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=$HOME/.openframe

# Go Development
export GOPATH=$HOME/go
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org

# Kubernetes Development
export KUBECONFIG=$HOME/.kube/config
export KUBECTL_EXTERNAL_DIFF="code --diff --wait"

# Docker Development
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Testing
export OPENFRAME_TEST_TIMEOUT=30m
export OPENFRAME_INTEGRATION_TESTS=true
```

### Shell Profile Configuration

Add to your `~/.bashrc`, `~/.zshrc`, or equivalent:

```bash
# OpenFrame CLI development function
openframe-dev() {
    cd $GOPATH/src/github.com/flamingo-stack/openframe-cli
    export OPENFRAME_DEV_MODE=true
    export OPENFRAME_LOG_LEVEL=debug
    echo "OpenFrame development environment activated"
}

# Quick build and test
openframe-test() {
    go build -o openframe ./main.go && ./openframe "$@"
}

# Lint and format code
openframe-lint() {
    goimports -w .
    golangci-lint run
}
```

## Editor Extensions and Plugins

### Essential Extensions

| Editor | Extension | Purpose |
|--------|-----------|---------|
| **VS Code** | Go (by Google) | Go language support |
| **VS Code** | Kubernetes | Kubernetes YAML support |
| **VS Code** | YAML | YAML language support |
| **VS Code** | GitLens | Enhanced Git integration |
| **GoLand** | Kubernetes | Kubernetes integration |
| **Vim/Neovim** | vim-go | Go development support |

### Optional but Helpful

- **Thunder Client** (VS Code): API testing
- **Docker** (VS Code): Container management
- **Remote - Containers** (VS Code): Development in containers
- **Live Share** (VS Code): Collaborative development

## Debugging Configuration

### VS Code Debug Configuration

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
      "args": ["bootstrap", "--verbose"],
      "env": {
        "OPENFRAME_DEV_MODE": "true",
        "OPENFRAME_LOG_LEVEL": "debug"
      },
      "showLog": true
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}",
      "env": {
        "OPENFRAME_TEST_MODE": "true"
      },
      "args": [
        "-test.v",
        "-test.run",
        "TestBootstrap"
      ]
    }
  ]
}
```

### Command Line Debugging

```bash
# Debug with delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the CLI
dlv debug ./main.go -- bootstrap --verbose

# Debug tests
dlv test ./internal/bootstrap
```

## Performance and Monitoring

### Profiling Tools

```bash
# Install pprof
go install github.com/google/pprof@latest

# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=.

# Profile memory usage
go test -memprofile=mem.prof -bench=.

# Analyze profiles
pprof cpu.prof
pprof mem.prof
```

### Benchmarking

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# Compare benchmarks
go install golang.org/x/perf/cmd/benchcmp@latest
benchcmp old.txt new.txt
```

## Development Workflow Automation

### Air for Live Reloading

Create `.air.toml` for live reloading:

```toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = ["bootstrap", "--verbose"]
  bin = "./tmp/openframe"
  cmd = "go build -o ./tmp/openframe ./main.go"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
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

Start live reloading:
```bash
air
```

## Troubleshooting Common Issues

### Go Module Issues
```bash
# Clear module cache
go clean -modcache

# Reinitialize modules
go mod tidy
go mod vendor
```

### VS Code Go Extension Issues
```bash
# Restart Go language server
Ctrl+Shift+P → "Go: Restart Language Server"

# Update Go tools
Ctrl+Shift+P → "Go: Install/Update Tools"
```

### Build Issues
```bash
# Clean build cache
go clean -cache

# Rebuild everything
go build -a ./...
```

## Next Steps

With your development environment configured:

1. **[Local Development Guide](local-development.md)** - Clone and run the project locally
2. **[Architecture Overview](../architecture/README.md)** - Understand the system design
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the contribution process

> 💡 **Pro Tip**: Set up your development environment incrementally. Start with basic Go support, then add linting, debugging, and automation tools as you become more comfortable with the codebase.