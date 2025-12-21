# First Steps Guide

Now that you have OpenFrame CLI installed and running, let's explore the key features and get you productive with your new Kubernetes development environment.

## Your First 5 Tasks

### 1. Explore Your Cluster

Start by understanding what was created during bootstrap:

```bash
# View cluster information
openframe cluster status

# List all available clusters  
openframe cluster list

# Check Kubernetes resources
kubectl get nodes
kubectl get namespaces
kubectl get pods --all-namespaces
```

**What you'll see:**
- K3d cluster with master and worker nodes
- System namespaces (kube-system, argocd, etc.)
- ArgoCD pods running and ready

### 2. Access the ArgoCD Interface

ArgoCD is your GitOps control center:

```bash
# Get the admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d && echo

# Access ArgoCD UI
open http://localhost:8080
# Login: admin / (password from above)
```

**Explore the ArgoCD UI:**
- Applications tab: See deployed applications
- Repositories: Connect your Git repos  
- Settings: Configure sync policies
- Clusters: View connected clusters

### 3. Deploy Your First Application

Let's deploy a sample application using OpenFrame:

```bash
# Create a simple application manifest
cat <<EOF > sample-app.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: sample-nginx
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
kubectl apply -f sample-app.yaml

# Watch it deploy
kubectl get applications -n argocd
kubectl get pods -w  # Watch pods being created
```

**Result**: You'll see a guestbook application deployed via GitOps!

### 4. Try Development Commands

OpenFrame includes powerful development tools:

```bash
# List available development commands
openframe dev --help

# If you have Telepresence installed, try intercept (optional)
openframe dev intercept --help

# Check what services are available for interception
kubectl get services
```

### 5. Manage Multiple Environments

Create additional clusters for different environments:

```bash
# Create a staging cluster
openframe cluster create staging-cluster --nodes=1 --skip-wizard

# Create a testing cluster with minimal resources
openframe cluster create test-cluster --nodes=1 --skip-wizard

# List all your clusters
openframe cluster list

# Switch between clusters
kubectl config get-contexts
kubectl config use-context k3d-staging-cluster
```

## Key Configuration Tasks

### Configure ArgoCD for Your Projects

#### Add Your Git Repository

1. **Via ArgoCD UI:**
   - Go to Settings → Repositories
   - Click "Connect Repo using HTTPS"
   - Enter your repository URL
   - Add credentials if private

2. **Via CLI:**
```bash
# Add a public repository
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: my-repo
  namespace: argocd
  labels:
    argocd.argoproj.io/secret-type: repository
stringData:
  type: git
  url: https://github.com/your-org/your-repo.git
EOF
```

#### Set Up App-of-Apps Pattern

Create a main application that manages other applications:

```bash
cat <<EOF > app-of-apps.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: app-of-apps
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/kubernetes-apps.git
    targetRevision: HEAD
    path: apps
  destination:
    server: https://kubernetes.default.svc
    namespace: argocd
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
EOF

kubectl apply -f app-of-apps.yaml
```

### Set Up Development Workflows

#### Enable Traffic Interception

If you installed Telepresence:

```bash
# Connect to cluster
telepresence connect

# List interceptable services
telepresence list

# Intercept a service (example)
telepresence intercept sample-nginx --port 8080:80
```

#### Configure Local Development

```bash
# Create a development namespace
kubectl create namespace development

# Set as default namespace for convenience
kubectl config set-context --current --namespace=development

# Create development resources
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: dev-config
  namespace: development
data:
  environment: "local-development"
  debug: "true"
EOF
```

## Common Configuration Patterns

### 1. Resource Quotas and Limits

Set up resource management for development:

```bash
cat <<EOF > dev-resources.yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: dev-quota
  namespace: development
spec:
  hard:
    requests.cpu: "2"
    requests.memory: 4Gi
    limits.cpu: "4"
    limits.memory: 8Gi
    persistentvolumeclaims: "4"
---
apiVersion: v1
kind: LimitRange
metadata:
  name: dev-limits
  namespace: development
spec:
  limits:
  - default:
      cpu: 200m
      memory: 256Mi
    defaultRequest:
      cpu: 100m
      memory: 128Mi
    type: Container
EOF

kubectl apply -f dev-resources.yaml
```

### 2. Ingress Configuration

Set up ingress for your applications:

```bash
cat <<EOF > ingress-setup.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: dev-ingress
  namespace: development
  annotations:
    kubernetes.io/ingress.class: "traefik"
spec:
  rules:
  - host: myapp.local
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
EOF

kubectl apply -f ingress-setup.yaml

# Add to /etc/hosts for local testing
echo "127.0.0.1 myapp.local" | sudo tee -a /etc/hosts
```

### 3. Secrets Management

Set up secret management for development:

