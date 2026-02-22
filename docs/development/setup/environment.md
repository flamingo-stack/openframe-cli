# Development Environment Setup

This guide walks you through setting up a complete development environment for OpenFrame CLI, including IDEs, tools, extensions, and configuration for maximum productivity.

## IDE and Editor Setup

### Visual Studio Code (Recommended)

VS Code provides excellent Go support and Kubernetes tooling:

#### Installation
```bash
# macOS
brew install --cask visual-studio-code

# Ubuntu/Debian
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update
sudo apt install code
```

#### Essential Extensions

Install these extensions for Go and Kubernetes development:

```bash
# Core Go development
code --install-extension golang.Go
code --install-extension golang.Go-nightly

# Kubernetes and YAML
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools
code --install-extension redhat.vscode-yaml

# Git and GitHub integration
code --install-extension GitHub.vscode-pull-request-github
code --install-extension eamodio.gitlens

# Code quality and testing
code --install-extension ms-vscode.test-adapter-converter
code --install-extension hbenl.vscode-test-explorer
code --install-extension SonarSource.sonarlint-vscode

# Docker and containers
code --install-extension ms-azuretools.vscode-docker

# Markdown and documentation
code --install-extension yzhang.markdown-all-in-one
code --install-extension bierner.markdown-mermaid

# Productivity
code --install-extension vscodevim.vim  # Optional: Vim keybindings
code --install-extension ms-vscode.vscode-json
```

#### VS Code Configuration

Create `.vscode/settings.json` in your workspace:

```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.gopath": "",
  "go.goroot": "",
  "go.lintTool": "golangci-lint",
  "go.lintFlags": [
    "--fast"
  ],
  "go.vetOnSave": "package",
  "go.formatTool": "goimports",
  "go.buildOnSave": "package",
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s",
  "go.coverOnSave": false,
  "go.coverOnSingleTest": true,
  "go.coverOnSingleTestFile": true,
  "go.coverOnTestPackage": true,
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  },
  "files.eol": "\n",
  "yaml.schemas": {
    "https://json.schemastore.org/kustomization": [
      "kustomization.yaml",
      "kustomization.yml"
    ],
    "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/crds/application-crd.yaml": [
      "**/argocd-apps/*.yaml",
      "**/argocd-apps/*.yml"
    ]
  },
  "kubernetes.namespace": "default",
  "kubernetes.defaultLogLevel": "info"
}
```

Create `.vscode/launch.json` for debugging:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Launch OpenFrame CLI",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["--help"],
      "env": {
        "GO_ENV": "development"
      }
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["bootstrap", "--verbose", "--non-interactive"],
      "env": {
        "GO_ENV": "development"
      }
    },
    {
      "name": "Debug Cluster Status",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "args": ["cluster", "status"],
      "env": {
        "GO_ENV": "development"
      }
    }
  ]
}
```

### GoLand/IntelliJ IDEA

For JetBrains IDEs:

#### Installation
```bash
# macOS
brew install --cask goland

# Or download from https://www.jetbrains.com/go/
```

#### Configuration
1. **Go Settings**: File → Settings → Go → GOROOT and GOPATH
2. **Code Style**: File → Settings → Editor → Code Style → Go
3. **Live Templates**: File → Settings → Editor → Live Templates
4. **Kubernetes Plugin**: File → Settings → Plugins → Install "Kubernetes"

### Vim/Neovim

For terminal-based editing:

#### Installation
```bash
# Install vim-go plugin manager
curl -fLo ~/.vim/autoload/plug.vim --create-dirs \
    https://raw.githubusercontent.com/junegunn/vim-plug/master/plug.vim
```

Add to `~/.vimrc`:
```vim
call plug#begin('~/.vim/plugged')
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }
Plug 'neoclide/coc.nvim', {'branch': 'release'}
Plug 'preservim/nerdtree'
Plug 'airblade/vim-gitgutter'
call plug#end()

" Go-specific settings
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_highlight_types = 1
let g:go_highlight_fields = 1
let g:go_highlight_functions = 1
let g:go_highlight_function_calls = 1
```

## Development Tools

### Go Tools Installation

Install essential Go development tools:

```bash
# Core Go tools
go install golang.org/x/tools/cmd/goimports@latest
go install golang.org/x/tools/cmd/godoc@latest
go install golang.org/x/tools/cmd/gofmt@latest

