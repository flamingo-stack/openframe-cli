# First Steps

You've installed OpenFrame CLI and completed the [quick start](quick-start.md). Here's what to do next to get comfortable with day-to-day usage.

## 1. Check the platform's status

Get a snapshot of cluster health and ArgoCD application readiness:

```bash
openframe app status
```

For a live-refreshing view that polls every few seconds:

```bash
openframe app status --watch
```

For an interactive, k9s-style dashboard where you can navigate applications, inspect details, and trigger syncs:

```bash
openframe app status --interactive
```

> `--watch` and `--interactive` require an interactive terminal and cannot be combined with `--output`/`--plain` modes.

## 2. Retrieve ArgoCD access credentials

To sign in to the ArgoCD UI and manage the platform visually:

```bash
openframe app access
```

This prints the admin username/password and port-forward instructions, e.g.:

```text
ArgoCD access
  Username: admin
  Password: <redacted>
Open the ArgoCD UI:
  1. kubectl port-forward -n argocd svc/argocd-server 8080:443
  2. open https://localhost:8080
```

## 3. Inspect your cluster(s)

List and inspect the clusters OpenFrame CLI manages:

```bash
# List all managed clusters
openframe cluster list

# Include externally-discovered cloud clusters (GKE/EKS) not created via this CLI
openframe cluster list --all

# Detailed cluster status (nodes, health)
openframe cluster status

# Switch kubectl context to a specific cluster
openframe cluster use <cluster-name>
```

## 4. Explore configuration options

If you skipped the interactive wizard during bootstrap, explore the flags available for a full cluster + install workflow:

```bash
openframe cluster create --help
openframe app install --help
```

Key things worth trying:

- `openframe cluster create my-gke --type gke --project my-project --region us-central1 --skip-wizard` — provision a cloud cluster non-interactively.
- `openframe app install --ref v1.4.0` — install a specific OpenFrame release ref instead of the default branch.
- `openframe app upgrade --ref v1.4.1` — move an existing install to a different git ref, or `--sync` to force a re-sync at the current ref.

## 5. Keep the CLI itself up to date

```bash
# Check for a newer CLI release without installing it
openframe update check

# Update to the latest release
openframe update

# Roll back to the previously installed binary
openframe update rollback
```

## Where to Get Help

- Run `openframe --help` or `openframe <command> --help` at any point — every command has detailed built-in help text and usage examples.
- Run `openframe prerequisites check --type <k3d|eks|gke>` if something isn't working — most failures are missing/misconfigured local tooling.
- For questions, discussion, and support, the project is community-supported via the **OpenMSP Slack community**: [join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA) or visit [openmsp.ai](https://www.openmsp.ai/).
- The OpenFrame platform code that this CLI deploys lives in [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant); its documentation is at [github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs](https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs).
