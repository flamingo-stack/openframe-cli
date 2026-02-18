# First Steps with OpenFrame CLI

Now that you have OpenFrame CLI installed and bootstrapped, let's explore the essential features and workflows that will accelerate your MSP development.

## 🎯 Your First 5 Actions

### 1. Explore Your Cluster

Start by understanding what was created during bootstrap:

```bash
# View cluster overview
openframe cluster status

# Detailed cluster information
openframe cluster status --detailed

# List all available clusters
openframe cluster list
```

**What You'll See:**
- Cluster name, status, and age
- Node information and resource usage
- Network configuration details
- Installed components and versions

### 2. Access the ArgoCD Dashboard

ArgoCD is your GitOps control center for managing deployments:

```bash
# Port forward ArgoCD server
kubectl port-forward -n argocd svc/argocd-server 8080:443

# Get the admin password
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d && echo
```

**Access:** https://localhost:8080
- **Username:** `admin`
- **Password:** Use the decoded secret above

**Explore the Dashboard:**
- View your application topology
- Monitor synchronization status  
- Explore GitOps repository connections
- Review deployment history

### 3. Understand Your Applications

Check what applications are deployed and their status:

```bash
# ArgoCD applications
kubectl get applications -n argocd

# All pods across namespaces
kubectl get pods --all-namespaces

# Services and ingresses
kubectl get services,ingress --all-namespaces
```

### 4. Try Development Workflows

Experience the developer-focused features:

```bash
# Interactive service intercept
openframe dev intercept

# Generate new service scaffolding
openframe dev scaffold
```

The intercept feature allows you to redirect traffic from a Kubernetes service to your local development environment, enabling rapid iteration and debugging.

### 5. Explore Chart Management

Understand how OpenFrame manages Helm charts and GitOps:

```bash
# View chart installation status
openframe chart install --dry-run

# Check ArgoCD connectivity
kubectl get applications -n argocd -o wide
```

## 🔧 Essential Configuration

### Environment Customization

Set up your preferred development environment:

```bash
# Create custom configuration directory
mkdir -p ~/.openframe/config

# Set environment variables
export OPENFRAME_CLUSTER_NAME="my-dev-cluster"
export OPENFRAME_NAMESPACE="openframe-dev"
export OPENFRAME_LOG_LEVEL="debug"
```

### Kubeconfig Setup

Ensure kubectl is properly configured:

```bash
# Verify current context
kubectl config current-context

# View all contexts
kubectl config get-contexts

# Switch between clusters if needed
kubectl config use-context k3d-openframe-local
```

## 🚀 Development Patterns

### Service Development Workflow

1. **Start with Intercepts**
   ```bash
   # List available services for intercept
   openframe dev intercept --list
   
   # Intercept a specific service
   openframe dev intercept my-service --port 8080
   ```

2. **Use Local Development**
   ```bash
   # Start local service development
   openframe dev scaffold my-new-service
   cd my-new-service
   ```

3. **Test and Iterate**
   - Make changes to your local service
   - Traffic is automatically routed through your local instance
   - Test with real cluster dependencies

### GitOps Integration

Understanding the GitOps workflow:

```bash
# View GitOps repository status
kubectl describe application app-of-apps -n argocd

# Force synchronization
kubectl patch application app-of-apps -n argocd \
  --type json -p '[{"op": "replace", "path": "/spec/syncPolicy", "value": {"automated": {"prune": true}}}]'
```

## 📊 Monitoring and Observability

### Cluster Health Checks

Regular monitoring commands:

```bash
# Node status and resources
kubectl top nodes 2>/dev/null || echo "Metrics server not available"

# Pod resource usage
kubectl top pods --all-namespaces 2>/dev/null || echo "Metrics server not available"

# Check system pods
kubectl get pods -n kube-system

# ArgoCD health
kubectl get pods -n argocd
```

### Application Status

Monitor your applications:

```bash
# Application sync status
kubectl get applications -n argocd -o custom-columns="NAME:.metadata.name,SYNC:.status.sync.status,HEALTH:.status.health.status"

# Detailed application info
kubectl describe application <app-name> -n argocd
```

