# How to install the OpenFrame platform onto a cluster

This tutorial shows an operator how to deploy the OpenFrame platform (ArgoCD plus the app-of-apps) onto a Kubernetes cluster using the `openframe` command-line tool.

## Prerequisites

Before you begin, check that:

- You have a working terminal with the `openframe` CLI binary available on your `$PATH` (downloaded from the [Releases](https://github.com/flamingo-stack/openframe-cli/releases/latest) page for your platform, for example `openframe-cli_windows_amd64.zip` on Windows).
- You already have a Kubernetes cluster to install onto. If you do not have one yet, create it first — the material for that is covered separately in [Create a Kubernetes cluster with OpenFrame CLI](./create-a-kubernetes-cluster-with-openframe-cli.md). The cluster can be one you created with `openframe cluster create`, one you created yourself, or any other cluster you can reach.
- Your machine meets the recommended capacity for the cluster you are installing into: at least 6 CPU cores and 24GB of RAM. If the cluster you pick falls short, the CLI warns you but still lets you continue.
- Your `kubectl` configuration (kubeconfig) already contains a context pointing at the target cluster, since the install command selects or accepts a kube-context rather than creating one.
- You know which GitHub ref (branch or release tag) of the platform chart repository you want deployed, if you do not want the default branch. The platform's source lives in the separate `flamingo-stack/openframe-oss-tenant` repository, which the CLI clones during installation — you do not need a local copy of it.

## Steps

1. Open a terminal and confirm the CLI is installed by checking that the `openframe` command is available on your `$PATH`.
2. Run `openframe app install` with no arguments to install onto a cluster interactively. The command shows the current kube-context in a dim header line and, when run bare (no cluster name, no `--non-interactive`, no `--dry-run`), prompts you to select which kube-context to install into.
3. Alternatively, run `openframe app install my-cluster` naming a specific cluster if you already know which one to target.
4. Alternatively, run `openframe app install --context <context-name>` to target an exact kube-context directly and skip the interactive selection prompt.
5. When prompted to select a kube-context, choose the context for the cluster you prepared in the prerequisites. If the selected cluster has fewer resources than recommended (around 6 cores / 24GB RAM), read the warning the CLI prints and decide whether to continue.
6. Watch the terminal as the command installs ArgoCD with its custom values, then installs the app-of-apps from the configured GitHub repository.
7. If you need to deploy a specific branch or release tag instead of the default, rerun the command with the `--ref` flag, for example `openframe app install --ref 1.0.48` for a release tag or `openframe app install --ref develop` for a branch.
8. If you are installing non-interactively (for example, in CI/CD) and already have an `openframe-helm-values.yaml` file prepared, run `openframe app install --non-interactive` to skip all prompts and reuse that file.
9. If you want to see what would be installed without making changes, run `openframe app install --dry-run` first.
10. If an installation needs to be forced even though charts already appear to exist, add the `--force` flag to the install command.
11. Wait for the command to finish waiting for all ArgoCD applications to become synced and healthy before considering the install complete.

## Expected result

When the install succeeds, ArgoCD and the app-of-apps are installed on the targeted cluster and all its ArgoCD applications report as synced and healthy. You can follow up by running `openframe app status` to see application sync/health, or `openframe app access` to get the ArgoCD admin credentials and how to reach the ArgoCD UI — both are documented separately in [Check the status of your OpenFrame cluster and app](./check-the-status-of-your-openframe-cluster-and-app.md).

## Troubleshooting

- **"`<command>` takes no positional arguments (got `<args>`) — use `--context` to target a cluster"**: you passed a stray positional argument on a command that targets a cluster by flag. Remove the extra argument or use `--context <name>` as the message suggests.
- **A branch-not-found style failure when deploying with `--ref`**: the `--ref` value you supplied does not exist in the chart repository. Re-check the branch name or release tag (release tags have no `v` prefix, e.g. `1.0.48`) and rerun the install.
- **"could not use context `<name>`"**: the `--context` value you passed does not match a context in your kubeconfig. Confirm the context name with your kubeconfig tooling and retry.
- **A warning that the cluster is smaller than recommended (~6 cores / 24GB RAM)**: this does not stop the install — it only warns. Continue if you accept the smaller capacity, or point the install at a larger cluster.
- **The command exits with a non-zero status after invoking an external tool**: the CLI preserves the failing tool's own exit code so automation can react to it; check the terminal output above the exit for the underlying message from that tool.
- For failures not covered here, the material does not show a specific message or fix, so check the terminal output of the `openframe app install` command directly for the error reported.
