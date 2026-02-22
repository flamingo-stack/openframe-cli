# Quick Start Guide

Get OpenFrame CLI up and running in 5 minutes! This guide will walk you through installing OpenFrame CLI and bootstrapping your first environment.

[![OpenFrame: 5-Minute MSP Platform Walkthrough - Cut Vendor Costs & Automate Ops](https://img.youtube.com/vi/er-z6IUnAps/maxresdefault.jpg)](https://www.youtube.com/watch?v=er-z6IUnAps)

## TL;DR - 5-Minute Setup

If you have all prerequisites installed, here's the fastest path:

```bash
# 1. Download and install OpenFrame CLI (choose your platform)
# Linux/macOS
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/

# Windows (download manually)
# Download: https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip

# 2. Verify installation
openframe --version

# 3. Bootstrap complete environment
openframe bootstrap

# 4. Check cluster status
openframe cluster status
```

That's it! Continue reading for detailed steps and explanations.

## Step 1: Install OpenFrame CLI

### Option A: Download Pre-built Binary (Recommended)

Choose your platform and download the latest release:

#### Linux (AMD64)
```bash
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/
chmod +x /usr/local/bin/openframe
```

#### Linux (ARM64)
```bash
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_arm64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/
chmod +x /usr/local/bin/openframe
```

#### macOS (AMD64)
```bash
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/
chmod +x /usr/local/bin/openframe
```

#### macOS (Apple Silicon)
```bash
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/
chmod +x /usr/local/bin/openframe
```

#### Windows (AMD64)
1. Download: https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
2. Extract the ZIP file
3. Move `openframe.exe` to a directory in your `PATH`
4. Open WSL2 terminal and verify access

### Option B: Build from Source

If you have Go 1.24.6+ installed:

```bash
# Clone the repository
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli

# Build the binary
go build -o openframe main.go

# Move to PATH
sudo mv openframe /usr/local/bin/
```

## Step 2: Verify Installation

Confirm OpenFrame CLI is installed correctly:

```bash
openframe --version
```

Expected output:
```text
openframe version v1.x.x (commit: abc123, built: 2024-01-01)
```

Check available commands:
```bash
openframe --help
```

You should see the main command groups:
- `bootstrap` - Complete environment setup
- `cluster` - Kubernetes cluster management
- `chart` - Helm chart and ArgoCD management
- `dev` - Development tools and workflows

## Step 3: Bootstrap Your First Environment

The bootstrap command creates a complete OpenFrame environment with a single command:

```bash
openframe bootstrap
```

### Interactive Bootstrap

The command will guide you through setup with prompts:

```text
🚀 OpenFrame Bootstrap Wizard

? Select deployment mode:
  > oss-tenant (Single tenant, open source)
    saas-tenant (Multi-tenant SaaS mode) 
    saas-shared (Shared SaaS infrastructure)

? Enable verbose logging? (y/N): y

🔍 Checking prerequisites...
✅ Docker is running
✅ kubectl is available  
✅ Helm is available
✅ K3D is available

🎯 Creating K3D cluster...
✅ Cluster 'openframe-local' created

🎭 Installing ArgoCD...
✅ ArgoCD installed and ready

📦 Installing application charts...
✅ App-of-apps synchronized
✅ All applications healthy

🎉 OpenFrame environment is ready!
```

### Non-Interactive Bootstrap

For scripts and CI/CD, use flags to skip prompts:

```bash
openframe bootstrap \
  --mode=oss-tenant \
  --non-interactive \
  --verbose
```

## Step 4: Verify Your Environment

### Check Cluster Status
```bash
openframe cluster status
```

Expected output:
```text
📊 Cluster Status: openframe-local

Cluster Info:
  Name: openframe-local
  Status: Running ✅
  Nodes: 1 (1 ready)
  Kubernetes Version: v1.28.6+k3s1

Resource Usage:
  CPU: 2 cores (25% used)
  Memory: 8Gi (45% used)  
  Storage: 50Gi (12% used)

Key Services:
  ArgoCD: Healthy ✅
  Traefik: Healthy ✅
  CoreDNS: Healthy ✅
```

### Access ArgoCD UI
The bootstrap process installs ArgoCD. Access the web interface:

```bash
# Get ArgoCD URL and credentials
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Then open: http://localhost:8080
- Username: `admin`  
- Password: (from the command above)

### View Running Applications
```bash
# List all pods across namespaces
kubectl get pods --all-namespaces

# Check ArgoCD applications
kubectl get applications -n argocd
```

## Step 5: Test Basic Functionality

### Create a Test Namespace
```bash
kubectl create namespace test-app
kubectl get namespaces
```

### Deploy a Simple Application
```bash
# Create a test deployment
kubectl create deployment nginx --image=nginx:latest -n test-app
kubectl expose deployment nginx --port=80 --target-port=80 -n test-app

# Check deployment
kubectl get pods -n test-app
kubectl get svc -n test-app
```

### Test Service Access
```bash
# Port forward to test connectivity
kubectl port-forward svc/nginx -n test-app 8081:80 &

# Test the service
curl http://localhost:8081

# Clean up
kill %1  # Stop port-forward
kubectl delete namespace test-app
```

## What Just Happened?

The bootstrap process created:

1. **K3D Cluster**: A lightweight Kubernetes cluster running in Docker
2. **ArgoCD**: GitOps deployment tool for application management  
3. **Traefik**: Ingress controller for routing traffic
4. **Core Services**: DNS, metrics, and monitoring components
5. **Application Templates**: Ready-to-use deployment patterns

## Expected Results

After successful bootstrap:

| Component | Status | Access Method |
|-----------|--------|---------------|
| **Kubernetes API** | ✅ Running | `kubectl` commands |
| **ArgoCD UI** | ✅ Running | Port forward to 8080 |
| **Traefik Dashboard** | ✅ Running | Port forward to 9000 |
| **Container Registry** | ✅ Running | Docker daemon |

## Troubleshooting

### Bootstrap Fails with Docker Error
```bash
# Check Docker status
docker ps
docker info

# Restart Docker if needed
sudo systemctl restart docker  # Linux
# or restart Docker Desktop on macOS/Windows
```

### Kubectl Cannot Connect
```bash
# Check kubeconfig
kubectl config current-context
kubectl config get-contexts

# Switch to openframe context if needed
kubectl config use-context k3d-openframe-local
```

### ArgoCD Not Accessible
```bash
# Check ArgoCD pods
kubectl get pods -n argocd

# Restart ArgoCD if needed
kubectl rollout restart deployment argocd-server -n argocd
```

### Port Already in Use
```bash
# Find and kill processes using required ports
lsof -ti:6443,8080,9000 | xargs kill -9

# Or use different ports
kubectl port-forward svc/argocd-server -n argocd 8081:443
```

## Next Steps

🎉 **Congratulations!** You now have a running OpenFrame environment. Here's what to explore next:

- **[First Steps Guide](first-steps.md)** - Learn key workflows and features
- **[Architecture Overview](../development/architecture/README.md)** - Understand how components work together  
- **[Development Setup](../development/setup/local-development.md)** - Configure your development environment

## Common Next Actions

### Deploy Your First Application
```bash
# Use ArgoCD to deploy from Git
openframe chart install my-app \
  --repo=https://github.com/your-org/your-app \
  --path=helm-chart
```

### Start Local Development
```bash
# Intercept a service for local development
openframe dev intercept my-service \
  --namespace=default \
  --port=8080:3000
```

### Explore the Environment
```bash
# List available commands
openframe --help

# Get cluster information
openframe cluster list
openframe cluster status

# Check chart installations
openframe chart list
```

## Getting Help

Need assistance? The OpenFrame community is here to help:

- **OpenMSP Slack**: [Join the community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **Documentation**: Browse other guides in this repository
- **GitHub Issues**: Report bugs or request features (but use Slack for general support)

> **Note**: All community support happens in Slack - we don't monitor GitHub Issues for support requests.