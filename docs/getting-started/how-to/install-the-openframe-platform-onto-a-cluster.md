# How to install the OpenFrame platform onto a cluster

This guide walks a platform operator through deploying the OpenFrame platform (ArgoCD plus the app-of-apps) onto a Kubernetes cluster using the `openframe` CLI, once a cluster already exists and `kubectl`/Helm can reach it.

## Prerequisites

Before you start, check that:

- The OpenFrame CLI binary (`openframe`) is installed and runs. You can download it from the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest) for your platform — for example, on Windows, download `openframe-cli_windows_amd64.zip`.
- You have a running, reachable Kubernetes cluster. This can be one you created with `openframe cluster create`, or any other cluster you can already reach with `kubectl` — the CLI does not require the cluster to have been created by it.
- Your machine meets the recommended capacity for the cluster you are installing into: roughly 6 CPU cores and 24GB RAM minimum, with 12 CPU cores and 32GB RAM recommended for a comfortable run. The CLI only warns if the cluster is smaller than recommended — it does not block the install.
- Docker, Helm, and (if you provisioned the cluster yourself) k3d, Terraform, or your cloud CLI (`gcloud`/`aws`) are available, since the `openframe app install` flow shells out to Helm during installation. Run `openframe prerequisites check` first if you are unsure what is missing.
- You know which kube-context points at the target cluster, or you are prepared to pick it interactively during the install.

## Steps

1. Open a terminal and confirm the CLI is on your `$PATH` by running `openframe --help`; this also shows the top-level command groups (`bootstrap`, `cluster`, `app`, `prerequisites`, `update`).
2. If you have not already, verify required tools are present by running `openframe prerequisites check` (add `--type k3d`, `--type eks`, or `--type gke` to match how your cluster was provisioned).
3. Start the installation by running `openframe app install`, optionally naming the cluster, for example `openframe app install my-cluster`. Add `--context <name>` to target a specific kube-context directly and skip the interactive cluster picker.
4. If you ran the command with no cluster name, `--non-interactive`, or `--dry-run`, and you were not prompted with `--context`, select the target kube-context from the interactive prompt that appears. The CLI reports the context you are installing into, for example "Installing OpenFrame into context ..." before proceeding.
5. If the CLI warns that the selected cluster is smaller than the recommended size (around 6 cores / 24GB RAM), decide whether to continue; the warning does not stop the install.
6. Wait while the CLI installs ArgoCD onto the cluster with its bundled configuration values.
7. Wait while the CLI clones the app-of-apps chart repository and installs it via Helm. By default this deploys the default platform branch; pass `--ref <branch-or-tag>` (for example `--ref 1.0.48`) to deploy a specific branch or release tag instead.
8. Wait while the CLI waits for all ArgoCD Applications to report a healthy, synced status before it finishes.
9. For a scripted or CI run instead of an interactive one, run `openframe app install --non-interactive`, which skips all prompts and uses the existing `openframe-helm-values.yaml` file in place of the interactive configuration wizard.
10. For a one-shot setup that both creates a local cluster and installs the platform, run `openframe bootstrap` instead of separate `cluster create` and `app install` commands.
11. After installation completes, check platform readiness with `openframe app status` (add `--watch` to keep monitoring, or `--interactive` for the dashboard view), and retrieve ArgoCD admin access details with `openframe app access`.

## Expected result

- `openframe --help` lists the `app` command group among the available commands.
- During an interactive install, the CLI prints a line naming the kube-context it is installing into, such as "Installing OpenFrame into context ...".
- If the cluster is under the recommended size, the CLI prints a warning naming the cluster context and the recommended minimum (around 6 cores / 24GB RAM), then continues anyway.
- When installation finishes successfully, `openframe app install` returns without an error, having installed ArgoCD and the app-of-apps and waited for all ArgoCD Applications to become ready.
- Running `openframe app status` afterward reports platform readiness — a table of ArgoCD applications and their sync/health status, plus a readiness summary.
- Running `openframe app access` prints ArgoCD admin credentials and instructions for reaching the ArgoCD UI.

## Troubleshooting

- **"<command> takes no positional arguments (got ...) — use --context to target a cluster"**: you passed a positional cluster name to a subcommand in the `app` group that does not accept one. Re-run the command using `--context <name>` instead to name the target cluster.
- **A `--ref` value that does not exist as a branch or tag in the chart repository**: the CLI validates the git ref before touching the cluster and fails fast with an actionable message rather than installing ArgoCD first. Correct the `--ref` value (the chart repository uses tags like `1.0.48`, without a `v` prefix) and re-run `openframe app install`.
- **Installation fails while waiting for ArgoCD Applications to become ready**: this occurs after ArgoCD and app-of-apps are already installed, so the CLI does not automatically retry (retrying would reinstall unnecessarily). Check platform status with `openframe app status` to see which applications are unhealthy, and investigate from there; the material does not show an automated recovery path for this failure beyond re-running status checks.
- **Required tools are missing (Docker, Helm, k3d, Terraform, cloud CLIs)**: run `openframe prerequisites check` or `openframe prerequisites install` with the appropriate `--type` flag to verify or install what the install flow needs before retrying `openframe app install`.

For creating a cluster first, see [Create a Kubernetes cluster with OpenFrame CLI](./create-a-kubernetes-cluster-with-openframe-cli.md). To verify the result afterward, see [Check the status of your OpenFrame cluster and app](./check-the-status-of-your-openframe-cluster-and-app.md).
