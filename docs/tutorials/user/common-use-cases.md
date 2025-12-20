# OpenFrame CLI - Common Use Cases

This guide covers the most common scenarios and workflows you'll encounter when using OpenFrame CLI for managing Kubernetes clusters and OpenFrame deployments.

## Use Case 1: Setting Up a New Development Environment

**Scenario**: You're a new team member who needs to quickly set up a local OpenFrame environment for development.

### Steps:

1. **One-command setup**:
   ```bash
   openframe bootstrap --deployment-mode oss-tenant
   ```

2. **Verify everything is running**:
   ```bash
   openframe cluster status
   kubectl get pods -A
   ```

3. **Access ArgoCD dashboard**:
   ```bash
   # Get the ArgoCD URL
   kubectl get ingress -n argocd
   
   # Get admin password
   kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d
   ```

**Expected Result**: A fully functional local environment with ArgoCD managing your applications.

---

## Use Case 2: Managing Multiple Clusters

**Scenario**: You need different clusters for different environments (dev, staging, testing).

### Steps:

1. **Create multiple clusters**:
   ```bash
   # Development cluster
   openframe cluster create dev-cluster
   
   # Staging cluster  
   openframe cluster create staging-cluster
   
   # Testing cluster
   openframe cluster create test-cluster
   ```

2. **List all your clusters**:
   ```bash
   openframe cluster list
   ```

3. **Check specific cluster status**:
   ```bash
   openframe cluster status dev-cluster
   ```

4. **Switch between clusters** (using kubectl):
   ```bash
   kubectl config use-context k3d-dev-cluster
   kubectl config use-context k3d-staging-cluster
   ```

### Best Practices:
- Use descriptive names for clusters
- Keep clusters resource-light for local development
- Use `openframe cluster cleanup` periodically to remove unused resources

---

## Use Case 3: Cleaning Up and Starting Fresh

**Scenario**: Your environment is corrupted or you want to start over completely.

### Steps:

1. **Delete a specific cluster**:
   ```bash
   openframe cluster delete my-cluster
   ```

2. **Clean up all unused resources**:
   ```bash
   openframe cluster cleanup
   ```

3. **Verify cleanup**:
   ```bash
   openframe cluster list
   docker ps  # Should show no k3d containers
   ```

4. **Start fresh**:
   ```bash
   openframe bootstrap
   ```

**Tip**: Use this when you encounter persistent issues or want to test installation procedures.

---

## Use Case 4: Installing OpenFrame on Existing Cluster

**Scenario**: You already have a Kubernetes cluster and want to add OpenFrame components.

### Steps:

1. **Ensure your cluster context is active**:
   ```bash
   kubectl config current-context
   ```

2. **Install OpenFrame charts only**:
   ```bash
   openframe chart install --deployment-mode saas-tenant
   ```

3. **Monitor installation progress**:
   ```bash
   # Watch ArgoCD installation
   kubectl get pods -n argocd -w
   
   # Check application sync status
   kubectl get applications -n argocd
   ```

**Note**: This assumes your existing cluster meets OpenFrame requirements.

---

## Use Case 5: Development with Traffic Intercept

**Scenario**: You're developing a microservice and want to intercept cluster traffic to your local development server.

### Steps:

1. **Ensure you have a running cluster with applications**:
   ```bash
   openframe cluster status
   ```

2. **Start traffic intercept**:
   ```bash
   openframe dev intercept my-service
   ```

3. **Run your local development server** on the intercepted port (instructions will be shown)

4. **Test that traffic is being intercepted** by making requests to your cluster

5. **Stop intercept when done**:
   ```bash
   # Follow the stop instructions provided by the intercept command
   ```

**Prerequisites**: 
- Telepresence must be installed (OpenFrame will check this)
- Your service must be deployed in the cluster

---

## Use Case 6: Live Development with Skaffold

**Scenario**: You want to develop with live reloading where code changes are automatically deployed to your cluster.

### Steps:

1. **Navigate to your project directory** with a `skaffold.yaml` file

2. **Start live development**:
   ```bash
   openframe dev scaffold my-app
   ```

3. **Make changes to your code** and watch them automatically deploy

4. **Stop when done** (usually Ctrl+C)

