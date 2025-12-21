# Prerequisites

Before you can use OpenFrame CLI, you'll need to install several prerequisite tools and ensure your system meets the minimum requirements. This guide will walk you through each requirement and help you verify your setup.

## System Requirements

| Requirement | Minimum Version | Recommended | Notes |
|-------------|----------------|-------------|--------|
| **Operating System** | Linux, macOS, Windows (WSL2) | Latest stable | Windows requires WSL2 or Docker Desktop |
| **RAM** | 8GB | 16GB+ | K3d clusters require sufficient memory |
| **CPU** | 2 cores | 4+ cores | More cores improve cluster performance |
| **Disk Space** | 10GB free | 20GB+ free | For container images and cluster data |
| **Network** | Internet access | Stable connection | Required for downloading images and charts |

## Required Software

### 1. Docker Engine

Docker is required for K3d to create lightweight Kubernetes clusters.

#### Installation

**macOS:**
```bash
# Using Homebrew
brew install --cask docker

# Or download Docker Desktop from https://docs.docker.com/desktop/mac/install/
```

**Linux (Ubuntu/Debian):**
```bash
# Update package index
sudo apt-get update

# Install Docker
sudo apt-get install docker.io

# Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

**Windows:**
```bash
# Install Docker Desktop for Windows
# Download from https://docs.docker.com/desktop/windows/install/
# Ensure WSL2 backend is enabled
```

#### Verification
```bash
docker --version
docker run hello-world
```

Expected output:
```
Docker version 20.10.0+
Hello from Docker!
```

### 2. kubectl

Kubernetes command-line tool for cluster interaction.

#### Installation

**macOS:**
```bash
# Using Homebrew
brew install kubectl

# Or using curl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

**Linux:**
```bash
# Download latest stable version
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

**Windows (WSL2):**
```bash
# Download kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/windows/amd64/kubectl.exe"
# Move to PATH directory
```

#### Verification
```bash
kubectl version --client
```

### 3. Helm

Package manager for Kubernetes applications.

#### Installation

**macOS:**
```bash
# Using Homebrew
brew install helm
```

**Linux:**
```bash
# Using the installer script
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

**Windows:**
```bash
# Using Chocolatey
choco install kubernetes-helm

# Or using Scoop
scoop install helm
```

#### Verification
```bash
helm version
```

### 4. K3d

Lightweight wrapper to run K3s (lightweight Kubernetes) in Docker.

#### Installation

**macOS/Linux:**
```bash
# Using the install script
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Or using Homebrew (macOS)
brew install k3d
```

**Windows:**
```bash
# Using Chocolatey
choco install k3d

# Or using Scoop
scoop install k3d
```

#### Verification
```bash
k3d version
```

## Optional Development Tools

These tools are required only if you plan to use the `openframe dev` commands:

### 5. Telepresence (Optional)

For local development with remote cluster connectivity.

#### Installation

**macOS:**
```bash
# Download and install
sudo curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.15.1/telepresence-darwin-amd64 -o /usr/local/bin/telepresence
sudo chmod a+x /usr/local/bin/telepresence
```

**Linux:**
```bash
# Download and install
sudo curl -fL https://app.getambassador.io/download/tel2oss/releases/download/v2.15.1/telepresence-linux-amd64 -o /usr/local/bin/telepresence
sudo chmod a+x /usr/local/bin/telepresence
```

#### Verification
```bash
telepresence version
```

### 6. Skaffold (Optional)

For continuous development workflows.

#### Installation

**macOS:**
```bash
# Using Homebrew
brew install skaffold

# Or using curl
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-darwin-amd64
sudo install skaffold /usr/local/bin/
```

**Linux:**
```bash
# Download and install
curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64
sudo install skaffold /usr/local/bin/
```

#### Verification
```bash
skaffold version
```

## Environment Variables

OpenFrame CLI doesn't require specific environment variables, but these can be helpful:

