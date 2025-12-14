# OpenFrame CLI - API Usage Examples

## Overview

The OpenFrame CLI provides a comprehensive command-line interface for managing Kubernetes clusters and OpenFrame deployments. This guide covers the main commands, usage patterns, and best practices.

## Main CLI Commands

### Cluster Management Commands

#### `openframe cluster create`
Creates a new K3d Kubernetes cluster with OpenFrame optimized configuration.

```bash
# Interactive cluster creation with wizard
openframe cluster create

# Create cluster with specific name
openframe cluster create --name my-cluster

# Create cluster with custom configuration
openframe cluster create --name dev-cluster --k3d-config ./k3d.yaml
```

#### `openframe cluster list`
Lists all available clusters and their status.

```bash
# List all clusters
openframe cluster list

# Example output:
# NAME         STATUS    CREATED
# dev-cluster  running   2 hours ago
# test-env     stopped   1 day ago
```

#### `openframe cluster status`
Shows detailed information about cluster health and components.

```bash
# Check status of current cluster
openframe cluster status

# Check specific cluster
openframe cluster status --name dev-cluster

# Get status with detailed output
openframe cluster status --verbose
```

### Bootstrap and Installation Commands

#### `openframe bootstrap`
Bootstraps OpenFrame components on an existing cluster.

```bash
# Bootstrap with OSS tenant mode
openframe bootstrap --deployment-mode=oss-tenant

# Bootstrap with custom values
openframe bootstrap --deployment-mode=oss-tenant --values ./custom-values.yaml

# Bootstrap specific components only
openframe bootstrap --components=argocd,monitoring
```

#### `openframe chart install`
Installs Helm charts and configures ArgoCD applications.

```bash
# Install default charts
openframe chart install

# Install with custom chart repository
openframe chart install --repo https://charts.example.com

# Install specific chart version
openframe chart install --version 1.2.3
```

### Development Commands

#### `openframe dev scaffold`
Runs Skaffold for continuous development and deployment.

```bash
# Start Skaffold with default configuration
openframe dev scaffold

# Scaffold with custom skaffold.yaml
openframe dev scaffold --config ./skaffold.yaml

# Scaffold specific services
openframe dev scaffold --modules=api,frontend
```

#### `openframe dev intercept`
Sets up Telepresence intercepts for local development.

```bash
# Intercept a service for local development
openframe dev intercept --service api-service --port 8080

# Intercept with custom headers
openframe dev intercept --service web-app --port 3000 --headers "x-dev-user: alice"

# List active intercepts
openframe dev intercept --list
```

## Common Use Cases

### 1. Setting Up a Development Environment

```bash
#!/bin/bash

# Complete development environment setup
echo "Setting up OpenFrame development environment..."

# Create new cluster
openframe cluster create --name dev-env

# Wait for cluster to be ready
openframe cluster status --wait

# Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# Verify installation
openframe cluster status --verbose
```

### 2. Local Service Development

```bash
#!/bin/bash

# Develop a microservice locally while connected to cluster
SERVICE_NAME="user-api"
LOCAL_PORT="8080"

# Start local development with hot reload
echo "Starting development for $SERVICE_NAME..."

# Set up intercept to route traffic to local instance
openframe dev intercept \
  --service $SERVICE_NAME \
  --port $LOCAL_PORT \
  --headers "x-dev-mode: true"

# In another terminal, start Skaffold for other services
openframe dev scaffold --modules=frontend,auth-service
```

### 3. Multi-Environment Management

```bash
#!/bin/bash

# Script to manage multiple environments
ENVIRONMENTS=("dev" "staging" "testing")

for env in "${ENVIRONMENTS[@]}"; do
  echo "Checking environment: $env"
  
  # Check if cluster exists
  if openframe cluster list | grep -q "$env"; then
    echo "✓ Cluster $env exists"
    openframe cluster status --name "$env"
  else
    echo "✗ Cluster $env missing - creating..."
    openframe cluster create --name "$env"
    openframe bootstrap --deployment-mode=oss-tenant
  fi
done
```

### 4. Cluster Cleanup and Maintenance

```bash
#!/bin/bash

# Maintenance script for cluster cleanup
CLUSTER_NAME="dev-cluster"

echo "Performing maintenance on $CLUSTER_NAME..."

# Clean up unused resources
openframe cluster cleanup --name $CLUSTER_NAME

# Restart cluster if needed
openframe cluster status --name $CLUSTER_NAME | grep -q "unhealthy" && {
  echo "Cluster unhealthy, restarting..."
  openframe cluster delete --name $CLUSTER_NAME
  openframe cluster create --name $CLUSTER_NAME
  openframe bootstrap --deployment-mode=oss-tenant
}
```

## Configuration

### Environment Variables

```bash
# Set default cluster name
export OPENFRAME_CLUSTER_NAME="my-cluster"

# Set custom kubeconfig path
export KUBECONFIG="$HOME/.kube/openframe-config"

# Enable debug logging
export OPENFRAME_DEBUG="true"

# Set default deployment mode
export OPENFRAME_DEPLOYMENT_MODE="oss-tenant"
```

### Configuration Files

Create `~/.openframe/config.yaml` for persistent settings:

