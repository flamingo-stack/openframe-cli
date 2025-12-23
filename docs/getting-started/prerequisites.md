# Prerequisites

Before installing OpenFrame CLI, ensure your system meets the requirements and has the necessary dependencies installed.

## System Requirements

### Minimum Requirements

| Resource | Requirement |
|----------|-------------|
| **RAM** | 24GB |
| **CPU** | 6 cores |
| **Disk Space** | 50GB free |
| **Operating System** | Windows 10+, macOS 10.15+, Ubuntu 18.04+, or equivalent Linux |

### Recommended Requirements

| Resource | Recommendation |
|----------|----------------|
| **RAM** | 32GB |
| **CPU** | 12 cores |
| **Disk Space** | 100GB free |
| **Operating System** | Latest stable versions |

> **⚠️ Important**: These requirements ensure smooth operation of Kubernetes clusters and all development tools. Lower specifications may result in performance issues.

## Required Software Dependencies

The OpenFrame CLI automatically checks and can install missing prerequisites, but you may want to install them manually for better control.

### Core Dependencies

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Docker** | 20.10+ | Container runtime for K3d clusters | [Install Docker](https://docs.docker.com/get-docker/) |
| **kubectl** | 1.24+ | Kubernetes command-line tool | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| **Helm** | 3.8+ | Kubernetes package manager | [Install Helm](https://helm.sh/docs/intro/install/) |
| **K3d** | 5.4+ | Lightweight Kubernetes in Docker | [Install K3d](https://k3d.io/) |

### Development Tools (Optional)

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| **Telepresence** | 2.10+ | Local service intercepts | [Install Telepresence](https://www.telepresence.io/docs/latest/install/) |
| **jq** | 1.6+ | JSON processing for CLI operations | [Install jq](https://stedolan.github.io/jq/download/) |
| **Git** | 2.30+ | Version control for GitOps workflows | [Install Git](https://git-scm.com/downloads) |

## Platform-Specific Installation

### Windows

1. **Download the installer** from the [releases page](https://github.com/flamingo-stack/openframe-cli/releases)
2. **Run the installer** as Administrator
3. **Add to PATH** (installer does this automatically)
4. **Verify installation** in Command Prompt or PowerShell:

```bash
openframe --version
```

### macOS

#### Using Homebrew (Recommended)
```bash
# Add the OpenFrame tap (if available)
brew tap flamingo-stack/openframe

# Install OpenFrame CLI
brew install openframe-cli
```

#### Manual Installation
1. Download the macOS binary from releases
2. Move to `/usr/local/bin/`:

```bash
sudo mv openframe-cli /usr/local/bin/openframe
sudo chmod +x /usr/local/bin/openframe
```

### Linux

#### Using Package Manager (Ubuntu/Debian)
```bash
# Download the .deb package
wget https://github.com/flamingo-stack/openframe-cli/releases/download/v1.0.0/openframe-cli_1.0.0_linux_amd64.deb

# Install the package
sudo dpkg -i openframe-cli_1.0.0_linux_amd64.deb
```

#### Manual Installation
```bash
# Download and extract
wget https://github.com/flamingo-stack/openframe-cli/releases/download/v1.0.0/openframe-cli_1.0.0_linux_amd64.tar.gz
tar -xzf openframe-cli_1.0.0_linux_amd64.tar.gz

# Move to system PATH
sudo mv openframe /usr/local/bin/
sudo chmod +x /usr/local/bin/openframe
```

## Environment Variables

OpenFrame CLI uses these environment variables for configuration:

| Variable | Purpose | Default | Example |
|----------|---------|---------|---------|
| `KUBECONFIG` | Kubernetes config file location | `~/.kube/config` | `/home/user/.kube/config` |
| `OPENFRAME_LOG_LEVEL` | Logging verbosity | `info` | `debug` |
| `OPENFRAME_CLUSTER_NAME` | Default cluster name | `openframe-local` | `my-dev-cluster` |
| `DOCKER_HOST` | Docker daemon connection | System default | `unix:///var/run/docker.sock` |

### Setting Environment Variables

#### Windows (PowerShell)
```bash
$env:KUBECONFIG = "C:\Users\username\.kube\config"
$env:OPENFRAME_LOG_LEVEL = "debug"
```

#### macOS/Linux (Bash)
```bash
export KUBECONFIG=~/.kube/config
export OPENFRAME_LOG_LEVEL=debug
```

To make these permanent, add them to your shell profile (`.bashrc`, `.zshrc`, etc.).

## Verification Commands

After installation, verify your setup with these commands:

### Verify OpenFrame CLI Installation
```bash
# Check version
openframe --version

# Verify command structure
openframe --help
```

### Verify Dependencies
```bash
# Check Docker
docker --version
docker ps

# Check kubectl
kubectl version --client

# Check Helm
helm version

# Check K3d
k3d --version
```

### Verify System Resources
```bash
# Check available memory (Linux/macOS)
free -h

# Check CPU cores (Linux/macOS)
nproc

# Check disk space
df -h
```

## Network Requirements

Ensure your system can access these external resources:

| Resource | Purpose | Ports |
|----------|---------|-------|
| **Docker Hub** | Container image pulls | 443 (HTTPS) |
| **GitHub** | Repository access | 443 (HTTPS), 22 (SSH) |
| **Kubernetes API** | Cluster communication | 6443, 8443 |
| **ArgoCD** | GitOps operations | 443 (HTTPS) |

### Firewall Considerations

If behind a corporate firewall, ensure these outbound connections are allowed:
- Docker registry access (docker.io, gcr.io, quay.io)
- Helm chart repositories
- Git repositories for ArgoCD
- Telepresence relay services

## Troubleshooting Common Issues

### Docker Permission Issues (Linux)
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in, then test
docker ps
```

### kubectl Connection Issues
```bash
# Check cluster connection
kubectl cluster-info

# Verify config
kubectl config current-context
```

### Resource Constraints
If you experience performance issues:
1. Close unnecessary applications
2. Increase Docker resource limits
3. Consider using a smaller K3d cluster configuration

## What's Next?

Once you've verified all prerequisites are met:

1. **[Quick Start](quick-start.md)** - Bootstrap your first OpenFrame environment
2. **[First Steps](first-steps.md)** - Learn essential OpenFrame operations

> **💡 Pro Tip**: The OpenFrame CLI includes an automatic prerequisite checker that runs before major operations. It will guide you through installing any missing dependencies.