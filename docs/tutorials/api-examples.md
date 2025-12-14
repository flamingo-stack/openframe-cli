# API Usage Examples

This guide provides practical examples for using OpenFrame CLI commands. The CLI provides a simple command-line interface for managing Kubernetes clusters and development workflows.

## Main Commands Overview

OpenFrame CLI is organized into several command groups:

- **Cluster Management**: Create, manage, and monitor K3d clusters
- **Chart Management**: Install and manage Helm charts and ArgoCD
- **Development Tools**: Scaffold development environments and intercept traffic
- **Bootstrap**: Full OpenFrame installation and configuration

## Authentication

OpenFrame CLI uses your local Kubernetes configuration (`~/.kube/config`) for cluster authentication. No additional authentication is required for local development clusters.

```bash
# Verify kubectl access
kubectl cluster-info

# OpenFrame will use the current kubectl context
openframe cluster status
```

## Common Use Cases

### 1. Setting Up a Development Environment

**Complete development setup from scratch:**

```bash
# Create a new cluster with interactive wizard
openframe cluster create

# Alternative: Create cluster with specific configuration
openframe cluster create --name my-dev-cluster --workers 2

# Verify cluster is running
openframe cluster status

# Bootstrap OpenFrame on the cluster
openframe bootstrap --deployment-mode=oss-tenant

# Check installation status
kubectl get pods -A
```

### 2. Managing Multiple Clusters

**Working with multiple development clusters:**

```bash
# List all clusters
openframe cluster list

# Create additional clusters for different projects
openframe cluster create --name project-a
openframe cluster create --name project-b

# Switch between clusters (uses kubectl context)
kubectl config use-context k3d-project-a
openframe cluster status

kubectl config use-context k3d-project-b
openframe cluster status
```

### 3. Development Workflow

**Using Skaffold for continuous development:**

```bash
# Navigate to your service directory
cd my-microservice/

# Start development with hot reload
openframe dev scaffold

# In another terminal, check running services
kubectl get services
```

**Traffic interception with Telepresence:**

```bash
# Intercept traffic to a specific service
openframe dev intercept --service user-service --port 8080

# Your local service on port 8080 will receive cluster traffic
# Start your local development server
go run main.go  # or npm start, etc.
```

### 4. Chart and Application Management

**Installing additional charts:**

```bash
# Install Helm charts with ArgoCD
openframe chart install

# Check ArgoCD applications
kubectl get applications -n argocd

# Access ArgoCD UI (get admin password first)
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### 5. Cluster Lifecycle Management

**Starting and stopping clusters:**

```bash
# Stop a cluster (preserves data)
openframe cluster stop --name my-dev-cluster

# Start a stopped cluster
openframe cluster start --name my-dev-cluster

# Clean up cluster resources
openframe cluster cleanup --name my-dev-cluster

# Delete a cluster completely
openframe cluster delete --name my-dev-cluster
```

## Error Handling Patterns

### Common Error Scenarios

**1. Cluster Creation Failures**

```bash
# Check if Docker is running
docker ps

# Verify K3d installation
k3d version

# Create with verbose output for debugging
openframe cluster create --verbose
```

**2. Bootstrap Failures**

```bash
# Check cluster status first
openframe cluster status

# Verify kubectl connectivity
kubectl get nodes

# Check system requirements
openframe cluster status --check-requirements
```

**3. Development Tool Issues**

```bash
# Skaffold not found
which skaffold
# Install if missing: https://skaffold.dev/docs/install/

# Telepresence connection issues
openframe dev intercept --debug
```

### Error Handling Example

```bash
#!/bin/bash

# Robust cluster creation script
create_cluster() {
    echo "Creating OpenFrame cluster..."
    
    # Check prerequisites
    if ! command -v docker &> /dev/null; then
        echo "Error: Docker is not installed or not running"
        exit 1
    fi
    
    if ! command -v k3d &> /dev/null; then
        echo "Error: K3d is not installed"
        exit 1
    fi
    
    # Create cluster with error handling
    if openframe cluster create --name dev-cluster; then
        echo "Cluster created successfully"
        
        # Bootstrap with retry logic
        for i in {1..3}; do
            echo "Bootstrap attempt $i..."
            if openframe bootstrap --deployment-mode=oss-tenant; then
                echo "Bootstrap completed successfully"
                break
            elif [ $i -eq 3 ]; then
                echo "Bootstrap failed after 3 attempts"
                exit 1
            else
                echo "Bootstrap failed, retrying in 30 seconds..."
                sleep 30
            fi
        done
    else
        echo "Failed to create cluster"
        exit 1
    fi
}

create_cluster
```

## Best Practices

### 1. Cluster Management

```bash
# Always check status before operations
openframe cluster status

# Use descriptive cluster names
openframe cluster create --name project-frontend-dev
openframe cluster create --name project-backend-dev

# Regular cleanup of unused clusters
openframe cluster list
openframe cluster delete --name old-cluster
```

### 2. Development Workflow

```bash
# Use project-specific clusters
cd project-a/
openframe cluster create --name project-a-dev
openframe bootstrap --deployment-mode=oss-tenant

# Keep development and testing separate
openframe cluster create --name project-a-test
openframe cluster create --name project-a-dev
```

### 3. Resource Management

```bash
# Monitor cluster resources
kubectl top nodes
kubectl top pods -A

# Clean up resources regularly
openframe cluster cleanup --name dev-cluster

# Stop clusters when not in use
openframe cluster stop --name weekend-project
```

### 4. Configuration Management

```bash
# Export cluster configuration
kubectl config view --minify > cluster-config.yaml

# Use environment-specific configurations
export KUBECONFIG=~/.kube/dev-config
openframe cluster status

export KUBECONFIG=~/.kube/prod-config
openframe cluster status
```

### 5. Debugging and Troubleshooting

```bash
# Enable verbose output for debugging
openframe --verbose cluster create

# Check system status
openframe cluster status --check-all

# View cluster logs
k3d cluster list
docker logs k3d-my-cluster-server-0
```

### 6. Automation Scripts

```bash
#!/bin/bash
# daily-dev-setup.sh

# Start development cluster
openframe cluster start --name daily-dev || {
    echo "Creating new development cluster..."
    openframe cluster create --name daily-dev
    openframe bootstrap --deployment-mode=oss-tenant
}

# Set kubectl context
kubectl config use-context k3d-daily-dev

# Start development services
openframe dev scaffold --background

echo "Development environment ready!"
echo "ArgoCD: http://localhost:8080"
echo "Grafana: http://localhost:3000"
```

### 7. Team Collaboration

```bash
# Shared cluster configuration
cat > team-cluster.yaml << EOF
name: team-shared
workers: 3
ports:
  - "8080:80@loadbalancer"
  - "8443:443@loadbalancer"
EOF

# Create cluster from config
openframe cluster create --config team-cluster.yaml

# Share cluster access
kubectl config view --flatten > shared-kubeconfig.yaml
# Securely share shared-kubeconfig.yaml with team
```

## Quick Reference

### Most Common Commands

```bash
# Quick start sequence
openframe cluster create
openframe bootstrap --deployment-mode=oss-tenant
openframe dev scaffold

# Daily usage
openframe cluster status
openframe cluster list
kubectl get pods -A

# Cleanup
openframe cluster cleanup
openframe cluster stop
```

### Useful Aliases

```bash
# Add to your ~/.bashrc or ~/.zshrc
alias of='openframe'
alias ofc='openframe cluster'
alias ofd='openframe dev'

# Usage
of cluster list
ofc status
ofd scaffold
```