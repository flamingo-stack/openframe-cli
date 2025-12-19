# Common Use Cases for OpenFrame CLI

This guide covers the most common scenarios and workflows when using OpenFrame CLI. Each use case includes step-by-step instructions, best practices, and troubleshooting tips.

## Use Case 1: Setting Up a Development Environment

**When to use**: You're a developer who needs a local Kubernetes environment for testing applications.

### Step-by-Step Guide

1. **Quick Bootstrap** (Recommended for beginners):
   ```bash
   openframe bootstrap dev-environment
   ```

2. **Manual Setup** (More control):
   ```bash
   # Create cluster
   openframe cluster create dev-environment
   
   # Install charts
   openframe chart install dev-environment
   ```

3. **Verify Setup**:
   ```bash
   openframe cluster status dev-environment
   ```

### Best Practices
- Use descriptive cluster names like `dev-environment`, `feature-testing`, or `staging`
- Keep development clusters separate from production workloads
- Regularly clean up unused clusters to save resources

### Troubleshooting
- **Issue**: Cluster won't start
- **Solution**: Check Docker memory allocation (increase to 4GB+ in Docker Desktop)

---

## Use Case 2: Testing Different Deployment Modes

**When to use**: You need to test applications in different OpenFrame configurations (OSS vs SaaS).

### Deployment Mode Options

| Mode | Description | When to Use |
|------|-------------|-------------|
| **oss-tenant** | Open Source single-tenant | Local development, testing |
| **saas-tenant** | SaaS single-tenant | Customer simulation, isolation testing |
| **saas-shared** | SaaS multi-tenant shared | Load testing, resource optimization |

### Step-by-Step Guide

1. **OSS Tenant Environment**:
   ```bash
   openframe bootstrap oss-test --deployment-mode=oss-tenant --non-interactive
   ```

2. **SaaS Tenant Environment**:
   ```bash
   openframe bootstrap saas-tenant-test --deployment-mode=saas-tenant --non-interactive
   ```

3. **Compare Configurations**:
   ```bash
   openframe cluster list
   openframe cluster status oss-test
   openframe cluster status saas-tenant-test
   ```

### Best Practices
- Use consistent naming conventions: `oss-*`, `saas-*`
- Document which mode is used for each environment
- Test deployment differences before production

---

## Use Case 3: CI/CD Pipeline Integration

**When to use**: Automating OpenFrame deployments in continuous integration workflows.

### Step-by-Step Guide

1. **Non-Interactive Bootstrap**:
   ```bash
   # In your CI script
   openframe bootstrap ci-test-$(date +%s) \
     --deployment-mode=oss-tenant \
     --non-interactive \
     --verbose
   ```

2. **Check Status Programmatically**:
   ```bash
   # Wait for cluster to be ready
   openframe cluster status ci-test-* || exit 1
   ```

3. **Cleanup After Tests**:
   ```bash
   # Clean up test cluster
   openframe cluster delete ci-test-*
   ```

### Best Practices
- Always use `--non-interactive` in CI
- Include `--verbose` for debugging
- Use unique cluster names (timestamps, build numbers)
- Always clean up resources after tests

### Example GitHub Actions Workflow
```yaml
name: Test with OpenFrame
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup OpenFrame
        run: |
          openframe bootstrap test-${{ github.run_id }} \
            --deployment-mode=oss-tenant \
            --non-interactive
      - name: Run tests
        run: |
          # Your test commands here
          kubectl get pods -A
      - name: Cleanup
        run: openframe cluster delete test-${{ github.run_id }}
```

---

## Use Case 4: Managing Multiple Environments

**When to use**: You need to work with multiple clusters for different projects or stages.

### Step-by-Step Guide

1. **Create Multiple Environments**:
   ```bash
   openframe cluster create project-a-dev
   openframe cluster create project-a-staging  
   openframe cluster create project-b-dev
   ```

2. **Install Charts on Specific Clusters**:
   ```bash
   openframe chart install project-a-dev
   openframe chart install project-a-staging
   ```

3. **Switch Between Environments**:
   ```bash
   # List all clusters
   openframe cluster list
   
   # Check specific cluster
   openframe cluster status project-a-dev
   ```

### Best Practices
- Use naming conventions: `{project}-{environment}`
- Document cluster purposes and owners
- Regular cleanup of unused clusters
- Consider resource limits when running multiple clusters

