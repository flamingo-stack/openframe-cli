# OpenFrame CLI - API Usage Examples

This guide provides practical examples for using the OpenFrame CLI tool to manage Kubernetes clusters and development workflows.

## Table of Contents

1. [Installation & Setup](#installation--setup)
2. [Authentication](#authentication)
3. [Main Commands](#main-commands)
4. [Common Use Cases](#common-use-cases)
5. [Error Handling](#error-handling)
6. [Best Practices](#best-practices)

## Installation & Setup

### Quick Installation

```bash
# macOS (ARM64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Verify installation
openframe --version
```

### Verify Setup

```bash
# Check system requirements
openframe cluster status

# View available commands
openframe --help
```

## Authentication

OpenFrame CLI uses your local Kubernetes configuration for authentication. Ensure you have proper access to your target clusters.

### Configure Kubernetes Context

```bash
# View current context
kubectl config current-context

# Switch context if needed
kubectl config use-context your-cluster-context

# Verify access
kubectl get nodes
```

## Main Commands

### Core CLI Functions

| Command Category | Function | Description |
|------------------|----------|-------------|
| Cluster Management | `cluster create/list/status/delete` | K3d cluster lifecycle |
| Chart Management | `chart install` | Helm chart installation |
| Bootstrap | `bootstrap` | Full OpenFrame setup |
| Development | `dev scaffold/intercept` | Development workflows |

## Common Use Cases

### 1. Complete Cluster Setup

**Create and bootstrap a new development cluster:**

```bash
# Step 1: Create cluster with interactive wizard
openframe cluster create

# Step 2: Wait for cluster to be ready
openframe cluster status

# Step 3: Bootstrap OpenFrame
openframe bootstrap --deployment-mode=oss-tenant

# Step 4: Verify installation
kubectl get pods -n openframe-system
```

### 2. Development Workflow

**Set up local development environment:**

```bash
# Start development with Skaffold
openframe dev scaffold --config=skaffold.yaml

# In another terminal, intercept service traffic
openframe dev intercept my-service --port=8080:80

# View intercepted traffic
curl http://localhost:8080/api/health
```

### 3. Cluster Management

**Manage multiple clusters:**

```bash
# List all clusters
openframe cluster list

# Get detailed cluster information
openframe cluster status --cluster-name=my-cluster

# Start/stop clusters as needed
openframe cluster start my-cluster
openframe cluster delete my-cluster --force
```

### 4. Chart Installation

**Install and manage Helm charts:**

```bash
# Install charts with ArgoCD
openframe chart install --chart-path=./charts --namespace=my-app

# Install specific chart
openframe chart install \
  --chart-name=my-service \
  --chart-version=1.2.3 \
  --values=values.prod.yaml
```

### 5. Production Bootstrap

**Bootstrap production-ready OpenFrame:**

```bash
# Bootstrap with production configuration
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --config-file=production.yaml \
  --timeout=30m

# Verify all components
openframe cluster status --verbose
```

## Error Handling

### Common Error Patterns

#### 1. Cluster Creation Failures

```bash
# Check Docker daemon
docker info || echo "Docker not running"

# Verify system resources
openframe cluster status --check-requirements

# Clean up failed cluster
openframe cluster cleanup --force
```

#### 2. Bootstrap Issues

```bash
# Check cluster readiness before bootstrap
kubectl get nodes --no-headers | grep -q "Ready" || {
  echo "Cluster not ready"
  exit 1
}

# Bootstrap with error recovery
openframe bootstrap --deployment-mode=oss-tenant --retry=3 || {
  echo "Bootstrap failed, checking logs..."
  kubectl logs -n openframe-system -l app=openframe
}
```

#### 3. Development Setup Problems

```bash
# Verify Skaffold configuration
skaffold diagnose --filename=skaffold.yaml

# Check Telepresence connectivity
openframe dev intercept --check-connection my-service || {
  echo "Telepresence connection failed"
  telepresence status
}
```

### Error Handling Script Template

```bash
#!/bin/bash
set -euo pipefail

# Function for error handling
handle_error() {
  echo "Error on line $1"
  openframe cluster status --verbose
  exit 1
}

trap 'handle_error ${LINENO}' ERR

# Your OpenFrame commands here
openframe cluster create
openframe bootstrap --deployment-mode=oss-tenant
```

## Best Practices

### 1. Cluster Management

**Use descriptive cluster names:**

```bash
# Good: descriptive names
openframe cluster create --name=feature-auth-service
openframe cluster create --name=staging-v2-testing

# Avoid: generic names
openframe cluster create --name=test123
```

**Regular cleanup:**

```bash
# Weekly cleanup script
#!/bin/bash
echo "Cleaning up old clusters..."
openframe cluster list --format=json | \
  jq -r '.[] | select(.age > "7d") | .name' | \
  xargs -I {} openframe cluster delete {}
```

### 2. Development Workflow

**Use configuration files:**

```bash
# Create reusable configuration
cat > openframe-config.yaml << EOF
cluster:
  name: my-project-dev
  nodes: 3
  ports:
    - "80:80"
    - "443:443"
bootstrap:
  deployment-mode: oss-tenant
  timeout: 20m
EOF

# Use configuration
openframe cluster create --config=openframe-config.yaml
```

**Environment-specific commands:**

```bash
# Development
openframe bootstrap --deployment-mode=oss-tenant --dev

# Staging
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --config-file=staging.yaml \
  --namespace=staging

# Production (external cluster)
openframe chart install \
  --chart-path=./production-charts \
  --values=values.prod.yaml \
  --dry-run  # Always dry-run first
```

### 3. Automation & CI/CD

**GitHub Actions integration:**

```yaml
# .github/workflows/openframe.yml
name: OpenFrame Development
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install OpenFrame CLI
        run: |
          curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
          sudo mv openframe /usr/local/bin/
      
      - name: Create test cluster
        run: |
          openframe cluster create --name=ci-test-${{ github.run_id }}
          
      - name: Bootstrap OpenFrame
        run: |
          openframe bootstrap --deployment-mode=oss-tenant --timeout=15m
          
      - name: Run tests
        run: |
          openframe dev scaffold --config=ci-skaffold.yaml
          
      - name: Cleanup
        if: always()
        run: |
          openframe cluster delete ci-test-${{ github.run_id }} --force
```

### 4. Monitoring & Debugging

**Health check script:**

```bash
#!/bin/bash
# openframe-health-check.sh

echo "=== OpenFrame Health Check ==="

# Check cluster status
echo "Cluster Status:"
openframe cluster status || exit 1

# Check key components
echo "Checking OpenFrame components..."
kubectl get pods -n openframe-system

# Check ArgoCD
echo "ArgoCD Status:"
kubectl get applications -n argocd

# Resource usage
echo "Resource Usage:"
kubectl top nodes
kubectl top pods -n openframe-system
```

### 5. Security Best Practices

**Use least-privilege principles:**

```bash
# Create service account for automation
kubectl create serviceaccount openframe-automation
kubectl create rolebinding openframe-automation \
  --clusterrole=view \
  --serviceaccount=default:openframe-automation

# Use specific namespaces
openframe chart install \
  --chart-name=my-app \
  --namespace=production \
  --create-namespace
```

**Secure configuration management:**

```bash
# Use external secret management
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --secrets-from=vault://my-vault/openframe-secrets
```

This documentation provides practical, copy-paste ready examples for all major OpenFrame CLI operations. For additional help, use `openframe --help` or visit the [official documentation](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs).