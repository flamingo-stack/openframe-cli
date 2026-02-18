# First Steps with OpenFrame CLI

Congratulations! You've successfully installed OpenFrame CLI and bootstrapped your first cluster. This guide will walk you through the essential next steps to get the most out of your OpenFrame environment.

## The 5 Essential First Steps

### 1. Explore Your Cluster

Start by understanding what was created during bootstrap:

```bash
# Check overall cluster health
./openframe cluster status

# List all available clusters
./openframe cluster list

# View detailed cluster information
./openframe cluster status --verbose
```

**What you'll see:**
- Cluster node status and resource usage
- ArgoCD deployment health
- Application sync status
- Network configuration details

### 2. Navigate the ArgoCD Interface

ArgoCD is your GitOps control center. Access it and explore:

```bash
# Get ArgoCD credentials (if you missed them)
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

**Navigate to:** `https://localhost:8080`

**Key areas to explore:**
- **Applications**: View all deployed services and their health
- **Repositories**: See connected Git repositories
- **Settings**: Review cluster configuration
- **Logs**: Monitor deployment activities

### 3. Understand the Application Architecture

OpenFrame follows the app-of-apps pattern. View your applications:

```bash
# List ArgoCD applications
kubectl get applications -n argocd

# Describe a specific application
kubectl describe application api-service -n argocd

# View application pods across all namespaces
kubectl get pods --all-namespaces
```

**Default Applications Include:**
- **API Service**: Core backend API
- **Gateway Service**: API gateway and routing
- **UI Service**: Frontend application
- **Monitoring**: Observability stack
- **Ingress**: Traffic management

### 4. Test Local Development Workflow

Set up your first development intercept:

```bash
# List available services for intercept
./openframe dev intercept --list

# Create an intercept for the API service
./openframe dev intercept api-service

# This will:
# • Set up Telepresence intercept
# • Route traffic to your local development server
# • Provide instructions for local development
```

**Benefits of intercepts:**
- Test changes against the full cluster environment
- Debug services with real dependencies
- Faster development iteration cycles

### 5. Practice Cluster Management

Learn the essential cluster operations:

```bash
# Create a new cluster for experimentation
./openframe cluster create dev-cluster --dry-run

# Actually create it
./openframe cluster create dev-cluster

# Switch between clusters
kubectl config get-contexts
kubectl config use-context k3d-dev-cluster

# Clean up when done
./openframe cluster delete dev-cluster
```

## Common Configuration Tasks

### Configure Git Integration

If you plan to use private repositories:

```bash
# Configure Git credentials for ArgoCD
kubectl create secret generic git-credentials \
  --from-literal=username=your-username \
  --from-literal=password=your-token \
  -n argocd
```

### Set Up Container Registry Access

For private container images:

```bash
# Create registry secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=your-username \
  --docker-password=your-token \
  --docker-email=your-email
```

### Customize Helm Values

Modify application configurations:

```bash
# Launch the chart installation wizard
./openframe chart install

# Follow the interactive prompts to:
# • Select deployment mode
# • Configure service parameters
# • Set resource limits
# • Enable monitoring features
```

## Essential Commands Reference

### Cluster Operations

```bash
# Cluster lifecycle
./openframe cluster create <name>     # Create new cluster
./openframe cluster list              # List all clusters  
./openframe cluster status <name>     # Check cluster health
./openframe cluster delete <name>     # Remove cluster

# Cluster maintenance
./openframe cluster cleanup           # Clean up resources
```

### Chart Management

```bash
# Chart operations
./openframe chart install             # Interactive installation
./openframe chart install --mode=ghcr # GHCR deployment mode
./openframe chart install --mode=local # Local development mode

# Chart status
kubectl get applications -n argocd    # View ArgoCD apps
kubectl get pods --all-namespaces     # View all pods
```

### Development Tools

```bash
# Service intercepts
./openframe dev intercept <service>   # Create intercept
./openframe dev intercept --list      # List available services
./openframe dev intercept --cleanup   # Remove all intercepts

# Service scaffolding
./openframe dev scaffold <service>    # Create new service template
```

