# Getting Started with OpenFrame CLI

Welcome to OpenFrame CLI! This guide will help you set up and start using OpenFrame CLI to manage Kubernetes clusters and development workflows.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following tools installed on your system:

### Required Dependencies

- **Docker** (version 20.10 or later)
  - [Install Docker Desktop](https://docs.docker.com/desktop/) for macOS/Windows
  - [Install Docker Engine](https://docs.docker.com/engine/install/) for Linux
- **kubectl** (Kubernetes command-line tool)
  - [Install kubectl](https://kubernetes.io/docs/tasks/tools/)
- **Helm** (version 3.0 or later)
  - [Install Helm](https://helm.sh/docs/intro/install/)

### Optional but Recommended

- **Git** - for cloning repositories and version control
- **curl** or **wget** - for downloading releases

### System Requirements

- **Operating System**: macOS, Linux, or Windows
- **Architecture**: AMD64 or ARM64
- **Memory**: At least 4GB RAM available for Docker
- **Disk Space**: At least 10GB free space

## Installation

### Option 1: Install from Release (Recommended)

Choose the appropriate command for your platform:

#### macOS (Apple Silicon/M1/M2)
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
1. Download the latest release from: https://github.com/flamingo-stack/openframe-cli/releases/latest
2. Extract the ZIP file
3. Add the extracted `openframe.exe` to your PATH

### Option 2: Build from Source

If you prefer to build from source or need the latest development version:

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the binary
go build -o openframe .

# Move to PATH (macOS/Linux)
sudo mv openframe /usr/local/bin/
```

### Verify Installation

Confirm the installation was successful:

```bash
openframe --help
```

You should see the OpenFrame CLI help output with available commands.

## Basic Configuration

### Docker Configuration

Ensure Docker is running and configured with sufficient resources:

```bash
# Check Docker is running
docker version

# Verify Docker has enough memory allocated (should be at least 4GB)
docker system info | grep "Total Memory"
```

### Kubectl Configuration

OpenFrame CLI will use your existing kubectl configuration. You can verify it's working:

```bash
# Check kubectl installation
kubectl version --client

# View current context (this will show an error if no cluster is configured yet - that's expected)
kubectl config current-context
```

## Running the Project Locally

### Step 1: Create Your First Cluster

OpenFrame CLI uses K3d to create lightweight Kubernetes clusters locally. Create your first cluster with the interactive wizard:

```bash
openframe cluster create
```

The interactive wizard will:
- Prompt for cluster name (or use default)
- Configure cluster settings
- Download and set up K3d if not already installed
- Create a local Kubernetes cluster
- Configure kubectl context

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
# List all OpenFrame clusters
openframe cluster list

# Check detailed cluster status
openframe cluster status

# Verify with kubectl
kubectl get nodes
```

### Step 3: Bootstrap OpenFrame

Install OpenFrame components on your cluster:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install necessary Helm charts
- Set up ArgoCD for GitOps
- Configure OpenFrame services
- Provide access URLs when complete

## First Steps / "Hello World"

### Basic Workflow Example

Here's a complete example of getting OpenFrame running:

```bash
# 1. Create a new cluster
openframe cluster create
# Follow the prompts (you can use defaults for most options)

# 2. Wait for cluster to be ready (usually 1-2 minutes)
openframe cluster status

# 3. Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# 4. Check that everything is running
kubectl get pods --all-namespaces

# 5. Access the ArgoCD UI (if installed)
# The bootstrap command will provide the URL and credentials
```

### Exploring Available Commands

```bash
# Get help for any command
openframe cluster --help
openframe bootstrap --help

# List all available commands
openframe --help

# Check version
openframe version
```

### Development Workflow Commands

Once you have a cluster running, explore development features:

```bash
# Scaffold a service for development (if you have a service to work on)
openframe dev scaffold

# List development tools
openframe dev --help
```

## Common Issues and Solutions

### Issue: Docker Not Running

**Error**: `Cannot connect to the Docker daemon`

**Solution**:
```bash
# Start Docker Desktop (macOS/Windows)
# Or start Docker service (Linux)
sudo systemctl start docker

# Verify Docker is running
docker ps
```

### Issue: Permission Denied When Installing

**Error**: `Permission denied` when moving binary to `/usr/local/bin/`

**Solution**:
```bash
# Use sudo for system directories
sudo mv openframe /usr/local/bin/

# Alternative: Install to user directory
mkdir -p ~/bin
mv openframe ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Issue: kubectl Not Found

**Error**: `kubectl: command not found`

**Solution**:
```bash
# Install kubectl first
# macOS with Homebrew
brew install kubectl

# Or download directly
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/$(uname -s | tr '[:upper:]' '[:lower:]')/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

### Issue: Cluster Creation Fails

**Error**: Cluster creation times out or fails

**Solution**:
```bash
# Check Docker resources
docker system info

# Clean up any existing clusters
openframe cluster cleanup

# Try creating with a different name
openframe cluster create --name my-test-cluster

# Check Docker logs for more details
docker logs k3d-my-cluster-name-server-0
```

### Issue: Port Already in Use

**Error**: `Port already in use` during cluster creation

**Solution**:
```bash
# List existing clusters
openframe cluster list

# Delete conflicting cluster
openframe cluster delete <cluster-name>

# Or use a different port
openframe cluster create --port 6444
```

### Issue: Insufficient Resources

**Error**: Cluster runs slow or components fail to start

**Solution**:
1. Increase Docker Desktop memory allocation (8GB recommended)
2. Close unnecessary applications
3. Check available disk space:
```bash
df -h
docker system df
```

### Getting Help

- **CLI Help**: Use `openframe --help` or `openframe <command> --help`
- **Documentation**: Visit the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs)
- **Issues**: Report bugs at the [GitHub repository](https://github.com/flamingo-stack/openframe-cli/issues)

## Next Steps

Now that you have OpenFrame CLI running:

1. **Explore the ArgoCD interface** - Access the URL provided after bootstrap
2. **Deploy your first application** - Try the development workflow commands
3. **Read the full documentation** - Learn about advanced features and configuration
4. **Join the community** - Contribute to the project or ask questions

You're ready to start developing with OpenFrame! 🚀