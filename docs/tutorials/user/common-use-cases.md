# Common Use Cases - OpenFrame CLI

This guide covers the most common scenarios you'll encounter when using OpenFrame CLI. Each use case includes step-by-step instructions, best practices, and troubleshooting tips.

## 🚀 Use Case 1: Setting Up a Development Environment

**Scenario**: You're a developer joining a project and need a complete OpenFrame development environment.

### Step-by-Step Process

1. **Quick Setup (Recommended)**:
   ```bash
   # One command to rule them all
   openframe bootstrap my-dev-env
   
   # Follow the interactive prompts:
   # 1. Select deployment mode: OSS Tenant (for development)
   # 2. Confirm cluster settings
   # 3. Wait for installation (~5-10 minutes)
   ```

2. **Custom Development Setup**:
   ```bash
   # Create cluster with development-optimized settings
   openframe cluster create dev-cluster --nodes 1 --k8s-version v1.25.0
   
   # Install charts with development mode
   openframe chart install dev-cluster --deployment-mode oss-tenant
   ```

### Best Practices
- Use **OSS Tenant** mode for local development
- Allocate at least **4GB RAM** to Docker
- Use descriptive cluster names like `projectname-dev`

### Common Issues
- **Slow startup**: First-time installation downloads many images
- **Port conflicts**: Use `openframe cluster cleanup` to free ports
- **Resource limits**: Increase Docker memory if cluster fails to start

---

## 🔧 Use Case 2: Managing Multiple Projects

**Scenario**: You work on multiple projects and need separate, isolated environments.

### Creating Project-Specific Clusters

```bash
# Project Alpha cluster
openframe bootstrap alpha-project --deployment-mode oss-tenant

# Project Beta cluster  
openframe bootstrap beta-project --deployment-mode saas-tenant

# Switch between projects
export KUBECONFIG=$HOME/.k3d/kubeconfig-alpha-project.yaml
kubectl config current-context
```

### Managing Multiple Environments

| Command | Purpose | Example |
|---------|---------|---------|
| `openframe cluster list` | Show all clusters | See which projects are running |
| `openframe cluster status <name>` | Check specific cluster | Health check for alpha-project |
| `openframe cluster delete <name>` | Remove old cluster | Clean up finished projects |

### Switching Between Projects

<details>
<summary>Click to expand: Advanced kubectl context management</summary>

```bash
# Set up aliases for quick switching
echo 'alias kc-alpha="export KUBECONFIG=$HOME/.k3d/kubeconfig-alpha-project.yaml"' >> ~/.bashrc
echo 'alias kc-beta="export KUBECONFIG=$HOME/.k3d/kubeconfig-beta-project.yaml"' >> ~/.bashrc
source ~/.bashrc

# Quick context switching
kc-alpha && kubectl get pods
kc-beta && kubectl get pods
```
</details>

---

## 🎯 Use Case 3: Live Development and Testing

**Scenario**: You want to develop applications with live reload and real-time feedback.

### Setting Up Development Workflows

1. **Traffic Interception** (for microservices):
   ```bash
   # Intercept traffic to your service
   openframe dev intercept my-api-service
   
   # Your local development server will receive live traffic
   # Make changes and see them instantly in the cluster
   ```

2. **Continuous Development** with Skaffold:
   ```bash
   # Start live development mode
   openframe dev skaffold my-cluster
   
   # Skaffold will:
   # - Build your code on changes
   # - Deploy automatically
   # - Stream logs to your terminal
   ```

### Development Workflow Tips

> **💡 Pro Tip**: Use intercept for debugging production issues and skaffold for new feature development.

```bash
# Recommended development cycle:
openframe dev intercept user-service    # Debug specific service
# ... fix issues locally ...
openframe dev skaffold my-cluster       # Test full application
# ... verify changes work end-to-end ...
```

---

## 🏗️ Use Case 4: Testing Different Deployment Modes

**Scenario**: You need to test how your application behaves in different OpenFrame deployment configurations.

### Deployment Mode Comparison

| Mode | Use Case | Resource Requirements | Best For |
|------|----------|--------------------|----------|
| **OSS Tenant** | Development, testing | Low (1-2 GB) | Local development |
| **SaaS Tenant** | Production simulation | Medium (4-6 GB) | Staging environments |
| **SaaS Shared** | Multi-tenancy testing | High (8+ GB) | Performance testing |

### Testing Each Mode

```bash
# Test OSS Tenant (development)
openframe bootstrap test-oss --deployment-mode oss-tenant --non-interactive

# Test SaaS Tenant (staging-like)
openframe bootstrap test-saas --deployment-mode saas-tenant --non-interactive

# Test SaaS Shared (production-like)  
openframe bootstrap test-shared --deployment-mode saas-shared --non-interactive

# Compare resource usage
docker stats
openframe cluster status test-oss
openframe cluster status test-saas
```

---

## 🔄 Use Case 5: CI/CD Integration

**Scenario**: You want to automate OpenFrame environments in your CI/CD pipeline.

### Non-Interactive Automation

