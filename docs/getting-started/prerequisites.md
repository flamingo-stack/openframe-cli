# Prerequisites

Before installing OpenFrame CLI, ensure your system meets the following requirements for optimal performance and compatibility.

## System Requirements

### Hardware Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| **RAM** | 24GB | 32GB |
| **CPU Cores** | 6 cores | 12 cores |
| **Disk Space** | 50GB free | 100GB free |
| **Network** | Stable internet connection | High-speed broadband |

> **Note**: Kubernetes clusters consume significant resources. The recommended specifications ensure smooth operation with multiple services and development workflows.

### Operating System Support

| OS | Version | Notes |
|----|---------|-------|
| **macOS** | 10.15+ (Catalina+) | Full feature parity as of v0.5.10 |
| **Linux** | Ubuntu 20.04+, RHEL 8+, Fedora 35+ | Primary development platform |
| **Windows** | Windows 10/11 with WSL2 | CLI compatibility achieved in v0.3.7 |

## Required Dependencies

OpenFrame CLI automatically detects and installs most dependencies, but some must be installed manually.

### Essential Tools

| Tool | Version | Installation | Verification |
|------|---------|--------------|-------------|
| **Docker** | 20.10+ | [Docker Desktop](https://docker.com) | `docker --version` |
| **Git** | 2.30+ | System package manager | `git --version` |
| **Go** (for development) | 1.21+ | [golang.org](https://golang.org) | `go version` |

### Auto-Installed Tools

The following tools are automatically installed by OpenFrame CLI when needed:

- **K3D** - Kubernetes cluster management
- **Helm** - Package management for Kubernetes
- **kubectl** - Kubernetes command-line tool
- **ArgoCD CLI** - GitOps workflow management
- **Telepresence** - Service mesh development
- **mkcert** - Local certificate authority

## Network Requirements

### Firewall and Port Access

OpenFrame CLI requires access to the following ports:

| Port Range | Purpose | Protocol |
|------------|---------|----------|
| 80, 443 | HTTP/HTTPS traffic | TCP |
| 6443 | Kubernetes API server | TCP |
| 8080-8090 | Development services | TCP |
| 30000-32767 | NodePort services | TCP |

### Internet Connectivity

The following external services must be accessible:

```text
- docker.io (Container registry)
- registry.k8s.io (Kubernetes images)  
- github.com (Source repositories)
- helm.sh (Helm charts)
- argoproj.io (ArgoCD resources)
```

## Environment Variables

Set the following environment variables for optimal operation:

### Required Variables

```bash
# Docker Desktop or Docker Engine
export DOCKER_HOST=unix:///var/run/docker.sock

# Kubernetes configuration directory
export KUBECONFIG="$HOME/.kube/config"
```

### Optional Configuration

```bash
# Custom cluster configuration
export OPENFRAME_CLUSTER_NAME="my-openframe"
export OPENFRAME_NAMESPACE="openframe-system"

# Development settings
export OPENFRAME_DEV_MODE="true"
export OPENFRAME_LOG_LEVEL="debug"
```

## User Account Requirements

### Permissions

Your user account needs the following permissions:

- **Docker**: Ability to run Docker containers
- **File System**: Write access to `/tmp` and `$HOME/.openframe/`
- **Network**: Bind to localhost ports for development

### macOS Specific

```bash
# Grant Docker socket access
sudo chmod 666 /var/run/docker.sock

# Allow mkcert to install certificates
sudo security add-trusted-cert -d system -k /System/Library/Keychains/SystemRootCertificates.keychain
```

### Linux Specific

```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Refresh group membership (or logout/login)
newgrp docker
```

### Windows/WSL2 Specific

```powershell
# Enable WSL2 and install Docker Desktop
wsl --install
wsl --set-default-version 2

# Configure WSL2 integration in Docker Desktop settings
```

## Verification Commands

Run these commands to verify your system is ready:

### Basic System Check

```bash
# Check system resources
echo "RAM: $(free -h | awk '/^Mem/ {print $2}' 2>/dev/null || echo 'N/A')"
echo "CPU: $(nproc 2>/dev/null || echo 'N/A') cores"
echo "Disk: $(df -h . | awk 'NR==2 {print $4}' 2>/dev/null || echo 'N/A') available"
```

### Docker Verification

```bash
# Verify Docker is running
docker run --rm hello-world

# Check Docker version and system info
docker --version
docker system info | grep -E "(Server Version|Memory|CPUs)"
```

### Network Connectivity

```bash
# Test external connectivity
curl -s https://docker.io/v2/ > /dev/null && echo "✓ Docker Hub accessible"
curl -s https://github.com > /dev/null && echo "✓ GitHub accessible"
curl -s https://helm.sh > /dev/null && echo "✓ Helm repository accessible"
```

## Troubleshooting Common Issues

### Docker Permission Denied

```bash
# Solution: Add user to docker group
sudo usermod -aG docker $USER
newgrp docker
```

### Insufficient Resources

If you encounter resource constraints:

1. **Close unnecessary applications**
2. **Increase Docker resource limits** in Docker Desktop settings
3. **Use smaller cluster configurations** with fewer nodes

### Network Proxy Issues

For corporate networks with proxies:

```bash
# Configure Docker proxy
mkdir -p ~/.docker
cat > ~/.docker/config.json << EOF
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy.company.com:8080",
      "httpsProxy": "http://proxy.company.com:8080"
    }
  }
}
EOF
```

## Development Environment (Optional)

For contributors and advanced users who want to build from source:

### Additional Tools

```bash
# Go development tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/goreleaser/goreleaser@latest

# Testing and validation
go install github.com/vektra/mockery/v2@latest
```

### IDE Recommendations

- **Visual Studio Code** with Go extension
- **GoLand** by JetBrains
- **Vim/Neovim** with vim-go plugin

## Ready to Install?

Once your system meets these prerequisites, proceed to the [Quick Start Guide](quick-start.md) for installation and initial setup.

If you encounter any issues during prerequisite setup, visit the OpenMSP community at https://www.openmsp.ai/ for assistance.