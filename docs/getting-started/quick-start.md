# Quick Start

This guide gets you from zero to a running local OpenFrame platform in about 5 minutes, using OpenFrame CLI's one-command `bootstrap` workflow.

## TL;DR Installation

### Windows

Download the AMD64 build directly:

```text
https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
```

Unzip the archive and run the `openframe` executable the same way you would run any other installer/binary on your system.

### macOS / Linux

Download the platform-appropriate archive from the [Releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest), unzip it, and place the `openframe` binary somewhere on your `$PATH` (e.g. `/usr/local/bin`).

If you have a Go toolchain available, you can alternatively install directly from source:

```bash
go install github.com/flamingo-stack/openframe-cli@latest
```

### Verify the install

```bash
openframe --version
```

## Hello World: Your First Cluster + Platform Install

Once the binary is on your `$PATH`, verify prerequisites and bootstrap a local environment in one step:

```bash
# 1. Check that Docker/k3d/helm are ready (auto-installs on macOS/Linux where possible)
openframe prerequisites check

# 2. Bootstrap: creates a local k3d cluster AND installs the OpenFrame platform
openframe bootstrap
```

`openframe bootstrap` runs interactively by default — it will:

1. Validate (or prompt for) a cluster name.
2. Create a local k3d cluster (Docker-backed Kubernetes-in-Docker).
3. Install ArgoCD via Helm.
4. Install the app-of-apps chart, which deploys the OpenFrame platform components.
5. Wait for all ArgoCD applications to reach a synced/healthy state.
6. Print a summary card with stage timings and access instructions.

For CI/automation, run it non-interactively (reusing an existing `openframe-helm-values.yaml`):

```bash
openframe bootstrap --non-interactive
```

## Expected Output

After a successful bootstrap, you should see a summary similar to:

```text
✓ Cluster created (k3d)
✓ ArgoCD installed
✓ app-of-apps synced and healthy
Bootstrap complete in Xm Ys

ArgoCD access:
  Username: admin
  Password: <redacted>
```

Confirm everything is healthy:

```bash
openframe app status
```

This reports cluster reachability plus the sync/health state of every ArgoCD-managed application, and a readiness summary.

To view ArgoCD admin credentials and UI access instructions at any time:

```bash
openframe app access
```

## Next Steps

- Follow the [First Steps guide](first-steps.md) to explore the platform, check status interactively, and learn common day-2 commands.
- Review the [Prerequisites guide](prerequisites.md) if any tool checks failed during bootstrap.
- Read the [Introduction](introduction.md) for a broader overview of what OpenFrame CLI manages.
