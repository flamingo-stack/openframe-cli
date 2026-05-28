# Prerequisites

Before using OpenFrame CLI, ensure your system meets the requirements and has the necessary tools installed.

## System Requirements

| Requirement | Minimum | Recommended | Notes |
|-------------|---------|-------------|-------|
| **Memory** | 24GB RAM | 32GB RAM | Kubernetes clusters require significant memory |
| **CPU** | 6 cores | 12 cores | Multi-core beneficial for parallel operations |
| **Disk Space** | 50GB free | 100GB free | Docker images and cluster data |
| **Operating System** | Linux, macOS, Windows | Linux/macOS | Windows requires WSL2 |

## Required Software

### Core Dependencies

| Tool | Minimum Version | Installation | Purpose |
|------|----------------|--------------|---------|
| **Docker** | 20.10+ | [Get Docker](https://docs.docker.com/get-docker/) | Container runtime for K3d clusters |
| **K3d** | 5.0+ | [Install K3d](https://k3d.io/v5.6.0/#installation) | Lightweight Kubernetes clusters |
| **kubectl** | 1.24+ | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) | Kubernetes command-line interface |
| **Helm** | 3.8+ | [Install Helm](https://helm.sh/docs/intro/install/) | Kubernetes package manager |

### OpenFrame CLI

Download the latest OpenFrame CLI for your platform:

| Platform | Download Link |
|----------|---------------|
| **Windows (AMD64)** | [openframe-cli_windows_amd64.zip](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip) |
| **Linux (AMD64)** | [openframe-cli_linux_amd64.tar.gz](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz) |
| **macOS (Intel)** | [openframe-cli_darwin_amd64.tar.gz](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz) |
| **macOS (Apple Silicon)** | [openframe-cli_darwin_arm64.tar.gz](https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz) |

### Development Tools (Optional)

For development workflows, install these additional tools:

| Tool | Purpose | Installation |
|------|---------|--------------|
| **Telepresence** | Traffic interception | [Install Telepresence](https://www.telepresence.io/docs/latest/install/) |
| **Skaffold** | Live development | [Install Skaffold](https://skaffold.dev/docs/install/) |

## Installation Steps

### 1. Install Core Dependencies

**Docker**
```bash
# Verify Docker installation
docker --version
docker ps
```

**K3d**
```bash
# Install K3d (Linux/macOS)
curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash

# Verify K3d
k3d version
```

**kubectl**
```bash
# Install kubectl (Linux)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Verify kubectl
kubectl version --client
```

**Helm**
```bash
# Install Helm (Linux/macOS)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verify Helm
helm version
```

### 2. Install OpenFrame CLI

**Linux/macOS**
```bash
# Download and extract
curl -L "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz" | tar xz

# Make executable and move to PATH
chmod +x openframe
sudo mv openframe /usr/local/bin/openframe

# Verify installation
openframe --version
```

**Windows (PowerShell)**
```bash
# Download the zip file
Invoke-WebRequest -Uri "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip" -OutFile "openframe-cli.zip"

# Extract and run
Expand-Archive -Path "openframe-cli.zip" -DestinationPath "."

# Add to PATH or run directly
./openframe.exe --version
```

## Environment Variables

Set these environment variables for optimal experience:

```bash
# Docker configuration
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration
export KUBECONFIG=$HOME/.kube/config

# Optional: Set default cluster name
export OPENFRAME_CLUSTER_NAME=openframe-dev
```

## Network Requirements

Ensure these network configurations:

- **Docker daemon** running and accessible
- **Port availability**: 80, 443, 6443, 8080-8090 (for cluster services)
- **Internet access** for downloading images and charts
- **DNS resolution** working correctly

## Verification Commands

Run these commands to verify your setup:

```bash
# Check Docker
docker run hello-world

# Check K3d
k3d cluster list

# Check kubectl
kubectl cluster-info

# Check Helm
helm repo list

# Check OpenFrame CLI
openframe --help
```

## Troubleshooting Common Issues

### Docker Issues
```bash
# Start Docker daemon (Linux)
sudo systemctl start docker

# Add user to docker group (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### K3d Issues
```bash
# Check Docker is running
docker ps

# Reset K3d if needed
k3d cluster delete --all
```

### Kubectl Issues
```bash
# Check kubeconfig
kubectl config get-contexts

# Reset kubeconfig if needed
kubectl config unset current-context
```

## Next Steps

Once prerequisites are installed:

1. **[Quick Start Guide](quick-start.md)** - Create your first cluster in 5 minutes
2. **[First Steps Guide](first-steps.md)** - Explore key features after setup

> 🔧 **Having issues?** Join our [OpenMSP Slack community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) for help and support.