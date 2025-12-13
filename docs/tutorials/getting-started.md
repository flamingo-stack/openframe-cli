# Getting Started with OpenFrame CLI

Welcome to OpenFrame CLI! This guide will help you set up and start using the OpenFrame CLI tool for managing Kubernetes clusters and development workflows.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following installed:

### Required
- **Docker** (version 20.10 or later)
  - [Install Docker Desktop](https://www.docker.com/products/docker-desktop/) for macOS/Windows
  - Use your package manager for Linux: `sudo apt install docker.io` or `sudo yum install docker`
- **kubectl** (Kubernetes command-line tool)
  - [Installation guide](https://kubernetes.io/docs/tasks/tools/install-kubectl/)

### Optional but Recommended
- **Helm** (version 3.0 or later) - for chart management
  - [Installation guide](https://helm.sh/docs/intro/install/)
- **Go** (version 1.19 or later) - only needed if building from source
  - [Download Go](https://golang.org/dl/)

### System Requirements
- **Operating System**: macOS, Linux, or Windows
- **Architecture**: AMD64 or ARM64
- **Memory**: At least 4GB RAM recommended
- **Disk Space**: 2GB free space for clusters and images

## Installation

### Option 1: Install from Release (Recommended)

Choose the command for your platform:

#### macOS
```bash
# For Apple Silicon (M1/M2)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# For Intel Macs
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

#### Linux
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

#### Windows
1. Download the Windows release from: https://github.com/flamingo-stack/openframe-cli/releases/latest
2. Extract the archive
3. Add the `openframe.exe` to your PATH

### Option 2: Install from Source

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the binary
go build -o openframe .

# Move to your PATH (Linux/macOS)
sudo mv openframe /usr/local/bin/
```

### Verify Installation

```bash
openframe --help
```

You should see the OpenFrame CLI help output with available commands.

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. The tool will automatically:

- Detect your system architecture and operating system
- Configure Docker integration for K3d clusters
- Set up kubectl contexts for cluster management

### Initial Setup Check

Run a system check to ensure everything is properly configured:

```bash
# Check if Docker is running
docker ps

# Verify kubectl is installed
kubectl version --client

# Check OpenFrame CLI version
openframe --version
```

## Running Your First Cluster

Let's create your first OpenFrame cluster and get it running locally.

### Step 1: Create a Cluster

```bash
openframe cluster create
```

This command will:
- Launch an interactive wizard to guide you through cluster creation
- Create a local K3d Kubernetes cluster
- Configure networking and storage
- Set up kubectl context

### Step 2: Verify Cluster Status

```bash
# List all clusters
openframe cluster list

# Check detailed cluster status
openframe cluster status

# Verify with kubectl
kubectl get nodes
```

### Step 3: Bootstrap OpenFrame

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This will install the core OpenFrame components on your cluster.

## Your First "Hello World"

Now let's deploy a simple application to verify everything is working:

### Step 1: Check Cluster is Ready

```bash
openframe cluster status
```

Look for status indicators showing the cluster is healthy.

### Step 2: Deploy a Test Application

```bash
# Create a simple nginx deployment
kubectl create deployment hello-world --image=nginx:latest

# Expose it as a service
kubectl expose deployment hello-world --port=80 --type=NodePort

# Check the deployment
kubectl get pods
kubectl get services
```

### Step 3: Access Your Application

```bash
# Get the service details
kubectl get service hello-world

# Port-forward to access locally
kubectl port-forward service/hello-world 8080:80
```

Open your browser to `http://localhost:8080` to see the nginx welcome page.

### Step 4: Clean Up

```bash
# Remove the test deployment
kubectl delete deployment hello-world
kubectl delete service hello-world
```

## Common Issues and Solutions

### Issue: "Docker daemon not running"

**Symptoms**: Error messages about Docker not being available

**Solution**:
```bash
# Start Docker Desktop (macOS/Windows)
# Or start Docker service (Linux)
sudo systemctl start docker

# Verify Docker is running
docker ps
```

### Issue: "kubectl not found"

**Symptoms**: Command not found errors when running kubectl

**Solution**:
```bash
# Install kubectl using package manager
# macOS with Homebrew
brew install kubectl

# Linux with snap
sudo snap install kubectl --classic

# Or download binary directly
# See: https://kubernetes.io/docs/tasks/tools/install-kubectl/
```

### Issue: "Permission denied" on macOS/Linux

**Symptoms**: Cannot execute openframe binary

**Solution**:
```bash
# Make the binary executable
chmod +x openframe

# Move to a directory in your PATH
sudo mv openframe /usr/local/bin/
```

### Issue: Cluster creation fails

**Symptoms**: Errors during `openframe cluster create`

**Solutions**:
```bash
# Check Docker has enough resources (4GB+ RAM recommended)
# Clean up any existing clusters
openframe cluster cleanup

# Try creating with explicit configuration
openframe cluster create --name=my-cluster

# Check Docker Desktop settings if on macOS/Windows
```

### Issue: "Port already in use"

**Symptoms**: Cannot bind to ports during cluster creation

**Solution**:
```bash
# Check what's using the ports (typically 6443, 8080, 8443)
lsof -i :6443
lsof -i :8080

# Stop conflicting services or use different ports
openframe cluster delete
openframe cluster create  # Will use different ports automatically
```

### Issue: kubectl context not set

**Symptoms**: kubectl commands don't work after cluster creation

**Solution**:
```bash
# List available contexts
kubectl config get-contexts

# Set the correct context (usually k3d-<cluster-name>)
kubectl config use-context k3d-openframe

# Verify
kubectl cluster-info
```

## Next Steps

Now that you have OpenFrame CLI running:

1. **Explore Commands**: Run `openframe --help` to see all available commands
2. **Install Charts**: Use `openframe chart install` to add Helm charts
3. **Development Workflow**: Try `openframe dev scaffold` for service development
4. **Read Documentation**: Check the [full documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)

## Getting Help

- **CLI Help**: `openframe --help` or `openframe <command> --help`
- **Documentation**: [OpenFrame Docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
- **Issues**: [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues)
- **Community**: [Contributing Guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md)

Happy coding with OpenFrame! 🚀