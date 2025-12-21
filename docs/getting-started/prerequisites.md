# Prerequisites

Before using OpenFrame CLI, ensure your system has the required tools and meets the system requirements. OpenFrame CLI includes automatic prerequisite checking, but this guide helps you prepare your environment.

## Required Software

### Core Requirements

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Docker** | 20.10+ | Container runtime for K3d clusters | [docker.com/get-started](https://docs.docker.com/get-started/) |
| **K3d** | 5.4+ | Lightweight Kubernetes in Docker | [k3d.io](https://k3d.io/v5.4.1/#installation) |
| **kubectl** | 1.25+ | Kubernetes cluster interaction | [kubernetes.io/docs/tasks/tools/](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.8+ | Chart installation and management | [helm.sh/docs/intro/install/](https://helm.sh/docs/intro/install/) |

### Development Tools (Optional)

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Traffic interception for local dev | [telepresence.io](https://www.telepresence.io/docs/latest/install/) |
| **Skaffold** | 2.0+ | Live development workflows | [skaffold.dev](https://skaffold.dev/docs/install/) |

## System Requirements

### Minimum Requirements
- **OS**: Linux (Ubuntu 20.04+), macOS (11+), Windows 10+ with WSL2
- **RAM**: 4GB available (8GB+ recommended)
- **CPU**: 2 cores (4+ recommended)
- **Disk**: 10GB free space for Docker images and clusters
- **Network**: Internet access for downloading images and charts

### Recommended Requirements
- **RAM**: 8GB+ for multiple clusters and development workflows
- **CPU**: 4+ cores for better performance
- **Disk**: 20GB+ free space with SSD storage
- **Network**: Stable broadband connection

## Account and Access Requirements

### Docker Hub Access
```bash
# Optional: Login for higher rate limits
docker login
```

### Git Access (for GitOps workflows)
- Git client configured with user credentials
- SSH key or token access to Git repositories (if using private repos)

## Environment Variables

OpenFrame CLI supports these optional environment variables:

```bash
# Customize cluster configuration
export OPENFRAME_CLUSTER_PREFIX="myteam"
export OPENFRAME_DEFAULT_DEPLOYMENT_MODE="oss-tenant"

# Docker configuration
export DOCKER_HOST="unix:///var/run/docker.sock"

# Development tools
export TELEPRESENCE_LOGIN_DOMAIN="your-domain.com"
```

## Installation Verification

### Quick Verification Script

Create a script to verify all prerequisites:

```bash
#!/bin/bash
echo "🔍 Checking OpenFrame CLI Prerequisites..."

# Check Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    echo "✅ Docker: $DOCKER_VERSION"
    
    # Check if Docker daemon is running
    if docker info &> /dev/null; then
        echo "✅ Docker daemon is running"
    else
        echo "❌ Docker daemon is not running - start Docker"
        exit 1
    fi
else
    echo "❌ Docker not found"
    exit 1
fi

# Check K3d
if command -v k3d &> /dev/null; then
    K3D_VERSION=$(k3d --version | head -1 | cut -d' ' -f3)
    echo "✅ K3d: $K3D_VERSION"
else
    echo "❌ K3d not found"
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    KUBECTL_VERSION=$(kubectl version --client --short 2>/dev/null | cut -d' ' -f3)
    echo "✅ kubectl: $KUBECTL_VERSION"
else
    echo "❌ kubectl not found"
fi

# Check Helm
if command -v helm &> /dev/null; then
    HELM_VERSION=$(helm version --short | cut -d' ' -f1)
    echo "✅ Helm: $HELM_VERSION"
else
    echo "❌ Helm not found"
fi

# Optional tools
echo ""
echo "📦 Optional Development Tools:"

if command -v telepresence &> /dev/null; then
    TEL_VERSION=$(telepresence version 2>/dev/null | grep "Client" | cut -d' ' -f2)
    echo "✅ Telepresence: $TEL_VERSION"
else
    echo "⚠️ Telepresence not found (optional for dev workflows)"
fi

if command -v skaffold &> /dev/null; then
    SKAF_VERSION=$(skaffold version --output=json 2>/dev/null | grep '"version"' | cut -d'"' -f4)
    echo "✅ Skaffold: $SKAF_VERSION"
else
    echo "⚠️ Skaffold not found (optional for dev workflows)"
fi

echo ""
echo "🎉 Prerequisites check complete!"
```

Save as `check-prerequisites.sh`, make executable, and run:

```bash
chmod +x check-prerequisites.sh
./check-prerequisites.sh
```

### Manual Verification Commands

```bash
# Verify Docker
docker --version
docker info

# Verify K3d
k3d --version

# Verify kubectl
kubectl version --client

# Verify Helm
helm version

# Optional: Verify development tools
telepresence version
skaffold version
```

## Platform-Specific Setup

### macOS
```bash
# Install via Homebrew
brew install docker k3d kubectl helm

# Optional development tools
brew install telepresence skaffold
```

### Ubuntu/Debian
```bash
# Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# K3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Helm
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
```

### Windows (WSL2)
```powershell
# Install Docker Desktop for Windows
# Follow: https://docs.docker.com/desktop/windows/install/

# In WSL2 terminal, follow Ubuntu instructions above
```

## Troubleshooting Common Issues

### Docker Permission Issues (Linux)
```bash
# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Verify access
docker run hello-world
```

### K3d Port Conflicts
```bash
# Check for port conflicts
netstat -tulpn | grep :80
netstat -tulpn | grep :443

# Kill conflicting processes if needed
sudo lsof -ti:80 | xargs kill -9
```

### kubectl Context Issues
```bash
# Reset kubectl configuration
kubectl config get-contexts
kubectl config use-context k3d-my-cluster
```

## Next Steps

Once all prerequisites are installed and verified:

1. **[Quick Start](./quick-start.md)** - Bootstrap your first OpenFrame environment
2. **[First Steps](./first-steps.md)** - Explore key features and workflows

> **💡 Pro Tip**: OpenFrame CLI will automatically check prerequisites when you run commands. If something is missing, it will provide specific installation guidance for your platform.

---

## Verification Checklist

- [ ] Docker installed and daemon running
- [ ] K3d installed and accessible
- [ ] kubectl installed and working
- [ ] Helm installed and accessible
- [ ] System meets minimum requirements (4GB RAM, 10GB disk)
- [ ] Network connectivity confirmed
- [ ] Optional: Telepresence and Skaffold installed for development workflows

Ready to proceed? Let's move to the [Quick Start](./quick-start.md) guide!