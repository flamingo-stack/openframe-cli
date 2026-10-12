## How to install the OpenFrame platform onto a cluster

This tutorial walks a cluster operator through deploying the OpenFrame platform (ArgoCD plus the app-of-apps) onto an existing Kubernetes cluster using the OpenFrame CLI, so the platform's applications come up and sync automatically.

## Prerequisites

Before you start, check that:

- You have an existing, reachable Kubernetes cluster. This can be one created with `openframe cluster create` (see [Create a Kubernetes cluster with OpenFrame CLI](./create-a-kubernetes-cluster-with-openframe-cli.md)), one you built yourself, or any other cluster you can reach — `openframe app install` works against any kube-context.
- Your machine meets the recommended capacity for the platform: roughly 6 CPU cores and 24GB RAM. If the selected cluster is smaller, the CLI warns you but lets you continue.
- The `openframe` CLI binary is installed and on your `$PATH`. If you are not sure which version you have, you can check with `openframe update check` (see [Update the OpenFrame CLI to the latest version](./update-the-openframe-cli-to-the-latest-version.md)).
- Helm is available, because the install step installs ArgoCD via Helm. The material does not show a separate manual Helm install step for this flow — the CLI's own install command verifies and drives this for you.
- Your kubeconfig contains a context pointing at the target cluster — the CLI resolves the cluster either from an explicit `--context` flag or from an interactive prompt that lists your kube-contexts.
- You know which Git ref (branch or release tag) of the OpenFrame platform chart repository you want deployed, if you don't want the default branch. Release tags follow a plain numeric scheme, for example `1.0.48` (no `v` prefix).

## Steps

1. Open a terminal and confirm the `openframe` command is available by running `openframe --help`.
2. Run `openframe app install` with no arguments to start the fully interactive flow, or run `openframe app install my-cluster` to target a specific cluster by name.
3. If you want to target a specific kube-context directly instead of being prompted, add the `--context` flag, for example `openframe app install --context my-cluster`.
4. When prompted and no cluster name or `--context` was given, choose the kube-context to install into from the interactive selector. The CLI reports the resource size of the chosen cluster against the recommended minimum (about 6 cores / 24GB RAM) and continues even if it is smaller, after printing a warning.
5. If you need to deploy a specific branch or release tag instead of the default, add the `--ref` flag, for example `openframe app install --ref develop` for a branch or `openframe app install --ref 1.0.48` for a release tag.
6. If you are running this non-interactively (for example in CI/CD) against an existing `openframe-helm-values.yaml` file, add the `--non-interactive` flag: `openframe app install --non-interactive`.
7. If you want to preview what would happen without making changes, add the `--dry-run` flag: `openframe app install --dry-run`.
8. Confirm the installation when prompted (this confirmation step is skipped automatically in `--non-interactive` and `--dry-run` modes).
9. Wait while the CLI installs ArgoCD with its custom values, installs the app-of-apps from the configured GitHub repository, and then waits for all ArgoCD applications to become ready.
10. Once the command finishes, check platform readiness by running `openframe app status` (see [Check the status of your OpenFrame cluster and app](./check-the-status-of-your-openframe-cluster-and-app.md)).
11. Retrieve ArgoCD admin credentials and UI access instructions by running `openframe app access`.

## Expected result

- While installation runs, you see progress messages for each stage: configuration mode (default or interactive), confirmation of the install target context (for example `Installing OpenFrame into context "my-cluster"`), certificate regeneration, the ArgoCD install, the app-of-apps install, and the wait for applications to sync.
- In dry-run mode, you see "Using existing configuration (dry-run mode)" and "Would run:" lines instead of real installation actions, and no cluster changes are made.
- In non-interactive mode, you see "Running in non-interactive mode" before the CLI loads your existing `openframe-helm-values.yaml`.
- If your chosen cluster is smaller than the recommended size, you see a warning naming the context and the recommended cores/RAM, followed by continued installation.
- On success, the app-of-apps applications in ArgoCD reach a synced and healthy state, which you can confirm afterwards with `openframe app status`.

## Troubleshooting

- **`"<command>" takes no positional arguments (got "<args>") — use --context to target a cluster`**: you passed a positional cluster name together with a command that only accepts `--context`. Re-run the command using `--context <name>` instead of a bare argument.
- **`no cluster selected — nothing was installed`**: no cluster was chosen during interactive selection (for example the prompt was cancelled, or no clusters were found). Re-run `openframe app install` and complete the cluster selection prompt, or pass `--context`/a cluster name explicitly.
- **`could not use context "<name>": ...`**: the `--context` value you supplied does not resolve to a usable kube-context. Check your kubeconfig and the context name, then retry.
- **`non-interactive configuration failed: ...`**: `--non-interactive` was used but no usable `openframe-helm-values.yaml` was found or it failed to load. Make sure the file exists in the expected location before using `--non-interactive`, or omit the flag to run the interactive wizard instead.
- **`helm values pre-flight failed: ...`**: the Helm values file failed validation before any cluster changes were made. Review the values file content referenced in the error and correct it, then re-run the install.
- **A branch-not-found style error on the chart repository ref**: the `--ref` you supplied does not exist in the chart repository. Re-run with a valid branch name or release tag (for example `1.0.48`).
- **Installation cancelled at the confirmation prompt**: if you answer "no" at the confirmation step, the CLI reports the installation as cancelled and exits non-zero; re-run `openframe app install` and confirm when ready.
