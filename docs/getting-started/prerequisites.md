# Prerequisites Guide

Before installing OpenFrame CLI, ensure your system meets the requirements and has the necessary dependencies installed.

## System Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **Operating System** | Linux, macOS, Windows (WSL2) | Linux/macOS | Windows requires WSL2 for Docker |
| **CPU** | 2 cores | 4+ cores | K3d clusters benefit from more cores |
| **Memory** | 4GB RAM | 8GB+ RAM | Kubernetes workloads are memory-intensive |
| **Disk Space** | 10GB free | 20GB+ free | Container images and cluster data |
| **Network** | Internet access | Stable broadband | For downloading images and charts |

## Required Software Dependencies

### Core Dependencies

These tools are **required** for OpenFrame CLI to function:

| Tool | Version | Installation | Verification |
|------|---------|-------------|---------------|
| **Docker** | 20.10+ | [docker.com](https://docker.com) | `docker --version` |
| **K3d** | 5.4+ | [k3d.io](https://k3d.io) | `k3d --version` |
| **kubectl** | 1.24+ | [kubernetes.io](https://kubernetes.io/docs/tasks/tools/) | `kubectl version --client` |
| **Helm** | 3.8+ | [helm.sh](https://helm.sh) | `helm version` |

### Development Dependencies (Optional)

These tools enable advanced development workflows but are not required for basic usage:

| Tool | Purpose | Installation | Verification |
|------|---------|-------------|---------------|
| **Telepresence** | Local service debugging | [telepresence.io](https://telepresence.io) | `telepresence version` |
| **Skaffold** | Live reload workflows | [skaffold.dev](https://skaffold.dev) | `skaffold version` |

## Installation Commands

### macOS (Homebrew)

```bash
# Install core dependencies
brew install docker k3d kubectl helm

# Start Docker Desktop
open -a Docker

# Verify Docker is running
docker ps

# Install optional development tools
brew install datawire/blackbird/telepresence
brew install skaffold
```

### Linux (Ubuntu/Debian)

```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install K3d
wget -q -O - https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm

# Restart shell or logout/login to activate docker group
newgrp docker
```

### Windows (WSL2)

> **Note**: OpenFrame CLI requires WSL2 on Windows. Install WSL2 and Ubuntu first.

```bash
# Inside WSL2 Ubuntu, follow the Linux instructions above
# Ensure Docker Desktop for Windows has WSL2 backend enabled
```

## Environment Variables

OpenFrame CLI uses these environment variables for configuration:

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `KUBECONFIG` | No | `~/.kube/config` | Kubernetes configuration file path |
| `OPENFRAME_LOG_LEVEL` | No | `info` | Logging level: debug, info, warn, error |
| `DOCKER_HOST` | No | Auto-detect | Docker daemon connection |

### Setting Environment Variables

```bash
# Optional: Set custom kubeconfig path
export KUBECONFIG="$HOME/.kube/openframe-config"

# Optional: Enable debug logging
export OPENFRAME_LOG_LEVEL=debug

# Add to ~/.bashrc or ~/.zshrc to persist
echo 'export OPENFRAME_LOG_LEVEL=debug' >> ~/.bashrc
```

## Account & Access Requirements

### Container Registry Access

OpenFrame CLI pulls public container images. No authentication required for basic usage.

For private registries or enterprise features:
- **Docker Hub**: Personal account for higher rate limits
- **Private Registry**: Credentials configured in Docker/Kubernetes

### Git Repository Access

If using custom charts or GitOps repositories:
- **SSH Keys**: Configure for private Git repositories
- **Git Credentials**: Set up for HTTPS authentication

## Verification Commands

Run these commands to verify your system is ready:

```bash
# Check Docker is running
docker ps
# Expected: Table of running containers (may be empty)

# Check K3d is installed
k3d version
# Expected: Version output for k3d, k3s, and containerd

# Check kubectl can connect (after cluster creation)
kubectl version --client
# Expected: Client version information

# Check Helm is functional
helm version
# Expected: Version information for Helm

# Test K3d cluster creation (cleanup test)
k3d cluster create test-cluster
k3d cluster delete test-cluster
# Expected: Successful creation and deletion messages
```

## Troubleshooting Common Issues

### Docker Issues

```bash
# Docker daemon not running
sudo systemctl start docker  # Linux
# or restart Docker Desktop   # macOS/Windows

# Permission denied
sudo usermod -aG docker $USER
newgrp docker  # or logout/login
```

### K3d Issues

```bash
# Port conflicts
k3d cluster create --port 8080:80@loadbalancer  # Use custom ports

# Insufficient resources
docker system prune  # Clean up unused containers
```

### Network Issues

```bash
# Test internet connectivity
curl -I https://registry-1.docker.io
# Expected: HTTP 200 response

# Check DNS resolution
nslookup docker.io
# Expected: IP address resolution
```

## Next Steps

Once all prerequisites are installed and verified:

1. **Continue to [Quick Start](./quick-start.md)** to install and run OpenFrame CLI
2. **Bookmark this page** for troubleshooting future issues
3. **Join the community** for help with specific setup challenges

> **💡 Tip**: Run the verification commands periodically to ensure your development environment stays healthy.