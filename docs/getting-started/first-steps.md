
# First Steps

This page walks you through what to do right after installing the OpenFrame CLI (`openframe`): the first commands to run, the configuration you'll typically touch, the key features worth exploring, and where to go if you get stuck.

If you have not installed the CLI yet, see [`./introduction.md`](./introduction.md) and [`./prerequisites.md`](./prerequisites.md) first. If you want the fastest path to a running platform, [`./quick-start.md`](./quick-start.md) covers the single-command bootstrap flow — this page instead walks through the individual steps and what each one configures.

> **Scope reminder:** the OpenFrame CLI (this repository) is the tool that provisions clusters and deploys the platform. The platform's application code — the services that actually make up OpenFrame — lives in the separate [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) repository, which the CLI clones and deploys for you via its app-of-apps Helm chart. You won't need to clone that repository yourself for the steps below.

## Before you begin: hardware

OpenFrame deploys a full Kubernetes-based platform (ArgoCD, the app-of-apps chart, and the services it manages), so the machine running the cluster needs real headroom:

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU Cores | 6 | 12 |
| Disk Space | 50 GB | 100 GB |

## The first 5 things to do after installation

### 1. Confirm the binary is installed and check its version

Run the root command with no arguments, or check the version metadata that `cmd/root.go` resolves at build time:

```bash
openframe --help
```

This prints the command tree: `bootstrap`, `cluster`, `app`, `prerequisites`, and `update`. If this prints help text instead of a "command not found" error, the install succeeded.

### 2. Check and install prerequisites

Before provisioning anything, verify your machine has the tools OpenFrame needs (Docker, k3d, Helm, and — for cloud clusters — Terraform, gcloud, or the AWS CLI):

```bash
openframe prerequisites check
```

This reports which required tools are present without changing anything. If something is missing:

```bash
openframe prerequisites install
```

On macOS and Linux this auto-installs missing tools. On Windows it prints manual installation links instead of installing automatically. You can scope either command to a specific cluster backend with `--type k3d`, `--type eks`, or `--type gke`.

Some of the tools this step installs — k3d, Helm, mkcert, Terraform, and infracost — are downloaded by the CLI itself through its internal download pipeline, pinned to exact versions and verified against SHA256 digests before use, never fetched via an unverified pipe-to-shell.

### 3. Create your first cluster

With prerequisites satisfied, create a cluster. For a local cluster (the fastest option, backed by k3d/Docker):

```bash
openframe cluster create
```

Run interactively (it will prompt you for a cluster name and options), or pass a name directly and skip the wizard:

```bash
openframe cluster create my-cluster --skip-wizard
```

For cloud clusters, pass `--type eks` or `--type gke`; cloud creation shows a Terraform plan (and an optional infracost estimate) before applying anything. A full, detailed walkthrough of this step — including the wizard options and non-interactive flags — continues in [`./how-to/create-a-new-kubernetes-cluster-with-openframe-cli.md`](./how-to/create-a-new-kubernetes-cluster-with-openframe-cli.md).

> If you'd rather combine cluster creation and platform installation into one command, `openframe bootstrap [cluster-name]` does both steps in sequence (with `--non-interactive` for CI, and `--verbose` for detailed ArgoCD sync progress) — see the bootstrap sequence in [`../reference/architecture/overview.md`](../reference/architecture/overview.md).

### 4. Install the OpenFrame platform onto the cluster

Once you have a reachable cluster (created with `openframe cluster create`, or any other cluster you point `--context` at), deploy ArgoCD and the app-of-apps chart:

```bash
openframe app install
```

or, targeting a specific cluster by name:

```bash
openframe app install my-cluster
```

This is where the CLI clones the `flamingo-stack/openframe-oss-tenant` repository (via the `GitHubRepo`/`GitHubBranch` settings in `internal/chart/models/appofapps.go`) and installs it through Helm and ArgoCD. Pass `--ref` to deploy a specific branch or release tag instead of the default. The full procedure, including configuration prompts, continues in [`./how-to/install-the-openframe-platform-onto-a-cluster.md`](./how-to/install-the-openframe-platform-onto-a-cluster.md).

### 5. Verify the platform came up healthy

