# OpenFrame CLI API Usage Guide

## Overview

OpenFrame CLI provides a command-line interface for managing Kubernetes clusters and development workflows. This guide covers the main commands, configuration options, and best practices for using the CLI effectively.

## Authentication & Configuration

### Environment Variables

OpenFrame CLI uses environment variables and local configuration files for authentication and settings:

```bash
# Set default cluster context
export OPENFRAME_CLUSTER="my-cluster"

# Set kubeconfig path (optional)
export KUBECONFIG="$HOME/.kube/config"

# Enable debug logging
export OPENFRAME_DEBUG=true
```

### Configuration File

The CLI stores configuration in `~/.openframe/config.yaml`:

```yaml
clusters:
  - name: "development"
    registry: "localhost:5000"
    ports:
      - "80:80"
      - "443:443"
    created_at: "2024-01-15T10:30:00Z"
```

## Main CLI Commands

### 1. Cluster Management

#### Create a New Cluster

```bash
# Interactive cluster creation
openframe cluster create

# Create with specific name
openframe cluster create --name my-dev-cluster

# Create with custom ports
openframe cluster create --name api-cluster --ports 8080:80,8443:443
```

#### List Clusters

```bash
# List all clusters
openframe cluster list

# Output example:
# NAME         STATUS    NODES   CREATED
# development  Running   1       2024-01-15T10:30:00Z
# staging      Stopped   1       2024-01-14T09:15:00Z
```

#### Check Cluster Status

```bash
# Check status of current cluster
openframe cluster status

# Check specific cluster
openframe cluster status --name development
```

#### Start/Stop Clusters

```bash
# Start a stopped cluster
openframe cluster start --name development

# Delete a cluster
openframe cluster delete --name development

# Clean up cluster resources
openframe cluster cleanup --name development
```

### 2. Bootstrap & Chart Management

#### Bootstrap OpenFrame

```bash
# Bootstrap with OSS tenant deployment
openframe bootstrap --deployment-mode=oss-tenant

# Bootstrap with custom values
openframe bootstrap --deployment-mode=oss-tenant --values custom-values.yaml

# Bootstrap with specific chart version
openframe bootstrap --deployment-mode=oss-tenant --chart-version=1.2.3
```

#### Install Charts

```bash
# Install Helm charts and ArgoCD
openframe chart install

# Install with custom namespace
openframe chart install --namespace openframe-system

# Install specific chart
openframe chart install --chart prometheus --namespace monitoring
```

### 3. Development Workflows

#### Scaffold Development Environment

```bash
# Run Skaffold for service development
openframe dev scaffold

# Scaffold specific service
openframe dev scaffold --service user-api

# Scaffold with custom config
openframe dev scaffold --config skaffold-dev.yaml
```

#### Traffic Interception

```bash
# Intercept service traffic with Telepresence
openframe dev intercept --service user-api

# Intercept with port mapping
openframe dev intercept --service user-api --port 8080:3000

# List active intercepts
openframe dev intercept --list
```

## Common Use Cases

### 1. Setting Up a Development Environment

```bash
#!/bin/bash
# Complete development setup script

# Create a new cluster
openframe cluster create --name dev-env

# Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# Verify installation
openframe cluster status --name dev-env

echo "Development environment ready!"
echo "Access dashboard at: https://localhost"
```

### 2. Multi-Environment Management

```bash
#!/bin/bash
# Manage multiple environments

# Create staging environment
openframe cluster create --name staging --ports 8080:80,8443:443

# Create production-like environment  
openframe cluster create --name prod-test --ports 9080:80,9443:443

# Switch between environments
export OPENFRAME_CLUSTER="staging"
openframe cluster status

export OPENFRAME_CLUSTER="prod-test"
openframe cluster status
```

### 3. CI/CD Pipeline Integration

```bash
#!/bin/bash
# Example CI/CD script

set -e

# Create ephemeral test cluster
CLUSTER_NAME="test-$(date +%s)"
openframe cluster create --name "$CLUSTER_NAME"

# Bootstrap minimal environment
openframe bootstrap --deployment-mode=oss-tenant

# Run tests
kubectl apply -f test-manifests/
kubectl wait --for=condition=ready pod -l app=test-runner --timeout=300s

# Run test suite
kubectl exec -it deployment/test-runner -- npm test

# Cleanup
openframe cluster delete --name "$CLUSTER_NAME"
```

### 4. Service Development with Telepresence

```bash
#!/bin/bash
# Local service development

# Start intercept for user service
openframe dev intercept --service user-api --port 3000:3000 &

# Start local development server
cd user-service/
npm run dev &

# The service will receive traffic from the cluster
# while running locally for fast development cycles

# Cleanup when done
openframe dev intercept --stop --service user-api
```

## Error Handling Patterns

### 1. Check Command Success

```bash
#!/bin/bash

if openframe cluster create --name test-cluster; then
    echo "✓ Cluster created successfully"
else
    echo "✗ Failed to create cluster"
    exit 1
fi
```