---

## Use Case 5: Development with Traffic Interception

**When to use**: You want to debug services locally while they run in the cluster.

### Step-by-Step Guide

1. **Setup Development Environment**:
   ```bash
   openframe bootstrap dev-intercept
   ```

2. **Intercept Traffic for a Service**:
   ```bash
   openframe dev intercept my-service
   ```

3. **Run Local Development Server**:
   ```bash
   # Terminal 1: Run intercepted traffic to local service
   ./my-service --port=8080
   
   # Terminal 2: Test the connection
   curl http://localhost:8080/health
   ```

### Best Practices
- Use unique service names
- Document which services are intercepted
- Clean up intercepts when done
- Monitor resource usage during development

---

## Use Case 6: Quick Environment Reset

**When to use**: Your environment is corrupted or you need a fresh start.

### Step-by-Step Guide

1. **Full Cleanup**:
   ```bash
   # Delete cluster
   openframe cluster delete my-cluster
   
   # Clean up Docker resources
   docker system prune -f
   ```

2. **Fresh Bootstrap**:
   ```bash
   openframe bootstrap my-cluster --verbose
   ```

3. **Verify Clean State**:
   ```bash
   openframe cluster list
   openframe cluster status my-cluster
   ```

### When to Reset
- Applications won't start properly
- ArgoCD is in a bad state
- Resource conflicts or port issues
- After major configuration changes

---

## Use Case 7: Resource Monitoring and Cleanup

**When to use**: Regular maintenance and resource optimization.

### Step-by-Step Guide

1. **Check All Clusters**:
   ```bash
   openframe cluster list
   ```

2. **Detailed Status Check**:
   ```bash
   for cluster in $(openframe cluster list --format=name); do
     echo "=== $cluster ==="
     openframe cluster status $cluster
   done
   ```

3. **Cleanup Unused Resources**:
   ```bash
   # Clean up specific cluster
   openframe cluster cleanup old-cluster
   
   # Or delete entirely
   openframe cluster delete old-cluster
   ```

### Maintenance Schedule
- **Daily**: Check cluster status
- **Weekly**: Clean up unused resources
- **Monthly**: Delete old test clusters

---

## Tips and Tricks

### Quick Commands Reference

```bash
# Fast bootstrap with defaults
alias of-boot="openframe bootstrap"

# Quick cluster status
alias of-status="openframe cluster status"

# List all clusters
alias of-list="openframe cluster list"

# Verbose mode for debugging
alias of-debug="openframe --verbose"
```

### Resource Management

<details>
<summary>Click to expand resource optimization tips</summary>

- **Limit concurrent clusters**: Run only what you need
- **Use cleanup commands**: Regular maintenance prevents issues
- **Monitor Docker resources**: Increase limits if needed
- **Delete unused clusters**: Free up system resources

</details>

### Advanced Configuration

<details>
<summary>Click to expand advanced setup options</summary>

```bash
# Custom configurations
export OPENFRAME_CONFIG_PATH="./custom-config"
export OPENFRAME_VERBOSE=true

# Development shortcuts
alias of="openframe"
alias of-dev="openframe dev"
alias of-intercept="openframe dev intercept"
```

</details>

## Troubleshooting Common Issues

| Problem | Command to Diagnose | Solution |
|---------|-------------------|----------|
| Cluster won't start | `openframe cluster status` | Restart Docker, check resources |
| ArgoCD not accessible | `kubectl get pods -n argocd` | Port-forward: `kubectl port-forward svc/argocd-server -n argocd 8080:443` |
| Services not deploying | `openframe cluster status --verbose` | Check ArgoCD sync status |
| Out of disk space | `docker system df` | Run `docker system prune -f` |
| Port conflicts | `netstat -tulpn \| grep 8080` | Change port or stop conflicting service |

## Getting Help

- **Interactive Help**: Run any command without arguments to see prompts
- **Command Help**: Use `--help` flag: `openframe cluster --help`
- **Verbose Output**: Add `--verbose` to see detailed execution logs
- **Status Checks**: Regular use of `status` commands helps identify issues early

> **Pro Tip**: OpenFrame CLI is designed to be forgiving. Most problems can be solved by deleting and recreating clusters, which only takes a few minutes.