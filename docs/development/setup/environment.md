# Development Environment Setup

Set up your development environment for optimal productivity when working on OpenFrame CLI. This guide covers IDE configuration, essential tools, editor extensions, and development workflow optimization.

## IDE Recommendations and Setup

### Option 1: Visual Studio Code (Recommended)

VS Code provides excellent Go support and integrates well with Kubernetes development.

#### Installation
```bash
# macOS with Homebrew
brew install --cask visual-studio-code

# Linux (Ubuntu/Debian)
wget -qO- https://packages.microsoft.com/keys/microsoft.asc | gpg --dearmor > packages.microsoft.gpg
sudo install -o root -g root -m 644 packages.microsoft.gpg /etc/apt/trusted.gpg.d/
sudo sh -c 'echo "deb [arch=amd64,arm64,armhf signed-by=/etc/apt/trusted.gpg.d/packages.microsoft.gpg] https://packages.microsoft.com/repos/code stable main" > /etc/apt/sources.list.d/vscode.list'
sudo apt update && sudo apt install code
```

#### Essential Extensions

Install these extensions for optimal Go and Kubernetes development:

| Extension | Purpose | Extension ID |
|-----------|---------|--------------|
| **Go** | Go language support | `golang.Go` |
| **Kubernetes** | K8s YAML editing and cluster management | `ms-kubernetes-tools.vscode-kubernetes-tools` |
| **YAML** | YAML syntax and validation | `redhat.vscode-yaml` |
| **Docker** | Dockerfile and container management | `ms-azuretools.vscode-docker` |
| **GitLens** | Advanced Git integration | `eamodio.gitlens` |
| **Error Lens** | Inline error display | `usernamehw.errorlens` |
| **Thunder Client** | API testing | `rangav.vscode-thunder-client` |

```bash
# Install all recommended extensions at once
code --install-extension golang.Go
code --install-extension ms-kubernetes-tools.vscode-kubernetes-tools  
code --install-extension redhat.vscode-yaml
code --install-extension ms-azuretools.vscode-docker
code --install-extension eamodio.gitlens
code --install-extension usernamehw.errorlens
code --install-extension rangav.vscode-thunder-client
```

#### VS Code Configuration

Create `.vscode/settings.json` in your project root:

```json
{
  "go.formatTool": "gofmt",
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.testFlags": ["-v"],
  "go.testTimeout": "30s",
  "go.coverOnSave": true,
  "go.coverageDecorator": {
    "type": "gutter",
    "coveredHighlightColor": "rgba(64,128,64,0.5)",
    "uncoveredHighlightColor": "rgba(128,64,64,0.5)"
  },
  "yaml.schemas": {
    "https://json.schemastore.org/github-workflow.json": ".github/workflows/*.yml",
    "kubernetes": ["*.k8s.yaml", "k8s/*.yaml"]
  },
  "files.associations": {
    "*.yaml": "yaml",
    "*.yml": "yaml"
  },
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

#### Debug Configuration

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
      "program": "${workspaceFolder}",
      "args": ["--help"],
      "cwd": "${workspaceFolder}",
      "env": {
        "OPENFRAME_DEBUG": "true"
      }
    },
    {
      "name": "Debug Bootstrap Command",
      "type": "go",
      "request": "launch", 
      "mode": "debug",
      "program": "${workspaceFolder}",
      "args": ["bootstrap", "--help"],
      "cwd": "${workspaceFolder}"
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/cluster"
    }
  ]
}
```

### Option 2: GoLand

JetBrains GoLand provides advanced refactoring and debugging capabilities.

#### Key Features for OpenFrame Development
- **Advanced Go support**: Refactoring, debugging, profiling
- **Kubernetes integration**: YAML editing, cluster browser
- **Database tools**: For testing database integrations
- **HTTP client**: API testing and development
- **Git integration**: Visual diff, merge conflicts

#### Recommended Plugins
- **Kubernetes**: Kubernetes support and YAML validation
- **Docker**: Container management and Dockerfile editing  
- **Makefile Language**: Syntax support for Makefiles
- **Mermaid**: Diagram preview (for documentation)

#### Configuration
```bash
# Set GOPATH and GOROOT in GoLand settings
# File -> Settings -> Go -> GOPATH
# File -> Settings -> Go -> GOROOT
```

