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
  UI. (`app` was previously `chart`; `chart` remains a hidden alias for
  backward compatibility.)
- `openframe prerequisites check|install [cluster|app]` — the prerequisite
  checks/installs as first-class commands.
- `openframe update` — self-update of the CLI binary (checksum + cosign verified,
  with `check` and `rollback`); see D6-adjacent tooling in
  `internal/shared/selfupdate`.

---

## D4 — `bootstrap` is a thin orchestrator

`openframe bootstrap [name] [--non-interactive] [--verbose]` stays as a single,
beginner-friendly command. Internally it only orchestrates:

```
prerequisites → cluster create → app install
```

It contains no business logic of its own — everything lives in the primitives.
`openframe bootstrap --non-interactive` reuses the existing `openframe-helm-values.yaml`
for the OSS tenant deployment.

---

## D5 — Cluster providers behind a unified interface

Cluster creation goes through a `Provider` interface parameterized by
**provider** (k3d local now; GKE/EKS later) and **target** (local vs cloud).
Only **k3d** is implemented; cloud providers return a clear "coming soon"
message. No new providers are added now — the interface exists so they can be
added later without touching the rest of the CLI.

For OSS the target is always **local** (k3d). SaaS targets (cloud) are future work.

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

## Platform support

- **macOS / Linux** — full support; prerequisites are checked and auto-installed.
- **Windows** — prerequisites are not auto-installed; the CLI prints a link to
  the documentation describing what to install and how (WSL2, Docker, etc.).

The primary audience is non-technical and semi-technical users, so every
interactive flow uses plain-language prompts, safe defaults, and confirmations
rather than raw errors.

---

## Target layout

```
cmd/
  cluster/         create, delete, list, status, cleanup
  app/             install, upgrade, status, access, uninstall   (alias: chart)
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
