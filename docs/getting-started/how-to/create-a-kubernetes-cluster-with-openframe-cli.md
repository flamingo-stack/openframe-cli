# How to create a Kubernetes cluster with OpenFrame CLI

A platform operator uses the OpenFrame CLI to provision a new Kubernetes cluster — locally with k3d, or in the cloud on GKE or EKS — as the first step before installing the OpenFrame platform onto it.

## Prerequisites

Check each of these before you start:

- You have a terminal with the `openframe` CLI binary installed and on your `$PATH`. If you have not installed it yet, that is covered by a separate document, not this one.
- For a **local cluster (k3d, the default)**: Docker is installed and running. The CLI can install k3d itself the first time you run `cluster create`, so you do not need to install k3d by hand.
- For a **GKE cluster**: you have (or can log in to) a Google Cloud identity with access to the target GCP project, and you know the project ID and region you want to use.
- For an **EKS cluster**: you have (or can configure) an AWS identity (default credentials or a named profile) with access to the target account, and you know the region you want to use.
- You understand that cloud clusters (GKE/EKS) are provisioned with Terraform and create real cloud resources that incur cost — the CLI shows a cost warning, and an estimate when `infracost` is available, before anything is created.
- You have enough free terminal space to read an interactive wizard, or you already know the flags you want to pass for a non-interactive run.

## Steps

1. Open a terminal and run `openframe cluster create` with no arguments to start a local k3d cluster using the interactive selection menu, or run `openframe cluster create my-cluster` to pre-fill the cluster name in that same menu.
2. At the selection menu, press **Enter** to accept "Quick start with defaults" if you want the CLI to create the cluster immediately with default settings, or choose the interactive configuration wizard option to step through cluster customization (name, type, node count, and — for GKE/EKS — region, project, profile, high availability, and spot instances).
3. To skip the menu entirely and create a cluster directly from flags (for example in CI or scripting), run `openframe cluster create [NAME] --skip-wizard` together with the flags you need, such as `--type`, `--nodes`, `--region`, `--project`, `--profile`, `--ha`, or `--spot`. The type defaults to `k3d` when `--type` is omitted.
4. To create a local cluster with a specific node count without the wizard, run a command such as `openframe cluster create my-cluster --nodes 3 --type k3d --skip-wizard`.
5. To create a cloud cluster on Google Kubernetes Engine, run a command such as `openframe cluster create my-gke --type gke --project my-project --region us-central1 --skip-wizard`.
6. To create a cloud cluster on Amazon EKS, run a command such as `openframe cluster create my-eks --type eks --region us-east-1 --skip-wizard` (add `--profile staging` if you need a named AWS profile instead of your default credentials).
7. Before a cloud cluster is created, respond to the prompts the CLI shows you: for GKE it offers to run `gcloud auth login` (and the Application Default Credentials login) if you are not authenticated yet; for EKS it confirms the AWS identity/account it is about to use before continuing.
8. If you want to preview what would be created without creating anything, add `--dry-run` to your command (for example `openframe cluster create my-gke --type gke --project my-project --region us-central1 --dry-run`). For GKE/EKS this runs a real Terraform plan and shows the resource footprint and, when `infracost` is available, an estimated monthly cost — without provisioning anything.
9. When asked to confirm a cloud create (the "terraform apply" confirmation step), review the plan shown and confirm to proceed, or decline to cancel without applying any changes.
10. Wait for the command to finish. For k3d this is fast; for GKE/EKS this runs a full Terraform apply and can take noticeably longer while cloud infrastructure is provisioned.
11. After the cluster is ready, run `openframe cluster status [NAME]` (optionally with `-o json` or `-o yaml`) to confirm the cluster's health, or `openframe cluster list` (add `--all` to also discover external GKE/EKS clusters) to see it listed.
12. Once the cluster exists, proceed to install the OpenFrame platform onto it — covered in a separate document — rather than as part of cluster creation itself.

## Expected result

- For a local k3d cluster, the create command finishes and the cluster is usable immediately; `openframe cluster list` shows it, and `openframe cluster status [NAME]` reports its health.
- For a GKE or EKS cluster, after you confirm the Terraform apply, the command finishes by merging the new cluster into your default kubeconfig and switching the current `kubectl` context to it, so the cluster is reachable right away without a separate `kubectl config use-context` step.
- If you ran with `--dry-run`, no cluster is created; instead you see a resource footprint ("Plan: no changes — the cluster already matches this configuration" when nothing would change) and, if `infracost` is available, an estimated monthly cost line.
- If you re-run `cluster create` with the same name after a cluster already exists, the existing cluster is left untouched and reused rather than recreated — delete it first if you want to start from scratch.

## Troubleshooting

- **"node count must be at least 1: `<N>`"** — you passed `--nodes` with `0` or a negative number in non-interactive mode; pass a positive node count.
- **"--min-nodes must be at least 1 (a zero-node floor is not supported; omit the flag for the default of 1)"** or the equivalent message for **`--max-nodes`** — you explicitly set the flag to `0`; omit the flag to use the default instead of passing zero.
- **"GCP project `<project>` is not among your accessible projects"** — the `--project` value for a GKE create does not match a project your gcloud identity can access; the message lists the accessible projects (or tells you to run `gcloud projects list` when there are many) so you can pick a valid one.
- **"AWS `<profile or default credentials>` cannot authenticate (aws sts get-caller-identity failed)"** — your AWS credentials or named `--profile` are not usable; fix your AWS configuration (or the profile) and re-run.
- **"cluster '`<name>`' already exists in region/project '`<region/project>`' but is not managed by openframe — refusing to touch it; pick another cluster name"** — a cluster with that name already exists in the cloud account/project outside of OpenFrame's tracking; choose a different cluster name.
- **"kubeconfig context '`<name>`' already exists and points at `<server>` — refusing to overwrite it"** — you already have an unrelated kubeconfig context with the same name; rename the existing context or pick another cluster name.
- **A failed or interrupted cloud create** — re-run the same `cluster create` command to resume: the CLI keeps the Terraform workspace and state for the cluster and resumes from where it left off, refreshed against your current flags. The failure message also gives you the exact path to the kept Terraform state and the `openframe cluster delete [NAME]` command if you want to tear it down instead of resuming.
- **A create that collides with an already-existing cloud resource terraform is not tracking** (shown as a remediation message naming the resource) — this happens when a previous create was interrupted after the cloud resource was created but before it was recorded; resolve it by either deleting the orphaned resource in the cloud console/CLI or importing it into the cluster's Terraform state as instructed in the message, then re-run create.
- **"cluster creation cancelled — no changes were applied"** — you declined the apply confirmation prompt; re-run `cluster create` and confirm if you want to proceed.
- **terraform not installed** — for a `--dry-run` on a cloud cluster, the CLI skips the plan preview with a message noting that terraform installs automatically on a real (non-dry-run) create; the material does not show a manual terraform installation step beyond this automatic install.
