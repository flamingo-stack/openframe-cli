# Prerequisites

Before installing and using the OpenFrame CLI, ensure your environment meets the following requirements. The CLI can automatically check and install most prerequisites on macOS and Linux — on Windows, you will be guided to the relevant documentation.

---

## System Requirements

| Resource | Minimum | Recommended |
|---|---|---|
| **RAM** | 24 GB | 32 GB |
| **CPU Cores** | 6 cores | 12 cores |
| **Disk Space** | 50 GB free | 100 GB free |
| **Operating System** | macOS, Linux, Windows (WSL2) | macOS or Linux |

> **Windows users:** The OpenFrame CLI runs natively on Windows but automatically forwards all operations into WSL2 (Windows Subsystem for Linux). You must have WSL2 installed and a Linux distro configured. The CLI will auto-install itself inside WSL when first run.

---

## Required Software

| Tool | Minimum Version | Purpose | Auto-Installed? |
|---|---|---|---|
| **Docker** | 24.x or newer | Container runtime for K3D clusters | ✅ macOS/Linux |
| **k3d** | 5.x or newer | Lightweight K3D cluster manager | ✅ macOS/Linux |
| **Helm** | 3.x or newer | Kubernetes package manager | ✅ macOS/Linux |
| **kubectl** | 1.28+ | Kubernetes CLI (optional — CLI uses client-go directly) | ❌ Manual |
| **WSL2** *(Windows only)* | Windows 10/11 | Linux environment on Windows | ❌ Manual |

> **Note:** Docker, k3d, and Helm can be installed automatically by running `openframe prerequisites install`. On Windows, the CLI will display documentation links for each missing tool.

---

## Operating System Details

### macOS

- macOS 12 (Monterey) or newer recommended
- [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/) or [OrbStack](https://orbstack.dev/) required
- Homebrew is recommended for manual tool management

### Linux

- Ubuntu 20.04+, Debian 11+, Fedora 36+, or any modern distribution
- Docker Engine (not just the CLI) must be running
- User must be in the `docker` group or have `sudo` access

### Windows (via WSL2)

- Windows 10 version 2004+ or Windows 11
- WSL2 enabled: run `wsl --install` in PowerShell as Administrator
- A Linux distro installed (Ubuntu recommended): `wsl --install -d Ubuntu`
- Docker Desktop for Windows with WSL2 backend enabled

---

## Account & Access Requirements

| Requirement | Details |
|---|---|
| **GitHub Access** | Required for downloading the `openframe-oss-tenant` chart and for self-updates |
| **GitHub Token** *(optional)* | Set `OPENFRAME_GITHUB_TOKEN` or `GITHUB_TOKEN` to avoid rate limiting |
| **Internet Access** | Required for downloading charts, container images, and updates |

---

## Environment Variables

The following environment variables are recognized by the CLI:

| Variable | Required | Description |
|---|---|---|
| `OPENFRAME_GITHUB_TOKEN` | Optional | GitHub personal access token (avoids API rate limits) |
| `GITHUB_TOKEN` | Optional | Standard GitHub token (also accepted) |
| `OPENFRAME_WSL_DISTRO` | Windows only | Target WSL distro name (default: WSL default distro) |
| `OPENFRAME_NO_WSL_FORWARD` | Windows only | Disable WSL forwarding (unsupported; use at your own risk) |
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Emergency only | Skip cosign signature verification during updates |
| `KUBECONFIG` | Optional | Path to kubeconfig file (default: `~/.kube/config`) |

---

## Verification Commands

Run these commands to verify your environment is ready before installing OpenFrame CLI:

### Check Docker

```bash
docker --version
docker ps
```

Expected output: Docker version and an empty container list (confirms Docker is running).

### Check k3d

```bash
k3d version
```

Expected output: `k3d version vX.Y.Z`

### Check Helm

```bash
helm version
```

Expected output: `version.BuildInfo{Version:"vX.Y.Z", ...}`

### Check available memory

```bash
# macOS
sysctl -n hw.memsize | awk '{print $1/1024/1024/1024 " GB"}'

# Linux
free -h
```

Ensure at least 24 GB RAM is available.

### Check disk space

```bash
df -h .
```

Ensure at least 50 GB free on the relevant partition.

### Run CLI prerequisite check (after installing OpenFrame CLI)

```bash
openframe prerequisites check
```

This is the most comprehensive check — the CLI will display exactly what is missing and how to fix it.

---

## Windows-Specific Setup

<details>
<summary>Expand Windows WSL2 Setup Steps</summary>

**Step 1: Enable WSL2**

Open PowerShell as Administrator and run:

```bash
wsl --install
```

**Step 2: Install Ubuntu distro**

```bash
wsl --install -d Ubuntu
```

**Step 3: Install Docker Desktop**

Download and install [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/install/). In Docker Desktop settings, enable:
- **WSL2 backend** (Settings → General → Use WSL2 based engine)
- **Ubuntu integration** (Settings → Resources → WSL Integration → Ubuntu)

**Step 4: Set WSL2 as default (if needed)**

```bash
wsl --set-default-version 2
wsl --set-default Ubuntu
```

**Step 5: Download the Windows CLI binary**

Download from: https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip

Extract and run the `.exe` — the CLI will automatically forward into WSL2.

</details>

---

## Next Steps

- Proceed to the [Quick Start Guide](quick-start.md) to install and run OpenFrame CLI
- Return to the [Introduction](introduction.md) for a feature overview
