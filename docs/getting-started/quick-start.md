# Quick Start Guide

Get up and running with OpenFrame CLI in under 5 minutes! This guide will take you from zero to a fully functional OpenFrame development environment.

## TL;DR - 5-Minute Setup

For those who want to jump straight in:

```bash
# 1. Install OpenFrame CLI
curl -sSL https://get.openframe.io | bash

# 2. Bootstrap complete environment
openframe bootstrap my-dev-cluster

# 3. Verify everything is working
openframe cluster status
kubectl get pods -A
```

That's it! Skip to [Verification](#verification) to confirm your setup.

## Step-by-Step Installation

### Step 1: Install OpenFrame CLI

Choose your preferred installation method:

#### Option A: One-Line Install (Recommended)
```bash
curl -sSL https://get.openframe.io | bash
```

#### Option B: Manual Download
```bash
# Download latest release
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64
chmod +x openframe-linux-amd64
sudo mv openframe-linux-amd64 /usr/local/bin/openframe
```

#### Option C: Build from Source
```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe
sudo mv openframe /usr/local/bin/
```

**Verify Installation:**
```bash
openframe --version
openframe --help
```

Expected output:
```
openframe version v1.0.0
Build: 2024-01-15
```

### Step 2: Bootstrap Your First Environment

The `bootstrap` command creates a complete OpenFrame environment in one step:

```bash
openframe bootstrap my-dev-cluster
```

You'll see output like this:

```
🚀 OpenFrame CLI

Starting bootstrap process...
✅ Prerequisites check passed
🔄 Creating cluster 'my-dev-cluster'...
✅ Cluster created successfully
🔄 Installing ArgoCD...
✅ ArgoCD installed
🔄 Deploying OpenFrame applications...
✅ All applications deployed

🎉 Bootstrap complete! Your OpenFrame environment is ready.

Next steps:
  kubectl get pods -A              # Check all pods
  openframe cluster status         # View cluster details  
  openframe dev intercept --help   # Explore development tools
```

**What just happened?**

1. ✅ **Prerequisites checked** - Verified Docker, kubectl, helm, k3d
2. 🏗️ **Cluster created** - K3d cluster with 3 nodes
3. ⚡ **ArgoCD installed** - GitOps controller for application management
4. 📦 **Apps deployed** - OpenFrame platform components via ArgoCD
5. 🔧 **Environment ready** - Full development environment configured

### Step 3: Verification

Let's make sure everything is working correctly:

```bash
# Check cluster status
openframe cluster status
```

Expected output:
```
📊 Cluster Status: my-dev-cluster

Cluster Info:
  Name: my-dev-cluster
  Type: k3d
  Status: ✅ Running
  Nodes: 3 (1 master, 2 workers)
  Kubernetes: v1.28.2+k3s1

ArgoCD Status:
  Status: ✅ Ready
  Applications: 3/3 Synced
  URL: https://localhost:8080

Resource Usage:
  CPU: 15%
  Memory: 2.1GB/8GB
```

```bash
# Check all pods are running
kubectl get pods -A
```

You should see pods in various namespaces:
```
NAMESPACE     NAME                                     READY   STATUS    RESTARTS
argocd        argocd-server-xxx                       1/1     Running   0
argocd        argocd-application-controller-xxx       1/1     Running   0  
openframe     openframe-api-xxx                       1/1     Running   0
openframe     openframe-ui-xxx                        1/1     Running   0
kube-system   coredns-xxx                             1/1     Running   0
```

```bash
# Access ArgoCD UI (optional)
kubectl port-forward svc/argocd-server -n argocd 8080:443 &
# Open https://localhost:8080 in your browser
# Username: admin
# Password: (get with) kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d
```

## Basic "Hello World" Example

Now let's deploy a simple application to test your environment:

### 1. Create a Sample Application

```bash
# Create a simple nginx deployment
kubectl create deployment hello-world --image=nginx:latest
kubectl expose deployment hello-world --port=80 --target-port=80 --type=NodePort
```

### 2. Access Your Application

