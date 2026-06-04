# Development Environment Setup

This guide walks you through configuring your local development environment for contributing to or extending the OpenFrame CLI.

---

## Required Development Tools

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.22+ | Primary language — compile and test the CLI |
| **Git** | 2.x+ | Version control |
| **Docker** | 20.x+ (daemon running) | K3D cluster container runtime |
| **k3d** | 5.x+ | Local Kubernetes clusters for integration testing |
| **kubectl** | 1.25+ | Kubernetes API client |
| **Helm** | 3.x+ | Kubernetes package manager |
| **mkcert** | Latest | Local TLS certificate generation |
| **make** | Any | Build task runner (optional, but helpful) |

---

## Installing Go

### Linux / macOS

```bash
# Download and install Go (replace VERSION with latest, e.g. 1.22.4)
curl -Lo go.tar.gz https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
source ~/.profile
```

Verify:

```bash
go version
```

### Windows (WSL2)

Install Go inside your WSL2 environment using the Linux instructions above. Do not install the Windows Go binary for CLI development — the project compiles inside the WSL2 Linux environment.

---

## IDE Recommendations

### Visual Studio Code (Recommended)

VS Code with the official Go extension provides the best experience for this project.

**Required Extensions:**

| Extension | ID | Purpose |
|---|---|---|
| Go | `golang.go` | Language server, debugging, testing |
| YAML | `redhat.vscode-yaml` | Helm values file editing |
| Docker | `ms-azuretools.vscode-docker` | Dockerfile and container management |

**Install all at once:**

```bash
code --install-extension golang.go
code --install-extension redhat.vscode-yaml
code --install-extension ms-azuretools.vscode-docker
```

**Recommended VS Code settings for this project** (`.vscode/settings.json`):

```json
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

### GoLand (JetBrains)

GoLand provides excellent built-in Go support with no additional plugins required. Recommended for developers who prefer a full IDE experience.

---

## Go Toolchain Setup

After installing Go, install these additional development tools:

```bash
# Install goimports (code formatter that also manages imports)
go install golang.org/x/tools/cmd/goimports@latest

# Install golangci-lint (comprehensive linter)
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.57.0

# Install govulncheck (security vulnerability scanner)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Verify the linter:

```bash
golangci-lint --version
```

---

## Environment Variables for Development

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`, or equivalent):

```bash
# Go workspace
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

# Optional: use a specific kubeconfig for development
export KUBECONFIG="$HOME/.kube/config"
```

For `saas-tenant` or `saas-shared` development, you may also need:

```bash
# GitHub credentials for private GHCR access (SaaS modes only)
export GITHUB_TOKEN="your-personal-access-token"
```

> **Security note:** Never commit tokens to version control. Use environment variables or a secrets manager.

---

## Docker Desktop Configuration (macOS / Windows)

For local K3D development, configure Docker Desktop with sufficient resources:

| Resource | Minimum | Recommended |
|---|---|---|
| Memory | 12 GB | 16+ GB |
| CPUs | 4 | 6+ |
| Disk | 50 GB | 100 GB |

On macOS: **Docker Desktop → Settings → Resources** and adjust sliders accordingly.

---

## WSL2-Specific Setup

If developing on Windows with WSL2:

```bash
# Increase inotify limits (required for Skaffold file watching inside WSL2)
echo "fs.inotify.max_user_watches=524288" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Verify Docker is accessible from WSL2
docker info
```

The OpenFrame CLI automatically detects WSL2 and applies platform-specific configuration (IP detection, kubeconfig permission fixes) when creating K3D clusters.

---

## Verifying Your Dev Environment

Run this checklist to confirm your environment is ready for development:

```bash
# Language
go version          # Should show 1.22+

# VCS
git --version       # Should show 2.x+

# Container runtime
docker info         # Should show Docker daemon info (not an error)

# Kubernetes tools
k3d version         # Should show k3d 5.x+
kubectl version --client   # Should show 1.25+
helm version        # Should show 3.x+

# Certificate tool
mkcert --version    # Should show a version

# Linter
golangci-lint --version   # Should show installed version
```

All checks passing? You're ready to move on to [Local Development](local-development.md).
