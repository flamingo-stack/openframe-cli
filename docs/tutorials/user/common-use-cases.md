# Common Use Cases for OpenFrame CLI

This guide covers the most common scenarios you'll encounter when using OpenFrame CLI for local Kubernetes development and GitOps workflows.

## Use Case 1: Setting Up a Development Environment

**Scenario**: You're a developer who needs a local Kubernetes environment for testing applications.

### Steps
1. **Quick Environment Setup**
   ```bash
   # One-command setup for development
   openframe bootstrap dev-environment --deployment-mode=oss-tenant
   ```

2. **Verify Your Environment**
   ```bash
   # Check cluster status
   openframe cluster status dev-environment
   
   # View all running services
   kubectl get pods --all-namespaces
   ```

3. **Access Development Tools**
   - Open ArgoCD: `https://argocd.local`
   - View cluster dashboard: `https://openframe.local`
   - Use kubectl: `kubectl cluster-info`

**Best Practices**:
- Use descriptive cluster names for different projects
- Set deployment mode based on your team size (oss-tenant for solo work)
- Enable verbose mode during initial setup to understand the process

---

## Use Case 2: Managing Multiple Project Environments

**Scenario**: You work on multiple projects and need isolated environments for each.

### Steps
1. **Create Project-Specific Clusters**
   ```bash
   # Frontend project
   openframe cluster create frontend-app
   openframe chart install frontend-app --deployment-mode=oss-tenant
   
   # Backend API project  
   openframe cluster create api-backend
   openframe chart install api-backend --deployment-mode=oss-tenant
   
   # Microservices project
   openframe cluster create microservices
   openframe chart install microservices --deployment-mode=saas-tenant
   ```

2. **List and Switch Between Environments**
   ```bash
   # View all clusters
   openframe cluster list
   
   # Check specific cluster status
   openframe cluster status frontend-app
   ```

3. **Switch kubectl Context**
   ```bash
   # Switch to specific cluster
   kubectl config use-context k3d-frontend-app
   
   # Verify current context
   kubectl config current-context
   ```

**Tips & Tricks**:
- Use consistent naming conventions: `projectname-env` (e.g., `myapp-dev`, `myapp-staging`)
- Document which cluster is used for what purpose
- Consider resource usage - each cluster consumes Docker resources

---

## Use Case 3: GitOps Workflow with ArgoCD

**Scenario**: You want to deploy applications using GitOps principles through ArgoCD.

### Steps
1. **Prepare Your Git Repository**
   ```bash
   # Your repo should have Kubernetes manifests
   my-app-repo/
   ├── k8s/
   │   ├── deployment.yaml
   │   ├── service.yaml
   │   └── ingress.yaml
   └── argocd/
       └── application.yaml
   ```

2. **Install Charts with Custom Repository**
   ```bash
   # Install with custom GitHub settings
   openframe chart install my-cluster \
     --github-repo=https://github.com/myorg/my-app-configs \
     --github-branch=main
   ```

3. **Access ArgoCD and Deploy Apps**
   - Navigate to `https://argocd.local`
   - Login with admin credentials (shown during installation)
   - Create new applications pointing to your repositories

4. **Monitor Deployments**
   ```bash
   # Watch ArgoCD sync status
   kubectl get applications -n argocd
   
   # View application logs
   kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller
   ```

**Best Practices**:
- Structure your repositories with clear environment folders
- Use ArgoCD application sets for managing multiple similar applications
- Tag your releases for easier rollbacks

---

## Use Case 4: Local Development with Traffic Interception

**Scenario**: You're developing a microservice and want to test it against a real cluster environment.

### Steps
1. **Set Up Development Environment**
   ```bash
   # Create cluster with development tools
   openframe bootstrap dev-cluster --deployment-mode=saas-tenant
   ```

2. **Deploy Your Application Stack**
   - Use ArgoCD to deploy all services except the one you're developing
   - Ensure your service has proper Kubernetes service definitions

3. **Start Local Development with Intercept**
   ```bash
   # Intercept traffic to your service
   openframe dev intercept my-service
   ```

4. **Run Your Service Locally**
   ```bash
   # Start your service locally on the intercepted port
   npm start  # or python app.py, go run main.go, etc.
   ```

**How It Works**:
- Telepresence intercepts traffic destined for your service in the cluster
- Redirects that traffic to your local development environment
- Other services in the cluster continue to work normally

---

## Use Case 5: CI/CD Pipeline Integration

**Scenario**: You want to integrate OpenFrame CLI into your automated testing pipeline.

### Steps
1. **Non-Interactive Setup Script**
   ```bash
   #!/bin/bash
   # ci-setup.sh
   
   # Create test environment
   openframe bootstrap test-env-$BUILD_NUMBER \
     --deployment-mode=oss-tenant \
     --non-interactive
   
   # Wait for services to be ready
   kubectl wait --for=condition=ready pod -l app=argocd-server -n argocd --timeout=300s
   ```

