# OpenFrame CLI API Usage Examples

This guide provides practical examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Main CLI Commands

The OpenFrame CLI provides several command groups for different operations:

- `openframe cluster` - Cluster lifecycle management
- `openframe chart` - Helm chart and ArgoCD operations  
- `openframe bootstrap` - Full OpenFrame installation
- `openframe dev` - Development workflow tools

## Authentication

OpenFrame CLI uses your local Kubernetes configuration for authentication:

```bash
# Ensure kubectl is configured for your cluster
kubectl config current-context

# OpenFrame will use the current kubectl context
openframe cluster status
```

For cloud clusters, ensure your cloud provider CLI is authenticated:

```bash
# AWS
aws configure

# Azure
az login

# Google Cloud
gcloud auth login
```

## Common Use Cases

### 1. Local Development Setup

Create a complete local development environment:

```bash
# Create a new K3d cluster with interactive wizard
openframe cluster create

# Alternatively, create with specific configuration
openframe cluster create --name my-cluster --ports "80:80@loadbalancer,443:443@loadbalancer"

# Verify cluster is running
openframe cluster status

# Bootstrap OpenFrame on the cluster
openframe bootstrap --deployment-mode=oss-tenant

# Start development with hot-reload
openframe dev scaffold --filename skaffold.yaml
```

### 2. Cluster Management

List and manage existing clusters:

```bash
# List all clusters
openframe cluster list

# Check specific cluster status
openframe cluster status --name my-cluster

# Start a stopped cluster
openframe cluster start --name my-cluster

# Delete a cluster when done
openframe cluster delete --name my-cluster
```

### 3. Chart Installation

Install and manage Helm charts:

```bash
# Install charts with ArgoCD
openframe chart install

# Install specific chart
openframe chart install --chart-name nginx-ingress --namespace ingress-system

# Check installation status
kubectl get applications -n argocd
```

### 4. Development Workflows

Use development tools for service testing:

```bash
# Scaffold development with Skaffold
openframe dev scaffold --filename skaffold.yaml --port-forward=true

# Intercept traffic for debugging
openframe dev intercept --service my-service --namespace default --port 8080

# Clean up development resources
openframe cluster cleanup
```

### 5. Production Bootstrap

Bootstrap OpenFrame for production environments:

```bash
# Bootstrap with external cluster
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --cluster-type=external \
  --domain=mycompany.com \
  --skip-cluster-creation=true

# Verify installation
kubectl get pods -n openframe-system
```

## Error Handling Patterns

### Check Command Exit Codes

```bash
#!/bin/bash

# Example script with proper error handling
if ! openframe cluster create --name test-cluster; then
    echo "Failed to create cluster"
    exit 1
fi

echo "Cluster created successfully"
```

### Handle Missing Dependencies

```bash
# Check if cluster exists before operations
if ! openframe cluster status --name my-cluster >/dev/null 2>&1; then
    echo "Cluster 'my-cluster' not found. Creating..."
    openframe cluster create --name my-cluster
fi

# Proceed with operations
openframe bootstrap --deployment-mode=oss-tenant
```

### Validate Cluster State

```bash
# Wait for cluster to be ready
timeout=300
elapsed=0

while [ $elapsed -lt $timeout ]; do
    if openframe cluster status --name my-cluster | grep -q "Running"; then
        echo "Cluster is ready"
        break
    fi
    
    echo "Waiting for cluster... ($elapsed/${timeout}s)"
    sleep 10
    elapsed=$((elapsed + 10))
done

if [ $elapsed -ge $timeout ]; then
    echo "Timeout waiting for cluster to be ready"
    exit 1
fi
```

### Recovery from Failed Operations

```bash
# Clean up and retry on failure
cleanup_and_retry() {
    echo "Cleaning up failed installation..."
    openframe cluster cleanup
    
    echo "Retrying bootstrap..."
    openframe bootstrap --deployment-mode=oss-tenant
}

# Try bootstrap with error handling
if ! openframe bootstrap --deployment-mode=oss-tenant; then
    echo "Bootstrap failed, attempting cleanup and retry..."
    cleanup_and_retry
fi
```

## Best Practices

### 1. Use Consistent Naming

```bash
# Use descriptive cluster names
openframe cluster create --name "project-dev-$(whoami)"

# Include environment in names
openframe cluster create --name myapp-staging
```

### 2. Environment-Specific Configuration

```bash
# Development environment
export OPENFRAME_CLUSTER_NAME="dev-cluster"
export OPENFRAME_DEPLOYMENT_MODE="oss-tenant"

openframe cluster create --name $OPENFRAME_CLUSTER_NAME
openframe bootstrap --deployment-mode=$OPENFRAME_DEPLOYMENT_MODE

# Use configuration files for complex setups
openframe bootstrap --config bootstrap-config.yaml
```

### 3. Resource Management

```bash
# Always clean up temporary clusters
trap 'openframe cluster delete --name temp-cluster' EXIT

openframe cluster create --name temp-cluster
# ... do work ...
# Cleanup happens automatically on exit
```

### 4. Monitoring and Validation

```bash
# Validate installation after bootstrap
validate_installation() {
    echo "Validating OpenFrame installation..."
    
    # Check core components
    kubectl get pods -n openframe-system
    kubectl get applications -n argocd
    
    # Verify services are responding
    if kubectl get service -n ingress-nginx ingress-nginx-controller >/dev/null 2>&1; then
        echo "✓ Ingress controller is running"
    else
        echo "✗ Ingress controller not found"
        return 1
    fi
    
    return 0
}

openframe bootstrap --deployment-mode=oss-tenant
validate_installation
```

### 5. Automation Scripts

```bash
#!/bin/bash
# complete-setup.sh - Full environment setup script

set -e  # Exit on any error

CLUSTER_NAME="${1:-dev-cluster}"
DEPLOYMENT_MODE="${2:-oss-tenant}"

echo "Setting up OpenFrame development environment..."
echo "Cluster: $CLUSTER_NAME"
echo "Mode: $DEPLOYMENT_MODE"

# Create cluster
echo "Creating cluster..."
openframe cluster create --name "$CLUSTER_NAME"

# Bootstrap OpenFrame
echo "Bootstrapping OpenFrame..."
openframe bootstrap --deployment-mode="$DEPLOYMENT_MODE"

# Wait for readiness
echo "Waiting for services to be ready..."
kubectl wait --for=condition=ready pod -l app=argocd-server -n argocd --timeout=300s

echo "✓ Environment setup complete!"
echo "Access your cluster with: kubectl config use-context k3d-$CLUSTER_NAME"
```

### 6. Development Workflow Integration

```bash
# Makefile example
.PHONY: dev-start dev-stop dev-clean

dev-start:
	@echo "Starting development environment..."
	openframe cluster create --name dev-cluster || true
	openframe bootstrap --deployment-mode=oss-tenant
	openframe dev scaffold --filename skaffold.yaml

dev-stop:
	@echo "Stopping development environment..."
	openframe cluster cleanup

dev-clean:
	@echo "Cleaning up development environment..."
	openframe cluster delete --name dev-cluster

dev-status:
	@echo "Development environment status:"
	openframe cluster status --name dev-cluster
```

These examples provide a solid foundation for integrating OpenFrame CLI into your development and deployment workflows. Always refer to `openframe --help` and `openframe [command] --help` for the most up-to-date command options and flags.