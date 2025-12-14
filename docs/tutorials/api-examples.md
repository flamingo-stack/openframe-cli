# API Usage Examples Guide

This guide provides practical examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Overview

The OpenFrame CLI provides a command-line interface for:
- Managing K3d Kubernetes clusters
- Installing and configuring OpenFrame components
- Development workflow tools (Skaffold, Telepresence)
- Chart and ArgoCD management

## Authentication

OpenFrame CLI uses your local Kubernetes configuration (`~/.kube/config`) for cluster authentication. No additional authentication is required for local development clusters.

```bash
# Verify kubectl access
kubectl config current-context

# If needed, set the correct context
kubectl config use-context <context-name>
```

## Core API Commands

### 1. Cluster Management

#### Create a New Cluster

```bash
# Interactive cluster creation (recommended)
openframe cluster create

# Create with specific name
openframe cluster create --name my-dev-cluster

# Create with custom configuration
openframe cluster create \
  --name production-like \
  --agents 3 \
  --port 8080
```

#### List and Inspect Clusters

```bash
# List all clusters
openframe cluster list

# Check detailed cluster status
openframe cluster status

# Get status for specific cluster
openframe cluster status --name my-cluster
```

#### Manage Cluster Lifecycle

```bash
# Stop a running cluster
openframe cluster stop --name my-cluster

# Start a stopped cluster
openframe cluster start --name my-cluster

# Delete a cluster (with confirmation)
openframe cluster delete --name my-cluster

# Force delete without confirmation
openframe cluster delete --name my-cluster --force
```

### 2. Bootstrap and Installation

#### Bootstrap OpenFrame

```bash
# Bootstrap with OSS tenant mode (recommended for development)
openframe bootstrap --deployment-mode=oss-tenant

# Bootstrap with custom values
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --namespace=openframe-system \
  --timeout=10m
```

#### Install Charts

```bash
# Install Helm charts and ArgoCD
openframe chart install

# Install with custom repository
openframe chart install --repo-url=https://charts.example.com

# Install specific chart version
openframe chart install --version=1.2.3
```

### 3. Development Workflows

#### Scaffold Development

```bash
# Run Skaffold for active development
openframe dev scaffold

# Scaffold with specific configuration
openframe dev scaffold --config=skaffold.yaml

# Scaffold with custom namespace
openframe dev scaffold --namespace=my-app
```

#### Service Interception

```bash
# Intercept service traffic with Telepresence
openframe dev intercept

# Intercept specific service
openframe dev intercept --service=my-service

# Intercept with custom port mapping
openframe dev intercept \
  --service=my-service \
  --port=8080:80
```

## Common Use Cases

### Use Case 1: Complete Development Environment Setup

```bash
#!/bin/bash

# 1. Create a new development cluster
echo "Creating development cluster..."
openframe cluster create --name dev-environment

# 2. Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
sleep 30

# 3. Bootstrap OpenFrame
echo "Bootstrapping OpenFrame..."
openframe bootstrap --deployment-mode=oss-tenant

# 4. Verify installation
echo "Checking cluster status..."
openframe cluster status

echo "Development environment ready!"
```

### Use Case 2: Multi-Environment Management

```bash
#!/bin/bash

# Create multiple environments
ENVIRONMENTS=("dev" "staging" "testing")

for env in "${ENVIRONMENTS[@]}"; do
    echo "Setting up $env environment..."
    
    # Create cluster
    openframe cluster create --name "$env-cluster"
    
    # Switch context
    kubectl config use-context "k3d-$env-cluster"
    
    # Bootstrap with environment-specific config
    openframe bootstrap \
        --deployment-mode=oss-tenant \
        --namespace="openframe-$env"
done
```

### Use Case 3: Development Workflow with Hot Reload

```bash
#!/bin/bash

# 1. Ensure cluster is running
openframe cluster status || {
    echo "Starting cluster..."
    openframe cluster start
}

# 2. Start development mode
echo "Starting Skaffold development mode..."
openframe dev scaffold --port-forward &

# 3. Wait for services to be ready
sleep 60

# 4. Set up service interception for debugging
echo "Setting up service interception..."
openframe dev intercept \
    --service=my-api \
    --port=3000:8080

echo "Development environment ready for coding!"
```

### Use Case 4: Cleanup and Resource Management

```bash
#!/bin/bash

# Cleanup script for development environments
echo "Cleaning up development resources..."

# Stop all running clusters
openframe cluster list | grep -E "^dev-|^test-" | while read cluster; do
    echo "Stopping cluster: $cluster"
    openframe cluster stop --name "$cluster"
done

# Clean up cluster resources
openframe cluster cleanup

# Remove unused Docker images
docker image prune -f

echo "Cleanup completed!"
```

## Error Handling Patterns

### 1. Cluster Creation Errors