### Bootstrap Operations

```bash
# Complete environment setup
./openframe bootstrap                 # Full bootstrap
./openframe bootstrap --cluster-only # Just cluster creation
./openframe bootstrap --charts-only  # Just chart installation
```

## Exploring Advanced Features

### Multi-Cluster Management

```bash
# Create multiple environments
./openframe cluster create staging
./openframe cluster create production

# Deploy to different clusters
kubectl config use-context k3d-staging
./openframe chart install --mode=staging

kubectl config use-context k3d-production  
./openframe chart install --mode=production
```

### Service Mesh Integration

If using service mesh features:

```bash
# Check service mesh status
kubectl get pods -n istio-system

# View service mesh configuration
kubectl get virtualservices --all-namespaces
kubectl get destinationrules --all-namespaces
```

### Monitoring and Observability

Access monitoring dashboards:

```bash
# Check monitoring stack
kubectl get pods -n monitoring

# Port forward to Grafana (if installed)
kubectl port-forward svc/grafana 3000:3000 -n monitoring

# Port forward to Prometheus (if installed)
kubectl port-forward svc/prometheus 9090:9090 -n monitoring
```

## Troubleshooting Your Environment

### Application Health Checks

```bash
# Check application status
kubectl get applications -n argocd

# View application logs
kubectl logs -f deployment/api-service -n default

# Debug failing pods
kubectl describe pod <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace>
```

### Network Connectivity

```bash
# Test service connectivity
kubectl exec -it <pod-name> -- curl http://api-service:8080/health

# Check ingress configuration  
kubectl get ingress --all-namespaces
kubectl describe ingress <ingress-name>
```

### Resource Issues

```bash
# Check resource usage
kubectl top nodes
kubectl top pods --all-namespaces

# View resource limits
kubectl describe limitrange --all-namespaces
kubectl describe resourcequota --all-namespaces
```

## Development Best Practices

### Local Development Setup

1. **Use Intercepts**: Always use `./openframe dev intercept` for local development
2. **Monitor Logs**: Keep application logs open while developing
3. **Test Incrementally**: Make small changes and test frequently
4. **Clean Up**: Remove intercepts when switching contexts

### Cluster Management

1. **Use Descriptive Names**: Name clusters by purpose (dev, staging, feature-xyz)
2. **Regular Cleanup**: Delete unused clusters to save resources
3. **Monitor Health**: Check cluster status regularly
4. **Backup Configurations**: Export important kubectl configs

### GitOps Workflow

1. **Small Commits**: Make incremental changes to application configs
2. **Monitor Sync**: Watch ArgoCD for sync status after changes
3. **Rollback Ready**: Know how to revert changes quickly
4. **Document Changes**: Comment on configuration modifications

## Where to Go Next

Based on your role and interests:

### For Developers
- Learn about service intercepts and local development workflows
- Explore the application architecture and API patterns
- Practice debugging and monitoring techniques

### For Platform Engineers  
- Study the chart management and deployment patterns
- Learn multi-cluster management strategies
- Explore security and monitoring configurations

### For DevOps Teams
- Master the bootstrap and cluster lifecycle operations
- Learn GitOps best practices with ArgoCD
- Explore CI/CD integration patterns

## Getting Help

When you need assistance:

1. **Built-in Help**: Use `./openframe <command> --help`
2. **Verbose Output**: Add `--verbose` to commands for detailed logging
3. **Community Support**: Join [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
4. **Documentation**: Refer to the architecture and development guides

## Summary

You've now completed the essential first steps with OpenFrame CLI:

✅ **Explored** your cluster and applications  
✅ **Navigated** the ArgoCD interface  
✅ **Understood** the application architecture  
✅ **Tested** local development workflows  
✅ **Practiced** cluster management operations  

The OpenFrame CLI is designed to grow with you. As you become more comfortable with these basics, explore the advanced features and development workflows that make OpenFrame a powerful platform for MSP operations.

Remember: The CLI includes interactive wizards and comprehensive help for every operation. Don't hesitate to experiment - clusters can be created and destroyed quickly, making it safe to try new approaches and configurations.