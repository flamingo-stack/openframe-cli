# OpenFrame CLI - API Usage Examples

This guide provides practical examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Main CLI Commands

The OpenFrame CLI provides three main command categories:

### 1. Cluster Management
- `openframe cluster create` - Create new K3d clusters
- `openframe cluster list` - List existing clusters
- `openframe cluster status` - Check cluster health
- `openframe cluster delete` - Remove clusters
- `openframe cluster start/stop` - Control cluster state
- `openframe cluster cleanup` - Clean up resources

### 2. Chart and Bootstrap Management
- `openframe chart install` - Install Helm charts
- `openframe bootstrap` - Full OpenFrame installation

### 3. Development Tools
- `openframe dev scaffold` - Skaffold integration
- `openframe dev intercept` - Telepresence integration

## Authentication

OpenFrame CLI uses your local kubectl configuration for cluster authentication. Ensure you have:

```bash
# Verify kubectl is configured
kubectl config current-context

# Check cluster access
kubectl get nodes
```

## Common Use Cases

### 1. Creating and Setting Up a New Cluster

**Basic cluster creation:**
```bash
# Create a cluster with interactive wizard
openframe cluster create

# Create with specific name
openframe cluster create --name my-dev-cluster

# Create with custom configuration
openframe cluster create \
  --name production-local \
  --workers 3 \
  --memory 4096 \
  --cpu 4
```

**Complete setup workflow:**
```bash
# 1. Create cluster
openframe cluster create --name openframe-dev

# 2. Verify cluster is ready
openframe cluster status --name openframe-dev

# 3. Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# 4. Verify installation
kubectl get pods -A
```

### 2. Managing Multiple Clusters

**List and switch between clusters:**
```bash
# List all OpenFrame clusters
openframe cluster list

# Get detailed status
openframe cluster status --name cluster-1
openframe cluster status --name cluster-2

# Switch kubectl context
kubectl config use-context k3d-cluster-1
```

**Cluster lifecycle management:**
```bash
# Start a stopped cluster
openframe cluster start --name my-cluster

# Stop a running cluster (preserves data)
openframe cluster stop --name my-cluster

# Delete cluster completely
openframe cluster delete --name my-cluster --force
```

### 3. Development Workflow

**Setting up development environment:**
```bash
# Create development cluster
openframe cluster create --name dev-env

# Install required charts
openframe chart install --chart-name argocd
openframe chart install --chart-name prometheus

# Start development with Skaffold
openframe dev scaffold --service my-service

# In another terminal, intercept traffic
openframe dev intercept --service my-service --port 8080
```

**Working with services:**
```bash
# Scaffold with specific configuration
openframe dev scaffold \
  --service user-service \
  --profile development \
  --port-forward 8080:8080

# Intercept with custom routing
openframe dev intercept \
  --service user-service \
  --port 8080 \
  --headers "x-dev-user=john"
```

### 4. Chart and Application Management

**Installing charts step by step:**
```bash
# Install ArgoCD first
openframe chart install \
  --chart-name argocd \
  --namespace argocd \
  --create-namespace

# Install monitoring stack
openframe chart install \
  --chart-name prometheus \
  --namespace monitoring \
  --values custom-values.yaml

# Verify installations
kubectl get pods -n argocd
kubectl get pods -n monitoring
```

**Bootstrap with different modes:**
```bash
# OSS tenant mode (default)
openframe bootstrap --deployment-mode=oss-tenant

# Development mode with additional tools
openframe bootstrap \
  --deployment-mode=development \
  --include-monitoring \
  --include-logging
```

## Error Handling Patterns

### 1. Cluster Creation Errors

```bash
# Check if cluster name already exists
if openframe cluster list | grep -q "my-cluster"; then
  echo "Cluster 'my-cluster' already exists"
  exit 1
fi

# Create with error handling
if ! openframe cluster create --name my-cluster; then
  echo "Failed to create cluster"
  # Cleanup partial resources
  openframe cluster cleanup --name my-cluster
  exit 1
fi
```

