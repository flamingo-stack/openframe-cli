# First Steps with OpenFrame CLI

Now that you have OpenFrame CLI installed and running, this guide walks you through the essential operations and features to get you productive quickly.

## Your First 5 Tasks

### 1. Explore Your Cluster

Start by understanding what was created during bootstrap:

```bash
# View cluster information
kubectl cluster-info

# List all namespaces
kubectl get namespaces

# Check cluster nodes
kubectl get nodes

# View cluster resources summary
kubectl top nodes  # Requires metrics-server
```

**What you'll see:**
- OpenFrame cluster running on K3d
- ArgoCD namespace with GitOps components
- System namespaces (kube-system, kube-public)
- Your applications deployed via ArgoCD

### 2. Access the ArgoCD Dashboard

ArgoCD is your GitOps control center. Access it to see your applications:

```bash
# Get the admin password
ARGOCD_PASSWORD=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)
echo "ArgoCD Password: $ARGOCD_PASSWORD"

# Port forward to access the UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

**Access ArgoCD:**
1. Open [https://localhost:8080](https://localhost:8080) (accept the self-signed certificate)
2. Login with username: `admin` and the password from above
3. Explore the applications dashboard

[![OpenFrame Preview Webinar](https://img.youtube.com/vi/bINdW0CQbvY/maxresdefault.jpg)](https://www.youtube.com/watch?v=bINdW0CQbvY)

### 3. Understand OpenFrame Applications

View the applications deployed by OpenFrame:

```bash
# List ArgoCD applications
kubectl get applications -n argocd

# Get detailed application info
kubectl describe application openframe-app -n argocd

# Check application sync status
argocd app list  # If argocd CLI is installed

# View application resources
kubectl get all -n openframe
```

**Key Applications:**
- **app-of-apps**: Master application managing other applications
- **openframe-core**: Core OpenFrame services
- **monitoring**: Observability stack (if enabled)
- **ingress**: Traffic routing and SSL termination

### 4. Learn Essential OpenFrame Commands

Master the core OpenFrame CLI commands:

```bash
# Cluster management
openframe cluster list           # List all clusters
openframe cluster status        # Check cluster health
openframe cluster delete <name> # Remove a cluster

# Chart operations  
openframe chart install --help  # View chart installation options

# Development tools
openframe dev scaffold --help   # Learn about scaffolding
openframe dev intercept --help  # Telepresence integration

# Bootstrap variations
openframe bootstrap --deployment-mode=saas-tenant  # Different deployment modes
```

### 5. Deploy Your First Application

Create and deploy a simple application using GitOps:

```bash
# Create application namespace
kubectl create namespace my-app

# Create a simple deployment
cat <<EOF > my-app.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: my-app
  labels:
    app: my-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
---
apiVersion: v1
kind: Service
metadata:
  name: my-app-service
  namespace: my-app
spec:
  selector:
    app: my-app
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
EOF

# Deploy the application
kubectl apply -f my-app.yaml

# Verify deployment
kubectl get pods -n my-app
kubectl get svc -n my-app
```

## Initial Configuration

### Configure kubectl Context

Ensure kubectl is properly configured:

```bash
# View current context
kubectl config current-context

# List all contexts
kubectl config get-contexts

# Switch context if needed
kubectl config use-context k3d-openframe-cluster
```

### Set Default Namespace (Optional)

```bash
# Set default namespace to avoid -n flags
kubectl config set-context --current --namespace=openframe

# Or use kubens if installed
kubens openframe
```

### Configure Shell Completion

Enable command completion for better productivity:

```bash
# For Bash
echo 'source <(openframe completion bash)' >> ~/.bashrc

# For Zsh  
echo 'source <(openframe completion zsh)' >> ~/.zshrc

# For Fish
openframe completion fish | source

# Reload your shell
source ~/.bashrc  # or ~/.zshrc
```

## Key Features to Explore

### Interactive Cluster Management

```bash
# Create a new cluster interactively
openframe cluster create

# Interactive chart installation
openframe chart install

# Guided bootstrap process
openframe bootstrap  # Without flags for full interactive experience
```

### Deployment Modes

Experiment with different deployment configurations:

```bash
# OSS Tenant (Single-tenant, open source)
openframe bootstrap --deployment-mode=oss-tenant

# SaaS Tenant (Multi-tenant SaaS)
openframe bootstrap --deployment-mode=saas-tenant  

