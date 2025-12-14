# API Usage Examples

This guide provides practical examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Table of Contents

- [Available Commands](#available-commands)
- [Authentication](#authentication)
- [Common Use Cases](#common-use-cases)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

## Available Commands

The OpenFrame CLI provides several main command groups:

### Cluster Management
- `openframe cluster create` - Create a new K3d cluster
- `openframe cluster list` - List all clusters
- `openframe cluster status` - Show cluster details
- `openframe cluster delete` - Delete a cluster
- `openframe cluster start` - Start a stopped cluster
- `openframe cluster cleanup` - Clean up cluster resources

### Chart Management
- `openframe chart install` - Install Helm charts and ArgoCD
- `openframe bootstrap` - Bootstrap full OpenFrame installation

### Development Tools
- `openframe dev scaffold` - Run Skaffold for service development
- `openframe dev intercept` - Intercept service traffic with Telepresence

## Authentication

The OpenFrame CLI uses your local Kubernetes configuration for authentication. Ensure your `kubectl` is configured correctly:

```bash
# Verify your current context
kubectl config current-context

# List available contexts
kubectl config get-contexts

# Switch context if needed
kubectl config use-context your-cluster-context
```

## Common Use Cases

### 1. Setting Up a Development Environment

Create and configure a complete development environment:

```bash
# Create a new cluster
openframe cluster create

# Follow the interactive wizard to configure:
# - Cluster name
# - Port mappings
# - Resource limits
# - Registry configuration

# Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# Verify installation
openframe cluster status
```

### 2. Managing Multiple Clusters

Work with multiple development clusters:

```bash
# List all clusters
openframe cluster list

# Example output:
# NAME                STATUS    NODES    VERSION
# openframe-dev       Running   1        v1.27.1-k3s1
# openframe-staging   Running   1        v1.27.1-k3s1

# Check specific cluster status
openframe cluster status openframe-dev

# Start a stopped cluster
openframe cluster start openframe-staging

# Clean up unused clusters
openframe cluster delete openframe-old
```

### 3. Installing and Managing Charts

Deploy applications and services:

```bash
# Install basic charts and ArgoCD
openframe chart install

# Bootstrap with specific deployment mode
openframe bootstrap --deployment-mode=oss-tenant

# Verify chart installation
kubectl get applications -n argocd
```

### 4. Development Workflow

Use development tools for active development:

```bash
# Start Skaffold for continuous development
openframe dev scaffold

# In another terminal, intercept service traffic
openframe dev intercept --service=my-service --port=8080

# This allows you to:
# - Run your service locally
# - Receive traffic from the cluster
# - Debug in your local environment
```

### 5. Cluster Lifecycle Management

Complete cluster lifecycle operations:

```bash
# Create cluster with custom configuration
openframe cluster create \
  --name=my-dev-cluster \
  --api-port=6443 \
  --http-port=80 \
  --https-port=443

# Perform regular maintenance
openframe cluster cleanup

# Stop cluster when not in use
openframe cluster stop my-dev-cluster

# Restart when needed
openframe cluster start my-dev-cluster

# Delete when no longer needed
openframe cluster delete my-dev-cluster
```

### 6. Troubleshooting and Monitoring

Check cluster health and debug issues:

```bash
# Get detailed cluster status
openframe cluster status --verbose

# List all resources
kubectl get all --all-namespaces

# Check cluster events
kubectl get events --all-namespaces --sort-by='.lastTimestamp'

# View logs for specific services
kubectl logs -n argocd deployment/argocd-server
```

## Error Handling

### Common Error Patterns

#### 1. Cluster Creation Failures

```bash
# If cluster creation fails
openframe cluster create
# Error: port 80 already in use

# Solution: Use different ports
openframe cluster create --http-port=8080 --https-port=8443
```

#### 2. Bootstrap Issues

```bash
# If bootstrap fails due to missing dependencies
openframe bootstrap --deployment-mode=oss-tenant
# Error: kubectl not found

# Solution: Install kubectl first
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

#### 3. Development Tool Issues

```bash
# If Skaffold fails to start
openframe dev scaffold
# Error: no Skaffold configuration found

# Solution: Ensure you're in a project with skaffold.yaml
ls -la | grep skaffold.yaml
# If missing, create or navigate to correct directory
```

### Error Recovery Scripts

Create a recovery script for common issues:

```bash
#!/bin/bash
# recovery.sh - OpenFrame error recovery

echo "Checking OpenFrame cluster health..."

# Check if cluster exists
if ! openframe cluster list | grep -q "Running"; then
    echo "No running clusters found. Creating new cluster..."
    openframe cluster create
fi

# Check if ArgoCD is running
if ! kubectl get deployment argocd-server -n argocd &>/dev/null; then
    echo "ArgoCD not found. Running bootstrap..."
    openframe bootstrap --deployment-mode=oss-tenant
fi

# Verify cluster status
openframe cluster status

echo "Recovery complete!"
```

## Best Practices

### 1. Cluster Naming Convention

Use descriptive, consistent naming:

```bash
# Good examples
openframe cluster create --name=feature-auth-dev
openframe cluster create --name=project-staging
openframe cluster create --name=team-integration

# Avoid
openframe cluster create --name=test1
openframe cluster create --name=cluster
```

### 2. Port Management

Plan port allocation to avoid conflicts:

```bash
# Development cluster
openframe cluster create \
  --name=dev \
  --api-port=6443 \
  --http-port=8080 \
  --https-port=8443

# Testing cluster
openframe cluster create \
  --name=test \
  --api-port=6444 \
  --http-port=8081 \
  --https-port=8444
```

### 3. Resource Management

Configure appropriate resource limits:

```bash
# For development (lightweight)
openframe cluster create --agents=1 --memory=2g

# For integration testing (more resources)
openframe cluster create --agents=2 --memory=4g
```

### 4. Regular Maintenance

Create maintenance scripts:

```bash
#!/bin/bash
# maintenance.sh - Regular OpenFrame maintenance

echo "Starting OpenFrame maintenance..."

# Clean up unused resources
openframe cluster cleanup

# Update cluster status
openframe cluster status

# Check for available updates
openframe version --check-updates

echo "Maintenance complete!"
```

### 5. Backup and Recovery

Backup important configurations:

```bash
#!/bin/bash
# backup.sh - Backup OpenFrame configurations

BACKUP_DIR="$HOME/.openframe/backups/$(date +%Y%m%d)"
mkdir -p "$BACKUP_DIR"

# Backup kubectl config
cp ~/.kube/config "$BACKUP_DIR/kubeconfig"

# Export cluster configurations
openframe cluster list --output=yaml > "$BACKUP_DIR/clusters.yaml"

# Export ArgoCD applications
kubectl get applications -n argocd -o yaml > "$BACKUP_DIR/argocd-apps.yaml"

echo "Backup saved to: $BACKUP_DIR"
```

### 6. Environment-Specific Configuration

Use environment variables for different setups:

```bash
# Development environment
export OPENFRAME_ENV=development
export OPENFRAME_REGISTRY=dev-registry.local
export OPENFRAME_NAMESPACE=dev

# Production environment
export OPENFRAME_ENV=production
export OPENFRAME_REGISTRY=prod-registry.company.com
export OPENFRAME_NAMESPACE=production

# Use in commands
openframe bootstrap --deployment-mode=oss-tenant --registry=$OPENFRAME_REGISTRY
```

### 7. Integration with CI/CD

Example GitHub Actions workflow:

```yaml
# .github/workflows/openframe.yml
name: OpenFrame Development

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Install OpenFrame CLI
      run: |
        curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
        sudo mv openframe /usr/local/bin/
    
    - name: Create test cluster
      run: |
        openframe cluster create --name=ci-test
        openframe bootstrap --deployment-mode=oss-tenant
    
    - name: Run tests
      run: |
        # Your test commands here
        kubectl get pods --all-namespaces
    
    - name: Cleanup
      if: always()
      run: |
        openframe cluster delete ci-test
```

This guide provides a comprehensive overview of using the OpenFrame CLI effectively. For more specific use cases or advanced configurations, refer to the [official documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs).