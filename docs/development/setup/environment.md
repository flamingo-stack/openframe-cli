# Development Environment Setup

This guide covers setting up a development environment for contributing to the OpenFrame CLI.

---

## Required Tools

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.21+ | Primary language runtime and toolchain |
| **Git** | 2.30+ | Version control |
| **Docker** | 24.x+ | Container runtime (required for integration tests) |
| **k3d** | 5.x+ | Local Kubernetes clusters (integration tests) |
| **Helm** | 3.x+ | Kubernetes package manager (integration tests) |
| **Make** | Any | Build automation (if Makefile is present) |

---

## Installing Go

### macOS

```bash
# Using Homebrew
brew install go

# Verify
go version
```

### Linux

```bash
# Download the latest Go release
curl -OL https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin

# Verify
go version
```

### Windows (WSL2)

Follow the Linux instructions inside your WSL2 terminal.

---

## IDE Recommendations

### Visual Studio Code (Recommended)

VS Code with the Go extension provides the best development experience for this project.

**Install the Go extension:**

```bash
code --install-extension golang.go
```

**Recommended VS Code extensions:**

| Extension | ID | Purpose |
|---|---|---|
| Go | `golang.go` | Go language support, debugging, testing |
| GitLens | `eamodio.gitlens` | Enhanced Git integration |
| YAML | `redhat.vscode-yaml` | YAML editing for Helm values |
| Docker | `ms-azuretools.vscode-docker` | Docker integration |
| Markdown All in One | `yzhang.markdown-all-in-one` | Documentation editing |

**Recommended `settings.json` for Go development:**

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "[go]": {
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

### GoLand (JetBrains)

GoLand offers excellent Go support with built-in refactoring tools. No additional plugins required — all Go features are built in.

### Neovim / Vim

Use `gopls` (Go language server) via `nvim-lspconfig`:

```bash
go install golang.org/x/tools/gopls@latest
```

---

## Go Environment Configuration

### Verify GOPATH and module mode

```bash
go env GOPATH
go env GOMODCACHE
go env GOFLAGS
```

The project uses Go modules (`go.mod`), so `GOFLAGS` should not set `-mod=vendor` unless you're working with a vendor directory.

### Configure GOPRIVATE (if needed)

If your environment restricts access to the Flamingo private modules, configure:

```bash
go env -w GOPRIVATE=github.com/flamingo-stack
```

---

## Linting and Code Quality Tools

Install the Go linting toolchain used in the project:

```bash
# golangci-lint (recommended)
curl -sSfL https://raw.githubusercontent.com/golangci-lint/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# goimports (import management + formatting)
go install golang.org/x/tools/cmd/goimports@latest

# staticcheck (additional static analysis)
go install honnef.co/go/tools/cmd/staticcheck@latest
```

Verify:

```bash
golangci-lint --version
goimports -h
staticcheck -version
```

---

## Environment Variables for Development

| Variable | Description | Example |
|---|---|---|
| `OPENFRAME_GITHUB_TOKEN` | GitHub token for API calls (avoids rate limits) | Your personal access token |
| `GOFLAGS` | Go build flags | `-v` for verbose builds |
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Skip cosign verification (dev/testing only) | `1` |

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
export OPENFRAME_GITHUB_TOKEN="your-github-token-here"
```

---

## Pre-Commit Hooks (Optional)

Setting up pre-commit hooks ensures code quality before every commit:

```bash
# Install pre-commit
pip install pre-commit
# or: brew install pre-commit

# Install hooks (from repo root)
pre-commit install
```

Alternatively, add a manual hook to `.git/hooks/pre-commit`:

```bash
#!/bin/sh
set -e
go vet ./...
goimports -l .
```

---

## Verifying Your Setup

Run these commands from the repository root to confirm your environment is ready:

```bash
# Verify Go version
go version

# Download dependencies
go mod download

# Verify all dependencies resolve
go mod verify

# Build the binary
go build -o openframe .

# Run unit tests
go test ./...

# Run vet
go vet ./...
```

A successful run of all the above indicates a correctly configured development environment.

---

## Next Steps

- Follow the [Local Development Guide](local-development.md) to clone, build, and run the CLI
- Review the [Architecture Overview](../architecture/README.md) to understand the codebase
