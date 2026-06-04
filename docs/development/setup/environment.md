# Development Environment Setup

This guide covers setting up your local machine for OpenFrame CLI development, including IDE configuration, Go toolchain setup, and recommended extensions.

---

## Go Toolchain

OpenFrame CLI requires **Go 1.21 or newer**.

```bash
# Verify your Go version
go version
# Expected: go version go1.21.x or higher

# Install Go (Linux)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Go (macOS)
brew install go

# Install Go (Windows WSL2)
# Run the Linux instructions above inside your WSL2 terminal
```

Ensure your `$GOPATH` and `$GOBIN` are configured:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export GOPATH=$HOME/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

---

## Recommended IDE: VS Code

[Visual Studio Code](https://code.visualstudio.com/) with the Go extension is the recommended editor for OpenFrame CLI development.

### Required Extensions

| Extension | ID | Purpose |
|-----------|-----|---------|
| **Go** | `golang.go` | Go language support, IntelliSense, debugging |
| **GitLens** | `eamodio.gitlens` | Enhanced Git integration |
| **YAML** | `redhat.vscode-yaml` | Helm values YAML editing |

Install all at once:

```bash
code --install-extension golang.go
code --install-extension eamodio.gitlens
code --install-extension redhat.vscode-yaml
```

### Recommended VS Code Settings

Create or update `.vscode/settings.json` in your project:

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testOnSave": false,
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "always"
    }
  }
}
```

---

## Alternative IDE: GoLand

[GoLand](https://www.jetbrains.com/go/) by JetBrains is a full-featured Go IDE with built-in debugging, refactoring, and Kubernetes support.

### Recommended GoLand Plugins

| Plugin | Purpose |
|--------|---------|
| **Kubernetes** | Kubernetes manifest support |
| **Go** (built-in) | Full Go language tooling |

---

## Code Quality Tools

Install the following tools for linting and formatting:

```bash
# goimports — import organizer (used by the formatter)
go install golang.org/x/tools/cmd/goimports@latest

# golangci-lint — multi-linter runner
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOBIN) latest

# Verify installations
goimports --version
golangci-lint --version
```

---

## Environment Variables for Development

The following environment variables are commonly used during development:

| Variable | Description | Example |
|----------|-------------|---------|
| `GOPATH` | Go workspace directory | `$HOME/go` |
| `GOBIN` | Go binaries directory | `$GOPATH/bin` |
| `KUBECONFIG` | Kubeconfig file path | `$HOME/.kube/config` |
| `CI` | Set to any value to enable CI/non-interactive mode | `true` |

---

## Shell Completions (Optional)

OpenFrame CLI supports shell completions via Cobra. Add them to speed up development testing:

```bash
# Bash completions
openframe completion bash > /etc/bash_completion.d/openframe

# Zsh completions
openframe completion zsh > "${fpath[1]}/_openframe"

# Fish completions
openframe completion fish > ~/.config/fish/completions/openframe.fish
```

---

## Kubernetes Development Tools

For working on the `dev` intercept and scaffold features, install these additional tools:

```bash
# Telepresence v2 (traffic intercept)
curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o telepresence
chmod +x telepresence && sudo mv telepresence /usr/local/bin/

# Skaffold (hot-reload dev sessions)
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
chmod +x skaffold && sudo mv skaffold /usr/local/bin/

# jq (JSON processing for intercept scripts)
sudo apt-get install jq  # Debian/Ubuntu
brew install jq           # macOS
```

---

## Hardware Requirements

| Tier | RAM | CPU Cores | Disk |
|------|-----|-----------|------|
| **Minimum** | 24 GB | 6 cores | 50 GB |
| **Recommended** | 32 GB | 12 cores | 100 GB |

Running a full K3D-based stack locally with ArgoCD is resource-intensive. The minimum specs are adequate for unit test development, but the recommended specs are needed for integration testing and full-stack development.

---

## Verifying Your Setup

Run this checklist to confirm your development environment is ready:

```bash
# Go version
go version

# Module downloads work
go env GOMODCACHE

# Linter available
golangci-lint --version

# Build the project
cd openframe-cli
go build ./...

# Run unit tests
go test ./... -short
```

If all commands succeed, your environment is ready for development. Continue to the [Local Development Guide](local-development.md) to clone the repo and start running the CLI.
