# OpenFrame CLI Common Use Cases

This guide covers the most common scenarios and workflows you'll encounter when using OpenFrame CLI for your development and deployment needs.

## Use Case 1: Setting Up a Development Environment

**Scenario**: You're a new developer who needs to quickly set up a local OpenFrame environment for feature development.

### Steps:

1. **Quick Bootstrap for Development**
   ```bash
   # Create development environment with OSS tenant mode
   openframe bootstrap my-dev --deployment-mode=oss-tenant --verbose
   ```

2. **Verify Your Setup**
   ```bash
   # Check cluster health
   openframe cluster status my-dev
   
   # Verify ArgoCD is running
   kubectl get pods -n argocd
   ```

3. **Access Your Environment**
   - ArgoCD UI will be available at the URL shown in bootstrap output
   - Use provided credentials to log in
   - Verify all applications are synced and healthy

### Best Practices:
- Use descriptive cluster names like `feature-auth` or `bugfix-123`
- Always run with `--verbose` flag initially to understand the process
- Keep your development cluster separate from testing environments

---

## Use Case 2: Managing Multiple Environments

**Scenario**: You need to manage separate environments for development, testing, and demos.

### Steps:

1. **Create Multiple Clusters**
   ```bash
   # Development environment
   openframe bootstrap dev-env --deployment-mode=oss-tenant
   
   # Testing environment  
   openframe bootstrap test-env --deployment-mode=saas-tenant
   
   # Demo environment
   openframe bootstrap demo-env --deployment-mode=oss-tenant
   ```

2. **Switch Between Environments**
   ```bash
   # List all your clusters
   openframe cluster list
   
   # Check specific cluster status
   openframe cluster status dev-env
   openframe cluster status test-env
   ```

3. **Clean Up When Done**
   ```bash
   # Remove old demo environment
   openframe cluster delete demo-env --force
   
   # Clean up unused resources
   openframe cluster cleanup dev-env
   ```

### Environment Naming Conventions:
| Environment | Naming Pattern | Example |
|-------------|---------------|---------|
| **Development** | `dev-<feature>` | `dev-auth`, `dev-ui` |
| **Testing** | `test-<version>` | `test-v2.1`, `test-staging` |
| **Demo** | `demo-<audience>` | `demo-client`, `demo-internal` |

---

## Use Case 3: Local Development with Live Services

**Scenario**: You want to develop a microservice locally while connecting to other services running in your cluster.

### Steps:

1. **Set Up Your Development Cluster**
   ```bash
   openframe bootstrap local-dev --deployment-mode=oss-tenant
   ```

2. **Intercept Traffic to Your Service**
   ```bash
   # Intercept traffic for the service you're developing
   openframe dev intercept my-service
   ```

3. **Start Your Local Development**
   ```bash
   # Run your service locally - traffic will be routed to your local instance
   # while other services continue running in the cluster
   npm start
   # or
   go run main.go
   ```

4. **Test Your Integration**
   - Make requests to your cluster endpoints
   - Your local service will receive the intercepted traffic
   - Other services in cluster handle non-intercepted requests

### Development Workflow Tips:
- Use meaningful intercept names: `openframe dev intercept user-auth-service`
- Always check what's being intercepted: `kubectl get intercepts`
- Stop intercepts when done: `telepresence quit`

---

## Use Case 4: Continuous Development Workflow

**Scenario**: You want automated deployment of your changes to a development cluster as you code.

### Steps:

1. **Prepare Your Cluster**
   ```bash
   openframe bootstrap skaffold-dev --deployment-mode=oss-tenant
   ```

2. **Start Continuous Development**
   ```bash
   # Start Skaffold development workflow
   openframe dev skaffold skaffold-dev
   ```

3. **Develop with Live Updates**
   - Make changes to your code
   - Skaffold automatically builds and deploys changes
   - See updates live in your cluster

### Workflow Benefits:
- ✅ Automatic rebuilds on file changes
- ✅ Live deployment to Kubernetes
- ✅ Log streaming from deployed pods
- ✅ Port forwarding for easy access

---

## Use Case 5: Installing OpenFrame on Existing Cluster

**Scenario**: You already have a Kubernetes cluster and want to add OpenFrame components only.

### Steps:

1. **Connect to Your Existing Cluster**
   ```bash
   # Make sure kubectl is configured for your cluster
   kubectl cluster-info
   ```

2. **Install Only the Charts**
   ```bash
   # Install ArgoCD and OpenFrame applications
   openframe chart install --deployment-mode=oss-tenant
   ```

3. **Customize Installation**
   ```bash
   # Install with specific branch
   openframe chart install --github-branch=develop --verbose
   
   # Non-interactive mode for automation
   openframe chart install --non-interactive
   ```

### When to Use This Approach:
- You have an existing K8s cluster (EKS, GKE, AKS)
- You want to add OpenFrame to a shared cluster
- You're doing production or staging deployments

---

## Use Case 6: Troubleshooting and Maintenance

