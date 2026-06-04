# First Steps

You've successfully bootstrapped your first OpenFrame environment. Here are the first 5 things to do to get comfortable with your new cluster and start using OpenFrame effectively.

---

## 1. Check Your Cluster Status

After bootstrap, verify everything is healthy:

```bash
openframe cluster status my-openframe
```

Use the `--detailed` flag for full node and application information:

```bash
openframe cluster status my-openframe --detailed
```

This will display:

- Cluster health and node count
- Running Kubernetes nodes
- ArgoCD application sync/health status for all deployed apps

To see a quick list of all managed clusters:

```bash
openframe cluster list
```

---

## 2. Explore the Available Commands

OpenFrame CLI is organized into four command groups. Get familiar with each:

```bash
# Cluster lifecycle management
openframe cluster --help

# Helm chart and ArgoCD operations
openframe chart --help

# Developer workflow tools
openframe dev --help

# Full environment bootstrap
openframe bootstrap --help
```

### Full Command Reference

| Command | What It Does |
|---|---|
| `openframe bootstrap [name]` | Bootstrap a complete environment (cluster + charts) |
| `openframe cluster create [name]` | Create a new K3D cluster |
| `openframe cluster delete [name]` | Delete a cluster |
| `openframe cluster list` | List all managed clusters |
| `openframe cluster status [name]` | Show cluster health and ArgoCD app status |
| `openframe cluster cleanup [name]` | Remove unused Docker images and resources |
| `openframe chart install [name]` | Install/reinstall ArgoCD + app-of-apps on a cluster |
| `openframe dev intercept [service]` | Start a Telepresence intercept for local development |
| `openframe dev skaffold [cluster]` | Run a Skaffold hot-reload workflow |

---

## 3. Configure Your Local Development Workflow

If you're a developer working on OpenFrame services, set up a service intercept for local development. This routes live Kubernetes traffic from a running service to your local machine.

### Start an Intercept

```bash
# Interactive mode — select service and port via wizard
openframe dev intercept

# Or specify directly
openframe dev intercept my-api-service --port 8080 --namespace development
```

The intercept command will:

1. Validate your kubectl context and cluster connectivity
2. Connect Telepresence to the cluster
3. Route traffic from the specified Kubernetes service to your local port

### Stop an Intercept

Intercepts are automatically cleaned up when you exit (`Ctrl+C`). The service handles signal cleanup gracefully.

---

## 4. Run the Skaffold Hot-Reload Workflow

For rapid iteration on services, use the Skaffold workflow. This automatically rebuilds and redeploys container images as you change code:

```bash
# Run Skaffold dev with optional cluster bootstrap
openframe dev skaffold my-openframe

# Skip bootstrap if cluster is already running
openframe dev skaffold my-openframe --skip-bootstrap

# With a custom helm values file
openframe dev skaffold my-openframe --helm-values ./values-dev.yaml
```

> The Skaffold workflow handles prerequisite checking, interactive service selection, cluster management, and Helm chart installation automatically.

---

## 5. Learn the Deployment Modes

OpenFrame supports three deployment configurations, selected at bootstrap time via `--deployment-mode`:

```bash
# Self-hosted OSS (most common)
openframe bootstrap --deployment-mode=oss-tenant

# SaaS managed tenant
openframe bootstrap --deployment-mode=saas-tenant

# Shared SaaS infrastructure
openframe bootstrap --deployment-mode=saas-shared
```

| Mode | When to Use |
|---|---|
| `oss-tenant` | Standard self-hosted OpenFrame — recommended for most operators |
| `saas-tenant` | When deploying as a managed tenant on the SaaS platform |
| `saas-shared` | Shared infrastructure layer for SaaS deployments |

---

## Common Initial Configuration

### Working with Multiple Clusters

You can manage multiple isolated environments simultaneously:

```bash
# Create separate clusters for different environments
openframe cluster create dev-cluster --nodes 2
openframe cluster create staging-cluster --nodes 4

# Install charts on each independently
openframe chart install dev-cluster --deployment-mode=oss-tenant
openframe chart install staging-cluster --deployment-mode=oss-tenant

# List all clusters
openframe cluster list
```

### Reinstalling Charts

If you need to reinstall or update charts on an existing cluster without recreating it:

```bash
openframe chart install my-openframe --deployment-mode=oss-tenant --force
```

### Cleaning Up Resources

When a cluster is no longer needed:

```bash
# Clean up Docker images and unused resources (keeps cluster)
openframe cluster cleanup my-openframe

# Fully delete the cluster
openframe cluster delete my-openframe
```

---

## Verbose Mode

For troubleshooting or to see detailed logs during any operation, add the `-v` / `--verbose` flag:

```bash
openframe bootstrap my-openframe --verbose
openframe cluster create my-openframe --verbose
openframe chart install my-openframe --verbose
```

Verbose mode shows:
- Detailed ArgoCD synchronization progress
- All Helm operations and outputs
- Full prerequisite check results

---

## Where to Get Help

> **Support is handled in the OpenMSP Slack community — not GitHub Issues.**

| Resource | Link |
|---|---|
| OpenMSP Community | [openmsp.ai](https://www.openmsp.ai/) |
| Slack Invite | [Join OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) |
| OpenFrame Platform | [openframe.ai](https://openframe.ai) |
| Flamingo | [flamingo.run](https://flamingo.run) |

In Slack you can:
- Ask questions about setup and configuration
- Report bugs and unexpected behavior
- Request new features
- Connect with other MSP operators and developers using OpenFrame

---

## What's Next

After getting comfortable with the basics, explore:

- The architecture documentation to understand how the CLI components work together
- The development setup guide for contributing to OpenFrame CLI
- Security best practices for managing credentials and certificates in your environment
