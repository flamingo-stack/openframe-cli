# Prerequisites

Before installing and using OpenFrame CLI, make sure your environment meets the requirements below. OpenFrame CLI can auto-install most missing tools for you on macOS and Linux (see [`openframe prerequisites`](#verifying-your-setup) below), but it's useful to know what's actually required.

## Required Software

The exact tool set depends on which cluster backend you plan to use.

| Tool | Required For | Notes |
|---|---|---|
| **Docker** | Local `k3d` clusters | Must be installed and running; OpenFrame CLI checks daemon health via `docker ps` |
| **k3d** | Local `k3d` clusters | Downloaded and checksum-verified automatically by the CLI if missing (macOS/Linux) |
| **Helm** | All cluster types | Used to install ArgoCD and the app-of-apps chart |
| **Terraform** | AWS `EKS` / GCP `GKE` clusters | Provisions cloud infrastructure |
| **AWS CLI** | AWS `EKS` clusters | Used for identity/auth and discovery |
| **gcloud CLI** + `gke-gcloud-auth-plugin` | GCP `GKE` clusters | Used for auth and cluster discovery |
| **infracost** (optional) | Cloud clusters | Shows a monthly cost estimate before `cluster create --dry-run`; the CLI offers to install/login interactively if missing |
| **mkcert** (optional) | Local TLS certificates | Downloaded automatically when local HTTPS certs are needed |
| **Go 1.26+** | Building from source only | Only required if you're compiling the CLI yourself; not needed to run a released binary |
| **WSL2 + Ubuntu distro** | Windows only | Required because Docker/k3d and Kubernetes networking run natively in WSL2, not on native Windows |

## System Requirements

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU Cores | 6 | 12 |
| Disk Space | 50 GB | 100 GB |

> **Note:** These figures reflect running the full OpenFrame platform (ArgoCD + app-of-apps + all platform services) on a local `k3d` cluster. Cloud (`EKS`/`GKE`) clusters shift most compute to the cloud provider but still require enough local resources to run Terraform, kubectl, and the CLI itself.

## Account / Access Requirements

- **Local (`k3d`)**: No external account needed — everything runs on your machine via Docker.
- **AWS EKS**: Valid AWS credentials with permissions to create EKS clusters, VPCs, IAM roles, and related resources (Terraform-managed). The CLI validates AWS identity via `openframe cluster create --type eks` flows.
- **GCP GKE**: A GCP project and authenticated `gcloud` session with permissions to create GKE clusters and related networking resources.
- **Self-update**: No account required; the CLI checks and downloads releases from the public [flamingo-stack/openframe-cli releases](https://github.com/flamingo-stack/openframe-cli/releases) page.

## Environment Variables

These are optional and only relevant in specific scenarios:

| Variable | Purpose |
|---|---|
| `OPENFRAME_WSL_DISTRO` | Targets a specific WSL distro on Windows instead of the WSL default |
| `OPENFRAME_NO_WSL_FORWARD` | Disables WSL forwarding on Windows and forces native execution (unsupported for cluster operations) |
| `OPENFRAME_UPDATE_INSECURE_SKIP_VERIFY` | Emergency escape hatch to skip cosign signature verification during self-update (not recommended) |
| `GITHUB_TOKEN` / `OPENFRAME_GITHUB_TOKEN` | Forwarded into WSL for authenticated GitHub API access (e.g., higher rate limits when checking releases) |

## Verifying Your Setup

OpenFrame CLI ships a built-in prerequisites checker, scoped to the cluster type you intend to use:

```bash
# Check tools required for the default local (k3d) cluster type
openframe prerequisites check

# Check tools required for an EKS cluster
openframe prerequisites check --type eks

# Check tools required for a GKE cluster
openframe prerequisites check --type gke

# Attempt to auto-install missing tools (macOS/Linux only)
openframe prerequisites install --type k3d
```

`openframe prerequisites check` reports which tools are satisfied, which are missing, and links to install docs. On Windows, `install` prints manual installation guidance rather than auto-installing, since local cluster tooling must run inside WSL2.

## Next Steps

Once your environment is ready, continue to the [Quick Start](quick-start.md) guide to bootstrap your first OpenFrame cluster.
