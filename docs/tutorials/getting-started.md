# Getting Started with OpenFrame CLI

A comprehensive guide to get you up and running with the OpenFrame CLI tool for managing Kubernetes clusters and development workflows.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following installed:

### Required Dependencies

- **Docker Desktop** (v4.0+)
  - [Download for macOS](https://desktop.docker.com/mac/main/amd64/Docker.dmg)
  - [Download for Windows](https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe)
  - [Install on Linux](https://docs.docker.com/engine/install/)

- **kubectl** (v1.24+)
  ```bash
  # macOS (via Homebrew)
  brew install kubectl
  
  # Windows (via Chocolatey)
  choco install kubernetes-cli
  
  # Linux
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl && sudo mv kubectl /usr/local/bin/
  ```

### Optional but Recommended

- **Helm** (v3.8+) - For chart management
  ```bash
  # macOS
  brew install helm
  
  # Windows
  choco install kubernetes-helm
  
  # Linux
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  ```

- **Git** - For source code management
- **Go** (v1.19+) - Only needed if building from source

## Installation

### Option 1: Install from Release (Recommended)

Choose the appropriate command for your platform:

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

# Windows (PowerShell as Administrator)
# Download manually from: https://github.com/flamingo-stack/openframe-cli/releases/latest
# Extract and add to your PATH
```

### Option 2: Build from Source

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the binary
go build -o openframe .

# Move to your PATH (macOS/Linux)
sudo mv openframe /usr/local/bin/
```

### Verify Installation

```bash
openframe version
openframe --help
```

## Basic Configuration

OpenFrame CLI uses sensible defaults and requires minimal configuration to get started.

### Initial Setup

1. **Verify Docker is running:**
   ```bash
   docker ps
   ```

2. **Check kubectl configuration:**
   ```bash
   kubectl config current-context
   ```

### Environment Variables (Optional)

You can customize behavior with these environment variables:

```bash
# Set default cluster name
export OPENFRAME_CLUSTER_NAME="my-dev-cluster"

# Set custom kubeconfig location
export KUBECONFIG="$HOME/.kube/config"

# Enable debug logging
export OPENFRAME_DEBUG=true
```

## Running Your First Cluster

### Step 1: Create a Cluster

Start the interactive cluster creation wizard:

```bash
openframe cluster create
```

This will guide you through:
- Cluster name selection
- Port configuration
- Resource allocation
- Network settings

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
# List all clusters
openframe cluster list

# Check detailed status
openframe cluster status

# Verify kubectl connectivity
kubectl get nodes
```

You should see output similar to:
```
NAME                    STATUS   ROLES                  AGE   VERSION
k3d-openframe-server-0  Ready    control-plane,master   1m    v1.25.3+k3s1
```

## Your First OpenFrame Application

### Step 1: Bootstrap OpenFrame

Install the core OpenFrame components:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install Helm charts
- Set up ArgoCD
- Configure the OpenFrame tenant

### Step 2: Verify Installation

Check that all components are running:

```bash
# Check ArgoCD status
kubectl get pods -n argocd

# Check OpenFrame components
kubectl get pods -n openframe-system

# Access ArgoCD UI (optional)
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Visit https://localhost:8080 (admin/password from ArgoCD docs)
```

### Step 3: Deploy a Sample Application

Create a simple application to test your setup:

```bash
# Create a sample deployment
kubectl create deployment hello-openframe --image=nginx:latest

# Expose the service
kubectl expose deployment hello-openframe --port=80 --type=LoadBalancer

# Check the service
kubectl get services hello-openframe
```

Access your application:
```bash
# Get the service URL
kubectl get svc hello-openframe

# Port forward to access locally
kubectl port-forward svc/hello-openframe 8081:80
# Visit http://localhost:8081
```

## Common Issues and Solutions

### Issue 1: Docker Not Running

**Symptoms:**
```
Error: Cannot connect to the Docker daemon
```

**Solution:**
1. Start Docker Desktop
2. Wait for Docker to fully initialize (green indicator)
3. Verify with `docker ps`

### Issue 2: Port Already in Use

**Symptoms:**
```
Error: port 8080 already in use
```

**Solution:**
```bash
# Find what's using the port
lsof -i :8080

# Kill the process (replace PID)
kill -9 <PID>

# Or use a different port during cluster creation
openframe cluster create --api-port 8081
```

### Issue 3: kubectl Context Issues

**Symptoms:**
```
Error: The connection to the server was refused
```

**Solution:**
```bash
# List available contexts
kubectl config get-contexts

# Switch to the correct context
kubectl config use-context k3d-<cluster-name>

# Verify connection
kubectl cluster-info
```

### Issue 4: Insufficient Resources

**Symptoms:**
```
Error: pods stuck in "Pending" state
```

**Solution:**
```bash
# Check resource usage
kubectl top nodes
kubectl describe pod <pod-name>

# Increase Docker Desktop resources:
# Docker Desktop → Settings → Resources → Advanced
# Increase CPU and Memory limits
```

### Issue 5: Chart Installation Failures

**Symptoms:**
```
Error: failed to install chart
```

**Solution:**
```bash
# Update Helm repositories
helm repo update

# Check Helm installation
helm version

# Retry bootstrap with verbose output
openframe bootstrap --deployment-mode=oss-tenant --verbose
```

## Next Steps

Now that you have OpenFrame CLI running:

1. **Explore Development Workflows:**
   ```bash
   openframe dev scaffold --help
   openframe dev intercept --help
   ```

2. **Learn Chart Management:**
   ```bash
   openframe chart install --help
   ```

3. **Read the Documentation:**
   - [OpenFrame Documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
   - [Contributing Guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md)

4. **Join the Community:**
   - Report issues on GitHub
   - Contribute improvements
   - Share your use cases

## Getting Help

- **CLI Help:** `openframe --help` or `openframe <command> --help`
- **GitHub Issues:** [Report bugs or request features](https://github.com/flamingo-stack/openframe-cli/issues)
- **Documentation:** [Full documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

Happy coding with OpenFrame! 🚀