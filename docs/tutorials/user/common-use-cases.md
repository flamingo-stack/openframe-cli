# OpenFrame CLI: Common Use Cases

This guide covers the most frequent scenarios and workflows you'll encounter when using OpenFrame CLI for Kubernetes cluster management and development.

## Overview

OpenFrame CLI is designed to handle various development and deployment scenarios. Whether you're a developer setting up a local environment, a DevOps engineer managing multiple clusters, or a team lead orchestrating deployments, this guide has you covered.

## 1. Setting Up a Local Development Environment

**Scenario**: You want to quickly spin up a local Kubernetes cluster for development and testing.

### Steps

1. **Quick Bootstrap** (Recommended for new users):
   ```bash
   openframe bootstrap my-dev-env
   ```
   This creates a cluster with ArgoCD and essential charts pre-installed.

2. **Manual Cluster Creation** (For more control):
   ```bash
   # Create cluster only
   openframe cluster create my-dev-cluster
   
   # Install charts separately
   openframe chart install --deployment-mode=oss-tenant
   ```

### Best Practices

- Use descriptive cluster names (e.g., `feature-auth-service`, `my-app-dev`)
- Keep development clusters lightweight - delete when not needed
- Use `oss-tenant` mode for local development

> **💡 Tip**: Add `--non-interactive` flag for scripted setups: `openframe bootstrap --non-interactive --deployment-mode=oss-tenant`

---

## 2. Managing Multiple Environments

**Scenario**: You need separate clusters for development, staging, and testing different features.

### Environment Strategy

| Environment | Cluster Name | Deployment Mode | Purpose |
|-------------|--------------|-----------------|---------|
| **Development** | `dev-main` | `oss-tenant` | Daily development work |
| **Feature Testing** | `feature-xyz` | `oss-tenant` | Testing specific features |
| **Staging** | `staging` | `saas-tenant` | Pre-production testing |
| **Demo** | `demo-env` | `saas-shared` | Client demonstrations |

### Workflow

```bash
# Create multiple environments
openframe bootstrap dev-main --deployment-mode=oss-tenant
openframe bootstrap staging --deployment-mode=saas-tenant
openframe bootstrap feature-auth --deployment-mode=oss-tenant

# List all environments
openframe cluster list

# Switch between environments
kubectl config use-context k3d-dev-main
kubectl config use-context k3d-staging
```

### Management Commands

```bash
# Check status of all clusters
for cluster in dev-main staging feature-auth; do
    echo "=== $cluster ==="
    openframe cluster status $cluster
done

# Cleanup unused feature environments
openframe cluster delete feature-old-xyz
```

---

## 3. Team Collaboration Setup

**Scenario**: Your team needs consistent development environments across different machines.

### Standardized Setup Script

Create a team setup script (`team-bootstrap.sh`):

```bash
#!/bin/bash
set -e

CLUSTER_NAME="team-dev-$(whoami)"
DEPLOYMENT_MODE="oss-tenant"

echo "Setting up team development environment..."
echo "Cluster: $CLUSTER_NAME"

# Bootstrap with team settings
openframe bootstrap "$CLUSTER_NAME" \
    --deployment-mode="$DEPLOYMENT_MODE" \
    --non-interactive

echo "✅ Team environment ready!"
echo "Share this script with team members for consistent setup"
```

### Team Best Practices

- **Naming Convention**: Use `team-dev-<username>` for personal clusters
- **Shared Configuration**: Store custom Helm values in version control
- **Resource Limits**: Set Docker resource limits to prevent conflicts
- **Cleanup Schedule**: Delete clusters older than 7 days

<details>
<summary>📋 Team Checklist</summary>

- [ ] All team members have prerequisites installed
- [ ] Shared cluster naming convention agreed upon
- [ ] Custom Helm values repository created
- [ ] Cleanup schedule established
- [ ] Documentation shared with team

</details>

---

## 4. CI/CD Integration

**Scenario**: Integrate OpenFrame CLI into your continuous integration pipeline.

### GitHub Actions Example

```yaml
name: Deploy to Test Environment
on:
  pull_request:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup OpenFrame CLI
        run: |
          wget https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64
          chmod +x openframe-linux-amd64
          sudo mv openframe-linux-amd64 /usr/local/bin/openframe
      
      - name: Create Test Environment
        run: |
          openframe bootstrap "pr-${{ github.event.number }}" \
            --deployment-mode=oss-tenant \
            --non-interactive
      
      - name: Run Tests
        run: |
          kubectl config use-context k3d-pr-${{ github.event.number }}
          # Your test commands here
      
      - name: Cleanup
        if: always()
        run: |
          openframe cluster delete "pr-${{ github.event.number }}" --force
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Setup Environment') {
            steps {
                sh '''
                    openframe bootstrap "build-${BUILD_NUMBER}" \
                        --deployment-mode=oss-tenant \
                        --non-interactive
                '''
            }
        }
        stage('Deploy & Test') {
            steps {
                sh '''
                    kubectl config use-context k3d-build-${BUILD_NUMBER}
                    # Your deployment and test commands
                '''
            }
        }
    }
    post {
        always {
            sh 'openframe cluster delete "build-${BUILD_NUMBER}" --force || true'
        }
    }
}
```

