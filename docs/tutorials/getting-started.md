# OpenFrame CLI - Getting Started Guide

Welcome to OpenFrame CLI! This guide will help you set up and start using the OpenFrame CLI tool for managing Kubernetes clusters and development workflows.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following tools installed:

### Required Dependencies
- **Docker** (v20.10+): Required for K3d cluster management
  - [Install Docker](https://docs.docker.com/get-docker/)
- **kubectl** (v1.21+): Kubernetes command-line tool
  - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)

### Optional but Recommended
- **Helm** (v3.8+): For advanced chart management
  - [Install Helm](https://helm.sh/docs/intro/install/)
- **git**: For source code management and development workflows

### System Requirements
- **Operating System**: macOS, Linux, or Windows
- **Architecture**: AMD64 or ARM64
- **Memory**: At least 4GB RAM (8GB recommended for development)
- **Disk Space**: At least 10GB free space

## Installation

### Option 1: Install from Release (Recommended)

Choose the command for your platform:

#### macOS (Apple Silicon)
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

#### macOS (Intel)
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

#### Linux (AMD64)
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

#### Windows (AMD64)
1. Download the Windows binary from the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest)
2. Extract the archive
3. Move `openframe.exe` to a directory in your PATH

### Option 2: Build from Source

If you have Go installed (v1.19+):

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # On Unix-like systems
```

### Verify Installation

```bash
openframe --version
```

You should see output showing the OpenFrame CLI version.

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. The tool will automatically detect your system and configure itself appropriately.

### Initial Setup Check

Verify your environment is ready:

```bash
# Check if Docker is running
docker version

# Check if kubectl is available
kubectl version --client

# Verify OpenFrame CLI can access Docker
openframe cluster list
```

## Running Your First Cluster

### Step 1: Create a Cluster

Create your first OpenFrame cluster using the interactive wizard:

```bash
openframe cluster create
```

This command will:
- Guide you through cluster configuration options
- Automatically detect your system capabilities
- Create a K3d-based Kubernetes cluster
- Configure networking and storage

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
# List all clusters
openframe cluster list

# Get detailed status of your cluster
openframe cluster status
```

You should see output similar to:
```
NAME                STATUS    WORKERS   CREATED
openframe-cluster   Running   1         2m ago
```

### Step 3: Test Kubernetes Access

Verify kubectl can connect to your cluster:

```bash
kubectl get nodes
kubectl get pods -A
```

## Your First OpenFrame Application

### Bootstrap OpenFrame Services

Install the core OpenFrame components:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install necessary Helm charts
- Set up ArgoCD for GitOps
- Configure OpenFrame services
- Provide you with access URLs

### Verify the Installation

Check that OpenFrame services are running:

```bash
kubectl get pods -n openframe
```

### Access the OpenFrame Dashboard

After bootstrapping, the CLI will display URLs to access various services. Look for output like:

```
✅ OpenFrame services are ready!

Access your services at:
- ArgoCD: http://localhost:8080
- OpenFrame Dashboard: http://localhost:3000

Use 'kubectl port-forward' if needed to access services locally.
```

## Next Steps

### Explore Available Commands

```bash
# See all available commands
openframe --help

# Get help for specific commands
openframe cluster --help
openframe bootstrap --help
```

### Development Workflow

For active development, try these commands:

```bash
# Set up development environment with Skaffold
openframe dev scaffold

# Intercept service traffic for debugging
openframe dev intercept
```

### Chart Management

```bash
# Install additional charts
openframe chart install

# See chart status
helm list -A
```

## Common Issues and Solutions

### Issue: "Docker not found" or "Docker daemon not running"

**Problem**: OpenFrame CLI cannot connect to Docker.

**Solution**:
1. Ensure Docker is installed and running
2. On Linux, make sure your user is in the `docker` group:
   ```bash
   sudo usermod -aG docker $USER
   # Log out and back in, or run:
   newgrp docker
   ```
3. On macOS/Windows, start Docker Desktop

### Issue: "kubectl: command not found"

**Problem**: kubectl is not installed or not in PATH.

**Solution**:
1. Install kubectl following the [official guide](https://kubernetes.io/docs/tasks/tools/)
2. Verify installation: `kubectl version --client`

### Issue: "Port already in use" during cluster creation

**Problem**: Required ports (80, 443, 6443) are already in use.

**Solution**:
1. Stop services using those ports, or
2. Create cluster with custom ports:
   ```bash
   openframe cluster create --api-port 6444 --http-port 8080 --https-port 8443
   ```

### Issue: "Insufficient resources" error

**Problem**: Not enough CPU/memory for cluster creation.

**Solution**:
1. Close unnecessary applications
2. Increase Docker Desktop resource limits (macOS/Windows)
3. Create a smaller cluster configuration

### Issue: Bootstrap fails with timeout errors

**Problem**: Network issues or slow image pulls.

**Solution**:
1. Check internet connection
2. Retry the bootstrap command:
   ```bash
   openframe bootstrap --deployment-mode=oss-tenant --timeout=10m
   ```
3. Pre-pull required images:
   ```bash
   docker pull argoproj/argocd:latest
   ```

### Issue: "Cluster not found" after creation

**Problem**: kubectl context is not set correctly.

**Solution**:
```bash
# List available contexts
kubectl config get-contexts

# Switch to your OpenFrame cluster context
kubectl config use-context k3d-openframe-cluster

# Or let OpenFrame CLI handle it
openframe cluster status
```

### Getting Help

If you encounter issues not covered here:

1. Check the [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
2. Run commands with `--verbose` flag for detailed output
3. Check cluster logs: `kubectl logs -n openframe <pod-name>`
4. Join the community discussions for support

## What's Next?

Now that you have OpenFrame CLI running:

- Explore the [full documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
- Learn about [development workflows](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/docs)
- Try deploying your first application
- Set up CI/CD pipelines with ArgoCD

Happy coding with OpenFrame! 🚀