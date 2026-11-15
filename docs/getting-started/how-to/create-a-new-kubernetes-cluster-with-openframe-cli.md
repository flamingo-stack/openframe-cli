# How to create a new Kubernetes cluster with OpenFrame CLI

at# How to create a new Kubernetes cluster with OpenFrame CLI

A platform operator uses the OpenFrame CLI's `openframe cluster create` command to provision a new Kubernetes cluster — locally on this machine, or in the cloud on GKE or EKS — as the first step before installing the OpenFrame platform onto it.

## Prerequisites

- You have the `openframe` CLI binary installed and on your `$PATH`. Windows users can get it from `https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip`; macOS/Linux users download the matching build from the project's Releases page, or run `go install github.com/flamingo-stack/openframe-cli@latest`.
- Your machine meets the hardware requirements for running a cluster locally: minimum 24 GB RAM, 6 CPU cores, 50 GB disk; recommended 32 GB RAM, 12 CPU cores, 100 GB disk.
- For a **local cluster** (the default): Docker is installed and running. The CLI can check and install the remaining local-cluster tools (k3d, helm) for you as part of the create flow.
- For a **cloud cluster on GKE**: you can run `gcloud auth login`, and the account has access to the target GCP project.
- For a **cloud cluster on EKS**: `aws sts get-caller-identity` succeeds for the AWS profile you plan to use (or your default credentials), meaning the CLI can authenticate as you.
- You are prepared for cloud clusters (GKE/EKS) to create real, billable cloud resources (VPC, node groups, load balancers) — Terraform is used under the hood and is installed automatically if missing.
- On Windows, you are running inside WSL2, or you understand that the native `openframe` build re-executes itself inside WSL2 automatically before any cluster command runs.

## Steps

1. Open a terminal and run `openframe cluster create` to start creating a cluster. Add a name as an argument, for example `openframe cluster create my-cluster`, if you want a specific cluster name instead of the default.
2. At the selection menu, choose how you want to configure the cluster:
   - Press **Enter** to accept **Quick start with defaults**, which creates a local k3d cluster with default settings immediately, or
   - Select the **interactive configuration wizard** to step through cluster options (cluster type, name, node count, and — for cloud types — region, project, and machine settings) one screen at a time.
3. If you are creating a cloud cluster (GKE or EKS) and have not already authenticated, respond to the prompt the CLI shows you:
   - For GKE, confirm the offer to run `gcloud auth login` (and the Application Default Credentials login) when asked.
   - For EKS, review the AWS identity the CLI announces (from `aws sts get-caller-identity`) and confirm it is the account you want billed.
4. Review the configuration summary the CLI prints before it provisions anything — cluster type, name, node count, and, for cloud types, region/project — to confirm it matches what you intended.
5. For a cloud cluster, review the Terraform plan the CLI shows (the resources it is about to create) and, if `infracost` is available or you accept the CLI's offer to install it, review the estimated monthly cost line before continuing.
6. Confirm the apply when prompted, so the CLI provisions the cluster (the local k3d cluster, or the Terraform apply for GKE/EKS).
7. Wait for the command to finish. Do not close the terminal while cluster nodes are being created — a cloud cluster create in particular can take several minutes.
8. If you want to skip the menu and wizard entirely next time (for example in a script or CI), re-run the command non-interactively, for example `openframe cluster create my-cluster --skip-wizard`, optionally adding flags such as `--nodes`, `--type`, `--region`, `--project`, or `--profile` to set values directly instead of being prompted.

## Expected result

- On success, the CLI leaves your local kubectl configuration pointing at the new cluster: the new cluster's context becomes the current context, so your next `kubectl` or `openframe` command targets it automatically.
- For a quick-start local cluster, the command returns you to the prompt with the cluster created and ready to use.
- For a cloud cluster (GKE/EKS), the CLI reports the plan had changes, applies them, and once complete writes the cluster's endpoint and credentials into your kubeconfig; if the plan shows no differences from an existing cluster, the CLI reports "Plan: no changes — the cluster already matches this configuration" and does not create anything new.
- The cluster itself has no OpenFrame platform installed yet — only the Kubernetes cluster exists. Installing the platform is a separate, later task.

## Troubleshooting

- **"node count must be at least 1: `<N>`"** — you passed `--nodes` with a value of zero or less in non-interactive mode. Re-run with a positive node count.
- **"--min-nodes must be at least 1 (a zero-node floor is not supported; omit the flag for the default of 1)"** or the equivalent message for `--max-nodes` — you explicitly set one of these flags to `0`. Omit the flag to use the default, or set it to a positive number.
- **"cluster '`<name>`' already exists in region '`<region>`' but is not managed by openframe — refusing to touch it; pick another cluster name"** (EKS), or the equivalent project-scoped message for GKE — a cluster with that name already exists outside this CLI's tracking. Choose a different cluster name.
- **"AWS `<profile or default credentials>` cannot authenticate (aws sts get-caller-identity failed)"** — your AWS credentials are not usable. Re-check your AWS CLI configuration or the `--profile` you passed before retrying.
- **"gcloud is not authenticated (run 'gcloud auth login' and 'gcloud auth application-default login')"** — accept the CLI's offer to run `gcloud auth login` when prompted, or run that command yourself first.
- **"GCP project '`<project>`' is not accessible with the current gcloud identity"**, or "GCP project `<project>` is not among your accessible projects: ..." — the `--project` value does not match a project your gcloud identity can reach; rerun with a project from the list the CLI shows, or check access in the Google Cloud Console.
- **"kubeconfig context '`<name>`' already exists and points at `<server>` — refusing to overwrite it; rename the existing context or pick another cluster name"** — a kubeconfig entry with the same name as your new cluster already points somewhere else. Rename the existing context yourself or choose a different cluster name for the new one.
- **A cloud create fails partway through** — the CLI reports where its Terraform state is kept (under your cluster's workspace) and tells you to either re-run `openframe cluster create` with the same name to resume, or run `openframe cluster delete` to tear down what was created so far.
- **A create is interrupted and a later retry reports a resource that already exists in the cloud ("ResourceInUseException" / "AlreadyExists" for AWS, or an HTTP 409/"alreadyExists" for GCP)** — the CLI explains this is the signature of a create interrupted after a cloud resource was made but before Terraform recorded it, and gives you two ways to resolve it: delete the orphaned resource in the cloud console (or with the matching delete command), or import it into the cluster's Terraform state with the `terraform import` command the CLI prints, then re-run create.
- **Cost estimate unavailable** — if `infracost` cannot produce an estimate, the CLI tells you to run `infracost auth login` (free, one-time) and re-run; this does not block cluster creation.
