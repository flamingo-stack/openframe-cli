## Prerequisites

Before you install and run `openframe` — the CLI that provisions Kubernetes clusters and deploys the OpenFrame platform onto them — make sure your machine satisfies the requirements below. OpenFrame CLI is a Go service published as `github.com/flamingo-stack/openframe-cli` (per the organization's dependency graph it has no upstream dependencies on other internal repositories, and no other repository in the graph currently consumes it — it operates as a standalone, leaf tool in this ecosystem).

> **Note on scope:** OpenFrame CLI provisions infrastructure (clusters) and deploys the OpenFrame platform via ArgoCD app-of-apps. The platform's own source lives in the separate **openframe-oss-tenant** repository (https://github.com/flamingo-stack/openframe-oss-tenant) — the CLI clones its chart repository at install time rather than bundling it.

### Required software and versions

The CLI itself is a single compiled Go binary — end users do **not** need a Go toolchain installed to run it. The table below separates what you need to **build the CLI from source** from what the CLI **shells out to at runtime** depending on the cluster type you choose.

| Tool | Minimum / Pinned Version | Required For | Notes |
|---|---|---|---|
| Go | 1.26.0 (toolchain `1.26.5`) | Building from source only | Declared in `go.mod`. The comment in `go.mod` warns that Go 1.26.0/1.26.1 can crash randomly on `windows-amd64` during GC stack scanning — the pinned `toolchain` directive keeps release builds off those versions. |
| Docker | Any recent Engine/Desktop release | `openframe cluster create` with a local **k3d** cluster | k3d provisions cluster nodes as Docker containers; gated by `internal/cluster/prerequisites/docker`. |
| k3d | Pinned version (auto-downloaded and checksum-verified) | Local cluster provisioning | Managed by `internal/cluster/prerequisites/k3d` and `internal/shared/download` (pinned in `internal/shared/download/pins.go`); the CLI can install it for you rather than requiring a pre-existing install. |
| Helm | Pinned version (auto-downloaded and checksum-verified) | ArgoCD install, app-of-apps chart operations | Managed by `internal/cluster/prerequisites/helm` (cluster-side) and `internal/chart/prerequisites/helm` (chart-side). |
| Terraform | Pinned version (auto-downloaded and checksum-verified) | **EKS** or **GKE** cloud cluster provisioning | Managed by `internal/cluster/prerequisites/terraform`; generates and applies the root modules in `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`. Requires Terraform `>= 1.15.0` per those templates. |
| gcloud CLI | Recent stable release | **GKE** cluster provisioning and `cluster use` (switches gcloud config too) | Checked via `internal/cluster/prerequisites/gcloud`; used for GCP authentication and project API enablement ahead of the Terraform run. |
| AWS CLI | Recent stable release | **EKS** cluster provisioning | Checked via `internal/cluster/prerequisites/aws`; the EKS kubeconfig path uses `aws eks get-token` for short-lived, credential-free cluster authentication (see `internal/cluster/providers/eks/kubeconfig.go`). |
| infracost | Pinned version (auto-downloaded and checksum-verified) | Cost preview before cloud cluster creation | Managed by `internal/cluster/prerequisites/infracost`. |
| mkcert / certificates tooling | Pinned version (auto-downloaded and checksum-verified) | Local TLS certificates for platform ingress | Managed by `internal/chart/prerequisites/certificates`. |

> Where the table says "auto-downloaded and checksum-verified," the CLI's own prerequisite framework (`internal/prerequisites`, `internal/cluster/prerequisites`, `internal/chart/prerequisites`) checks for the tool and, on macOS/Linux, can install the pinned version itself via `internal/shared/download`. On Windows, the framework falls back to **docs-only guidance** rather than auto-installing (see `internal/platform`).

Run `openframe prerequisites check` (or `check --type k3d|eks|gke`) at any time to see exactly which of these your machine still needs — see Verification commands below.

### System requirements

The OpenFrame platform workloads that get deployed onto the cluster (not the CLI binary itself, which is lightweight) are the reason for these minimums:

| Tier | RAM | CPU cores | Disk |
|---|---|---|---|
| Minimum | 24 GB | 6 cores | 50 GB free |
| Recommended | 32 GB | 12 cores | 100 GB free |

These apply to the machine running the local k3d cluster (Docker's resource allocation must be raised to match — Docker Desktop's default resource limits are usually too low). For cloud clusters (EKS/GKE), the Terraform templates size the **node pool**, not your workstation: `internal/cluster/providers/eks/templates/main.tf` defaults to `m7i-flex.large` instances (chosen specifically because it is Free-Tier-eligible) with 1–4 autoscaled nodes (3 desired), and `internal/cluster/providers/gke/templates/main.tf` defaults to `e2-standard-4` with the same 1–4 range. Your workstation only needs enough headroom to run the CLI, Terraform, and the cloud provider CLI comfortably.

Supported operating systems: macOS, Linux, and Windows. On Windows, cluster-access operations may need to run through WSL2 — `internal/shared/wsllauncher` forwards the native Windows binary into WSL2 when required, and `internal/platform/wsl_hint.go` surfaces guidance when a WSL-specific issue is detected.

### Account / access requirements

Depending on which cluster type and ingress mode you plan to use, gather the following before you start:

- **Local (k3d):** No cloud account needed — only a working Docker daemon.
- **AWS (EKS):** An AWS account and credentials the AWS CLI can resolve (profile or environment credentials), with permission to create VPCs, EKS clusters, managed node groups, and IAM roles. `internal/cluster/discovery/eks.go` and `internal/cluster/discovery/auth.go` handle AWS identity resolution; `cmd/cluster/aws_identity.go` exposes an explicit identity-check command.
- **GCP (GKE):** A GCP project with billing enabled, and `gcloud` authenticated against an account with permission to enable compute/container APIs and create VPCs, subnets, Cloud Router/NAT, and GKE clusters. `internal/cluster/discovery/gke.go` handles GKE discovery and auth.
- **Git access:** Network access to clone the app-of-apps chart repository used by `internal/chart/providers/git` during `openframe app install` — this pulls deployment charts from an external repository rather than bundling them in the CLI.
- **Docker registry (optional):** If your deployment needs private image pull credentials, `internal/chart/ui/configuration/docker.go` interactively collects a registry username, password, and email during chart configuration. These are registered with the CLI's secret redactor (`internal/shared/redact`) so they never appear in verbose logs.
- **Ngrok account (optional):** Only required if you choose external ingress instead of localhost-only access. `internal/chart/ui/configuration/ingress.go` walks you through Ngrok domain, API key, and auth token collection, with links to the Ngrok dashboard for domain creation, API keys, and tunnel auth tokens; it also lets you configure a CIDR-based IP allowlist for the tunnel.
- **Code signing (release maintainers only):** The `scripts/sign-binary.sh` script used in release builds requires Apple Developer ID credentials (`APPLE_ID_USERNAME`, `APPLE_ID_PASSWORD`, `APPLE_TEAM_ID`, plus a configured keychain) for macOS notarization, and Azure Trusted Signing credentials (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SIGNING_ENDPOINT`, `AZURE_CODE_SIGNING_ACCOUNT_NAME`, `AZURE_CERTIFICATE_PROFILE_NAME`) for Windows Authenticode signing. This only applies if you are building and signing releases yourself; it is not needed to simply run the CLI.

### Environment variables

| Variable | Purpose | Required? |
|---|---|---|
| `$KUBECONFIG` | Points to the kubeconfig file the CLI reads/merges cluster contexts into. `internal/cluster/providers/eks/kubeconfig.go` explicitly honors `$KUBECONFIG` when merging an EKS cluster's context into your default kubeconfig, and refuses to overwrite an existing context of the same name pointing at a different server. | Optional — defaults to the standard kubeconfig location if unset. |
| `OPENFRAME_SIGN` | Gates whether `scripts/sign-binary.sh` performs any signing at all; the release workflow sets this to `1`. Local builds and CI compile checks leave it unset, making the script a no-op. | Only relevant to release maintainers. |
| `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SIGNING_ENDPOINT`, `AZURE_CODE_SIGNING_ACCOUNT_NAME`, `AZURE_CERTIFICATE_PROFILE_NAME` | Azure Trusted Signing credentials consumed by `scripts/sign-binary.sh` for Windows binary signing. | Only relevant to release maintainers (`OPENFRAME_SIGN=1`). |
| `APPLE_ID_USERNAME`, `APPLE_ID_PASSWORD`, `APPLE_TEAM_ID`, `SIGNING_IDENTITY`, `KEYCHAIN_PATH` | Apple notarization credentials consumed by `scripts/sign-binary.sh` for macOS binary signing/notarization. | Only relevant to release maintainers (`OPENFRAME_SIGN=1`). |

Day-to-day cluster and platform operations do not require you to set any environment variables yourself beyond an optional `$KUBECONFIG` override — the CLI resolves kube-context, cloud credentials, and tool paths through its own prerequisite and discovery logic (`internal/k8s`, `internal/cluster/discovery`).

### Verification commands

Run these from a terminal after installing the binary (see the platform-specific install steps in the introduction/quick-start material) to confirm your environment is ready:

```bash
# Confirm the CLI is installed and resolvable on PATH
openframe --help

# Check every prerequisite relevant to local (k3d) clusters
openframe prerequisites check

# Check prerequisites scoped to a specific cluster type
openframe prerequisites check --type k3d
openframe prerequisites check --type eks
openframe prerequisites check --type gke

# Let the CLI install any missing pinned tools itself (macOS/Linux)
openframe prerequisites install --type k3d
```

```bash
# Independently confirm the underlying tools the CLI shells out to
docker version
helm version
k3d version
terraform version
gcloud --version
aws --version
```

```bash
# Confirm kube-context resolution once a cluster exists
openframe cluster status
```

A clean `openframe prerequisites check` run (no missing-tool warnings) means you can proceed to cluster creation. If any check fails, re-run the equivalent `openframe prerequisites install --type <k3d|eks|gke>` on macOS/Linux, or follow the docs-only guidance the CLI prints on Windows before continuing.
