# Cloud Clusters (EKS / GKE)

Besides local k3d clusters, `openframe cluster create` can provision managed
Kubernetes clusters in AWS (EKS) or Google Cloud (GKE) using Terraform under
the hood. The CLI installs its own verified Terraform binary and generates the
infrastructure code for you — no Terraform knowledge required.

> **Cost warning.** Cloud clusters create billed resources: a managed control
> plane (~$73/month on both providers), VM nodes, and NAT/networking. The CLI
> shows this warning before creating and requires you to re-type the cluster
> name before deleting.

## Prerequisites

Checked and installed automatically on `cluster create`:

| Type | Tools | You provide |
|------|-------|-------------|
| eks  | terraform (pinned, verified), AWS CLI | working AWS credentials (`aws configure` or `--profile`) |
| gke  | terraform (pinned, verified), gcloud, gke-gcloud-auth-plugin | `gcloud auth login` + a GCP project |

Credentials are preflighted before anything is created (`aws sts
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

Provisioning takes ~10–20 minutes; the CLI streams per-resource progress.
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
