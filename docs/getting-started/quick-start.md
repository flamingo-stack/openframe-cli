# Quick Start Guide

Get OpenFrame up and running in under 5 minutes. This guide covers the fastest path to a working OpenFrame environment.

---

## TL;DR — 5-Minute Setup

```bash
# 1. Download the CLI for your platform (see below)
# 2. Check prerequisites
openframe prerequisites check

# 3. Bootstrap a full OpenFrame environment
openframe bootstrap
```

That's it. The `bootstrap` command creates a local K3D cluster, installs ArgoCD, deploys the full OpenFrame platform, and waits for everything to become healthy.

---

## Step 1: Download the OpenFrame CLI

Choose your platform:

### macOS (Apple Silicon / Intel)

```bash
# Apple Silicon (M1/M2/M3)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Intel
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

### Linux (amd64)

```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

### Windows (amd64)

Download: https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip

Extract the ZIP archive and run the `openframe.exe` — it will automatically forward commands into WSL2.

### Browse All Releases

Visit [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases) for all available platform binaries.

---

## Step 2: Verify Installation

```bash
openframe --version
```

Expected output:

```text
openframe version v1.x.x (abc1234) built on 2024-xx-xx
```

---

## Step 3: Check Prerequisites

```bash
openframe prerequisites check
```

The CLI will inspect your environment and report the status of required tools:

```text
✓ Docker    - running
✓ k3d       - v5.x.x
✓ Helm      - v3.x.x
```

If any prerequisites are missing, install them automatically:

```bash
openframe prerequisites install
```

> **Windows users:** Auto-install is not supported on native Windows. The CLI will display documentation links for each missing tool instead.

---

## Step 4: Bootstrap OpenFrame

Run the interactive bootstrap wizard:

```bash
openframe bootstrap
```

The wizard will guide you through:

1. **Cluster name** — the name for your local K3D cluster (default: `openframe-dev`)
2. **Configuration mode** — default settings or interactive customization
3. **Branch/version** — which OpenFrame release to deploy

To use all defaults without prompts (e.g. in CI):

```bash
openframe bootstrap --non-interactive
```

### Expected Output

```text
  ___                  ___
 / _ \ _ __   ___ _ __|  _|_ __ __ _ _ __ ___   ___
| | | | '_ \ / _ \ '_ \ |_| '__/ _` | '_ ` _ \ / _ \
| |_| | |_) |  __/ | | |  _| | | (_| | | | | | |  __/
 \___/| .__/ \___|_| |_|_| |_|  \__,_|_| |_| |_|\___|
      |_|

✓ Prerequisites validated
✓ Creating cluster: openframe-dev
✓ Cluster ready
✓ Installing ArgoCD
✓ ArgoCD ready
✓ Deploying OpenFrame platform
✓ Waiting for applications...
✓ All applications Healthy + Synced

Bootstrap complete! 🎉
```

---

## Step 5: Check Status

After bootstrapping, verify everything is running:

```bash
openframe app status
```

```bash
openframe cluster status
```

---

## What Was Installed?

After a successful `openframe bootstrap`, you have:

| Component | Description |
|---|---|
| **K3D cluster** | A local lightweight Kubernetes cluster named `openframe-dev` |
| **ArgoCD** | GitOps continuous delivery engine managing your platform |
| **OpenFrame platform** | The full OSS tenant chart from [openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant) |

---

## Common Next Actions

After bootstrap completes, you may want to:

```bash
# View all available commands
openframe --help

# Check cluster list
openframe cluster list

# Get access information
openframe app access

# Upgrade to a new OpenFrame version
openframe app upgrade

# Keep the CLI itself up to date
openframe update
```

---

## Troubleshooting Quick Fixes

| Problem | Solution |
|---|---|
| `docker: command not found` | Install Docker: `openframe prerequisites install` |
| `connection refused` | Check `openframe cluster status` — the cluster may not be running |
| `context deadline exceeded` | Network/resource issue; wait and retry, or check system resources |
| Missing `openframe-helm-values.yaml` | The bootstrap wizard will create it for you in non-interactive mode |
| Permission denied on binary | `chmod +x /usr/local/bin/openframe` |

---

## Next Steps

- Read the [First Steps Guide](first-steps.md) for what to explore after your first bootstrap
- Review the [Prerequisites Guide](prerequisites.md) if you encounter environment issues
- Visit the [OpenMSP community](https://www.openmsp.ai/) for help and discussion
