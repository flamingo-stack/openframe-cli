# Quick Start

Get a fully operational OpenFrame Kubernetes environment running in minutes using a single command.

[![OpenFrame v0.3.7 - Enhanced Developer Experience](https://img.youtube.com/vi/O8hbBO5Mym8/maxresdefault.jpg)](https://www.youtube.com/watch?v=O8hbBO5Mym8)

---

## TL;DR — 5-Minute Setup

```bash
# 1. Download the CLI binary for your platform (example: Linux AMD64)
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz
tar -xzf openframe-cli_linux_amd64.tar.gz
chmod +x openframe
sudo mv openframe /usr/local/bin/openframe

# 2. Bootstrap a complete environment (interactive mode)
openframe bootstrap

# 3. That's it — your cluster and full OpenFrame stack are running!
```

> **Windows (AMD64) users:** Download from [https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip), extract, and run the installer the same way as other platforms.

---

## Step-by-Step Installation

### Step 1 — Download the CLI

Choose the binary for your operating system:

| Platform | Download |
|----------|----------|
| Linux AMD64 | `openframe-cli_linux_amd64.tar.gz` |
| macOS (Apple Silicon) | `openframe-cli_darwin_arm64.tar.gz` |
| macOS (Intel) | `openframe-cli_darwin_amd64.tar.gz` |
| Windows AMD64 | `openframe-cli_windows_amd64.zip` |

All releases are available at:
[https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases)

### Step 2 — Install the Binary

```bash
# Linux / macOS
tar -xzf openframe-cli_<os>_<arch>.tar.gz
chmod +x openframe
sudo mv openframe /usr/local/bin/openframe

# Verify installation
openframe --version
```

### Step 3 — Bootstrap Your Environment

The `bootstrap` command does everything in one shot:

1. Checks all prerequisites
2. Creates a K3D local Kubernetes cluster
3. Installs ArgoCD via Helm
4. Clones the OpenFrame Helm chart repository
5. Deploys the app-of-apps stack
6. Waits for all ArgoCD Applications to reach Healthy+Synced state

```bash
# Interactive mode — the wizard asks for cluster name and deployment mode
openframe bootstrap

# Non-interactive OSS tenant setup (recommended for first-time)
openframe bootstrap my-cluster --deployment-mode=oss-tenant

# Verbose output to see ArgoCD sync progress
openframe bootstrap my-cluster --deployment-mode=oss-tenant -v
```

---

## Expected Output

When bootstrap runs successfully, you'll see output similar to:

```text
  ___                    _____
 / _ \ _ __   ___ _ __ |  ___| __ __ _ _ __ ___   ___
| | | | '_ \ / _ | '_ \| |_ | '__/ _` | '_ ` _ \ / _ \
| |_| | |_) |  __| | | |  _|| | | (_| | | | | | |  __/
 \___/| .__/ \___|_| |_|_|  |_|  \__,_|_| |_| |_|\___|
      |_|

✓ Prerequisites validated
✓ Cluster "my-cluster" created
✓ ArgoCD installed
✓ App-of-apps deployed
✓ All applications synced and healthy

Your OpenFrame environment is ready!
```

---

## Basic "Hello World" — Verify Your Environment

After bootstrap completes, run these commands to confirm everything is working:

```bash
# List your cluster
openframe cluster list

# Check cluster status
openframe cluster status my-cluster

# Check what's running in Kubernetes
kubectl get pods -A
```

Expected cluster list output:

```text
NAME          STATE    AGENTS   SERVER
my-cluster    running  1        1
```

---

## Build from Source (Alternative)

If you prefer to build from source instead of using a binary release:

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the binary
go build -o openframe ./main.go

# Run it
./openframe --version
```

---

## CI/CD Non-Interactive Mode

For automated pipelines, use the `--non-interactive` flag:

```bash
openframe bootstrap my-env \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose
```

The `CI` environment variable also automatically enables non-interactive mode when set.

---

## Available Bootstrap Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--deployment-mode` | One of `oss-tenant`, `saas-tenant`, `saas-shared` | Interactive prompt |
| `--non-interactive` | Skip all interactive prompts | `false` |
| `-v`, `--verbose` | Show detailed output including ArgoCD sync | `false` |
| `--silent` | Suppress all output except errors | `false` |

---

## Next Steps

After completing the quick start:

- Follow the **[First Steps Guide](first-steps.md)** to explore key features and run your first development workflow
- Review the **[Prerequisites Guide](prerequisites.md)** if you encountered any tool-missing errors
