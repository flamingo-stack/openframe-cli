# Quick Start Guide

Get OpenFrame CLI up and running in under 5 minutes with this streamlined setup guide. This covers the essentials to bootstrap your first OpenFrame environment.

## TL;DR - 5-Minute Setup

```bash
# 1. Download OpenFrame CLI
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/

# 2. Bootstrap complete environment
openframe bootstrap --deployment-mode=oss-tenant

# 3. Verify installation
kubectl get pods -A
```

That's it! Your OpenFrame environment is ready to use.

## Step-by-Step Installation

### Step 1: Download OpenFrame CLI

Choose your platform and download the latest release:

**Linux (x86_64):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-arm64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-darwin-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-darwin-arm64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

**Windows (WSL2):**
```bash
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64 -o openframe
chmod +x openframe
sudo mv openframe /usr/local/bin/
```

### Step 2: Verify Installation

```bash
# Check OpenFrame CLI version
openframe --version

# View available commands
openframe --help
```

**Expected output:**
```text
OpenFrame CLI v1.0.0 (commit: abc123, built: 2024-01-01T00:00:00Z)
```

### Step 3: Bootstrap Your Environment

OpenFrame CLI's `bootstrap` command creates a complete Kubernetes environment with ArgoCD and OpenFrame applications:

**Interactive Mode (Recommended for first-time users):**
```bash
openframe bootstrap
```

**Non-Interactive Mode (For automation):**
```bash
openframe bootstrap --deployment-mode=oss-tenant --non-interactive
```

**With Custom Cluster Name:**
```bash
openframe bootstrap my-dev-cluster --deployment-mode=oss-tenant
```

### Step 4: Monitor Bootstrap Progress

The bootstrap process will display progress indicators:

```text
🚀 OpenFrame Bootstrap Starting...

📋 Prerequisites Check
 ✅ Docker: 20.10.21 (running)
 ✅ K3d: 5.4.6 
 ✅ Helm: 3.10.2
 ✅ kubectl: 1.25.4
 
🏗️  Creating Cluster: openframe-cluster
 ✅ Cluster configuration generated
 ✅ K3d cluster created successfully
 ✅ Kubeconfig updated
 
📦 Installing Charts
 ✅ ArgoCD installed (namespace: argocd)
 ✅ OpenFrame app-of-apps deployed
 ✅ Waiting for applications to sync...
 
🎉 Bootstrap Complete!
```

## Verify Your Installation

### Check Kubernetes Cluster

```bash
# View cluster info
kubectl cluster-info

# Check all pods are running
kubectl get pods -A

# View OpenFrame applications
kubectl get applications -n argocd
```

### Access ArgoCD Dashboard

```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port-forward to access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Open [https://localhost:8080](https://localhost:8080) in your browser:
- **Username:** `admin`
- **Password:** (from command above)

### Explore OpenFrame Applications

```bash
# List all applications
kubectl get applications -n argocd

# Check application status
kubectl describe application openframe-app -n argocd

# View application pods
kubectl get pods -n openframe
```

## Hello World Example

Create a simple application to test your OpenFrame environment:

### Create a Test Namespace

```bash
kubectl create namespace hello-world
```

### Deploy a Sample Application

```bash
# Create a simple deployment
cat <<EOF | kubectl apply -f -
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
      - name: hello
        image: nginx:alpine
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world-service
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

### Test the Application

```bash
# Check pods are running
kubectl get pods -n hello-world

# Get service endpoint
kubectl get svc -n hello-world

# Port forward to test locally
kubectl port-forward svc/hello-world-service -n hello-world 8081:80
```

Open [http://localhost:8081](http://localhost:8081) to see the nginx welcome page.

## Expected Output Summary

After successful bootstrap, you should have:

| Component | Status | Location |
|-----------|--------|----------|
| **K3d Cluster** | ✅ Running | `k3d cluster list` |
| **ArgoCD** | ✅ Installed | `kubectl get pods -n argocd` |
| **OpenFrame Apps** | ✅ Deployed | `kubectl get applications -n argocd` |
| **Kubeconfig** | ✅ Updated | `~/.kube/config` |

## Cleanup (Optional)

To remove the OpenFrame environment:

```bash
# Delete the cluster and all resources
openframe cluster delete openframe-cluster

# Or use cleanup command for thorough removal
openframe cluster cleanup openframe-cluster --force
```

## Troubleshooting

### Bootstrap Fails
```bash
# Run with verbose logging
openframe bootstrap --deployment-mode=oss-tenant -v

# Check Docker is running
docker ps

# Check available resources
free -h
df -h
```

### Pods Not Starting
```bash
# Check pod status
kubectl get pods -A

# View pod logs
kubectl logs -n argocd deployment/argocd-server

# Describe problematic pods
kubectl describe pod <pod-name> -n <namespace>
```

### Port Conflicts
```bash
# Find processes using ports
sudo netstat -tlpn | grep :6443
sudo netstat -tlpn | grep :8080

# Use different ports
kubectl port-forward svc/argocd-server -n argocd 9090:443
```

## Next Steps

Congratulations! You now have a working OpenFrame environment. Here's what to explore next:

### Immediate Next Steps
1. **[First Steps](first-steps.md)** - Learn key OpenFrame operations
2. **[Access ArgoCD](https://localhost:8080)** - Explore GitOps workflows  
3. **Deploy Applications** - Try deploying your own applications

### Advanced Topics
- **[Development Workflows](../development/setup/local-development.md)** - Set up local development
- **[Architecture Overview](../development/architecture/overview.md)** - Understanding OpenFrame internals
- **[Contributing](../development/contributing/guidelines.md)** - Contribute to OpenFrame

### Get Help
- Run `openframe --help` for command documentation
- Check `openframe <command> --help` for specific command usage
- View logs with `openframe bootstrap -v` for detailed output

---

**🎉 Success!** You've successfully set up OpenFrame CLI and bootstrapped your first environment. The platform is now ready for application deployments and GitOps workflows.