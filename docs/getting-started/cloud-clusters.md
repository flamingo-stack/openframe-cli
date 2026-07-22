# Cloud Clusters (EKS / GKE)

> Looking for a step-by-step walkthrough? See the
> [GKE Workflow](./gke-workflow.md) — this page is the reference.
> AWS EKS creation is currently gated behind a coming-soon banner; the
> sections below describe both providers for when it is enabled.

Besides local k3d clusters, `openframe cluster create` can provision managed
Kubernetes clusters in AWS (EKS) or Google Cloud (GKE) using Terraform under
the hood. The CLI installs its own verified Terraform binary and generates the
infrastructure code for you — no Terraform knowledge required.

> **Cost warning.** Cloud clusters create billed resources: a managed control
> plane, VM nodes, and NAT/networking. The CLI shows a warning with the
> provider's pricing page before creating — and, in an interactive
> `--dry-run` it offers to install [infracost](https://www.infracost.io)
> (verified pinned download; the one-time free `infracost auth login` is also offered in-CLI) and shows
> a monthly estimate — and requires you to re-type the cluster name
> before deleting. Pricing: [GKE](https://cloud.google.com/kubernetes-engine/pricing)
> · [EKS](https://aws.amazon.com/eks/pricing/).

## Prerequisites

Checked and installed automatically on `cluster create`:

| Type | Tools | You provide |
|------|-------|-------------|
| eks  | terraform (pinned, verified), AWS CLI | working AWS credentials (`aws configure` or `--profile`) |
| gke  | terraform (pinned, verified), gcloud, gke-gcloud-auth-plugin | `gcloud auth login` + a GCP project |

**You do not need to log in beforehand.** When a command needs Google Cloud
access (`create --type gke`, `list --all`, `use`), the CLI checks your gcloud
auth state and, in an interactive session, offers to run `gcloud auth login`
(and `gcloud auth application-default login` for Terraform) right there — one
flow, no manual steps. Non-interactive sessions (CI) never prompt and fail
with the exact command to run instead.

Credentials are additionally preflighted before anything is created (`aws sts
get-caller-identity` / `gcloud auth print-access-token`), so a broken login
fails in seconds, not mid-provisioning.

## Creating a cluster

Interactive (wizard asks for type, region, instance type):

```bash
openframe cluster create
```

Non-interactive:

```bash
# AWS EKS
openframe cluster create my-eks --type eks --region us-east-1 --skip-wizard

# Google GKE
openframe cluster create my-gke --type gke --project my-project --region us-central1 --skip-wizard
```

Useful flags: `--machine-type`, `--min-nodes` / `--max-nodes`, `--spot`,
`--profile` (AWS), `--nodes` (initial size), `--version` (`<major>.<minor>`,
e.g. `1.33`).

In interactive sessions the CLI first shows the full Terraform plan and asks
for approval (the `terraform apply` shape; what you approve is exactly what
runs — non-interactive sessions auto-approve). Provisioning then takes ~10–20
minutes; the CLI streams per-resource progress. GKE nodes are private (no
external IPs, egress via Cloud NAT) with a public control-plane endpoint, so
the flow works in organizations enforcing `restrict_vm_external_ips`.
When it finishes, your kubeconfig gets a context named after the cluster and
it becomes the current context — `kubectl get nodes` just works
(authentication runs through short-lived tokens via `aws eks get-token` /
`gke-gcloud-auth-plugin`; no static credentials are stored).

## Previewing without creating

`--dry-run` runs a real `terraform plan` and prints the resource footprint
without creating anything (and without registering the cluster):

```bash
openframe cluster create my-eks --type eks --region us-east-1 --skip-wizard --dry-run
# Plan: 47 to add, 0 to change, 0 to destroy
```

## Where the state lives

Each cloud cluster owns a workspace in `~/.openframe/clusters/<name>/`: the
generated Terraform module and the state file. The state is the only pointer
to your billed cloud resources — the workspace is never deleted on a failed
create, only after a successful delete.

- **A create failed or was interrupted?** Re-run the same `cluster create` —
  it resumes where it stopped.
- **Want the state to survive your machine?** Pass a remote backend at
  create time: `--backend-config s3://bucket/prefix` (EKS) or
  `--backend-config gcs://bucket/prefix` (GKE).

## Day-2 commands

```bash
openframe cluster list                # local + cloud clusters
openframe cluster list --all          # + external clusters discovered in your GCP projects
openframe cluster use my-gke          # switch kubectl context (and gcloud configuration)
openframe cluster status my-eks
openframe cluster delete my-eks       # terraform destroy; asks to re-type the name
openframe app install                 # install OpenFrame onto the current context
```

`cluster use` works for external (discovered) GKE clusters too: it fetches
credentials via gcloud when the kubeconfig has no entry yet, and activates
the gcloud configuration matching the cluster's project.

`cluster delete --force` skips the typed confirmation (for CI). `cluster
cleanup` does not apply to cloud clusters — use `delete`.

## Troubleshooting

- **"AWS ... cannot authenticate" / "gcloud is not authenticated"** — fix
  credentials (`aws configure`, `gcloud auth login`) and re-run; nothing was
  created.
- **Create failed mid-way** — the error names the workspace directory. Re-run
  `cluster create <name>` to resume, or `cluster delete <name>` to tear down
  what was partially created.
- **Verbose Terraform output** — add `--verbose` to stream Terraform's own
  logs during create/delete.
