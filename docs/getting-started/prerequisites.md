# Prerequisites

Before using OpenFrame CLI, ensure your system meets the minimum requirements and has the necessary tools installed. OpenFrame CLI includes automatic detection and installation capabilities for most dependencies.

## System Requirements

### Minimum Hardware Requirements
| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **RAM** | 24GB | 32GB | Required for Kubernetes clusters and ArgoCD |
| **CPU** | 6 cores | 12 cores | Multi-core needed for container orchestration |
| **Disk Space** | 50GB | 100GB | Container images and cluster data |
| **Network** | Stable internet | High-speed broadband | For downloading images and charts |

### Operating System Support

| Platform | Version | Architecture | Status |
|----------|---------|--------------|---------|
| **Linux** | Ubuntu 20.04+ | x86_64, ARM64 | ✅ Fully Supported |
| **Linux** | CentOS/RHEL 8+ | x86_64, ARM64 | ✅ Fully Supported |
| **macOS** | 10.15+ (Catalina) | x86_64, Apple Silicon | ✅ Fully Supported |
| **Windows** | Windows 10+ | x86_64 | ✅ Supported via WSL2 |

> **Windows Note**: Windows users must have WSL2 (Windows Subsystem for Linux) installed and configured. OpenFrame CLI runs within the Linux subsystem.

## Required Software Dependencies

OpenFrame CLI automatically detects and can install the following tools. You can also install them manually:

### Core Dependencies

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Docker** | 20.10+ | Container runtime for K3d clusters | ✅ Yes |
| **K3d** | 5.0+ | Lightweight Kubernetes distribution | ✅ Yes |
| **Helm** | 3.8+ | Kubernetes package manager | ✅ Yes |
| **kubectl** | 1.25+ | Kubernetes CLI tool | ✅ Yes |

### Development Tools (Optional)

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Local development intercepts | ✅ Yes |
| **jq** | 1.6+ | JSON processing for scripts | ✅ Yes |
| **Git** | 2.30+ | Version control for GitOps | ❌ Manual |

## Manual Installation Guide

If you prefer to install dependencies manually, follow these platform-specific guides:

### Linux (Ubuntu/Debian)

```bash
# Update package manager
sudo apt-get update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install K3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl

# Install additional tools
sudo apt-get install -y jq git
```

### macOS

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install required tools
brew install docker k3d helm kubectl jq git

# Start Docker Desktop
open /Applications/Docker.app
```

### Windows (WSL2)

```bash
# Inside WSL2 terminal
# Install Docker (use Docker Desktop for Windows)
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install other tools
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/kubectl
```

## Environment Variables

Configure these environment variables for optimal experience:

### Required Variables

```bash
# Docker daemon socket (usually automatic)
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration (set by kubectl)
export KUBECONFIG=$HOME/.kube/config
```

### Optional Variables

```bash
# OpenFrame CLI configuration
export OPENFRAME_LOG_LEVEL=info
export OPENFRAME_CLUSTER_NAME=my-cluster

# Development tools
export TELEPRESENCE_LOG_LEVEL=info
```

Add these to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
echo 'export OPENFRAME_LOG_LEVEL=info' >> ~/.bashrc
source ~/.bashrc
```

## Account and Access Requirements

### Git Repository Access

For GitOps workflows, ensure you have:

- **Git credentials** configured for repository access
- **SSH keys** set up for private repositories
- **Personal Access Tokens** for GitHub/GitLab integrations

```bash
# Configure Git credentials
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Test Git access
git clone https://github.com/your-org/your-repo.git
```

### Container Registry Access

If using private registries:

```bash
# Docker Hub login
docker login

# GitHub Container Registry
docker login ghcr.io -u your-username

# Custom registry
docker login your-registry.com
```

## Verification Commands

Run these commands to verify your system is ready:

### Basic System Check

```bash
# Check system resources
free -h  # Memory
df -h    # Disk space
nproc    # CPU cores
```

### Tool Verification

```bash
# Verify Docker
docker --version
docker run hello-world

# Verify K3d
k3d version

# Verify Helm
helm version

# Verify kubectl
kubectl version --client

# Verify optional tools
telepresence version 2>/dev/null || echo "Telepresence not installed"
jq --version
git --version
```

### Network Connectivity

```bash
# Test internet connectivity
curl -I https://github.com
curl -I https://registry-1.docker.io

# Test DNS resolution
nslookup github.com
nslookup registry-1.docker.io
```

## Automatic Prerequisite Check

OpenFrame CLI includes built-in prerequisite checking:

```bash
# Download OpenFrame CLI (see Quick Start guide)
# The CLI will automatically check prerequisites on first run
openframe bootstrap

# Manual prerequisite check
openframe cluster create --check-only
```

The CLI will:
1. ✅ Detect missing tools
2. 🔧 Offer to install them automatically  
3. ⚠️ Warn about system resource constraints
4. 🚫 Prevent execution if critical requirements are missing

## Troubleshooting Common Issues

### Docker Permission Denied

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Restart terminal or run
newgrp docker
```

### K3d Port Conflicts

```bash
# Check for port conflicts
sudo netstat -tlpn | grep :6443
sudo netstat -tlpn | grep :80

# Kill conflicting processes if safe
sudo pkill -f :6443
```

### Memory Insufficient

```bash
# Check available memory
free -h

# Close unnecessary applications
# Consider upgrading RAM to 32GB for optimal performance
```

### WSL2 Docker Integration

```bash
# Ensure Docker Desktop is running
# Enable WSL2 integration in Docker Desktop settings
# Restart WSL2: wsl --shutdown, then reopen terminal
```

## Next Steps

Once your system meets all prerequisites:

1. **[Quick Start](quick-start.md)** - Get OpenFrame running in 5 minutes
2. **[First Steps](first-steps.md)** - Explore OpenFrame features
3. **[Development Setup](../development/setup/environment.md)** - Advanced configuration

---

**Need Help?** If you encounter issues during prerequisite setup:
- Check the specific tool's documentation
- Run OpenFrame CLI with `--verbose` flag for detailed diagnostics
- Consult the [troubleshooting section](../development/troubleshooting/common-issues.md)