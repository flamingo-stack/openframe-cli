# Quick Start Guide

Get OpenFrame CLI up and running in 5 minutes with this streamlined setup guide. This guide uses the bootstrap command for the fastest experience.

## TL;DR Installation Steps

```bash
# 1. Download OpenFrame CLI (replace with actual download method)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/

# 2. Verify installation
openframe --version

# 3. Bootstrap complete environment
openframe bootstrap my-first-cluster

# 4. Verify cluster is running
kubectl get nodes
kubectl get pods -A
```

> **⚠️ Prerequisites Required**: Ensure you have completed the [Prerequisites](./prerequisites.md) setup before proceeding.

## Step-by-Step Quick Setup

### Step 1: Install OpenFrame CLI

Choose your platform:

<details>
<summary><strong>macOS</strong></summary>

```bash
# Via Homebrew (if available)
brew install flamingo-stack/tap/openframe-cli

# Or download binary
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-darwin-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```
</details>

<details>
<summary><strong>Linux</strong></summary>

```bash
# Download and install
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/

# Verify installation
openframe --version
```
</details>

<details>
<summary><strong>Windows</strong></summary>

```powershell
# In PowerShell
Invoke-WebRequest -Uri "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-windows-amd64.exe" -OutFile "openframe.exe"

# Move to PATH location or add current directory to PATH
```
</details>

### Step 2: Bootstrap Your First Environment

The bootstrap command creates a complete OpenFrame environment:

```bash
# Interactive bootstrap (recommended for first-time users)
openframe bootstrap my-first-cluster
```

You'll see an interactive wizard:

```
   ___                   _____                        
  / _ \ _ __   ___ _ __  |  ___| __ __ _ _ __ ___   ___ 
 | | | | '_ \ / _ \ '_ \ | |_ | '__/ _` | '_ ` _ \ / _ \
 | |_| | |_) |  __/ | | |  _|| | | (_| | | | | | |  __/
  \___/| .__/ \___|_| |_|_|  |_|  \__,_|_| |_| |_|\___|
       |_|                                            

🚀 Welcome to OpenFrame CLI Bootstrap

? Cluster name: my-first-cluster
? Deployment mode: 
  ▸ OSS Tenant (single-tenant, open source)
    SaaS Tenant (single-tenant, managed)
    SaaS Shared (multi-tenant, managed)

? Number of worker nodes: 3
? Enable ArgoCD UI: Yes
? Port configuration: 
  ▸ Default (80, 443, 8080)
    Custom

✅ Creating cluster...
✅ Installing ArgoCD...
✅ Deploying OpenFrame charts...

🎉 Bootstrap complete! Your OpenFrame environment is ready.
```

### Step 3: Verify Installation

Check that your cluster is running:

```bash
# Check cluster nodes
kubectl get nodes

# Expected output:
# NAME                         STATUS   ROLES                  AGE   VERSION
# k3d-my-first-cluster-server-0   Ready    control-plane,master   2m    v1.27.4+k3s1
# k3d-my-first-cluster-agent-0    Ready    <none>                 2m    v1.27.4+k3s1
# k3d-my-first-cluster-agent-1    Ready    <none>                 2m    v1.27.4+k3s1
# k3d-my-first-cluster-agent-2    Ready    <none>                 2m    v1.27.4+k3s1

# Check ArgoCD installation
kubectl get pods -n argocd

# Expected output:
# NAME                                 READY   STATUS    RESTARTS   AGE
# argocd-application-controller-0      1/1     Running   0          2m
# argocd-dex-server-xxx               1/1     Running   0          2m
# argocd-redis-xxx                    1/1     Running   0          2m
# argocd-repo-server-xxx              1/1     Running   0          2m
# argocd-server-xxx                   1/1     Running   0          2m
```

### Step 4: Access ArgoCD UI

```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Open in browser: https://localhost:8080
# Username: admin
# Password: (from command above)
```

## Basic "Hello World" Example

Deploy a simple application to verify everything works:

```bash
# Create a test namespace
kubectl create namespace hello-world

# Deploy nginx
kubectl create deployment hello-nginx --image=nginx --namespace=hello-world

# Expose the service
kubectl expose deployment hello-nginx --port=80 --target-port=80 --namespace=hello-world

# Check deployment
kubectl get pods -n hello-world

# Expected output:
# NAME                           READY   STATUS    RESTARTS   AGE
# hello-nginx-xxx               1/1     Running   0          30s
```

### Test Application Access

```bash
# Port forward to test
kubectl port-forward -n hello-world deployment/hello-nginx 8081:80

# Open in browser: http://localhost:8081
# Should see nginx welcome page
```

## Expected Output and Results

After successful bootstrap, you should have:

### ✅ Cluster Components
- K3d cluster with specified number of nodes
- Traefik ingress controller (included with K3d)
- Local storage provisioner

### ✅ ArgoCD Installation  
- ArgoCD server and components in `argocd` namespace
- ArgoCD CLI access configured
- Admin user with generated password

### ✅ OpenFrame Charts
- Application configuration via app-of-apps pattern
- GitOps workflows configured
- Ingress routing setup

### ✅ Development Tools Ready
- kubectl context automatically configured
- Helm tiller (if needed) installed
- Ready for Telepresence/Skaffold workflows

## Quick Validation Checklist

Verify your environment with these commands:

```bash
# Cluster health
openframe cluster status my-first-cluster

# Available clusters
openframe cluster list

# ArgoCD applications
kubectl get applications -n argocd

# Ingress configuration
kubectl get ingress -A
```

## Common Quick Start Issues

### Issue: Docker Not Running
```bash
# Error: Cannot connect to the Docker daemon
# Solution: Start Docker
sudo systemctl start docker  # Linux
# or start Docker Desktop on macOS/Windows
```

### Issue: Port Already in Use
```bash
# Error: Port 80 or 443 already in use
# Solution: Stop conflicting services or use custom ports
openframe bootstrap my-cluster --port-offset=1000
```

### Issue: ArgoCD Pods Not Starting
```bash
# Check pod events
kubectl describe pod -n argocd argocd-server-xxx

# Common causes: insufficient resources, image pull issues
# Solution: Ensure minimum 4GB RAM available
```

## What You've Accomplished

In just 5 minutes, you've:

1. **Installed** OpenFrame CLI
2. **Created** a complete Kubernetes environment
3. **Deployed** ArgoCD for GitOps workflows  
4. **Verified** everything is working
5. **Tested** with a simple application

## Next Steps After Quick Start

Now that you have a working environment, explore these areas:

1. **[First Steps](./first-steps.md)** - Learn about key OpenFrame features
2. **[Architecture Overview](../development/architecture/overview.md)** - Understand the components
3. **[Development Setup](../development/setup/local-development.md)** - Set up for development

### Advanced Exploration

```bash
# Explore cluster management
openframe cluster list
openframe cluster status my-first-cluster

# Explore chart management  
openframe chart --help

# Explore development tools
openframe dev --help
```

## Cleanup (Optional)

To remove the environment when done:

```bash
# Delete cluster and all resources
openframe cluster delete my-first-cluster

# Cleanup unused resources
openframe cluster cleanup
```

---

> **🎉 Congratulations!** You've successfully set up your first OpenFrame environment. The bootstrap command automated the complex process of cluster creation, ArgoCD installation, and chart deployment.

Ready to dive deeper? Continue to [First Steps](./first-steps.md) to learn about OpenFrame's key features and workflows!