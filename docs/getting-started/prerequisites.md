
# Prerequisites

This page lists what you need **before** installing and running the OpenFrame CLI (`openframe`), the Go-based command-line tool that provisions Kubernetes clusters (k3d, GKE, EKS) and deploys the OpenFrame platform onto them via ArgoCD's app-of-apps pattern.

> The OpenFrame platform application code (the services the CLI deploys) lives in a separate repository, [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant). This CLI clones that repository's Helm chart at install time — it is not vendored here. See `docs/reference/architecture/overview.md` for how the two repositories relate.

## What the CLI is built on

Per `go.mod`, the CLI is a single Go module (`github.com/flamingo-stack/openframe-cli`, Go `1.26.0`, toolchain `1.26.5`). It is distributed as a standalone binary via GitHub Releases, so **you do not need Go installed to use it** — only to build it from source. Per the organization's code graph, this repository has **no upstream dependencies and no downstream consumers** among internally-published artifacts: it is a self-contained CLI, not a library another repository imports.

## Required software and versions

| Tool | Why it's needed | Minimum / pinned version | Required for |
|---|---|---|---|
| Go | Build the CLI from source (`go build`, `go install`) | `1.26.5` (toolchain pinned in `go.mod`; `1.26.0`/`1.26.1` have a known Windows GC crash) | Building from source only — not needed to run a released binary |
| Docker | Backing engine for local k3d clusters | Any current Docker Desktop / Docker Engine with a running daemon | `openframe cluster create` (k3d), `openframe bootstrap` |
| k3d | Local Kubernetes cluster provisioning | Auto-downloaded and checksum-verified by the CLI itself (see `internal/shared/download/pins.go`) | `openframe cluster create --type k3d`, `openframe bootstrap` |
| Helm | Installs ArgoCD and the app-of-apps chart | Auto-downloaded and checksum-verified by the CLI | `openframe app install`, `openframe app upgrade` |
| Terraform | Provisions GKE/EKS via the generated root modules | `>= 1.15.0` (required by `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`); auto-downloaded by the CLI | `openframe cluster create --type eks\|gke` |
| gcloud CLI | GCP authentication and project/API checks for GKE | Current version, authenticated | `openframe cluster create --type gke`, GKE discovery |
| AWS CLI | AWS authentication for EKS | Current version, configured with credentials | `openframe cluster create --type eks`, EKS discovery |
| infracost | Optional cost estimate shown before a cloud `terraform apply` | Auto-downloaded and checksum-verified by the CLI | Optional, shown during `openframe cluster create` for EKS/GKE |
| kubectl context tooling (via `client-go`) | The CLI talks to the Kubernetes API natively — no separate `kubectl` binary required | n/a (built in) | All cluster-status and app operations |

> Docker, k3d, Helm, Terraform, and infracost binaries are pinned to exact versions with SHA256 digests in `internal/shared/download/pins.go` and verified by `internal/shared/download/verify.go` before use — never fetched via an unverified `curl \| bash`. Run `openframe prerequisites check` to see exactly which of these your machine is missing.

## System requirements

The CLI itself is a lightweight Go binary, but the local k3d cluster and the OpenFrame platform workloads it deploys (ArgoCD plus the app-of-apps services) need real headroom on the host running Docker.

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU Cores | 6 | 12 |
| Disk Space | 50 GB | 100 GB |

Supported operating systems:

- **macOS** and **Linux** — the CLI runs natively; `internal/prerequisites/runner.go` auto-installs missing tools directly.
- **Windows** — per `docs/reference/architecture/overview.md`, the native Windows build re-executes itself inside **WSL2** (`internal/shared/wsllauncher`) before any cluster command runs, because the local cluster (Docker + k3d) and the native Kubernetes client (`client-go`) need to share a network namespace that a native Windows process cannot reach reliably. On Windows, `openframe prerequisites install` prints manual installation links rather than auto-installing (prerequisites are auto-installed only on macOS/Linux, per `internal/prerequisites/runner.go`).

Download the Windows build directly if you are not building from source:

```text
https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
```

## Account / access requirements

Depending on which cluster backend you target, you need:

