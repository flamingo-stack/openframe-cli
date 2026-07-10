# First Steps with OpenFrame CLI

With the CLI installed and an environment bootstrapped, here are the core commands and workflows.

## Explore the CLI

```bash
openframe --help
openframe cluster --help
openframe app --help
openframe update --help
```

Command groups:

- **bootstrap** — create a cluster and install the platform in one step
- **cluster** — k3d cluster lifecycle
- **app** — install, upgrade, inspect, and remove the OpenFrame app-of-apps deployment
- **prerequisites** — check and install required tools
- **update** — self-update the CLI
- **completion** — generate shell completion scripts

## Cluster Management

```bash
openframe cluster create              # create a k3d cluster (interactive wizard)
openframe cluster create dev -n 3     # named cluster with 3 nodes
openframe cluster list                # list clusters (add -o json|yaml)
openframe cluster status              # cluster health
openframe cluster delete dev -f       # delete without confirmation
openframe cluster cleanup             # remove leftover resources
```

`cluster create` flags: `--type/-t` (`k3d` only; cloud is coming soon), `--nodes/-n` (default 3), `--version`, `--skip-wizard`, `--dry-run`.

## Platform Deployment

`openframe app` manages the OpenFrame deployment. `app install` clones the `openframe-oss-tenant` repo and helm-installs the `app-of-apps` chart (helm release `app-of-apps`), which creates an ArgoCD root Application named `argocd-apps` that fans out to all child applications.

```bash
openframe app install                        # install into the current cluster
openframe app install --non-interactive      # reuse the existing openframe-helm-values.yaml
openframe app install --dry-run              # preview without applying

openframe app status                         # deployment status (add -o text|json|yaml)
openframe app upgrade --sync                 # force an ArgoCD re-sync
openframe app upgrade --prune                # re-sync and prune removed resources
openframe app access                         # print ArgoCD URL, admin creds, port-forward cmd
openframe app uninstall -y                   # remove the deployment
```

Key `app install` flags: `--github-repo`, `--ref/-r`, `--context/-c`, `--cert-dir`, `--non-interactive`, `--dry-run`, `--force/-f`.

`app install` deploys the OpenFrame platform app-of-apps — it does not install arbitrary charts.

## Access ArgoCD

```bash
openframe app access
```

This prints the ArgoCD URL, the admin username and password, and the `kubectl port-forward svc/argocd-server` command. Run the port-forward, then open the printed URL.

## Self-Update

The CLI updates itself from signed releases:

```bash
openframe update check          # is a newer release available? (add -o json|yaml)
openframe update                # download a checksum- and cosign-verified release, replace the binary (keeps a backup)
openframe update 1.4.2          # switch to a specific version (up or down)
openframe update rollback       # revert to the backed-up binary, offline
```

Opt into a daily auto-update (same-major only; disabled in CI / non-interactive):

```bash
export OPENFRAME_AUTO_UPDATE=1
```

## Common Workflows

### Recreate an environment

```bash
openframe cluster delete <name> -f
openframe bootstrap
```

### Inspect running workloads

```bash
kubectl get pods --all-namespaces
kubectl get applications -n argocd
```

## Getting Help

- **OpenMSP Slack**: [Join the community](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)

All support happens in Slack — we don't monitor GitHub Issues.
