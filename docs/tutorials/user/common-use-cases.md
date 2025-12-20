# OpenFrame CLI - Common Use Cases

This guide covers the most common scenarios you'll encounter when using OpenFrame CLI for Kubernetes development and deployment management.

## Table of Contents

1. [Setting Up Development Environments](#1-setting-up-development-environments)
2. [Managing Multiple Clusters](#2-managing-multiple-clusters)
3. [Deploying Applications with ArgoCD](#3-deploying-applications-with-argocd)
4. [Local Development Workflows](#4-local-development-workflows)
5. [Team Collaboration Setup](#5-team-collaboration-setup)
6. [CI/CD Integration](#6-cicd-integration)
7. [Troubleshooting Common Issues](#7-troubleshooting-common-issues)

## 1. Setting Up Development Environments

### Scenario: "I need a clean development environment for my project"

**Quick Solution:**
```bash
# One-command setup with sensible defaults
openframe bootstrap my-project-dev
```

**Step-by-Step Approach:**
```bash
# Create cluster with custom configuration
openframe cluster create my-project-dev

# Install ArgoCD and development tools
openframe chart install my-project-dev

# Verify everything is running
openframe cluster status my-project-dev
```

**Best Practices:**
- Use descriptive cluster names that include project and environment type
- Start with OSS tenant mode for individual development
- Enable verbose mode during initial setup to understand what's happening

---

## 2. Managing Multiple Clusters

### Scenario: "I'm working on multiple projects and need separate environments"

**Creating Multiple Clusters:**
```bash
# Project A - microservices development
openframe cluster create project-a-dev --nodes 3

# Project B - AI/ML workloads (more resources)
openframe cluster create project-b-dev --nodes 5

# Shared testing environment
openframe cluster create shared-test
```

**Switching Between Clusters:**
```bash
# List all your clusters
openframe cluster list

# Check specific cluster status
openframe cluster status project-a-dev

# Quick cluster overview
kubectl config get-contexts
```

**Cluster Cleanup:**
```bash
# Clean unused resources in a cluster
openframe cluster cleanup project-a-dev

# Remove entire cluster when project is done
openframe cluster delete project-a-dev
```

**Tips & Tricks:**
- Use consistent naming conventions: `{project}-{environment}`
- Regularly clean up unused clusters to save system resources
- Keep a development cluster for quick testing

---

## 3. Deploying Applications with ArgoCD

### Scenario: "I want to deploy my application using GitOps principles"

**After Bootstrap Setup:**
```bash
# Your ArgoCD is already running, access the UI
openframe cluster status my-cluster
# Note the ArgoCD URL and credentials from the output
```

**Access ArgoCD Dashboard:**
1. Open the ArgoCD URL shown in cluster status (typically https://localhost:8080)
2. Login with the provided admin credentials
3. Create new applications through the UI

**Common ArgoCD Operations:**
- **Sync Applications**: Click "Sync" in ArgoCD UI or use `argocd app sync my-app`
- **Check Status**: Monitor deployment status in the Applications dashboard
- **Rollback**: Use ArgoCD's revision history to rollback deployments

**Best Practices:**
- Keep your application manifests in Git repositories
- Use separate repositories for different environments
- Configure webhooks for automatic syncing

---

## 4. Local Development Workflows

### Scenario: "I want to test my code changes against a real Kubernetes environment"

**Traffic Interception with Telepresence:**
```bash
# Intercept traffic to your service for local development
openframe dev intercept my-service

# This redirects service traffic to your local machine
# Run your local development server on the same port
```

**Continuous Development with Skaffold:**
```bash
# Set up live reloading for your application
openframe dev scaffold my-service

# Skaffold will watch your code and redeploy automatically
```

**Development Workflow Example:**
1. Start with a running cluster: `openframe cluster status my-dev`
2. Intercept the service: `openframe dev intercept api-service`
3. Run your local development server
4. Make code changes and test against the live environment
5. Stop interception when done

**Screenshot Placeholder:**
```
[Screenshot: ArgoCD Dashboard showing deployed applications]
[Screenshot: Telepresence intercept in action]
```

---

## 5. Team Collaboration Setup

### Scenario: "My team needs a shared development environment"

**Shared Cluster Setup:**
```bash
# Create a shared cluster for team development
openframe bootstrap team-shared --deployment-mode=saas-shared
```

**Team Member Onboarding:**
```bash
# Each team member runs this to connect to the shared cluster
kubectl config use-context team-shared
```

**Best Practices for Teams:**
- Use `saas-shared` deployment mode for multi-team environments
- Establish naming conventions for namespaces and resources
- Set up RBAC policies for appropriate access control
- Create separate clusters for different stages (dev, staging, production)

**Resource Sharing Guidelines:**
| Resource Type | Sharing Strategy |
|---------------|------------------|
| **Namespaces** | One per team member or feature |
| **Secrets** | Shared at cluster level |
| **ConfigMaps** | Environment-specific |
| **Storage** | Shared with careful naming |

---

## 6. CI/CD Integration

### Scenario: "I want to integrate OpenFrame with my CI/CD pipeline"

**GitHub Actions Example:**
```yaml
name: Deploy to OpenFrame
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Setup OpenFrame
        run: |
          # Install OpenFrame CLI
          curl -sSL https://get.openframe.dev | bash
          
          # Create/update cluster
          openframe bootstrap ci-cluster \
            --deployment-mode=oss-tenant \
            --non-interactive \
            --verbose
```

**Non-Interactive Bootstrap:**
```bash
# Perfect for CI/CD - no prompts, uses defaults
openframe bootstrap \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose
```

**Jenkins Pipeline Example:**
```groovy
pipeline {
    agent any
    stages {
        stage('Setup Environment') {
            steps {
                sh '''
                    openframe cluster create ci-${BUILD_NUMBER}
                    openframe chart install ci-${BUILD_NUMBER}
                '''
            }
        }
        stage('Deploy') {
            steps {
                sh 'kubectl apply -f deployment.yaml'
            }
        }
        stage('Cleanup') {
            post {
                always {
                    sh 'openframe cluster delete ci-${BUILD_NUMBER}'
                }
            }
        }
    }
}
```

---

## 7. Troubleshooting Common Issues

### Issue: "My cluster won't start"

**Symptoms:**
- `openframe cluster create` fails
- Error messages about Docker or k3d

**Troubleshooting Steps:**
```bash
# Check prerequisites
openframe cluster create --dry-run

# Verify Docker is running
docker ps

# Check available resources
docker system df

# Try with verbose logging
openframe cluster create my-test --verbose
```

### Issue: "ArgoCD installation fails"

**Symptoms:**
- Chart install hangs or fails
- ArgoCD pods not starting

**Solutions:**
```bash
# Check cluster resources
kubectl get nodes
kubectl get pods -A

# Restart the installation
openframe cluster cleanup my-cluster
openframe chart install my-cluster --verbose

# Check ArgoCD status
kubectl get pods -n argocd
```

### Issue: "Cannot access applications"

**Symptoms:**
- URLs not responding
- Connection timeouts

**Troubleshooting:**
```bash
# Check cluster status
openframe cluster status my-cluster

# Verify services are running
kubectl get services -A

# Check port forwarding
kubectl get pods -n argocd
kubectl port-forward -n argocd svc/argocd-server 8080:443
```

### Common Error Messages

| Error | Likely Cause | Solution |
|-------|--------------|----------|
| "Docker daemon not available" | Docker not running | Start Docker Desktop |
| "Port already in use" | Conflicting services | Stop other services or use different ports |
| "Insufficient resources" | Low system resources | Increase Docker memory/CPU limits |
| "Network unreachable" | Network issues | Check firewall and network configuration |

## Quick Reference Commands

```bash
# Essential commands for daily use
openframe cluster list                    # See all clusters
openframe cluster status <name>          # Check cluster health
openframe cluster cleanup <name>         # Clean up resources
openframe cluster delete <name>          # Remove cluster completely

# Development commands
openframe dev intercept <service>        # Local development
openframe bootstrap --non-interactive    # CI/CD deployment
```

## Performance Tips

- **Resource Management**: Regularly clean up unused clusters with `openframe cluster cleanup`
- **Development Speed**: Use intercept instead of full deployments for faster iteration
- **System Resources**: Monitor Docker resource usage, especially with multiple clusters
- **Network**: Use local registries for faster image pulls during development

## Next Steps

- **Advanced Development**: Learn about [traffic interception patterns](../dev/getting-started-dev.md)
- **Production Deployment**: Explore production-ready configurations
- **Monitoring**: Set up observability tools in your clusters
- **Security**: Configure RBAC and security policies

> **Pro Tip**: Bookmark the cluster status command (`openframe cluster status`) - you'll use it often to get access URLs and check system health!