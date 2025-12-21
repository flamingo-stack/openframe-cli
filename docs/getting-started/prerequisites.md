# Prerequisites Guide

Before starting with OpenFrame CLI, ensure your system meets all requirements and has the necessary tools installed.

## System Requirements

### Operating System Support

| OS | Version | Status |
|-----|---------|--------|
| **macOS** | 10.15+ (Catalina or newer) | ✅ Fully Supported |
| **Ubuntu** | 18.04+ | ✅ Fully Supported |
| **Debian** | 10+ | ✅ Fully Supported |
| **CentOS/RHEL** | 8+ | ✅ Supported |
| **Windows** | 10+ (with WSL2) | ⚠️ Use WSL2 Environment |

### Hardware Requirements

| Resource | Minimum | Recommended | Notes |
|----------|---------|-------------|-------|
| **CPU** | 2 cores | 4+ cores | For local K8s clusters |
| **Memory** | 4 GB RAM | 8+ GB RAM | K3d clusters require significant memory |
| **Storage** | 5 GB free | 10+ GB free | For Docker images and cluster data |
| **Network** | Internet connection | Stable broadband | For downloading dependencies |

## Required Software

### Core Dependencies

These tools are **automatically checked and installed** by OpenFrame CLI when needed:

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Docker** | 20.0+ | Container runtime for K3d clusters | ❌ Manual |
| **K3d** | 5.0+ | Lightweight Kubernetes clusters | ✅ Auto |
| **kubectl** | 1.25+ | Kubernetes cluster management | ✅ Auto |
| **Helm** | 3.8+ | Kubernetes package manager | ✅ Auto |

### Development Tools (Optional)

These are installed automatically when using development features:

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Local development traffic interception | ✅ Auto |
| **Skaffold** | 2.0+ | Continuous development workflows | ✅ Auto |

## Manual Installation Steps

### 1. Install Docker

#### macOS
```bash
# Using Homebrew
brew install --cask docker

# Or download from https://docs.docker.com/desktop/mac/install/
```

#### Ubuntu/Debian
```bash
# Update package index
sudo apt-get update

# Install Docker
sudo apt-get install -y docker.io

# Add user to docker group
sudo usermod -aG docker $USER

# Restart session or run:
newgrp docker
```

#### CentOS/RHEL
```bash
# Install Docker
sudo dnf install -y docker

# Start and enable Docker
sudo systemctl start docker
sudo systemctl enable docker

# Add user to docker group  
sudo usermod -aG docker $USER
```

### 2. Verify Docker Installation

```bash
# Check Docker version
docker --version

# Test Docker functionality
docker run hello-world

# Verify Docker daemon is running
docker info
```

**Expected Output:**
```text
Docker version 24.0.0 or higher
Hello from Docker! (success message)
Server Version: 24.0.0 (in docker info)
```

## Environment Variables

### Required Variables

Set these in your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Docker environment (usually auto-configured)
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration directory
export KUBECONFIG=$HOME/.kube/config

# OpenFrame CLI configuration (optional)
export OPENFRAME_CONFIG_DIR=$HOME/.openframe
```

### Optional Configuration

```bash
# Default cluster name for bootstrap
export OPENFRAME_DEFAULT_CLUSTER=my-dev-cluster

# Skip interactive prompts in CI/CD
export OPENFRAME_NON_INTERACTIVE=true

# Enable verbose logging by default
export OPENFRAME_VERBOSE=true
```

## Account Requirements

### Container Registry Access

No special account setup required - OpenFrame CLI uses public container images.

### GitHub Access (Optional)

For accessing private OpenFrame repositories or contributing:

- GitHub account with SSH key configured
- Personal Access Token (PAT) for API access
- Repository access permissions

## Network Configuration

### Required Network Access

| Service | Port/Protocol | Purpose |
|---------|---------------|---------|
| **Docker Hub** | HTTPS (443) | Pulling container images |
| **GitHub** | HTTPS (443) | Downloading tools and charts |
| **K3d Registry** | TCP (5000-5100) | Local container registry |
| **Kubernetes API** | TCP (6443) | Cluster API access |

### Firewall Considerations

Ensure these port ranges are open for local development:

```bash
# K3d cluster ports
6443/tcp    # Kubernetes API server
80/tcp      # HTTP ingress
443/tcp     # HTTPS ingress
5000-5100/tcp  # Container registry
```

## Verification Commands

Run these commands to verify your system is ready:

### System Check Script

```bash
#!/bin/bash
echo "=== OpenFrame Prerequisites Check ==="

# Check Docker
echo -n "Docker: "
if docker --version >/dev/null 2>&1; then
    echo "✅ $(docker --version)"
else
    echo "❌ Not installed"
fi

# Check Docker daemon
echo -n "Docker daemon: "
if docker info >/dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
fi

# Check available memory
echo -n "Memory: "
free_mb=$(free -m | awk 'NR==2{printf "%.0f", $7}')
if [ $free_mb -ge 2000 ]; then
    echo "✅ ${free_mb}MB available"
else
    echo "⚠️ ${free_mb}MB available (recommend 2GB+)"
fi

# Check disk space
echo -n "Disk space: "
available=$(df -h ~ | awk 'NR==2{print $4}')
echo "📊 ${available} available"

echo "=== Prerequisites Check Complete ==="
```

Save as `check-prerequisites.sh` and run:

```bash
chmod +x check-prerequisites.sh
./check-prerequisites.sh
```

## Troubleshooting

### Common Issues

#### Docker Permission Denied
```bash
# Error: permission denied while trying to connect to the Docker daemon
sudo usermod -aG docker $USER
newgrp docker
# Or restart your terminal session
```

#### Port Already in Use
```bash
# Find process using port 6443
sudo lsof -i :6443

# Kill the process if safe to do so
sudo kill <PID>
```

#### Insufficient Memory
```bash
# Check current memory usage
free -h

# Close unnecessary applications
# Consider increasing Docker memory limits in Docker Desktop
```

## Next Steps

Once all prerequisites are installed and verified:

1. **[Quick Start Guide](./quick-start.md)** - Get OpenFrame running in 5 minutes
2. **[Development Environment Setup](../development/setup/environment.md)** - Configure your dev environment  
3. **[First Steps](./first-steps.md)** - Essential post-installation tasks

> **💡 Tip**: OpenFrame CLI will automatically install most tools when you first run commands. The bootstrap process includes built-in prerequisite checking!

---

**System ready?** Let's move to the [Quick Start Guide](./quick-start.md)! 🚀