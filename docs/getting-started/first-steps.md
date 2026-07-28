# First Steps

You've successfully bootstrapped an OpenFrame environment. Here are the first 5 things to explore and configure to get the most out of your installation.

---

## 1. Verify Your Environment

Start by confirming the state of your cluster and platform:

```bash
# Check cluster status
openframe cluster status

# Check application status
openframe app status
```

### Cluster Status Output

```text
NAME             STATUS    NODES    VERSION
openframe-dev    running   3        v1.29.x
```

### App Status Output

The app status command aggregates both Kubernetes cluster health and ArgoCD application sync state into a single unified report. You'll see each deployed application with its sync and health status.

```bash
# For machine-readable output (e.g., in scripts)
openframe cluster list --output json
openframe cluster status --output yaml
```

---

## 2. Access the OpenFrame Platform

Get access information for the deployed OpenFrame services:

```bash
openframe app access
```

This command displays the URLs and connection details for the OpenFrame platform running in your local cluster.

> **Tip:** Bookmark the displayed URLs for quick access to the OpenFrame web interface and ArgoCD dashboard.

---

## 3. Explore the Cluster Commands

The `cluster` command group (also aliased as `k`) manages your Kubernetes cluster lifecycle:

```bash
# List all managed clusters
openframe cluster list

# Get detailed status of a cluster
openframe cluster status

# Create an additional cluster with a custom name
openframe cluster create my-second-cluster

# Delete a cluster
openframe cluster delete my-second-cluster

# Clean up leftover resources from a failed cluster
openframe cluster cleanup
```

> **Shorthand:** `openframe k list` is equivalent to `openframe cluster list`.

---

## 4. Explore the App Commands

The `app` command group manages the OpenFrame platform deployment:

```bash
# Install OpenFrame on an existing cluster
openframe app install

# Check application status
openframe app status

# Upgrade to a different OpenFrame version/branch
openframe app upgrade

# Uninstall OpenFrame from a cluster
openframe app uninstall

# Show access details
openframe app access
```

### Upgrading OpenFrame

There are two upgrade modes:

```bash
# Mode 1: Switch to a different git ref (branch, tag, or commit)
openframe app upgrade --ref v2.0.0

# Mode 2: Force re-sync of the current ref
openframe app upgrade --force-sync
```

---

## 5. Keep the CLI Up to Date

OpenFrame CLI includes a built-in self-update mechanism with cryptographic verification:

```bash
# Check if an update is available
openframe update --check

# Apply the latest update
openframe update

# Roll back to the previous version if needed
openframe update --rollback
```

> **Security note:** All updates are verified using [Sigstore/cosign](https://docs.sigstore.dev/cosign/overview/) against the official GitHub Actions release workflow. Only binaries produced by the `flamingo-stack/openframe-cli` release pipeline are accepted.

---

## Initial Configuration: The Helm Values File

When you ran `openframe bootstrap`, a configuration file called `openframe-helm-values.yaml` was created in your working directory. This file controls the OpenFrame platform deployment:

```bash
# View the generated configuration
cat openframe-helm-values.yaml
```

Key configurable areas include:

| Section | Description |
|---|---|
| `branch` | The OpenFrame git ref (branch, tag) to deploy |
| `docker` | Container registry settings |
| `ingress` | Ingress hostname and TLS configuration |
| `argocd` | ArgoCD Helm value overrides |

To apply changes to an existing deployment:

```bash
openframe app upgrade
```

---

## Verbose and Silent Modes

Control the CLI's output verbosity:

```bash
# Show detailed debug output (ArgoCD sync events, Helm operations, etc.)
openframe bootstrap --verbose

# Suppress all non-error output (perfect for scripts)
openframe bootstrap --silent

# Machine-readable output for cluster commands
openframe cluster list --output json
```

---

## Running in CI/CD

For automated pipelines, use `--non-interactive` to skip all prompts:

```bash
# Full non-interactive bootstrap
openframe bootstrap --non-interactive

# With a specific cluster name
openframe bootstrap --non-interactive my-ci-cluster
```

The CLI reads from an existing `openframe-helm-values.yaml` file in the current directory when running non-interactively.

---

## Getting Help

Every command has built-in help:

```bash
# General help
openframe --help

# Help for a specific command
openframe bootstrap --help
openframe cluster create --help
openframe app install --help
openframe update --help
```

---

## Community & Support

- **OpenMSP Slack:** [https://www.openmsp.ai/](https://www.openmsp.ai/) — Join for help, discussions, and announcements
- **Slack invite:** [https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **OpenFrame platform repo:** [https://github.com/flamingo-stack/openframe-oss-tenant](https://github.com/flamingo-stack/openframe-oss-tenant)
- **CLI releases:** [https://github.com/flamingo-stack/openframe-cli/releases](https://github.com/flamingo-stack/openframe-cli/releases)

---

## Summary: First Steps Checklist

- [ ] Verified cluster status with `openframe cluster status`
- [ ] Checked app status with `openframe app status`
- [ ] Accessed the platform with `openframe app access`
- [ ] Explored `openframe cluster --help` and `openframe app --help`
- [ ] Reviewed `openframe-helm-values.yaml` configuration file
- [ ] Ran `openframe update --check` to see if a newer version is available
- [ ] Joined the [OpenMSP Slack](https://www.openmsp.ai/) community
