# Prerequisites

Before running OpenFrame CLI, ensure your system meets the hardware requirements and has all required software installed. The CLI will also perform its own prerequisite checks and provide guided installation instructions for any missing tools.

---

## Hardware Requirements

| Tier | RAM | CPU Cores | Disk Space |
|---|---|---|---|
| **Minimum** | 24 GB | 6 cores | 50 GB |
| **Recommended** | 32 GB | 12 cores | 100 GB |

> K3D runs Kubernetes nodes as Docker containers. Insufficient memory is the most common cause of failed bootstraps — ensure Docker has access to at least 16 GB RAM.

---

## Required Software

The OpenFrame CLI depends on the following tools being installed on your system. The CLI will check for these automatically and provide installation guidance if any are missing.

### Core Prerequisites (Cluster Operations)

| Tool | Minimum Version | Purpose | Install Guide |
|---|---|---|---|
| **Docker** | 20.10+ | Container runtime for K3D nodes | [docs.docker.com](https://docs.docker.com/get-docker/) |
| **kubectl** | 1.26+ | Kubernetes CLI for cluster interaction | [kubernetes.io/docs](https://kubernetes.io/docs/tasks/tools/) |
| **k3d** | 5.6+ | Lightweight Kubernetes via K3D | [k3d.io](https://k3d.io/#installation) |
| **Helm** | 3.12+ | Kubernetes package manager | [helm.sh](https://helm.sh/docs/intro/install/) |

### Additional Prerequisites (Chart Installation)

| Tool | Minimum Version | Purpose | Install Guide |
|---|---|---|---|
| **Git** | 2.30+ | Clone app-of-apps chart repositories | [git-scm.com](https://git-scm.com/downloads) |
| **mkcert** | 1.4+ | Generate local TLS certificates | [github.com/FiloSottile/mkcert](https://github.com/FiloSottile/mkcert) |

### Developer Prerequisites (Dev Workflows Only)

| Tool | Minimum Version | Purpose | Install Guide |
|---|---|---|---|
| **Telepresence** | 2.x | Route K8s traffic to local machine | [telepresence.io](https://www.telepresence.io/docs/latest/install/) |
| **Skaffold** | 2.x | Hot-reload workflow for services | [skaffold.dev](https://skaffold.dev/docs/install/) |
| **jq** | 1.6+ | JSON processing for dev scripts | [jqlang.github.io](https://jqlang.github.io/jq/download/) |

---

## Operating System Support

| Platform | Status | Notes |
|---|---|---|
| **macOS** (Intel / Apple Silicon) | ✅ Fully supported | Docker Desktop or Colima recommended |
| **Linux** (Ubuntu, Debian, RHEL, Arch) | ✅ Fully supported | Native Docker Engine |
| **Windows** (WSL2) | ✅ Supported | Requires WSL2 with Docker Desktop or Docker Engine in WSL2 |

> **Windows users**: The CLI includes special WSL2 integration for Docker, IP detection, and inotify limits. Ensure WSL2 is enabled and Docker is accessible from within your WSL2 distribution.

---

## OpenFrame CLI Binary

Download the latest `openframe` CLI binary for your platform:

| Platform | Architecture | Download |
|---|---|---|
| **macOS** | ARM64 (Apple Silicon) | [openframe-cli_darwin_arm64](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz) |
| **macOS** | AMD64 (Intel) | [openframe-cli_darwin_amd64](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz) |
| **Linux** | AMD64 | [openframe-cli_linux_amd64](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz) |
| **Linux** | ARM64 | [openframe-cli_linux_arm64](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_arm64.tar.gz) |
| **Windows** | AMD64 | [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip) |

---

## Environment Variables

The following environment variables are recognized by the CLI:

| Variable | Purpose | Example |
|---|---|---|
| `OPENFRAME_FANCY_LOGO` | Enable/disable fancy terminal logo rendering | `true` / `false` |
| `NO_COLOR` | Disable all colored terminal output | `1` |
| `TERM` | Terminal type detection for UI rendering | `xterm-256color` |
| `KUBECONFIG` | Path to kubeconfig file (standard Kubernetes) | `~/.kube/config` |

---

## Verification Commands

Run these commands to confirm your environment is ready before running the CLI:

```bash
# Verify Docker is running
docker info

# Verify kubectl is available
kubectl version --client

# Verify k3d is installed
k3d version

# Verify Helm is installed
helm version

# Verify Git is installed
git --version

# Verify mkcert is installed
mkcert --version
```

Expected output example for Docker:

```text
Client:
 Version:           24.0.5
 API version:       1.43
 ...
Server: Docker Engine - Community
 Engine:
  Version:          24.0.5
```

---

## Automatic Prerequisite Checking

The CLI performs automatic prerequisite validation before every operation. If a required tool is missing, it will display:

1. Which tool is missing
2. Platform-specific installation instructions
3. An option to attempt automatic installation

```text
⚠  Prerequisites check failed:
   Missing: k3d
   Install: brew install k3d   (macOS)
            curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash   (Linux)
```

> You do not need to manually install all tools before first use — the CLI will guide you.

---

## Memory Requirements for K3D

The CLI checks available system memory before cluster creation. A typical OpenFrame deployment requires:

- **Minimum**: 8 GB available to Docker
- **Recommended**: 16 GB+ available to Docker

On macOS with Docker Desktop, ensure the memory limit in Docker Desktop preferences is set to at least 16 GB.

---

## Next Steps

- Follow the [Quick Start Guide](quick-start.md) to bootstrap your first environment
- Read the [First Steps Guide](first-steps.md) once your environment is running
