# Common Use Cases for OpenFrame CLI

This guide covers the most frequently asked questions and common scenarios when using OpenFrame CLI. Whether you're setting up development environments, managing multiple clusters, or troubleshooting issues, you'll find practical solutions here.

## Top 10 Common Use Cases

### 1. Setting Up a Development Environment

**Scenario**: You need a local Kubernetes environment for application development.

**Solution**: Use the bootstrap command for a complete setup:

```bash
# Quick development environment
openframe bootstrap dev-env --deployment-mode=oss-tenant

# With verbose output to see what's happening
openframe bootstrap dev-env --deployment-mode=oss-tenant --verbose
```

**What you get**:
- Local K3d cluster named "dev-env"
- ArgoCD installed and configured
- OpenFrame application stack deployed
- Ready-to-use development environment

### 2. Creating Multiple Isolated Environments

**Scenario**: You need separate clusters for different projects or environments (dev, staging, testing).

**Solution**: Create multiple named clusters:

```bash
# Create project-specific environments
openframe cluster create frontend-dev --nodes 3
openframe cluster create backend-dev --nodes 5
openframe cluster create integration-test --nodes 2

# List all your clusters
openframe cluster list
```

**Managing multiple clusters**:
```bash
# Switch between clusters using kubectl context
kubectl config get-contexts
kubectl config use-context k3d-frontend-dev

# Check status of specific cluster
openframe cluster status frontend-dev
```

### 3. Working with Different Deployment Modes

**Scenario**: You need to test different OpenFrame configurations (OSS vs SaaS).

**Solution**: Use different deployment modes:

```bash
# OSS single-tenant (simplest setup)
openframe bootstrap oss-demo --deployment-mode=oss-tenant

# SaaS multi-tenant setup
openframe bootstrap saas-demo --deployment-mode=saas-tenant

# Shared SaaS environment
openframe bootstrap shared-demo --deployment-mode=saas-shared
```

**When to use each mode**:
| Mode | Best For | Use Case |
|------|----------|----------|
| **oss-tenant** | Learning, simple development | Personal projects, tutorials |
| **saas-tenant** | Multi-tenancy development | Building tenant-aware features |
| **saas-shared** | Production-like testing | Performance testing, integration |

### 4. Installing Charts on Existing Clusters

**Scenario**: You have an existing cluster and want to add OpenFrame applications.

**Solution**: Use the chart install command:

```bash
# Install on current kubectl context
openframe chart install --deployment-mode=oss-tenant

# Install with verbose output
openframe chart install --deployment-mode=saas-tenant --verbose

# Non-interactive installation (for automation)
openframe chart install --deployment-mode=oss-tenant --non-interactive
```

### 5. Cleaning Up Resources

**Scenario**: Your cluster is in a bad state or you want to start fresh.

**Solution**: Use cleanup and delete commands:

```bash
# Clean up cluster resources but keep cluster
openframe cluster cleanup my-cluster

# Completely delete a cluster
openframe cluster delete my-cluster

# Clean up all stopped clusters (Docker containers)
docker system prune
```

**Step-by-step cleanup process**:
1. List current clusters: `openframe cluster list`
2. Cleanup specific cluster: `openframe cluster cleanup cluster-name`
3. If problems persist: `openframe cluster delete cluster-name`
4. Recreate: `openframe bootstrap cluster-name --deployment-mode=oss-tenant`

### 6. Troubleshooting Failed Installations

**Scenario**: Bootstrap or chart installation fails partway through.

**Solution**: Diagnose and fix common issues:

```bash
# 1. Run with verbose output to see details
openframe bootstrap my-cluster --deployment-mode=oss-tenant --verbose

# 2. Check prerequisites
docker info
kubectl version --client
helm version

# 3. Check cluster status
openframe cluster status my-cluster
kubectl get nodes
kubectl get pods --all-namespaces

# 4. Clean up and retry if needed
openframe cluster cleanup my-cluster
openframe bootstrap my-cluster --deployment-mode=oss-tenant
```

### 7. Accessing Applications and UIs

**Scenario**: You need to access ArgoCD, applications, or other services in your cluster.

**Solution**: Use port forwarding and service discovery:

```bash
# Access ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Then open https://localhost:8080

# Get ArgoCD admin password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d

# List all services
kubectl get svc --all-namespaces

# Port forward to any service
kubectl port-forward svc/service-name -n namespace 8080:80
```

**Common service ports**:
| Service | Namespace | Port | Local Access |
|---------|-----------|------|--------------|
| ArgoCD | argocd | 443 | https://localhost:8080 |
| Grafana | monitoring | 80 | http://localhost:3000 |
| Prometheus | monitoring | 9090 | http://localhost:9090 |

### 8. Using Dry Run Mode

**Scenario**: You want to see what commands will be executed without actually running them.

**Solution**: Use the `--dry-run` flag:

```bash
# See what bootstrap would do
openframe bootstrap test-cluster --deployment-mode=oss-tenant --dry-run

# Preview cluster creation
openframe cluster create new-cluster --nodes 5 --dry-run

# Preview chart installation
openframe chart install --deployment-mode=saas-tenant --dry-run
```

### 9. Automating with Non-Interactive Mode

**Scenario**: You need to use OpenFrame CLI in scripts or CI/CD pipelines.

