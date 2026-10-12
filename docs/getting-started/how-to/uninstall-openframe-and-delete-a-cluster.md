# How to uninstall OpenFrame and delete a cluster

This guide walks a platform operator through removing the OpenFrame application from a cluster, and then (optionally) deleting the cluster itself, using the `openframe` command-line tool.

## Prerequisites

- The `openframe` CLI is installed and runnable (check with `openframe --help`).
- You have a Kubernetes cluster that OpenFrame was installed onto, either created with `openframe cluster create`/`openframe bootstrap`, or any other cluster you can already reach with `kubectl`.
- Your kube-context is set to the cluster you intend to act on. You can check the active context by running `openframe cluster status`, or pass the correct context explicitly with the `--context` flag on commands that support it.
- For a cloud cluster (GKE or EKS), the underlying cloud CLI (`gcloud` or `aws`) is authenticated, since deleting a cloud cluster uses Terraform and the cloud provider's APIs.
- You know whether you only want to remove the OpenFrame application (keeping the cluster running) or also delete the cluster entirely — these are two separate, independent operations.

## Steps

1. Open a terminal and confirm which cluster you are targeting by running `openframe cluster status` (or note the `--context` value you plan to pass), since the uninstall command acts on the current kube-context unless you override it.
2. Run `openframe app uninstall` to remove the OpenFrame application from the cluster. Add `--context <name>` to target a specific cluster instead of the current context, for example `openframe app uninstall --context k3d-openframe-dev`.
3. When prompted "Remove ArgoCD and all OpenFrame apps from the current kube-context? The cluster itself is kept.", type the confirmation to proceed, or re-run the command with `--yes` to skip the prompt in an automated session (for example `openframe app uninstall --yes`).
4. Optionally, add the `--delete-namespace` flag to also delete the `argocd` namespace once the application and Helm releases are removed, for example `openframe app uninstall --yes --delete-namespace`.
5. Wait for the command to report that it uninstalled OpenFrame; this step only removes the ArgoCD applications and Helm releases — the cluster is left running and reachable.
6. Decide whether you also want to delete the cluster. If you only wanted to remove the OpenFrame application, stop here; the cluster remains available for `openframe app install` to reinstall OpenFrame later.
7. To delete the cluster, run `openframe cluster delete` with no arguments to pick the cluster interactively from a list, or run `openframe cluster delete <NAME>` to target one by name directly.
8. If the cluster is a local cluster, confirm the interactive yes/no prompt to proceed, or pass `--force` to skip it in an automated session, for example `openframe cluster delete my-cluster --force`.
9. If the cluster is a cloud cluster (GKE or EKS), type the cluster's exact name when prompted to confirm the destructive deletion, since cloud clusters require this stronger, typed-name confirmation instead of a simple yes/no (unless `--force` is passed).
10. Wait for `openframe cluster delete` to finish; it stops any intercepts, deletes the cluster, cleans up associated Docker resources (for a local cluster) or cloud infrastructure (for a cloud cluster), and removes the cluster from the CLI's cluster list.

## Expected result

- After `openframe app uninstall` completes successfully, you see a success message reporting how many applications and Helm releases were removed, for example a line stating the number of application(s) and Helm release(s) removed. If you passed `--delete-namespace`, a further success line confirms the `argocd` namespace was deleted. A final info line states that the cluster was left running, and that you can reinstall at any time with `openframe app install`.
- If you cancel the uninstall confirmation prompt, the CLI prints "Uninstall cancelled." and exits without changing anything.
- After `openframe cluster delete` completes successfully, the CLI prints a friendly success message confirming the cluster was deleted, and the cluster no longer appears in `openframe cluster list`.
- If you decline the cloud-cluster typed-name confirmation (the name you typed does not match), the CLI prints "Cluster name did not match — nothing was deleted" and leaves the cluster untouched.

## Troubleshooting

- **"`<command>` takes no positional arguments (got `<args>`) — use `--context` to target a cluster"**: you passed a cluster name directly after `openframe app uninstall` (or a similar `app` subcommand). These commands only accept a target cluster through the `--context` flag; re-run the command using `--context <name>` instead of a positional argument.
- **"refusing to destroy cloud cluster '`<name>`' non-interactively; pass `--force` to confirm"**: you ran `openframe cluster delete` against a GKE or EKS cluster from a non-interactive session (such as a script or CI job) without `--force`. Re-run the command with `--force` to confirm the destructive deletion explicitly.
- **"Cluster name did not match — nothing was deleted"**: during the cloud-cluster confirmation prompt, the text you typed did not match the cluster's name exactly. Run `openframe cluster delete <NAME>` again and type the name exactly as shown.
- **Uninstall fails with "could not connect to the cluster"**: the CLI could not resolve or reach the kube-context you specified (or the current one). Verify your context with `openframe cluster status` and make sure the cluster is online before retrying.
- **Uninstall fails with "uninstall failed: ..."**: the underlying ArgoCD or Helm removal step failed. The material does not show a specific remedy for every underlying cause; review the printed error message for the failing step and retry once the condition it describes (for example, cluster connectivity) is resolved.
