# Quick Start

Install `openframe` and bootstrap a local OpenFrame environment.

[![OpenFrame 5-Minute Walkthrough](https://img.youtube.com/vi/er-z6IUnAps/maxresdefault.jpg)](https://www.youtube.com/watch?v=er-z6IUnAps)

## TL;DR

```bash
# Install (Linux/macOS example below)
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/

openframe --version
openframe bootstrap
openframe cluster status
```

## Step 1: Install the CLI

### Option A: Pre-built binary (recommended)

Download the release for your platform, then move it onto your `PATH`:

```bash
# Linux amd64  (swap linux_amd64 for linux_arm64, darwin_amd64, or darwin_arm64)
curl -fsSL https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar -xz
sudo mv openframe /usr/local/bin/
```

On **Windows**, download `openframe-cli_windows_amd64.zip` from the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest), extract it, and put `openframe.exe` on your `PATH`. The CLI re-execs itself inside WSL2, so run `openframe ...` normally.

Once installed, the CLI can update itself — see [Step 5](#step-5-keep-the-cli-up-to-date).

### Option B: Build from source

Requires Go 1.24+:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

## Step 2: Verify Installation

```bash
openframe --version
openframe --help
```

## Step 3: Bootstrap Your Environment

`openframe bootstrap` checks prerequisites, installs any missing tools, creates a k3d cluster, and installs the OpenFrame platform via the ArgoCD app-of-apps:

```bash
openframe bootstrap
```

The interactive wizard prompts for the deployment mode (`oss-tenant`, `saas-tenant`, or `saas-shared`) and other options. For CI or scripting, run non-interactively — `--non-interactive` requires `--deployment-mode`:

```bash
openframe bootstrap --deployment-mode oss-tenant --non-interactive
```

## Step 4: Verify Your Environment

Check cluster health:

```bash
openframe cluster status
```

Check the platform deployment:

```bash
openframe app status
```

Get the ArgoCD URL, admin credentials, and port-forward command:

```bash
openframe app access
```

This prints the `kubectl port-forward svc/argocd-server` command and the admin login. Run it, then open the printed URL.

## Step 5: Keep the CLI Up to Date

```bash
openframe update check      # see if a newer release is available
openframe update            # download a verified release and replace the binary
```

See [First Steps](first-steps.md#self-update) for rollback and auto-update.

## Troubleshooting

### Docker not running

```bash
docker ps        # must succeed
docker info
```

Start Docker Desktop (macOS/Windows) or `sudo systemctl restart docker` (Linux).

### kubectl can't connect

```bash
kubectl config current-context   # should be k3d-<cluster-name>
kubectl config get-contexts
```

### ArgoCD not reachable

```bash
openframe app access             # re-print URL, credentials, and port-forward
kubectl get pods -n argocd
```

## Next Steps

- **[First Steps](first-steps.md)** — Core commands and workflows

Need help? [OpenMSP Slack](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
