# GKE Workflow — from zero to a running cluster

The complete, in-order flow for working with Google Kubernetes Engine through
the OpenFrame CLI. You need: the `openframe` binary, a Google account with
access to a GCP project, and a browser for the login. Everything else —
Terraform, gcloud, the auth plugin, credentials — the CLI sets up itself.

> **Costs.** A GKE cluster bills real money: a cluster management fee, VM
> nodes, and networking — see the
> [GKE pricing page](https://cloud.google.com/kubernetes-engine/pricing).
> In an interactive `--dry-run`, the CLI offers to install
> [infracost](https://www.infracost.io) (verified pinned download) and then
> shows a monthly estimate — even the one-time free
> `infracost auth login` is offered right inside the CLI. The CLI warns before creating and requires re-typing the
> cluster name before deleting.

## 0. (Optional) Preview what would be created

A dry-run computes a real `terraform plan` without creating anything and
without registering the cluster:

```bash
openframe cluster create my-gke --type gke \
  --project my-project --region us-central1 --skip-wizard --dry-run
#   +   module.network.module.vpc.google_compute_network.network
#   +   module.network.module.subnets.google_compute_subnetwork.subnetwork["us-central1/my-gke-subnet"]
#   +   module.gke.google_container_cluster.primary
#   +   module.gke.google_container_node_pool.pools["default"]
#   ...
# Plan: 11 to add, 0 to change, 0 to destroy
```

If Terraform is not installed yet, the preview is skipped with a note — it
installs automatically on a real create.

## 1. Create the cluster

One command; interactive wizard or flags.

**Wizard** (prompts: name → type `gke` → project → region → instance type →
node count → confirmation with a cost warning):

```bash
openframe cluster create
```

**Flags** (defaults: `e2-standard-4`, 3 nodes, latest GKE version):

```bash
openframe cluster create my-gke --type gke \
  --project my-project --region us-central1 --skip-wizard
```

Useful extras: `--machine-type e2-standard-8`, `--min-nodes 1 --max-nodes 6`,
`--spot` (spot-capacity nodes, typically 60–90% off the node cost — good for
test clusters), `--version 1.33`, `--nodes 4`, `--ha` (regional control plane
and nodes — note that a regional node count is **per zone**: `--nodes 3 --ha`
provisions 3 × 3 zones = 9 nodes, and the CLI displays exactly that math).

What happens, in order — no manual steps in between:

1. **Tools**: terraform (pinned, checksum-verified, into `~/.openframe/bin`),
   gcloud, and `gke-gcloud-auth-plugin` are checked and installed if missing.
2. **Login**: if gcloud is not authenticated, the CLI offers to run
   `gcloud auth login` right there (browser opens); then, because Terraform
   uses Application Default Credentials, it offers
   `gcloud auth application-default login` too. CI/non-interactive sessions
   never get prompts — they fail with the exact command to run.
3. **Preflight**: project access is verified; the required project APIs
   (Compute Engine, Kubernetes Engine) are enabled with one idempotent
   `gcloud services enable` — deliberately outside Terraform, so no cluster's
   state owns project-level toggles in a shared project; and the CLI refuses
   to proceed if a cluster with this name already exists in the project but
   was not created by openframe (it will never touch clusters it does not
   own).
4. **Plan & confirm** (interactive sessions): the full Terraform plan is
   shown — every resource to be created and the summary line — and you are
   asked to approve it before anything is applied, exactly like
   `terraform apply`. What you approve is what runs (the saved plan is
   applied, not a re-plan). Declining a brand-new create leaves no trace.
   Non-interactive sessions auto-approve, as before.
5. **Provision** (~10–15 min): a dedicated VPC with pod/service ranges, a
   Cloud NAT for egress, and a **zonal** GKE cluster (regional with `--ha`)
   with **private nodes** (no external IPs — compatible with orgs enforcing
   `restrict_vm_external_ips`) behind a public control-plane endpoint,
   streamed as per-resource progress lines. Add `--verbose` for raw
   Terraform output; the full stream is also appended to
   `~/.openframe/clusters/my-gke/terraform/terraform.log` either way.
6. **Kubeconfig**: a context named exactly `my-gke` is merged into your
   kubeconfig (existing contexts are never overwritten) and made current.

## 2. Verify

```bash
kubectl get nodes                    # exec-auth via gke-gcloud-auth-plugin
openframe cluster status my-gke
openframe cluster list
```

## 3. Install OpenFrame onto it

```bash
openframe app install                # targets the current kubectl context
openframe app status
openframe app access                 # ArgoCD URL + credentials
```

## 4. Day-2 operations

```bash
# See everything: local k3d + openframe-managed + EXTERNAL clusters
# discovered in the GCP projects of your gcloud configurations
openframe cluster list --all
# NAME              TYPE  SOURCE     STATUS   NODES  CONTEXT                        PROJECT              CREATED
# my-gke            gke   openframe  Ready    3      my-gke                         my-project           2026-07-21 14:02
# tenant-cluster-1  gke   external   Running  3      connectgateway_..._tenant-...  tenant-runners-db9z  —

# Switch kubectl (and the matching gcloud configuration) to any of them:
openframe cluster use my-gke
openframe cluster use tenant-cluster-1   # external: credentials are fetched
                                         # via gcloud if not present yet
```

External clusters are strictly **read-only** for openframe: they show up in
`list --all`/`status`/`use`, but `delete` and `cleanup` refuse them.

## 5. If a create fails or is interrupted

The workspace and Terraform state under `~/.openframe/clusters/my-gke/` are
kept — they are the only pointer to the billed resources. Two options:

```bash
openframe cluster create my-gke --type gke \
  --project my-project --region us-central1 --skip-wizard   # resumes
openframe cluster delete my-gke                             # tears down
```

Want the state to survive your machine? Create with
`--backend-config gcs://my-bucket/clusters/my-gke` (remote state in GCS).

## 6. Delete

```bash
openframe cluster delete my-gke
# → asks you to re-type "my-gke", then:
#   1. app namespaces are deleted first, so the CSI driver reclaims the
#      PVC-backed Persistent Disks while the nodes still run (those disks
#      live outside the terraform state and would otherwise survive as
#      billable orphans)
#   2. terraform destroy removes the cluster, node pool, and VPC
#   3. any disk that still survived is swept up: listed, and deleted with
#      your consent (interactive prompt, or automatically with --force)
#   4. the workspace and kubeconfig context are cleaned up
```

`--force` skips the typed confirmation and consents to the orphan-disk
cleanup (CI). `cluster cleanup` does not apply to cloud clusters — use
`delete`.

## Troubleshooting

| Symptom | What to do |
| --- | --- |
| "gcloud is not authenticated" (CI) | run `gcloud auth login` and `gcloud auth application-default login` in an interactive session |
| "project ... is not accessible" | check the project ID and your IAM role (`gcloud projects describe <project>`) |
| "already exists ... not managed by openframe" | the name is taken by a cluster openframe does not own — pick another name |
| "kubeconfig context ... refusing to overwrite" | a same-named context points elsewhere — rename it or pick another cluster name |
| create failed mid-way | re-run the same create to resume, or delete to tear down (state is never lost) |
| want Terraform's own logs | add `--verbose`, or read `~/.openframe/clusters/<name>/terraform/terraform.log` — every apply/destroy appends its full stream there |

See [Cloud Clusters](./cloud-clusters.md) for the reference (flags, state
model, EKS status) and `docs/architecture/decisions.md` (D5, D7, D8) for the
design rationale.