**Scenario**: Your environment is having issues and you need to diagnose and fix problems.

### Diagnostic Commands:

```bash
# Check overall cluster health
openframe cluster status my-cluster

# View all clusters and their status
openframe cluster list

# Check specific pod issues
kubectl get pods -A
kubectl describe pod <pod-name> -n <namespace>

# Check ArgoCD application status
kubectl get applications -n argocd
```

### Common Fixes:

1. **Restart Problematic Services**
   ```bash
   kubectl rollout restart deployment <deployment-name> -n <namespace>
   ```

2. **Clean Up Resources**
   ```bash
   # Clean unused images and containers
   openframe cluster cleanup my-cluster
   ```

3. **Recreate Cluster if Needed**
   ```bash
   # Delete and recreate (will lose data!)
   openframe cluster delete my-cluster --force
   openframe bootstrap my-cluster --deployment-mode=oss-tenant
   ```

---

## Use Case 7: Team Collaboration Setup

**Scenario**: Multiple team members need consistent development environments.

### Standardized Setup Script:

```bash
#!/bin/bash
# team-setup.sh - Shared script for team environments

# Standard cluster name format
CLUSTER_NAME="${USER}-dev-$(date +%m%d)"

# Create standardized environment
openframe bootstrap $CLUSTER_NAME \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose

echo "✅ Team environment ready: $CLUSTER_NAME"
echo "Share this cluster name with your team: $CLUSTER_NAME"
```

### Team Best Practices:

| Practice | Command Example | Benefit |
|----------|----------------|---------|
| **Naming Convention** | `${USER}-dev-feature` | Easy identification |
| **Shared Configuration** | `--non-interactive` flag | Consistent setups |
| **Documentation** | `--verbose` logs | Debugging help |
| **Resource Cleanup** | Daily `cleanup` commands | Resource management |

---

## Use Case 8: CI/CD Integration

**Scenario**: You want to use OpenFrame CLI in your automated pipelines.

### GitHub Actions Example:

```yaml
name: Deploy to OpenFrame
on: [push]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Install Prerequisites
        run: |
          # Install k3d, helm, kubectl
          curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
          
      - name: Setup OpenFrame Environment
        run: |
          # Download and install OpenFrame CLI
          curl -L -o openframe https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64
          chmod +x openframe
          
      - name: Deploy Application
        run: |
          # Create test environment
          ./openframe bootstrap test-${{ github.sha }} \
            --deployment-mode=oss-tenant \
            --non-interactive \
            --verbose
```

### CI/CD Best Practices:
- Always use `--non-interactive` flag
- Use unique names with commit SHA or build number
- Clean up environments after tests
- Store logs for debugging

---

## Tips and Tricks

### ⚡ Quick Commands Reference

```bash
# Super quick development setup
openframe bootstrap --deployment-mode=oss-tenant

# Check if everything is running
openframe cluster list && kubectl get pods -A

# Quick cleanup
openframe cluster cleanup $(openframe cluster list --quiet)

# Get cluster info quickly
alias of-status='openframe cluster status'
alias of-list='openframe cluster list'
```

### 🔧 Useful Aliases

Add these to your `~/.bashrc` or `~/.zshrc`:

```bash
# OpenFrame shortcuts
alias of='openframe'
alias ofb='openframe bootstrap'
alias ofc='openframe cluster'
alias ofd='openframe dev'

# Kubernetes shortcuts for OpenFrame
alias kgp='kubectl get pods -A'
alias kgs='kubectl get services -A'
alias kga='kubectl get applications -n argocd'
```

### 📊 Monitoring Your Environments

Create a simple monitoring script:

```bash
#!/bin/bash
# monitor-clusters.sh

echo "=== OpenFrame Cluster Status ==="
openframe cluster list

echo -e "\n=== Resource Usage ==="
for cluster in $(openframe cluster list --quiet); do
    echo "Cluster: $cluster"
    kubectl --context k3d-$cluster top nodes 2>/dev/null || echo "  Unable to get metrics"
done

echo -e "\n=== ArgoCD Status ==="
kubectl get applications -n argocd 2>/dev/null || echo "No ArgoCD found"
```

---

## Getting Help

### When Things Go Wrong

1. **Check Prerequisites**: Ensure Docker, k3d, helm, and kubectl are working
2. **Use Verbose Mode**: Add `--verbose` to see detailed output
3. **Check Logs**: Use `kubectl logs` to investigate pod issues
4. **Clean and Retry**: Use cleanup commands before retrying operations
5. **Ask for Help**: Use `openframe <command> --help` for command-specific help

### Community Resources

- 📖 **Built-in Help**: `openframe --help` and `openframe <command> --help`
- 🔍 **Troubleshooting**: Check cluster status and pod logs
- 💬 **Community**: Join discussions and share your use cases
- 🐛 **Bug Reports**: Report issues with detailed reproduction steps

---

**🚀 Ready to tackle your specific use case?** Start with the examples above and adapt them to your needs!