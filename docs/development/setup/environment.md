# Development Environment Setup

This guide covers the IDE configuration, development tools, and editor extensions recommended for contributing to OpenFrame CLI.

---

## Required Development Tools

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.21+ | Primary language runtime and toolchain |
| **Git** | 2.30+ | Version control |
| **Docker** | 20.10+ | Required for running K3D clusters during development |
| **k3d** | 5.6+ | Local Kubernetes cluster provider |
| **kubectl** | 1.26+ | Kubernetes CLI for cluster interaction |
| **Helm** | 3.12+ | Kubernetes package manager |
| **mkcert** | 1.4+ | Local TLS certificate generation |
| **Make** | 3.81+ | Build automation (optional but recommended) |

---

## Installing Go

### macOS

```bash
brew install go
```

### Linux

```bash
# Download and install Go (replace version as needed)
curl -L https://go.dev/dl/go1.21.13.linux-amd64.tar.gz | sudo tar -C /usr/local -xz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

### Verify

```bash
go version
# go version go1.21.x linux/amd64
```

---

## Recommended IDE

### Visual Studio Code

VS Code with the Go extension is the recommended IDE for OpenFrame CLI development.

**Install the Go extension:**

```bash
code --install-extension golang.go
```

**Recommended VS Code extensions:**

| Extension | ID | Purpose |
|---|---|---|
| Go | `golang.go` | Go language support, IntelliSense, debugging |
| GitLens | `eamodio.gitlens` | Enhanced Git integration |
| YAML | `redhat.vscode-yaml` | YAML editing with schema validation |
| Docker | `ms-azuretools.vscode-docker` | Docker file editing and container management |
| Kubernetes | `ms-kubernetes-tools.vscode-kubernetes-tools` | Kubernetes manifest editing |

**Recommended VS Code settings for Go development:**

```json
{
  "go.useLanguageServer": true,
  "go.lintOnSave": "package",
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "go.testOnSave": false,
  "go.coverOnSave": false,
  "editor.formatOnSave": true,
  "[go]": {
    "editor.defaultFormatter": "golang.go",
    "editor.codeActionsOnSave": {
      "source.organizeImports": "explicit"
    }
  }
}
```

### GoLand (JetBrains)

GoLand is a fully featured Go IDE that works well with this codebase. No additional plugins are required beyond the built-in Go support.

---

## Go Tools Setup

Install the standard Go development tools used in this project:

```bash
# Install golangci-lint (linter)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goimports (import formatting)
go install golang.org/x/tools/cmd/goimports@latest

# Install govulncheck (vulnerability scanner)
go install golang.org/x/vuln/cmd/govulncheck@latest
```

Verify the tools are in your `PATH`:

```bash
golangci-lint version
goimports --help
govulncheck --help
```

---

## Environment Variables for Development

Set the following environment variables in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) for a comfortable development experience:

```bash
# Go workspace
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin"

# Optional: Enable fancy logo in terminal
export OPENFRAME_FANCY_LOGO=true

# Optional: Disable color output in CI-like environments
# export NO_COLOR=1
```

Apply the changes:

```bash
source ~/.bashrc   # or ~/.zshrc for zsh users
```

---

## Docker Configuration

The OpenFrame CLI creates K3D clusters as Docker containers. Ensure Docker has sufficient resources:

### Docker Desktop (macOS / Windows)

1. Open Docker Desktop → **Settings** → **Resources**
2. Set **Memory** to at least **16 GB** (24 GB recommended)
3. Set **CPUs** to at least **6** (12 recommended)
4. Apply and Restart

### Docker Engine (Linux)

Docker Engine on Linux uses the host system's resources directly. Ensure your machine meets the [hardware requirements](../../getting-started/prerequisites.md).

---

## Terminal Recommendations

The OpenFrame CLI uses rich terminal UI (pterm, promptui) that works best with a modern terminal emulator:

| Platform | Recommended Terminal |
|---|---|
| macOS | iTerm2, Warp, or macOS Terminal with zsh |
| Linux | Any xterm-256color compatible terminal (GNOME Terminal, Alacritty, Kitty) |
| Windows | Windows Terminal + WSL2 |

Ensure your terminal supports 256 colors:

```bash
echo $TERM
# Should output: xterm-256color or similar
```

---

## WSL2 Setup (Windows Only)

If developing on Windows, use WSL2:

```bash
# Install WSL2 (PowerShell as Administrator)
wsl --install

# Set WSL2 as default
wsl --set-default-version 2

# Install Ubuntu (or your preferred distro)
wsl --install -d Ubuntu
```

After installing WSL2, follow the Linux installation steps for Go, Docker Engine, and other tools inside your WSL2 distribution.

> The OpenFrame CLI includes built-in WSL2 support for Docker daemon detection, IP detection, and inotify limit configuration.

---

## Verifying Your Environment

Run this checklist to confirm your development environment is ready:

```bash
# 1. Go toolchain
go version

# 2. Git
git --version

# 3. Docker
docker info

# 4. kubectl
kubectl version --client

# 5. k3d
k3d version

# 6. Helm
helm version

# 7. Linting tools
golangci-lint version

# 8. Clone and build the CLI (see Local Development guide)
```

---

## Next Steps

- Proceed to [Local Development](local-development.md) to clone, build, and run the CLI from source
