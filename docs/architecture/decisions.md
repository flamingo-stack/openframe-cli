# Architecture Decisions

This document records the key design decisions for the OpenFrame CLI restructure.
It is the authoritative reference for how the CLI is organized and why.

Status: **accepted** · Last updated: 2026-06-24

---

## Context

OpenFrame CLI is used by operators and semi-technical users to stand up OpenFrame
on Kubernetes. The primary supported path is **OSS** (a local cluster, no
credentials). SaaS modes come later. The CLI is being restructured into three
clearly isolated abstractions so each can be used on its own.

---

## D1 — Three isolated abstractions: cluster, app, prerequisites

The CLI is organized around three independent concerns:

- **cluster** — make a Kubernetes cluster (local now; cloud later).
- **app** — deploy the OpenFrame application (Helm chart → ArgoCD → apps) onto a
  cluster that already exists and is online.
- **prerequisites** — check and install the tools each of the above needs.

**Rule:** the `app` subsystem must not import cluster-creation code. It only
talks to a cluster through a small Kubernetes access API (list contexts, check
health, check resources). This lets a user who already has a cluster (their own,
or one made outside OpenFrame) install the app into it, and lets a user create a
cluster without installing anything.

---

## D2 — OSS-tenant is the only deployment

The CLI supports a single deployment: **oss-tenant**. The app is always installed
from the public `openframe-oss-tenant` chart repository, which requires no
credentials. There is no `--deployment-mode` flag; `--non-interactive` simply
reuses the existing `openframe-helm-values.yaml`.

| deployment   | chart repository                | credentials |
|--------------|---------------------------------|-------------|
| `oss-tenant` | `openframe-oss-tenant` (public) | none        |

The cluster is always a local k3d cluster.

---

## D3 — Commands: `cluster` and `app` are the two primitives

- `openframe cluster create|delete|list|status|cleanup` — cluster lifecycle.
  `create` **only creates the cluster**; it never installs the app. (Verb is
  `create`; there is no `apply`.) `cleanup` removes unused cluster resources.
- `openframe app install|upgrade|status|access|uninstall` — installs and operates
  the OpenFrame app on an existing, online cluster. `upgrade` re-deploys the
  app-of-apps at a new git ref (`--ref`) or forces an ArgoCD hard refresh + sync
  (`--sync`); `access` prints the ArgoCD admin credentials and how to open the
  UI. (`app` was previously named `chart`.)
- `openframe prerequisites check|install [cluster|app]` — the prerequisite
  checks/installs as first-class commands.
- `openframe update` — self-update of the CLI binary (checksum + cosign verified,
  with `check` and `rollback`); see D6-adjacent tooling in
  `internal/shared/selfupdate`.

---

## D4 — `bootstrap` is a thin orchestrator

`openframe bootstrap [name] [--non-interactive] [--verbose]` stays as a single,
beginner-friendly command. Internally it only orchestrates:

```text
prerequisites → cluster create → app install
```

It contains no business logic of its own — everything lives in the primitives.
`openframe bootstrap --non-interactive` reuses the existing `openframe-helm-values.yaml`
for the OSS tenant deployment.

---

## D5 — Cluster providers behind a unified interface

Cluster creation goes through a `Provider` interface with three backends:
**k3d** (local), **EKS**, and **GKE** (cloud). Backends are selected via the
`provider.New(type)` factory, keyed on `ClusterConfig.Type`; the rest of the
CLI never knows which backend runs. Cloud providers additionally implement
`Planner` (`--dry-run` renders a real `terraform plan` footprint).

The cloud backends share one terraform engine (D7/D8): each generates a
pinned, self-contained root module on the public `terraform-aws-modules` /
`terraform-google-modules` modules and drives `terraform` via terraform-exec.
Kubeconfig entries carry no static credentials — auth runs through the
provider CLI exec plugins (`aws eks get-token`, `gke-gcloud-auth-plugin`),
with the context named after the cluster so exact-match context resolution
works unchanged.

For OSS the default remains **local** (k3d); cloud clusters are an explicit
`--type eks|gke` opt-in with a cost warning and a typed-name confirmation on
delete.

---

## D6 — No dependency on the ArgoCD Go module (use the dynamic client)

ArgoCD is **not importable as a Go library** (its `go.mod` uses a local
`replace => ./gitops-engine`), which previously pinned the entire Kubernetes
*server* tree (`k8s.io/kubernetes`) into this CLI.

The CLI reads ArgoCD `Application` resources through the Kubernetes **dynamic
client** (unstructured, GVR `argoproj.io/v1alpha1 applications`) instead of the
typed argo-cd clientset. Benefits:

- **version-agnostic** — compatible with whatever ArgoCD version is deployed,
  including the latest;
- removes the largest supply-chain dependency;
- unblocks keeping `k8s.io/*` on the latest stable release.

---

## D7 — Terraform (BUSL) as the provisioning engine, installed verified

Cloud clusters are provisioned with **HashiCorp Terraform**, not OpenTofu.
BUSL 1.1 only restricts "hosted or embedded" offerings **competitive with
HashiCorp's products**; this CLI uses terraform as an internal tool to
provision the user's own infrastructure, which is not a competitive offering.
The binary is installed like every other prerequisite: a pinned version with
SHA256 verification into `~/.openframe/bin` (no curl-pipe-bash, no sudo). An
already-installed `terraform` on PATH in `~/.openframe/bin` is preferred.

If a server-side scenario ever provisions clusters *as a service* with
terraform, that is a different BUSL use profile and needs its own review.

## D8 — Local terraform state in per-cluster workspaces

Each cloud cluster owns a workspace under `~/.openframe/clusters/<name>/`:
the generated root module, `terraform.tfvars.json`, local state, and a
`cluster.json` registry record (type, status, endpoint/CA). The registry is
what makes cloud clusters visible to `list`/`status`/`delete` without cloud
API calls, and the state file is the only pointer to billed resources — so a
workspace is **never deleted on a failed apply**, only after a successful
destroy. Re-running `create` resumes an interrupted apply.

Remote state is opt-in via `--backend-config s3://bucket/prefix` (EKS) or
`gcs://bucket/prefix` (GKE) for users who need the state to survive the
machine that created the cluster.

## Platform support

- **macOS / Linux** — full support; prerequisites are checked and auto-installed.
- **Windows** — prerequisites are not auto-installed; the CLI prints a link to
  the documentation describing what to install and how (WSL2, Docker, etc.).

The primary audience is non-technical and semi-technical users, so every
interactive flow uses plain-language prompts, safe defaults, and confirmations
rather than raw errors.

---

## Target layout

```text
cmd/
  cluster/         create, delete, list, status, cleanup
  app/             install, upgrade, status, access, uninstall
  prerequisites/   check, install
  bootstrap/       orchestrator (prerequisites → cluster create → app install)
  update/          self-update: (update), check, rollback
internal/
  cluster/provider/   Provider interface + Target(local|cloud) + k3d impl
  cluster/            cluster lifecycle (service + k3d provider)
  chart/              helm/argocd/git providers + app-of-apps install
  k8s/                cluster-access API: contexts, rest.Config, health, resources
  prerequisites/      OS-aware checker/installer framework
  platform/           OS detection + Windows/WSL2 doc hints
  shared/             executor, errors, ui, redact, files, config, flags,
                      download (pinned tools), selfupdate, wsllauncher
docs/                 all documentation
```
