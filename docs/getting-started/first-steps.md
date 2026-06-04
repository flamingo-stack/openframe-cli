# First Steps

You've bootstrapped your first OpenFrame environment — here's what to do next to explore its key capabilities.

[![OpenFrame Product Walkthrough (Beta Access)](https://img.youtube.com/vi/awc-yAnkhIo/maxresdefault.jpg)](https://www.youtube.com/watch?v=awc-yAnkhIo)

---

## Step 1 — Verify Your Cluster

Start by confirming your cluster is healthy and all applications are in sync.

```bash
# List all managed clusters
openframe cluster list

# Get detailed status of your cluster
openframe cluster status my-cluster

# Confirm all pods are running
kubectl get pods -A
```

You should see your K3D cluster in `running` state and all ArgoCD-managed pods in the `Running` or `Completed` state.

---

## Step 2 — Explore the Cluster Commands

The `cluster` command group (alias `k`) manages your K3D cluster lifecycle:

```bash
# Create a new cluster interactively
openframe cluster create

# Create a named cluster with defaults
openframe cluster create my-dev-cluster

# List all clusters
openframe cluster list

# Get status of a specific cluster
openframe cluster status my-dev-cluster

# Clean up unused resources
openframe cluster cleanup

# Delete a cluster
openframe cluster delete my-dev-cluster
```

> **Tip:** Use the short alias `openframe k` instead of `openframe cluster` to save keystrokes.

---

## Step 3 — Explore the Chart Commands

The `chart` command group (alias `c`) manages ArgoCD and the app-of-apps Helm charts independently of the full bootstrap:

```bash
# Install charts on an existing cluster interactively
openframe chart install

# Install charts with a specific deployment mode
openframe chart install --deployment-mode=oss-tenant

# Install non-interactively
openframe chart install --deployment-mode=oss-tenant --non-interactive
```

This is useful when you already have a cluster and only need to (re)install the OpenFrame application stack.

---

## Step 4 — Set Up a Local Development Intercept

The `dev intercept` command lets you route live Kubernetes traffic for a specific service to your local machine. This is the fastest way to develop and debug a service without rebuilding containers.

```bash
# Interactive intercept setup (recommended for first time)
openframe dev intercept

# The wizard will guide you through:
# 1. Select which cluster to work with
# 2. Choose a namespace
# 3. Enter the service name to intercept
# 4. Choose the Kubernetes port
# 5. Enter your local port

# Or use flags directly
openframe dev intercept my-api-service \
  --port 8080 \
  --namespace development
```

When the intercept is active, any traffic hitting `my-api-service` in the cluster is forwarded to your local port `8080`. Press `Ctrl+C` to stop and cleanly disconnect.

---

## Step 5 — Run a Skaffold Dev Session

For a full hot-reload development loop, use the `dev skaffold` command:

```bash
# Discover skaffold.yaml and start a dev session
openframe dev skaffold

# Target a specific cluster
openframe dev skaffold my-cluster
```

This command:
1. Discovers `skaffold.yaml` files in your project
2. Optionally bootstraps a fresh cluster if needed
3. Runs `skaffold dev` with live rebuild-on-change

---

## Quick Reference — All Top-Level Commands

```bash
# Show the help menu and OpenFrame logo
openframe

# Get help for any command
openframe [command] --help

# Show version
openframe --version
```

| Command | Alias | Description |
|---------|-------|-------------|
| `openframe bootstrap` | — | Full one-shot setup (cluster + charts) |
| `openframe cluster` | `k` | Cluster lifecycle management |
| `openframe chart` | `c` | Helm chart and ArgoCD management |
| `openframe dev` | `d` | Local development workflows |

---

## Common Initial Configuration

### Verbose Mode

Add `-v` or `--verbose` to any command for detailed output — especially useful for watching ArgoCD sync progress:

```bash
openframe bootstrap my-cluster --deployment-mode=oss-tenant -v
```

### Silent Mode

Use `--silent` to suppress all output except errors — useful for scripting:

```bash
openframe cluster create my-cluster --silent
```

### Using Global Flags

Global flags apply to all commands:

```bash
openframe --verbose cluster create my-cluster
openframe --silent chart install --deployment-mode=oss-tenant
```

---

## Explore the Configuration Wizard

When you run `openframe chart install` or `openframe bootstrap` without `--non-interactive`, the CLI launches a multi-step configuration wizard covering:

1. **Deployment Mode** — Select `oss-tenant`, `saas-tenant`, or `saas-shared`
2. **Git Branch** — Choose which branch of the OpenFrame chart repo to deploy
3. **Docker Registry** — Configure container image source (GHCR or custom registry)
4. **Ingress** — Set domain and ingress routing configuration
5. **SaaS Credentials** — Provide API tokens if using SaaS mode

---

## Where to Get Help

If you run into issues or have questions:

- **OpenMSP Community Slack:** The primary support channel — [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **OpenMSP Website:** [https://www.openmsp.ai/](https://www.openmsp.ai/)
- **OpenFrame Platform:** [https://openframe.ai](https://openframe.ai)

> All bug reports, feature requests, and community discussions happen in the OpenMSP Slack — not GitHub Issues or GitHub Discussions.

---

## What's Next?

Now that you're familiar with the basics:

- Explore the [architecture documentation](../development/architecture/README.md) to understand how the CLI is structured
- Set up your [development environment](../development/setup/environment.md) to contribute to OpenFrame CLI
- Review [security best practices](../development/security/README.md) before deploying to production
