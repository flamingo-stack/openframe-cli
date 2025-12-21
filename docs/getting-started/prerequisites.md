# Prerequisites

Before installing and using OpenFrame CLI, ensure your system meets the following requirements. This guide will help you verify that all necessary dependencies are properly installed and configured.

## System Requirements

| Component | Minimum Version | Recommended | Platform Support |
|-----------|----------------|-------------|------------------|
| **Operating System** | Linux, macOS, Windows | Linux/macOS | All major platforms |
| **Memory (RAM)** | 4GB | 8GB+ | Required for K3d clusters |
| **Disk Space** | 2GB free | 5GB+ | For Docker images and data |
| **Network** | Internet access | Stable connection | For pulling images and charts |

## Required Software Dependencies

### 1. Go Programming Language

OpenFrame CLI is built with Go and requires Go 1.23.0 or later.

| Software | Version Required | Installation | Verification Command |
|----------|------------------|--------------|---------------------|
| **Go** | 1.23.0+ | [Download from golang.org](https://golang.org/download/) | `go version` |

**Verification:**
```bash
go version
# Expected output: go version go1.23.x linux/amd64 (or your platform)
```

### 2. Docker Engine

Required for K3d clusters and container operations.

| Software | Version Required | Installation | Verification Command |
|----------|------------------|--------------|---------------------|
| **Docker** | 20.10+ | [Docker Desktop](https://www.docker.com/products/docker-desktop) or [Docker Engine](https://docs.docker.com/engine/install/) | `docker version` |

**Installation Options:**
- **macOS/Windows**: Docker Desktop (includes Docker Engine + GUI)
- **Linux**: Docker Engine (CLI only, lighter weight)

**Verification:**
```bash
docker version
# Expected: Client and Server versions displayed

docker run hello-world
# Expected: "Hello from Docker!" message
```

### 3. Kubernetes Tools (Auto-installed)

These tools are automatically installed by OpenFrame CLI when needed:

| Tool | Purpose | Auto-Install | Manual Install |
|------|---------|-------------|----------------|
| **K3d** | Lightweight Kubernetes | ✅ Yes | [k3d.io](https://k3d.io) |
| **Helm** | Package manager | ✅ Yes | [helm.sh](https://helm.sh/docs/intro/install/) |
| **kubectl** | Kubernetes CLI | ✅ Yes | [kubernetes.io](https://kubernetes.io/docs/tasks/tools/) |

> **Note**: OpenFrame CLI includes prerequisite checks and will install missing tools automatically when possible.

## Optional Development Tools

These tools enhance the development experience but are not required for basic usage:

| Tool | Purpose | Installation | When Needed |
|------|---------|--------------|-------------|
| **Skaffold** | Live reloading | [skaffold.dev](https://skaffold.dev/docs/install/) | `openframe dev` commands |
| **Telepresence** | Traffic interception | [telepresence.io](https://www.telepresence.io/docs/latest/install/) | Traffic debugging |
| **ArgoCD CLI** | GitOps management | [argoproj.github.io](https://argo-cd.readthedocs.io/en/stable/cli_installation/) | Advanced GitOps workflows |

## Environment Variables

The following environment variables can be set to customize OpenFrame CLI behavior:

| Variable | Purpose | Default | Example |
|----------|---------|---------|---------|
| `OPENFRAME_CONFIG_DIR` | Configuration directory | `$HOME/.openframe` | `/opt/openframe/config` |
| `DOCKER_HOST` | Docker daemon address | System default | `unix:///var/run/docker.sock` |
| `KUBECONFIG` | Kubernetes config file | `$HOME/.kube/config` | `/path/to/kubeconfig` |

**Setting Environment Variables:**

```bash
# Bash/Zsh
export OPENFRAME_CONFIG_DIR="$HOME/.openframe"
echo 'export OPENFRAME_CONFIG_DIR="$HOME/.openframe"' >> ~/.bashrc

# Fish
set -gx OPENFRAME_CONFIG_DIR "$HOME/.openframe"
echo 'set -gx OPENFRAME_CONFIG_DIR "$HOME/.openframe"' >> ~/.config/fish/config.fish
```

## Account and Access Requirements

### GitHub Access (Optional)

For GitOps workflows and private chart repositories:

- **Public repositories**: No authentication required
- **Private repositories**: Personal Access Token (PAT) with repo permissions

### Container Registry Access (Optional)

For private container images:

- **Public images**: No authentication required  
- **Private registries**: Registry credentials configured with Docker

## Network Requirements

| Requirement | Purpose | Ports |
|-------------|---------|-------|
| **Internet Access** | Download images, charts, tools | 80, 443 |
| **Docker Registry Access** | Pull container images | 443 |
| **GitHub API Access** | Clone repositories, download releases | 443 |
| **Local Port Availability** | K3d cluster and services | 6443, 8080-8090 |

## Verification Checklist

Run these commands to verify your system is ready:

### Basic System Check

```bash
# Check Go installation
go version
# ✅ Should show Go 1.23.0+

# Check Docker
docker --version
docker ps
# ✅ Should show Docker version and running containers (if any)

# Check available disk space
df -h
# ✅ Should show at least 2GB free space

# Check memory
free -h  # Linux
vm_stat  # macOS
# ✅ Should show at least 4GB total RAM
```

### Network Connectivity Check

```bash
# Test internet connectivity
curl -I https://github.com
# ✅ Should return HTTP 200

# Test Docker Hub access
docker pull hello-world
# ✅ Should download successfully

# Check available ports (Linux/macOS)
netstat -tuln | grep -E "(6443|8080)"
# ✅ Should show no conflicts on these ports
```

### Directory Permissions

```bash
# Ensure Docker daemon is accessible
docker info
# ✅ Should show Docker system information without permission errors

# Check home directory write access
touch $HOME/.openframe-test && rm $HOME/.openframe-test
# ✅ Should create and delete file without errors
```

## Troubleshooting Common Issues

### Docker Permission Issues (Linux)

```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify docker access
docker ps
```

### Port Conflicts

```bash
# Find processes using required ports
sudo lsof -i :6443
sudo lsof -i :8080

# Kill conflicting processes if safe to do so
sudo kill -9 <PID>
```

### Go Installation Issues

```bash
# Verify GOPATH and GOROOT
go env GOPATH
go env GOROOT

# Update PATH if needed
export PATH=$PATH:/usr/local/go/bin
```

## Ready to Continue?

Once you've verified all prerequisites are met, you're ready to proceed:

- **Next**: [Quick Start Guide](./quick-start.md) for rapid installation and first cluster
- **Alternative**: [Local Development Setup](../development/setup/local-development.md) for comprehensive development environment

## Need Help?

If you encounter issues during prerequisite setup:

1. **Check the troubleshooting section above**
2. **Review official tool documentation** for detailed installation guides
3. **Open an issue** on GitHub with your system details and error messages

> **💡 Pro Tip**: The `openframe` CLI includes built-in prerequisite checking. Running any command will validate your setup and provide helpful error messages for missing dependencies.