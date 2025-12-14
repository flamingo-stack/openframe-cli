# Getting Started with OpenFrame CLI

Welcome to OpenFrame CLI! This guide will help you set up and start using the CLI tool to manage OpenFrame Kubernetes clusters and development workflows.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

### Required Software
- **Docker** (version 20.10 or later)
  - Download from [docker.com](https://www.docker.com/get-started)
  - Verify: `docker --version`
- **kubectl** (Kubernetes command-line tool)
  - Install via [official documentation](https://kubernetes.io/docs/tasks/tools/)
  - Verify: `kubectl version --client`

### System Requirements
- **Operating System**: macOS, Linux, or Windows
- **Memory**: At least 4GB RAM (8GB recommended for local clusters)
- **Disk Space**: 2GB free space for cluster images and data

### Optional but Recommended
- **Helm** (for advanced chart management)
- **Git** (for development workflows)

## Installation

Choose the installation method that works best for you:

### Option 1: Quick Install (Recommended)

Select your platform and run the appropriate command:

**macOS (Apple Silicon/ARM64):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**macOS (Intel/AMD64):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**Linux (AMD64):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**Windows (AMD64):**
1. Visit the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest)
2. Download `openframe-cli_windows_amd64.zip`
3. Extract and add `openframe.exe` to your PATH

### Option 2: Build from Source

If you have Go installed (version 1.19 or later):

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # On Unix systems
```

### Verify Installation

Confirm the installation was successful:

```bash
openframe --version
```

You should see output similar to:
```
OpenFrame CLI v1.0.0
```

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. However, you may want to customize a few settings:

### Docker Configuration

Ensure Docker is running and accessible:

```bash
docker ps
```

If you see an error, start Docker Desktop or the Docker daemon.

### Kubernetes Context

OpenFrame CLI will automatically manage K3d clusters, but you can check your current Kubernetes context:

```bash
kubectl config current-context
```

## Running the Project Locally

Let's create your first OpenFrame cluster and get it running:

### Step 1: Create a Local Cluster

Run the interactive cluster creation wizard:

```bash
openframe cluster create
```

The wizard will guide you through:
- Cluster name selection
- Port configuration
- Resource allocation
- Network settings

Example interaction:
```
✓ Enter cluster name (default: openframe): my-cluster
✓ Enter HTTP port (default: 8080): 8080
✓ Enter HTTPS port (default: 8443): 8443
✓ Creating cluster "my-cluster"...
✓ Cluster created successfully!
```

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
openframe cluster status
```

Expected output:
```
Cluster: my-cluster
Status: Running
Nodes: 1/1 ready
Kubernetes Version: v1.27.4+k3s1
```

### Step 3: List All Clusters

View all your clusters:

```bash
openframe cluster list
```

## First Steps - Your "Hello World"

Now that you have a running cluster, let's bootstrap OpenFrame:

### Bootstrap OpenFrame

Install the core OpenFrame components:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This will:
1. Install required Helm charts
2. Set up ArgoCD for GitOps
3. Configure core OpenFrame services
4. Validate the installation

The process takes 2-3 minutes. You'll see progress indicators:

```
✓ Installing ArgoCD...
✓ Installing OpenFrame charts...
✓ Configuring GitOps repository...
✓ Validating installation...
✓ Bootstrap complete!
```

### Access Your Installation

Once bootstrapped, access the OpenFrame dashboard:

```bash
# Get the dashboard URL
kubectl get ingress -n openframe

# Or use port-forwarding
kubectl port-forward -n openframe svc/openframe-dashboard 8080:80
```

Visit `http://localhost:8080` in your browser to see your OpenFrame installation.

## Common Issues and Solutions

### Issue 1: Docker Not Running

**Error:**
```
Error: Cannot connect to Docker daemon
```

**Solution:**
Start Docker Desktop or the Docker daemon:
- **macOS/Windows**: Start Docker Desktop application
- **Linux**: `sudo systemctl start docker`

### Issue 2: Port Already in Use

**Error:**
```
Error: Port 8080 is already in use
```

**Solution:**
Either stop the service using the port or choose different ports:
```bash
# Find what's using the port
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Or create cluster with different ports
openframe cluster create --http-port=8081 --https-port=8444
```

### Issue 3: Insufficient Memory

**Error:**
```
Error: Not enough memory to create cluster
```

**Solution:**
1. Close unnecessary applications
2. Increase Docker's memory limit (Docker Desktop → Settings → Resources)
3. Use a smaller cluster configuration

### Issue 4: kubectl Not Found

**Error:**
```
Error: kubectl command not found
```

**Solution:**
Install kubectl following the [official guide](https://kubernetes.io/docs/tasks/tools/), or use the cluster's built-in kubectl:
```bash
openframe cluster kubectl -- get nodes
```

### Issue 5: Cluster Creation Fails

**Error:**
```
Error: Failed to create cluster
```

**Solution:**
1. Clean up any partial clusters:
   ```bash
   openframe cluster cleanup
   ```
2. Ensure Docker has enough resources
3. Check Docker daemon logs for specific errors
4. Try creating with a different name:
   ```bash
   openframe cluster create --name=test-cluster
   ```

## Next Steps

Now that you have OpenFrame running locally, explore these areas:

1. **Development Workflows**: Try `openframe dev scaffold` for service development
2. **Chart Management**: Use `openframe chart install` to add more components  
3. **Documentation**: Visit the [OpenFrame docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs) for advanced topics
4. **Community**: Join discussions and contribute to the project

## Getting Help

If you encounter issues not covered here:

1. Check the [troubleshooting docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
2. Use `openframe --help` or `openframe <command> --help` for command-specific help
3. View cluster logs: `openframe cluster logs`
4. File an issue on [GitHub](https://github.com/flamingo-stack/openframe-cli/issues)

Happy coding with OpenFrame! 🚀