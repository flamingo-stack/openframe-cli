# Prerequisites Guide

Before using OpenFrame CLI, ensure your system meets the requirements and has the necessary tools installed.

## System Requirements

| Component | Minimum | Recommended | Notes |
|-----------|---------|-------------|-------|
| **Operating System** | Linux, macOS, Windows | Linux/macOS | Windows requires WSL2 for optimal experience |
| **Memory (RAM)** | 8 GB | 16 GB | K3d clusters require adequate memory |
| **Disk Space** | 10 GB free | 20 GB free | For Docker images and cluster data |
| **CPU** | 2 cores | 4 cores | Better performance with more cores |
| **Internet** | Required | Stable connection | For downloading images and charts |

## Required Software

### Core Dependencies

| Tool | Version | Installation | Purpose |
|------|---------|--------------|---------|
| **Docker** | 20.10+ | [Install Docker](https://docs.docker.com/get-docker/) | Container runtime for K3d |
| **kubectl** | 1.24+ | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) | Kubernetes command-line tool |
| **Helm** | 3.8+ | [Install Helm](https://helm.sh/docs/intro/install/) | Package manager for Kubernetes |
| **K3d** | 5.4+ | [Install K3d](https://k3d.io/v5.4.6/#installation) | Lightweight Kubernetes in Docker |

### Development Tools (Optional)

| Tool | Version | Installation | Purpose |
|------|---------|--------------|---------|
| **Telepresence** | 2.10+ | [Install Telepresence](https://www.telepresence.io/docs/latest/install/) | Traffic interception for `dev` commands |
| **Skaffold** | 2.0+ | [Install Skaffold](https://skaffold.dev/docs/install/) | Development workflows |
| **Git** | 2.30+ | [Install Git](https://git-scm.com/downloads) | Repository management |

## Installation Instructions

### 1. Docker Installation

#### Linux (Ubuntu/Debian)
```bash
# Update package index
sudo apt-get update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Add user to docker group
sudo usermod -aG docker $USER

# Verify installation
docker --version
```

#### macOS
```bash
# Install using Homebrew
brew install --cask docker

# Or download Docker Desktop from:
# https://docs.docker.com/desktop/install/mac-install/

# Verify installation
docker --version
```

#### Windows
1. Install [Docker Desktop for Windows](https://docs.docker.com/desktop/install/windows-install/)
2. Enable WSL2 integration
3. Verify with: `docker --version`

### 2. kubectl Installation

#### Linux
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
kubectl version --client
```

#### macOS
```bash
brew install kubectl
kubectl version --client
```

#### Windows
```bash
# Using Chocolatey
choco install kubernetes-cli

# Or using Scoop
scoop install kubectl

kubectl version --client
```

### 3. Helm Installation

#### Linux/macOS
```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
helm version
```

#### Windows
```bash
# Using Chocolatey
choco install kubernetes-helm

# Using Scoop
scoop install helm

helm version
```

### 4. K3d Installation

#### Linux/macOS
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
k3d version
```

#### Windows
```bash
# Using Chocolatey
choco install k3d

# Using Scoop
scoop install k3d

k3d version
```

## Optional Development Tools

### Telepresence (for `dev intercept` commands)

#### All Platforms
```bash
# Download and install the latest version
curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o telepresence
sudo install -o root -g root -m 0755 telepresence /usr/local/bin/telepresence

# Verify installation
telepresence version
```

### Skaffold (for `dev skaffold` commands)

#### Linux/macOS
```bash
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/
skaffold version
```

#### Windows
```bash
# Using Chocolatey
choco install skaffold

skaffold version
```

## Environment Configuration

### Docker Configuration

Ensure Docker daemon is running and properly configured:

```bash
# Start Docker service (Linux)
sudo systemctl start docker
sudo systemctl enable docker

# Verify Docker can run containers
docker run hello-world

# Configure Docker for K3d (allocate sufficient resources)
# On Docker Desktop: Settings > Resources > Advanced
# - Memory: 4GB minimum (8GB recommended)
# - CPU: 2 cores minimum (4 cores recommended)
```

### Environment Variables

Set up optional environment variables for customization:

```bash
# Default cluster configuration
export OPENFRAME_DEFAULT_CLUSTER_NAME="openframe-dev"
export OPENFRAME_DEFAULT_NODES=3
export OPENFRAME_DEFAULT_DEPLOYMENT_MODE="oss-tenant"

# Development tool configuration
export TELEPRESENCE_NAMESPACE="default"
export SKAFFOLD_DEFAULT_REPO="localhost:5000"

# Add to ~/.bashrc or ~/.zshrc to persist
echo 'export OPENFRAME_DEFAULT_CLUSTER_NAME="openframe-dev"' >> ~/.bashrc
```

## Verification Commands

Run these commands to verify your setup is complete:

### Core Tools Verification
```bash
# Check Docker
docker --version && docker info > /dev/null 2>&1 && echo "✅ Docker OK" || echo "❌ Docker issue"

# Check kubectl
kubectl version --client --short && echo "✅ kubectl OK" || echo "❌ kubectl issue"

# Check Helm
helm version --short && echo "✅ Helm OK" || echo "❌ Helm issue"

# Check K3d
k3d version && echo "✅ K3d OK" || echo "❌ K3d issue"
```

### Development Tools Verification (Optional)
```bash
# Check Telepresence
telepresence version && echo "✅ Telepresence OK" || echo "❌ Telepresence not installed"

# Check Skaffold  
skaffold version && echo "✅ Skaffold OK" || echo "❌ Skaffold not installed"

# Check Git
git --version && echo "✅ Git OK" || echo "❌ Git issue"
```

### Complete Verification Script
```bash
#!/bin/bash
echo "🔍 OpenFrame Prerequisites Check"
echo "================================"

# Check required tools
declare -a tools=("docker" "kubectl" "helm" "k3d")
declare -a optional=("telepresence" "skaffold" "git")

echo "📋 Required Tools:"
for tool in "${tools[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool installed"
    else
        echo "❌ $tool missing - required"
    fi
done

echo ""
echo "🔧 Optional Tools:"
for tool in "${optional[@]}"; do
    if command -v $tool &> /dev/null; then
        echo "✅ $tool installed"
    else
        echo "⚠️  $tool not installed - optional for dev commands"
    fi
done

# Check Docker is running
if docker info &> /dev/null; then
    echo "✅ Docker daemon running"
else
    echo "❌ Docker daemon not running"
fi

echo ""
echo "🎯 Ready for OpenFrame CLI installation!"
```

## Account Requirements

### No External Accounts Required

OpenFrame CLI works entirely with local tools - no external service accounts or subscriptions are needed for basic functionality.

### Optional Integrations

If you plan to use advanced features:

| Service | Purpose | Required For |
|---------|---------|-------------|
| **GitHub Account** | Repository access for app-of-apps | Custom chart repositories |
| **Container Registry** | Custom image storage | Private container images |
| **Cloud Provider** | Remote cluster management | Production deployments |

## Troubleshooting Common Issues

### Docker Issues

**Problem**: Permission denied while connecting to Docker daemon
```bash
# Solution: Add user to docker group and restart
sudo usermod -aG docker $USER
newgrp docker
# Or log out and back in
```

**Problem**: Docker Desktop not starting on macOS/Windows
```text
Solution: 
1. Restart Docker Desktop application
2. Increase memory allocation in settings
3. Disable antivirus real-time scanning temporarily
```

### K3d Issues

**Problem**: K3d cluster creation fails
```bash
# Check Docker is running
docker info

# Check available resources
docker system df

# Clean up old clusters
k3d cluster list
k3d cluster delete --all
```

### Network Issues

**Problem**: Can't download images or charts
```bash
# Check internet connectivity
ping google.com

# Check proxy settings if behind corporate firewall
echo $HTTP_PROXY
echo $HTTPS_PROXY

# Configure Docker proxy if needed
# Edit ~/.docker/config.json
```

## What's Next?

✅ **Prerequisites Complete!** You're ready to proceed to:

1. **[Quick Start Guide](./quick-start.md)** - Install and run OpenFrame in 5 minutes
2. **[First Steps Guide](./first-steps.md)** - Explore OpenFrame features
3. **[Development Setup](../development/setup/environment.md)** - Configure development environment

> 💡 **Tip**: Run the verification script above to double-check your setup before proceeding to the Quick Start guide.