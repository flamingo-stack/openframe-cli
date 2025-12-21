# Quick Start Guide

Get OpenFrame CLI up and running in under 5 minutes! This guide will take you from installation to your first working OpenFrame environment.

## TL;DR - 5-Minute Setup

```bash
# 1. Install OpenFrame CLI
curl -sSL https://install.openframe.io | bash

# 2. Bootstrap complete environment
openframe bootstrap

# 3. Access ArgoCD UI
# Follow the on-screen instructions to access the web interface
```

That's it! You now have a complete OpenFrame environment running locally.

## Step-by-Step Installation

### Step 1: Install OpenFrame CLI

Choose your preferred installation method:

**Option A: Install Script (Recommended)**
```bash
# Download and install latest version
curl -sSL https://install.openframe.io | bash

# Verify installation
openframe --version
```

**Option B: Manual Installation**
```bash
# Download binary for your platform
wget https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64

# Make executable and move to PATH
chmod +x openframe-linux-amd64
sudo mv openframe-linux-amd64 /usr/local/bin/openframe

# Verify installation
openframe --version
```

**Option C: Build from Source**
```bash
# Clone repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build binary
go build -o openframe main.go

# Move to PATH
sudo mv openframe /usr/local/bin/

# Verify installation
openframe --version
```

### Step 2: Bootstrap Your Environment

The fastest way to get started is with the bootstrap command:

```bash
# Interactive bootstrap (recommended for first-time users)
openframe bootstrap

# Non-interactive with specific deployment mode
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive
```

The bootstrap process will:
1. ✅ Check prerequisites (Docker, kubectl, helm, k3d)
2. 🔧 Create a K3d cluster
3. 📦 Install ArgoCD
4. 🚀 Deploy OpenFrame applications
5. 🌐 Configure port forwarding

**Expected Output:**
```
🚀 OpenFrame CLI

✅ Checking prerequisites...
✅ Docker is running
✅ kubectl found
✅ helm found  
✅ k3d found

🔧 Creating cluster 'openframe-dev'...
✅ Cluster created successfully

📦 Installing ArgoCD...
✅ ArgoCD installed and ready

🚀 Deploying OpenFrame applications...
✅ Applications synced successfully

🌐 Setting up port forwarding...
✅ ArgoCD UI: http://localhost:8080

🎉 Bootstrap complete! Your OpenFrame environment is ready.

Next steps:
- Access ArgoCD UI: http://localhost:8080
- Username: admin
- Password: [generated password shown here]
```

### Step 3: Verify Installation

Let's make sure everything is working correctly:

**Check Cluster Status:**
```bash
openframe cluster status
```

Expected output:
```
🚀 OpenFrame CLI

📊 Cluster Status: openframe-dev

✅ Cluster is running
✅ ArgoCD is healthy
✅ All pods are ready

Resources:
- Nodes: 1
- Namespaces: 4
- Pods: 8 (8 running, 0 pending, 0 failed)
- Services: 6
```

**List Available Clusters:**
```bash
openframe cluster list
```

**Access ArgoCD Web Interface:**
1. Open your browser to `http://localhost:8080`
2. Username: `admin`
3. Password: Use the password shown in bootstrap output
4. You should see the ArgoCD dashboard with your applications

### Step 4: Test Basic Functionality

**Create a Simple Application:**

Create a test application to verify everything works:

```bash
# Create test namespace
kubectl create namespace test-app

# Deploy a simple application
kubectl create deployment hello-world --image=nginx:alpine -n test-app
kubectl expose deployment hello-world --port=80 --type=ClusterIP -n test-app

# Verify deployment
kubectl get pods -n test-app
```

**Test Port Forwarding:**
```bash
# Forward port to access the application
kubectl port-forward deployment/hello-world 8081:80 -n test-app &

# Test the application
curl http://localhost:8081
```

You should see the nginx welcome page HTML.

## Common Bootstrap Options

The bootstrap command supports several useful options:

### Interactive Mode (Default)
```bash
# Prompts for deployment mode and cluster name
openframe bootstrap
```

### Quick Setup with Defaults
```bash
# Use OSS tenant mode with default cluster name
openframe bootstrap --deployment-mode=oss-tenant
```

### Custom Cluster Name
```bash
# Create cluster with specific name
openframe bootstrap production-cluster
```

### CI/CD Mode
```bash
# Fully automated setup for scripts
openframe bootstrap ci-cluster \
  --deployment-mode=saas-shared \
  --non-interactive \
  --verbose
```

### Verbose Output
```bash
# Show detailed progress including ArgoCD sync logs
openframe bootstrap --verbose
```

## Quick Command Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `openframe bootstrap` | Complete environment setup | `openframe bootstrap dev` |
| `openframe cluster list` | Show all clusters | `openframe cluster list` |
| `openframe cluster status` | Check cluster health | `openframe cluster status` |
| `openframe chart install` | Install charts manually | `openframe chart install` |
| `openframe dev intercept` | Local development mode | `openframe dev intercept my-service` |
| `openframe cluster cleanup` | Clean up resources | `openframe cluster cleanup` |

## Expected Results

After completing the quick start, you should have:

✅ **OpenFrame CLI installed** and accessible from command line  
✅ **Local K3d cluster running** with OpenFrame components  
✅ **ArgoCD deployed** and accessible at http://localhost:8080  
✅ **OpenFrame applications synced** and healthy  
✅ **Port forwarding configured** for easy access  

## Troubleshooting Quick Fixes

### Bootstrap Fails at Prerequisites
```bash
# Check what's missing
openframe bootstrap --dry-run

# Install missing tools (see prerequisites guide)
# Then retry bootstrap
openframe bootstrap
```

### Port Already in Use
```bash
# Use different ports
openframe bootstrap --api-port 6444

# Or stop conflicting services
sudo lsof -i :6443
kill -9 <PID>
```

### ArgoCD UI Not Accessible
```bash
# Restart port forwarding
kubectl port-forward svc/argocd-server -n argocd 8080:443 &

# Check if ArgoCD is running
kubectl get pods -n argocd
```

### Cluster Creation Fails
```bash
# Clean up any existing clusters
k3d cluster delete openframe-dev

# Try again with verbose output
openframe bootstrap --verbose
```

## Next Steps

Congratulations! You now have OpenFrame CLI running. Here's what to do next:

### Immediate Next Steps
1. **[First Steps Guide](first-steps.md)** - Essential configuration and features
2. **Explore ArgoCD UI** - Get familiar with the GitOps interface
3. **Try Development Commands** - Use `openframe dev --help` to see development options

### For Development
4. **[Development Environment Setup](../development/setup/environment.md)** - Configure your IDE and tools
5. **[Local Development Guide](../development/setup/local-development.md)** - Learn development workflows

### For Operations
6. **[Architecture Overview](../development/architecture/overview.md)** - Understand the system design
7. **[Testing Guide](../development/testing/overview.md)** - Learn about testing strategies

## Getting Help

If you encounter any issues:

- **Command Help**: `openframe <command> --help`
- **Verbose Output**: Add `--verbose` to any command
- **Check Status**: `openframe cluster status`
- **Clean Slate**: `openframe cluster cleanup --all` then retry

---

> 🎉 **Success!** You've successfully installed OpenFrame CLI and bootstrapped your first environment. You're now ready to explore the full capabilities of OpenFrame!