# Linting and code quality
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/kisielk/errcheck@latest

# Testing tools
go install github.com/rakyll/gotest@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/AlekSi/gocov-xml@latest

# Documentation tools
go install golang.org/x/tools/cmd/godoc@latest

# Binary management
go install github.com/cosmtrek/air@latest  # Hot reload
go install github.com/goreleaser/goreleaser@latest  # Release management
```

### Configure golangci-lint

Create `.golangci.yml` in the project root:

```yaml
run:
  timeout: 5m
  issues-exit-code: 1
  tests: true

linters-settings:
  golint:
    min-confidence: 0
  gocyclo:
    min-complexity: 15
  goimports:
    local-prefixes: github.com/flamingo-stack/openframe-cli
  govet:
    check-shadowing: true
  errcheck:
    check-type-assertions: true
    check-blank: true
  unused:
    check-exported: false

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - goconst
    - gocritic
    - gofmt
    - goimports
    - golint
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - rowserrcheck
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gosec
        - dupl
  exclude-use-default: false
  exclude:
    - "Error return value of .((os\.)?std(out|err)\..*|.*Close|.*Flush|os\.Remove(All)?|.*print.*|os\.(Un)?Setenv). is not checked"
```

### Shell Configuration

Add helpful aliases and functions to your shell profile:

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.fish
# OpenFrame development aliases
alias of="go run main.go"
alias oft="go test ./..."
alias ofb="go build -o openframe main.go"
alias ofl="golangci-lint run"

# Go development aliases
alias gob="go build"
alias got="go test"
alias gor="go run"
alias gof="go fmt ./..."
alias goi="goimports -w ."
alias gol="golangci-lint run"

# Kubernetes aliases for development
alias k="kubectl"
alias kgp="kubectl get pods"
alias kgs="kubectl get svc"
alias kgd="kubectl get deployment"
alias kdp="kubectl describe pod"
alias kl="kubectl logs"
alias kex="kubectl exec -it"

# OpenFrame specific kubectl contexts
alias kof="kubectl config use-context k3d-openframe-local"
alias kctx="kubectl config current-context"
alias kns="kubectl config set-context --current --namespace"

# Docker aliases
alias d="docker"
alias dc="docker compose"
alias dps="docker ps"
alias di="docker images"
alias dl="docker logs"

# Git aliases for development workflow
alias gs="git status"
alias ga="git add"
alias gc="git commit"
alias gp="git push"
alias gl="git pull"
alias gb="git branch"
alias gco="git checkout"
alias gd="git diff"
```

## Environment Variables

Set up development environment variables:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)

# Go development
export GOPATH="$HOME/go"
export GOROOT="$(go env GOROOT)"
export PATH="$GOPATH/bin:$GOROOT/bin:$PATH"

# OpenFrame CLI development
export OPENFRAME_LOG_LEVEL="debug"
export OPENFRAME_CONFIG_DIR="$HOME/.openframe-dev"
export OPENFRAME_DEV_MODE="true"

# Kubernetes development
export KUBECONFIG="$HOME/.kube/config"
export KUBECTL_EXTERNAL_DIFF="diff -u"

# Docker development
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Development tools
export EDITOR="code"  # or vim, nano, etc.
export BROWSER="chrome"  # or firefox, safari, etc.

# Testing
export GO_TEST_TIMEOUT="30s"
export COVERAGE_OUTPUT="coverage.out"
```

## Kubernetes Development Setup

### Configure kubectl contexts

```bash
# Create separate contexts for development
kubectl config set-context openframe-dev \
  --cluster=k3d-openframe-local \
  --user=admin@k3d-openframe-local

kubectl config set-context openframe-test \
  --cluster=k3d-openframe-test \
  --user=admin@k3d-openframe-test

# Use development context by default
kubectl config use-context openframe-dev
```

### Install Kubernetes development tools

```bash
# Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# K3D
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Stern for log streaming
brew install stern  # macOS
# or download from https://github.com/stern/stern/releases

