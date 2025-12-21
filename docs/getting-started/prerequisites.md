# Prerequisites

Before you can use OpenFrame CLI effectively, you'll need to ensure your system has the required software and meets the minimum requirements. This guide will walk you through all the prerequisites and help you verify your setup.

## System Requirements

### Minimum Specifications

| Requirement | Minimum | Recommended | Notes |
|-------------|---------|-------------|-------|
| **CPU** | 2 cores | 4+ cores | For cluster nodes and container builds |
| **RAM** | 4 GB | 8+ GB | Kubernetes requires significant memory |
| **Disk Space** | 10 GB | 20+ GB | For container images and cluster data |
| **OS** | Linux/macOS/Windows | Linux/macOS | WSL2 required on Windows |

### Supported Operating Systems

| OS | Version | Notes |
|-----|---------|-------|
| **Linux** | Ubuntu 18.04+, CentOS 7+, RHEL 8+ | Native Docker support |
| **macOS** | 10.14+ (Mojave) | Docker Desktop required |
| **Windows** | Windows 10 with WSL2 | Use WSL2 for best performance |

## Required Software

The following tools are **required** for OpenFrame CLI to function properly:

### 🐳 Docker

Docker is essential for running K3d clusters and container management.

| OS | Installation Method | Version Required |
|----|-------------------|-----------------|
| **Linux** | Package manager or Docker convenience script | 20.10+ |
| **macOS** | [Docker Desktop for Mac](https://docs.docker.com/desktop/mac/install/) | 4.0+ |
| **Windows** | [Docker Desktop with WSL2](https://docs.docker.com/desktop/windows/install/) | 4.0+ |

**Installation:**
```bash
# Linux (Ubuntu/Debian)
curl -fsSL https://get.docker.com | bash
sudo usermod -aG docker $USER

# macOS (Homebrew)
brew install --cask docker

# Verify installation
docker --version
docker run hello-world
```

### ⚡ K3d

K3d creates lightweight Kubernetes clusters using Docker containers.

```bash
# Install K3d
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Verify installation
k3d --version
```

### ☸️ kubectl

Kubernetes command-line tool for cluster management.

```bash
# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# macOS
brew install kubectl

# Windows (using Chocolatey)
choco install kubernetes-cli

# Verify installation
kubectl version --client
```

### ⛵ Helm

Helm is the Kubernetes package manager, required for chart installations.

```bash
# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# macOS
brew install helm

# Verify installation
helm version
```

## Optional but Recommended Tools

These tools enhance your development experience with OpenFrame CLI:

### 🔗 Telepresence (for development workflows)

```bash
# macOS
brew install datawire/blackbird/telepresence

# Linux
sudo curl -fL https://app.getambassador.io/download/tel2/linux/amd64/latest/telepresence -o /usr/local/bin/telepresence
sudo chmod a+x /usr/local/bin/telepresence

# Verify
telepresence version
```

### 🏗️ Skaffold (for CI/CD workflows)

```bash
# Linux/macOS
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/

# macOS
brew install skaffold

# Verify
skaffold version
```

## Environment Variables

OpenFrame CLI may require these environment variables depending on your setup:

```bash
# Docker configuration (if using non-default daemon)
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration (optional - kubectl default)
export KUBECONFIG=$HOME/.kube/config

# OpenFrame specific (optional)
export OPENFRAME_CONFIG_PATH=$HOME/.openframe
export OPENFRAME_LOG_LEVEL=info
```

Add these to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
echo 'export KUBECONFIG=$HOME/.kube/config' >> ~/.bashrc
echo 'export OPENFRAME_LOG_LEVEL=info' >> ~/.bashrc
source ~/.bashrc
```

## Network Requirements

### Port Requirements

OpenFrame CLI and its tools require access to these ports:

| Port Range | Purpose | Required For |
|------------|---------|--------------|
| `6443` | Kubernetes API server | kubectl access |
| `8080-8090` | Development services | Local debugging |
| `30000-32767` | NodePort services | Service exposure |
| `80, 443` | HTTP/HTTPS traffic | Web applications |

### Internet Access

The following domains must be accessible:

- `docker.io` - Docker Hub for images
- `k8s.io` - Kubernetes releases
- `helm.sh` - Helm charts
- `github.com` - Source code and releases
- `gcr.io` - Google Container Registry

## Verification Commands

Run these commands to verify your setup is ready:

<details>
<summary>📋 <strong>Complete Verification Script</strong></summary>

```bash
#!/bin/bash

echo "🔍 OpenFrame CLI Prerequisites Checker"
echo "======================================"

# Check Docker
if command -v docker &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
    if docker ps &> /dev/null; then
        echo "   ✅ Docker daemon is running"
    else
        echo "   ❌ Docker daemon is not running"
    fi
else
    echo "❌ Docker: Not installed"
fi

# Check K3d
if command -v k3d &> /dev/null; then
    echo "✅ K3d: $(k3d --version)"
else
    echo "❌ K3d: Not installed"
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    echo "✅ kubectl: $(kubectl version --client --short)"
else
    echo "❌ kubectl: Not installed"
fi

# Check Helm
if command -v helm &> /dev/null; then
    echo "✅ Helm: $(helm version --short)"
else
    echo "❌ Helm: Not installed"
fi

# Check optional tools
echo ""
echo "Optional Tools:"
if command -v telepresence &> /dev/null; then
    echo "✅ Telepresence: $(telepresence version --output json | jq -r '.client.version' 2>/dev/null || telepresence version)"
else
    echo "⚠️  Telepresence: Not installed (optional for development)"
fi

if command -v skaffold &> /dev/null; then
    echo "✅ Skaffold: $(skaffold version)"
else
    echo "⚠️  Skaffold: Not installed (optional for CI/CD)"
fi

# System resources
echo ""
echo "System Resources:"
echo "💾 Available disk space: $(df -h / | awk 'NR==2 {print $4}')"
echo "🧠 Available memory: $(free -h 2>/dev/null | awk 'NR==2 {print $7}' || echo 'Unknown')"

echo ""
echo "🎉 Prerequisites check complete!"
```

</details>

### Quick Check Commands

```bash
# Essential tools
docker --version && echo "Docker: ✅" || echo "Docker: ❌"
k3d --version && echo "K3d: ✅" || echo "K3d: ❌"  
kubectl version --client && echo "kubectl: ✅" || echo "kubectl: ❌"
helm version && echo "Helm: ✅" || echo "Helm: ❌"

# Test Docker functionality
docker run hello-world && echo "Docker functionality: ✅" || echo "Docker functionality: ❌"

# Check available resources
echo "Free disk space: $(df -h . | awk 'NR==2 {print $4}')"
echo "Free memory: $(free -h | awk 'NR==2 {print $7}' 2>/dev/null || echo 'Check manually')"
```

## Troubleshooting Common Issues

### Docker Issues

**Problem**: Permission denied when running Docker commands
```bash
# Solution: Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in, or run:
newgrp docker
```

**Problem**: Docker daemon not running
```bash
# Linux
sudo systemctl start docker
sudo systemctl enable docker

# macOS: Start Docker Desktop application
```

### K3d Issues

**Problem**: K3d clusters fail to start
```bash
# Check Docker is running
docker ps

# Clean up any failed clusters  
k3d cluster delete --all

# Ensure sufficient resources
docker system prune
```

### Network Issues

**Problem**: Cannot pull images or charts
```bash
# Test internet connectivity
curl -I https://docker.io
curl -I https://k8s.io

# Check DNS resolution
nslookup docker.io
```

## Next Steps

Once all prerequisites are installed and verified:

1. **[Install OpenFrame CLI](quick-start.md#installation)**
2. **[Run your first bootstrap](quick-start.md#bootstrap-your-first-environment)**
3. **[Explore first steps](first-steps.md)**

---

> **💡 Pro Tip**: Save the verification script above as `check-prereqs.sh` and run it anytime you're troubleshooting setup issues. Most OpenFrame CLI problems stem from missing or misconfigured prerequisites.