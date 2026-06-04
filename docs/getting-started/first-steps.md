# First Steps

Congratulations on your first successful OpenFrame bootstrap! This guide walks you through the first five things to do after your environment is up and running.

[![OpenFrame v0.3.7 - Enhanced Developer Experience](https://img.youtube.com/vi/O8hbBO5Mym8/maxresdefault.jpg)](https://www.youtube.com/watch?v=O8hbBO5Mym8)

---

## 1. Explore Your Cluster

Start by familiarizing yourself with what was created.

### List All Clusters

```bash
openframe cluster list
```

This displays a formatted table of all K3D clusters managed by OpenFrame, including their names, types, and states.

### Check Detailed Status

```bash
openframe cluster status my-cluster --detailed
```

The `--detailed` flag shows:
- Cluster node health and roles
- All ArgoCD applications and their sync/health status
- Any applications that need attention

### Check Without App Details

```bash
openframe cluster status my-cluster --no-apps
```

Useful for a quick cluster health check without the full ArgoCD application list.

---

## 2. Interact With the Cluster Using kubectl

OpenFrame automatically manages your `kubeconfig` when creating clusters. After bootstrap, you can use `kubectl` directly:

```bash
# View all namespaces
kubectl get namespaces

# Check ArgoCD applications
kubectl get applications -n argocd

# Check all pods across namespaces
kubectl get pods --all-namespaces
```

> **Tip:** If you manage multiple clusters, use `kubectl config get-contexts` to see which context is currently active, and `kubectl config use-context k3d-my-cluster` to switch.

---

## 3. Install Charts on an Existing Cluster

If you want to re-install or update the OpenFrame charts on a cluster that already exists, use the `chart install` command:

```bash
# Interactive mode — launches the configuration wizard
openframe chart install

# Direct mode — specify cluster and deployment mode
openframe chart install my-cluster --deployment-mode=oss-tenant --verbose
```

### Configuration Wizard

When running interactively, the configuration wizard guides you through:

1. **Deployment mode selection**: `oss-tenant`, `saas-tenant`, or `saas-shared`
2. **Configuration mode**: Default (recommended) or interactive (custom branch, Docker, ingress)
3. **GitHub repository settings** (if using a custom branch)
4. **Docker registry configuration** (for SaaS modes)
5. **Ingress configuration** (domain and routing settings)

---

## 4. Set Up a Developer Intercept

One of OpenFrame CLI's most powerful features for developers is **Telepresence service intercepts** — the ability to route live Kubernetes traffic to your local machine for debugging.

### Interactive Intercept Setup

```bash
openframe dev intercept
```

The wizard prompts you to:
1. Select the target namespace
2. Select the service to intercept
3. Specify the local port to forward traffic to

### Direct Intercept

```bash
# Intercept a specific service on port 3000 in the development namespace
openframe dev intercept my-api-service --port 3000 --namespace development

# Intercept with environment variable export and custom header
openframe dev intercept my-api-service --port 8080 --env-file .env.local --header "x-user=developer"
```

When an intercept is active, all traffic sent to `my-api-service` inside the cluster is forwarded to your local process on the specified port.

> **Prerequisites:** Telepresence must be installed. See [Prerequisites](prerequisites.md) for installation links.

---

## 5. Run a Live-Reload Development Session

Use Skaffold to deploy your service with automatic live reloading every time you save a file:

```bash
# Interactive — prompts you to select a Skaffold config and cluster
openframe dev skaffold

# Direct — specify cluster name
openframe dev skaffold my-cluster

# Skip bootstrapping if cluster is already set up
openframe dev skaffold my-cluster --skip-bootstrap

# With a custom Helm values file
openframe dev skaffold my-cluster --helm-values ./my-values.yaml
```

Skaffold watches your source files, rebuilds your container image on change, and redeploys to the cluster automatically — giving you a tight inner-loop development cycle inside Kubernetes.

> **Prerequisites:** Skaffold must be installed. If it is missing, the CLI will offer to install it automatically.

---

## Common Initial Configuration

### Verbose Output

For any command, add `-v` or `--verbose` to get detailed logs including ArgoCD sync progress:

```bash
openframe bootstrap --verbose
openframe chart install my-cluster --verbose
```

### Silent Mode

Suppress all output except errors (useful in scripts):

```bash
openframe cluster list --silent
```

### Force Delete

If a cluster deletion gets stuck, use `--force`:

```bash
openframe cluster delete my-cluster --force
```

### Cluster Cleanup

Remove unused Docker images and resources from cluster nodes:

```bash
openframe cluster cleanup my-cluster
```

---

## Deployment Mode Summary

| Mode | What Gets Deployed | When to Use |
|---|---|---|
| `oss-tenant` | Full self-hosted OpenFrame stack | Default — self-hosted MSP setup |
| `saas-tenant` | SaaS tenant configuration | When connecting to hosted Flamingo infrastructure |
| `saas-shared` | Shared SaaS platform | Multi-tenant shared platform deployment |

---

## Getting Help

Every command has built-in help:

```bash
openframe --help
openframe cluster --help
openframe cluster create --help
openframe chart install --help
openframe dev intercept --help
```

For community support and discussions, join the **OpenMSP Slack**:

👉 [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

- 🌐 OpenMSP Community: [https://www.openmsp.ai/](https://www.openmsp.ai/)
- 🌐 OpenFrame: [https://openframe.ai](https://openframe.ai)
- 🌐 Flamingo: [https://flamingo.run](https://flamingo.run)
