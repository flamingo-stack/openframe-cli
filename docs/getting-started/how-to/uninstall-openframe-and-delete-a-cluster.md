# How to uninstall OpenFrame and delete a cluster

This guide shows the person operating an OpenFrame deployment how to remove the OpenFrame platform from a cluster and, if the cluster itself is no longer needed, delete the cluster entirely.

## Prerequisites

- The OpenFrame CLI is installed and the `openframe` command runs from your terminal. See [How to update the OpenFrame CLI to the latest version](./update-the-openframe-cli-to-the-latest-version.md) if you are not sure which version you have.
- You have a kube-context pointing at the cluster you want to clean up. Check your current context before running any command below — `openframe app uninstall` and `openframe app upgrade` warn you that "a dim line naming the current kube-context" is shown before destructive operations, as "the cheapest guard against installing onto the wrong cluster."
- You know whether the cluster was created as a local **k3d** cluster or a cloud cluster (**GKE** or **EKS**). You can check with `openframe cluster list` or `openframe cluster status [NAME]`, described in [How to check the status of your OpenFrame cluster and app](./check-the-status-of-your-openframe-cluster-and-app.md).
- For a cloud cluster (GKE or EKS), you have the credentials and CLI tools the cluster's provider needs already working (`gcloud` for GKE, `aws` for EKS), since deleting it tears down billed infrastructure.
- You are prepared for destructive, typically irreversible actions: removing OpenFrame deletes its ArgoCD applications and Helm releases, and deleting a cluster removes the cluster and, for cloud clusters, the underlying infrastructure.

## Steps

1. Decide what you want to remove: run `openframe app uninstall` to remove OpenFrame but keep the cluster running, or continue to cluster deletion if you want the cluster gone too. The command's own description states: "Deletes the ArgoCD applications and uninstalls the ArgoCD and app-of-apps Helm releases. The cluster itself is NOT deleted."
2. If you want to target a cluster other than your current kube-context, add the `--context` flag, for example `openframe app uninstall --context k3d-openframe-dev`. Do not pass the cluster name as a plain argument — the command rejects positional arguments and tells you to use `--context` instead, to avoid acting on the wrong cluster.
3. Respond to the confirmation prompt: "Remove ArgoCD and all OpenFrame apps from [the current kube-context / context "NAME"]? The cluster itself is kept." Confirm to proceed, or answer no to cancel.
4. To skip this prompt in a script or CI job, add the `--yes` flag (for example `openframe app uninstall --yes`). Without a terminal to prompt in, the command refuses to run unless `--yes` is given.
5. Optionally add `--delete-namespace` (for example `openframe app uninstall --yes --delete-namespace`) if you also want the `argocd` namespace removed after the applications and Helm releases are gone.
6. Watch the uninstall messages: `Uninstalling OpenFrame...`, followed by a success line reporting how many applications and Helm releases were removed, and — if you asked for it — a line confirming the `argocd` namespace was deleted.
7. If you only wanted OpenFrame removed, stop here. The CLI's own closing message tells you how to bring it back: "The cluster was left running. Re-install any time with: openframe app install" — see [How to install the OpenFrame platform onto a cluster](./install-the-openframe-platform-onto-a-cluster.md) when you are ready.
8. To also remove the cluster, run `openframe cluster delete [NAME]`. Omitting the name triggers interactive selection; the command prunes Docker resources, deletes the cluster, and removes its configuration.
9. If you do not know the exact cluster name, run `openframe cluster delete` with no argument and pick the cluster from the interactive list the command presents.
10. If the cluster is a cloud cluster (GKE or EKS), re-type the cluster's name exactly when asked to confirm. This stronger confirmation step exists specifically for cloud clusters — local clusters do not require it.
11. To skip the cloud confirmation prompt in automation, add `--force` (for example `openframe cluster delete my-cluster --force`). In a non-interactive session without `--force`, the command refuses outright rather than hang on a prompt or guess an answer.
12. Watch for the operation-start and operation-success messages the command prints for the delete action, naming the cluster and, for a successful run, its type (k3d, GKE, or EKS).

## Expected result

- After `openframe app uninstall` succeeds, you see a line such as `Removed N application(s) and N Helm release(s).`, optionally `Deleted the argocd namespace.`, and finally `The cluster was left running. Re-install any time with: openframe app install`.
- If you cancel the uninstall confirmation, you see `Uninstall cancelled.` and nothing is removed.
- After `openframe cluster delete` succeeds, the CLI reports the operation as successful for the named cluster and its detected type (k3d, GKE, or EKS); the cluster, its Docker resources (for k3d), and its configuration entry are gone.
- If you decline the cloud-cluster typed-name confirmation, you see `Cluster name did not match — nothing was deleted`, and the cluster is left untouched.

## Troubleshooting

- **`"app uninstall" takes no positional arguments (got "...") — use --context to target a cluster`**: you passed a cluster name directly after `uninstall`. Re-run the command with `--context <name>` instead.
- **`refusing to destroy cloud cluster '<name>' non-interactively; pass --force to confirm`**: you ran `openframe cluster delete` against a GKE or EKS cluster from a non-interactive session (for example a CI job) without `--force`. Add `--force` if you are certain, or run the command interactively so you can re-type the cluster name to confirm.
- **`Cluster name did not match — nothing was deleted`**: you were asked to re-type the cloud cluster's name to confirm deletion and what you typed did not match exactly. Run `openframe cluster delete <name>` again and type the name precisely, or pass `--force` to skip the prompt.
- **The command seems to hang waiting for input**: a destructive command such as `openframe app uninstall` is waiting for a yes/no confirmation. Either answer the prompt, or re-run with `--yes` (for uninstall) or `--force` (for cloud cluster delete) if you intend to automate it.
- **You uninstalled OpenFrame but still see the cluster running**: this is expected — `openframe app uninstall` only removes OpenFrame (ArgoCD and its apps, and optionally the `argocd` namespace); it never deletes the cluster. Use `openframe cluster delete` separately if you also want the cluster gone.
- The material does not describe how to recover a cluster or its data once `openframe cluster delete` has completed; treat the deletion as final.
