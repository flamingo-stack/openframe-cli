# OpenFrame CLI: Common Use Cases & Practical Examples

This guide covers the most common scenarios for using OpenFrame CLI, with step-by-step instructions and real-world examples.

## 1. Setting Up a Development Environment

**Scenario**: You're a developer who needs a local Kubernetes environment for testing applications.

### Quick Setup

```bash
# One-command development setup
openframe bootstrap my-dev-env --deployment-mode=oss-tenant --non-interactive
```

### Step-by-Step Setup

```bash
# Step 1: Create development cluster
openframe cluster create dev-cluster

# Step 2: Install development tools
openframe chart install --deployment-mode=oss-tenant

# Step 3: Verify everything is running
kubectl get pods -A
```

### What You Get

- K3d cluster with development-friendly settings
- ArgoCD for GitOps deployments
- Basic monitoring and logging
- Development tools integration

### Development Workflow

```bash
# Set up traffic interception for local development
openframe dev intercept my-app --port 3000

# Deploy with live reloading
openframe dev skaffold my-app

# Access your local app at localhost:3000
```

---

## 2. Creating a Demo Environment

**Scenario**: You need to quickly demo OpenFrame capabilities to stakeholders or clients.

### Demo Setup

```bash
# Create a clean demo environment
openframe bootstrap demo --deployment-mode=oss-tenant

# Wait for all components to be ready
kubectl wait --for=condition=Ready pods --all -A --timeout=300s
```

### Demo Checklist

- [ ] ArgoCD dashboard accessible
- [ ] Sample applications deployed
- [ ] Monitoring dashboards available
- [ ] GitOps workflows functioning

### Demo Commands

```bash
# Show cluster status
openframe cluster status demo

# Display running applications
kubectl get applications -n argocd

# Port forward to access services locally
kubectl port-forward -n argocd svc/argocd-server 8080:443
```

---

## 3. Multi-Environment Management

**Scenario**: You manage multiple environments (dev, staging, prod) and need isolated clusters.

### Environment Strategy

| Environment | Cluster Name | Mode | Purpose |
|-------------|--------------|------|---------|
| Development | `dev-cluster` | oss-tenant | Feature development |
| Staging | `staging-cluster` | oss-tenant | Pre-production testing |
| Production | `prod-cluster` | enterprise | Live applications |

### Setup Multiple Environments

```bash
# Development environment
openframe bootstrap dev-cluster --deployment-mode=oss-tenant

# Staging environment  
openframe bootstrap staging-cluster --deployment-mode=oss-tenant

# Production environment (enterprise mode)
openframe bootstrap prod-cluster --deployment-mode=enterprise
```

### Environment Management

```bash
# List all clusters
openframe cluster list

# Switch between environments
kubectl config use-context k3d-dev-cluster
kubectl config use-context k3d-staging-cluster

# Check status of all environments
for cluster in dev-cluster staging-cluster prod-cluster; do
  echo "=== $cluster ==="
  openframe cluster status $cluster
done
```

---

## 4. Application Development with Traffic Interception

**Scenario**: You're developing a microservice that needs to interact with other services in the cluster.

### Setup Service Interception

```bash
# Start your development cluster
openframe bootstrap dev-cluster

# List available services
kubectl get services -A

# Intercept traffic to your service
openframe dev intercept user-service --port 8080 --namespace default
```

### Development Workflow

1. **Run your service locally** on port 8080
2. **All cluster traffic** to `user-service` routes to your local instance  
3. **Test integrations** with other cluster services
4. **Debug and develop** with full cluster context

```bash
# Example: Run your local service
cd ~/my-microservice
npm start  # Runs on localhost:8080

# Traffic from the cluster now routes to your local service
# Test with: kubectl run test --image=curlimages/curl -it --rm -- \
#   curl http://user-service.default.svc.cluster.local/api/users
```

### Stop Interception

