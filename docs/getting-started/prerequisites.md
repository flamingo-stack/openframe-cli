# Prerequisites

Before installing and using OpenFrame CLI, ensure your system meets the following requirements and has the necessary dependencies installed.

## System Requirements

### Hardware Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| **RAM** | 24GB | 32GB |
| **CPU Cores** | 6 cores | 12 cores |
| **Disk Space** | 50GB free | 100GB free |
| **Architecture** | x86_64, ARM64 | x86_64, ARM64 |

### Operating System Support

| OS | Version | Status |
|---|---------|--------|
| **Linux** | Ubuntu 20.04+, CentOS 8+, RHEL 8+ | ✅ Fully Supported |
| **macOS** | 11+ (Big Sur and later) | ✅ Fully Supported |
| **Windows** | Windows 10/11 with WSL2 | ⚠️ Supported via WSL2 |

> **Windows Users**: OpenFrame CLI requires WSL2 for proper Kubernetes integration. Native Windows support is not currently available.

## Required Dependencies

### Core Dependencies

These tools must be installed before using OpenFrame CLI:

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Docker** | 20.10+ | Container runtime for K3D clusters | [Docker Install Guide](https://docs.docker.com/get-docker/) |
| **kubectl** | 1.25+ | Kubernetes command-line tool | [kubectl Install Guide](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.10+ | Kubernetes package manager | [Helm Install Guide](https://helm.sh/docs/intro/install/) |
| **K3D** | 5.0+ | Lightweight Kubernetes clusters | [K3D Install Guide](https://k3d.io/v5.4.6/#installation) |

### Development Dependencies (Optional)

Required only if using development features (`openframe dev` commands):

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Service intercepts for local development | [Telepresence Install](https://www.telepresence.io/docs/latest/install/) |
| **jq** | 1.6+ | JSON processing for dev scripts | [jq Install Guide](https://jqlang.github.io/jq/download/) |

## Installation Verification

### Check Docker
```bash
docker --version
docker ps
```

Expected output:
```text
Docker version 20.10.0 or higher
CONTAINER ID   IMAGE     COMMAND   CREATED   STATUS   PORTS   NAMES
```

### Check kubectl
```bash
kubectl version --client
```

Expected output:
```text
Client Version: version.Info{Major:"1", Minor:"25"+...}
```

### Check Helm
```bash
helm version
```

Expected output:
```text
version.BuildInfo{Version:"v3.10.0"+...}
```

### Check K3D
```bash
k3d version
```

Expected output:
```text
k3d version v5.0.0+
```

### Check Telepresence (Optional)
```bash
telepresence version
```

Expected output:
```text
Client: v2.10.0+
```

### Check jq (Optional)
```bash
jq --version
```

Expected output:
```text
jq-1.6+
```

## Network Requirements

### Outbound Connectivity
OpenFrame CLI requires internet access for:
- Pulling Docker images
- Downloading Helm charts
- Accessing Git repositories
- Installing prerequisites

### Port Requirements
| Port Range | Protocol | Purpose |
|------------|----------|---------|
| 80, 443 | TCP | HTTPS/HTTP for downloads |
| 6443 | TCP | Kubernetes API server |
| 30000-32767 | TCP | Kubernetes NodePort range |
| 2376, 2377 | TCP | Docker daemon (if remote) |

### Firewall Considerations
Ensure your firewall allows:
- Docker daemon communication
- Kubernetes cluster communication
- Outbound HTTPS connections

## Environment Variables

Set these environment variables for optimal experience:

### Required
```bash
# Docker daemon configuration
export DOCKER_HOST="unix:///var/run/docker.sock"

# Kubernetes configuration
export KUBECONFIG="$HOME/.kube/config"
```

### Optional
```bash
# OpenFrame CLI configuration
export OPENFRAME_LOG_LEVEL="info"
export OPENFRAME_CONFIG_DIR="$HOME/.openframe"

# Development tools
export TELEPRESENCE_LOGIN_DOMAIN="auth.datawire.io"
```

## Account Requirements

### Container Registry Access
- **Public registries**: Docker Hub, GHCR.io (no authentication required)
- **Private registries**: Configure Docker credentials if using private images

### Git Repository Access
- **Public repositories**: No authentication required
- **Private repositories**: Configure SSH keys or personal access tokens

## Quick Setup Script

For convenience, here's a script to verify all prerequisites:

```bash
#!/bin/bash
# prerequisites-check.sh

echo "🔍 Checking OpenFrame CLI prerequisites..."

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check version
check_version() {
    local tool="$1"
    local min_version="$2"
    local current_version="$3"
    
    echo "  $tool: $current_version (required: $min_version+)"
}

errors=0

# Check Docker
if command_exists docker; then
    docker_version=$(docker --version | cut -d' ' -f3 | cut -d',' -f1)
    check_version "Docker" "20.10.0" "$docker_version"
    if ! docker ps >/dev/null 2>&1; then
        echo "  ❌ Docker daemon is not running or accessible"
        ((errors++))
    fi
else
    echo "  ❌ Docker not found"
    ((errors++))
fi

# Check kubectl
if command_exists kubectl; then
    kubectl_version=$(kubectl version --client -o json 2>/dev/null | jq -r '.clientVersion.gitVersion' 2>/dev/null || echo "unknown")
    check_version "kubectl" "1.25.0" "$kubectl_version"
else
    echo "  ❌ kubectl not found"
    ((errors++))
fi

# Check Helm
if command_exists helm; then
    helm_version=$(helm version --short | cut -d'+' -f1)
    check_version "Helm" "3.10.0" "$helm_version"
else
    echo "  ❌ Helm not found"
    ((errors++))
fi

# Check K3D
if command_exists k3d; then
    k3d_version=$(k3d version | grep k3d | cut -d' ' -f2)
    check_version "K3D" "5.0.0" "$k3d_version"
else
    echo "  ❌ K3D not found"
    ((errors++))
fi

# Check optional tools
echo ""
echo "📋 Optional development tools:"

if command_exists telepresence; then
    telepresence_version=$(telepresence version --output=json 2>/dev/null | jq -r '.client.version' 2>/dev/null || "unknown")
    check_version "Telepresence" "2.10.0" "$telepresence_version"
else
    echo "  ⚠️  Telepresence not found (optional for dev workflows)"
fi

if command_exists jq; then
    jq_version=$(jq --version | cut -d'-' -f2)
    check_version "jq" "1.6" "$jq_version"
else
    echo "  ⚠️  jq not found (optional for dev scripts)"
fi

echo ""
if [ $errors -eq 0 ]; then
    echo "✅ All required prerequisites are installed!"
    echo "🚀 You're ready to install OpenFrame CLI"
else
    echo "❌ $errors required dependencies are missing"
    echo "📖 Please install missing dependencies before proceeding"
    exit 1
fi
```

Save this as `prerequisites-check.sh`, make it executable, and run:

```bash
chmod +x prerequisites-check.sh
./prerequisites-check.sh
```

## Troubleshooting Common Issues

### Docker Permission Issues
If you get permission denied errors:

```bash
# Add your user to the docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker
```

### kubectl Not Finding Config
```bash
# Create kubeconfig directory
mkdir -p ~/.kube

# Verify KUBECONFIG environment variable
echo `$KUBECONFIG`
```

### K3D Installation Issues on macOS
```bash
# If using Homebrew
brew install k3d

# If using curl
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
```

### WSL2 Setup on Windows
1. Enable WSL2: `wsl --install`
2. Install Ubuntu from Microsoft Store
3. Install Docker Desktop with WSL2 backend
4. Install all prerequisites inside WSL2 environment

## Next Steps

Once all prerequisites are installed and verified:

1. **[Quick Start Guide](quick-start.md)** - Install OpenFrame CLI and bootstrap your first environment
2. **[First Steps Guide](first-steps.md)** - Explore key features and workflows

Need help? Join our community:
- **OpenMSP Slack**: [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)