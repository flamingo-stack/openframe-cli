# API Usage Examples

The OpenFrame CLI provides a comprehensive set of commands for managing Kubernetes clusters and development workflows. This guide covers practical usage examples for all major CLI operations.

## Installation and Setup

### Quick Installation

```bash
# Install latest release (macOS ARM64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz && sudo mv openframe /usr/local/bin/

# Verify installation
openframe --version
```

### Build from Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

## Core API Commands

### 1. Cluster Management

#### Create a New Cluster

```bash
# Interactive cluster creation (recommended)
openframe cluster create

# Create with specific name
openframe cluster create --name my-cluster

# Create with custom configuration
openframe cluster create \
  --name production-cluster \
  --agents 3 \
  --registry-port 5001
```

**Example Output:**
```
✓ Creating K3d cluster 'my-cluster'
✓ Waiting for cluster to be ready
✓ Cluster 'my-cluster' created successfully
```

#### List All Clusters

```bash
# List all clusters
openframe cluster list

# List with detailed output
openframe cluster list --output wide
```

**Example Output:**
```
NAME         STATUS   AGENTS   CREATED
my-cluster   running  1        2 hours ago
dev-cluster  stopped  2        1 day ago
```

#### Check Cluster Status

```bash
# Check specific cluster
openframe cluster status my-cluster

# Check current context cluster
openframe cluster status
```

**Example Output:**
```
Cluster: my-cluster
Status: running
Nodes: 2/2 ready
Version: v1.28.2+k3s1
Registry: localhost:5000
```

#### Manage Cluster Lifecycle

```bash
# Start a stopped cluster
openframe cluster start my-cluster

# Stop a running cluster
openframe cluster stop my-cluster

# Delete a cluster
openframe cluster delete my-cluster

# Clean up cluster resources
openframe cluster cleanup
```

### 2. Chart Management

#### Install Helm Charts

```bash
# Install with default configuration
openframe chart install

# Install specific chart
openframe chart install --chart prometheus

# Install with custom values
openframe chart install \
  --chart grafana \
  --values ./custom-values.yaml \
  --namespace monitoring
```

#### Bootstrap OpenFrame

```bash
# Bootstrap with OSS tenant mode
openframe bootstrap --deployment-mode=oss-tenant

# Bootstrap with custom configuration
openframe bootstrap \
  --deployment-mode=enterprise \
  --namespace openframe-system \
  --timeout 10m
```

**Example Output:**
```
✓ Installing ArgoCD
✓ Configuring OpenFrame applications
✓ Waiting for deployments to be ready
✓ OpenFrame bootstrap completed successfully
```

### 3. Development Tools

#### Scaffold Development Environment

```bash
# Run Skaffold for current directory
openframe dev scaffold

# Scaffold specific service
openframe dev scaffold --service my-api

# Scaffold with custom configuration
openframe dev scaffold \
  --config ./skaffold-dev.yaml \
  --namespace development
```

#### Service Intercept with Telepresence

```bash
# Intercept service traffic
openframe dev intercept my-service

# Intercept with port mapping
openframe dev intercept my-service \
  --port 8080:80 \
  --namespace production

# List active intercepts
openframe dev intercept --list
```

## Authentication

The OpenFrame CLI uses your current Kubernetes context for authentication. Ensure you have proper cluster access:

```bash
# Check current context
kubectl config current-context

# Switch context if needed
kubectl config use-context my-cluster

# Verify access
kubectl get nodes
```

## Common Use Cases

### 1. Local Development Setup

```bash
# Complete local development setup
openframe cluster create --name dev-local
openframe bootstrap --deployment-mode=oss-tenant
openframe chart install --chart prometheus --chart grafana

# Verify everything is running
openframe cluster status
kubectl get pods --all-namespaces
```

### 2. Multi-Environment Management

```bash
# Create environments
openframe cluster create --name staging
openframe cluster create --name production

# Switch between environments
kubectl config use-context k3d-staging
openframe bootstrap --deployment-mode=oss-tenant

kubectl config use-context k3d-production
openframe bootstrap --deployment-mode=enterprise
```