| Target | Access required |
|---|---|
| k3d (local) | Docker daemon access on the local machine (no cloud account needed) |
| GKE | A Google Cloud project with the Compute Engine and Kubernetes Engine APIs enabled, and `gcloud` authenticated with permissions to create VPCs, subnets, Cloud NAT, and a GKE cluster (see `internal/cluster/providers/gke/templates/main.tf`) |
| EKS | An AWS account/identity with permissions to create a VPC, an EKS cluster, a managed node group, and IAM roles (see `internal/cluster/providers/eks/templates/main.tf`); AWS CLI credentials configured |
| OpenFrame platform install | Network access to clone the `flamingo-stack/openframe-oss-tenant` repository (public, no credentials required per `internal/chart/utils/config/builder.go`) |
| `openframe update` | Network access to GitHub Releases; update binaries are verified via Sigstore/cosign keyless signatures (`internal/shared/selfupdate/cosign.go`), no GitHub account needed to consume releases |

> Docker registry credentials and ngrok credentials (for external ingress) are **optional**, collected only if you choose "Input custom Docker credentials" or "Use ngrok for External visibility" during `openframe app install`'s interactive configuration wizard (`internal/chart/ui/configuration/docker.go`, `internal/chart/ui/configuration/ingress.go`). If you need ngrok, you will be prompted to register at `https://dashboard.ngrok.com` and obtain an API key and auth token — refer to your own ngrok account for these values; this project does not issue them.

For AWS cost estimation, the EKS root module (`internal/cluster/providers/eks/templates/main.tf`) defaults to the Free-Tier-eligible `m7i-flex.large` instance type specifically so a new AWS account's Free plan does not reject the node group.

## Environment variables

Most environment variables used by the CLI are set automatically by the release/signing pipeline and are **not** something an end user needs to configure to run `openframe`. They are documented here for completeness and for anyone building/signing releases from source:

| Variable | Used by | Purpose |
|---|---|---|
| `OPENFRAME_SIGN` | `scripts/sign-binary.sh` | Must be `1` to trigger binary signing; unset or any other value makes the script a no-op, so local builds never attempt to sign |
| `SIGNING_IDENTITY`, `KEYCHAIN_PATH`, `APPLE_ID_USERNAME`, `APPLE_ID_PASSWORD`, `APPLE_TEAM_ID` | `scripts/sign-binary.sh` (macOS) | codesign + notarytool signing/notarization of macOS binaries during release |
| `JSIGN_JAR`, `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SIGNING_ENDPOINT`, `AZURE_CODE_SIGNING_ACCOUNT_NAME`, `AZURE_CERTIFICATE_PROFILE_NAME` | `scripts/sign-binary.sh` (Windows) | Azure Trusted Signing (Authenticode) of Windows binaries during release |
| `CGO_ENABLED` | `Makefile` (`build`, `build-all`, `test-race`) | `0` for reproducible cross-compiled builds; `1` only when running `make test-race` (the race detector requires cgo) |

None of these need to be set for a normal end-user install or for day-to-day `openframe` usage — they apply to the release pipeline that builds and signs the binaries you download.

## Verification commands

Run these before your first `openframe` command to confirm your machine is ready.

Confirm Go is available (only needed if building from source):

```bash
go version
```

Build the CLI for your current platform (output goes to `build/`):

```bash
make build
```

Or install directly via Go:

```bash
go install github.com/flamingo-stack/openframe-cli@latest
```

Check that the resulting binary runs and report its version:

```bash
openframe --version
```

Let the CLI check all tool prerequisites for your target cluster type without installing or changing anything:

```bash
openframe prerequisites check
```

Narrow the check to a specific cluster backend:

```bash
openframe prerequisites check --type k3d
openframe prerequisites check --type gke
openframe prerequisites check --type eks
```

Install any missing prerequisites (auto-installs on macOS/Linux; prints manual links on Windows):

```bash
openframe prerequisites install
```

Confirm Docker is reachable (required for k3d clusters):

```bash
docker info
```

Confirm cloud CLIs are authenticated, if you plan to use GKE or EKS:

```bash
gcloud auth list
aws sts get-caller-identity
```

Once prerequisites are satisfied, verify end-to-end readiness with a local bootstrap:

```bash
openframe bootstrap
```

This single command creates a local k3d cluster and installs the full OpenFrame platform (ArgoCD + app-of-apps), exercising the Docker, k3d, and Helm prerequisites together.

## Next

With prerequisites verified, continue to the quick start guide, or jump directly into the how-to guide for creating a new Kubernetes cluster with the OpenFrame CLI.
