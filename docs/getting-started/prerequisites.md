# Prerequisites

Before installing and using OpenFrame CLI, make sure your system meets the requirements below. OpenFrame CLI ships with a built-in `prerequisites` command that can check (and, on macOS/Linux, auto-install) most of these tools for you.

## Required Software

The exact tool set depends on which cluster type you plan to use (`k3d` for local, `eks` for AWS, `gke` for GCP).

| Tool | Required for | Notes |
|---|---|---|
| Docker | `k3d` (local clusters) | Must be **installed and running** — the daemon, not just the CLI |
| k3d | `k3d` (local clusters) | Kubernetes-in-Docker; auto-installed via Homebrew (macOS) or a pinned, checksum-verified binary (Linux) |
| Helm | `k3d`, `eks`, `gke` | Used to install ArgoCD and the app-of-apps chart |
| Terraform | `eks`, `gke` (cloud clusters) | Minimum version `>= 1.15.0`; always installed as a pinned, SHA256-verified binary (not via package managers) |
| AWS CLI | `eks` | Required for identity/credential resolution against AWS |
| gcloud CLI + `gke-gcloud-auth-plugin` | `gke` | Required for GCP authentication and `kubectl` credential plugin support |
| infracost (optional) | cloud cluster cost estimates | Optional; if missing, a generic pricing hint is shown instead |

> On **Windows**, native auto-install is not supported for local (`k3d`) prerequisites — the CLI forwards k3d/Docker-based cluster operations into WSL2 (Ubuntu). Cloud cluster provisioning (EKS/GKE via Terraform) works natively on Windows.

## System Requirements

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU Cores | 6 | 12 |
| Disk Space | 50 GB | 100 GB |

> These figures reflect running a full local OpenFrame platform install (ArgoCD + app-of-apps) inside a k3d cluster on your machine. Cloud cluster deployments (EKS/GKE) shift most resource consumption to the cloud provider, but the CLI host still needs enough local resources to run Docker, Terraform, and Helm operations.

## Account / Access Requirements

| Cluster type | Access needed |
|---|---|
| `k3d` (local) | None — runs entirely on your local Docker daemon |
| `eks` | An AWS account with credentials configured (via the AWS CLI / environment) and permissions to create EKS clusters, VPCs, and related IAM resources |
| `gke` | A GCP project with billing enabled and permissions to create GKE clusters and related networking resources |

## Environment Variables

| Variable | Purpose | Required |
|---|---|---|
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Escape hatch to bypass cosign signature verification during self-update | No — not recommended, emergency use only |
| `OPENFRAME_AUTO_UPDATE` | Set to `1` to opt in to automatic daily update checks (skipped in CI/non-interactive shells) | No |

Cloud credentials themselves (AWS/GCP) are picked up from your existing AWS CLI / gcloud CLI configuration rather than dedicated OpenFrame-specific environment variables.

## Verification Commands

Once OpenFrame CLI is installed, verify your environment is ready before creating a cluster:

```bash
# Check prerequisites for the default (local k3d) cluster type
openframe prerequisites check

# Check prerequisites for an EKS cluster
openframe prerequisites check --type eks

# Check prerequisites for a GKE cluster
openframe prerequisites check --type gke

# Auto-install missing tools where supported (macOS/Linux)
openframe prerequisites install --type k3d
```

If any tool is missing, `prerequisites check` returns a non-zero exit code and prints the exact `install --type ...` command to fix it, along with the reason (e.g., "installed but not running" for Docker) and a documentation link for manual setup where auto-install isn't available.

## Next Steps

Once your environment passes the prerequisites check, continue to the quick-start guide to create your first cluster and install the OpenFrame platform.
