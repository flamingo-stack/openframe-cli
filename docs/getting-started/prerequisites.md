# Prerequisites

Before installing or building the OpenFrame CLI, make sure your machine satisfies the requirements below. The CLI (binary name `openframe`) provisions Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and installs the OpenFrame platform onto them via ArgoCD's app-of-apps pattern — so its own prerequisites fall into two layers: what you need to **run the CLI itself**, and what the CLI **checks for and can auto-install** before cluster/app operations.

> This document covers the `flamingo-stack/openframe-cli` repository. The OpenFrame platform that the CLI deploys (the application stack itself) lives in the separate `flamingo-stack/openframe-oss-tenant` repository — refer to its own documentation at `https://github.com/flamingo-stack/openframe-oss-tenant/tree/main/docs` for platform-level (not CLI-level) requirements.

## Required Software and Versions

| Tool | Why it's needed | Required by | Notes |
|---|---|---|---|
| Go | Build the CLI from source | Building from source | `go.mod` pins `go 1.26.0` with `toolchain go1.26.5` — go1.26.0/1.26.1 crash intermittently on windows-amd64 (`golang/go#77975`); go1.26.2+ fixes it |
| Git | Clone the repository; `go-git` is also vendored for app-of-apps repo cloning at runtime | Building from source, `app install` | CLI uses `github.com/go-git/go-git/v5` internally for cloning the app-of-apps chart repository |
| Docker | Backing runtime for local k3d clusters | `cluster create` (k3d type) | Checked by the CLI's cluster prerequisites gate before provisioning a k3d cluster |
| k3d | Creates/manages local Kubernetes clusters | `cluster create` (k3d type), `bootstrap` | Downloaded and checksum-verified by the CLI itself via its pinned-tool downloader if missing |
| Helm | Installs ArgoCD and the app-of-apps chart | `app install`, `app upgrade`, `bootstrap` | CLI shells out to the Helm CLI (`internal/chart/providers/helm`); also pinned/verified via the downloader |
| Terraform | Provisions cloud clusters (EKS/GKE) | `cluster create --type eks\|gke` | Driven via `terraform-exec`/`terraform-json`; root modules require Terraform `>= 1.15.0` (see `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`) |
| AWS CLI | AWS authentication/identity checks for EKS | `cluster create --type eks` | Checked under `internal/cluster/prerequisites/aws` |
| gcloud CLI | GCP authentication for GKE | `cluster create --type gke` | Checked under `internal/cluster/prerequisites/gcloud` |
| infracost (optional) | Cost preview before provisioning cloud clusters | `cluster create` (cloud types) | Checked under `internal/cluster/prerequisites/infracost` |
| mkcert / certificates tooling | Local TLS certificate generation | `app install` (localhost ingress) | Managed by `internal/chart/prerequisites/certificates` |

> All platform CLI tools (k3d, Helm, Terraform, mkcert) that the CLI downloads itself are checksum-verified and version-pinned — see `internal/shared/download/pins.go` and `internal/shared/download/verify.go`. You do not need to pre-install these; the CLI's `openframe prerequisites install` can fetch them for you on macOS/Linux.

## System Requirements

Because the CLI's primary workflow (`openframe bootstrap` / `openframe cluster create`) stands up a full Kubernetes cluster and the OpenFrame platform stack locally via Docker and k3d, your machine needs resources sized for that workload, not just for running a small Go binary:

| Tier | RAM | CPU cores | Disk |
|---|---|---|---|
| Minimum | 24 GB | 6 cores | 50 GB free |
| Recommended | 32 GB | 12 cores | 100 GB free |

