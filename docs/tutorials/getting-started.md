# Getting Started with OpenFrame CLI

OpenFrame CLI is a modern command-line tool that simplifies managing Kubernetes clusters and development workflows. This guide will help you install, configure, and run your first OpenFrame cluster.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following tools installed on your system:

### Required Dependencies

- **Docker**: Required for running K3d clusters
  - [Install Docker Desktop](https://www.docker.com/products/docker-desktop/) (macOS/Windows)
  - [Install Docker Engine](https://docs.docker.com/engine/install/) (Linux)
- **kubectl**: Kubernetes command-line tool
  ```bash
  # macOS (using Homebrew)
  brew install kubectl
  
  # Linux
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
  
  # Windows (using Chocolatey)
  choco install kubernetes-cli
  ```

### Optional but Recommended

- **Helm**: Package manager for Kubernetes (automatically installed by OpenFrame if missing)
- **Git**: For cloning repositories and version control

### System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: ARM64 or AMD64
- **RAM**: Minimum 4GB available for Docker
- **Disk Space**: At least 2GB free space

## Installation

### Option 1: Install from Release (Recommended)

Choose the appropriate command for your platform:

**macOS (Apple Silicon)**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**Linux (AMD64)**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**Windows (AMD64)**
1. Download the Windows release from [GitHub Releases](https://github.com/flamingo-stack/openframe-cli/releases/latest)
2. Extract the archive
3. Add the executable to your PATH

### Option 2: Build from Source

If you have Go installed and want the latest development version:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # Optional: install globally
```

### Verify Installation

Confirm OpenFrame CLI is installed correctly:

```bash
openframe --version
```

You should see output similar to:
```
OpenFrame CLI v1.0.0
```

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. However, you can customize its behavior:

### Environment Variables

Set these environment variables if needed:

```bash
# Custom cluster name prefix (optional)
export OPENFRAME_CLUSTER_PREFIX="my-dev"

# Custom Docker registry (optional)
export OPENFRAME_REGISTRY="your-registry.com"
```

### Verify Dependencies

Check that all required tools are available:

```bash
# Check Docker
docker --version

# Check kubectl
kubectl version --client

# OpenFrame will automatically check dependencies when you run commands
openframe cluster create --dry-run
```

## Running Your First OpenFrame Cluster

### Step 1: Create a Cluster

Start by creating your first K3d cluster:

```bash
openframe cluster create
```

OpenFrame will guide you through an interactive setup process:

1. **Cluster Name**: Choose a name (or use the default)
2. **Port Configuration**: Select ports for services
3. **Resource Allocation**: Configure CPU/memory limits

Example output:
```
🎯 Creating OpenFrame cluster...
📋 Cluster name: openframe-dev
🔧 Configuring K3d cluster with 1 server, 2 agents
✅ Cluster 'openframe-dev' created successfully
🔌 Cluster accessible at: https://localhost:6443
```

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
openframe cluster status
```

Expected output:
```
📊 Cluster Status: openframe-dev
┌─────────────────┬─────────────────┬──────────┬─────────┐
│ NODE            │ STATUS          │ ROLES    │ VERSION │
├─────────────────┼─────────────────┼──────────┼─────────┤
│ k3d-openframe-dev-server-0  │ Ready    │ control-plane,master │ v1.27.4+k3s1 │
│ k3d-openframe-dev-agent-0   │ Ready    │ <none>               │ v1.27.4+k3s1 │
│ k3d-openframe-dev-agent-1   │ Ready    │ <none>               │ v1.27.4+k3s1 │
└─────────────────┴─────────────────┴──────────┴─────────┘
```

### Step 3: Bootstrap OpenFrame

Install the OpenFrame platform on your cluster:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install Helm charts
- Set up ArgoCD for GitOps
- Configure ingress and networking
- Deploy core OpenFrame services

The process takes 3-5 minutes. You'll see progress updates:
```
🚀 Bootstrapping OpenFrame...
📦 Installing Helm charts...
🎯 Configuring ArgoCD...
✅ Bootstrap completed successfully
🌐 OpenFrame UI available at: https://openframe.local
```

## "Hello World" - Your First Application

Now that OpenFrame is running, let's deploy a simple application:

### Step 1: Access the OpenFrame UI

Open your browser and navigate to the URL shown in the bootstrap output (typically `https://openframe.local`).

### Step 2: Deploy a Sample App

Use the CLI to deploy a sample application:

```bash
# Create a simple nginx deployment
kubectl create deployment hello-openframe --image=nginx:latest

# Expose it as a service
kubectl expose deployment hello-openframe --port=80 --type=LoadBalancer

# Check the deployment
kubectl get pods,services
```

### Step 3: Access Your Application

```bash
# Get the service URL
kubectl get service hello-openframe

# Port-forward to access locally
kubectl port-forward service/hello-openframe 8080:80
```

Visit `http://localhost:8080` to see your running application.

## Next Steps

Now that you have OpenFrame running, explore these features:

- **Development Workflow**: Try `openframe dev scaffold` for rapid development
- **Service Mesh**: Intercept traffic with `openframe dev intercept`
- **Multiple Clusters**: Create additional clusters with `openframe cluster create`
- **Chart Management**: Install additional services with `openframe chart install`

## Common Issues and Solutions

### Issue: Docker Not Running
**Error**: `Cannot connect to the Docker daemon`

**Solution**:
```bash
# Start Docker Desktop (macOS/Windows)
# Or start Docker service (Linux)
sudo systemctl start docker
```

### Issue: Port Already in Use
**Error**: `Port 6443 is already allocated`

**Solution**:
```bash
# Check what's using the port
sudo lsof -i :6443

# Use a different port range
openframe cluster create --api-port=6444
```

### Issue: Insufficient Resources
**Error**: `insufficient memory` or cluster fails to start

**Solution**:
```bash
# Check Docker resources
docker system df

# Increase Docker memory in Docker Desktop settings
# Or clean up unused resources
docker system prune -a
```

### Issue: kubectl Not Found
**Error**: `kubectl: command not found`

**Solution**:
```bash
# Install kubectl (see Prerequisites section)
# Or use the kubectl bundled with Docker Desktop
alias kubectl="docker run --rm -i -t -v ~/.kube:/root/.kube -v $(pwd):/workspace -w /workspace bitnami/kubectl:latest"
```

### Issue: Cluster Creation Hangs
**Error**: Cluster creation process appears stuck

**Solution**:
```bash
# Cancel the process (Ctrl+C)
# Clean up any partial installation
openframe cluster cleanup

# Try creating with debug output
openframe cluster create --verbose
```

### Issue: Bootstrap Fails
**Error**: `failed to install charts` or ArgoCD setup fails

**Solution**:
```bash
# Check cluster resources
kubectl get nodes
kubectl get pods --all-namespaces

# Retry bootstrap with more verbose output
openframe bootstrap --deployment-mode=oss-tenant --verbose

# If still failing, recreate cluster
openframe cluster delete
openframe cluster create
```

### Getting Help

- **CLI Help**: Run `openframe --help` or `openframe [command] --help`
- **Documentation**: Visit the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
- **Issues**: Report bugs on [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **Community**: Join discussions in the project repository

### Cleanup

When you're done experimenting, clean up resources:

```bash
# Delete your cluster
openframe cluster delete

# Clean up any remaining resources
openframe cluster cleanup
```

You're now ready to start developing with OpenFrame! Check out the full documentation for advanced features and deployment options.