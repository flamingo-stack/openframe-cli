# Cloud Clusters (EKS / GKE)

> Looking for a step-by-step walkthrough? See the
> [GKE Workflow](./gke-workflow.md) — this page is the reference.

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

**AWS identity is vetted before anything runs.** An EKS create (or its
`--dry-run`) first resolves which identity it is about to use — your
`--profile`, or the default credential chain — and shows the actual account
and ARN:

- *Interactive*: you must confirm it (`Use AWS profile 'staging' — account
  123456789012, arn:… for this EKS operation?`); declining aborts and lists
  your other configured profiles. The wizard additionally offers a picker of
  the profiles found in your AWS config.
- *Non-interactive (CI)*: never prompts — passing `--profile` (or having a
  working default chain) is the consent, and the account in use is printed so
  CI logs show whose account was billed.
- *Nothing configured at all* (no named profiles and the default chain cannot
  authenticate): the command fails with a pointer to the official AWS guide —
  [Configuration and credential file settings](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html).
  Set up credentials, then re-run (optionally with `--profile <name>`).

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

Useful flags: `--machine-type`, `--min-nodes` / `--max-nodes` (autoscaler
bounds; defaults 1 / 4, must be at least 1 — an explicit 0 is rejected),
`--spot` (spot-capacity nodes, typically 60–90% off the node cost — the cost
warning suggests it for test clusters), `--profile` (AWS), `--nodes` (initial
size), `--version` (`<major>.<minor>`, e.g. `1.33`), `--ha` (GKE: regional
control plane and nodes; the node count is then **per zone**, and every
summary shows the `N per zone × 3 zones` math).

In interactive sessions the CLI first shows the full Terraform plan and asks
for approval (the `terraform apply` shape; what you approve is exactly what
runs — non-interactive sessions auto-approve). Provisioning then takes ~10–20
minutes; the CLI streams per-resource progress. GKE nodes are private (no
external IPs, egress via Cloud NAT) with a public control-plane endpoint, so
the flow works in organizations enforcing `restrict_vm_external_ips`. EKS
clusters get a dedicated VPC (2 AZs, nodes in private subnets behind a single
NAT gateway), the core addons (`vpc-cni`, `kube-proxy`, `coredns`) and the
`aws-ebs-csi-driver` addon with a default gp3 StorageClass, so networking and
PersistentVolumeClaims work out of the box. The generated Terraform pins the
upstream EKS/VPC modules to exact versions — upstream default changes arrive
only with a deliberate CLI release, never mid-`create`. The default node type
is `m7i-flex.large`, which is Free-Tier-eligible: a brand-new AWS account (its
Free plan refuses non-eligible instance types) can run the documented flow
unchanged.
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

The preview authenticates like a real create (including the AWS identity
vetting above), but has classic `terraform plan` semantics: nothing is
written — not to the cloud, and not to the workspace. Over an existing
(failed/interrupted) workspace the preview shows what a **resume** would
actually apply: the module and variables are regenerated from the current CLI
and your current flags, planned against the workspace's saved state in a
throwaway directory.

## Where the state lives

Each cloud cluster owns a workspace in `~/.openframe/clusters/<name>/`: the
generated Terraform module, the state file, and a `terraform.log` that every
apply/destroy appends its full output stream to (so a long operation leaves a
record beyond the terminal). The state is the only pointer to your billed
cloud resources — the workspace is never deleted on a failed create, only
after a successful delete.

- **A create failed or was interrupted?** Re-run the same `cluster create` —
  it resumes where it stopped.
- **Want the state to survive your machine?** Pass a remote backend at
  create time: `--backend-config s3://bucket/prefix` (EKS) or
  `--backend-config gcs://bucket/prefix` (GKE).

## Day-2 commands

```bash
openframe cluster list                # local + cloud clusters
openframe cluster list --all          # + external clusters discovered in your GCP projects / AWS profiles
openframe cluster use my-gke          # switch kubectl context (and gcloud configuration)
openframe cluster status my-eks
openframe cluster delete my-eks       # terraform destroy; asks to re-type the name
openframe app install                 # install OpenFrame onto the current context
```

`cluster use` works for external (discovered) GKE clusters too: it fetches
credentials via gcloud when the kubeconfig has no entry yet, and activates
the gcloud configuration matching the cluster's project.

`cluster delete` tears down more than the terraform state: application
namespaces are removed first so PVC-backed disks/volumes are reclaimed while
the nodes still run, and anything that survives the destroy is swept up
afterwards — listed and deleted with your consent. `--force` skips the typed
confirmation and consents to that sweep (for CI). `cluster cleanup` does not
apply to cloud clusters — use `delete`.

## Troubleshooting

- **"no usable AWS configuration found"** — the default credential chain
  cannot authenticate and no named profiles exist. Set up credentials first
  (official guide: [Configuration and credential file settings](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)),
  then re-run — optionally with `--profile <name>`.
- **"AWS ... cannot authenticate" / "gcloud is not authenticated"** — fix
  credentials (`aws configure`, `gcloud auth login`) and re-run; nothing was
  created. For AWS the error lists your other configured profiles when the
  selected one is broken.
- **Create failed mid-way** — the error names the workspace directory. Re-run
  `cluster create <name>` to resume, or `cluster delete <name>` to tear down
  what was partially created.
- **Verbose Terraform output** — add `--verbose` to stream Terraform's own
  logs during create/delete. Either way, the full stream of every
  apply/destroy is appended to
  `~/.openframe/clusters/<name>/terraform/terraform.log`, and a failed
  operation names that path.
