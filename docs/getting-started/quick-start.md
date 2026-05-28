# Quick Start Guide

Get your first OpenFrame environment running in under 5 minutes with this streamlined guide.

[![OpenFrame Product Walkthrough (Beta Access)](https://img.youtube.com/vi/awc-yAnkhIo/maxresdefault.jpg)](https://www.youtube.com/watch?v=awc-yAnkhIo)

## TL;DR - 5-Minute Setup

```bash
# 1. Bootstrap complete environment
openframe bootstrap my-first-cluster

# 2. Verify cluster is running
openframe cluster status

# 3. Check ArgoCD installation
kubectl get pods -n argocd
```

That's it! You now have a fully functional Kubernetes cluster with ArgoCD GitOps capabilities.

## Step-by-Step Walkthrough

### Step 1: Bootstrap Your Environment

The `bootstrap` command combines cluster creation and chart installation into one seamless operation:

```bash
openframe bootstrap
```

This command will:
1. Show the OpenFrame logo and welcome message
2. Prompt you for cluster configuration
3. Create a K3d cluster with your settings
4. Install ArgoCD with GitOps configuration
5. Set up the app-of-apps pattern

**Expected output:**
```text
🚀 OpenFrame CLI - MSP Kubernetes Environment Manager

✓ Creating cluster: openframe-dev
✓ Installing ArgoCD charts
✓ Configuring app-of-apps pattern
✓ Environment ready!

Cluster: openframe-dev
ArgoCD URL: https://localhost:8080
```

### Step 2: Verify Your Installation

Check that your cluster is running correctly:

```bash
# View cluster status
openframe cluster status

# List all clusters
openframe cluster list

# Check Kubernetes nodes
kubectl get nodes
```

**Expected output:**
```text
┌─────────────────┬────────┬─────────┬───────────┬─────────┐
│ Cluster Name    │ Status │ Nodes   │ Version   │ Age     │
├─────────────────┼────────┼─────────┼───────────┼─────────┤
│ openframe-dev   │ Ready  │ 1/1     │ v1.27.4   │ 2m      │
└─────────────────┴────────┴─────────┴───────────┴─────────┘
```

### Step 3: Explore ArgoCD

Verify ArgoCD is installed and running:

```bash
# Check ArgoCD pods
kubectl get pods -n argocd

# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Port forward to access ArgoCD UI (optional)
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Access ArgoCD at `https://localhost:8080` with:
- Username: `admin`
- Password: (from the command above)

## Basic "Hello World" Example

Let's deploy a simple application to verify everything works:

```bash
# Create a simple deployment
kubectl create deployment hello-openframe --image=nginx

# Expose the deployment
kubectl expose deployment hello-openframe --port=80 --target-port=80

# Check the deployment
kubectl get deployments
kubectl get services
```

**Expected output:**
```text
NAME              READY   UP-TO-DATE   AVAILABLE   AGE
hello-openframe   1/1     1            1           30s

NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
hello-openframe   ClusterIP   10.43.123.456   <none>        80/TCP    20s
```

## Configuration Options

### Non-Interactive Mode

For CI/CD or automated deployments:

```bash
# Skip all prompts with predefined deployment mode
openframe bootstrap --deployment-mode=oss-tenant --non-interactive
```

### Verbose Mode

For detailed logging and troubleshooting:

```bash
# Show detailed progress including ArgoCD sync
openframe bootstrap --verbose
```

### Custom Cluster Name

```bash
# Specify custom cluster name
openframe bootstrap my-custom-cluster
```

## Deployment Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `oss-tenant` | Open-source tenant configuration | Development and testing |
| `saas-tenant` | SaaS tenant setup | Multi-tenant environments |
| `saas-shared` | SaaS shared configuration | Shared service platforms |

## What You Just Created

Your OpenFrame environment includes:

### Infrastructure Components
- **K3d Cluster**: Lightweight Kubernetes cluster running in Docker
- **ArgoCD**: GitOps continuous delivery platform
- **Helm Charts**: Pre-configured application templates

### Network Configuration
- **Cluster Network**: Internal pod-to-pod communication
- **Service Discovery**: DNS-based service resolution
- **Load Balancing**: K3d load balancer for external access

### GitOps Setup
- **App-of-Apps Pattern**: ArgoCD managing multiple applications
- **Automated Sync**: Continuous deployment from Git repositories
- **Declarative Configuration**: Infrastructure as code approach

## Quick Verification Checklist

Ensure everything is working:

- [ ] `openframe cluster status` shows "Ready"
- [ ] `kubectl get nodes` shows cluster nodes
- [ ] `kubectl get pods -n argocd` shows ArgoCD pods running
- [ ] ArgoCD UI accessible at `https://localhost:8080`
- [ ] Sample deployment successful

## What's Next?

Now that your environment is running:

1. **[First Steps Guide](first-steps.md)** - Explore key features and capabilities
2. **Development Setup** - Configure your development workflow
3. **Application Deployment** - Deploy your first real application via GitOps

## Cleaning Up

When you're done experimenting:

```bash
# Delete the cluster
openframe cluster delete openframe-dev

# Clean up Docker resources
openframe cluster cleanup
```

## Troubleshooting

### Cluster Creation Failed
```bash
# Check Docker is running
docker ps

# Try with verbose logging
openframe bootstrap --verbose
```

### ArgoCD Not Accessible
```bash
# Check ArgoCD pods
kubectl get pods -n argocd

# Restart port-forward
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### Network Issues
```bash
# Reset cluster networking
openframe cluster delete openframe-dev
openframe bootstrap --verbose
```

> 🎉 **Congratulations!** You've successfully set up your first OpenFrame environment. Head to the [First Steps Guide](first-steps.md) to explore what you can do next.