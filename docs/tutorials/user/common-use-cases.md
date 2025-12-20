# Common Use Cases for OpenFrame CLI

This guide covers the most common scenarios you'll encounter when using OpenFrame CLI for Kubernetes development and deployment management.

## Use Case 1: Setting Up a New Development Environment

**Scenario**: You need to create a fresh Kubernetes environment for a new project or team member.

### Step-by-Step Instructions

1. **Quick Setup with Bootstrap**
```bash
# Interactive setup (asks for preferences)
openframe bootstrap

# Or specify options directly
openframe bootstrap my-project --deployment-mode=oss-tenant
```

2. **Verify the Environment**
```bash
# Check cluster health
openframe cluster status my-project

# Verify ArgoCD is running
kubectl get applications -n argocd
```

### Best Practices
- Use descriptive cluster names that match your project
- Choose `oss-tenant` mode for most development scenarios
- Run `--verbose` flag if you need to troubleshoot setup issues

---

## Use Case 2: Managing Multiple Development Clusters

**Scenario**: You're working on multiple projects and need separate Kubernetes environments.

### How to Manage Multiple Clusters

1. **Create Project-Specific Clusters**
```bash
# Create clusters for different projects
openframe cluster create frontend-project
openframe cluster create api-project  
openframe cluster create staging-env
```

2. **List and Switch Between Clusters**
```bash
# See all your clusters
openframe cluster list

# Check specific cluster status
openframe cluster status frontend-project
```

3. **Clean Up When Done**
```bash
# Delete unused clusters
openframe cluster delete old-project

# Clean up Docker images
openframe cluster cleanup
```

### Tips and Tricks
- Use consistent naming conventions (e.g., `project-env` format)
- Regularly clean up unused clusters to save disk space
- Use `kubectl config get-contexts` to see which cluster you're currently connected to

---

## Use Case 3: Local Development with Traffic Interception

**Scenario**: You want to develop and test a microservice locally while it's integrated with a full Kubernetes environment.

### Setting Up Local Development

1. **Prepare Your Service**
```bash
# Make sure your cluster is running
openframe cluster status

# Set up traffic interception for your service
openframe dev intercept my-service
```

2. **Start Local Development**
```bash
# Your local code will now receive traffic from the cluster
# Run your service locally (example with Node.js)
npm start

# Or with Go
go run main.go

# Or with Python
python app.py
```

### What This Enables
- Test your local changes against real cluster dependencies
- Debug issues with full integration context
- Develop faster without rebuilding containers

### Troubleshooting Traffic Interception
- Ensure Telepresence is installed and working
- Check that your local service is running on the correct port
- Verify network connectivity between local machine and cluster

---

## Use Case 4: Continuous Development with Live Reloading

**Scenario**: You want automatic rebuilding and deployment when you make code changes.

### Using Skaffold Integration

1. **Set Up Skaffold Workflow**
```bash
# Configure Skaffold for your service
openframe dev skaffold my-service

# This monitors your code for changes and rebuilds/redeploys automatically
```

2. **Configure Your Project**
Create a `skaffold.yaml` in your project root:
```yaml
apiVersion: skaffold/v2beta24
kind: Config
build:
  artifacts:
  - image: my-service
    docker:
      dockerfile: Dockerfile
deploy:
  kubectl:
    manifests:
    - k8s/*.yaml
```

### Benefits
- Immediate feedback on code changes
- No manual rebuild/redeploy steps
- Faster development iteration cycles

---

## Use Case 5: Installing Additional Applications

**Scenario**: You need to add new applications or services to your existing cluster.

### Using Chart Management

1. **Install Applications via ArgoCD**
```bash
# Install charts on existing cluster
openframe chart install my-cluster

# This sets up ArgoCD if not already present
```

2. **Manage Application Deployments**
```bash
# Check ArgoCD applications
kubectl get applications -n argocd

# View application sync status
argocd app list
```

### Adding Custom Applications
- Add your application manifests to your GitOps repository
- ArgoCD will automatically sync and deploy changes
- Use ArgoCD UI for visual monitoring of deployments