```bash
# Get the service details
kubectl get svc hello-world

# Port forward to access locally
kubectl port-forward svc/hello-world 8080:80 &

# Test the application
curl http://localhost:8080
```

You should see the nginx welcome page HTML.

### 3. Clean Up

```bash
# Remove the test application
kubectl delete deployment hello-world
kubectl delete service hello-world
```

## Expected Results

After completing the quick start, you should have:

✅ **OpenFrame CLI installed** and working  
✅ **Kubernetes cluster** running with 3 nodes  
✅ **ArgoCD deployed** and managing applications  
✅ **OpenFrame platform** components running  
✅ **kubectl configured** to access your cluster  
✅ **Sample app deployed** and accessible  

## Interactive Mode vs Non-Interactive

### Interactive Mode (Default)
Perfect for learning and development:
```bash
# Shows menus and prompts for configuration
openframe bootstrap

# Example prompts you'll see:
# ? Select deployment mode: 
#   oss-tenant (recommended for development)
#   saas-shared (multi-tenant)  
#   saas-tenant (single tenant)
# ? Enable verbose logging? (y/N)
```

### Non-Interactive Mode
Great for automation and CI/CD:
```bash
# Skips all prompts, uses defaults or flags
openframe bootstrap my-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose
```

## Available Deployment Modes

| Mode | Use Case | Description |
|------|----------|-------------|
| **oss-tenant** | Development, learning | Single-tenant open source deployment |
| **saas-shared** | Multi-tenancy testing | Shared infrastructure for multiple tenants |
| **saas-tenant** | Production-like | Dedicated tenant infrastructure |

## Quick Commands Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `openframe bootstrap` | Complete setup | `openframe bootstrap dev-cluster` |
| `openframe cluster list` | Show clusters | `openframe cluster list` |
| `openframe cluster delete` | Remove cluster | `openframe cluster delete dev-cluster` |
| `openframe chart install` | Install charts only | `openframe chart install` |
| `openframe dev intercept` | Traffic interception | `openframe dev intercept api-service` |

## Troubleshooting Quick Start Issues

### Installation Issues

**Problem**: Command not found after installation
```bash
# Check if binary is in PATH
which openframe

# If not found, check installation location
find /usr/local/bin -name "openframe*"

# Add to PATH if needed
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Problem**: Permission denied
```bash
# Fix permissions
sudo chmod +x /usr/local/bin/openframe

# Or install to user directory
mkdir -p ~/bin
mv openframe ~/bin/
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
```

### Bootstrap Issues

**Problem**: Prerequisites check failed
```bash
# Run detailed prerequisites check
openframe cluster create --check-prerequisites-only

# Install missing tools (see prerequisites.md)
```

**Problem**: Cluster creation failed
```bash
# Check Docker status
docker ps

# Clean up any partial clusters
k3d cluster delete my-dev-cluster
openframe cluster cleanup

# Try again with verbose output
openframe bootstrap my-dev-cluster --verbose
```

**Problem**: ArgoCD installation failed
```bash
# Check cluster status first
kubectl get nodes

# Try chart installation separately
openframe chart install --verbose

# Check ArgoCD namespace
kubectl get pods -n argocd
```

## Next Steps

🎉 Congratulations! You now have a working OpenFrame environment. Here's what to explore next:

### Immediate Next Steps
1. **[First Steps Guide](first-steps.md)** - Explore key features and workflows
2. **[Development Setup](../development/setup/environment.md)** - Configure your IDE and tools
3. **[Architecture Overview](../development/architecture/overview.md)** - Understand the system design

### Learn More
- **Cluster Management**: `openframe cluster --help`
- **Chart Operations**: `openframe chart --help`  
- **Development Tools**: `openframe dev --help`
- **Configuration**: Check `~/.openframe/config.yaml`

### Join the Community
- 📖 Browse additional documentation
- 🐛 Report issues on GitHub
- 💬 Join our Discord community
- ⭐ Star us on GitHub if this was helpful!

---

> **💡 Pro Tip**: Keep your first cluster around for experimentation. You can always create additional clusters with `openframe bootstrap another-cluster` to test different configurations without affecting your main development environment.