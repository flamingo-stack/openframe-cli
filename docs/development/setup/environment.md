# Development Environment Setup

This guide helps you configure your development environment for contributing to OpenFrame CLI. We'll set up your IDE, development tools, and workspace for an optimal development experience.

## 🎯 Overview

Setting up a proper development environment will help you:
- Write code efficiently with proper syntax highlighting and debugging
- Maintain code quality with linting and formatting
- Run and test OpenFrame CLI locally
- Contribute to the project following established conventions

## 📋 Prerequisites

Before setting up your development environment, ensure you have completed:
- ✅ [OpenFrame Prerequisites](../../getting-started/prerequisites.md)
- ✅ [Quick Start Guide](../../getting-started/quick-start.md) (for understanding the tool)

### Additional Development Requirements

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| **Go** | 1.21+ | Primary language | [Install Go](https://golang.org/dl/) |
| **Make** | Latest | Build automation | Platform package manager |
| **golangci-lint** | Latest | Code linting | [Install golangci-lint](https://golangci-lint.run/usage/install/) |

## 🛠️ IDE Setup

### Visual Studio Code (Recommended)

VS Code provides excellent Go support and integrates well with OpenFrame CLI development.

#### 1. Install VS Code Extensions

```bash
# Install via command line
code --install-extension golang.Go
code --install-extension ms-vscode.vscode-json
code --install-extension redhat.vscode-yaml  
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension GitLab.gitlab-workflow
```

**Essential Extensions:**
| Extension | Purpose | Benefits |
|-----------|---------|----------|
| **Go** | Go language support | Syntax highlighting, debugging, testing |
| **YAML** | YAML file support | Kubernetes manifest editing |
| **Kubernetes** | K8s integration | Cluster management within IDE |
| **GitLab Workflow** | Git integration | Enhanced Git workflows |

#### 2. Configure VS Code Settings

Create `.vscode/settings.json` in your project:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.useLanguageServer": true,
  "go.testFlags": ["-v", "-race"],
  "go.testTimeout": "300s",
  "go.coverOnSave": true,
  "go.coverOnSaveMode": "package",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.exclude": {
    "**/vendor": true,
    "**/node_modules": true,
    "**/.git": true,
    "**/dist": true,
    "**/build": true
  },
  "yaml.schemas": {
    "https://json.schemastore.org/kustomization": "kustomization.yaml",
    "https://raw.githubusercontent.com/argoproj/argo-cd/master/manifests/crds/application-crd.yaml": "argocd/application.yaml"
  }
}
```

#### 3. Configure Debugging

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
      "args": ["bootstrap", "--verbose", "--dry-run"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    },
    {
      "name": "Debug Cluster Create",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "create", "test-cluster", "--verbose"],
      "env": {
        "OPENFRAME_LOG_LEVEL": "debug"
      }
    }
  ]
}
```

#### 4. Set Up Tasks

Create `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Go: Build",
      "type": "shell",
      "command": "go",
      "args": ["build", "-o", "openframe", "main.go"],
      "group": "build",
      "presentation": {
        "echo": true,
        "reveal": "silent",
        "focus": false,
        "panel": "shared"
      }
    },
    {
      "label": "Go: Test All",
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
      "label": "Go: Lint",
      "type": "shell",
      "command": "golangci-lint",
      "args": ["run"],
      "group": "test",
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

### GoLand/IntelliJ IDEA

For JetBrains IDE users:

#### 1. Install Go Plugin
- Go to **File > Settings > Plugins**
- Search for and install **Go** plugin
- Restart IDE

#### 2. Configure Go Settings
- **File > Settings > Go > GOROOT**: Point to your Go installation
- **File > Settings > Go > GOPATH**: Set to your workspace
- **File > Settings > Go > Build Tags**: Add `integration` for integration tests

#### 3. Set Up Run Configurations

Create run configurations for:
- **Main Application**: `main.go` with various arguments
- **Tests**: Package or directory-level test execution
- **Benchmarks**: Performance testing configurations

## 🔧 Development Tools Setup

### Go Tools Installation

```bash
# Install essential Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Install testing tools
go install github.com/onsi/ginkgo/v2/ginkgo@latest
go install gotest.tools/gotestsum@latest

# Verify installation
goimports -h
golangci-lint --version
dlv version
```

### Git Hooks Setup

Set up pre-commit hooks to ensure code quality:

```bash
# Create git hooks directory
mkdir -p .git/hooks

# Create pre-commit hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Format code
goimports -w .

# Run lints
golangci-lint run

# Run tests
go test -short ./...

echo "Pre-commit checks passed!"
EOF

