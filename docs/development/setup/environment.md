# Development Environment Setup

Set up your development environment for contributing to OpenFrame CLI. This guide covers IDE configuration, required tools, and development-specific settings.

## Go Development Environment

### Go Installation

OpenFrame CLI requires Go 1.21 or later:

**Linux/macOS:**
```bash
# Download and install Go
curl -LO https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin

# Reload shell
source ~/.bashrc
```

**macOS (via Homebrew):**
```bash
brew install go@1.21
```

**Windows:**
Download installer from [https://golang.org/dl/](https://golang.org/dl/)

**Verify Installation:**
```bash
go version  # Should show 1.21.x or later
go env GOPATH
go env GOROOT
```

### Go Environment Configuration

```bash
# Set Go modules on (default in Go 1.16+)
go env -w GO111MODULE=on

# Configure Go proxy for faster downloads
go env -w GOPROXY=https://proxy.golang.org,direct

# Enable checksum database
go env -w GOSUMDB=sum.golang.org
```

## IDE Setup

### Visual Studio Code (Recommended)

**Install VS Code:**
- Download from [https://code.visualstudio.com/](https://code.visualstudio.com/)
- Or via package manager: `brew install --cask visual-studio-code`

**Essential Extensions:**
```bash
# Install via command line
code --install-extension golang.Go
code --install-extension ms-vscode.vscode-go
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension ms-vscode.makefile-tools
code --install-extension redhat.vscode-yaml
```

**Recommended Extensions:**
```bash
# Additional useful extensions
code --install-extension GitHub.copilot                # AI assistance
code --install-extension eamodio.gitlens              # Git integration  
code --install-extension ms-vscode.test-adapter-converter # Testing
code --install-extension humao.rest-client            # API testing
code --install-extension ms-vscode.hexdump           # Binary file viewing
```

**VS Code Settings (`.vscode/settings.json`):**
```json
{
    "go.testTimeout": "60s",
    "go.testFlags": ["-v", "-race"],
    "go.lintTool": "golangci-lint",
    "go.lintOnSave": "package",
    "go.formatTool": "goimports",
    "go.useLanguageServer": true,
    "go.autocompleteUnimportedPackages": true,
    "go.gocodePackageLookupMode": "go",
    "go.gotoSymbol.includeImports": true,
    "go.testExplorer.enable": true,
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.eol": "\\n",
    "kubernetes-tools": {
        "vs-kubernetes.draft-path": "",
        "vs-kubernetes.helm-path": "/usr/local/bin/helm",
        "vs-kubernetes.kubectl-path": "/usr/local/bin/kubectl"
    }
}
```

**VS Code Tasks (`.vscode/tasks.json`):**
```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "go: build",
            "type": "shell", 
            "command": "go build -o bin/openframe main.go",
            "group": {
                "kind": "build",
                "isDefault": true
            }
        },
        {
            "label": "go: test",
            "type": "shell",
            "command": "go test ./...",
            "group": {
                "kind": "test",
                "isDefault": true
            }
        },
        {
            "label": "make: test-integration",
            "type": "shell",
            "command": "make test-integration"
        }
    ]
}
```

### GoLand/IntelliJ IDEA

**Install GoLand:**
- Download from [https://www.jetbrains.com/go/](https://www.jetbrains.com/go/)
- Or install IntelliJ IDEA with Go plugin

**Configuration:**
1. **Go SDK**: Configure Go SDK path (`/usr/local/go`)
2. **GOPATH**: Set GOPATH to `$HOME/go`
3. **Code Style**: Enable `gofmt` on save
4. **Inspections**: Enable Go-specific code inspections
5. **File Watchers**: Set up for automatic formatting

**Useful Plugins:**
- Kubernetes
- Makefile Language
- YAML/Ansible Support
- GitToolBox

### Alternative Editors

**Vim/Neovim:**
```vim
" Install vim-go plugin
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }

" Essential vim-go settings
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
let g:go_highlight_build_constraints = 1
```

**Emacs:**
```elisp
;; Install go-mode and related packages
(use-package go-mode
  :config
  (setq gofmt-command "goimports")
  (add-hook 'before-save-hook 'gofmt-before-save))
```

## Development Tools

### Required Development Tools

```bash
# Install development dependencies
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/golang/mock/mockgen@latest
go install golang.org/x/tools/cmd/godoc@latest
go install github.com/go-delve/delve/cmd/dlv@latest

# Verify installations
goimports -h
golangci-lint --version
mockgen -version
dlv version
```

### Code Quality Tools

**golangci-lint Configuration (`.golangci.yml`):**
```yaml
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
    - maligned
    - prealloc
    - gochecknoglobals

run:
  skip-dirs:
    - vendor
    - testdata
  skip-files:
    - ".*\\.pb\\.go"
    - ".*_mock\\.go"

issues:
  exclude-use-default: false
  exclude:
    - "should have a package comment"
```

### Testing Tools

```bash
# Install testing utilities
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html@latest

# Coverage visualization
go install github.com/boumenot/gocover-cobertura@latest
```

### Build Tools

**Makefile for Development:**
```makefile
.PHONY: dev-deps build test lint fmt vet coverage

# Development dependencies
dev-deps:
	@echo "Installing development dependencies..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang/mock/mockgen@latest

# Build
build:
	go build -o bin/openframe main.go

# Testing
test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Code quality
lint:
	golangci-lint run

fmt:
	goimports -w .

vet:
	go vet ./...

# Development
dev: fmt vet lint test build

clean:
	rm -f bin/openframe coverage.out coverage.html
```

## Docker Development Environment

For containerized development:

**Development Dockerfile:**
```dockerfile
FROM golang:1.21-alpine AS development

RUN apk add --no-cache git make bash curl

# Install development tools
RUN go install golang.org/x/tools/cmd/goimports@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install github.com/golang/mock/mockgen@latest

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Default to development mode
CMD ["make", "dev"]
```

**Docker Compose for Development:**
```yaml
version: '3.8'
services:
  openframe-dev:
    build: 
      context: .
      target: development
    volumes:
      - .:/workspace
      - go-mod-cache:/go/pkg/mod
      - ~/.gitconfig:/etc/gitconfig:ro
    working_dir: /workspace
    command: tail -f /dev/null  # Keep container running

volumes:
  go-mod-cache:
```

## Environment Variables for Development

```bash
# Add to ~/.bashrc, ~/.zshrc, or shell profile

# Go development
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on

# OpenFrame development
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug

# Testing
export OPENFRAME_TEST_TIMEOUT=60s
export K3D_FIX_MOUNTS=1  # For Docker Desktop on macOS

# Docker development
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1
```

## Git Configuration for Development

```bash
# Essential Git configuration
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Useful aliases for OpenFrame development  
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.unstage 'reset HEAD --'
git config --global alias.visual '!gitk'

# Line ending configuration
git config --global core.autocrlf false  # Preserve line endings
git config --global core.eol lf         # Use LF line endings

# Enable commit signing (recommended)
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID
```

**Git Hooks for Development:**
```bash
# Install pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Format code
echo "Formatting code..."
goimports -w .

# Run linter
echo "Running linter..."
golangci-lint run

# Run tests
echo "Running tests..."
go test ./...

echo "Pre-commit checks passed!"
EOF

chmod +x .git/hooks/pre-commit
```

## Debugging Configuration

### VS Code Debugging (`.vscode/launch.json`)

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
            "args": ["bootstrap", "--deployment-mode=oss-tenant", "--verbose"],
            "env": {
                "OPENFRAME_LOG_LEVEL": "debug"
            },
            "cwd": "${workspaceFolder}"
        },
        {
            "name": "Debug Tests",
            "type": "go",
            "request": "launch",
            "mode": "test",
            "program": "${workspaceFolder}/internal/cluster",
            "args": ["-test.v"]
        },
        {
            "name": "Attach to Process",
            "type": "go",
            "request": "attach",
            "mode": "local",
            "processId": "${command:pickProcess}"
        }
    ]
}
```

### Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug main application
dlv debug main.go -- bootstrap --deployment-mode=oss-tenant

# Debug tests
dlv test ./internal/cluster

# Debug with breakpoints
dlv debug main.go
(dlv) break main.main
(dlv) continue
```

