# Prerequisites

## System Requirements

### Hardware

A full local OpenFrame platform is resource-intensive. Recommended host:

| Resource | Recommended |
|----------|-------------|
| RAM | 24 GB |
| CPU | 6 cores |
| Disk | 50 GB free |
| Architecture | x86_64 or ARM64 |

### Operating System

| OS | Notes |
|----|-------|
| Linux | Ubuntu 20.04+, RHEL/CentOS 8+ |
| macOS | 11 (Big Sur) or later |
| Windows | Windows 10/11 with WSL2 |

On Windows, `openframe` auto-forwards the whole invocation into WSL2 and runs as a Linux binary — just run `openframe ...` normally. There's no need to `wsl -d Ubuntu` first.

## Dependencies

**Docker is the only tool you must install and run yourself.** It provides the container runtime that k3d clusters run on. See the [Docker install guide](https://docs.docker.com/get-docker/).

The CLI installs everything else automatically:

- **kubectl, k3d, helm** — downloaded as verified, version-pinned binaries into `~/.openframe/bin`
- **mkcert** — used to issue a locally-trusted certificate for the localhost HTTPS ingress. `mkcert -install` modifies the OS trust store (and may prompt for sudo), so it is skipped in non-interactive mode.

## Checking and Installing

Verify Docker is running:

```bash
docker ps
```

Report on all prerequisites:

```bash
openframe prerequisites check
```

Install anything that's missing:

```bash
openframe prerequisites install
```

`openframe bootstrap` also runs these checks and installs missing tools before creating a cluster.

## Network Requirements

Outbound internet access is needed to pull container images, download Helm charts, clone the platform repository, and fetch pinned tool binaries.

## Next Steps

- **[Quick Start](quick-start.md)** — Install the CLI and bootstrap your first environment
- **[First Steps](first-steps.md)** — Explore core commands and workflows

Need help? [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
