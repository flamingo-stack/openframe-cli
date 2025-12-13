# OpenFrame CLI - API Usage Examples

This guide provides comprehensive examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Available Commands Overview

The OpenFrame CLI provides three main command categories:

| Category | Purpose | Commands |
|----------|---------|----------|
| **Cluster Management** | Manage K3d clusters | `create`, `list`, `status`, `delete`, `start`, `cleanup` |
| **Chart Management** | Install and manage Helm charts | `install`, `bootstrap` |
| **Development** | Development workflow tools | `scaffold`, `intercept` |

## Authentication

OpenFrame CLI uses your local Kubernetes configuration and doesn't require separate authentication. Ensure you have:

- `kubectl` configured and accessible
- Docker running (for K3d clusters)
- Proper permissions for cluster operations

```bash
# Verify prerequisites
kubectl version --client
docker version
```

## Cluster Management Examples

### Creating a New Cluster

```bash
# Create cluster with interactive wizard
openframe cluster create

# Create cluster with specific name
openframe cluster create --name my-dev-cluster

# Create cluster with custom configuration
openframe cluster create --name production --nodes 3 --memory 4g
```

### Listing and Checking Clusters

```bash
# List all clusters
openframe cluster list

# Example output:
# NAME          STATUS    NODES    AGE
# dev-cluster   Running   1        2h
# test-env      Stopped   2        1d

# Get detailed cluster status
openframe cluster status

# Check specific cluster
openframe cluster status --name dev-cluster
```

### Managing Cluster Lifecycle

```bash
# Start a stopped cluster
openframe cluster start --name dev-cluster

# Delete a cluster (with confirmation)
openframe cluster delete --name old-cluster

# Clean up cluster resources
openframe cluster cleanup --name dev-cluster
```

## Chart Management Examples

### Installing Charts and ArgoCD

```bash
# Install basic charts
openframe chart install

# Install with specific configuration
openframe chart install --namespace openframe-system

# Install ArgoCD only
openframe chart install --argocd-only
```

### Bootstrap Complete OpenFrame Installation

```bash
# Bootstrap with OSS tenant mode
openframe bootstrap --deployment-mode=oss-tenant

# Bootstrap with custom namespace
openframe bootstrap --deployment-mode=oss-tenant --namespace=my-openframe

# Bootstrap with specific cluster
openframe bootstrap --deployment-mode=oss-tenant --cluster=production
```

## Development Workflow Examples

### Service Development with Skaffold

```bash
# Start development with Skaffold
openframe dev scaffold

# Scaffold specific service
openframe dev scaffold --service=user-api

# Scaffold with custom skaffold.yaml
openframe dev scaffold --config=./custom-skaffold.yaml
```

### Traffic Interception with Telepresence

```bash
# Intercept service traffic
openframe dev intercept --service=payment-service

# Intercept with specific port
openframe dev intercept --service=payment-service --port=8080

# Intercept with local development server
openframe dev intercept --service=payment-service --local-port=3000
```

## Error Handling Patterns

### Common Error Scenarios

```bash
# Handle cluster creation failures
openframe cluster create --name test-cluster
# Error: cluster 'test-cluster' already exists
# Solution: Use different name or delete existing cluster

# Handle missing dependencies
openframe bootstrap --deployment-mode=oss-tenant
# Error: kubectl not found in PATH
# Solution: Install kubectl and ensure it's accessible

# Handle insufficient resources
openframe cluster create --nodes 5 --memory 8g
# Error: insufficient system resources
# Solution: Reduce resource requirements or free up system resources
```

### Error Response Format

OpenFrame CLI provides structured error messages:

```bash
# Example error output
Error: Failed to create cluster 'dev-cluster'
Cause: Port 6443 already in use
Solution: Stop existing cluster or use different port with --api-port flag

# Verbose error information
openframe cluster create --verbose
```

## Best Practices

### 1. Environment-Specific Configurations

```bash
# Use environment-specific naming
openframe cluster create --name dev-$(whoami)
openframe cluster create --name staging-v1.2.0

# Set up environment variables
export OPENFRAME_CLUSTER=dev-cluster
export OPENFRAME_NAMESPACE=openframe-dev
```

### 2. Resource Management

```bash
# Check system resources before creating clusters
docker system df
kubectl top nodes

# Use appropriate resource limits
openframe cluster create --memory 2g --cpus 2  # For development
openframe cluster create --memory 8g --cpus 4  # For testing
```

### 3. Development Workflow Optimization

```bash
# Combine commands for faster setup
openframe cluster create --name dev && \
openframe bootstrap --deployment-mode=oss-tenant && \
openframe dev scaffold

# Use aliases for common operations
alias ofc="openframe cluster"
alias ofb="openframe bootstrap --deployment-mode=oss-tenant"
alias ofd="openframe dev"
```

### 4. Monitoring and Debugging

```bash
# Regular cluster health checks
openframe cluster status --verbose

# Monitor cluster resources
kubectl top nodes
kubectl top pods --all-namespaces

# Debug development workflows
openframe dev scaffold --debug
openframe dev intercept --verbose
```

### 5. Cleanup and Maintenance

```bash
# Regular cleanup routine
openframe cluster cleanup --all
docker system prune -f

# Backup important configurations
kubectl get configmaps -o yaml > cluster-config-backup.yaml
helm list --all-namespaces > installed-charts.txt
```

## Complete Workflow Example

Here's a complete example of setting up a development environment:

```bash
#!/bin/bash
set -e

echo "Setting up OpenFrame development environment..."

# 1. Create development cluster
echo "Creating cluster..."
openframe cluster create --name dev-openframe --memory 4g

# 2. Verify cluster is ready
echo "Checking cluster status..."
openframe cluster status --name dev-openframe

# 3. Bootstrap OpenFrame
echo "Bootstrapping OpenFrame..."
openframe bootstrap --deployment-mode=oss-tenant --namespace=openframe-system

# 4. Wait for services to be ready
echo "Waiting for services..."
kubectl wait --for=condition=ready pod -l app=argocd-server -n argocd --timeout=300s

# 5. Start development workflow
echo "Starting development environment..."
openframe dev scaffold --service=my-service

echo "Development environment ready!"
echo "Access ArgoCD: kubectl port-forward svc/argocd-server -n argocd 8080:443"
```

## Troubleshooting Common Issues

### Cluster Creation Issues

```bash
# Issue: Port conflicts
openframe cluster create --api-port=6444 --name dev-cluster

# Issue: Insufficient memory
openframe cluster create --memory 1g --name minimal-cluster

# Issue: Docker not running
sudo systemctl start docker  # Linux
open /Applications/Docker.app  # macOS
```

### Bootstrap Issues

```bash
# Issue: Missing cluster context
kubectl config current-context
kubectl config use-context k3d-dev-cluster

# Issue: Insufficient permissions
kubectl auth can-i create pods --all-namespaces
```

For more detailed troubleshooting, run commands with the `--verbose` flag to see detailed execution logs.