**Solution**: Use non-interactive mode with explicit parameters:

```bash
# Non-interactive bootstrap
openframe bootstrap ci-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose

# Script example
#!/bin/bash
set -e

echo "Creating development environment..."
openframe bootstrap dev-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive

echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

echo "Environment ready!"
kubectl get nodes
```

### 10. Managing Cluster Lifecycle

**Scenario**: You need to start, stop, or manage long-running development clusters.

**Solution**: Use cluster management commands:

```bash
# Create a persistent development cluster
openframe cluster create persistent-dev --nodes 3

# Check status anytime
openframe cluster status persistent-dev

# Stop cluster (keeps data)
docker stop $(docker ps -q --filter "label=app=k3d" --filter "label=k3d.cluster=persistent-dev")

# Start stopped cluster
docker start $(docker ps -aq --filter "label=app=k3d" --filter "label=k3d.cluster=persistent-dev")

# Complete status overview
openframe cluster list
```

## Best Practices

### 🎯 Naming Conventions
Use descriptive cluster names that indicate purpose:
- `frontend-dev`, `backend-dev` for component-specific development
- `feature-auth`, `feature-payments` for feature development
- `staging`, `testing` for environment purposes

### 🔄 Resource Management
Monitor your system resources:
```bash
# Check Docker resource usage
docker stats

# Monitor cluster resource usage
kubectl top nodes
kubectl top pods --all-namespaces

# Clean up unused resources
docker system prune
```

### 📊 Monitoring Cluster Health
Regular health checks:
```bash
# Quick health check script
#!/bin/bash
echo "=== Cluster Status ==="
openframe cluster list

echo "=== Node Status ==="
kubectl get nodes

echo "=== Pod Status ==="
kubectl get pods --all-namespaces | grep -v Running | grep -v Completed || echo "All pods healthy!"

echo "=== Storage Usage ==="
df -h
```

## Troubleshooting Quick Reference

### Common Error Messages

| Error | Cause | Solution |
|-------|-------|----------|
| "Docker daemon not running" | Docker not started | Start Docker Desktop or `sudo systemctl start docker` |
| "cluster already exists" | Name collision | Use different name or delete existing: `openframe cluster delete <name>` |
| "port 8080 already in use" | Port conflict | Kill process using port or use different port |
| "timeout waiting for condition" | Slow system/network | Increase timeout or check resources |
| "permission denied" | Docker permissions | Add user to docker group: `sudo usermod -aG docker $USER` |

### Debug Commands

```bash
# Enable maximum verbosity
openframe bootstrap test --deployment-mode=oss-tenant --verbose

# Check Docker status
docker version
docker info

# Check Kubernetes status
kubectl cluster-info
kubectl get nodes -o wide
kubectl describe node

# Check ArgoCD status
kubectl get pods -n argocd
kubectl logs -n argocd deployment/argocd-server

# Network troubleshooting
kubectl get svc --all-namespaces
kubectl get ingress --all-namespaces
```

## Tips and Tricks

### ⚡ Speed Up Development

1. **Keep clusters running**: Don't delete development clusters daily
2. **Use specific contexts**: `kubectl config use-context k3d-dev-cluster`
3. **Alias common commands**:
   ```bash
   alias of='openframe'
   alias k='kubectl'
   alias kgp='kubectl get pods'
   alias kgs='kubectl get svc'
   ```

### 🛠 Development Workflow

```bash
# 1. Create feature branch environment
openframe bootstrap feature-login --deployment-mode=oss-tenant

# 2. Develop and test
# ... your development work ...

# 3. Clean up when done
openframe cluster delete feature-login
```

### 🔍 Advanced Debugging

<details>
<summary>Click to expand advanced debugging techniques</summary>

```bash
# Inspect cluster configuration
k3d config show k3d-my-cluster

# View all cluster resources
kubectl api-resources

# Debug networking
kubectl get endpoints --all-namespaces
kubectl describe service service-name -n namespace

# Check resource quotas and limits
kubectl describe limits --all-namespaces
kubectl top nodes

# Export cluster state
kubectl get all --all-namespaces -o yaml > cluster-state.yaml
```

</details>

## Automation Examples

### CI/CD Pipeline Integration

```yaml
# GitHub Actions example
name: OpenFrame CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup OpenFrame environment
        run: |
          # Install OpenFrame CLI
          curl -LO https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli-linux-amd64
          chmod +x openframe-cli-linux-amd64
          sudo mv openframe-cli-linux-amd64 /usr/local/bin/openframe
          
          # Bootstrap environment
          openframe bootstrap ci-test \
            --deployment-mode=oss-tenant \
            --non-interactive \
            --verbose
      
      - name: Run tests
        run: |
          kubectl wait --for=condition=Ready nodes --all --timeout=300s
          # Your tests here
      
      - name: Cleanup
        run: |
          openframe cluster delete ci-test
```

---

## Next Steps

- **Development Workflows**: Explore `openframe dev` commands for advanced development features
- **Developer Guide**: Check out the [Developer Getting Started Guide](../dev/getting-started-dev.md)
- **Architecture**: Learn about the system in [Architecture Overview](../dev/architecture-overview-dev.md)

> **💡 Pro Tip**: Bookmark this page and use it as a reference while working with OpenFrame CLI. Most common scenarios are covered here with copy-paste ready solutions!