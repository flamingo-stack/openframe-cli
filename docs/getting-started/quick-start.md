# Quick Start Guide

Get OpenFrame CLI running in 5 minutes with this streamlined setup guide.

## TL;DR - One Command Setup

```bash
# Install OpenFrame CLI (download from releases)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/

# Bootstrap complete environment
openframe bootstrap
```

That's it! 🎉 Your local Kubernetes cluster with ArgoCD is ready.

## Step-by-Step Setup

### 1. Install OpenFrame CLI

#### Option A: Download Binary (Recommended)

**Linux:**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**macOS:**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-darwin-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri "https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-windows-amd64.exe" -OutFile "openframe.exe"
# Move to a directory in your PATH
```

#### Option B: Build from Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

### 2. Verify Installation

```bash
openframe --help
```

Expected output:
```text
🚀 OpenFrame CLI - Kubernetes cluster management made simple

Usage:
  openframe [command]

Available Commands:
  bootstrap   Bootstrap complete OpenFrame environment
  cluster     Manage Kubernetes clusters (alias: k)
  chart       Manage Helm charts (alias: c)
  dev         Development tools (alias: d)
  help        Help about any command

Flags:
  -h, --help     help for openframe
  -v, --version  version for openframe

Use "openframe [command] --help" for more information about a command.
```

### 3. Bootstrap Your First Environment

#### Option A: Interactive Mode (Recommended)

```bash
openframe bootstrap
```

This will show an interactive menu:

```text
🚀 OpenFrame CLI v1.0.0

Welcome to OpenFrame! Let's set up your development environment.

? Select deployment mode:
  ▸ OSS Tenant (Open Source)
    SaaS Tenant (Multi-tenant)
    SaaS Shared (Shared infrastructure)

? Enter cluster name: (openframe-dev)
? Number of worker nodes: (2)

🔍 Checking prerequisites...
✅ Docker is running
✅ kubectl installed
✅ helm installed
✅ k3d installed

🚀 Creating cluster 'openframe-dev'...
✅ K3d cluster created successfully

📦 Installing ArgoCD...
✅ ArgoCD installed
✅ App-of-apps configured

🎉 Environment ready!

ArgoCD UI: http://localhost:8080
Username: admin
Password: (auto-generated - see below)
```

#### Option B: Non-Interactive Mode

```bash
# Quick defaults
openframe bootstrap my-cluster --deployment-mode=oss-tenant --non-interactive

# Custom configuration
openframe bootstrap my-cluster \
  --deployment-mode=oss-tenant \
  --nodes=3 \
  --non-interactive \
  --verbose
```

### 4. Verify Your Environment

#### Check Cluster Status
```bash
openframe cluster status
```

Expected output:
```text
🔍 Cluster Status: openframe-dev

Cluster Info:
  Name: openframe-dev
  Status: Running
  Nodes: 3 (1 master, 2 workers)
  K3s Version: v1.27.3+k3s1

Network:
  API Server: https://0.0.0.0:6443
  Load Balancer: 0.0.0.0:8080-8090

Services:
  ✅ ArgoCD (http://localhost:8080)
  ✅ Kubernetes API
  ✅ CoreDNS
  ✅ Traefik

Recent Activity:
  2 minutes ago: Cluster created
  1 minute ago: ArgoCD installed
  30 seconds ago: App-of-apps deployed
```

#### Check Kubernetes Resources
```bash
kubectl get nodes
kubectl get pods -A
kubectl get svc -A
```

#### Access ArgoCD UI
```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Open ArgoCD UI
open http://localhost:8080  # macOS
# Or browse to http://localhost:8080 on any platform
```

## Hello World Example

Let's deploy a simple application to test your setup:

### 1. Create a Test Application

```bash
# Create namespace
kubectl create namespace hello-world

# Deploy simple nginx application
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
  namespace: hello-world
spec:
  replicas: 2
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: nginx
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world
  namespace: hello-world
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
EOF
```

### 2. Test the Application

```bash
# Check deployment status
kubectl get pods -n hello-world

# Get service URL
kubectl get svc -n hello-world

# Test the application (may take a moment for LoadBalancer)
curl http://localhost:8081  # Adjust port based on service output
```

Expected output:
```html
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
...
```

### 3. View in ArgoCD (Optional)

1. Access ArgoCD at http://localhost:8080
2. Login with admin and the password from step 4 above
3. You should see your applications in the UI

## Expected Results

After completing the quick start, you should have:

### ✅ Running Components

| Component | Status | Access |
|-----------|--------|--------|
| **K3d Cluster** | Running | `kubectl` commands |
| **ArgoCD** | Deployed | http://localhost:8080 |
| **Kubernetes API** | Active | kubectl configured |
| **Load Balancer** | Ready | Services accessible via localhost |

### ✅ Verified Capabilities

- **Cluster Management**: Create/delete/list clusters
- **Application Deployment**: Deploy workloads with kubectl
- **GitOps Ready**: ArgoCD installed and configured
- **Development Ready**: Ready for traffic interception and live development

### ✅ Key Files Created

- **Kubeconfig**: `~/.kube/config` updated with cluster context
- **ArgoCD Config**: ArgoCD server and admin credentials
- **Cluster State**: K3d cluster configuration stored locally

## Quick Commands Reference

| Task | Command | Description |
|------|---------|-------------|
| **Check cluster** | `openframe cluster status` | View cluster information |
| **List clusters** | `openframe cluster list` | Show all clusters |
| **Delete cluster** | `openframe cluster delete` | Remove cluster |
| **Install charts** | `openframe chart install` | Install ArgoCD manually |
| **Access ArgoCD** | Open http://localhost:8080 | ArgoCD web interface |

## Troubleshooting Quick Issues

### Cluster Creation Fails

```bash
# Check prerequisites
docker info
k3d version

# Check existing clusters
k3d cluster list

# Clean up if needed
k3d cluster delete openframe-dev
openframe bootstrap --deployment-mode=oss-tenant
```

### ArgoCD Not Accessible

```bash
# Check ArgoCD pods
kubectl get pods -n argocd

# Port-forward if needed
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Get admin password again
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### Kubectl Context Issues

```bash
# Check current context
kubectl config current-context

# Switch to OpenFrame cluster
kubectl config use-context k3d-openframe-dev

# List all contexts
kubectl config get-contexts
```

## Cleanup (When Done Testing)

To remove the test environment:

```bash
# Delete test application
kubectl delete namespace hello-world

# Delete entire cluster
openframe cluster delete openframe-dev

# Or delete all clusters
k3d cluster delete --all
```

## What's Next?

🎉 **Congratulations!** You now have a working OpenFrame environment. 

### Next Steps:

1. **[First Steps Guide](./first-steps.md)** - Explore OpenFrame's key features
2. **[Development Setup](../development/setup/local-development.md)** - Set up your development workflow  
3. **[Architecture Overview](../development/architecture/overview.md)** - Understand how OpenFrame works

### Common Next Actions:

- **Deploy your application**: Use ArgoCD to deploy your own applications
- **Set up traffic interception**: Use `openframe dev intercept` for local development
- **Configure GitOps**: Connect ArgoCD to your Git repositories
- **Scale your cluster**: Add more nodes or create additional clusters

> 💡 **Pro Tip**: Keep your cluster running and explore the `openframe dev` commands to see traffic interception in action!