```bash
# Stop intercepting traffic
openframe dev intercept user-service --stop
```

---

## 5. Continuous Integration Setup

**Scenario**: You need OpenFrame CLI to work in your CI/CD pipeline for automated testing.

### CI-Friendly Commands

```bash
# Non-interactive bootstrap for CI
openframe bootstrap ci-test-cluster \
  --deployment-mode=oss-tenant \
  --non-interactive \
  --verbose

# Wait for cluster to be ready
while ! kubectl get nodes | grep -q Ready; do
  echo "Waiting for cluster..."
  sleep 10
done

# Run your tests
kubectl apply -f test-manifests/
kubectl wait --for=condition=Ready pods -l app=test-app --timeout=300s

# Cleanup after tests
openframe cluster delete ci-test-cluster --force
```

### GitHub Actions Example

```yaml
name: OpenFrame CI Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install OpenFrame CLI
        run: |
          curl -L -o openframe https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-linux-amd64
          chmod +x openframe
          sudo mv openframe /usr/local/bin/
      
      - name: Setup test cluster
        run: |
          openframe bootstrap test-cluster \
            --deployment-mode=oss-tenant \
            --non-interactive \
            --verbose
      
      - name: Run tests
        run: |
          kubectl apply -f k8s/test/
          kubectl wait --for=condition=Ready pods -l app=my-app
          
      - name: Cleanup
        run: openframe cluster delete test-cluster --force
```

---

## 6. Learning Kubernetes and GitOps

**Scenario**: You're learning Kubernetes and want to experiment with GitOps patterns.

### Learning Setup

```bash
# Create a learning environment
openframe bootstrap k8s-learning --deployment-mode=oss-tenant

# Explore what was created
kubectl get all -A
kubectl get applications -n argocd
```

### Exploration Commands

```bash
# Understanding cluster structure
kubectl get nodes
kubectl get namespaces
kubectl describe node k3d-k8s-learning-server-0

# Exploring ArgoCD
kubectl get applications -n argocd
kubectl describe application -n argocd app-of-apps

# Examining deployed workloads
kubectl get deployments -A
kubectl get services -A
kubectl get ingresses -A
```

### Hands-On Learning

1. **Access ArgoCD Dashboard**:
   ```bash
   kubectl port-forward -n argocd svc/argocd-server 8080:443
   # Open https://localhost:8080
   ```

2. **Deploy a Test Application**:
   ```bash
   kubectl create deployment nginx --image=nginx
   kubectl expose deployment nginx --port=80 --type=NodePort
   ```

3. **Practice GitOps**:
   - Fork the OpenFrame applications repository
   - Modify application configurations
   - Watch ArgoCD sync changes automatically

---

## 7. Troubleshooting and Maintenance

**Scenario**: Your OpenFrame cluster needs troubleshooting or maintenance.

### Health Checks

```bash
# Check cluster health
openframe cluster status my-cluster

# Comprehensive cluster check
kubectl get nodes
kubectl get pods -A | grep -v Running
kubectl top nodes  # Resource usage
kubectl top pods -A  # Pod resource usage
```

### Common Issues & Solutions

<details>
<summary><strong>ArgoCD Applications Not Syncing</strong></summary>

**Symptoms**: Applications show "OutOfSync" status

**Solutions**:
```bash
# Force sync all applications
kubectl get applications -n argocd -o name | \
  xargs -I {} kubectl patch {} -n argocd \
  --type merge -p '{"operation":{"sync":{"prune":true}}}'

# Check ArgoCD server logs
kubectl logs -n argocd deployment/argocd-server
```
</details>

<details>
<summary><strong>Pods Stuck in Pending State</strong></summary>

**Symptoms**: Pods won't schedule to nodes

**Solutions**:
```bash
# Check node resources
kubectl describe nodes

# Check pod events
kubectl describe pod <pod-name> -n <namespace>

# Increase cluster resources if needed
openframe cluster delete my-cluster
openframe cluster create my-cluster --nodes=3  # Multi-node
```
</details>

