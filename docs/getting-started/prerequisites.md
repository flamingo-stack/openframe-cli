# Prerequisites

Before you can run OpenFrame CLI, you need several tools installed and configured on your system. The CLI validates these prerequisites automatically before any operation — if something is missing it will tell you exactly how to install it.

---

## System Requirements

| Tier | RAM | CPU Cores | Disk Space |
|------|-----|-----------|------------|
| **Minimum** | 24 GB | 6 cores | 50 GB |
| **Recommended** | 32 GB | 12 cores | 100 GB |

> **Note:** Running a full K3D-based Kubernetes stack with ArgoCD and the OpenFrame application suite is memory-intensive. The minimum specs will work for single-service development, but the recommended specs are required for stable full-stack operation.

---

## Required Software

All of the following tools must be installed before the CLI can run. The prerequisite checker validates each one at startup.

| Tool | Minimum Version | Purpose | Check Command |
|------|----------------|---------|---------------|
| **Go** | 1.21+ | Build the CLI from source (if not using a binary release) | `go version` |
| **Docker** | 20.10+ | Runs the K3D container nodes; daemon must be running | `docker info` |
| **k3d** | 5.x | Creates and manages K3D (K3s-in-Docker) clusters | `k3d version` |
| **kubectl** | 1.25+ | Interacts with Kubernetes clusters | `kubectl version --client` |
| **Helm** | 3.x | Installs ArgoCD and app-of-apps charts | `helm version` |
| **Git** | 2.x | Clones the app-of-apps chart repository from GitHub | `git --version` |
| **mkcert** | 1.4+ | Generates locally-trusted TLS certificates for ingress | `mkcert --version` |

### For Development Workflows (Optional)

These are only required when using `openframe dev` commands:

| Tool | Purpose | Check Command |
|------|---------|---------------|
| **Telepresence** | 2.x | Routes live Kubernetes traffic to local processes | `telepresence version` |
| **Skaffold** | Latest | Hot-reload development sessions | `skaffold version` |
| **jq** | 1.6+ | JSON processing used by intercept workflows | `jq --version` |

---

## Platform-Specific Requirements

### Linux / macOS

No additional requirements. Ensure Docker Desktop (macOS) or Docker Engine (Linux) is running before invoking the CLI.

### Windows (WSL2)

OpenFrame CLI supports Windows via WSL2. The following additional setup is needed:

1. **WSL2 enabled** with an Ubuntu distribution installed
2. **Docker Desktop** configured with WSL2 backend enabled
3. All tools above installed **inside WSL2** (not the native Windows environment)

> **Windows users:** The CLI auto-detects WSL2 and wraps commands appropriately. Run `openframe` from within your WSL2 terminal session.

---

## Environment Variables

The following environment variables affect CLI behavior. None are required for basic operation, but may be needed in specific setups:

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` |
| `OPENFRAME_VERBOSE` | Enable verbose output globally | `false` |
| `CI` | Disables interactive prompts in CI mode | unset |

> When `CI` is set (to any value), the CLI switches to non-interactive mode and skips prompts that require keyboard input.

---

## Account & Access Requirements

### GitHub Access

The CLI clones the OpenFrame app-of-apps Helm chart from a GitHub repository during `chart install` and `bootstrap`. You need:

- Network access to `github.com`
- If using a private repository: a valid GitHub token or SSH key configured

### Container Registry Access

For custom deployments using `saas-tenant` or `saas-shared` mode, you may need credentials for:

- **GitHub Container Registry (GHCR):** `ghcr.io`
- **Custom Docker registry:** configured during the installation wizard

---

## Verification Commands

Run the following to confirm all prerequisites are installed and working:

```bash
# Verify Go installation
go version

# Verify Docker is installed and daemon is running
docker info

# Verify k3d
k3d version

# Verify kubectl
kubectl version --client

# Verify Helm
helm version

# Verify Git
git --version

# Verify mkcert
mkcert --version

# Optional: dev workflow tools
telepresence version
skaffold version
jq --version
```

If any command fails or returns an error, install the missing tool using your system package manager or the official installation instructions.

---

## Quick Install References

<details>
<summary>Install k3d</summary>

```bash
# Using the official install script
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Or with Homebrew (macOS/Linux)
brew install k3d
```

</details>

<details>
<summary>Install kubectl</summary>

```bash
# Linux (AMD64)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl && sudo mv kubectl /usr/local/bin/

# macOS
brew install kubectl
```

</details>

<details>
<summary>Install Helm</summary>

```bash
# Using the official install script
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Or with Homebrew
brew install helm
```

</details>

<details>
<summary>Install mkcert</summary>

```bash
# macOS
brew install mkcert
mkcert -install

# Linux (using pre-built binary)
curl -LO https://github.com/FiloSottile/mkcert/releases/latest/download/mkcert-v1.4.4-linux-amd64
chmod +x mkcert-v1.4.4-linux-amd64 && sudo mv mkcert-v1.4.4-linux-amd64 /usr/local/bin/mkcert
mkcert -install
```

</details>

<details>
<summary>Install Telepresence</summary>

```bash
# macOS/Linux
curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o telepresence
chmod +x telepresence && sudo mv telepresence /usr/local/bin/
```

</details>

---

## Automated Prerequisite Check

The CLI runs its own prerequisite checker before every cluster or chart operation. If you prefer to check manually:

```bash
# The CLI will check and report missing tools when you run any command
openframe cluster list
```

Missing tools will be reported with platform-specific installation instructions printed directly in the terminal.

---

## Next Steps

Once all prerequisites are verified, follow the [Quick Start Guide](quick-start.md) to bootstrap your first OpenFrame environment.