## Performance Profiling

```bash
# CPU profiling
go test -cpuprofile cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling  
go test -memprofile mem.prof ./...
go tool pprof mem.prof

# Benchmark profiling
go test -bench=. -cpuprofile cpu.prof -memprofile mem.prof ./...
```

## Development Verification

Run this checklist to verify your environment is properly set up:

```bash
# 1. Go version and environment
go version                    # Should be 1.21+
go env GOPATH                # Should be set
go env GOROOT                # Should be set

# 2. Development tools
goimports -h                 # Code formatting
golangci-lint --version      # Linting
mockgen -version            # Mock generation
dlv version                 # Debugger

# 3. Build and test
make build                  # Should compile successfully
make test                   # Should pass all tests
make lint                   # Should pass linting

# 4. IDE integration
code --version              # VS Code (if using)
# OR
goland.sh                   # GoLand (if using)

# 5. Docker development (optional)
docker --version            # Docker for containers
docker-compose --version    # Container orchestration
```

## Next Steps

With your development environment configured:

1. **[Local Development](local-development.md)** - Clone and build the project
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase
3. **[Testing Guide](../testing/overview.md)** - Run and write tests
4. **[Contributing Guidelines](../contributing/guidelines.md)** - Submit your first PR

---

**Environment Ready!** Your development setup is now complete. You can efficiently develop, test, and debug OpenFrame CLI with all the necessary tools and configurations in place.