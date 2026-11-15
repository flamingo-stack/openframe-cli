# How to install the OpenFrame platform onto a cluster

This tutorial shows an operator how to install the OpenFrame platform (ArgoCD plus the app-of-apps) onto an existing Kubernetes cluster using the `openframe` command-line tool, or how to create a cluster and install the platform in one step.

## Prerequisites

- You have the `openframe` CLI binary available on your machine. For Windows, download the AMD64 build from `https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip` and run the installer the same way you would on other operating systems. For macOS/Linux, download the matching build from the [Releases page](https://github.com/flamingo-stack/openframe-cli/releases) and place it on your `$PATH`, or install it from source with `go install github.com/flamingo-stack/openframe-cli@latest`.
- Your machine meets the hardware requirements the platform needs to run comfortably:

  | Resource | Minimum | Recommended |
  |---|---|---|
  | RAM | 24 GB | 32 GB |
  | CPU Cores | 6 | 12 |
  | Disk Space | 50 GB | 100 GB |

- You already have a reachable Kubernetes cluster — one created with `openframe cluster create`, made by you directly, or any other cluster you can reach — unless you plan to use `openframe bootstrap`, which creates a local cluster for you as part of the same step.
- Docker is installed and running if you intend to let the CLI create a local cluster (k3d-backed).
- Helm is available to the CLI; the install flow regenerates certificates and installs Helm releases automatically, but the underlying tools (Docker, k3d, Helm, Terraform, gcloud, or the AWS CLI, depending on your target) must be present. Run `openframe prerequisites check` first to confirm what is missing.
- You know which kube-context you want to install into if you are not going to let the CLI prompt you interactively. Check your current contexts with your usual kubeconfig tooling before you begin.

## Steps

1. Open a terminal and confirm the CLI is on your `$PATH` by running `openframe` with no arguments; it should print its help text and command list instead of a "command not found" error.
2. Check what the installer needs by running `openframe prerequisites check` (optionally with `--type k3d`, `--type eks`, or `--type gke` to match your target), and resolve anything it reports as missing with `openframe prerequisites install`.
3. Decide how you will get a cluster:
   - If you do not yet have a cluster, run `openframe bootstrap [cluster-name]` to create a local k3d cluster and install the full OpenFrame platform (ArgoCD and the app-of-apps) in one step. Add `--non-interactive` for an unattended/CI run, or `--verbose` to see detailed ArgoCD sync progress.
   - If you already have a cluster (local or cloud), continue with `openframe app install` below.
4. If you already have a cluster, run `openframe app install` to install ArgoCD and the app-of-apps onto it:
   - Run `openframe app install` with no arguments for interactive mode: the CLI prompts you to pick a kube-context, warns you if the selected cluster is smaller than the recommended ~6 cores / 24 GB, and proceeds once you confirm.
   - Run `openframe app install my-cluster` to target a specific cluster by name.
   - Add `--context <name>` to target an exact kube-context directly, skipping interactive selection.
   - Add `--non-interactive` to skip all prompts and install using an existing `openframe-helm-values.yaml` file (for CI/CD).
   - Add `--ref <branch-or-tag>` (for example `--ref develop` or `--ref 1.0.48`) to deploy a specific branch or release tag of the platform instead of the default branch.
   - Add `--dry-run` to see what would be installed without executing anything.
   - Add `--force` to force installation even if charts already exist.
5. Watch the terminal while the install runs: it installs ArgoCD first, then the app-of-apps from the GitHub repository, then waits for every ArgoCD Application it creates to become ready. On an interactive terminal this appears as a live "Installing OpenFrame applications" panel with a progress bar and a count such as `14/17 ready`, alongside the names and statuses of applications still not ready.
6. Once the command finishes, check what was installed by running `openframe app status` (add `--watch` to keep watching, or `--interactive` to open the interactive status view).
7. Retrieve ArgoCD access by running `openframe app access`, which prints the ArgoCD admin credentials and the port-forward command needed to reach the ArgoCD UI in your browser.

## Expected result

- A successful `openframe app install` (or the install stage of `openframe bootstrap`) ends with every ArgoCD application reported as ready, shown during the run as a progress bar reaching `N/N ready` and finishing with a success line (optionally followed by which applications took longest to become ready).
- Running `openframe app status` afterward reports the cluster as reachable and each ArgoCD application's sync and health status.
- Running `openframe app access` prints the ArgoCD admin credentials and the exact port-forward command to reach the ArgoCD UI.

## Troubleshooting

- **"`<command>` takes no positional arguments (got `<args>`) — use `--context` to target a cluster"**: you passed a stray positional argument to a command (such as `app status` or `app uninstall`) that only accepts a cluster via `--context`. Remove the extra argument and pass `--context <name>` instead.
- **"could not use context `<name>`"**: the `--context` value you passed does not match a usable kube-context. Verify the context name against your kubeconfig and retry.
- **"Cluster `<name>` is smaller than recommended (~6 cores / 24 GB RAM). Continuing anyway."**: this is a warning, not a failure — the interactive install detected your cluster is below the recommended size but will continue. Expect slower installs or instability on severely undersized clusters.
- **Applications stuck as Degraded with a message that "will not recover on its own"**: the CLI has detected a workload that has stayed Degraded and synced for longer than its recovery window, and it stops waiting rather than running out the full timeout. The error names the affected application and recommends inspecting it with `kubectl get pods -n <namespace>` and `kubectl describe application <name> -n argocd`.
- **An application reports it "cannot render its manifests from the deployed revision" and that the error "is deterministic — retrying or waiting cannot fix it"**: the `--ref` you deployed likely predates the chart layout this CLI expects. Use a different branch or release tag whose chart matches this CLI's expectations.
- **The command exits with a specific non-zero code instead of a generic failure**: the CLI preserves the exit code of whatever external tool (Helm, Terraform, k3d, etc.) failed underneath it, so the number itself is a clue to which external command failed; re-run with `--verbose` for more detail.
- The material does not show a documented command for recovering an ArgoCD repo-server stuck in `CrashLoopBackOff` beyond what the CLI retries automatically — if the install still fails after a retry, inspect ArgoCD pods directly with `kubectl` as the error messages suggest.

The platform's application code (the app-of-apps chart and the services it deploys) lives in the separate `flamingo-stack/openframe-oss-tenant` repository, which this CLI clones at install time; this tutorial only covers running the `openframe` CLI against it.
