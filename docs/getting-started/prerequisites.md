# Prerequisites

Before installing and using OpenFrame CLI, ensure your system meets the following requirements. The CLI includes automatic prerequisite checking, but it's helpful to understand what's needed upfront.

## Required Software and Versions

| Software | Minimum Version | Purpose | Installation |
|----------|-----------------|---------|--------------|
| **Docker** | 20.10+ | Container runtime for K3d clusters | [Get Docker](https://docs.docker.com/get-docker/) |
| **K3d** | 5.4+ | Lightweight Kubernetes distribution | [Install K3d](https://k3d.io/v5.4.6/#installation) |
| **kubectl** | 1.25+ | Kubernetes command-line tool | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.10+ | Package manager for Kubernetes | [Install Helm](https://helm.sh/docs/intro/install/) |
| **Telepresence** | 2.10+ (optional) | Traffic interception for development | [Install Telepresence](https://www.telepresence.io/docs/latest/install/) |
| **Skaffold** | 2.0+ (optional) | Live development workflows | [Install Skaffold](https://skaffold.dev/docs/install/) |

> **📝 Note**: Telepresence and Skaffold are only required if you plan to use the `openframe dev` commands for local development workflows.

## System Requirements

### **Hardware**
- **RAM**: Minimum 4GB, recommended 8GB+ for full development workflows
- **CPU**: 2+ cores recommended
- **Disk Space**: 10GB+ free space for Docker images and cluster data
- **Network**: Internet access for downloading images and charts

### **Operating Systems**
- **Linux**: Ubuntu 20.04+, CentOS 7+, or equivalent
- **macOS**: 10.15+ (Catalina or newer)
- **Windows**: Windows 10/11 with WSL2

### **Container Runtime**
- Docker Desktop (recommended) or compatible Docker engine
- Docker daemon must be running and accessible
- User must have permission to run Docker commands

## Account and Access Requirements

### **Container Registries**
- Access to public Docker registries (docker.io, ghcr.io)
- For private deployments: credentials for private registries

### **Network Access**
- Outbound HTTPS access (ports 443, 80)
- Access to Helm chart repositories
- Access to ArgoCD repositories (if using GitOps)

### **Kubernetes Access**
- No pre-existing Kubernetes cluster required (K3d creates local clusters)
- For production deployments: appropriate cluster access and permissions

## Environment Variables (Optional)

While not required for basic operation, these environment variables can customize OpenFrame CLI behavior:

| Variable | Purpose | Default | Example |
|----------|---------|---------|---------|
| `KUBECONFIG` | Kubernetes configuration file path | `~/.kube/config` | `/path/to/kubeconfig` |
| `K3D_FIX_DNS` | Fix DNS issues in K3d | Not set | `1` |
| `DOCKER_HOST` | Docker daemon connection | Default socket | `tcp://localhost:2376` |
| `HELM_CACHE_HOME` | Helm cache directory | `~/.cache/helm` | `/tmp/helm-cache` |

> **💡 Pro Tip**: Set `K3D_FIX_DNS=1` if you experience DNS resolution issues in your K3d clusters.

## Verification Commands

Run these commands to verify your system is ready for OpenFrame CLI:

### Check Docker
```bash
docker --version
docker info
docker run --rm hello-world
```

Expected output should show Docker version 20.10+ and successful container execution.

### Check K3d
```bash
k3d version
```

Expected output should show K3d version 5.4+ and available commands.

### Check kubectl
```bash
kubectl version --client
```

Expected output should show kubectl client version 1.25+.

### Check Helm
```bash
helm version
```

Expected output should show Helm version 3.10+.

### Check Development Tools (Optional)
```bash
# Only if you plan to use dev commands
telepresence version
skaffold version
```

## Quick Verification Script

Save this script as `check-prerequisites.sh` and run it to verify all requirements:

```bash
#!/bin/bash

echo "🔍 Checking OpenFrame CLI Prerequisites..."
echo

# Check Docker
echo "📦 Checking Docker..."
if command -v docker &> /dev/null; then
    docker_version=$(docker --version | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo "✅ Docker $docker_version found"
    if ! docker info &> /dev/null; then
        echo "⚠️  Docker daemon not running or not accessible"
    fi
else
    echo "❌ Docker not found"
fi

# Check K3d
echo "🐳 Checking K3d..."
if command -v k3d &> /dev/null; then
    k3d_version=$(k3d version | grep k3d | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
    echo "✅ K3d $k3d_version found"
else
    echo "❌ K3d not found"
fi

# Check kubectl
echo "☸️  Checking kubectl..."
if command -v kubectl &> /dev/null; then
    kubectl_version=$(kubectl version --client --short 2>/dev/null | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
    echo "✅ kubectl $kubectl_version found"
else
    echo "❌ kubectl not found"
fi

# Check Helm
echo "⎈  Checking Helm..."
if command -v helm &> /dev/null; then
    helm_version=$(helm version --short | grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+')
    echo "✅ Helm $helm_version found"
else
    echo "❌ Helm not found"
fi

echo
echo "🚀 Prerequisites check complete!"
```

Make it executable and run:
```bash
chmod +x check-prerequisites.sh
./check-prerequisites.sh
```

## Installation Help

### Quick Install Commands

**Docker (Ubuntu/Debian):**
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

**K3d:**
```bash
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

**kubectl:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
```

**Helm:**
```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### macOS (using Homebrew):
```bash
brew install docker
brew install k3d
brew install kubectl
brew install helm
```

## Troubleshooting

### Common Issues

**Docker Permission Denied:**
```bash
sudo usermod -aG docker $USER
newgrp docker
```

**K3d DNS Issues:**
```bash
export K3D_FIX_DNS=1
```

**kubectl Not Found After Installation:**
```bash
export PATH=$PATH:~/.local/bin
echo 'export PATH=$PATH:~/.local/bin' >> ~/.bashrc
```

## Next Steps

Once you've verified all prerequisites are installed:

1. **[Quick Start](quick-start.md)** - Install and run OpenFrame CLI
2. **[First Steps](first-steps.md)** - Learn essential commands and workflows

> **🔧 Need Help?** The OpenFrame CLI includes built-in prerequisite checking. Run any command and it will validate dependencies automatically.