### 2. Bootstrap Validation

```bash
# Verify cluster is ready before bootstrap
openframe cluster status --name my-cluster --wait-ready --timeout 300

# Bootstrap with validation
if ! openframe bootstrap --deployment-mode=oss-tenant --validate; then
  echo "Bootstrap failed, checking logs..."
  kubectl logs -n openframe-system -l app=openframe
  exit 1
fi
```

### 3. Development Tool Errors

```bash
# Check if service exists before intercepting
if ! kubectl get service my-service >/dev/null 2>&1; then
  echo "Service 'my-service' not found"
  exit 1
fi

# Intercept with timeout
timeout 30 openframe dev intercept \
  --service my-service \
  --port 8080 || {
  echo "Intercept setup timed out"
  exit 1
}
```

## Best Practices

### 1. Cluster Naming and Organization

```bash
# Use descriptive names with environment
openframe cluster create --name "project-dev"
openframe cluster create --name "project-staging" 
openframe cluster create --name "project-testing"

# Use consistent naming patterns
openframe cluster create --name "${PROJECT_NAME}-${ENVIRONMENT}-$(date +%Y%m%d)"
```

### 2. Resource Management

```bash
# Always check resource requirements
openframe cluster create \
  --name large-cluster \
  --workers 3 \
  --memory 8192 \
  --cpu 6 \
  --disk 50GB

# Monitor resource usage
openframe cluster status --name my-cluster --show-resources
```

### 3. Development Workflow Automation

**Create a development script:**
```bash
#!/bin/bash
set -e

CLUSTER_NAME="dev-$(whoami)"
SERVICE_NAME="${1:-my-service}"

# Setup development environment
echo "Setting up development environment..."

# Create cluster if it doesn't exist
if ! openframe cluster list | grep -q "$CLUSTER_NAME"; then
  openframe cluster create --name "$CLUSTER_NAME"
fi

# Wait for cluster to be ready
openframe cluster status --name "$CLUSTER_NAME" --wait-ready

# Bootstrap if not already done
if ! kubectl get namespace openframe-system >/dev/null 2>&1; then
  openframe bootstrap --deployment-mode=development
fi

# Start development tools
echo "Starting development for service: $SERVICE_NAME"
openframe dev scaffold --service "$SERVICE_NAME" &
sleep 10
openframe dev intercept --service "$SERVICE_NAME" --port 8080

echo "Development environment ready!"
```

### 4. Cleanup and Maintenance

```bash
# Regular cleanup script
#!/bin/bash

# Remove old development clusters
for cluster in $(openframe cluster list --format json | jq -r '.[] | select(.age > "7d") | .name'); do
  echo "Removing old cluster: $cluster"
  openframe cluster delete --name "$cluster" --force
done

# Cleanup Docker resources
docker system prune -f
k3d registry delete k3d-registry || true
```

### 5. Configuration Management

**Using configuration files:**
```bash
# Create cluster config file
cat > cluster-config.yaml << EOF
name: my-project
workers: 2
memory: 4096
cpu: 4
ports:
  - "8080:80@loadbalancer"
  - "8443:443@loadbalancer"
registry:
  create: true
  host: "registry.localhost"
  port: 5000
EOF

# Use configuration
openframe cluster create --config cluster-config.yaml
```

### 6. Troubleshooting Commands

```bash
# Debug cluster issues
openframe cluster status --name my-cluster --verbose

# Check cluster logs
kubectl logs -n kube-system -l k8s-app=kube-dns

# Validate bootstrap
openframe bootstrap --deployment-mode=oss-tenant --dry-run --validate

# Network connectivity test
kubectl run test-pod --image=busybox --rm -it --restart=Never -- nslookup kubernetes
```

This documentation provides a comprehensive guide for using the OpenFrame CLI effectively. Always refer to `openframe --help` or `openframe <command> --help` for the latest command options and flags.