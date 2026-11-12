# How to create a Kubernetes cluster with OpenFrame CLI

This guide shows the person installing OpenFrame (an administrator or operator at a terminal) how to use the `openframe` command to provision a new Kubernetes cluster — a local `k3d` cluster, or a cloud GKE or EKS cluster — ready for the OpenFrame platform to be installed onto it.

## Prerequisites

- The OpenFrame CLI binary is installed and runs as `openframe` from a terminal (check with `openframe --help`).
- For a local cluster: Docker is installed and running, since the `k3d` backend provisions the cluster inside Docker.
- For a cloud cluster:
  - On GKE: you have (or can log into) a Google Cloud account and know the GCP project and region you want to use.
  - On EKS: you have (or can configure) AWS credentials/profile and know the AWS region you want to use.
  - Terraform is required for cloud clusters; the CLI installs it automatically as part of the type-aware prerequisite check if it is missing.
- You know what you want to name the cluster, or are prepared to accept the default name `openframe-dev`.
- You understand that a GKE or EKS cluster creates real cloud resources that incur costs, and that these costs are only estimated, not charged, during a dry run.

## Steps

1. Open a terminal and confirm the CLI is on your `$PATH` by running `openframe --help`.
2. Run `openframe cluster create` with no arguments to see the cluster creation selection menu, which offers **Quick start with defaults** (press Enter) or the **interactive configuration wizard**.
3. To create a cluster with a specific name, run `openframe cluster create my-cluster` instead, replacing `my-cluster` with the name you want; the CLI validates the name before continuing.
4. If you want full control over cluster type, provider, and sizing without the menu, run `openframe cluster create` with `--skip-wizard` and the relevant flags, for example:
   - `openframe cluster create --nodes 3 --type k3d --skip-wizard` for a local cluster.
   - `openframe cluster create my-gke --type gke --project my-project --region us-central1 --skip-wizard` for a GKE cluster.
   - `openframe cluster create my-eks --type eks --region us-east-1 --skip-wizard` for an EKS cluster (add `--profile staging` to use a named AWS profile).
5. If you chose a cloud cluster type (GKE or EKS), review the configuration summary the CLI prints when `--skip-wizard` or `--dry-run` is used, which shows the settings before anything is created.
6. Optionally, add `--dry-run` to any cloud cluster command first to preview the Terraform plan and an estimated monthly cost without creating anything; re-run the same command without `--dry-run` once you are satisfied.
7. For a GKE cluster, complete any `gcloud auth login` (and application-default login) prompt the CLI offers, so Terraform can authenticate against your Google Cloud project.
8. For an EKS cluster, confirm the AWS identity the CLI announces (the account `aws sts get-caller-identity` resolves) before it proceeds, so you are provisioning in the expected AWS account.
9. Wait while the CLI runs the prerequisite checks for the chosen cluster type, then provisions the cluster — for `k3d` this happens through Docker; for `gke`/`eks` this runs Terraform `init`, `plan`, and `apply` against the generated Terraform workspace.
10. If prompted to approve the Terraform apply plan during a cloud cluster create, review the plan summary shown and confirm to proceed.
11. Once creation finishes, run `openframe cluster status my-cluster` (or `openframe cluster status` for the default name) to confirm the cluster is healthy and ready for the next step, installing the OpenFrame platform.

## Expected result

- For `openframe cluster create` run with no arguments, you see a selection menu offering a quick-start default option and an interactive wizard option.
- For a non-interactive create (`--skip-wizard`), the CLI prints a configuration summary showing the cluster type, name, node count, and (for cloud types) region/project before creating anything.
- For a cloud cluster create with `--dry-run`, the CLI prints a Terraform plan diff and, where available, an estimated monthly cost (through `infracost`); it reports "Plan: no changes — the cluster already matches this configuration" if nothing would change.
- After a successful create, the cluster's context is merged into your kubeconfig and set as the current context automatically, and `openframe cluster status` reports the cluster as healthy.
- If a cluster with the chosen name already exists, the CLI leaves it untouched and reuses it rather than recreating it.

## Troubleshooting

- **"node count must be at least 1: `<n>`"** — you explicitly passed `--nodes` with a value of zero or less; omit the flag (default is 3) or pass a positive integer.
- **"--min-nodes must be at least 1 (a zero-node floor is not supported; omit the flag for the default of 1)"** or **"--max-nodes must be at least 1 (omit the flag for the default of 4)"** — you explicitly set one of these flags to `0`; omit the flag instead to use the default.
- **"GCP project `<project>` is not among your accessible projects"** (or, with more than 15 accessible projects, a message pointing you to run `gcloud projects list`) — the `--project` value for a GKE create does not match a project your current gcloud identity can access; pick one of the listed projects or re-authenticate.
- **"cluster '`<name>`' already exists in region/project '`<region-or-project>`' but is not managed by openframe — refusing to touch it; pick another cluster name"** — a cluster with that name already exists in the cloud account outside of OpenFrame's tracking; choose a different name.
- **"AWS `<profile/identity>` cannot authenticate (aws sts get-caller-identity failed)"** — your AWS credentials or named profile are not usable; fix your AWS CLI configuration and re-run.
- **"gcloud is not authenticated (run 'gcloud auth login' and 'gcloud auth application-default login')"** — complete both gcloud login steps before retrying.
- **"kubeconfig context '`<name>`' already exists and points at `<server>` — refusing to overwrite it; rename the existing context or pick another cluster name"** — an unrelated kubeconfig context already uses that cluster name; rename it or choose a different cluster name for the new create.
- **A cloud create fails partway through** — the CLI reports that the Terraform state is kept in the cluster's workspace directory, and that re-running `openframe cluster create` with the same name resumes the attempt, or `openframe cluster delete` tears it down; if the failure message mentions a resource that "already exists" in AWS or GCP but is not tracked by Terraform, follow the CLI's remediation steps (delete the orphaned resource in the cloud console, or import it into the Terraform state as shown in the message) before retrying.
- **Terraform is not installed** — for a `--dry-run` on a cloud cluster the CLI simply skips the plan preview and tells you so; for a real (non-dry-run) create, the CLI installs Terraform automatically as part of its prerequisite check.
- If none of these messages match what you see, the material available does not document that specific failure; check the exact error text printed by the CLI for further guidance.
