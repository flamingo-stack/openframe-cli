# First Steps

You've bootstrapped a cluster and installed OpenFrame. Here's what to explore next.

## 1. Inspect Your Cluster

List all clusters managed by OpenFrame CLI, and check their health:

```bash
openframe cluster list
openframe cluster status my-cluster
```

Use `--all` on `list` to also discover external (non-OpenFrame-managed) EKS/GKE clusters visible to your cloud credentials:

```bash
openframe cluster list --all
```

## 2. Switch Between Clusters

If you manage more than one cluster, switch your active kubectl context with:

```bash
openframe cluster use my-cluster
```

## 3. Monitor Platform Health

Beyond a one-shot check, use the watch or interactive modes to keep an eye on ArgoCD application sync/health state while the platform stabilizes:

```bash
openframe app status --watch
openframe app status --interactive
```

Machine-readable output is available for scripting:

```bash
openframe app status --output json
```

## 4. Get ArgoCD Access

Retrieve the ArgoCD admin username/password and instructions for port-forwarding into the UI:

```bash
openframe app access
```

```text
ArgoCD access
  Username: admin
  Password: <generated-password>
Open the ArgoCD UI:
  1. kubectl port-forward -n argocd svc/argocd-server 8080:443
  2. open https://localhost:8080
```

## 5. Upgrade or Reconfigure the Platform

To move to a different branch/release ref, or force a re-sync:

```bash
openframe app upgrade my-cluster --ref v1.3.0
openframe app upgrade my-cluster --sync
```

## Common Day-2 Operations

| Task | Command |
|---|---|
| Reinstall platform on an existing cluster | `openframe app install --non-interactive` |
| Remove the platform but keep the cluster | `openframe app uninstall --yes` |
| Free disk space by pruning old images | `openframe cluster cleanup my-cluster` |
| Tear down a cluster entirely | `openframe cluster delete my-cluster --force` |
| Check for CLI updates | `openframe update check` |
| Update the CLI itself | `openframe update` |
| Roll back a bad CLI update | `openframe update rollback` |

## Exploring Global Flags

These flags apply across most command groups:

| Flag | Effect |
|---|---|
| `--verbose` / `-v` | Timestamped debug output, including ArgoCD sync progress |
| `--silent` | Only errors are printed |
| `--plain` | Sequential output without spinners (useful for log files/CI) |
| `-o json` / `-o yaml` | Machine-readable output where supported (e.g., `cluster list`, `app status`, `app access`) |

## Where to Get Help

- Run `openframe --help` or `openframe <command> --help` for built-in usage information on any command.
- For architecture and internals, see the [Development documentation](../development/README.md).
- For community support, join the OpenMSP Slack community: [openmsp.ai](https://www.openmsp.ai/) ([invite link](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)).

This project does not use GitHub Issues or Discussions for support — all questions and discussion happen in the OpenMSP Slack community.