**Prerequisites**:
- Skaffold configuration file in your project
- Container registry accessible from your cluster
- Proper build and deploy configurations

---

## Use Case 7: CI/CD Integration

**Scenario**: You want to use OpenFrame CLI in automated scripts or CI/CD pipelines.

### Non-Interactive Setup:

```bash
# Automated cluster setup
openframe bootstrap \
  --deployment-mode oss-tenant \
  --non-interactive \
  --verbose

# Or just cluster creation
openframe cluster create ci-cluster \
  --skip-wizard \
  --nodes 1 \
  --type k3d
```

### In CI/CD Scripts:

```bash
#!/bin/bash
set -e

# Setup environment
openframe bootstrap --deployment-mode oss-tenant --non-interactive

# Run tests
kubectl apply -f test-resources.yaml
kubectl wait --for=condition=ready pod -l app=test --timeout=300s

# Cleanup
openframe cluster cleanup
```

**Best Practices**:
- Always use `--non-interactive` in scripts
- Use `--verbose` for better logging in CI
- Include proper cleanup steps
- Set timeouts for cluster operations

---

## Use Case 8: Troubleshooting Issues

**Scenario**: Something isn't working correctly and you need to debug.

### Debugging Steps:

1. **Enable verbose output**:
   ```bash
   openframe cluster status --verbose
   ```

2. **Check cluster health**:
   ```bash
   kubectl get nodes
   kubectl get pods -A
   kubectl top nodes  # Resource usage
   ```

3. **Check ArgoCD status**:
   ```bash
   kubectl get applications -n argocd
   kubectl logs -n argocd -l app.kubernetes.io/name=argocd-server
   ```

4. **Review OpenFrame components**:
   ```bash
   kubectl get ingress -A
   kubectl get services -A
   ```

### Common Issues and Solutions:

| Problem | Check | Solution |
|---------|-------|----------|
| Cluster won't start | `docker ps` | Restart Docker, check resources |
| Apps not syncing | ArgoCD logs | Check repository access, credentials |
| Can't access services | Ingress status | Verify ingress controller, DNS |
| Slow performance | `kubectl top nodes` | Increase Docker resources |
| Port conflicts | `netstat -tulpn` | Stop conflicting services |

---

## Use Case 9: Upgrading and Maintenance

**Scenario**: You need to update OpenFrame CLI or maintain your clusters.

### Upgrade Process:

1. **Check current version**:
   ```bash
   openframe --version
   ```

2. **Download latest version** and replace binary

3. **Verify upgrade**:
   ```bash
   openframe --version
   openframe cluster list  # Should still show your clusters
   ```

### Regular Maintenance:

```bash
# Clean up unused resources weekly
openframe cluster cleanup

# Check cluster health
openframe cluster status

# Update Helm repositories
helm repo update

# Check for ArgoCD updates
kubectl get pods -n argocd
```

---

## Tips and Best Practices

### Resource Management
- **Monitor Docker resources**: OpenFrame clusters can be resource-intensive
- **Use appropriate cluster sizes**: Start small and scale as needed
- **Regular cleanup**: Run `openframe cluster cleanup` periodically

### Security
- **Change default passwords**: Always change ArgoCD admin password
- **Use proper RBAC**: Configure appropriate permissions
- **Keep tools updated**: Regularly update OpenFrame CLI and dependencies

### Development Workflow
- **Use descriptive names**: Name clusters and resources clearly
- **Version control configs**: Keep your OpenFrame configurations in git
- **Document setups**: Document any custom configurations for team members

### Automation
- **Script common tasks**: Create scripts for repetitive operations
- **Use environment variables**: Configure defaults via environment
- **Implement health checks**: Add monitoring to automated workflows

---

## Getting More Help

### Command-Specific Help
```bash
openframe bootstrap --help
openframe cluster create --help
openframe dev intercept --help
```

### Verbose Output
Add `--verbose` to any command for detailed logging:
```bash
openframe bootstrap --verbose
```

### Community Resources
- **Documentation**: Check the project wiki for advanced configurations
- **Issues**: Report bugs or request features on GitHub
- **Discussions**: Join community discussions for tips and tricks

---

*Need help with a specific scenario not covered here? Check the [Getting Started Guide](getting-started.md) or reach out to the community!*