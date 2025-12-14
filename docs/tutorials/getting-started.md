# Getting Started with OpenFrame CLI

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. This guide will help you get up and running with your first OpenFrame cluster.

## Prerequisites

Before installing OpenFrame CLI, ensure you have the following installed on your system:

### Required
- **Docker Desktop** (v4.0 or later)
  - [Download for macOS](https://docs.docker.com/desktop/mac/install/)
  - [Download for Windows](https://docs.docker.com/desktop/windows/install/)
  - [Download for Linux](https://docs.docker.com/desktop/linux/install/)
- **kubectl** (v1.24 or later)
  ```bash
  # macOS
  brew install kubectl
  
  # Linux
  curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
  sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
  
  # Windows (using Chocolatey)
  choco install kubernetes-cli
  ```

### Optional (but recommended)
- **Helm** (v3.8 or later) - for chart management
  ```bash
  # macOS
  brew install helm
  
  # Linux/Windows
  curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
  ```

## Installation

### Option 1: Install from Release (Recommended)

Choose the appropriate command for your platform:

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/
```

**macOS (Intel):**
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
1. Download from the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest)
2. Extract the archive
3. Add the executable to your PATH

### Option 2: Build from Source

If you have Go 1.19+ installed:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/  # macOS/Linux
```

### Verify Installation

```bash
openframe --version
```

You should see output similar to:
```
OpenFrame CLI v1.0.0
```

## Basic Configuration

OpenFrame CLI works out of the box with sensible defaults. However, you can customize the configuration:

### Check System Requirements

```bash
openframe cluster status
```

This command will verify that Docker is running and kubectl is available.

### Set Default Cluster Name (Optional)

```bash
export OPENFRAME_CLUSTER_NAME="my-cluster"
```

## Running Your First OpenFrame Cluster

### Step 1: Create a Cluster

Run the interactive cluster creation wizard:

```bash
openframe cluster create
```

The CLI will guide you through:
- Cluster name selection
- Port configuration
- Resource allocation

**Example interaction:**
```
? Enter cluster name: my-openframe-cluster
? Select cluster type: Development (k3d)
? HTTP port [8080]: 8080
? HTTPS port [8443]: 8443
✓ Creating cluster my-openframe-cluster...
✓ Cluster created successfully!
```

### Step 2: Verify Cluster Status

```bash
openframe cluster list
```

Expected output:
```
NAME                 STATUS    NODES    AGE
my-openframe-cluster running   1/1      2m
```

Check detailed status:
```bash
openframe cluster status
```

### Step 3: Bootstrap OpenFrame

Install the core OpenFrame components:

```bash
openframe bootstrap --deployment-mode=oss-tenant
```

This command will:
- Install ArgoCD
- Set up the OpenFrame application stack
- Configure ingress controllers
- Deploy monitoring tools

**Note:** The bootstrap process takes 5-10 minutes depending on your internet connection.

## Your First "Hello World"

### Access the OpenFrame Dashboard

Once bootstrapping is complete:

1. **Get the dashboard URL:**
   ```bash
   openframe cluster status
   ```
   
2. **Access the dashboard:**
   - Open your browser to `http://localhost:8080`
   - You should see the OpenFrame welcome page

### Deploy a Sample Application

1. **Create a sample app:**
   ```bash
   kubectl create deployment hello-world --image=nginx:alpine
   kubectl expose deployment hello-world --port=80 --target-port=80
   ```

2. **Access your app:**
   ```bash
   kubectl port-forward service/hello-world 8081:80
   ```
   
3. **Open http://localhost:8081** in your browser to see the nginx welcome page.

### Explore Available Commands

```bash
# List all available commands
openframe --help

# Get help for specific commands
openframe cluster --help
openframe bootstrap --help
```

## Common Issues and Solutions

### Issue: "Docker not found" error

**Problem:** Docker Desktop is not running or not installed.

**Solution:**
1. Install Docker Desktop from the official website
2. Start Docker Desktop
3. Verify with: `docker version`

### Issue: "kubectl not found" error

**Problem:** kubectl is not installed or not in PATH.

**Solution:**
```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install kubectl /usr/local/bin/
```

### Issue: Port already in use

**Problem:** Default ports (8080, 8443) are already occupied.

**Solution:**
```bash
# Create cluster with custom ports
openframe cluster create --http-port=8090 --https-port=8453
```

### Issue: Bootstrap fails with timeout

**Problem:** Slow internet or resource constraints.

**Solution:**
1. Ensure stable internet connection
2. Increase Docker Desktop resources (4GB+ RAM recommended)
3. Retry the bootstrap:
   ```bash
   openframe bootstrap --deployment-mode=oss-tenant --timeout=20m
   ```

### Issue: Permission denied when moving binary

**Problem:** Insufficient permissions for installation.

**Solution:**
```bash
# Create local bin directory
mkdir -p ~/.local/bin
mv openframe ~/.local/bin/

# Add to PATH in your shell profile
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

### Issue: Cluster creation hangs

**Problem:** Network or Docker configuration issues.

**Solution:**
1. Check Docker Desktop status
2. Clean up any existing clusters:
   ```bash
   openframe cluster cleanup
   ```
3. Restart Docker Desktop and try again

## Next Steps

Now that you have OpenFrame running:

1. **Explore the Documentation:** Visit the [OpenFrame documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs) for advanced usage
2. **Set up Development Workflow:** Try `openframe dev scaffold` for application development
3. **Install Additional Charts:** Use `openframe chart install` to add more services
4. **Join the Community:** Check the project repository for contribution guidelines

## Getting Help

- **CLI Help:** `openframe --help` or `openframe [command] --help`
- **Documentation:** [GitHub Repository](https://github.com/flamingo-stack/openframe-cli)
- **Issues:** [Report bugs or request features](https://github.com/flamingo-stack/openframe-cli/issues)

Congratulations! You now have a fully functional OpenFrame development environment running locally. 🎉