---

## 5. Troubleshooting and Recovery

**Scenario**: Your cluster is in a bad state or you need to recover from issues.

### Common Recovery Scenarios

#### Cluster Not Responding

```bash
# Check cluster status
openframe cluster status my-cluster

# If cluster exists but not responding
kubectl config use-context k3d-my-cluster
kubectl get nodes
kubectl get pods -A

# Nuclear option: recreate cluster
openframe cluster delete my-cluster
openframe bootstrap my-cluster
```

#### ArgoCD Applications Stuck

```bash
# Check ArgoCD status
kubectl get applications -A

# Force sync applications
kubectl patch application my-app -p '{"operation":{"sync":{}}}' --type merge

# Or reinstall charts
openframe chart install --deployment-mode=oss-tenant
```

#### Resource Cleanup

```bash
# Full cluster cleanup (preserves cluster)
openframe cluster cleanup my-cluster --cleanup-helm --cleanup-docker

# Clean up Docker resources system-wide
docker system prune -af
docker volume prune -f
```

### Emergency Recovery Commands

| Issue | Command | Description |
|-------|---------|-------------|
| **Cluster won't start** | `openframe cluster delete && openframe bootstrap` | Complete recreation |
| **Out of disk space** | `docker system prune -af` | Clean Docker resources |
| **Port conflicts** | `openframe cluster delete && docker ps` | Find conflicting containers |
| **Corrupted state** | `rm -rf ~/.openframe && openframe bootstrap` | Reset all configuration |

---

## 6. Resource Management and Optimization

**Scenario**: You want to optimize resource usage and manage multiple clusters efficiently.

### Resource Monitoring

```bash
# Check Docker resource usage
docker stats

# Check cluster resource usage
kubectl top nodes
kubectl top pods -A

# List all k3d clusters and their resources
k3d cluster list
```

### Optimization Tips

1. **Cluster Sizing**:
   ```bash
   # Create smaller clusters for development
   openframe cluster create small-dev --agents=1
   ```

2. **Regular Cleanup**:
   ```bash
   # Weekly cleanup script
   #!/bin/bash
   # Delete clusters older than 7 days
   openframe cluster list --format=json | jq -r '.[] | select(.age > "7d") | .name' | xargs -I {} openframe cluster delete {}
   ```

3. **Resource Limits**:
   - Set Docker memory limit to 4GB minimum
   - Reserve 1 CPU core for host system
   - Monitor disk usage in `/var/lib/docker`

---

## 7. Advanced Configurations

**Scenario**: You need custom configurations for specific use cases.

### Custom Helm Values

Create `custom-values.yaml`:

```yaml
# ArgoCD customizations
argocd:
  server:
    service:
      type: LoadBalancer
  configs:
    repositories:
      - url: https://github.com/my-org/my-charts
        type: git
```

Apply with:
```bash
openframe chart install --values=custom-values.yaml
```

### Environment-Specific Settings

<details>
<summary>🔧 Production-Ready Settings</summary>

```bash
# Production-like cluster
openframe bootstrap prod-test \
    --deployment-mode=saas-shared \
    --agents=3 \
    --registry-create \
    --port=443:443@loadbalancer
```

</details>

### Network Configurations

```bash
# Cluster with custom networking
openframe cluster create network-test \
    --subnet=172.20.0.0/16 \
    --api-port=6550
```

---

## Tips and Tricks

### Quick Commands

```bash
# Alias for common operations
alias of-list='openframe cluster list'
alias of-status='openframe cluster status'
alias of-bootstrap='openframe bootstrap --non-interactive'

# Quick cluster switch
function switch-cluster() {
    kubectl config use-context "k3d-$1"
    echo "Switched to cluster: $1"
}
```

### Productivity Shortcuts

- **Tab Completion**: Many shells support tab completion for kubectl contexts
- **Context Switching**: Use `kubectx` tool for easier context switching
- **Resource Monitoring**: Use `k9s` for interactive cluster monitoring

### Automation Scripts

Create a `daily-dev-setup.sh`:

```bash
#!/bin/bash
DATE=$(date +%Y%m%d)
CLUSTER_NAME="daily-dev-$DATE"

openframe bootstrap "$CLUSTER_NAME" --non-interactive
echo "Today's development environment: $CLUSTER_NAME"
```

---

## Getting Help

When you encounter issues:

1. **Check Status**: Always start with `openframe cluster status`
2. **Verbose Mode**: Add `-v` flag for detailed output
3. **Logs**: Check container logs with `docker logs k3d-<cluster>-server-0`
4. **Community**: Ask questions in our community channels

Remember: OpenFrame CLI is designed to be forgiving - most issues can be resolved by recreating the cluster with `openframe bootstrap`!