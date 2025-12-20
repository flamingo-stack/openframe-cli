# OpenFrame CLI Common Use Cases

This guide covers the most common scenarios you'll encounter when using OpenFrame CLI for Kubernetes development. Each use case includes step-by-step instructions and practical examples.

## 🎯 Top Use Cases Overview

| Use Case | Time Required | Complexity |
|----------|---------------|------------|
| [Quick Development Environment](#1-quick-development-environment) | 3-5 minutes | Beginner |
| [Managing Multiple Projects](#2-managing-multiple-projects) | 5-10 minutes | Intermediate |
| [CI/CD Pipeline Integration](#3-cicd-pipeline-integration) | 2-3 minutes | Intermediate |
| [Local Service Development](#4-local-service-development) | 10-15 minutes | Advanced |
| [Cleaning Up Resources](#5-cleaning-up-resources) | 1-2 minutes | Beginner |
| [Troubleshooting Deployments](#6-troubleshooting-deployments) | 5-10 minutes | Intermediate |

---

## 1. Quick Development Environment

**Scenario**: You need to quickly spin up an OpenFrame environment for development or testing.

### Step-by-Step Process

```bash
# One command to set up everything
openframe bootstrap
```

**What happens behind the scenes:**
1. Interactive deployment mode selection appears
2. Choose your preferred deployment type (OSS Tenant recommended for development)
3. K3d cluster gets created with sensible defaults
4. ArgoCD is installed and configured
5. OpenFrame applications are deployed via GitOps

### Verification Steps
```bash
# Check cluster health
openframe cluster status

# View all running services
kubectl get all --all-namespaces

# Access ArgoCD dashboard
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

**💡 Pro Tip**: Bookmark `https://localhost:8080` for quick ArgoCD access after port-forwarding.

---

## 2. Managing Multiple Projects

**Scenario**: You're working on multiple projects and need separate, isolated Kubernetes environments.

### Creating Project-Specific Clusters

```bash
# Project A cluster
openframe bootstrap project-a-cluster

# Project B cluster  
openframe bootstrap project-b-cluster

# List all your clusters
openframe cluster list
```

### Switching Between Projects

```bash
# Switch to Project A
kubectl config use-context k3d-project-a-cluster

# Verify you're on the right cluster
kubectl config current-context

# Switch to Project B
kubectl config use-context k3d-project-b-cluster
```

### Best Practices for Multi-Project Setup

| Practice | Command | Why |
|----------|---------|-----|
| **Descriptive names** | `openframe bootstrap frontend-app` | Easy identification |
| **Regular cleanup** | `openframe cluster cleanup` | Free up resources |
| **Context awareness** | `kubectl config current-context` | Avoid wrong deployments |

<details>
<summary>Advanced: Custom Resource Allocation</summary>

```bash
# Create cluster with specific resources
openframe cluster create large-project --nodes 3 --skip-wizard

# Configure memory/CPU limits in your project
kubectl apply -f resource-limits.yaml
```

</details>

---

## 3. CI/CD Pipeline Integration

**Scenario**: You need to create clusters automatically in CI/CD pipelines without user interaction.

### Non-Interactive Mode

```bash
# Full automation - no prompts
openframe bootstrap ci-cluster --deployment-mode=oss-tenant --non-interactive

# With verbose logging for CI logs
openframe bootstrap ci-cluster --deployment-mode=oss-tenant --non-interactive -v
```

### Sample GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test with OpenFrame
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup OpenFrame
        run: |
          # Install openframe CLI
          curl -L <download-url> -o openframe
          chmod +x openframe
          sudo mv openframe /usr/local/bin/
          
      - name: Create Test Environment
        run: |
          openframe bootstrap test-env \
            --deployment-mode=oss-tenant \
            --non-interactive \
            --verbose
            
      - name: Run Tests
        run: |
          kubectl get pods --all-namespaces
          # Your test commands here
          
      - name: Cleanup
        if: always()
        run: openframe cluster delete test-env --force
```

### Docker-Based CI Setup

```bash
# Run in Docker for consistent environments
docker run --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/workspace \
  openframe/cli:latest \
  bootstrap ci-test --non-interactive --deployment-mode=saas-shared
```

---

## 4. Local Service Development

**Scenario**: You want to develop and test a specific service locally while it runs in the Kubernetes cluster.

### Traffic Interception with Telepresence

```bash
# Intercept traffic to your service
openframe dev intercept my-service

# Your local service will now receive cluster traffic
# Start your local development server on the intercepted port
```

### Live Reloading Development

```bash
# Set up Skaffold for live code reloading
openframe dev skaffold my-service

# Code changes will automatically deploy to the cluster
```

### Typical Development Workflow

1. **Start intercept**: `openframe dev intercept api-service`
2. **Run locally**: Start your service locally (e.g., `npm start`, `go run main.go`)
3. **Test integration**: Make requests to cluster services that will route to your local instance
4. **Stop intercept**: Press `Ctrl+C` to stop traffic interception

**Real-world Example:**
```bash
# Terminal 1: Start traffic interception
openframe dev intercept user-api

# Terminal 2: Run your local development server
cd ~/projects/user-api
npm run dev  # Runs on port 3000

# Terminal 3: Test the integration
curl http://user-api.local-cluster.dev/users  # Routes to your local server
```

---

## 5. Cleaning Up Resources

**Scenario**: You need to free up disk space and system resources by removing unused clusters and containers.

### Regular Cleanup Commands

```bash
# Remove all unused OpenFrame clusters
openframe cluster cleanup

# Delete a specific cluster
openframe cluster delete old-project-cluster

# Force delete without confirmation
openframe cluster delete test-cluster --force

# Remove Docker containers and images (careful!)
docker system prune
```

### Automated Cleanup Script

```bash
#!/bin/bash
# cleanup-old-clusters.sh

echo "🧹 Cleaning up OpenFrame clusters older than 7 days..."

# List clusters and their creation dates
openframe cluster list --format=json | \
jq -r '.[] | select(.age > "7d") | .name' | \
while read cluster; do
  echo "Deleting old cluster: $cluster"
  openframe cluster delete "$cluster" --force
done

echo "✅ Cleanup complete!"
```

### Storage Monitoring

| What to Check | Command | Typical Size |
|---------------|---------|--------------|
| **Docker images** | `docker images` | 1-5GB per cluster |
| **Container volumes** | `docker volume ls` | 500MB-2GB |
| **Kubernetes data** | `du -sh ~/.kube` | 10-100MB |

---

## 6. Troubleshooting Deployments

**Scenario**: Something isn't working correctly and you need to diagnose the issue.

### Common Diagnostic Commands

```bash
# Check overall cluster health
openframe cluster status

# View recent events
kubectl get events --sort-by='.lastTimestamp'

# Check pod logs
kubectl logs -f deployment/my-app

# Describe resource issues
kubectl describe pod problematic-pod-name
```

### ArgoCD Sync Issues

```bash
# Check ArgoCD application status
kubectl get applications -n argocd

# Force sync an application
kubectl patch app my-app -n argocd -p='{"operation":{"sync":{"syncStrategy":{"force":true}}}}' --type=merge

# Access ArgoCD dashboard for visual debugging
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### Network Connectivity Issues

```bash
# Test internal DNS resolution
kubectl run test-pod --image=busybox --rm -it -- nslookup my-service

# Check service endpoints
kubectl get endpoints

# Test service connectivity
kubectl run curl-test --image=curlimages/curl --rm -it -- curl http://my-service:8080
```

### Resource Problems

```bash
# Check node resources
kubectl top nodes

# Check pod resource usage
kubectl top pods --all-namespaces

# View resource limits and requests
kubectl describe nodes
```

## 🔧 Tips and Tricks

### Speed Up Common Operations

| Task | Shortcut | Full Command |
|------|----------|--------------|
| **List clusters** | `openframe cluster list` | Same |
| **Quick status** | `openframe cluster status` | Same |
| **Force delete** | `openframe cluster delete NAME --force` | Same |

### Useful Aliases

Add these to your `.bashrc` or `.zshrc`:

```bash
# OpenFrame shortcuts
alias of='openframe'
alias ofc='openframe cluster'
alias ofb='openframe bootstrap'

# Kubernetes shortcuts
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
```

### Environment Variables

```bash
# Set default cluster name
export OPENFRAME_DEFAULT_CLUSTER="my-dev-cluster"

# Enable verbose logging by default
export OPENFRAME_VERBOSE=true

# Skip confirmation prompts
export OPENFRAME_AUTO_APPROVE=true
```

## 🚨 Common Pitfalls to Avoid

| Problem | Solution |
|---------|----------|
| **Running out of Docker resources** | Increase Docker memory limit to 4GB+ |
| **Port conflicts** | Stop other services using ports 80, 443, 6443 |
| **Wrong Kubernetes context** | Always check: `kubectl config current-context` |
| **Stale clusters** | Regular cleanup with `openframe cluster cleanup` |
| **Permission issues** | Ensure Docker daemon is running and accessible |

---

**Need more help?** Check out our [Developer Getting Started Guide](../dev/getting-started-dev.md) for advanced configuration and development workflows.