# kubectx and kubens for context switching
brew install kubectx  # macOS
# or download from https://github.com/ahmetb/kubectx/releases

# K9s for cluster management
brew install k9s  # macOS
# or download from https://github.com/derailed/k9s/releases
```

## Testing Environment Setup

### Configure test databases and services

```bash
# Create test namespace
kubectl create namespace openframe-test

# Install test dependencies
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install test-redis bitnami/redis \
  --namespace openframe-test \
  --set auth.enabled=false

# Install test monitoring
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.1/deploy/static/provider/kind/deploy.yaml
```

### Test configuration

Create `test.env` for test environment variables:

```bash
# Test environment configuration
export OPENFRAME_TEST_MODE=true
export OPENFRAME_TEST_CLUSTER="k3d-openframe-test"
export OPENFRAME_TEST_NAMESPACE="openframe-test"
export OPENFRAME_TEST_TIMEOUT="60s"

# Test database connections
export TEST_REDIS_URL="redis://localhost:6379"
export TEST_DATABASE_URL="postgres://localhost:5432/openframe_test"

# Integration test settings
export INTEGRATION_TEST_ENABLED=true
export E2E_TEST_ENABLED=false  # Disable by default
```

## Performance and Debugging Tools

### Install profiling tools

```bash
# Go profiling tools
go install github.com/google/pprof@latest
go install github.com/uber/go-torch@latest

# Memory and performance analysis
go install github.com/pkg/profile@latest
```

### Configure debugging

Add debug configuration to your project:

```go
// debug/profile.go
// +build debug

package debug

import (
    "github.com/pkg/profile"
    "os"
)

func init() {
    if os.Getenv("CPUPROFILE") != "" {
        defer profile.Start().Stop()
    }
    if os.Getenv("MEMPROFILE") != "" {
        defer profile.Start(profile.MemProfile).Stop()
    }
}
```

## Pre-commit Hooks

Set up git hooks for code quality:

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

  - repo: https://github.com/dnephin/pre-commit-golang
    rev: v0.5.1
    hooks:
      - id: go-fmt
      - id: go-imports
      - id: go-vet-mod
      - id: go-unit-tests-mod
      - id: golangci-lint-mod
EOF

# Install the hooks
pre-commit install
```

Or create a simple git hook manually:

```bash
# Create .git/hooks/pre-commit
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for OpenFrame CLI

set -e

echo "Running pre-commit checks..."

# Format code
echo "Running gofmt..."
gofmt -l -w .

# Organize imports
echo "Running goimports..."
goimports -l -w .

# Run linter
echo "Running golangci-lint..."
golangci-lint run

# Run tests
echo "Running tests..."
go test -short ./...

echo "Pre-commit checks passed!"
EOF

chmod +x .git/hooks/pre-commit
```

## Troubleshooting

### Common Go Issues

#### GOPATH/GOROOT Problems
```bash
# Check Go environment
go env

# Reset Go environment
go env -w GOPATH=""
go env -w GOROOT=""
```

#### Module Issues
```bash
# Clean module cache
go clean -modcache

# Reinstall dependencies
go mod tidy
go mod download
```

### Kubernetes Issues

#### Context Not Found
```bash
# List available contexts
kubectl config get-contexts

# Create new context
kubectl config set-context openframe-dev \
  --cluster=k3d-openframe-local \
  --user=admin@k3d-openframe-local
```

#### Tools Not Found
```bash
# Verify PATH includes Go bin directory
echo $PATH | grep go

# Add Go bin to PATH
export PATH="$GOPATH/bin:$PATH"
```

## Next Steps

Your development environment is now ready! Continue with:

1. **[Local Development Guide](local-development.md)** - Clone and run OpenFrame CLI locally
2. **[Architecture Overview](../architecture/README.md)** - Understand the system design
3. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn the development workflow

## Additional Resources

- **Go Documentation**: https://golang.org/doc/
- **Kubernetes Documentation**: https://kubernetes.io/docs/
- **Cobra CLI Documentation**: https://cobra.dev/
- **VS Code Go Extension**: https://marketplace.visualstudio.com/items?itemName=golang.Go
- **OpenMSP Slack**: [Join the community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)