```bash
# Fully automated cluster creation
openframe bootstrap ci-cluster \
  --deployment-mode oss-tenant \
  --non-interactive \
  --verbose

# For GitHub Actions
name: Test with OpenFrame
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Setup OpenFrame
        run: |
          curl -sSL https://install.openframe.io | bash
          openframe bootstrap test-env --non-interactive --deployment-mode oss-tenant
      
      - name: Run Tests
        run: |
          kubectl apply -f tests/manifests/
          kubectl wait --for=condition=ready pod -l app=test-app
```

### CI/CD Best Practices

- Use `--non-interactive` flag for automation
- Set explicit deployment modes
- Include cleanup steps:
  ```bash
  # Cleanup after tests
  openframe cluster delete ci-cluster
  ```

---

## 🚨 Use Case 6: Troubleshooting and Recovery

**Scenario**: Something went wrong and you need to diagnose and fix issues.

### Common Troubleshooting Scenarios

#### Scenario 6a: Cluster Won't Start

```bash
# Check Docker resources
docker system df
docker system prune  # If low on space

# Verbose cluster creation for debugging
openframe cluster create debug-cluster --verbose

# Check k3d logs
k3d cluster list
docker logs k3d-debug-cluster-server-0
```

#### Scenario 6b: ArgoCD Not Working

```bash
# Check ArgoCD status
kubectl get pods -n argocd
kubectl describe pod -n argocd -l app.kubernetes.io/name=argocd-server

# Restart ArgoCD components
kubectl rollout restart deployment/argocd-server -n argocd

# Reset ArgoCD password
kubectl -n argocd delete secret argocd-initial-admin-secret
kubectl rollout restart deployment/argocd-server -n argocd
```

#### Scenario 6c: Complete Environment Reset

```bash
# Nuclear option: clean everything
openframe cluster list  # See all clusters
openframe cluster delete cluster1
openframe cluster delete cluster2
openframe cluster cleanup  # Clean Docker resources

# Fresh start
openframe bootstrap fresh-start
```

### Troubleshooting Commands Reference

| Problem | Command | Expected Result |
|---------|---------|----------------|
| Check cluster health | `openframe cluster status` | All components running |
| View detailed logs | `openframe bootstrap --verbose` | Detailed progress info |
| Clean Docker resources | `openframe cluster cleanup` | Free disk space |
| Reset everything | `openframe cluster delete --all` | Clean slate |

---

## 📊 Use Case 7: Performance Testing and Monitoring

**Scenario**: You want to monitor resource usage and test performance limits.

### Resource Monitoring

```bash
# Monitor cluster resources
kubectl top nodes
kubectl top pods --all-namespaces

# Docker resource usage
docker stats

# Check cluster capacity
openframe cluster status my-cluster --verbose
```

### Load Testing Setup

```bash
# Create cluster optimized for load testing
openframe cluster create load-test --nodes 3

# Install monitoring stack
openframe chart install load-test --deployment-mode saas-shared

# Apply load testing tools
kubectl apply -f https://raw.githubusercontent.com/kubernetes/examples/master/guestbook/redis-master-deployment.yaml
```

---

## 🎯 Tips and Tricks

### Quick Commands for Daily Use

```bash
# My daily OpenFrame aliases (add to ~/.bashrc)
alias ofb='openframe bootstrap'           # Quick bootstrap
alias ofl='openframe cluster list'        # List clusters  
alias ofs='openframe cluster status'      # Cluster status
alias ofi='openframe dev intercept'       # Quick intercept
alias ofc='openframe cluster cleanup'     # Clean resources

# Quick context switching
alias ofctx='kubectl config current-context'
```

### Environment Optimization

> **🔧 Performance Tip**: For faster startups, pre-pull common images:

```bash
# Pre-pull images to avoid delays
docker pull rancher/k3s:latest
docker pull argoproj/argocd:latest
docker pull coredns/coredns:latest
```

### Backup and Restore

```bash
# Backup cluster configuration
kubectl get all --all-namespaces -o yaml > cluster-backup.yaml

# Save cluster state
openframe cluster status my-cluster > cluster-state.txt

# Quick restore (recreate from config)
openframe bootstrap restored-cluster --non-interactive
kubectl apply -f cluster-backup.yaml
```

---

## 📚 Next Steps

Now that you're familiar with common use cases:

1. **Explore Advanced Features**: Check out the [Developer Architecture Guide](../dev/architecture-overview-dev.md)
2. **Customize Your Setup**: Learn about advanced configuration options
3. **Join the Community**: Share your use cases and learn from others
4. **Contribute**: Help improve OpenFrame CLI with your feedback

## Getting Help

For specific use cases not covered here:
- 📖 **Detailed Documentation**: [docs.openframe.io](https://docs.openframe.io)
- 💬 **Ask the Community**: [Discord](https://discord.gg/openframe)  
- 🎯 **Request Examples**: [GitHub Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)

---

**Need more help?** Each command supports `--help` for detailed options:
```bash
openframe --help
openframe cluster create --help
openframe dev intercept --help
```