### Option 3: Neovim/Vim

For terminal-based development with powerful Go support.

#### Essential Plugins
```lua
-- Using lazy.nvim plugin manager
{
  'fatih/vim-go',
  'neovim/nvim-lspconfig',
  'ray-x/go.nvim',
  'nvim-treesitter/nvim-treesitter',
  'lewis6991/gitsigns.nvim',
  'kubernetes/yaml'
}
```

#### Go Configuration  
```lua
-- In your init.lua
require('go').setup({
  gofmt = 'gofumpt',
  max_line_len = 100,
  tag_transform = false,
  test_dir = '',
  comment_placeholder = '',
  lsp_cfg = true,
  lsp_gofumpt = true,
  lsp_on_attach = true,
  dap_debug = true,
})
```

## Required Development Tools

### Core Development Tools

| Tool | Version | Installation | Purpose |
|------|---------|-------------|---------|
| **Go** | 1.23+ | [golang.org](https://golang.org/dl/) | Primary language |
| **Git** | 2.30+ | [git-scm.com](https://git-scm.com/) | Version control |
| **Make** | 3.81+ | System package manager | Build automation |
| **Docker** | 20.10+ | [docker.com](https://docker.com/) | Container testing |

#### Go Installation and Configuration
```bash
# Download and install Go 1.23+
wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz

# Configure environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
source ~/.bashrc

# Verify installation
go version
go env GOPATH
go env GOROOT
```

### Go Development Tools

Install essential Go tools for development:

```bash
# Code formatting and imports
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest

# Linting and static analysis  
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install honnef.co/go/tools/cmd/staticcheck@latest

# Testing and coverage
go install github.com/rakyll/gotest@latest
go install github.com/axw/gocov/gocov@latest
go install github.com/matm/gocov-html@latest

# Documentation
go install golang.org/x/tools/cmd/godoc@latest

# Debugging
go install github.com/go-delve/delve/cmd/dlv@latest

# Dependency management
go install github.com/psampaz/go-mod-outdated@latest
```

### Kubernetes Development Tools

| Tool | Installation | Purpose |
|------|-------------|---------|
| **kubectl** | `curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"` | Kubernetes CLI |
| **k3d** | `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh \| bash` | Lightweight Kubernetes |
| **helm** | `curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \| bash` | Package manager |
| **skaffold** | `curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && sudo install skaffold /usr/local/bin/` | Development workflow |

### Quality and Testing Tools

```bash
# Install testing and quality tools
go install gotest.tools/gotestsum@latest     # Better test output
go install github.com/vektra/mockery/v2@latest  # Mock generation
go install github.com/securecodewarrior/sast-scan@latest  # Security scanning
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest   # Complexity analysis
```

## Environment Variables for Development

Set up your development environment variables:

### Essential Variables
```bash
# Add to ~/.bashrc or ~/.zshrc
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on
export CGO_ENABLED=0

# OpenFrame CLI specific
export OPENFRAME_DEBUG=true
export OPENFRAME_CONFIG_DIR=$HOME/.openframe-dev
export OPENFRAME_TEST_MODE=true

# Docker configuration
export DOCKER_BUILDKIT=1
export COMPOSE_DOCKER_CLI_BUILD=1

# Kubernetes configuration
export KUBECONFIG=$HOME/.kube/config
export KUBE_EDITOR="code --wait"

# Development convenience
export EDITOR="code --wait"
export GIT_EDITOR="code --wait"
```

### Development-specific Configuration
```bash
# Create development config
mkdir -p $HOME/.openframe-dev
echo 'debug: true' > $HOME/.openframe-dev/config.yaml
echo 'log_level: debug' >> $HOME/.openframe-dev/config.yaml
echo 'test_mode: true' >> $HOME/.openframe-dev/config.yaml
```

## Editor Extensions and Plugins

### VS Code Extensions (Detailed)

#### Go Extension Configuration
```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.alternateTools": {
    "gofmt": "gofumpt",
    "goimports": "goimports"
  },
  "go.lintFlags": [
    "--config=.golangci.yml"
  ],
  "go.buildFlags": [
    "-v"
  ],
  "go.testEnvVars": {
    "OPENFRAME_TEST_MODE": "true"
  }
}
```

#### Kubernetes Extension Configuration
```json
{
  "vs-kubernetes": {
    "vs-kubernetes.crd-code-completion": "enabled",
    "vs-kubernetes.knative-disable-linting": false,
    "vs-kubernetes.disable-linting": false
  }
}
```

### Shell Integration

#### Bash/Zsh Completions
```bash
# Add command completions
openframe completion bash > /etc/bash_completion.d/openframe
kubectl completion bash > /etc/bash_completion.d/kubectl
helm completion bash > /etc/bash_completion.d/helm

# Or for zsh
openframe completion zsh > ~/.oh-my-zsh/completions/_openframe
kubectl completion zsh > ~/.oh-my-zsh/completions/_kubectl
helm completion zsh > ~/.oh-my-zsh/completions/_helm
```

#### Useful Aliases
```bash
# Add to shell profile
alias of='openframe'
alias k='kubectl'
alias h='helm'
alias d='docker'

# Development specific
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofdev='openframe dev'

# Testing aliases
alias got='go test'
alias gotv='go test -v'
alias gotc='go test -cover'

# Build aliases
alias build='make build'
alias test='make test'
alias lint='make lint'
alias clean='make clean'
```

## Development Workflow Optimization

### Git Configuration
```bash
# Configure Git for development
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
git config --global core.editor "code --wait"
git config --global merge.tool "code"
git config --global diff.tool "code"

# Useful Git aliases
git config --global alias.co checkout
git config --global alias.br branch  
git config --global alias.ci commit
git config --global alias.st status
git config --global alias.lg "log --oneline --graph --decorate"
```

### Pre-commit Hooks
```bash
# Install pre-commit hooks
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Format code
gofmt -s -w .
goimports -w .

# Run tests  
go test ./...

# Lint code
golangci-lint run
EOF

chmod +x .git/hooks/pre-commit
```

### Build Optimization
```bash
# Create development Makefile targets
cat >> Makefile << 'EOF'
.PHONY: dev-setup
dev-setup: ## Set up development environment
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

.PHONY: dev-test  
dev-test: ## Run tests in development mode
	OPENFRAME_TEST_MODE=true go test -v ./...

.PHONY: dev-build
dev-build: ## Build with development flags
	go build -ldflags="-X main.version=dev-$(shell git rev-parse --short HEAD)" -o openframe-dev

.PHONY: dev-clean
dev-clean: ## Clean development artifacts
	rm -f openframe-dev
	docker system prune -f
	k3d cluster delete --all
EOF
```

## Verification

Verify your development environment is properly configured:

### Tool Verification
```bash
# Check all required tools
echo "=== Go ===" && go version
echo "=== Git ===" && git --version  
echo "=== Docker ===" && docker --version
echo "=== kubectl ===" && kubectl version --client
echo "=== k3d ===" && k3d version
echo "=== Helm ===" && helm version --short

# Check Go tools
echo "=== Go Tools ==="
which gofmt goimports golangci-lint staticcheck dlv
```

### Build Test
```bash
# Test basic build
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
make dev-setup
make dev-build
./openframe-dev --help
```

### IDE Test
```bash
# Open project in your IDE
code .  # VS Code
# or
idea . # GoLand
```

You should see:
- ✅ Go syntax highlighting working
- ✅ Auto-completion functioning  
- ✅ Error detection and linting active
- ✅ Debug configuration available

## Troubleshooting

### Common Issues

#### Go Module Issues
```bash
# Clean Go modules
go clean -modcache
go mod download
go mod verify
```

#### IDE Issues
```bash
# Reset VS Code Go tools
# Command Palette -> Go: Install/Update Tools
# Select all tools and install
```

#### Permission Issues
```bash
# Fix Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

## Next Steps

With your environment set up:

1. **[Clone and Build](./local-development.md)** - Get the code and run your first build
2. **[Architecture Overview](../architecture/overview.md)** - Understand the codebase structure  
3. **[Testing Guide](../testing/overview.md)** - Learn the testing strategy
4. **[Contributing](../contributing/guidelines.md)** - Make your first contribution

Your development environment is now optimized for OpenFrame CLI development! 🚀