After installation finishes (or while it's in progress), check status:

```bash
openframe app status
```

This reports cluster reachability plus ArgoCD application sync/health. Add `--watch` to keep it refreshing, or `--interactive` to open a k9s-style terminal UI for browsing application state live. The detailed status workflow continues in [`./how-to/check-the-status-of-your-openframe-cluster-and-apps.md`](./how-to/check-the-status-of-your-openframe-cluster-and-apps.md).

Once the platform is healthy, get the ArgoCD admin credentials and the port-forward command to reach its UI:

```bash
openframe app access
```

## Common initial configuration

### Helm values file

Non-default installs are driven by a Helm values file named `openframe-helm-values.yaml`, read from and written to your current working directory (see `internal/chart/utils/config/paths.go`). When you run `openframe app install` interactively, the configuration wizard (`internal/chart/ui/configuration/wizard.go`) offers two modes:

- **Default configuration** — uses the existing `openframe-helm-values.yaml` as-is, or creates one with defaults if none exists.
- **Interactive configuration** — walks you through three sections in order: branch, Docker registry, and ingress.

```text
Configuration Mode
> Default configuration
  Interactive configuration
```

#### Branch configuration

Controls which branch of `flamingo-stack/openframe-oss-tenant` gets deployed (`internal/chart/ui/configuration/branch.go`). You can keep the current branch or specify a custom one — useful for testing a feature branch of the platform before it merges.

#### Docker registry configuration

Lets you supply Docker registry credentials (username, password/token, email) if your deployment needs to pull from a private registry (`internal/chart/ui/configuration/docker.go`). Credentials you enter here are registered with the CLI's secret redactor so they never appear in verbose logs or error output.

#### Ingress configuration

Chooses how the platform is exposed (`internal/chart/ui/configuration/ingress.go`):

- **Localhost** — local-only visibility, no external exposure.
- **Ngrok** — external visibility through an ngrok tunnel. This path asks you to register an ngrok account, create a domain, generate an API key, and add a tunnel authtoken (all via the ngrok dashboard), then optionally configure a CIDR IP allowlist for restricting access.

> API keys, auth tokens, and Docker passwords entered during the wizard are registered as secrets immediately at the point of collection — before they can reach any log output.

### Kube-context targeting

Commands under `app` (install, upgrade, status, access, uninstall) take **no positional arguments** for targeting — use `--context` explicitly instead, to avoid accidentally operating against whatever cluster is the current kube-context. This is enforced deliberately: passing a stray cluster name to a destructive command like `uninstall` used to run silently against the current context.

### Switching the active cluster context

If you manage multiple clusters, switch which one `kubectl`/the CLI targets:

```bash
openframe cluster use my-cluster
```

For GKE clusters this also updates your local gcloud configuration.

## Exploring key features

### Cluster lifecycle management

The `cluster` command group covers the full lifecycle beyond creation:

```text
openframe cluster create|delete|list|status|use|cleanup
```

- `list [--all]` — lists clusters the CLI manages; `--all` additionally discovers external GKE/EKS clusters read-only.
- `status [NAME]` — detailed health, nodes, and installed applications, with `-o json|yaml` for scripting.
- `cleanup [NAME]` — prunes unused container images from cluster nodes without touching installed apps.
- `delete [NAME]` — destroys a cluster; cloud deletes require typed name confirmation unless `--force`.

### Platform upgrades

Two distinct upgrade modes are available under `openframe app upgrade`:

- **Mode 1** — re-deploy at a new `--ref` (a different branch or tag of the platform).
- **Mode 2** — force a re-sync of the current ref via ArgoCD, using `--sync` (optionally with `--prune`).

The full upgrade procedure, including when to use each mode, continues in [`./how-to/upgrade-openframe-cli-and-its-installed-apps.md`](./how-to/upgrade-openframe-cli-and-its-installed-apps.md).

### Self-update

The CLI can update itself:

```text
openframe update [version]|check|rollback
```

- `update [version]` — updates to the latest release, or switches to a specific tagged version (up or down). Every download is checksum- and signature-verified via Sigstore/cosign, anchored through TUF to the public-good Sigstore trust root.
- `update check` — reports update availability without changing anything.
- `update rollback` — reverts to the previously-installed version offline, with no download.

### Teardown

When you're done with a cluster, or want to remove just the platform while keeping the cluster running:

```bash
openframe app uninstall
```

removes ArgoCD and the app-of-apps Helm releases/Applications, in a safe order (Applications deleted first, then Helm releases, then finalizers cleared), leaving the cluster itself intact.

To remove the cluster entirely as well, continue to [`./how-to/uninstall-openframe-and-delete-a-cluster.md`](./how-to/uninstall-openframe-and-delete-a-cluster.md).

### Windows users: WSL2 under the hood

On Windows, cluster and platform operations transparently re-execute inside WSL2 before running, because the local cluster (Docker + k3d) and the native Kubernetes client need to share a network namespace that a native Windows process can't reach reliably. You don't need to do anything differently — just run `openframe` commands as usual from Windows; the WSL launcher handles the handoff and auto-installs the Linux binary into WSL if it isn't there yet.

## Architecture at a glance

For a deeper look at how these commands map to the CLI's internal services (cluster provisioning, chart/ArgoCD installation, prerequisites, self-update) and how they talk to Docker, Kubernetes, cloud provider APIs, and GitHub Releases, see [`../reference/architecture/overview.md`](../reference/architecture/overview.md) and [`../reference/architecture/ecosystem.md`](../reference/architecture/ecosystem.md). Diagram sources are also available under [`../diagrams/architecture/README.md`](../diagrams/architecture/README.md).

Per the organization's repository graph, OpenFrame CLI publishes a single Go module (`github.com/flamingo-stack/openframe-cli`) and has no recorded upstream dependencies or downstream consumers within the organization — it's a standalone, user-facing binary distributed via GitHub Releases and `go install`, not a library other services import.

## Where to get help

This project does not use GitHub Issues or GitHub Discussions. All community support, questions, and discussion happen on the OpenMSP Slack community:

- OpenMSP: https://www.openmsp.ai/
- Slack invite: https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA

For background on the wider Flamingo platform and OpenFrame product, see https://flamingo.run and https://openframe.ai.

If you want to go deeper on contributing to the CLI itself, see [`../development/README.md`](../development/README.md), the environment and local-development setup guides at [`../development/setup/environment.md`](../development/setup/environment.md) and [`../development/setup/local-development.md`](../development/setup/local-development.md), the architecture notes at [`../development/architecture/README.md`](../development/architecture/README.md), the security guidance at [`../development/security/README.md`](../development/security/README.md), the testing guide at [`../development/testing/README.md`](../development/testing/README.md), and the contribution guidelines at [`../development/contributing/guidelines.md`](../development/contributing/guidelines.md).

> **Remember:** the platform application code you're deploying lives in [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant), not in this CLI repository — bug reports or questions about the deployed services themselves ultimately trace back there, but for CLI issues and general platform questions the OpenMSP Slack is the shared channel either way.