### 3. Service Development Workflow

```bash
# Start development
openframe cluster create --name service-dev
openframe bootstrap --deployment-mode=oss-tenant

# Begin development with hot reload
openframe dev scaffold --service user-api

# In another terminal, intercept traffic for testing
openframe dev intercept user-api --port 3000:8080
```

### 4. Cluster Cleanup and Maintenance

```bash
# Regular cleanup routine
openframe cluster cleanup

# Full cluster reset
openframe cluster delete --all
openframe cluster create --name fresh-start
openframe bootstrap --deployment-mode=oss-tenant
```

## Error Handling Patterns

### Common Error Scenarios

#### Cluster Creation Failures

```bash
# If cluster creation fails
openframe cluster delete failed-cluster  # Clean up
openframe cluster create --name new-cluster

# Check Docker daemon
docker ps  # Ensure Docker is running
```

#### Bootstrap Failures

```bash
# Check cluster readiness first
openframe cluster status

# Bootstrap with verbose output
openframe bootstrap --deployment-mode=oss-tenant --verbose

# Check ArgoCD status
kubectl get pods -n argocd
```

#### Development Tool Issues

```bash
# Verify Skaffold installation
skaffold version

# Check Telepresence connectivity
telepresence status

# Reset development environment
openframe cluster cleanup
openframe dev scaffold --reset
```

### Error Response Format

```bash
# Typical error output format
Error: failed to create cluster 'my-cluster'
Cause: port 5000 already in use
Solution: use --registry-port flag with different port

# Example fix
openframe cluster create --name my-cluster --registry-port 5001
```

## Best Practices

### 1. Cluster Naming

```bash
# Use descriptive names
openframe cluster create --name project-dev
openframe cluster create --name project-staging
openframe cluster create --name project-prod

# Avoid generic names
# ❌ openframe cluster create --name test
# ✅ openframe cluster create --name user-service-test
```

### 2. Resource Management

```bash
# Regular cleanup
openframe cluster cleanup  # Run weekly

# Monitor resource usage
openframe cluster status --resources

# Delete unused clusters
openframe cluster list
openframe cluster delete old-cluster
```

### 3. Development Workflow

```bash
# Always verify cluster before development
openframe cluster status

# Use consistent bootstrap configuration
openframe bootstrap --deployment-mode=oss-tenant

# Separate clusters for different services
openframe cluster create --name api-dev
openframe cluster create --name frontend-dev
```

### 4. Configuration Management

```bash
# Store configurations in version control
openframe chart install --values ./charts/values-dev.yaml

# Use environment-specific configurations
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --config ./configs/development.yaml
```

### 5. Monitoring and Debugging

```bash
# Enable verbose output for debugging
openframe cluster create --name debug-cluster --verbose

# Check logs for troubleshooting
kubectl logs -n argocd deployment/argocd-server

# Use status commands regularly
openframe cluster status
kubectl get pods --all-namespaces
```

## Advanced Usage

### Scripting with OpenFrame CLI

```bash
#!/bin/bash
# setup-dev-environment.sh

set -e

echo "Setting up development environment..."

# Create cluster
openframe cluster create --name dev-env

# Wait for cluster ready
while [[ $(openframe cluster status dev-env --output json | jq -r '.status') != "running" ]]; do
  echo "Waiting for cluster..."
  sleep 5
done

# Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# Install monitoring
openframe chart install --chart prometheus --chart grafana

echo "Development environment ready!"
```

### Integration with CI/CD

```yaml
# .github/workflows/test.yml
- name: Setup OpenFrame cluster
  run: |
    openframe cluster create --name ci-test
    openframe bootstrap --deployment-mode=oss-tenant
    
- name: Run tests
  run: |
    openframe dev scaffold --service api --test-mode
    
- name: Cleanup
  run: |
    openframe cluster delete ci-test
```

This documentation provides comprehensive examples for all OpenFrame CLI operations, enabling developers to quickly implement cluster management and development workflows.