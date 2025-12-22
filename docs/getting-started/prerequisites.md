# Prerequisites

Before you can use OpenFrame CLI, you'll need to ensure your system meets the requirements and has the necessary tools installed. OpenFrame CLI automates much of the prerequisite installation, but understanding these requirements will help you troubleshoot any issues.

## System Requirements

### Operating System Support

| Platform | Status | Versions |
|----------|--------|----------|
| **macOS** | ✅ Fully Supported | macOS 10.15+ (Catalina and newer) |
| **Linux** | ✅ Fully Supported | Ubuntu 18.04+, CentOS 7+, Debian 10+ |
| **Windows** | ✅ Fully Supported | Windows 10/11 with WSL2 |

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **RAM** | 4 GB | 8 GB+ |
| **CPU** | 2 cores | 4 cores+ |
| **Disk Space** | 10 GB free | 20 GB+ free |
| **Network** | Internet connection | Stable broadband |

> **Note**: Kubernetes clusters (even lightweight ones like K3d) are resource-intensive. The recommended specifications will provide a much better experience.

## Required Software

OpenFrame CLI will automatically install missing prerequisites when possible, but you can install them manually for better control.

### Core Dependencies

| Tool | Purpose | Auto-Install | Manual Install |
|------|---------|--------------|----------------|
| **Docker** | Container runtime for K3d | ❌ | [Download Docker](https://docs.docker.com/get-docker/) |
| **k3d** | Lightweight Kubernetes | ✅ | `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh \| bash` |
| **kubectl** | Kubernetes CLI | ✅ | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | Package manager | ✅ | [Install Helm](https://helm.sh/docs/intro/install/) |

### Development Tools (Optional)

| Tool | Purpose | Auto-Install | When Needed |
|------|---------|--------------|-------------|
| **Telepresence** | Traffic intercepts | ✅ | Using `openframe dev intercept` |
| **Skaffold** | Live development | ✅ | Using `openframe dev scaffold` |
| **jq** | JSON processing | ✅ | Various CLI operations |
| **Git** | Version control | ❌ | GitOps workflows |

### Docker Installation

Docker is the only prerequisite that **must** be installed manually, as it requires system-level access.

#### macOS
```bash
# Using Homebrew
brew install --cask docker

# Or download Docker Desktop from https://docker.com
```

#### Linux (Ubuntu/Debian)
```bash
# Install Docker Engine
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add your user to docker group
sudo usermod -aG docker $USER
# Log out and back in for group changes to take effect

# Start Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

#### Windows
1. Install Docker Desktop from [docker.com](https://docker.com)
2. Enable WSL2 integration
3. Restart Windows after installation

## Account Requirements

### Git Provider Access

For GitOps functionality, you'll need access to a Git repository:

| Provider | Requirements | Purpose |
|----------|-------------|----------|
| **GitHub** | Personal access token with repo permissions | Storing ArgoCD applications |
| **GitLab** | Deploy key or access token | Self-hosted or GitLab.com |
| **Bitbucket** | App password or SSH key | Repository access |
| **Generic Git** | SSH key or HTTPS credentials | Any Git server |

### Container Registry (Optional)

For custom applications:

| Registry | Access Type | Use Case |
|----------|-------------|----------|
| **Docker Hub** | Username/password or token | Public images |
| **GitHub Container Registry** | Personal access token | Private images |
| **Amazon ECR** | AWS credentials | AWS-based workflows |
| **Google Container Registry** | Service account | GCP-based workflows |

## Environment Variables

OpenFrame CLI respects standard environment variables:

### Required (Auto-configured)
```bash
# These are set automatically by OpenFrame CLI
export KUBECONFIG="$HOME/.kube/config"  # Kubernetes config
export PATH="$PATH:$HOME/.local/bin"     # Local tools
```

### Optional Configuration
```bash
# Docker configuration
export DOCKER_HOST="unix:///var/run/docker.sock"

# Proxy settings (if behind corporate firewall)
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"
export NO_PROXY="localhost,127.0.0.1,.local"

# Git configuration for ArgoCD
export GIT_USERNAME="your-username"
export GIT_TOKEN="your-access-token"
```

## Verification Commands

Before using OpenFrame CLI, verify your system is ready:

### Check Docker
```bash
# Verify Docker is running
docker --version
docker ps

# Expected output:
# Docker version 24.0.0+
# CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES
```

### Check System Resources
```bash
# Check available memory (should be 4GB+)
free -h

# Check disk space (should have 10GB+ free)
df -h

# Check CPU cores (should be 2+)
nproc
```

### Check Network Connectivity
```bash
# Verify internet access
ping -c 3 google.com

# Test Docker Hub connectivity
docker pull hello-world

# Test GitHub connectivity (for GitOps)
curl -s https://api.github.com/zen
```

## Common Issues and Solutions

### Docker Not Running
```bash
# Error: Cannot connect to the Docker daemon
# Solution: Start Docker service

# Linux
sudo systemctl start docker

# macOS
open /Applications/Docker.app

# Windows
# Start Docker Desktop from Start menu
```

### Insufficient Resources
```bash
# Error: Cluster creation fails with resource errors
# Solution: Free up system resources

# Check running containers
docker ps
docker system prune  # Remove unused containers/images

# Check system load
htop  # or top on macOS
```

### Network/Proxy Issues
```bash
# Error: Cannot download prerequisites
# Solution: Configure proxy settings

# Set proxy for current session
export HTTP_PROXY="http://proxy.company.com:8080"
export HTTPS_PROXY="http://proxy.company.com:8080"

# Configure Docker proxy (create ~/.docker/config.json)
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy.company.com:8080",
      "httpsProxy": "http://proxy.company.com:8080"
    }
  }
}
```

### Permission Issues
```bash
# Error: Permission denied when accessing Docker
# Solution: Add user to docker group