```bash
# Optional: Set default deployment mode
export OPENFRAME_DEPLOYMENT_MODE=oss-tenant

# Optional: Default cluster name
export OPENFRAME_DEFAULT_CLUSTER=dev-cluster

# Docker environment (if not using Docker Desktop)
export DOCKER_HOST=unix:///var/run/docker.sock
```

## Account Requirements

### GitHub Account (Optional)
- Required only if you plan to use private repositories for ArgoCD applications
- Personal access token may be needed for private repositories

### Container Registry Access (Optional)
- Required only if you use private container registries
- Docker Hub, GCR, ECR, or other registry credentials

## Network Requirements

| Service | Ports | Description |
|---------|--------|-------------|
| **K3d API Server** | 6443 | Kubernetes API access |
| **ArgoCD UI** | 8080 | ArgoCD web interface |
| **Local Development** | 8000-9000 | Common development ports |
| **Docker Registry** | 5000 | Local registry (if used) |

> **Note**: OpenFrame CLI will automatically configure port forwarding for services like ArgoCD.

## Verification Script

Run this script to verify all prerequisites are properly installed:

```bash
#!/bin/bash

echo "🔍 Checking OpenFrame CLI Prerequisites..."
echo

# Check Docker
if command -v docker &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
    if docker info &> /dev/null; then
        echo "   Docker daemon is running"
    else
        echo "❌ Docker daemon is not running"
        exit 1
    fi
else
    echo "❌ Docker not found"
    exit 1
fi

# Check kubectl
if command -v kubectl &> /dev/null; then
    echo "✅ kubectl: $(kubectl version --client --short)"
else
    echo "❌ kubectl not found"
    exit 1
fi

# Check Helm
if command -v helm &> /dev/null; then
    echo "✅ Helm: $(helm version --short)"
else
    echo "❌ Helm not found"
    exit 1
fi

# Check K3d
if command -v k3d &> /dev/null; then
    echo "✅ K3d: $(k3d version | head -n 1)"
else
    echo "❌ K3d not found"
    exit 1
fi

# Check optional tools
echo
echo "📋 Optional Development Tools:"

if command -v telepresence &> /dev/null; then
    echo "✅ Telepresence: $(telepresence version 2>/dev/null || echo 'Installed')"
else
    echo "⚠️  Telepresence not found (optional for dev commands)"
fi

if command -v skaffold &> /dev/null; then
    echo "✅ Skaffold: $(skaffold version --output=plain)"
else
    echo "⚠️  Skaffold not found (optional for dev commands)"
fi

echo
echo "🎉 All required prerequisites are installed!"
echo "You're ready to install OpenFrame CLI!"
```

Save this script as `check-prerequisites.sh` and run:

```bash
chmod +x check-prerequisites.sh
./check-prerequisites.sh
```

## Common Issues

### Docker Permission Denied
**Issue**: `permission denied while trying to connect to the Docker daemon socket`

**Solution**:
```bash
sudo usermod -aG docker $USER
newgrp docker
# Or restart your terminal session
```

### K3d Port Already in Use
**Issue**: `port 6443 already in use`

**Solution**:
```bash
# Check what's using the port
sudo lsof -i :6443

# Kill the process or use a different port
k3d cluster create --api-port 6444
```

### WSL2 Docker Issues
**Issue**: Docker commands fail in Windows WSL2

**Solution**:
1. Ensure Docker Desktop is running
2. Enable WSL2 integration in Docker Desktop settings
3. Restart WSL2: `wsl --shutdown` then reopen terminal

## Next Steps

Once you have all prerequisites installed and verified:

1. **[Quick Start Guide](quick-start.md)** - Get OpenFrame CLI running in 5 minutes
2. **[First Steps](first-steps.md)** - Essential configuration and usage
3. **[Development Setup](../development/setup/environment.md)** - For contributing to OpenFrame CLI

---

> 💡 **Pro Tip**: Bookmark this prerequisites page - you can use the verification script anytime to ensure your environment is properly configured for OpenFrame CLI.