# Make executable
chmod +x .git/hooks/pre-commit
```

### Makefile Setup

Create a `Makefile` for common development tasks:

```makefile
# Build variables
BINARY_NAME=openframe
VERSION=$(shell git describe --tags --always)
COMMIT=$(shell git rev-parse --short HEAD)
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.Date=$(DATE)"

.PHONY: all build clean test coverage lint fmt help

all: lint test build

## Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) main.go

## Run tests
test:
	$(GOTEST) -v -race ./...

## Run tests with coverage
coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Run linter
lint:
	golangci-lint run

## Format code
fmt:
	goimports -w .
	gofmt -w .

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## Download dependencies
deps:
	$(GOMOD) tidy
	$(GOMOD) download

## Install development tools
install-tools:
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Display help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
```

## ⚙️ Environment Configuration

### Environment Variables

Set up your development environment variables:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)

# OpenFrame Development
export OPENFRAME_LOG_LEVEL=debug
export OPENFRAME_CONFIG_DIR=~/.openframe-dev
export OPENFRAME_DEV_MODE=true

# Go Development
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org

# Development Tools
export PATH=$PATH:$(go env GOPATH)/bin

# Docker Development
export DOCKER_BUILDKIT=1

# Kubernetes Development
export KUBECONFIG=~/.kube/config
```

### Development Configuration

Create a development-specific configuration file:

```bash
# Create config directory
mkdir -p ~/.openframe-dev

# Create development config
cat > ~/.openframe-dev/config.yaml << EOF
log:
  level: debug
  format: text

development:
  enabled: true
  cluster:
    prefix: "dev-"
    auto_cleanup: true
  
testing:
  integration_enabled: true
  mock_external_calls: false
EOF
```

## 🧪 Testing Environment

### Test Configuration

Set up your testing environment:

```bash
# Create test config
cat > test.env << EOF
OPENFRAME_LOG_LEVEL=debug
OPENFRAME_TEST_MODE=true
KUBECONFIG=/tmp/test-kubeconfig
DOCKER_HOST=unix:///var/run/docker.sock
EOF
```

### Integration Test Setup

For running integration tests:

```bash
# Install test dependencies
go install gotest.tools/gotestsum@latest

# Set up test cluster (optional, for integration tests)
k3d cluster create openframe-test --agents 2

# Export kubeconfig for tests
k3d kubeconfig get openframe-test > /tmp/test-kubeconfig
```

## 📊 Code Quality Tools

### golangci-lint Configuration

Create `.golangci.yml`:

```yaml
run:
  timeout: 5m
  issues-exit-code: 1

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - golint
    - ineffassign
    - misspell
    - errcheck
    - gosimple
    - staticcheck
    - unused
    - typecheck

linters-settings:
  goimports:
    local-prefixes: github.com/flamingo-stack/openframe-cli

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - errcheck
```

### Code Coverage

Set up coverage reporting:

```bash
# Run tests with coverage
make coverage

# View coverage in browser
open coverage.html

# Check coverage threshold (optional)
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
```

## 🚀 Development Workflow

### Daily Development Routine

1. **Start Development Session**:
   ```bash
   # Update dependencies
   make deps
   
   # Run linter and tests
   make lint test
   
   # Start development
   code .  # or your preferred IDE
   ```

2. **Development Cycle**:
   ```bash
   # Make changes
   # Run tests frequently
   make test
   
   # Format and lint before commit
   make fmt lint
   
   # Build and test locally
   make build
   ./openframe --help
   ```

3. **Before Committing**:
   ```bash
   # Final checks
   make all
   
   # Commit changes
   git add .
   git commit -m "feat: add new feature"
   ```

## 🔍 Debugging Tips

### Local Debugging

```bash
# Build with debug info
go build -gcflags="all=-N -l" -o openframe-debug main.go

# Run with debugger
dlv exec ./openframe-debug -- cluster create test --verbose

# Or use IDE debugging with configurations above
```

### Troubleshooting Common Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| **Import errors** | `cannot find package` | Run `go mod tidy` |
| **Lint failures** | Code quality issues | Run `make fmt lint` |
| **Test failures** | Unit tests failing | Check test environment setup |
| **Build errors** | Compilation fails | Verify Go version and dependencies |

## 📚 Additional Resources

### Documentation
- **[Go Documentation](https://golang.org/doc/)** - Language reference
- **[VS Code Go](https://marketplace.visualstudio.com/items?itemName=golang.Go)** - Extension documentation
- **[golangci-lint](https://golangci-lint.run/)** - Linter configuration

### Development Tools
- **[Delve Debugger](https://github.com/go-delve/delve)** - Go debugging
- **[GoLand IDE](https://www.jetbrains.com/go/)** - Professional Go IDE
- **[Go Tools](https://golang.org/x/tools)** - Additional Go tooling

---

*Environment setup complete? Let's move on to [local development](local-development.md) to start running OpenFrame CLI from source!*