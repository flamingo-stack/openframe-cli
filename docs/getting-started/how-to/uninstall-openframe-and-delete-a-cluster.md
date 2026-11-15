# How to uninstall OpenFrame and delete a cluster

This guide shows the person operating an OpenFrame deployment (e.g., a platform admin or developer who ran `openframe bootstrap` or `openframe cluster create`) how to remove the OpenFrame application from a cluster and, if needed, delete the cluster itself, using the `openframe` CLI.

## Prerequisites

Before you start, confirm the following:

- You have the `openframe` CLI installed and on your `$PATH` (check with `openframe --version`). If you don't have it yet, see [Create a new Kubernetes cluster with OpenFrame CLI](./create-a-new-kubernetes-cluster-with-openframe-cli.md) and [Install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md) for how it was set up.
- You know which cluster you want to act on. Run `openframe cluster list` to see the clusters the CLI manages, or `openframe cluster list --all` to also discover external GKE/EKS clusters.
- Your current kube-context points at the cluster you intend to change — the CLI prints a dim line naming the active kube-context before `app` subcommands run. Use the `--context` flag on `openframe app uninstall` if you need to target a context other than the current one.
- For a cloud cluster (GKE or EKS), the CLI's type-aware prerequisite check requires the matching tools (Terraform plus `gcloud` for GKE, or the AWS CLI for EKS) to be available, the same checks used when the cluster was created.
- You understand the scope of each command: `openframe app uninstall` removes ArgoCD and the OpenFrame apps but leaves the cluster running; `openframe cluster delete` destroys the cluster and all its resources. Decide which one (or both, in order) you need.

## Steps

1. Check which clusters exist by running `openframe cluster list` (add `--all` to also list external GKE/EKS clusters discovered read-only).
2. Remove the OpenFrame application from the cluster by running `openframe app uninstall`. Add `--context <name>` to target a cluster other than the current kube-context, and add `--delete-namespace` if you also want the `argocd` namespace removed.
3. Confirm the destructive action when prompted: "Remove ArgoCD and all OpenFrame apps from the current kube-context (or from context "<name>")? The cluster itself is kept." Type `y` to continue, or re-run the command with the `--yes` (`-y`) flag to skip this prompt in automation.
4. Wait for the uninstall to run. The CLI prints "Uninstalling OpenFrame..." while it deletes the ArgoCD Applications, then uninstalls the `app-of-apps` Helm release, then the ArgoCD Helm release itself, in that order.
5. Decide whether the cluster itself should also be deleted. If you only wanted the OpenFrame application removed, stop here — the cluster keeps running and you can reinstall later with `openframe app install`.
6. To delete the cluster as well, run `openframe cluster delete [NAME]`. Omit the name to pick the cluster interactively from the list of managed clusters.
7. Confirm the cluster deletion when prompted. For a local (k3d) cluster this is the standard yes/no confirmation. For a cloud cluster (GKE or EKS) you must re-type the exact cluster name to confirm, because destroying cloud infrastructure is irreversible; pass `--force` only in automation where no interactive prompt can run.
8. Wait for the CLI to finish stopping intercepts, deleting the cluster, cleaning up Docker resources (for k3d), and removing the cluster's configuration entry.
9. Optionally, if you only want to reclaim disk space from a k3d cluster's nodes without removing the cluster or its apps, run `openframe cluster cleanup [NAME]` instead of deleting anything — this prunes unused container images only.

## Expected result

- After `openframe app uninstall` completes successfully, you see a success message of the form "Removed `<N>` application(s) and `<N>` Helm release(s)." followed, if `--delete-namespace` was used, by "Deleted the argocd namespace." and then "The cluster was left running. Re-install any time with: openframe app install".
- If you answered "no" to the uninstall confirmation, you see "Uninstall cancelled." and nothing is changed.
- After `openframe cluster delete` completes successfully, the CLI shows a friendly success message for the delete operation and the cluster no longer appears in `openframe cluster list`.
- If you declined the cloud-cluster name confirmation, you see "Cluster name did not match — nothing was deleted" and the cluster is left untouched.

## Troubleshooting

- **"`<command>` takes no positional arguments (got `<args>`) — use --context to target a cluster"**: you passed a stray argument to `openframe app uninstall` (which only accepts `--context`, not a positional cluster name). Re-run the command using `--context <name>` instead.
- **"refusing to destroy cloud cluster '`<name>`' non-interactively; pass --force to confirm"**: you ran `openframe cluster delete` against a GKE or EKS cluster in a non-interactive session without `--force`. Add `--force` only when you are certain, since there is no prompt to catch a mistake.
- **"could not connect to the cluster"**: `openframe app uninstall` could not resolve a working connection to the target cluster context. Verify your kube-context (or the `--context` value you passed) actually points at a reachable cluster before retrying.
- **"uninstall failed: ..."**: the uninstall stopped partway through deleting ArgoCD Applications, uninstalling a Helm release, or clearing finalizers. The material does not show a general recovery procedure for this case beyond re-running the command; check the specific error text printed after this prefix for the underlying cause.
- **"failed to list clusters" / "failed to detect cluster type"**: `openframe cluster delete` or `openframe cluster cleanup` could not read the cluster registry or determine whether the cluster is k3d, EKS, or GKE. The material does not document a specific fix beyond confirming the cluster you named still exists and is visible to `openframe cluster list`.
