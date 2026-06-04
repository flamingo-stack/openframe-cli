# Quick Start

Get a fully operational OpenFrame environment running in under 5 minutes using the `openframe bootstrap` command.

[![OpenFrame Product Walkthrough (Beta Access)](https://img.youtube.com/vi/awc-yAnkhIo/maxresdefault.jpg)](https://www.youtube.com/watch?v=awc-yAnkhIo)

---

## TL;DR

```bash
# 1. Download the CLI for your platform (example: macOS Apple Silicon)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
chmod +x openframe
sudo mv openframe /usr/local/bin/

# 2. Bootstrap a complete environment interactively
openframe bootstrap
```

That's it. The `bootstrap` command will:

1. Check and guide installation of any missing prerequisites (Docker, k3d, kubectl, Helm, Git, mkcert)
2. Create a local K3D Kubernetes cluster
3. Install ArgoCD via Helm
4. Clone and deploy the app-of-apps GitOps chart
5. Wait for all ArgoCD applications to become Healthy and Synced

---

## Step 1 — Install the CLI

### macOS (Apple Silicon)

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

### macOS (Intel)

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

### Linux (AMD64)

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

### Linux (ARM64)

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_arm64.tar.gz | tar xz
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

### Windows (AMD64)

1. Download: [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip)
2. Extract the zip archive
3. Move `openframe.exe` to a directory in your `PATH` (e.g., `C:\Windows\System32\` or a custom tools folder)

> **Windows users**: Run all commands from a WSL2 terminal for the best experience.

---

## Step 2 — Verify Installation

```bash
openframe --help
```

Expected output:

```text
   ___                    ___
  / _ \ _ __   ___ _ __  |  _|_ __ __ _ _ __ ___   ___
 | | | | '_ \ / _ \ '_ \ | |_| '__/ _` | '_ ` _ \ / _ \
 | |_| | |_) |  __/ | | ||  _| | | (_| | | | | | |  __/
  \___/| .__/ \___|_| |_||_| |_|  \__,_|_| |_| |_|\___|
       |_|

Usage:
  openframe [command]

Available Commands:
  bootstrap   Bootstrap a complete OpenFrame environment
  chart       Manage OpenFrame Helm charts
  cluster     Manage Kubernetes clusters
  dev         Development workflow tools
  help        Help about any command

Flags:
  -h, --help      help for openframe
  -v, --verbose   Enable verbose output

Use "openframe [command] --help" for more information about a command.
```

---

## Step 3 — Bootstrap Your First Environment

### Interactive Mode (Recommended for First-Time Users)

```bash
openframe bootstrap
```

The interactive wizard will guide you through:

1. **Cluster name** — Enter a name for your K3D cluster (e.g., `my-openframe`)
2. **Deployment mode** — Select `oss-tenant` for a standard self-hosted deployment
3. **Configuration mode** — Choose `default` for sensible defaults, or `interactive` for full customization
4. **Branch / Docker / Ingress** — Customize if using interactive configuration mode

### Non-Interactive Mode (CI/CD)

```bash
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive
```

---

## Step 4 — Watch the Bootstrap Progress

The bootstrap command will display real-time progress:

```text
✓ Checking prerequisites...
✓ Creating K3D cluster: my-openframe
✓ Configuring kubeconfig
✓ Installing ArgoCD via Helm
✓ Cloning app-of-apps chart repository
✓ Installing app-of-apps chart
⏳ Waiting for ArgoCD applications to sync...
  ● openframe-core        Healthy ✓
  ● openframe-frontend    Progressing...
  ● openframe-db          Healthy ✓
✓ All applications Healthy and Synced
✓ Bootstrap complete!
```

> Typical bootstrap time: **5–15 minutes** depending on network speed and hardware.

---

## Step 5 — Verify Your Environment

```bash
# Check cluster status
openframe cluster status my-openframe

# List all clusters
openframe cluster list
```

---

## Alternative: Step-by-Step Approach

If you prefer more control, run cluster creation and chart installation separately:

```bash
# Step 1: Create cluster only
openframe cluster create my-openframe --nodes 4 --skip-wizard

# Step 2: Install charts on the cluster
openframe chart install my-openframe --deployment-mode=oss-tenant
```

---

## Expected Result

After a successful bootstrap you will have:

- A running K3D Kubernetes cluster
- ArgoCD installed and accessible
- All OpenFrame applications deployed and synced via GitOps
- A local kubeconfig configured to access the cluster

---

## Troubleshooting Quick Reference

| Issue | Likely Cause | Fix |
|---|---|---|
| Docker not found | Docker not installed or not running | Start Docker Desktop / Docker daemon |
| Memory errors | Insufficient RAM allocated to Docker | Increase Docker memory limit to 16 GB+ |
| Bootstrap hangs on ArgoCD sync | Slow image pulls on first run | Wait — large images take time on first bootstrap |
| `k3d: command not found` | k3d not installed | CLI will offer to install it; or run `brew install k3d` |
| Port conflicts | Another service using required ports | Stop conflicting services before bootstrap |

---

## Next Steps

- Read the [First Steps Guide](first-steps.md) to explore your newly bootstrapped environment
- Check the [Prerequisites Guide](prerequisites.md) if you encounter tool-related errors
