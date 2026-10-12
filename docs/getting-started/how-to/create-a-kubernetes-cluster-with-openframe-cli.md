# How to create a Kubernetes cluster with OpenFrame CLI

This guide shows the person setting up OpenFrame (a developer, operator, or evaluator) how to provision a Kubernetes cluster — locally with k3d, or in the cloud on GKE or EKS — using the `openframe cluster create` command.

## Prerequisites

- The OpenFrame CLI binary (`openframe`) is installed and runs when you type `openframe` in your terminal. See `../prerequisites.md` for setup details.
- For a **local cluster** (the default): Docker is installed and running on your machine. The CLI can install `k3d` and `helm` for you automatically when they are missing — it prompts before doing so.
- For a **cloud cluster**:
  - **GKE**: you have a Google Cloud project you can access, and `gcloud` is installed (the CLI can prompt you to run `gcloud auth login` and set up application-default credentials when needed).
  - **EKS**: you have working AWS credentials (a default profile or a named profile) that can pass `aws sts get-caller-identity`.
  - `terraform` is required for cloud clusters — the CLI installs it automatically if it's missing.
  - Cloud clusters create billed resources in your account. Review the pricing of the provider before proceeding.
- You are in a terminal with enough free resources on the machine running the cluster — for a local cluster, the hardware should have at least 24GB RAM, 6 CPU cores, and 50GB disk free (32GB RAM / 12 CPU cores / 100GB disk recommended).
- You know the name you want to give the cluster (or you're fine with the default name `openframe-dev`).

## Steps

1. Open a terminal and run `openframe cluster create` to start cluster creation. Add a name as an argument if you want one other than the default, for example `openframe cluster create my-cluster`.
2. At the selection menu that appears, choose how to configure the cluster:
   - Press Enter to accept **Quick start with defaults**, which creates the cluster immediately with default settings, or
   - Select the **interactive configuration wizard** option to step through cluster customization (cluster type, node count, Kubernetes version, and — for cloud types — region/project/machine type) one screen at a time.
3. If you want to skip the menu entirely and create the cluster directly from flags, run `openframe cluster create` with the `--skip-wizard` flag instead, combined with any of the following flags as needed:
   - `--nodes <N>` to set the number of nodes (default 3)
   - `--type k3d|gke|eks` to choose the backend (default `k3d`)
   - `--region <region>` and `--project <project>` for GKE, or `--region <region>` and `--profile <aws-profile>` for EKS
   - `--min-nodes` / `--max-nodes` to bound autoscaling for cloud clusters
   - `--spot` to use spot/preemptible instances on a cloud cluster
4. If you want to preview what would happen without creating anything, add `--dry-run` to the command. For a cloud cluster this runs a real Terraform plan and shows the resources it would create and (when available) an estimated monthly cost, without provisioning anything.
5. When prompted, confirm any tool installation the CLI offers (for example installing `k3d`, `helm`, or `terraform`) by answering the interactive yes/no prompt, or let it proceed automatically if you ran with `--skip-wizard`.
6. For a **GKE** cluster, if you are not already authenticated, follow the prompt offering to run `gcloud auth login` (and the application-default login) right from the CLI.
7. For an **EKS** cluster, confirm the AWS identity shown when the CLI asks you to verify which AWS account/profile will be billed before it provisions anything.
8. For a cloud cluster that isn't a dry run, review the Terraform plan diff and cost information the CLI prints, then confirm the apply when prompted (the CLI only provisions what you approve).
9. Wait for cluster creation to finish. A local k3d cluster finishes in under a minute; a cloud cluster (GKE/EKS) takes several minutes while Terraform provisions networking and node infrastructure.
10. After creation succeeds, run `openframe cluster status` (optionally `openframe cluster status <name>`) to confirm the cluster is healthy and check its nodes before installing the OpenFrame platform on it.

## Expected result

- On success, the CLI merges the new cluster's access into your default kubeconfig and switches your current `kubectl` context to it, so you don't need any extra configuration step to reach the cluster afterward.
- For a `--skip-wizard` or `--dry-run` run (or with `--verbose`), the CLI first prints a configuration summary showing the chosen cluster type, name, node count, and other settings before proceeding.
- Running `--dry-run` on a cloud cluster either prints `Plan: no changes — the cluster already matches this configuration`, or a plan diff listing the resources Terraform would create, followed by a cost estimate line such as `Estimated monthly cost (infracost): ...` when the `infracost` tool is available, or a cost warning with a link to the provider's pricing page when it is not.
- After a real (non-dry-run) create finishes, `openframe cluster status` reports the cluster as healthy with its node list — this is your signal the cluster is ready for `openframe app install` (see `./install-the-openframe-platform-onto-a-cluster.md`).

## Troubleshooting

- **"node count must be at least 1: `<N>`"** — you explicitly passed `--nodes` with a value of zero or a negative number; pass a positive node count or omit the flag to use the default of 3.
- **"--min-nodes must be at least 1 (a zero-node floor is not supported; omit the flag for the default of 1)"** or **"--max-nodes must be at least 1 (omit the flag for the default of 4)"** — you passed one of these flags explicitly set to `0`; either give it a positive value or remove the flag.
- **"cluster '`<name>`' already exists in region/project '`<x>`' but is not managed by openframe — refusing to touch it; pick another cluster name"** — a cluster with that name already exists in the target cloud account outside OpenFrame's tracking; choose a different name.
- **"AWS `<profile>` cannot authenticate (aws sts get-caller-identity failed)"** — your AWS credentials for that profile aren't valid; fix your AWS CLI configuration and re-run.
- **"gcloud is not authenticated (run 'gcloud auth login' and 'gcloud auth application-default login')"** — accept the CLI's offer to run these commands for you, or run them yourself first.
- **"GCP project `<project>` is not among your accessible projects"** — the `--project` value doesn't match a project your gcloud identity can access; the message lists the valid options (or tells you to run `gcloud projects list` when there are many).
- **"kubeconfig context '`<name>`' already exists and points at `<server>` — refusing to overwrite it; rename the existing context or pick another cluster name"** — a different cluster or tool already owns that context name in your kubeconfig; pick a different cluster name.
- **"cluster creation cancelled — no changes were applied"** — you answered no to the apply confirmation prompt; re-run the command and confirm when you're ready to proceed.
- If a cloud create is interrupted (for example with Ctrl+C) partway through, the CLI prints a hint that the Terraform state is kept locally and that you can simply re-run `openframe cluster create` with the same name to resume, or run `openframe cluster delete` to tear down what was created so far — see `./uninstall-openframe-and-delete-a-cluster.md`.
- If a resource already exists in the cloud provider but Terraform isn't tracking it (shown as a "ResourceInUseException"/"AlreadyExists" or GCP 409-style error), the CLI's message explains this happens when a create was interrupted after the cloud resource was made but before its state was saved, and gives you two options: delete the orphaned resource in the provider's console, or import it into the cluster's Terraform state with the `terraform import` command shown in the message.