```bash
#!/bin/bash

create_cluster() {
    local cluster_name="$1"
    
    echo "Creating cluster: $cluster_name"
    
    if ! openframe cluster create --name "$cluster_name"; then
        echo "Error: Failed to create cluster $cluster_name"
        
        # Check if cluster already exists
        if openframe cluster list | grep -q "$cluster_name"; then
            echo "Cluster $cluster_name already exists"
            return 0
        fi
        
        # Check Docker daemon
        if ! docker info >/dev/null 2>&1; then
            echo "Error: Docker daemon is not running"
            return 1
        fi
        
        return 1
    fi
    
    echo "Cluster $cluster_name created successfully"
}
```

### 2. Bootstrap Failures

```bash
#!/bin/bash

bootstrap_with_retry() {
    local max_attempts=3
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        echo "Bootstrap attempt $attempt of $max_attempts"
        
        if openframe bootstrap --deployment-mode=oss-tenant; then
            echo "Bootstrap successful!"
            return 0
        fi
        
        echo "Bootstrap failed, attempt $attempt"
        
        if [ $attempt -lt $max_attempts ]; then
            echo "Retrying in 30 seconds..."
            sleep 30
        fi
        
        ((attempt++))
    done
    
    echo "Bootstrap failed after $max_attempts attempts"
    return 1
}
```

### 3. Service Check and Validation

```bash
#!/bin/bash

validate_cluster_health() {
    local cluster_name="$1"
    
    # Check cluster status
    if ! openframe cluster status --name "$cluster_name" >/dev/null 2>&1; then
        echo "Error: Cluster $cluster_name is not accessible"
        return 1
    fi
    
    # Check if nodes are ready
    if ! kubectl get nodes | grep -q "Ready"; then
        echo "Error: No ready nodes found"
        return 1
    fi
    
    # Check if OpenFrame is installed
    if ! kubectl get pods -n openframe-system >/dev/null 2>&1; then
        echo "Warning: OpenFrame may not be installed"
        return 2
    fi
    
    echo "Cluster $cluster_name is healthy"
    return 0
}
```

## Best Practices

### 1. Environment Configuration

```bash
# Use environment variables for configuration
export OPENFRAME_CLUSTER_NAME="my-dev-cluster"
export OPENFRAME_NAMESPACE="openframe-dev"
export OPENFRAME_TIMEOUT="15m"

# Create cluster with environment settings
openframe cluster create \
    --name "$OPENFRAME_CLUSTER_NAME" \
    --namespace "$OPENFRAME_NAMESPACE"
```

### 2. Resource Cleanup

```bash
#!/bin/bash

# Always clean up resources after use
cleanup() {
    echo "Performing cleanup..."
    openframe dev intercept --stop 2>/dev/null || true
    openframe cluster cleanup 2>/dev/null || true
}

# Set trap for cleanup on script exit
trap cleanup EXIT

# Your development work here
openframe dev scaffold
```

### 3. Configuration Management

```bash
# Create reusable configuration files
cat > .openframe-config << EOF
cluster:
  name: development
  agents: 2
  port: 8080

bootstrap:
  deployment-mode: oss-tenant
  namespace: openframe-system
  timeout: 10m
EOF

# Use configuration file
openframe cluster create --config=.openframe-config
```

### 4. Logging and Monitoring

```bash
#!/bin/bash

# Enable verbose logging for troubleshooting
export OPENFRAME_LOG_LEVEL=debug

# Log all operations
log_operation() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a openframe.log
}

log_operation "Starting cluster creation"
openframe cluster create --name dev-cluster 2>&1 | tee -a openframe.log

log_operation "Checking cluster status"
openframe cluster status 2>&1 | tee -a openframe.log
```

### 5. Version and Compatibility Checks

```bash
#!/bin/bash

check_prerequisites() {
    # Check OpenFrame CLI version
    local version=$(openframe --version)
    echo "OpenFrame CLI version: $version"
    
    # Check Docker
    if ! command -v docker >/dev/null 2>&1; then
        echo "Error: Docker is required but not installed"
        return 1
    fi
    
    # Check kubectl
    if ! command -v kubectl >/dev/null 2>&1; then
        echo "Error: kubectl is required but not installed"
        return 1
    fi
    
    echo "All prerequisites met"
    return 0
}

# Run prerequisite check before operations
check_prerequisites || exit 1
```

## Advanced Usage

### Custom Helm Values

```bash
# Create custom values file
cat > custom-values.yaml << EOF
global:
  environment: development
  debug: true

service:
  replicas: 2
  resources:
    requests:
      memory: "256Mi"
      cpu: "250m"
EOF

# Bootstrap with custom values
openframe bootstrap \
    --deployment-mode=oss-tenant \
    --values=custom-values.yaml
```

### Integration with CI/CD

```bash
#!/bin/bash

# CI/CD pipeline integration
if [ "$CI" = "true" ]; then
    # Use non-interactive mode
    export OPENFRAME_NON_INTERACTIVE=true
    
    # Set resource limits for CI
    openframe cluster create \
        --name ci-cluster \
        --agents 1 \
        --memory 2g
else
    # Interactive mode for local development
    openframe cluster create
fi
```

This guide provides comprehensive examples for using the OpenFrame CLI effectively. For additional help, use `openframe --help` or `openframe <command> --help` for command-specific documentation.