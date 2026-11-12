## Overview

Before you install or build the OpenFrame CLI, make sure your machine has the right tooling in place. The CLI (`openframe`) is a Go-based binary whose job is to provision a Kubernetes cluster (k3d locally, or EKS/GKE in the cloud via Terraform) and then deploy the [OpenFrame platform](https://openframe.ai) onto it — the platform's own source lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository, which this CLI clones and installs via ArgoCD's app-of-apps pattern at install/upgrade time.

Per the organization's dependency graph (`get_repo_ecosystem`, coverage: full), this repository has no recorded upstream dependencies on other internal repositories and no recorded downstream consumers — it publishes a single Go module (`github.com/flamingo-stack/openframe-cli`) and is a standalone, user-facing binary rather than a library other services import.

> For the full command reference and how these pieces fit together, see [`../reference/architecture/overview.md`](../reference/architecture/overview.md) and [`../reference/architecture/ecosystem.md`](../reference/architecture/ecosystem.md).

## Required Software and Versions

| Software | Minimum Version | Why it's needed | Verify with |
|---|---|---|---|
| Go | 1.26.0 (toolchain pinned to `1.26.5` in `go.mod`) | Building the CLI from source; `go.mod` pins the toolchain because Go 1.26.0/1.26.1 have a known crash on windows-amd64 (fixed in 1.26.2) | `go version` |
| Docker | Recent stable release | Required by the local k3d cluster provider (`internal/cluster/prerequisites/docker/docker.go`); k3d nodes run as Docker containers | `docker version` |
| k3d | Pinned version (checked/installed automatically by the CLI) | Creates and manages local Kubernetes clusters (`internal/cluster/providers/k3d`) | `k3d version` |
| Helm | Pinned version (checked/installed automatically by the CLI) | Installs ArgoCD and the app-of-apps chart (`internal/chart/providers/helm`) | `helm version` |
| Terraform | `>= 1.15.0` | Provisions EKS/GKE clusters (`internal/cluster/providers/terraform`, `internal/cluster/providers/eks`, `internal/cluster/providers/gke`) | `terraform version` |
| kubectl / kubeconfig access | Any recent version compatible with your cluster | Reading cluster health and contexts (`internal/k8s`) | `kubectl version --client` |
| `aws` CLI | Recent stable release | Required only for EKS clusters (`internal/cluster/prerequisites/aws/aws.go`, `cmd/cluster/aws_identity.go`) | `aws --version` |
| `gcloud` CLI | Recent stable release | Required only for GKE clusters (`internal/cluster/prerequisites/gcloud/gcloud.go`) | `gcloud --version` |
| infracost | Pinned version (checked/installed automatically by the CLI) | Cost estimation before cloud cluster creation (`internal/cluster/prerequisites/infracost`) | `infracost --version` |
| golangci-lint | Recent stable release (contributors only) | Linting (`make lint` in the `Makefile`) | `golangci-lint --version` |

> Several of these — k3d, Helm, Terraform, and infracost — are **pinned and checksum-verified** by the CLI itself at `internal/shared/download/pins.go` rather than installed via `curl | bash`. Running `openframe prerequisites check` (or letting `openframe cluster create` / `openframe bootstrap` run their preflight) tells you exactly what is missing and installs what it can.

### Notes on Go toolchain

`go.mod` specifies:

```text
module github.com/flamingo-stack/openframe-cli

go 1.26.0
toolchain go1.26.5
```

The `toolchain` line exists specifically to keep released Windows binaries off the buggy 1.26.0/1.26.1 runtimes (a GC stack-scanning bug that corrupts return addresses at random on windows-amd64). If you build from source, let Go resolve this toolchain automatically rather than pinning an older one.

## System Requirements

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU cores | 6 | 12 |
| Disk space | 50 GB | 100 GB |

These figures account for running a local k3d cluster alongside the full OpenFrame platform stack (the platform itself — Kafka, MongoDB, Cassandra, and related services — is provisioned onto the cluster by the CLI, not run by the CLI process itself). Cloud cluster creation (EKS/GKE) shifts most compute load off your local machine, but Docker, Terraform, and the CLI's own operations still need headroom locally.

### Operating systems

- **macOS** and **Linux** are supported natively — Docker, k3d, and `client-go` run directly.
- **Windows**: the native Windows build forwards execution into **WSL2** (`internal/shared/wsllauncher`), because Docker, k3d, and `client-go` do not work reliably on native Windows. Install and enable WSL2 before running cluster-creation commands on Windows.

## Account / Access Requirements

Depending on which cluster backend you use, you'll need different credentials:

| Cluster type | Access required |
|---|---|
| Local (k3d) | A working local Docker daemon — no cloud account needed |
| AWS (EKS) | An AWS account and configured credentials/profile the `aws` CLI can use (`cmd/cluster/aws_identity.go` verifies the active identity before provisioning) |
| GCP (GKE) | A GCP project and `gcloud` authenticated session (`internal/cluster/discovery/gke.go`, `internal/cluster/prerequisites/gcloud`) |

For installing the OpenFrame platform itself onto a cluster, `openframe app install` clones the app-of-apps chart repository from the external [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository (referenced as `RepoOSSTenant` in `internal/chart/models/repos.go`) — this is a public repository clone, not an authenticated API dependency, but network access to GitHub is required.

If you intend to use a private Docker registry with your deployment, the interactive configuration wizard (`internal/chart/ui/configuration/docker.go`) will prompt for registry username, password/token, and email — collected only when you explicitly choose "Input custom Docker credentials"; otherwise no registry credentials are required.

> Refer to your environment configuration for any specific account credentials your organization requires — this document does not define or assume any default usernames or passwords.

## Environment Variables

| Variable | Purpose | Required? |
|---|---|---|
| `KUBECONFIG` | Overrides the default kubeconfig file location; respected when merging/removing cluster contexts (see `internal/cluster/providers/eks/kubeconfig.go`'s `mergeIntoDefaultKubeconfig`) | Optional — defaults to the standard kubeconfig path if unset |
| `OPENFRAME_SIGN` | Enables binary signing during release builds (`scripts/sign-binary.sh`); only relevant to maintainers cutting a release, never to end users or local builds | Optional, maintainers only |

Cloud-provider credentials (AWS access keys/profiles, GCP application-default credentials) are resolved through the standard `aws` and `gcloud` CLI mechanisms rather than OpenFrame-specific environment variables — configure those tools per their own documentation before running cloud cluster commands.

> Docker registry credentials and Ngrok tokens entered through the interactive configuration wizard are **never** read from environment variables; they are collected via masked terminal prompts and immediately registered with the CLI's secret redactor (`internal/shared/redact`) so they never appear in verbose logs.

## Verification Commands

Run these checks before your first `openframe` command:

```bash
# Go toolchain (only needed if building from source)
go version

# Docker (required for local k3d clusters)
docker version

# Kubernetes CLI (optional but useful for manual inspection)
kubectl version --client

# Cloud CLIs (only if you plan to create EKS or GKE clusters)
aws --version
gcloud --version
```

The CLI also ships its own built-in readiness check, which is the most reliable way to confirm you're ready — it inspects the current OS/architecture and reports exactly which required tools are missing or outdated:

```bash
# Check prerequisites for the default (k3d) cluster type
openframe prerequisites check

# Check prerequisites for a specific cluster type
openframe prerequisites check --type k3d
openframe prerequisites check --type eks
openframe prerequisites check --type gke

# Attempt to install missing prerequisites automatically
openframe prerequisites install --type k3d
```

This maps to the `internal/prerequisites/runner.go` and `internal/cluster/prerequisites/checker.go` logic, which is OS-aware and shared across the `cluster`, `app`, and `prerequisites` command groups — the same check also runs automatically as a preflight step inside `openframe bootstrap` and `openframe cluster create`.

## Building From Source (Optional)

If you are contributing to the CLI rather than just using a released binary, confirm your build toolchain with:

```bash
# Build the binary for your current platform into build/
make build

# Run unit tests (vet on, excludes integration tests)
make test-unit

# Run go vet across all packages
make vet
```

`make test-integration` and `make test-tfvalidate` have additional prerequisites (a real Docker + k3d environment, and network access to download Terraform providers, respectively) and are opt-in via Go build tags — they are not required for basic CLI usage.

## Next Steps

Once your environment passes `openframe prerequisites check`, continue to the quick-start guide to create your first cluster and install the platform.