---

## Use Case 6: Environment Cleanup and Maintenance

**Scenario**: Your development machine is running low on space or you need to clean up old resources.

### Regular Maintenance Commands

1. **Clean Up Unused Resources**
```bash
# Remove unused Docker images
openframe cluster cleanup

# Delete old clusters you no longer need
openframe cluster delete old-cluster-name
```

2. **Check Resource Usage**
```bash
# See all clusters and their resource usage
openframe cluster list

# Get detailed status of a specific cluster
openframe cluster status my-cluster --detailed
```

### Maintenance Schedule
> **Recommended**: Run cleanup weekly or when disk space is low

---

## Use Case 7: Troubleshooting Common Issues

**Scenario**: Something isn't working as expected and you need to diagnose the problem.

### Diagnostic Commands

1. **Check Overall System Health**
```bash
# Verify prerequisites are installed
openframe cluster create --dry-run

# Check cluster status with verbose output
openframe cluster status --verbose
```

2. **Examine Specific Issues**
```bash
# Check if pods are running
kubectl get pods --all-namespaces

# Look at recent events
kubectl get events --sort-by=.metadata.creationTimestamp

# Check ArgoCD application status
kubectl get applications -n argocd
```

### Common Problems and Solutions

| Problem | Symptoms | Solution |
|---------|----------|----------|
| **Cluster won't start** | `k3d cluster create` fails | Check Docker is running and has enough resources |
| **ArgoCD not accessible** | Can't reach ArgoCD UI | Run `kubectl port-forward svc/argocd-server -n argocd 8080:443` |
| **Services not deploying** | Pods stuck in Pending state | Check resource constraints with `kubectl describe pod <name>` |
| **Traffic interception fails** | Local service not receiving traffic | Verify Telepresence installation and service configuration |
| **Skaffold build fails** | Build errors during development | Check Dockerfile and build context |

---

## Use Case 8: Team Collaboration

**Scenario**: Your team needs consistent development environments and deployment processes.

### Setting Up Team Standards

1. **Standardized Bootstrap Process**
```bash
# Create team documentation with standardized commands
openframe bootstrap project-name --deployment-mode=oss-tenant --non-interactive

# Share this in your team's setup documentation
```

2. **Version Control Integration**
- Add `helm-values.yaml` to your repository for consistent configurations
- Use GitOps with ArgoCD for shared application deployments
- Document cluster naming conventions

### Best Practices for Teams
- Use consistent cluster naming patterns
- Share Skaffold configurations via version control
- Regularly update OpenFrame CLI across the team
- Document any custom configurations or workflows

---

## Quick Reference: Essential Commands

<details>
<summary>Expand for command cheat sheet</summary>

### Cluster Management
```bash
openframe bootstrap                          # Complete environment setup
openframe cluster create my-cluster         # Create new cluster
openframe cluster list                      # Show all clusters
openframe cluster status my-cluster         # Check cluster health
openframe cluster delete my-cluster         # Remove cluster
openframe cluster cleanup                   # Clean up resources
```

### Development Tools
```bash
openframe dev intercept my-service          # Traffic interception
openframe dev skaffold my-service           # Live reloading
```

### Chart Management
```bash
openframe chart install                     # Install ArgoCD
```

### Kubernetes Commands
```bash
kubectl get pods                           # Check pod status
kubectl get applications -n argocd         # Check ArgoCD apps
kubectl port-forward svc/argocd-server -n argocd 8080:443  # Access ArgoCD UI
```

</details>

---

## Getting More Help

- **Command Documentation**: Run any command with `--help` for detailed information
- **Verbose Mode**: Add `--verbose` to commands for detailed logging
- **Status Checks**: Use `openframe cluster status` regularly to monitor health
- **Community**: Join our community channels for tips and support

Now you're equipped to handle the most common OpenFrame CLI scenarios! Each use case builds on the previous ones, so start with the basics and gradually explore more advanced workflows as your needs grow.