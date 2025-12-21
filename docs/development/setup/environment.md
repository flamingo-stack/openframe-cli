# Development Environment Setup

This guide will help you set up a complete development environment for contributing to OpenFrame CLI. Follow these steps to configure your IDE, install required tools, and set up editor extensions for optimal productivity.

## Prerequisites

Before setting up your development environment, ensure you have completed the [Prerequisites Guide](../../getting-started/prerequisites.md) for the basic tools (Docker, kubectl, helm, k3d).

## Development Tools

### Core Development Requirements

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Go** | 1.19+ | Core language | [Install Go](https://golang.org/doc/install) |
| **Git** | 2.30+ | Version control | [Install Git](https://git-scm.com/downloads) |
| **Make** | 3.8+ | Build automation | Usually pre-installed on Unix systems |
| **golangci-lint** | 1.50+ | Code linting | [Install golangci-lint](https://golangci-lint.run/usage/install/) |
| **goimports** | Latest | Import formatting | `go install golang.org/x/tools/cmd/goimports@latest` |

### Optional but Recommended

| Tool | Purpose | Installation |
|------|---------|--------------|
| **dlv (Delve)** | Go debugger | `go install github.com/go-delve/delve/cmd/dlv@latest` |
| **air** | Live reload for Go | `go install github.com/cosmtrek/air@latest` |
| **goreleaser** | Release automation | [Install GoReleaser](https://goreleaser.com/install/) |
| **mockgen** | Mock generation for tests | `go install github.com/golang/mock/mockgen@latest` |

## IDE Configuration

### Visual Studio Code (Recommended)

#### Required Extensions

```bash
# Install via VS Code extensions marketplace or command palette
code --install-extension golang.go
code --install-extension ms-vscode.vscode-yaml
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension redhat.vscode-yaml
```

#### VS Code Settings

Create or update `.vscode/settings.json` in your project:

```json
{
    "go.lintTool": "golangci-lint",
    "go.lintFlags": [
        "--fast"
    ],
    "go.useLanguageServer": true,
    "go.buildTags": "integration",
    "go.testTags": "integration",
    "go.formatTool": "goimports",
    "go.generateTestsFlags": [
        "-template=testify"
    ],
    "files.exclude": {
        "**/vendor": true,
        "**/tmp": true,
        "**/.git": true
    },
    "yaml.schemas": {
        "kubernetes": [
            "k8s/**/*.yaml",
            "k8s/**/*.yml"
        ]
    },
    "kubernetes.outputFormat": "yaml"
}
```

#### VS Code Tasks

Create `.vscode/tasks.json`:

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
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
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
            "name": "Debug OpenFrame Bootstrap",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["bootstrap", "--deployment-mode=oss-tenant", "--verbose"],
            "console": "integratedTerminal"
        },
        {
            "name": "Debug OpenFrame Cluster Create",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/main.go",
            "args": ["cluster", "create", "test-cluster", "--skip-wizard"],
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

#### Configuration Steps

1. **Install Go Plugin** (if using IntelliJ IDEA)
2. **Configure Go SDK**:
   - File → Settings → Languages & Frameworks → Go → GOROOT
   - Point to your Go installation directory

3. **Set up Code Style**:
   - File → Settings → Editor → Code Style → Go
   - Enable "Use tab character" and set tab size to 4

4. **Configure External Tools**:
   - File → Settings → Tools → External Tools
   - Add tools for golangci-lint, goimports, etc.

#### GoLand Run Configurations

Create run configurations for common tasks:

```xml
<!-- .idea/runConfigurations/Bootstrap_Test.xml -->
<component name="ProjectRunConfigurationManager">
  <configuration default="false" name="Bootstrap Test" type="GoApplicationRunConfiguration" factoryName="Go Application">
    <module name="openframe-cli" />
    <working_directory value="$PROJECT_DIR$" />
    <parameters value="bootstrap test-cluster --deployment-mode=oss-tenant --verbose" />
    <kind value="PACKAGE" />
    <package value="." />
    <directory value="$PROJECT_DIR$" />
    <filePath value="$PROJECT_DIR$" />
    <method v="2" />
  </configuration>
</component>
```

### Vim/Neovim

#### Required Plugins

Using vim-plug:

```vim
" .vimrc or init.vim
call plug#begin()
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'preservim/nerdtree'
Plug 'ctrlpvim/ctrlp.vim'
Plug 'tpope/vim-fugitive'
call plug#end()

" Go configuration
let g:go_fmt_command = "goimports"
let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_function_calls = 1
```

#### CoC Configuration

Create `~/.config/nvim/coc-settings.json`:

```json
{
    "go.goplsPath": "gopls",
    "go.goplsArgs": ["-remote=auto"],
    "languageserver": {
        "golang": {
            "command": "gopls",
            "rootPatterns": ["go.mod", ".vim/", ".git/", ".hg/"],
            "filetypes": ["go"]
        }
    }
}
```

## Environment Variables

### Required Development Variables

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Go configuration
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

# OpenFrame development
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=$HOME/.config/openframe

# Development cluster defaults
export OPENFRAME_DEV_CLUSTER_NAME="openframe-dev"
export OPENFRAME_DEV_NODES=3

# Testing configuration
export OPENFRAME_TEST_TIMEOUT=10m
export OPENFRAME_TEST_CLEANUP=true

# Enable go mod cache
export GOCACHE=$HOME/.cache/go-build
export GOMODCACHE=$GOPATH/pkg/mod
```

### Development-Specific Variables

```bash
# Enable debug logging for development
export DEBUG=true
export LOG_LEVEL=debug

# Local registry for testing
export OPENFRAME_LOCAL_REGISTRY=localhost:5000

# Development tools
export EDITOR=code  # or vim, emacs, etc.
export BROWSER=firefox  # for opening ArgoCD UI

# Git configuration for commits
export GIT_AUTHOR_NAME="Your Name"
export GIT_AUTHOR_EMAIL="your.email@example.com"
```

## Project Configuration

### Git Configuration

Set up Git hooks and configuration:

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Set up Git hooks (if available)
cp scripts/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit

# Configure Git for the project
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Set up useful Git aliases
git config alias.st status
git config alias.co checkout
git config alias.br branch
git config alias.up "pull --rebase"
```

### EditorConfig

Create `.editorconfig` for consistent formatting:

```ini
# .editorconfig
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true

[*.go]
indent_style = tab
indent_size = 4

[*.{yaml,yml}]
indent_style = space
indent_size = 2

[*.{json,js,ts}]
indent_style = space
indent_size = 2

[*.md]
trim_trailing_whitespace = false
```

### golangci-lint Configuration

Create `.golangci.yml`:

```yaml
# .golangci.yml
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  gocyclo:
    min-complexity: 15
  golint:
    min-confidence: 0
  govet:
    check-shadowing: true
  maligned:
    suggest-new: true
  dupl:
    threshold: 100

linters:
  enable:
    - gocyclo
    - golint
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
  disable:
    - maligned
    - prealloc

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gocyclo
        - errcheck
        - dupl
        - gosec
```

## Build Tools Setup

### Make Configuration

Ensure your Makefile supports development workflows:

```makefile
# Makefile
.PHONY: help build test lint clean dev-setup

# Default target
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  test       - Run all tests"
	@echo "  lint       - Run linters"
	@echo "  clean      - Clean build artifacts"
	@echo "  dev-setup  - Set up development environment"

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go mod download
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "Development environment ready!"

# Build
build:
	go build -o bin/openframe .

# Testing
test:
	go test -v ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting
lint:
	golangci-lint run

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
```

### Air Configuration (Live Reload)

Create `.air.toml` for live reloading during development:

```toml
# .air.toml
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/openframe"
  cmd = "go build -o ./tmp/openframe ."
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
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

## Testing Environment

### Test Database Setup (if needed)

```bash
# Set up test environment
export OPENFRAME_TEST_ENV=true
export OPENFRAME_TEST_CLUSTER_PREFIX="test-"

# Create test configuration
mkdir -p ~/.config/openframe/test
cat > ~/.config/openframe/test/config.yaml <<EOF
cluster:
  defaultName: test-cluster
  defaultNodes: 1
  cleanup: true
testing:
  timeout: 5m
  verbose: true
EOF
```

### Integration Test Environment

```bash
# Set up integration testing
export INTEGRATION_TEST=true
export TEST_CLUSTER_NAME="integration-test"

# Create integration test script
cat > scripts/integration-test.sh <<'EOF'
#!/bin/bash
set -e

echo "Setting up integration test environment..."

# Clean up any existing test clusters
k3d cluster delete $TEST_CLUSTER_NAME || true

# Run integration tests
go test -tags=integration ./tests/integration/...

# Clean up
k3d cluster delete $TEST_CLUSTER_NAME || true
EOF

chmod +x scripts/integration-test.sh
```

## Verification

### Environment Check Script

Create a script to verify your development setup:

```bash
#!/bin/bash
# scripts/check-dev-env.sh

echo "🔍 Checking OpenFrame development environment..."
echo "================================================"

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go $(go version | awk '{print $3}')"
else
    echo "❌ Go not installed"
fi

# Check required tools
tools=("git" "make" "docker" "kubectl" "helm" "k3d")
for tool in "${tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool installed"
    else
        echo "❌ $tool missing"
    fi
done

# Check Go tools
go_tools=("goimports" "golangci-lint" "dlv")
for tool in "${go_tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool (Go tool)"
    else
        echo "⚠️  $tool not installed (optional)"
    fi
done

# Check environment variables
if [[ -n "$GOPATH" ]]; then
    echo "✅ GOPATH set to $GOPATH"
else
    echo "⚠️  GOPATH not set (using defaults)"
fi

# Check project structure
if [[ -f "go.mod" ]]; then
    echo "✅ Go module found"
else
    echo "❌ Not in OpenFrame project directory"
fi

echo ""
echo "🎯 Development environment check complete!"
```

### Build and Test Verification

```bash
# Verify the environment works
make dev-setup
make build
make test
make lint

# Test basic functionality
./bin/openframe --help
./bin/openframe cluster --help
```

## IDE-Specific Features

### VS Code Snippets

Create `.vscode/go.code-snippets`:

```json
{
    "Cobra Command": {
        "prefix": "cobra-cmd",
        "body": [
            "func get${1:Command}Cmd() *cobra.Command {",
            "\tcmd := &cobra.Command{",
            "\t\tUse:   \"${2:command}\",",
            "\t\tShort: \"${3:Short description}\",",
            "\t\tLong:  `${4:Long description}`,",
            "\t\tRunE: func(cmd *cobra.Command, args []string) error {",
            "\t\t\treturn ${5:nil}",
            "\t\t},",
            "\t}",
            "\treturn cmd",
            "}"
        ],
        "description": "Create a new Cobra command"
    }
}
```

### GoLand Live Templates

Create live templates for common patterns:
- File → Settings → Editor → Live Templates
- Add template group "OpenFrame"
- Add templates for common functions

## What's Next?

Now that your development environment is set up:

1. **[Local Development Guide](./local-development.md)** - Learn to run OpenFrame locally
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase
3. **[Testing Guide](../testing/overview.md)** - Learn testing practices
4. **[Contributing Guidelines](../contributing/guidelines.md)** - Start contributing

Your development environment is now ready for OpenFrame CLI development! 🎉