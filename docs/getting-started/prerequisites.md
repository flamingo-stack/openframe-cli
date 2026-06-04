# Prerequisites

Before you can install and use OpenFrame CLI, you need to ensure your system meets the hardware requirements and has the required tools installed. The CLI automatically checks for prerequisites and will guide you through installing anything that is missing.

---

## Hardware Requirements

| Tier | RAM | CPU Cores | Disk Space |
|---|---|---|---|
| **Minimum** | 24 GB | 6 cores | 50 GB |
| **Recommended** | 32 GB | 12 cores | 100 GB |

> **Note:** K3D runs Kubernetes nodes as Docker containers. Insufficient memory is the most common cause of failed bootstraps. Always ensure you have at least 24 GB of RAM available before running `openframe bootstrap`.

---

## Required Software

The following tools must be installed and accessible on your `$PATH` before running cluster or chart operations. The OpenFrame CLI prerequisite checker will validate all of these at startup.

### Cluster Prerequisites

| Tool | Minimum Version | Purpose | Install Guide |
|---|---|---|---|
| **Docker** | 20.x+ (daemon running) | Container runtime for K3D nodes | [docker.com/get-docker](https://docs.docker.com/get-docker/) |
| **kubectl** | 1.25+ | Kubernetes API client | [kubernetes.io/docs/tasks/tools](https://kubernetes.io/docs/tasks/tools/) |
| **k3d** | 5.x+ | K3D cluster management | [k3d.io](https://k3d.io/#installation) |
| **Helm** | 3.x+ | Kubernetes package manager | [helm.sh/docs/intro/install](https://helm.sh/docs/intro/install/) |

### Chart Prerequisites

| Tool | Minimum Version | Purpose | Install Guide |
|---|---|---|---|
| **Git** | 2.x+ | Cloning app-of-apps repositories | [git-scm.com/downloads](https://git-scm.com/downloads) |
| **Helm** | 3.x+ | Helm chart operations | [helm.sh/docs/intro/install](https://helm.sh/docs/intro/install/) |
| **mkcert** | Latest | Local TLS certificate generation | [github.com/FiloSottile/mkcert](https://github.com/FiloSottile/mkcert) |

### Developer Workflow Prerequisites (Optional)

These are only required if you use `openframe dev` commands:

| Tool | Purpose | Install Guide |
|---|---|---|
| **Telepresence** | Service traffic interception | [telepresence.io/docs/install](https://www.telepresence.io/docs/install/client) |
| **Skaffold** | Live-reload development sessions | [skaffold.dev/docs/install](https://skaffold.dev/docs/install/) |

---

## Operating System Support

| OS | Status | Notes |
|---|---|---|
| **Linux** (x86_64, arm64) | ✅ Fully supported | Primary development platform |
| **macOS** (Intel & Apple Silicon) | ✅ Fully supported | Docker Desktop required |
| **Windows** (via WSL2) | ✅ Supported | WSL2 + Docker Desktop required; WSL2 IP detection is automatic |

### Windows Requirements

Windows users must use **WSL2** (Windows Subsystem for Linux 2). The CLI includes first-class WSL2 support with automatic IP detection and inotify limit configuration.

1. Install [WSL2](https://docs.microsoft.com/en-us/windows/wsl/install)
2. Install [Docker Desktop for Windows](https://www.docker.com/products/docker-desktop/) and enable WSL2 integration
3. Install the OpenFrame CLI inside your WSL2 environment or use the Windows AMD64 binary

**Windows AMD64 binary**: [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip)

---

## Go Build Requirements (Source Builds Only)

If you are building from source, you also need:

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.22+ | Compiling the CLI from source |

---

## Environment Variables

The following environment variables may be relevant depending on your deployment mode:

| Variable | Required For | Description |
|---|---|---|
| `KUBECONFIG` | All cluster operations | Path to your kubeconfig file (defaults to `~/.kube/config`) |
| GitHub Personal Access Token | `saas-tenant` / `saas-shared` modes | Accessing private GHCR container registry |

> **Tip:** For `oss-tenant` mode (the default self-hosted deployment), no special environment variables are required beyond a working Docker daemon.

---

## Verification Commands

Run these commands to confirm all prerequisites are installed and ready:

```bash
# Verify Docker is running
docker info

# Verify kubectl
kubectl version --client

# Verify k3d
k3d version

# Verify Helm
helm version

# Verify Git
git --version

# Verify mkcert
mkcert --version
```

All commands should return version information without errors. If Docker is not running, start it before proceeding.

---

## Automatic Prerequisite Checking

When you run any `cluster` or `chart` command, the CLI automatically runs a prerequisite check. Missing tools are listed with platform-specific installation instructions:

```bash
openframe cluster create
# If prerequisites are missing, you will see:
# ✗ Missing: k3d
# Install: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

You don't need to memorize all installation commands — the CLI's built-in checker will guide you interactively.

---

## Next Steps

Once your prerequisites are confirmed, continue to the [Quick Start Guide](quick-start.md) to bootstrap your first OpenFrame environment.