**Supported operating systems:**
- **Linux** — native support, `cluster prerequisites install` can auto-install missing tools.
- **macOS** — native support (Intel and Apple Silicon), auto-install supported.
- **Windows** — supported via the native `openframe` binary, which forwards cluster operations into WSL2 for Docker/k3d access (see `internal/shared/wsllauncher`). Prerequisite checks on Windows are documentation-only (the CLI reports what's missing rather than auto-installing); WSL2 with a Linux distribution is required for local k3d clusters.

> Cloud cluster types (EKS, GKE) do not run Kubernetes locally, so the memory/CPU pressure is much lower — but you still need Terraform, cloud CLIs, and valid cloud credentials. Disk usage for Terraform workspaces and provider plugins is modest compared to a local cluster.

## Account / Access Requirements

| Requirement | Needed for | Details |
|---|---|---|
| GitHub access to `flamingo-stack/openframe-cli` releases | Installing a pre-built binary | Releases are published at `https://github.com/flamingo-stack/openframe-cli/releases` |
| Docker daemon running and reachable | Local k3d clusters | The CLI checks Docker availability as part of its k3d prerequisite set (`internal/cluster/prerequisites/docker`) |
| AWS account + credentials (IAM user/role with EKS, VPC, EC2, IAM permissions) | `cluster create --type eks` | The Terraform root module (`internal/cluster/providers/eks/templates/main.tf`) provisions a dedicated VPC, an EKS cluster, a managed node group, and the EBS CSI driver addon; your AWS identity is also granted cluster-admin via `enable_cluster_creator_admin_permissions` so kubeconfig access works immediately after creation |
| GCP project + credentials with Compute/Container Engine permissions | `cluster create --type gke` | The GKE Terraform root module (`internal/cluster/providers/gke/templates/main.tf`) provisions a VPC with secondary ranges, Cloud NAT, and a private-nodes GKE cluster; required project APIs (compute, container) must be enabled before Terraform runs |
| Docker registry credentials (optional) | `app install` with a private image registry | Collected interactively by `internal/chart/ui/configuration/docker.go`'s `DockerConfigurator`; values are masked on input and registered with the secret redactor so they never leak into logs |
| Ngrok account (optional) | `app install` with external ingress | Required only if you choose Ngrok-based ingress during the interactive wizard (`internal/chart/ui/configuration/ingress.go`); sign up, create a domain, and generate an API key + auth token at the Ngrok dashboard before running the wizard |
| OpenMSP Slack community access (optional) | Getting help, reporting issues | This project does not use GitHub Issues or Discussions — community support happens on the OpenMSP Slack (`https://www.openmsp.ai/`) |

> Never commit AWS, GCP, Docker registry, or Ngrok credentials to source control or helm values files tracked in git. Refer to your environment configuration for how secrets should be supplied in your setup.

## Environment Variables

| Variable | Purpose | Used by |
|---|---|---|
| `$KUBECONFIG` | Points to the kubeconfig file the CLI merges cluster contexts into (or reads from) | EKS/GKE kubeconfig merge logic (`internal/cluster/providers/eks/kubeconfig.go`), `internal/k8s` context resolution |
| `OPENFRAME_SIGN` | Gates whether `scripts/sign-binary.sh` performs binary signing | Release builds only (`scripts/sign-binary.sh`) — irrelevant to normal CLI usage, relevant only if you build release artifacts yourself |

Additional environment variables are consumed only inside the release signing pipeline (`scripts/sign-binary.sh`) and are not needed for day-to-day CLI usage: Apple notarization credentials (`APPLE_ID_USERNAME`, `APPLE_ID_PASSWORD`, `APPLE_TEAM_ID`, plus `SIGNING_IDENTITY`/`KEYCHAIN_PATH` set up earlier in that pipeline) for macOS, and Azure Trusted Signing credentials (`AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_SIGNING_ENDPOINT`, `AZURE_CODE_SIGNING_ACCOUNT_NAME`, `AZURE_CERTIFICATE_PROFILE_NAME`, plus `JSIGN_JAR`) for Windows. These only matter if you are maintaining the release pipeline itself, not if you are installing or using the CLI.

No environment variables are required simply to run `openframe` commands interactively — cluster targeting is normally done via the `--context` flag or the interactive cluster selector, and Helm values are read from `openframe-helm-values.yaml` in the working directory by default (`internal/chart/utils/config/paths.go`).

## Verification Commands

Run these checks before using the CLI, substituting the tool names relevant to the cluster type you intend to use:

```bash
# Go toolchain (only needed if building from source)
go version

# Git
git --version

# Docker (required for local k3d clusters)
docker version
docker info

# k3d (if pre-installed; otherwise the CLI can install it)
k3d version

# Helm (if pre-installed; otherwise the CLI can install it)
helm version

# Terraform (required for EKS/GKE cluster types)
terraform version

# AWS CLI and credentials (EKS only)
aws --version
aws sts get-caller-identity

# gcloud CLI and credentials (GKE only)
gcloud --version
gcloud auth list
```

Once the CLI binary itself is installed, let it tell you what else it needs:

```bash
# Check prerequisites for the local k3d path
openframe prerequisites check --type k3d

# Check prerequisites for a cloud cluster type
openframe prerequisites check --type eks
openframe prerequisites check --type gke

# Auto-install missing tools (macOS/Linux only; Windows reports guidance instead)
openframe prerequisites install --type k3d

# Confirm the CLI itself is installed and see pinned dependency versions
openframe --version
```

```text
Exit codes for `openframe prerequisites check` follow the CLI's general
exit-code-fidelity contract: 0 means every required tool for the selected
--type was found and is usable; a non-zero code means at least one check
failed, and the failing tool(s) are named in the output.
```

### Windows-specific verification

On Windows, the native `openframe` binary forwards cluster operations into WSL2. Verify WSL2 and a Linux distribution are set up before running cluster commands:

```bash
wsl --status
wsl --list --verbose
```

If Docker and k3d are missing inside your WSL2 distribution, the Windows build of `openframe` surfaces documentation-only guidance (rather than auto-installing) — follow the printed instructions to install them inside WSL2.

Once your environment passes the checks above, continue to the quick-start guide to create your first cluster.