<details>
<summary><strong>Services Not Accessible</strong></summary>

**Symptoms**: Cannot reach services via ingress or port-forward

**Solutions**:
```bash
# Check service endpoints
kubectl get endpoints -A

# Verify ingress configuration
kubectl get ingress -A
kubectl describe ingress <ingress-name>

# Test service connectivity
kubectl run debug --image=curlimages/curl -it --rm -- \
  curl http://my-service.default.svc.cluster.local
```
</details>

### Cluster Maintenance

```bash
# Update cluster applications
kubectl get applications -n argocd
# Use ArgoCD UI to sync or refresh applications

# Restart cluster services
kubectl rollout restart deployment -n argocd argocd-server
kubectl rollout restart deployment -n kube-system coredns

# Clean up resources
kubectl delete pods --field-selector=status.phase=Failed -A
```

---

## 8. Advanced Configuration

**Scenario**: You need custom configurations for specific requirements.

### Custom Cluster Configuration

```bash
# Cluster with custom ports and multiple nodes
openframe cluster create advanced-cluster \
  --nodes=3 \
  --port 8080:80 \
  --port 8443:443

# Cluster with resource constraints
openframe cluster create constrained-cluster \
  --memory=2g \
  --cpus=2
```

### Environment Variables

```bash
# Configure default settings
export OPENFRAME_CLUSTER_PREFIX="company"
export OPENFRAME_DEFAULT_NODES="2"
export OPENFRAME_VERBOSE="true"

openframe bootstrap  # Uses environment defaults
```

### Configuration Files

Create `.openframe.yaml` in your project directory:

```yaml
cluster:
  defaultName: "project-cluster"
  nodes: 2
  ports:
    - "8080:80"
    - "8443:443"
deployment:
  mode: "oss-tenant"
  monitoring: true
  ingress: true
dev:
  interceptPort: 3000
  skaffoldProfile: "dev"
```

---

## Best Practices & Tips

### 🎯 Performance Tips

- **Resource Allocation**: Ensure Docker has at least 4GB RAM allocated
- **Multiple Nodes**: Use multi-node clusters for testing distributed scenarios
- **Cleanup**: Regularly clean up unused clusters with `openframe cluster cleanup`

### 🔐 Security Considerations

- Use different clusters for different security contexts
- Regularly update OpenFrame CLI to get security patches
- Don't expose production credentials in development clusters

### 📊 Monitoring & Observability

```bash
# Enable verbose logging for troubleshooting
openframe --verbose bootstrap my-cluster

# Monitor resource usage
kubectl top nodes
kubectl top pods -A

# Check application sync status
kubectl get applications -n argocd -w
```

### 🚀 Automation Tips

- Use `--non-interactive` flag for scripting
- Leverage environment variables for consistent configurations
- Implement proper cleanup in CI/CD pipelines

---

## Quick Reference Commands

| Task | Command |
|------|---------|
| **Quick start** | `openframe bootstrap` |
| **List clusters** | `openframe cluster list` |
| **Cluster status** | `openframe cluster status <name>` |
| **Delete cluster** | `openframe cluster delete <name>` |
| **Install charts** | `openframe chart install` |
| **Start intercept** | `openframe dev intercept <service> --port <port>` |
| **Stop intercept** | `openframe dev intercept <service> --stop` |
| **Help** | `openframe --help` |
| **Verbose output** | `openframe --verbose <command>` |

## What's Next?

- **Advanced Development**: Learn about custom ArgoCD applications
- **Production Deployment**: Explore enterprise deployment modes  
- **Integration**: Connect with your CI/CD pipelines
- **Monitoring**: Set up observability and alerting

> **💡 Pro Tip**: Use `openframe cluster list` regularly to keep track of your environments and clean up unused clusters to save resources.