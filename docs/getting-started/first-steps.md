# First Steps with OpenFrame CLI

Now that you have OpenFrame CLI running, let's explore the key features and workflows that will make you productive immediately.

## Your First 5 Actions

### 1. Explore Your Cluster

Start by understanding what you've created:

```bash
# Get detailed cluster information
openframe cluster status

# List all available clusters
openframe cluster list

# Check cluster nodes and resources
kubectl get nodes -o wide
kubectl get namespaces
```

**What you'll see:**
- Cluster health status and resource usage
- Available namespaces including `argocd`
- Node information and networking details

### 2. Navigate the ArgoCD Interface

Access your GitOps dashboard:

```bash
# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# Access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

Open `https://localhost:8080` and explore:
- **Applications**: View deployed applications and their sync status
- **Repositories**: See connected Git repositories
- **Clusters**: Manage target deployment clusters
- **Settings**: Configure projects, repositories, and RBAC

### 3. Deploy Your First Application via GitOps

Create a simple application using ArgoCD:

```bash
# Create application manifest
cat << EOF > my-first-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: guestbook
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/argoproj/argocd-example-apps.git
    targetRevision: HEAD
    path: guestbook
  destination:
    server: https://kubernetes.default.svc
    namespace: default
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
EOF

# Apply the application
kubectl apply -f my-first-app.yaml

# Watch the deployment
kubectl get pods -w
```

### 4. Try Development Workflows

Explore the development commands:

```bash
# View available development tools
openframe dev --help

# List available development commands
openframe dev
```

**Available workflows:**
- **Traffic Interception**: Route cluster traffic to local development
- **Live Reloading**: Deploy with automatic updates on code changes

### 5. Practice Resource Management

Learn essential management commands:

```bash
# Monitor cluster health
openframe cluster status

# Clean up unused resources
openframe cluster cleanup

# View cluster resource usage
kubectl top nodes
kubectl top pods --all-namespaces
```

## Essential Configuration

### Set Up Your Development Environment

Create a development workspace:

```bash
# Create development namespace
kubectl create namespace dev

# Set as default namespace
kubectl config set-context --current --namespace=dev
```

### Configure Git Integration

For GitOps workflows, ensure Git is configured:

```bash
# Configure Git (if not already done)
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"

# Generate SSH key for Git repositories (if needed)
ssh-keygen -t ed25519 -C "your.email@example.com"
```

## Exploring Key Features

### Cluster Management Features

| Feature | Command | Purpose |
|---------|---------|---------|
| **Status Monitoring** | `openframe cluster status` | Check cluster health and resources |
| **Multi-Cluster** | `openframe cluster list` | Manage multiple development clusters |
| **Resource Cleanup** | `openframe cluster cleanup` | Remove unused Docker images and resources |
| **Safe Deletion** | `openframe cluster delete <name>` | Remove clusters with confirmation |

### GitOps Capabilities

- **Automated Deployment**: Applications sync automatically from Git
- **Self-Healing**: ArgoCD corrects configuration drift
- **Rollback Support**: Easy rollback to previous versions
- **Multi-Environment**: Manage dev, staging, and production environments

### Development Workflows

```bash
# Example: Set up traffic interception (requires Telepresence)
openframe dev intercept my-service --port 8080

# Example: Use live development (requires Skaffold)
openframe dev skaffold my-app
```

## Common Initial Configuration

### 1. Customize Cluster Settings

Create a cluster with custom configuration:

```bash
# Interactive cluster creation with custom settings
openframe cluster create my-custom-cluster
```

You'll be prompted for:
- Node configuration
- Network settings  
- Resource limits
- Add-on installations

### 2. Set Up Persistent Storage

Configure persistent volumes for stateful applications:

```bash
# Create storage class
kubectl apply -f - << EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: local-storage
provisioner: kubernetes.io/no-provisioner
volumeBindingMode: WaitForFirstConsumer
EOF
```

### 3. Configure Network Policies

Set up basic security with network policies:

```bash
# Create default deny-all policy
kubectl apply -f - << EOF
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny
  namespace: dev
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
EOF
```

## Helpful Commands Reference

### Daily Operations
```bash
# Quick cluster health check
openframe cluster status

# View all running applications
kubectl get deployments --all-namespaces

# Check ArgoCD application status
kubectl get applications -n argocd

# Monitor cluster resources
kubectl top nodes && kubectl top pods --all-namespaces
```

### Troubleshooting
```bash
# View cluster events
kubectl get events --sort-by='.lastTimestamp'

# Check pod logs
kubectl logs <pod-name> -f

# Describe problematic resources
kubectl describe pod <pod-name>
kubectl describe node <node-name>
```

### Cleanup Commands
```bash
# Remove failed pods
kubectl delete pods --field-selector=status.phase=Failed --all-namespaces

# Clean up completed jobs
kubectl delete jobs --field-selector=status.successful=1 --all-namespaces

# Full cluster cleanup
openframe cluster cleanup
```

## Where to Get Help

### Community Resources
- **OpenMSP Slack**: [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **OpenFrame Website**: [openframe.ai](https://openframe.ai)
- **Flamingo Stack**: [flamingo.run](https://flamingo.run)

### Documentation
- **ArgoCD Documentation**: [argoproj.github.io/argo-cd](https://argo-cd.readthedocs.io/)
- **K3d Documentation**: [k3d.io](https://k3d.io/)
- **Kubernetes Documentation**: [kubernetes.io/docs](https://kubernetes.io/docs/)

### Command Help
```bash
# Get help for any command
openframe --help
openframe bootstrap --help
openframe cluster --help
openframe dev --help

# Get command usage examples
openframe bootstrap --help | grep -A 10 "Examples:"
```

## Next Steps in Your Journey

Now that you're familiar with the basics:

1. **Set up a real application** using GitOps workflows
2. **Explore advanced features** like traffic interception and live development
3. **Configure CI/CD pipelines** that deploy to your cluster
4. **Join the community** to share experiences and get help
5. **Contribute back** by reporting issues or suggesting improvements

> 🚀 **Pro Tip**: Start small with simple applications and gradually add complexity as you become more comfortable with the GitOps workflow and Kubernetes patterns.