sudo usermod -aG docker $USER
# Log out and back in
```

## Quick Verification Script

Run this script to check all prerequisites:

```bash
#!/bin/bash

echo "🔍 Checking OpenFrame CLI prerequisites..."

# Check Docker
if command -v docker &> /dev/null && docker ps &> /dev/null; then
    echo "✅ Docker: Running"
else
    echo "❌ Docker: Not running or not installed"
fi

# Check system resources
MEMORY_GB=$(free -g | awk '/^Mem:/{print $2}')
if [ "$MEMORY_GB" -ge 4 ]; then
    echo "✅ Memory: ${MEMORY_GB}GB (sufficient)"
else
    echo "⚠️ Memory: ${MEMORY_GB}GB (recommended: 4GB+)"
fi

# Check disk space
DISK_GB=$(df -BG . | awk 'NR==2 {print $4}' | sed 's/G//')
if [ "$DISK_GB" -ge 10 ]; then
    echo "✅ Disk Space: ${DISK_GB}GB free"
else
    echo "⚠️ Disk Space: ${DISK_GB}GB free (recommended: 10GB+)"
fi

# Check internet connectivity
if ping -c 1 google.com &> /dev/null; then
    echo "✅ Network: Connected"
else
    echo "❌ Network: No internet connection"
fi

echo ""
echo "🚀 Ready to install OpenFrame CLI? Run: curl -sSL https://install.openframe.dev | bash"
```

## Next Steps

Once your prerequisites are in place:

1. **[Quick Start](quick-start.md)** - Install OpenFrame CLI and create your first cluster
2. **[First Steps](first-steps.md)** - Explore the key features after installation

> **Tip**: OpenFrame CLI will check and install most prerequisites automatically. If you encounter issues, refer back to this guide for troubleshooting steps.

---

**Previous**: [Introduction](introduction.md) | **Next**: [Quick Start](quick-start.md)