## 🎮 Interactive Features

### Wizard-Driven Operations

OpenFrame CLI provides interactive wizards for complex operations:

```bash
# Interactive cluster creation
openframe cluster create

# Interactive chart installation
openframe chart install

# Interactive bootstrap with custom options
openframe bootstrap --interactive
```

### Command Discovery

Explore available commands:

```bash
# Top-level commands
openframe --help

# Cluster-specific commands  
openframe cluster --help

# Development commands
openframe dev --help

# Chart management commands
openframe chart --help
```

## 🛠️ Common Workflows

### Daily Development Routine

1. **Morning Setup**
   ```bash
   # Check cluster status
   openframe cluster status
   
   # Verify applications are healthy
   kubectl get pods --all-namespaces | grep -v Running
   ```

2. **Development Work**
   ```bash
   # Start service intercept
   openframe dev intercept my-service
   
   # Work on your local code
   # All traffic to my-service routes to your local instance
   ```

3. **Testing and Validation**
   ```bash
   # Check application logs
   kubectl logs -f deployment/my-service -n my-namespace
   
   # Monitor ArgoCD sync status
   kubectl get applications -n argocd
   ```

### Troubleshooting Workflow

When things go wrong:

```bash
# Check cluster health
openframe cluster status --detailed

# Examine pod issues
kubectl describe pod <pod-name> -n <namespace>

# View recent events
kubectl get events --sort-by='.lastTimestamp' --all-namespaces

# Restart problematic services
kubectl rollout restart deployment/<deployment-name> -n <namespace>
```

## 🎯 Next Steps and Learning Path

### Immediate Actions

1. **Explore Service Mesh**
   - Learn about Telepresence intercepts
   - Understand traffic routing and debugging

2. **GitOps Mastery**
   - Study ArgoCD application patterns
   - Learn about app-of-apps architecture

3. **Development Efficiency**
   - Master the scaffold generation
   - Understand hot-reload workflows

### Advanced Topics

After mastering the basics, explore:

- **Multi-cluster Management**: Managing multiple OpenFrame environments
- **Custom Chart Development**: Creating your own Helm charts
- **CI/CD Integration**: Automating OpenFrame in pipelines
- **Security Best Practices**: Securing your MSP platform

## 📚 Additional Resources

### Documentation Deep Dive

Explore specialized documentation:

- **[Architecture Overview](../development/architecture/README.md)** - Understanding system design
- **[Development Environment](../development/setup/environment.md)** - Advanced IDE setup
- **[Local Development](../development/setup/local-development.md)** - Source code development
- **[Security Guidelines](../development/security/README.md)** - Security best practices

### Video Learning

[![OpenFrame v0.5.2: Live Demo of AI-Powered IT Management for MSPs](https://img.youtube.com/vi/a45pzxtg27k/maxresdefault.jpg)](https://www.youtube.com/watch?v=a45pzxtg27k)

### Community and Support

- **OpenMSP Community**: https://www.openmsp.ai/
- **Slack Workspace**: Join for real-time help and discussions
- **Feature Requests**: Share ideas and vote on enhancements

## 🏆 Mastery Checklist

Track your OpenFrame CLI proficiency:

**Beginner Level:**
- [ ] Successfully bootstrap an environment
- [ ] Access ArgoCD dashboard
- [ ] Run basic cluster commands
- [ ] Understand pod and service concepts

**Intermediate Level:**
- [ ] Perform service intercepts
- [ ] Generate service scaffolds
- [ ] Navigate GitOps workflows
- [ ] Troubleshoot common issues

**Advanced Level:**
- [ ] Customize bootstrap configurations
- [ ] Manage multiple clusters
- [ ] Integrate with CI/CD pipelines
- [ ] Contribute to OpenFrame development

You're now equipped with the fundamental knowledge to leverage OpenFrame CLI effectively. Continue exploring the platform's capabilities and building amazing MSP solutions!