# SaaS Shared (Shared infrastructure)
openframe bootstrap --deployment-mode=saas-shared
```

### Development Workflows

If you're developing applications for OpenFrame:

```bash
# Scaffold a new application
openframe dev scaffold my-new-app

# Set up Telepresence for local development
openframe dev intercept my-service
```

## Common Operations

### Checking System Health

```bash
# Overall cluster health
kubectl get nodes
kubectl get pods -A | grep -v Running

# OpenFrame specific health
kubectl get applications -n argocd
kubectl get pods -n openframe

# ArgoCD health
kubectl get pods -n argocd
```

### Viewing Logs

```bash
# OpenFrame application logs
kubectl logs -n openframe deployment/openframe-core

# ArgoCD logs
kubectl logs -n argocd deployment/argocd-server

# Follow logs in real-time
kubectl logs -f -n openframe deployment/openframe-core
```

### Resource Monitoring

```bash
# Node resource usage
kubectl top nodes

# Pod resource usage  
kubectl top pods -A

# Specific namespace usage
kubectl top pods -n openframe
```

## Useful Aliases and Shortcuts

Add these to your shell profile for productivity:

```bash
# Add to ~/.bashrc or ~/.zshrc
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
alias kgn='kubectl get namespaces'
alias kdp='kubectl describe pod'
alias kl='kubectl logs'
alias kpf='kubectl port-forward'

# OpenFrame specific
alias of='openframe'
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofch='openframe chart'
```

## Integration Points

### Git Repository Setup

For GitOps workflows, you'll want to connect your Git repositories:

```bash
# Configure Git for ArgoCD
kubectl create secret generic repo-credentials \
  --from-literal=url=https://github.com/your-org/your-repo \
  --from-literal=username=your-username \
  --from-literal=password=your-token \
  -n argocd

# Label the secret for ArgoCD
kubectl label secret repo-credentials -n argocd argocd.argoproj.io/secret-type=repository
```

### Container Registry Access

For private container registries:

```bash
# Create registry secret
kubectl create secret docker-registry my-registry-secret \
  --docker-server=your-registry.com \
  --docker-username=your-username \
  --docker-password=your-password \
  --docker-email=your-email@domain.com \
  -n your-namespace
```

## Troubleshooting Quick Reference

### Common Issues and Solutions

| Issue | Command | Solution |
|-------|---------|----------|
| **Pod not starting** | `kubectl describe pod <name>` | Check events and resource constraints |
| **Service unreachable** | `kubectl get svc` | Verify service selectors and endpoints |
| **ArgoCD sync failed** | Check ArgoCD UI | Review application logs and Git repository |
| **Cluster unresponsive** | `k3d cluster stop/start` | Restart the cluster |
| **Port conflicts** | `sudo netstat -tlpn \| grep :port` | Find and stop conflicting processes |

### Quick Diagnostics

```bash
# Comprehensive cluster check
kubectl get all -A

# Check for failed pods
kubectl get pods -A | grep -v Running

# View recent events
kubectl get events -A --sort-by='.lastTimestamp' | tail -20

# Check cluster resources
kubectl describe nodes
```

## Where to Get Help

### Built-in Help

```bash
# Command help
openframe --help
openframe cluster --help
openframe bootstrap --help

# Verbose output for troubleshooting
openframe bootstrap -v
openframe cluster create -v
```

### Documentation

- **Architecture**: `../development/architecture/overview.md`
- **Advanced Setup**: `../development/setup/environment.md`  
- **Troubleshooting**: `../development/troubleshooting/common-issues.md`
- **Contributing**: `../development/contributing/guidelines.md`

### Community Resources

- **GitHub Issues**: Report bugs and feature requests
- **Discussions**: Community support and questions
- **Documentation**: Browse the complete documentation

## Next Learning Paths

Choose your next area of focus:

### For Platform Engineers
- **[Architecture Deep Dive](../development/architecture/overview.md)** - Understand internal components
- **Advanced Cluster Management** - Multi-cluster setups
- **Custom Deployment Modes** - Extend OpenFrame configurations

### For Developers  
- **[Local Development Setup](../development/setup/local-development.md)** - Development environment
- **Telepresence Integration** - Local debugging workflows
- **Application Scaffolding** - Generate application templates

### For DevOps Teams
- **GitOps Workflows** - Advanced ArgoCD usage
- **Monitoring and Observability** - Metrics and logging
- **CI/CD Integration** - Automated deployments

---

**🎯 You're Ready!** You now have a solid foundation with OpenFrame CLI. The next step is diving deeper into specific areas based on your role and interests. Each area builds upon these fundamentals.