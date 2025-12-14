# OpenFrame CLI - Getting Started Guide

Welcome to OpenFrame CLI! This guide will help you get up and running with OpenFrame for local Kubernetes development.

## Prerequisites

Before you begin, ensure you have the following tools installed on your system:

### Required Software

- **Docker** (version 20.0 or later)
  - [Install Docker Desktop](https://www.docker.com/products/docker-desktop/)
  - Verify installation: `docker --version`

- **kubectl** (Kubernetes command-line tool)
  - [Installation guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
  - Verify installation: `kubectl version --client`

- **Helm** (version 3.0 or later)
  - [Installation guide](https://helm.sh/docs/intro/install/)
  - Verify installation: `helm version`

### System Requirements

- **Memory**: At least 4GB of available RAM
- **Disk Space**: 10GB of free disk space
- **Operating System**: macOS, Linux, or Windows

## Installation

Choose one of the installation methods below:

### Option 1: Install from Release (Recommended)

Download the pre-built binary for your platform:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Windows (AMD64) - Use PowerShell
Invoke-WebRequest -Uri "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip" -OutFile "openframe-cli.zip"
Expand-Archive -Path "openframe-cli.zip" -DestinationPath "."
# Move openframe.exe to a directory in your PATH
```

### Option 2: Build from Source

If you have Go installed (version 1.19 or later):

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # On Linux/macOS
```

### Verify Installation

```bash
openframe --help
```

You should see the OpenFrame CLI help output.

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. The tool will automatically detect your system capabilities and guide you through setup.

### Initial Setup Check

Run the system check to ensure all prerequisites are met:

```bash
openframe cluster status
```

This command will verify that Docker is running and all required tools are available.

## Running the Project Locally

### Step 1: Create Your First Cluster

Create a local Kubernetes cluster using the interactive wizard:

```bash
openframe cluster create
```

The wizard will guide you through:
- Cluster naming
- Configuration options
- Resource allocation

### Step 2: Verify Cluster Creation

Check that your cluster is running:

```bash
# List all clusters
openframe cluster list

# Check specific cluster status
openframe cluster status
```

You should see output similar to:
```
Cluster: my-openframe-cluster
Status: Running
Nodes: 1
Kubernetes Version: v1.27.x
```

### Step 3: Bootstrap OpenFrame

Install the OpenFrame platform components:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This will:
- Install necessary Helm charts
- Set up ArgoCD for GitOps
- Configure the OpenFrame platform

## First Steps - "Hello World" Equivalent

Let's deploy a simple application to verify everything is working:

### 1. Check Cluster Connectivity

```bash
kubectl get nodes
```

You should see your K3d cluster node listed.

### 2. Verify OpenFrame Installation

```bash
kubectl get pods -A
```

You should see OpenFrame components running in various namespaces.

### 3. Access ArgoCD Dashboard (Optional)

If ArgoCD was installed during bootstrap:

```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Navigate to `http://localhost:8080` and login with:
- Username: `admin`
- Password: (from the command above)

### 4. Deploy a Test Application

Create a simple test deployment:

```bash
kubectl create deployment hello-openframe --image=nginx:alpine
kubectl expose deployment hello-openframe --port=80 --type=NodePort
```

Check the deployment:

```bash
kubectl get pods
kubectl get services
```

## Development Workflow

### Scaffolding with Skaffold

For active development, use Skaffold:

```bash
openframe dev scaffold
```

### Service Interception with Telepresence

To intercept traffic for debugging:

```bash
openframe dev intercept
```

## Common Issues and Solutions

### Issue 1: Docker Not Running

**Problem**: `Error: Docker is not running`

**Solution**: 
```bash
# Start Docker Desktop or Docker daemon
# On macOS: Start Docker Desktop application
# On Linux: sudo systemctl start docker
```

### Issue 2: Insufficient Resources

**Problem**: Cluster creation fails with resource errors

**Solution**:
- Ensure Docker has at least 4GB RAM allocated
- Close unnecessary applications
- Check Docker Desktop resource settings

### Issue 3: Port Conflicts

**Problem**: `Port already in use` errors during cluster creation

**Solution**:
```bash
# Check what's using the port
lsof -i :6443  # or the specific port mentioned

# Delete existing clusters if needed
openframe cluster delete <cluster-name>
```

### Issue 4: kubectl Context Issues

**Problem**: `kubectl` commands fail or connect to wrong cluster

**Solution**:
```bash
# List available contexts
kubectl config get-contexts

# Switch to your OpenFrame cluster context
kubectl config use-context k3d-<your-cluster-name>
```

### Issue 5: Bootstrap Fails

**Problem**: `openframe bootstrap` command fails

**Solution**:
```bash
# Clean up and retry
openframe cluster cleanup
openframe cluster delete <cluster-name>
openframe cluster create
openframe bootstrap --deployment-mode=oss-tenant
```

### Issue 6: Windows Path Issues

**Problem**: Binary not found after installation on Windows

**Solution**:
1. Add the directory containing `openframe.exe` to your system PATH
2. Or run commands from the directory containing the binary
3. Restart your terminal/PowerShell after PATH changes

## Next Steps

Now that you have OpenFrame running locally, you can:

1. **Explore the Documentation**: Visit the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs) for detailed guides
2. **Deploy Applications**: Use ArgoCD to deploy and manage applications
3. **Development**: Use `openframe dev` commands for local development workflows
4. **Monitoring**: Check cluster status regularly with `openframe cluster status`

## Getting Help

- Run `openframe --help` for command documentation
- Use `openframe <command> --help` for specific command help
- Check the [GitHub repository](https://github.com/flamingo-stack/openframe-cli) for issues and discussions
- Review the [contributing guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md) if you'd like to contribute

Happy coding with OpenFrame! 🚀