```yaml
# ~/.openframe/config.yaml
defaultCluster: "dev-cluster"
deploymentMode: "oss-tenant"
k3dConfig:
  image: "rancher/k3s:v1.28.2-k3s1"
  ports:
    - "80:80@loadbalancer"
    - "443:443@loadbalancer"
charts:
  repository: "https://charts.openframe.dev"
  timeout: "10m"
development:
  skaffoldConfig: "./skaffold.yaml"
  telepresenceConfig: "./telepresence.yaml"
```

## Error Handling Patterns

### 1. Command Execution Errors

```bash
#!/bin/bash

# Function to handle command errors gracefully
run_openframe_command() {
  local cmd="$1"
  local description="$2"
  
  echo "Running: $description"
  
  if ! openframe $cmd; then
    echo "❌ Failed: $description"
    echo "Command: openframe $cmd"
    
    # Check common issues
    case "$cmd" in
      "cluster create"*)
        echo "💡 Troubleshooting tips:"
        echo "  - Check Docker is running: docker info"
        echo "  - Verify k3d installation: k3d version"
        echo "  - Check available ports: netstat -tulpn"
        ;;
      "bootstrap"*)
        echo "💡 Troubleshooting tips:"
        echo "  - Verify cluster is running: openframe cluster status"
        echo "  - Check cluster connectivity: kubectl cluster-info"
        echo "  - Ensure sufficient resources: kubectl top nodes"
        ;;
    esac
    
    return 1
  fi
  
  echo "✅ Success: $description"
  return 0
}

# Usage example
run_openframe_command "cluster create --name test" "Creating test cluster" || exit 1
run_openframe_command "bootstrap --deployment-mode=oss-tenant" "Bootstrapping OpenFrame" || exit 1
```

### 2. Cluster Health Checks

```bash
#!/bin/bash

# Comprehensive cluster health check
check_cluster_health() {
  local cluster_name="$1"
  
  echo "🔍 Checking cluster health: $cluster_name"
  
  # Check cluster status
  if ! openframe cluster status --name "$cluster_name" >/dev/null 2>&1; then
    echo "❌ Cluster $cluster_name is not accessible"
    return 1
  fi
  
  # Check if cluster is running
  local status=$(openframe cluster list | grep "$cluster_name" | awk '{print $2}')
  if [ "$status" != "running" ]; then
    echo "⚠️  Cluster $cluster_name status: $status"
    
    if [ "$status" = "stopped" ]; then
      echo "🔄 Starting cluster..."
      openframe cluster start --name "$cluster_name"
    else
      echo "❌ Unexpected cluster status: $status"
      return 1
    fi
  fi
  
  echo "✅ Cluster $cluster_name is healthy"
  return 0
}

# Usage
check_cluster_health "dev-cluster" || {
  echo "❌ Health check failed"
  exit 1
}
```

## Best Practices

### 1. Resource Management

```bash
#!/bin/bash

# Best practices for resource management

# Always specify resource limits when creating clusters
openframe cluster create \
  --name production \
  --memory 8g \
  --cpus 4

# Use cleanup commands regularly
openframe cluster cleanup --name dev-cluster

# Monitor cluster resources
openframe cluster status --verbose
```

### 2. Development Workflow

```bash
#!/bin/bash

# Recommended development workflow

# 1. Create dedicated development cluster
openframe cluster create --name "dev-$(whoami)"

# 2. Bootstrap with development-friendly settings
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --dev-mode

# 3. Use intercepts for active development
openframe dev intercept \
  --service my-service \
  --port 8080 \
  --env-file .env.local

# 4. Run scaffolding for related services
openframe dev scaffold --modules=dependencies
```

### 3. Production Deployment

```bash
#!/bin/bash

# Production deployment best practices

# Use specific versions for reproducible deployments
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --version=1.2.3 \
  --values=production-values.yaml

# Verify deployment health
openframe cluster status --wait --timeout=10m

# Set up monitoring and alerts
openframe chart install \
  --chart=monitoring \
  --values=monitoring-config.yaml
```

### 4. Troubleshooting Commands

```bash
#!/bin/bash

# Common troubleshooting commands

# Get detailed cluster information
openframe cluster status --verbose --debug

# Check cluster logs
openframe cluster logs --follow

# Restart problematic services
openframe cluster cleanup --services=argocd,monitoring

# Reset cluster to clean state
openframe cluster delete --name problematic-cluster
openframe cluster create --name problematic-cluster
openframe bootstrap --deployment-mode=oss-tenant
```

## Integration Examples

### CI/CD Pipeline Integration

```yaml
# .github/workflows/test.yml
name: Test with OpenFrame

on: [push, pull_request]

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
          openframe cluster create --name ci-test-${{ github.run_id }}
          
      - name: Bootstrap OpenFrame
        run: |
          openframe bootstrap --deployment-mode=oss-tenant --wait
          
      - name: Run tests
        run: |
          # Your test commands here
          kubectl apply -f test-resources/
          
      - name: Cleanup
        if: always()
        run: |
          openframe cluster delete --name ci-test-${{ github.run_id }}
```

This comprehensive guide provides practical, copy-paste ready examples for using the OpenFrame CLI effectively in various scenarios, from basic development workflows to production deployments and CI/CD integration.