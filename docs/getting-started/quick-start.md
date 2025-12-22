# Quick Start Guide

Get your first OpenFrame cluster up and running in under 5 minutes! This guide will take you from zero to a fully functional Kubernetes cluster with GitOps capabilities.

## TL;DR - 5-Minute Setup

For the impatient, here's the complete quick start:

```bash
# 1. Install OpenFrame CLI
curl -sSL https://install.openframe.dev | bash

# 2. Create cluster with ArgoCD
openframe bootstrap

# 3. Verify everything works
kubectl get pods -A
```

That's it! Your cluster is ready for development and GitOps workflows.

## Step-by-Step Installation

### Step 1: Install OpenFrame CLI

Choose your preferred installation method:

#### Option A: Automated Script (Recommended)
```bash
curl -sSL https://install.openframe.dev | bash
```

#### Option B: Manual Download
```bash
# Download latest release
curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64
chmod +x openframe-cli-linux-amd64
sudo mv openframe-cli-linux-amd64 /usr/local/bin/openframe

# Verify installation
openframe --version
```

#### Option C: Build from Source
```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

### Step 2: Verify Installation

```bash
# Check OpenFrame CLI is installed
openframe --help

# Expected output:
# OpenFrame CLI - Kubernetes cluster bootstrapping and development tool
# 
# Usage:
#   openframe [command]
# 
# Available Commands:
#   bootstrap   Bootstrap a complete development environment
#   chart       Manage Helm charts and ArgoCD
#   cluster     Manage Kubernetes clusters
#   dev         Development tools and workflows
#   help        Help about any command
```

### Step 3: Bootstrap Your First Environment

The `bootstrap` command creates a complete development environment in one step:

```bash
# Start the interactive bootstrap wizard
openframe bootstrap
```

You'll see the OpenFrame logo and be guided through configuration:

```bash
🎯 OpenFrame CLI - Bootstrap Environment

✨ This wizard will create a complete development environment with:
   • K3d Kubernetes cluster
   • ArgoCD for GitOps
   • Development tools setup

📋 Cluster Configuration:
   Name: [my-cluster] 
   Nodes: [3] 
   Port: [8080] 
   
📋 GitOps Configuration:
   Repository: [https://github.com/your-org/argocd-apps] 
   Branch: [main] 
   Path: [applications] 

🚀 Creating cluster 'my-cluster'...
✅ Cluster created successfully
✅ ArgoCD installed
✅ GitOps configured

🎉 Your environment is ready!
```

## What You Just Created

Your bootstrap command created:

### Kubernetes Cluster
- **K3d cluster** with 3 nodes (1 server, 2 workers)
- **LoadBalancer** accessible on localhost:8080
- **kubectl context** automatically configured
- **Storage classes** for persistent volumes

### GitOps Infrastructure
- **ArgoCD server** with web UI
- **App-of-Apps pattern** configured
- **Git repository** connection established
- **Auto-sync** enabled for applications

### Development Tools
- **Prerequisites** automatically installed (kubectl, helm, k3d)
- **Telepresence** ready for traffic intercepts
- **Skaffold** configured for live development

## Verify Your Installation

Let's make sure everything is working:

### Check Cluster Status
```bash
# List all clusters
openframe cluster list

# Expected output:
# NAME        STATUS   NODES   VERSION   CREATED
# my-cluster  running  3/3     v1.28.x   2 minutes ago

# Check cluster details
openframe cluster status my-cluster

# View all running pods
kubectl get pods -A
```

### Access ArgoCD Web UI
```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port-forward to access UI (in another terminal)
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Open browser to https://localhost:8080
# Username: admin
# Password: (output from first command)
```

### Test Development Workflow
```bash
# Create a sample application
mkdir hello-openframe && cd hello-openframe

# Create simple deployment
cat <<EOF > deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-app
  template:
    metadata:
      labels:
        app: hello-app
    spec:
      containers:
      - name: hello
        image: nginx:alpine
        ports:
        - containerPort: 80
EOF

# Deploy to cluster
kubectl apply -f deployment.yaml

# Verify deployment
kubectl get pods
```

## Your First Development Workflow

Now let's try a common development workflow:

### 1. Create a Service for Testing
```bash
# Expose the deployment
kubectl expose deployment hello-app --port=80 --type=LoadBalancer

# Check service
kubectl get services
```

### 2. Set Up Traffic Intercept (Optional)
```bash
# Install a sample app that you can intercept
kubectl create deployment echo-server --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment echo-server --port=8080 --type=LoadBalancer

# Start Telepresence intercept (requires Telepresence installed)
openframe dev intercept echo-server --port=3000
```

### 3. Test with Skaffold (Optional)
```bash
# Initialize Skaffold in your project
openframe dev scaffold init

# Start live development
openframe dev scaffold dev
```

## Expected Results

After completing the quick start, you should have:

✅ **Running cluster** with 3 nodes
✅ **ArgoCD dashboard** accessible at https://localhost:8080
✅ **kubectl configured** to talk to your cluster
✅ **Sample application** deployed and running
✅ **Development tools** ready for use

### Validation Checklist

Run these commands to verify everything:

```bash
# ✅ Cluster is healthy
kubectl get nodes
# Should show 3 nodes in Ready state

# ✅ ArgoCD is running
kubectl get pods -n argocd
# Should show all argocd-* pods in Running state

# ✅ Your application is deployed
kubectl get deployment hello-app
# Should show 1/1 ready replicas

# ✅ OpenFrame tools are working
openframe cluster list
# Should show your cluster with "running" status
```

## Common Quick Start Issues

### Issue: Docker not running
```bash
# Error: Cannot connect to the Docker daemon
# Solution: Start Docker
sudo systemctl start docker  # Linux
open /Applications/Docker.app  # macOS
```

### Issue: Port already in use
```bash
# Error: Port 8080 already in use
# Solution: Use different port
openframe cluster create --api-port 8081
```

### Issue: Insufficient resources
```bash
# Error: Cluster creation fails
# Solution: Check available resources
docker system prune  # Free up disk space
free -h              # Check available RAM
```

### Issue: Network connectivity
```bash
# Error: Cannot pull images
# Solution: Check internet connection
curl -I https://docker.io  # Test registry access
```

## What's Next?

Congratulations! You now have a fully functional development environment. Here's what to explore next:

### Immediate Next Steps
1. **[First Steps Guide](first-steps.md)** - Learn the 5 essential things to do after installation
2. **[Architecture Overview](../development/architecture/overview.md)** - Understand how everything fits together
3. **[Development Setup](../development/setup/environment.md)** - Configure your IDE and tools

### Common Workflows
- **Deploy Applications**: Use ArgoCD to deploy apps from Git
- **Local Development**: Set up Skaffold for hot-reload development
- **Service Debugging**: Use Telepresence to intercept traffic
- **Cluster Management**: Create additional clusters for different environments

### Video Walkthrough

Want to see this in action? Check out our walkthrough video:

[![OpenFrame v0.3.4 - Enhanced Stability & Cross-Platform Support](https://img.youtube.com/vi/h9ZxyeYTBPE/maxresdefault.jpg)](https://www.youtube.com/watch?v=h9ZxyeYTBPE)

## Clean Up (Optional)

If you want to start fresh:

```bash
# Delete the cluster
openframe cluster delete my-cluster

# Clean up Docker resources
docker system prune
```

---

**Previous**: [Prerequisites](prerequisites.md) | **Next**: [First Steps](first-steps.md)

> 💡 **Pro Tip**: Save time by aliasing common commands:
> ```bash
> alias of="openframe"
> alias k="kubectl"
> # Now you can use: of cluster list, k get pods
> ```