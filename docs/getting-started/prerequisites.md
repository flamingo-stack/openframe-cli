# Prerequisites Guide

Before installing OpenFrame CLI, ensure your system meets all requirements and has the necessary tools installed. This guide will help you prepare your environment for a smooth installation experience.

## System Requirements

### Hardware Requirements

| Resource | Minimum | Recommended | Purpose |
|----------|---------|-------------|---------|
| **RAM** | 24GB | 32GB | Kubernetes containers, ArgoCD, and application workloads |
| **CPU Cores** | 6 cores | 12 cores | Container orchestration and concurrent processing |
| **Disk Space** | 50GB free | 100GB free | Container images, logs, persistent volumes |
| **Network** | Broadband | High-speed | Container image downloads and Git operations |

### Operating System Support

OpenFrame CLI supports all major operating systems:

| OS | Version | Architecture | Notes |
|-------|---------|--------------|-------|
| **Windows** | 10+ (64-bit) | AMD64 | PowerShell 5.1+ required |
| **macOS** | 10.15+ (Catalina) | Intel/Apple Silicon | Homebrew recommended |
| **Linux** | Modern distributions | AMD64, ARM64 | Ubuntu 20.04+, CentOS 8+, etc. |

## Required Software Dependencies

### Core Dependencies

These tools are **required** and must be installed before using OpenFrame CLI:

| Tool | Version | Purpose | Installation |
|------|---------|---------|-------------|
| **Docker** | 20.10+ | Container runtime | [Install Docker](https://docs.docker.com/get-docker/) |
| **kubectl** | 1.21+ | Kubernetes CLI | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.8+ | Package manager | [Install Helm](https://helm.sh/docs/intro/install/) |
| **Git** | 2.30+ | Version control | [Install Git](https://git-scm.com/downloads) |

### Optional Dependencies

These tools enhance functionality but are not strictly required:

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **k3d** | 5.4+ | Local cluster provider | ✅ Yes |
| **jq** | 1.6+ | JSON processing | ✅ Yes |
| **Telepresence** | 2.10+ | Service intercepts | ✅ Yes |

> **Note**: OpenFrame CLI can automatically install optional dependencies when needed, but manual installation provides better control and faster setup times.

## Platform-Specific Setup

### Windows Setup

1. **Install Docker Desktop**
   ```bash
   # Download and install Docker Desktop for Windows
   # Enable WSL 2 backend for better performance
   ```

2. **Install Windows Package Manager (Optional)**
   ```bash
   # Install winget if not available
   winget install Microsoft.WindowsPackageManager
   ```

3. **Install dependencies via winget**
   ```bash
   winget install Docker.DockerDesktop
   winget install Kubernetes.kubectl
   winget install Helm.Helm
   winget install Git.Git
   ```

### macOS Setup

1. **Install Homebrew** (if not already installed)
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

2. **Install dependencies via Homebrew**
   ```bash
   brew install docker kubectl helm git
   brew install --cask docker
   ```

3. **Start Docker Desktop**
   ```bash
   open /Applications/Docker.app
   ```

### Linux Setup

#### Ubuntu/Debian

```bash
# Update package index
sudo apt update

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install Git
sudo apt install git
```

#### CentOS/RHEL/Fedora

```bash
# Install Docker
sudo dnf install docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install Git
sudo dnf install git
```

## Account and Access Requirements

### Required Accounts

| Service | Purpose | Required |
|---------|---------|----------|
| **GitHub Account** | Source code access and releases | ✅ Yes |
| **Docker Hub** | Container image pulls (rate limits) | ⭕ Recommended |

### Optional Services

| Service | Purpose | Benefit |
|---------|---------|---------|
| **Container Registry** | Private image storage | Enhanced security |
| **Git Provider** | Custom chart repositories | Advanced workflows |
| **Cloud Provider** | Remote cluster deployment | Production usage |

## Environment Variables

### Required Variables

Set these environment variables in your shell profile:

```bash
# Add to ~/.bashrc, ~/.zshrc, or equivalent
export KUBECONFIG=~/.kube/config
export PATH=$PATH:/usr/local/bin
```

### Optional Variables

```bash
# Docker configuration
export DOCKER_BUILDKIT=1

# Kubernetes configuration  
export KUBECTL_EXTERNAL_DIFF="colordiff -N -u"

# Development configuration
export OPENFRAME_LOG_LEVEL=info
export OPENFRAME_CONFIG_DIR=~/.openframe
```

## Verification Commands

Run these commands to verify your environment is ready:

### Core Tools Verification

```bash
# Check Docker
docker --version
docker run hello-world

# Check kubectl
kubectl version --client

# Check Helm
helm version

# Check Git
git --version
```

### System Resources Verification

```bash
# Check available RAM (Linux/macOS)
free -h

# Check available disk space
df -h

# Check CPU cores
nproc
```

### Network Connectivity

```bash
# Test GitHub connectivity
curl -I https://github.com

# Test Docker Hub connectivity  
docker pull hello-world

# Test Helm repository access
helm repo add stable https://charts.helm.sh/stable
helm repo update
```

## Troubleshooting Common Issues

### Docker Issues

**Problem**: "Docker daemon not running"
```bash
# Windows/macOS: Start Docker Desktop
# Linux: Start Docker service
sudo systemctl start docker
```

**Problem**: Permission denied
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in, or:
newgrp docker
```

### kubectl Issues

**Problem**: "connection refused"
```bash
# Verify kubectl configuration
kubectl config view
kubectl config current-context
```

### Resource Issues

**Problem**: Insufficient memory
```bash
# Check memory usage
free -h
# Close unnecessary applications
# Consider upgrading hardware
```

## Ready to Install?

Once you've completed all prerequisites, you can proceed to the [Quick Start Guide](quick-start.md) to install and configure OpenFrame CLI.

### Pre-Installation Checklist

- [ ] Hardware requirements met (24GB+ RAM, 6+ cores, 50GB+ disk)
- [ ] Operating system supported
- [ ] Docker installed and running
- [ ] kubectl installed and accessible
- [ ] Helm 3.8+ installed
- [ ] Git installed and configured
- [ ] Environment variables configured
- [ ] Network connectivity verified
- [ ] User permissions configured

### Need Help?

If you encounter issues during setup:

1. **Check the [troubleshooting section](#troubleshooting-common-issues)** above
2. **Join our community**: [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
3. **Review platform-specific documentation** for detailed setup guides

---

*Prerequisites complete? Let's [get started with the quick installation](quick-start.md)!*