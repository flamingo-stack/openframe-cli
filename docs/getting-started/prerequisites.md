# Prerequisites Guide

Before installing and using OpenFrame CLI, ensure your system meets the requirements and has the necessary tools installed. OpenFrame CLI includes automatic prerequisite checking and can install missing tools when needed.

## System Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **Operating System** | Linux, macOS, Windows | Linux/macOS | Windows support via WSL2 |
| **Memory** | 4 GB RAM | 8 GB RAM | For K3D cluster overhead |
| **Disk Space** | 5 GB free | 10 GB free | Container images and cluster data |
| **Network** | Internet connection | Stable broadband | For image pulls and Git operations |
| **Architecture** | x86_64, ARM64 | x86_64 | ARM64 support available |

## Required Software

The following tools are **required** for OpenFrame CLI to function properly:

### Docker
```bash
# Check if Docker is installed and running
docker --version
docker info

# Expected output:
# Docker version 24.0.0 or later
# Server running status
```

**Installation**: [Get Docker](https://docs.docker.com/get-docker/)
**Purpose**: Container runtime for K3D clusters and application deployment

### kubectl
```bash
# Verify kubectl installation
kubectl version --client

# Expected output:
# Client Version: v1.28.0 or later
```

**Installation**: [Install kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
**Purpose**: Kubernetes command-line interface for cluster interaction

### K3D
```bash
# Check K3D installation
k3d version

# Expected output:
# k3d version v5.6.0 or later
```

**Installation**: [Install K3D](https://k3d.io/v5.6.0/#installation)
**Purpose**: Lightweight Kubernetes distribution for local development

### Helm
```bash
# Verify Helm installation
helm version

# Expected output:
# version.BuildInfo{Version:"v3.12.0" or later}
```

**Installation**: [Install Helm](https://helm.sh/docs/intro/install/)
**Purpose**: Kubernetes package manager for chart installation

## Optional Tools

These tools provide enhanced functionality but are **not required** for basic operations:

### Telepresence (for Development Tools)
```bash
# Check Telepresence installation (optional)
telepresence version

# Expected output:
# Telepresence 2.15.0 or later
```

**Installation**: [Install Telepresence](https://www.telepresence.io/docs/latest/install/)
**Purpose**: Traffic interception for local development against remote clusters

### jq (for Development Tools)
```bash
# Verify jq installation (optional)
jq --version

# Expected output:
# jq-1.6 or later
```

**Installation**: [Install jq](https://jqlang.github.io/jq/download/)
**Purpose**: JSON processing for scaffold operations and debugging

### Git
```bash
# Check Git installation
git --version

# Expected output:
# git version 2.30.0 or later
```

**Installation**: [Install Git](https://git-scm.com/downloads)
**Purpose**: Required for GitOps workflows and chart repositories

## Environment Variables

Configure these environment variables for optimal operation:

### Required
```bash
# Docker daemon socket (usually auto-detected)
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration directory
export KUBECONFIG=$HOME/.kube/config
```

### Optional
```bash
# Custom helm repository cache
export HELM_CACHE_HOME=$HOME/.cache/helm

# Telepresence configuration (if using dev tools)
export TELEPRESENCE_LOGIN_DOMAIN=your-domain.com

# K3D cluster configuration
export K3D_FIX_DNS=1  # For DNS resolution on some systems
```

## Verification Commands

Run these commands to verify your system is ready for OpenFrame CLI:

### Complete System Check
```bash
# Verify all tools are available and functional
docker info > /dev/null && echo "✅ Docker: OK" || echo "❌ Docker: Not available"
kubectl version --client > /dev/null && echo "✅ kubectl: OK" || echo "❌ kubectl: Not available"
k3d version > /dev/null && echo "✅ K3D: OK" || echo "❌ K3D: Not available"
helm version > /dev/null && echo "✅ Helm: OK" || echo "❌ Helm: Not available"
```

### Docker Validation
```bash
# Test Docker functionality
docker run --rm hello-world

# Should output: "Hello from Docker!"
```

### Network Connectivity
```bash
# Test access to required registries
docker pull nginx:alpine
docker pull rancher/k3s:latest

# Should complete without errors
```

## Account Requirements

### Container Registries

| Registry | Purpose | Access Required |
|----------|---------|-----------------|
| **Docker Hub** | Base container images | Public access (no auth needed) |
| **Rancher Registry** | K3S images | Public access (no auth needed) |
| **GitHub Container Registry** | OpenFrame images | Public access for OSS tenant |

### GitHub (Optional)
- **Personal Access Token**: Required for private chart repositories
- **SSH Keys**: For Git operations if using SSH URLs
- **Repository Access**: Read access to chart repositories

## Troubleshooting

### Common Issues

**Docker not running**:
```bash
# Start Docker daemon
sudo systemctl start docker  # Linux
open /Applications/Docker.app  # macOS
```

**kubectl not configured**:
```bash
# Initialize kubectl config
mkdir -p $HOME/.kube
touch $HOME/.kube/config
```

**K3D permission issues**:
```bash
# Add user to docker group (Linux)
sudo usermod -aG docker $USER
# Log out and back in
```

**Memory constraints**:
```bash
# Check available memory
free -h  # Linux
vm_stat  # macOS

# For low memory systems, adjust K3D settings
export K3D_MEMORY_LIMIT=2g
```

### Getting Help

If you encounter issues during setup:

1. **Built-in validation**: OpenFrame CLI will check prerequisites and provide specific error messages
2. **Verbose output**: Use `--verbose` flag to see detailed installation attempts
3. **Manual installation**: Install tools manually using official documentation links above

---

> **Next Step**: System ready? Continue to [Quick Start](quick-start.md) to create your first OpenFrame cluster in under 5 minutes.