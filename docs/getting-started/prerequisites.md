# Prerequisites

Before you can use OpenFrame CLI effectively, ensure your system meets the following requirements. The CLI includes smart prerequisite checking and can auto-install many tools, but some manual setup may be required.

## System Requirements

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| **RAM** | 24GB | 32GB |
| **CPU** | 6 cores | 12 cores |
| **Disk Space** | 50GB free | 100GB free |
| **Network** | Stable internet connection | High-speed broadband |

### Operating System Support

| Platform | Status | Notes |
|----------|---------|-------|
| **Linux** | ✅ Full Support | All distributions |
| **macOS** | ✅ Full Support | Intel and Apple Silicon |
| **Windows** | ✅ WSL2 Required | Windows 10/11 with WSL2 |

> **Windows Users**: OpenFrame CLI requires WSL2 (Windows Subsystem for Linux 2). The CLI automatically detects WSL environments and adapts command execution accordingly.

## Required Software

### Core Dependencies

These tools are **required** and will be auto-installed by the CLI if missing:

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Go** | 1.19+ | CLI runtime | ❌ Manual |
| **Docker** | 20.10+ | Container runtime | ❌ Manual |
| **K3D** | 5.0+ | Kubernetes clusters | ✅ Auto |
| **Helm** | 3.8+ | Package management | ✅ Auto |
| **kubectl** | 1.24+ | Kubernetes CLI | ✅ Auto |

### Development Tools (Optional)

These tools enhance the development experience:

| Tool | Version | Purpose | Auto-Install |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Service intercepts | ✅ Auto |
| **jq** | 1.6+ | JSON processing | ✅ Auto |
| **Git** | 2.30+ | Version control | ❌ Manual |

## Installation Verification

### Manual Prerequisites Check

Before running OpenFrame CLI, verify these core components are installed:

#### 1. Go Installation

```bash
go version
```

**Expected output:**
```text
go version go1.19.0 linux/amd64
```

If Go is not installed:
- **Linux/macOS**: Visit [golang.org/dl](https://golang.org/dl)
- **Windows WSL2**: Install inside WSL2 environment

#### 2. Docker Installation

```bash
docker --version
docker ps
```

**Expected output:**
```text
Docker version 20.10.21, build baeda1f
CONTAINER ID   IMAGE   COMMAND   CREATED   STATUS   PORTS   NAMES
```

If Docker is not running:
```bash
# Linux
sudo systemctl start docker
sudo systemctl enable docker

# macOS
# Start Docker Desktop application

# Windows WSL2
# Start Docker Desktop with WSL2 integration enabled
```

#### 3. WSL2 (Windows Only)

```bash
wsl --version
```

**Expected output:**
```text
WSL version: 1.0.3.0
```

If WSL2 is not installed, follow [Microsoft's WSL2 installation guide](https://docs.microsoft.com/en-us/windows/wsl/install).

## Environment Variables

OpenFrame CLI uses these environment variables when available:

| Variable | Description | Required | Example |
|----------|-------------|----------|---------|
| `HOME` | User home directory | ✅ Yes | `/home/username` |
| `PATH` | Executable search path | ✅ Yes | `/usr/local/bin:/usr/bin` |
| `KUBECONFIG` | Kubectl configuration | ❌ Optional | `~/.kube/config` |
| `DOCKER_HOST` | Docker daemon socket | ❌ Optional | `unix:///var/run/docker.sock` |

### Setting Environment Variables

**Linux/macOS:**
```bash
export KUBECONFIG=$HOME/.kube/config
```

**Windows (PowerShell):**
```powershell
$env:KUBECONFIG = "$env:USERPROFILE\.kube\config"
```

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.) for persistence.

## Network Requirements

### Required Connectivity

OpenFrame CLI needs internet access for:

- **Container Images**: Pull from GHCR, Docker Hub, Quay.io
- **Chart Repositories**: Helm charts from various sources
- **GitHub Integration**: Clone and sync repositories
- **Tool Downloads**: Auto-install prerequisites

### Firewall Considerations

Ensure these ports are accessible:

| Port | Purpose | Direction |
|------|---------|-----------|
| **80/443** | HTTPS traffic (charts, images) | Outbound |
| **22** | Git SSH operations | Outbound |
| **6443** | Kubernetes API server | Inbound |
| **8080-8090** | Development services | Inbound |

### Corporate Networks

If you're behind a corporate firewall:

1. Configure Docker daemon proxy settings
2. Set Git HTTP/HTTPS proxy configuration
3. Configure npm/yarn proxy (if applicable)
4. Ensure container registries are accessible

## Account Requirements

### GitHub Account (Recommended)

While not strictly required, a GitHub account enables:
- Private repository access for chart installations
- SSH key authentication for Git operations
- GitHub Container Registry (GHCR) access

### Container Registry Access

For production deployments, ensure access to:
- **GitHub Container Registry (GHCR)**: ghcr.io
- **Docker Hub**: hub.docker.com
- **Quay.io**: quay.io

## Verification Commands

Run these commands to verify your environment is ready:

### System Check Script

```bash
#!/bin/bash
echo "=== OpenFrame CLI Prerequisites Check ==="

# Check Go
if command -v go &> /dev/null; then
    echo "✅ Go: $(go version)"
else
    echo "❌ Go: Not found - install from https://golang.org/dl"
fi

# Check Docker
if command -v docker &> /dev/null && docker ps &> /dev/null; then
    echo "✅ Docker: $(docker --version)"
else
    echo "❌ Docker: Not found or not running"
fi

# Check system resources
echo "📊 System Resources:"
echo "   RAM: $(free -h | grep '^Mem:' | awk '{print $2}') total"
echo "   Disk: $(df -h / | awk 'NR==2 {print $4}') free"

# Check network connectivity
if curl -s --connect-timeout 5 https://github.com &> /dev/null; then
    echo "✅ Network: GitHub accessible"
else
    echo "❌ Network: Cannot reach GitHub"
fi

echo "=== Prerequisites check complete ==="
```

Save as `check-prereqs.sh`, make executable, and run:

```bash
chmod +x check-prereqs.sh
./check-prereqs.sh
```

## Troubleshooting Common Issues

### Docker Permission Denied (Linux)

```bash
sudo usermod -aG docker $USER
newgrp docker
# Or logout and login again
```

### WSL2 Integration Issues (Windows)

1. Open Docker Desktop
2. Go to Settings → Resources → WSL Integration
3. Enable integration with your WSL2 distro

### Port Conflicts

If ports 6443 or 8080-8090 are in use:
```bash
# Check which process is using a port
sudo lsof -i :6443
sudo netstat -tulpn | grep 6443

# Kill conflicting processes if needed
sudo kill -9 <process_id>
```

### Corporate Proxy Issues

Configure Docker proxy in `/etc/docker/daemon.json`:
```json
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy.corporate.com:8080",
      "httpsProxy": "http://proxy.corporate.com:8080"
    }
  }
}
```

## Next Steps

Once you've verified all prerequisites:

1. **[Quick Start Guide](quick-start.md)** - Install and run OpenFrame CLI
2. **[First Steps](first-steps.md)** - Explore core features

The OpenFrame CLI will automatically check for missing tools and attempt to install them during first run. If any critical prerequisites are missing, the CLI will provide specific installation instructions for your platform.