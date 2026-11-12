# How to uninstall OpenFrame and delete a cluster

This guide walks a cluster operator through removing the OpenFrame platform from a cluster and, optionally, deleting the cluster itself using the OpenFrame CLI (`openframe`).

## Prerequisites

- The OpenFrame CLI (`openframe`) is installed and on your `$PATH`. You can confirm this by running `openframe --help` and seeing the command list.
- You have a working `kubectl` context pointing at the cluster you want to clean up, or you know the cluster's kube-context name to pass with `--context`.
- The cluster is reachable (Helm and the Kubernetes API must respond) if you intend to run `openframe app uninstall`.
- For a local cluster created with `openframe cluster create` or `openframe bootstrap`, Docker is running.
- For a cloud cluster (EKS or GKE), you have valid cloud credentials configured (AWS or gcloud) and know the exact cluster name as it appears in `openframe cluster list`, since cloud deletions ask you to re-type the name.
- You understand the difference between the two operations covered here: `openframe app uninstall` removes the OpenFrame application (ArgoCD and its apps) but leaves the cluster running; `openframe cluster delete` removes the cluster itself and its underlying infrastructure.

## Steps

1. Open a terminal and confirm which cluster/context you are about to act on by running `kubectl config current-context`, or decide which `--context` value you will pass to the CLI.
2. Run `openframe app uninstall` to remove the OpenFrame application from the cluster. Add `--context <context-name>` if you want to target a context other than the current one, for example `openframe app uninstall --context k3d-openframe-dev`.
3. When prompted "Remove ArgoCD and all OpenFrame apps from the current kube-context? The cluster itself is kept.", confirm by answering yes, or avoid the prompt entirely by adding the `--yes` flag (`openframe app uninstall --yes`), which is useful for automation.
4. Optionally add the `--delete-namespace` flag (for example `openframe app uninstall --yes --delete-namespace`) if you also want the `argocd` namespace removed once the applications and Helm releases are gone.
5. Wait for the CLI to report "Uninstalling OpenFrame..." followed by a summary of how many applications and Helm releases were removed. At this point the OpenFrame application is gone but the cluster is still running.
6. Decide whether you also want to delete the cluster. If you only wanted to remove OpenFrame and keep the cluster, stop here — the CLI already told you it left the cluster running and that you can reinstall with `openframe app install`.
7. To delete the cluster, run `openframe cluster delete [NAME]`. Pass the cluster's name as the positional argument (for example `openframe cluster delete my-cluster`), or run `openframe cluster delete` with no name to pick the cluster from an interactive selection list.
8. Confirm the deletion when prompted. For a local (k3d) cluster this is a general confirmation; for a cloud cluster (EKS or GKE) you must re-type the exact cluster name to confirm, because destroying cloud infrastructure is irreversible. You can skip either prompt non-interactively by adding `--force` (for example `openframe cluster delete my-cluster --force`).
9. Wait for the CLI to show the operation start message, run the type-aware prerequisite checks for that cluster's type, and perform the deletion, which stops intercepts, deletes the cluster, cleans up Docker resources (for k3d) or the underlying cloud infrastructure (for EKS/GKE), and removes the cluster's local configuration.
10. Confirm success by watching for the CLI's operation-success message for the delete action against your cluster name.

## Expected result

- After `openframe app uninstall` completes successfully, you see an "Uninstalling OpenFrame..." message followed by a line such as "Removed `<N>` application(s) and `<N>` Helm release(s)." If you passed `--delete-namespace`, you also see a confirmation that the `argocd` namespace was deleted. The CLI then prints that the cluster was left running and that you can reinstall any time with `openframe app install`.
- After `openframe cluster delete` completes successfully, the CLI shows an operation-start message for the delete action, followed by an operation-success message once the cluster and its associated resources (Docker resources for k3d, or cloud infrastructure for EKS/GKE) have been removed.
- If you cancel the `openframe app uninstall` confirmation prompt, the CLI prints "Uninstall cancelled." and makes no changes.
- If you decline a cloud cluster's typed-name confirmation during `openframe cluster delete` (the name you typed does not match), the CLI prints "Cluster name did not match — nothing was deleted" and the cluster is left untouched.

## Troubleshooting

- **"`<command>` takes no positional arguments (got `<args>`) — use `--context` to target a cluster"** — you passed a stray argument to an `app` subcommand such as `uninstall`. Remove the extra argument and use the `--context` flag instead to target a specific cluster.
- **"could not connect to the cluster"** during `openframe app uninstall` — the CLI could not resolve or reach the kube-context. Verify `kubectl config current-context` points at a reachable cluster, or pass the correct `--context` value.
- **"uninstall failed: ..."** — the uninstall stopped partway through (for example while removing ArgoCD applications, uninstalling a Helm release, or clearing application finalizers). Re-run `openframe app uninstall` once the underlying issue is resolved; the operation is designed to be safe to retry since it removes resources in a fixed order (Applications, then Helm releases, then finalizers, then optionally the namespace).
- **"refusing to destroy cloud cluster '`<name>`' non-interactively; pass `--force` to confirm"** — you ran `openframe cluster delete` against an EKS or GKE cluster in a non-interactive session without `--force`. Re-run the command with `--force`, or run it interactively so you can type the cluster name to confirm.
- **"failed to list clusters" / "failed to detect cluster type"** — the CLI could not read your cluster configuration or determine whether the target is k3d, EKS, or GKE. Confirm the cluster name is correct and that any required cloud credentials (AWS or gcloud) are configured, since the material does not describe further diagnostics beyond this message.
- The material does not show how to verify the Helm releases or namespace were actually removed from the cluster afterward (for example via `kubectl` or `helm` commands); check your cluster directly if you need that confirmation.