2. **Deploy Test Applications**
   ```bash
   # Install test charts
   openframe chart install test-env-$BUILD_NUMBER \
     --github-repo=$TEST_REPO_URL \
     --github-branch=$BRANCH_NAME \
     --non-interactive
   ```

3. **Run Tests Against Environment**
   ```bash
   # Run your test suite against the cluster
   pytest integration_tests/ --cluster=test-env-$BUILD_NUMBER
   ```

4. **Clean Up**
   ```bash
   # Clean up after tests
   openframe cluster delete test-env-$BUILD_NUMBER
   ```

**Pipeline Benefits**:
- Each build gets isolated test environment
- Tests run against real Kubernetes services
- Automatic cleanup prevents resource leaks

---

## Use Case 6: Team Collaboration Setup

**Scenario**: Your team needs shared development environments with consistent configurations.

### Steps
1. **Create Team Configuration File**
   ```yaml
   # team-config.yaml
   deployment_mode: saas-shared
   github_repo: https://github.com/ourteam/app-configs
   github_branch: develop
   cluster_name: team-shared-dev
   ```

2. **Standardized Team Setup**
   ```bash
   # Each team member runs the same setup
   openframe bootstrap team-shared-dev \
     --deployment-mode=saas-shared \
     --github-repo=https://github.com/ourteam/app-configs \
     --github-branch=develop
   ```

3. **Share Access Information**
   ```bash
   # Export cluster configuration for sharing
   kubectl config view --flatten > team-kubeconfig.yaml
   
   # Team members can import this configuration
   export KUBECONFIG=team-kubeconfig.yaml
   ```

**Collaboration Tips**:
- Use shared Git repositories for Kubernetes manifests
- Establish naming conventions for branches and environments
- Document cluster access procedures for new team members

---

## Use Case 7: Upgrading and Maintenance

**Scenario**: You need to update OpenFrame components or clean up old environments.

### Steps
1. **Check Current Status**
   ```bash
   # List all clusters
   openframe cluster list
   
   # Check status of each
   openframe cluster status my-cluster
   ```

2. **Update Components**
   ```bash
   # Force reinstall charts with latest versions
   openframe chart install my-cluster --force
   ```

3. **Clean Up Unused Resources**
   ```bash
   # Remove specific cluster
   openframe cluster delete old-project
   
   # Clean up all unused resources
   openframe cluster cleanup
   ```

4. **Backup Important Data**
   ```bash
   # Export important configurations
   kubectl get configmaps -o yaml > backup-configs.yaml
   kubectl get secrets -o yaml > backup-secrets.yaml
   ```

---

## Troubleshooting Common Issues

<details>
<summary><strong>Cluster Creation Fails</strong></summary>

**Symptoms**: `openframe cluster create` fails with Docker errors

**Solutions**:
1. Check Docker is running: `docker ps`
2. Verify disk space: `df -h`
3. Clean up Docker: `docker system prune -f`
4. Try with verbose logging: `openframe cluster create --verbose`

</details>

<details>
<summary><strong>ArgoCD Not Accessible</strong></summary>

**Symptoms**: Cannot access `https://argocd.local`

**Solutions**:
1. Check ArgoCD pods: `kubectl get pods -n argocd`
2. Verify ingress: `kubectl get ingress -n argocd`
3. Check local DNS: Add `127.0.0.1 argocd.local` to `/etc/hosts`
4. Wait for sync: ArgoCD takes 2-3 minutes to fully start

</details>

<details>
<summary><strong>Port Conflicts</strong></summary>

**Symptoms**: "Port already in use" errors during cluster creation

**Solutions**:
1. Stop conflicting services: `sudo lsof -i :80,443`
2. Use different ports: Modify K3d configuration
3. Delete existing clusters: `openframe cluster list` and `openframe cluster delete`

</details>

## Advanced Tips

### Performance Optimization
- **Resource Limits**: Adjust Docker Desktop resources (4GB+ RAM recommended)
- **Cluster Sizing**: Use fewer nodes for development (`--nodes 1`)
- **Background Services**: Stop unused clusters to free resources

### Development Workflow
- **Hot Reloading**: Use `openframe dev skaffold` for automatic rebuilds
- **Service Mesh**: Enable Istio/Linkerd through ArgoCD applications
- **Monitoring**: Deploy Prometheus/Grafana via GitOps

### Security Best Practices
- **Secrets Management**: Use external secret operators in production
- **Network Policies**: Test network policies in development clusters
- **RBAC**: Practice role-based access control configurations

---

## Next Steps

Now that you understand common use cases:

1. **Explore Advanced Features**: Try the [Developer Getting Started](../dev/getting-started-dev.md) guide
2. **Learn Architecture**: Read the [Architecture Overview](../dev/architecture-overview-dev.md)
3. **Customize Configuration**: Modify Helm values and ArgoCD applications
4. **Join the Community**: Share your use cases and get help from other users

**Happy developing!** 🚀