```bash
# Create development secrets
kubectl create secret generic app-secrets \
  --from-literal=database-url="postgresql://localhost:5432/myapp" \
  --from-literal=api-key="dev-api-key-123" \
  --namespace=development

# Create config map for non-sensitive config
kubectl create configmap app-config \
  --from-literal=log-level="debug" \
  --from-literal=feature-flag="true" \
  --namespace=development
```

## Exploring Key Features

### Cluster Management Features

| Feature | Command | Purpose |
|---------|---------|---------|
| **Status Monitoring** | `openframe cluster status` | Real-time cluster health |
| **Resource Cleanup** | `openframe cluster cleanup` | Remove unused resources |
| **Multi-cluster** | `openframe cluster list` | Manage multiple environments |
| **Quick Recreation** | `openframe cluster delete && openframe cluster create` | Fresh environment |

### Chart Management Features

| Feature | Command | Purpose |
|---------|---------|---------|
| **ArgoCD Installation** | `openframe chart install` | Manual ArgoCD setup |
| **Chart Updates** | `openframe chart install --upgrade` | Update charts |
| **Custom Values** | `openframe chart install --values values.yaml` | Custom configuration |

### Development Features

| Feature | Command | Purpose |
|---------|---------|---------|
| **Traffic Interception** | `openframe dev intercept` | Local development debugging |
| **Live Development** | `openframe dev skaffold` | Continuous deployment |
| **Port Forwarding** | `kubectl port-forward` | Access cluster services |

## Development Best Practices

### 1. Namespace Organization

```bash
# Create environment-specific namespaces
kubectl create namespace development
kubectl create namespace staging  
kubectl create namespace testing

# Set default namespace for development
kubectl config set-context --current --namespace=development
```

### 2. Label Resources

```bash
# Label resources for organization
kubectl label namespace development environment=dev team=backend
kubectl label pods --all app.kubernetes.io/part-of=myapp
```

### 3. Use ConfigMaps and Secrets

```bash
# Separate configuration from code
kubectl create configmap app-config --from-file=config/
kubectl create secret generic app-secrets --from-env-file=secrets.env
```

### 4. Monitor Resource Usage

```bash
# Monitor cluster resources
kubectl top nodes
kubectl top pods

# Check resource quotas
kubectl describe resourcequota -n development
```

## Common Workflows

### Daily Development Workflow

```bash
# 1. Start your development session
openframe cluster status
kubectl config use-context k3d-openframe-dev

# 2. Deploy your changes
kubectl apply -f k8s/
# Or use ArgoCD sync via UI

# 3. Test your application
kubectl port-forward svc/my-service 8080:80
curl http://localhost:8080

# 4. Debug if needed
kubectl logs -f deployment/my-app
kubectl describe pod my-app-xxx
```

### Multi-Environment Workflow

```bash
# 1. Develop locally
kubectl config use-context k3d-openframe-dev
kubectl apply -f manifests/

# 2. Test in staging
kubectl config use-context k3d-staging
kubectl apply -f manifests/

# 3. Production deployment via GitOps
git push origin main  # Triggers ArgoCD sync
```

## Where to Get Help

### Built-in Help

```bash
# Command-specific help
openframe bootstrap --help
openframe cluster create --help
openframe dev --help

# Kubernetes help
kubectl explain deployment
kubectl explain service
```

### Common Commands Reference

| Task | Command |
|------|---------|
| **Check cluster health** | `openframe cluster status` |
| **View all resources** | `kubectl get all --all-namespaces` |
| **Debug pod issues** | `kubectl describe pod <pod-name>` |
| **View logs** | `kubectl logs -f <pod-name>` |
| **Access shell** | `kubectl exec -it <pod-name> -- /bin/bash` |
| **Port forward** | `kubectl port-forward svc/<service> 8080:80` |

### Troubleshooting Resources

- **Kubernetes Events**: `kubectl get events --sort-by=.metadata.creationTimestamp`
- **Resource Status**: `kubectl get pods,svc,ingress`
- **ArgoCD Applications**: Check ArgoCD UI for sync status
- **Cluster Resources**: `kubectl describe nodes`

## Next Steps

Now that you've completed your first steps:

### For Application Development
1. **[Local Development Guide](../development/setup/local-development.md)** - Set up your development environment
2. **[Architecture Overview](../development/architecture/overview.md)** - Understand OpenFrame's architecture
3. **[Testing Guide](../development/testing/overview.md)** - Learn testing best practices

### For DevOps/Platform Work
1. **[Contributing Guidelines](../development/contributing/guidelines.md)** - Contribute to OpenFrame
2. **[Environment Setup](../development/setup/environment.md)** - Advanced development configuration

### For Learning More
- Explore ArgoCD documentation for GitOps workflows
- Learn Kubernetes concepts with your running cluster
- Experiment with different deployment patterns
- Try the traffic interception features for debugging

🎉 **Congratulations!** You're now ready to be productive with OpenFrame CLI. Your local Kubernetes environment is configured and ready for development, testing, and learning.