### 2. Validate Cluster State

```bash
#!/bin/bash

check_cluster_ready() {
    local cluster_name="$1"
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for cluster '$cluster_name' to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if openframe cluster status --name "$cluster_name" | grep -q "Running"; then
            echo "✓ Cluster is ready"
            return 0
        fi
        
        echo "Attempt $attempt/$max_attempts: Cluster not ready yet..."
        sleep 10
        ((attempt++))
    done
    
    echo "✗ Cluster failed to become ready within timeout"
    return 1
}

# Usage
openframe cluster create --name my-cluster
check_cluster_ready "my-cluster"
```

### 3. Handle Missing Dependencies

```bash
#!/bin/bash

check_dependencies() {
    local missing_deps=()
    
    command -v docker >/dev/null 2>&1 || missing_deps+=("docker")
    command -v kubectl >/dev/null 2>&1 || missing_deps+=("kubectl")
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo "✗ Missing dependencies: ${missing_deps[*]}"
        echo "Please install the required tools before running OpenFrame CLI"
        exit 1
    fi
    
    echo "✓ All dependencies found"
}

check_dependencies
openframe cluster create
```

### 4. Graceful Error Recovery

```bash
#!/bin/bash

cleanup_on_error() {
    local cluster_name="$1"
    echo "Cleaning up due to error..."
    openframe cluster delete --name "$cluster_name" 2>/dev/null || true
}

deploy_with_cleanup() {
    local cluster_name="test-$(date +%s)"
    
    # Set trap to cleanup on error
    trap "cleanup_on_error $cluster_name" ERR
    
    # Create and configure cluster
    openframe cluster create --name "$cluster_name"
    openframe bootstrap --deployment-mode=oss-tenant
    
    # If we get here, clear the trap
    trap - ERR
    echo "✓ Deployment successful"
}

deploy_with_cleanup
```

## Best Practices

### 1. Use Configuration Files

Store cluster configurations in version control:

```yaml
# .openframe/cluster-configs/development.yaml
name: development
ports:
  - "80:80"
  - "443:443"
  - "5432:5432"  # Database access
registry: localhost:5000
```

```bash
# Use configuration file
openframe cluster create --config .openframe/cluster-configs/development.yaml
```

### 2. Environment-Specific Scripts

Create wrapper scripts for different environments:

```bash
#!/bin/bash
# scripts/dev-setup.sh

set -e

echo "Setting up development environment..."

# Create cluster with development settings
openframe cluster create --name development \
    --ports 80:80,443:443,5432:5432

# Bootstrap with development values
openframe bootstrap --deployment-mode=oss-tenant \
    --values configs/dev-values.yaml

# Set up port forwarding for databases
kubectl port-forward svc/postgres 5432:5432 &
kubectl port-forward svc/redis 6379:6379 &

echo "Development environment ready!"
echo "Database: localhost:5432"
echo "Redis: localhost:6379"
echo "Web UI: https://localhost"
```

### 3. Resource Management

Monitor and manage cluster resources:

```bash
#!/bin/bash
# Resource monitoring script

monitor_resources() {
    echo "=== Cluster Resources ==="
    kubectl top nodes 2>/dev/null || echo "Metrics server not available"
    kubectl top pods --all-namespaces 2>/dev/null || echo "Pod metrics not available"
    
    echo "=== Disk Usage ==="
    docker system df
    
    echo "=== Active Clusters ==="
    openframe cluster list
}

# Run monitoring
monitor_resources

# Cleanup if needed
echo "Clean up unused resources? (y/N)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    docker system prune -f
    openframe cluster cleanup --all
fi
```

### 4. Debugging and Logging

Enable debug output when troubleshooting:

```bash
#!/bin/bash

# Enable debug mode
export OPENFRAME_DEBUG=true

# Run command with verbose output
openframe cluster create --name debug-cluster --verbose

# Check logs
kubectl logs -n kube-system -l app=k3d
```

### 5. Backup and Recovery

Implement backup strategies for important data:

```bash
#!/bin/bash
# Backup script

backup_cluster() {
    local cluster_name="$1"
    local backup_dir="backups/$(date +%Y%m%d_%H%M%S)"
    
    mkdir -p "$backup_dir"
    
    # Export cluster configuration
    openframe cluster status --name "$cluster_name" > "$backup_dir/cluster-status.txt"
    
    # Backup Kubernetes resources
    kubectl get all --all-namespaces -o yaml > "$backup_dir/all-resources.yaml"
    
    # Backup persistent volumes
    kubectl get pv,pvc --all-namespaces -o yaml > "$backup_dir/volumes.yaml"
    
    echo "Backup created in $backup_dir"
}

# Usage
backup_cluster "production"
```

This guide provides comprehensive examples for using OpenFrame CLI effectively. Adapt these patterns to your specific use cases and environments.