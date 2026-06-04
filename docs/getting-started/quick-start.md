# Quick Start

Get a fully operational OpenFrame Kubernetes environment running in under 5 minutes.

[![Getting Started with OpenFrame](https://img.youtube.com/vi/-_56_qYvMWk/maxresdefault.jpg)](https://www.youtube.com/watch?v=-_56_qYvMWk)

---

## TL;DR — One-Command Bootstrap

If you already have Docker, kubectl, k3d, Helm, Git, and mkcert installed, this is all you need:

```bash
openframe bootstrap
```

The interactive wizard will guide you through the rest. For a fully automated, non-interactive setup:

```bash
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive
```

---

## Step 1: Install the OpenFrame CLI

### Linux / macOS

Download the latest release for your platform from the [GitHub Releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest) and place the binary on your `$PATH`:

```bash
# Example for Linux amd64
curl -Lo openframe https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64
chmod +x openframe
sudo mv openframe /usr/local/bin/openframe
```

```bash
# Example for macOS arm64 (Apple Silicon)
curl -Lo openframe https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64
chmod +x openframe
sudo mv openframe /usr/local/bin/openframe
```

### Windows (AMD64)

1. Download [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip)
2. Extract the archive
3. Move `openframe.exe` to a directory on your `$PATH` (or run from its extracted location)

### Build from Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe main.go
sudo mv openframe /usr/local/bin/openframe
```

---

## Step 2: Verify the Installation

```bash
openframe --version
```

Expected output:

```text
dev (none) built on unknown
```

(Release builds will show the actual version number.)

```bash
openframe --help
```

You should see the OpenFrame ASCII logo and the list of available commands.

---

## Step 3: Bootstrap Your First Environment

### Option A — Interactive Mode (Recommended for First-Time Users)

```bash
openframe bootstrap
```

The wizard will ask you to:

1. **Enter a cluster name** (or accept the default)
2. **Select a deployment mode** (`oss-tenant` is the default self-hosted option)
3. **Choose default or custom configuration** (default is recommended to start)

The bootstrap process will:

- Check and install prerequisites (Docker, kubectl, k3d, Helm, Git, mkcert)
- Create a K3D Kubernetes cluster
- Install ArgoCD via Helm
- Clone your app-of-apps repository
- Install the OpenFrame app-of-apps chart
- Wait for all ArgoCD applications to reach `Healthy + Synced`

### Option B — Non-Interactive / CI Mode

```bash
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive
```

This bypasses all prompts and uses defaults for every option. Suitable for CI/CD pipelines.

---

## Step 4: Verify the Environment

After bootstrap completes, check your cluster status:

```bash
openframe cluster list
```

```bash
openframe cluster status my-cluster --detailed
```

Expected output shows your cluster nodes as `Ready` and all ArgoCD applications as `Healthy`.

---

## Expected Output

A successful bootstrap looks like this:

```text
  ____                 _____ _
 / __ \___  ___ ____  / ___/| |
/ /_/ / _ \/ -_) _ \/ /__  |_|
\____/ .__/\__/_//_/\___/  (_)
    /_/  CLI v1.x.x

✔ Checking prerequisites...
✔ Docker: running
✔ kubectl: installed
✔ k3d: installed
✔ Helm: installed
✔ Creating cluster: my-cluster
✔ Cluster created successfully
✔ Installing ArgoCD...
✔ ArgoCD deployed
✔ Cloning app-of-apps repository...
✔ Installing app-of-apps chart...
✔ Waiting for applications to sync...
✔ All applications: Healthy + Synced

Environment ready! 🚀
```

---

## Quick Cluster Commands

Once your cluster is running, here are the most useful commands:

```bash
# List all clusters
openframe cluster list

# Check cluster status with app details
openframe cluster status my-cluster --detailed

# Install charts on an existing cluster
openframe chart install my-cluster --deployment-mode=oss-tenant

# Delete a cluster when done
openframe cluster delete my-cluster
```

---

## Next Steps

After your first successful bootstrap:

- Follow the [First Steps Guide](first-steps.md) to explore key features
- Review [Prerequisites](prerequisites.md) to understand all system requirements
- Join the [OpenMSP Community Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) for support and updates
