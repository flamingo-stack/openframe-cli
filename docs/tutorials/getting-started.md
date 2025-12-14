# Getting Started with OpenFrame CLI

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. This guide will help you get up and running with your first OpenFrame cluster.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following tools installed on your system:

### Required Tools

- **Docker Desktop** (version 4.0+)
  - [Download for macOS](https://docs.docker.com/desktop/mac/install/)
  - [Download for Windows](https://docs.docker.com/desktop/windows/install/)
  - [Download for Linux](https://docs.docker.com/desktop/linux/install/)

- **kubectl** (Kubernetes command-line tool)
  ```bash
  # macOS (using Homebrew)
  brew install kubectl
  
  # Windows (using Chocolatey)
  choco install kubernetes-cli
  
  # Linux (using package manager)
  sudo apt-get install kubectl  # Ubuntu/Debian
  ```

### Optional but Recommended

- **Helm** (for advanced chart management)
  ```bash
  # macOS
  brew install helm
  
  # Windows
  choco install kubernetes-helm
  
  # Linux
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  ```

## Installation

### Option 1: Install from Release (Recommended)

Choose the command for your operating system:

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
1. Download the latest release from [GitHub Releases](https://github.com/flamingo-stack/openframe-cli/releases/latest)
2. Extract the ZIP file
3. Add the `openframe.exe` to your PATH

### Option 2: Build from Source

If you have Go installed (version 1.19+):

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # macOS/Linux
```

### Verify Installation

Check that OpenFrame CLI is installed correctly:

```bash
openframe --version
```

You should see version information displayed.

## Basic Configuration

OpenFrame CLI works out of the box with minimal configuration. The tool automatically detects your system environment and configures itself accordingly.

### Check System Requirements

Run the built-in system check to ensure everything is ready:

```bash
openframe cluster create --dry-run
```

This will validate that Docker is running and all prerequisites are met without actually creating a cluster.

## Running Your First Cluster

### Step 1: Create a Cluster

Start the interactive cluster creation wizard:

```bash
openframe cluster create
```

The wizard will guide you through:
- Cluster name selection
- Resource allocation
- Network configuration

For a quick start with default settings, you can also run:

```bash
openframe cluster create --name my-first-cluster
```

### Step 2: Verify Cluster Status

Check that your cluster is running:

```bash
openframe cluster list
```

You should see output similar to:
```
NAME               STATUS    NODES    AGE
my-first-cluster   Running   1        2m
```

Get detailed cluster information:

```bash
openframe cluster status --name my-first-cluster
```

### Step 3: Access Your Cluster

The cluster is automatically configured with `kubectl`. Test the connection:

```bash
kubectl get nodes
```

You should see your cluster node(s) listed.

## Your First OpenFrame Application

### Bootstrap OpenFrame Platform

Install the OpenFrame platform components on your cluster:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install necessary Helm charts
- Configure ArgoCD for GitOps workflows
- Set up monitoring and observability tools

### Verify the Installation

Check that OpenFrame components are running:

```bash
kubectl get pods --all-namespaces
```

Look for pods in namespaces like `openframe-system`, `argocd`, and `monitoring`.

### Access the Dashboard

Once bootstrapped, you can access the OpenFrame dashboard:

```bash
# Get the dashboard URL
openframe cluster status --name my-first-cluster
```

The status command will display URLs for accessing various services including the main dashboard.

## Essential Commands

Here are the most commonly used commands to get you started:

```bash
# Cluster management
openframe cluster list                    # List all clusters
openframe cluster status                  # Show current cluster details
openframe cluster delete --name <name>    # Delete a cluster
openframe cluster start --name <name>     # Start a stopped cluster

# Development workflows
openframe dev scaffold                    # Set up development environment
openframe dev intercept                  # Debug services with Telepresence

# Chart management
openframe chart install                  # Install additional charts

# Get help
openframe --help                         # Show all available commands
openframe cluster --help                 # Show cluster-specific commands
```

## Common Issues and Solutions

### Issue: "Docker not found" or "Docker not running"

**Solution:**
1. Ensure Docker Desktop is installed and running
2. On Linux, make sure your user is in the `docker` group:
   ```bash
   sudo usermod -aG docker $USER
   # Log out and log back in
   ```

### Issue: "kubectl: command not found"

**Solution:**
Install kubectl using the instructions in the Prerequisites section above.

### Issue: "Permission denied" when moving binary to `/usr/local/bin/`

**Solution:**
Either run the command with `sudo` or install to a user directory:

```bash
# Install to user directory
mkdir -p ~/.local/bin
mv openframe ~/.local/bin/
export PATH="$HOME/.local/bin:$PATH"  # Add to your shell profile
```

### Issue: Cluster creation fails with port conflicts

**Solution:**
Check if other applications are using ports 80, 443, or 6443:

```bash
# Find processes using these ports
lsof -i :80
lsof -i :443
lsof -i :6443

# Stop conflicting services or use different ports
openframe cluster create --api-port 6444 --http-port 8080 --https-port 8443
```

### Issue: Bootstrap fails with timeout errors

**Solution:**
1. Ensure sufficient system resources (4GB+ RAM recommended)
2. Check internet connectivity for downloading charts
3. Increase timeout values:
   ```bash
   openframe bootstrap --deployment-mode=oss-tenant --timeout=10m
   ```

### Issue: Can't access services after bootstrap

**Solution:**
1. Wait a few minutes for all pods to be ready
2. Check pod status: `kubectl get pods --all-namespaces`
3. Restart the cluster if needed: `openframe cluster start --name <cluster-name>`

## Next Steps

Now that you have OpenFrame CLI running:

1. **Explore the Documentation**: Visit the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs) for advanced features
2. **Try Development Workflows**: Use `openframe dev scaffold` to set up a development environment
3. **Deploy Your First Application**: Follow the application deployment guides in the main documentation
4. **Join the Community**: Check out the [contributing guidelines](https://github.com/flamingo-stack/openframe-oss-tenant/blob/main/CONTRIBUTING.md) to get involved

## Getting Help

- Run `openframe --help` for command-line help
- Visit the [GitHub Issues](https://github.com/flamingo-stack/openframe-cli/issues) page to report bugs or request features
- Check the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs) for detailed guides

Happy coding with OpenFrame! 🚀