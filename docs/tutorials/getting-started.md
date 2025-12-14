# Getting Started with OpenFrame CLI

OpenFrame CLI is a modern command-line tool that simplifies the management of OpenFrame Kubernetes clusters and development workflows. This guide will walk you through setting up your first OpenFrame cluster.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

### Required
- **Docker** (version 20.10 or later)
  - [Install Docker Desktop](https://docs.docker.com/get-docker/) for your platform
  - Verify installation: `docker --version`

- **kubectl** (Kubernetes command-line tool)
  - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
  - Verify installation: `kubectl version --client`

### Optional but Recommended
- **Helm** (version 3.0 or later) - for chart management
  - [Install Helm](https://helm.sh/docs/intro/install/)
  - Verify installation: `helm version`

- **Git** - for cloning repositories and development workflows
  - [Install Git](https://git-scm.com/downloads)

## Installation

### Option 1: Download Pre-built Binary (Recommended)

Choose the appropriate command for your platform:

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
3. Add the `openframe.exe` to your PATH

### Option 2: Build from Source

If you have Go installed (version 1.19 or later):

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # On Unix systems
```

### Verify Installation

```bash
openframe --version
```

You should see the version information displayed.

## Basic Configuration

OpenFrame CLI works out of the box without additional configuration, but you can customize its behavior:

### Environment Variables

Set these optional environment variables to customize behavior:

```bash
# Set default cluster name prefix
export OPENFRAME_CLUSTER_PREFIX="my-dev"

# Set default namespace
export OPENFRAME_NAMESPACE="openframe-system"
```

### Docker Configuration

Ensure Docker is running and accessible:

```bash
# Test Docker connectivity
docker ps

# If using Docker Desktop, ensure Kubernetes is enabled (optional)
```

## Running Your First OpenFrame Cluster

### Step 1: Create a Cluster

Start the interactive cluster creation wizard:

```bash
openframe cluster create
```

The wizard will guide you through:
- Cluster naming
- Port configuration
- Resource allocation
- Network settings

**Example interaction:**
```
? Enter cluster name: my-first-cluster
? Select cluster type: Development
? Enable ingress? Yes
? API server port: 6443
Creating cluster 'my-first-cluster'...
✓ Cluster created successfully
```

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
# List all clusters
openframe cluster list

# Get detailed status
openframe cluster status my-first-cluster
```

Expected output:
```
Name: my-first-cluster
Status: Running
Nodes: 1
Kubernetes Version: v1.27.3+k3s1
API Endpoint: https://127.0.0.1:6443
```

### Step 3: Bootstrap OpenFrame

Install the OpenFrame platform on your cluster:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This will:
- Install necessary Helm charts
- Set up ArgoCD for GitOps
- Configure ingress and networking
- Deploy core OpenFrame components

## Your First "Hello World"

Once your cluster is bootstrapped, let's verify everything is working:

### 1. Check Pod Status

```bash
# Check all OpenFrame pods are running
kubectl get pods -A

# Specifically check OpenFrame namespace
kubectl get pods -n openframe-system
```

### 2. Access the Dashboard

```bash
# Get the dashboard URL
openframe cluster status my-first-cluster

# Port-forward to access locally (if needed)
kubectl port-forward -n openframe-system svc/openframe-dashboard 8080:80
```

Open http://localhost:8080 in your browser to see the OpenFrame dashboard.

### 3. Deploy a Sample Application

Create a simple test deployment:

```bash
# Create a test namespace
kubectl create namespace hello-world

# Deploy a simple nginx pod
kubectl run hello-nginx --image=nginx --port=80 -n hello-world

# Expose it as a service
kubectl expose pod hello-nginx --port=80 --target-port=80 --type=ClusterIP -n hello-world

# Verify it's running
kubectl get pods -n hello-world
```

## Common Issues and Solutions

### Issue 1: Docker Not Running

**Error:** `Cannot connect to the Docker daemon`

**Solution:**
```bash
# Start Docker Desktop or Docker service
# On macOS/Windows: Start Docker Desktop application
# On Linux:
sudo systemctl start docker
```

### Issue 2: Port Already in Use

**Error:** `Port 6443 is already in use`

**Solution:**
```bash
# Check what's using the port
lsof -i :6443

# Either stop the conflicting service or use a different port
openframe cluster create --api-port 6444
```

### Issue 3: Kubectl Context Issues

**Error:** `Unable to connect to cluster`

**Solution:**
```bash
# List available contexts
kubectl config get-contexts

# Switch to the correct context
kubectl config use-context k3d-my-first-cluster

# Or let OpenFrame set it for you
openframe cluster status my-first-cluster
```

### Issue 4: Bootstrap Hangs or Fails

**Error:** Bootstrap process appears stuck

**Solution:**
```bash
# Check cluster resources
kubectl top nodes
kubectl get events --sort-by=.metadata.creationTimestamp

# If cluster is low on resources, try:
openframe cluster delete my-first-cluster
openframe cluster create --memory 4g --cpus 2
```

### Issue 5: Permission Denied (macOS/Linux)

**Error:** `Permission denied` when moving binary

**Solution:**
```bash
# Create local bin directory if it doesn't exist
mkdir -p ~/.local/bin

# Move binary there instead
mv openframe ~/.local/bin/

# Add to PATH in your shell profile
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Next Steps

Now that you have OpenFrame CLI running:

1. **Explore the documentation**: Check the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

2. **Try development workflows**:
   ```bash
   openframe dev scaffold --help
   openframe dev intercept --help
   ```

3. **Manage charts and applications**:
   ```bash
   openframe chart install --help
   ```

4. **Join the community**: Visit the [project repository](https://github.com/flamingo-stack/openframe-cli) for updates and support

## Getting Help

- Run `openframe --help` for command reference
- Use `openframe [command] --help` for specific command help
- Check the [issues page](https://github.com/flamingo-stack/openframe-cli/issues) for known problems
- Review the [contributing guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md) to get involved

Happy coding with OpenFrame! 🚀