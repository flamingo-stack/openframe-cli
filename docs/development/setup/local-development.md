
# Local Development

This guide covers cloning, building, running, and debugging `openframe-cli` — the Go-based command-line tool that provisions Kubernetes clusters (locally via k3d, or in the cloud via GKE/EKS through Terraform) and deploys the OpenFrame platform onto them using ArgoCD's app-of-apps pattern.

> The platform application code the CLI deploys lives in a separate repository, [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant). This repository, `flamingo-stack/openframe-cli`, is the standalone CLI binary only — per the organization's dependency graph it publishes a single Go module (`github.com/flamingo-stack/openframe-cli`) and has no recorded upstream dependencies or downstream consumers within the organization. Treat it as its own project, not a subdirectory of `openframe-oss-tenant`.

## Prerequisites

Before cloning, confirm your machine meets the minimum requirements the CLI's cloud and local cluster providers expect:

| Resource | Minimum | Recommended |
|---|---|---|
| RAM | 24 GB | 32 GB |
| CPU Cores | 6 | 12 |
| Disk Space | 50 GB | 100 GB |

See `docs/getting-started/prerequisites.md` for the full environment checklist before you begin. The CLI itself requires:

- **Go 1.26.0** or later — `go.mod` declares `go 1.26.0` with `toolchain go1.26.5`. The toolchain pin matters: Go 1.26.0/1.26.1 crash intermittently on windows-amd64 during GC stack scanning, and the fix only shipped in 1.26.2, so Go resolves the toolchain from `go.mod` automatically if your local `go` is older.
- **Docker** — required for the k3d local-cluster provider.
- **git** — the CLI shells out to clone the app-of-apps chart repository at install time.

Tool-specific prerequisites for cluster creation (k3d, helm, terraform, gcloud, aws) are checked and, where supported, auto-installed by the CLI's own `prerequisites` subsystem — see `internal/prerequisites/runner.go` and `internal/cluster/prerequisites/`. You do not need to pre-install those to build or test the CLI itself; they only matter when you run cluster-provisioning commands against a real target.

## Clone and Setup

Clone the repository from its canonical location:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

Verify your Go toolchain and download module dependencies:

```bash
go version
go mod download
```

`go.mod` lists the full dependency set, including `spf13/cobra` (command routing), `kubernetes/client-go` (native Kubernetes API access), `pterm/pterm` and `charmbracelet/huh`/`bubbletea` (terminal UI), and `sigstore/sigstore-go` (self-update signature verification). No `go mod tidy` is required for a fresh clone, but the `Makefile` exposes a `tidy` target that fails if `go mod tidy` would change `go.mod`/`go.sum` — run it before committing a dependency change:

```bash
make tidy
```

### Building the binary

The `Makefile` is the primary build entry point. For your current platform:

```bash
make build
```

This builds with `CGO_ENABLED=0` and `-trimpath` (dropping absolute build paths for reproducibility, matching the release build), placing the binary at `build/openframe-<GOOS>-<GOARCH>` (or with a `.exe` suffix on Windows). `GOOS`/`GOARCH` are auto-detected from `go env`.

To cross-compile every release platform at once (matching what the release pipeline ships):

```bash
make build-all
```

This produces `build/openframe-linux-amd64`, `build/openframe-linux-arm64`, `build/openframe-darwin-amd64`, `build/openframe-darwin-arm64`, `build/openframe-windows-amd64.exe`, and `build/openframe-windows-arm64.exe`.

To remove build artifacts:

```bash
make clean
```

`clean` only removes the platform-suffixed binaries it produced (it deliberately avoids a broad glob that would also delete tracked files such as `openframe-helm-values.example.yaml`).

Run `make help` at any time to list every available target with a short description.

### Alternative: install via `go install`

For a quick local install without cloning (pulls the latest published module):

```bash
go install github.com/flamingo-stack/openframe-cli@latest
```

### Windows

Windows users can download the prebuilt AMD64 archive directly from the Releases page instead of building from source:

```text
https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_windows_amd64.zip
```

Note that on Windows, the native build re-executes itself inside WSL2 (`internal/shared/wsllauncher`) before any cluster command runs — the local cluster (Docker + k3d) and the native Kubernetes client (`client-go`) need to share a network namespace that a native Windows process cannot reach reliably. This happens transparently; you do not need to invoke WSL yourself.

## Running Locally

Once built, invoke the binary directly from `build/`, or run it straight from source with `go run`:

```bash
# Run the freshly built binary
./build/openframe-$(go env GOOS)-$(go env GOARCH) --help

# Or run directly from source, no build step
go run . --help
```

The CLI's top-level command groups, defined under `cmd/`:

```text
openframe bootstrap                    # k3d cluster + platform install in one step
openframe cluster create|delete|list|status|use|cleanup
openframe app install|upgrade|status|access|uninstall
openframe prerequisites check|install
openframe update [version]|check|rollback
```

A typical local development loop — bring up a disposable k3d cluster and install the platform in one step:

```bash
go run . bootstrap my-dev-cluster --verbose
```

`bootstrap` (see `cmd/bootstrap/bootstrap.go`, `internal/bootstrap/service.go`) chains `cluster create` and `app install` into one staged flow with a progress tracker. `--non-interactive` skips prompts and reuses an existing `openframe-helm-values.yaml`, which is the mode to use in CI/automation; `--verbose` prints ArgoCD sync progress and other debug detail.

To inspect platform state without reinstalling:

```bash
go run . app status --context k3d-my-dev-cluster
go run . app status --watch          # live-refreshing view, Ctrl+C to exit
go run . app status --interactive    # k9s-style TUI (internal/app/status/tui)
go run . app access                  # print ArgoCD admin credentials + UI access steps
```

`app status` and `app access` are annotated read-only (`Annotations: {"readonly": "true"}` in `cmd/app/status.go` / `cmd/app/access.go`) and accept `--output json|yaml` for machine-readable output via `-o`. `--watch`/`--interactive` require an interactive terminal and cannot combine with `--output` or `--plain`.

To tear the platform down without deleting the underlying cluster:

```bash
go run . app uninstall --context k3d-my-dev-cluster
```

`uninstall` prompts for confirmation unless `--yes` is passed, deletes ArgoCD Applications and Helm releases in a safe order (`internal/app/uninstall/uninstall.go`), and leaves the cluster running.

## Hot Reload / Watch Mode

`openframe-cli` is a compiled Go CLI with no frontend build step and no `package.json` in this repository, so there is no bundler-style hot-module-reload workflow. The practical equivalents during development are:

- **`go run .`** — compiles and runs in one step on every invocation, which is the fastest "edit, then re-run" loop for iterating on command logic. Combine with `--verbose` to see the extra debug output most commands gate behind that flag.
- **`app status --watch`** — a live-refreshing terminal view (`cmd/app/status.go`, `watchStatus`) that polls the platform every 3 seconds and re-renders in place. This is the closest thing to a "watch mode" the CLI ships, though it watches cluster state rather than your source tree.
- **`app status --interactive`** — opens an interactive TUI (`internal/app/status/tui/tui.go`) for navigating ArgoCD applications, inspecting details, and triggering syncs without re-running separate commands.

If you want automatic rebuild-on-save while editing Go source, use a generic Go file watcher in your editor or terminal (for example an editor's built-in "run on save" feature) that invokes `go run .` or `make build` — this is a personal workflow choice, not something the repository's `Makefile` or scripts provide out of the box.

## Debug Configuration

### `--verbose` and `--silent`

Nearly every command honors the root-level `--verbose` flag, which surfaces extra detail — most visibly ArgoCD sync progress during `app install`/`app upgrade`/`bootstrap`. Command groups that define their own `PersistentPreRunE` (such as `app`) must explicitly re-apply the global output flags via `ui.ApplyGlobalOutputFlags(cmd)`, because Cobra only runs the closest parent's hook — this is enforced as an active code-review rule (`OPENFRAM-005-10`) on this repository, so verbose/silent behavior is consistent across every subtree.

```bash
go run . app install --verbose
go run . bootstrap --non-interactive --verbose
```

`--silent` suppresses normal output; all user-facing output in this codebase is required to go through `pterm` printers or the shared `internal/shared/ui` package rather than raw `fmt.Print*` (rule `OPENFRAM-007`), specifically so `--silent`/`--plain` and test output redirection work uniformly.

### Dry runs

Destructive or state-changing commands support `--dry-run` to preview an action without executing it:

```bash
go run . app install --dry-run
go run . app upgrade --ref main --dry-run
go run . app upgrade --sync --dry-run
```

For `app upgrade`'s force-sync mode, `--dry-run` lists which ArgoCD applications are out of sync without triggering a sync (`previewOutOfSync` in `cmd/app/upgrade.go`).

### Exit codes

`main.go` preserves a failed external command's exit code rather than collapsing every failure to a generic `1`: it unwraps `*executor.CommandError` (see `internal/shared/executor/executor.go`) and propagates the original code when it is a valid, non-zero Unix exit code. This matters when scripting around the CLI (e.g. in CI) — a `helm`/`terraform`/`k3d` failure surfaces its real exit status rather than an opaque `1`. Errors already shown to the user carry an `AlreadyHandledError` sentinel (`internal/shared/errors`) so `main` does not print them a second time.

### Debugging with a Go debugger

Because the CLI shells out to external tools (Docker, k3d, Helm, Terraform, gcloud, aws) through the `executor.CommandExecutor` abstraction (`internal/shared/executor/executor.go`), you can attach a standard Go debugger (such as Delve) to `go run .` or a `make build` binary and step through command logic exactly as you would for any Go program — no special launch wrapper is required. Useful breakpoint targets when diagnosing a specific flow:

- `internal/bootstrap/service.go` — the `bootstrap` orchestration (cluster create → app install, with a step tracker).
- `internal/cluster/service.go` — `ClusterService`, which resolves the right provider (k3d/EKS/GKE) and drives create/delete/status.
- `internal/chart/services/installer.go` — the core install sequence (ref preflight → ArgoCD install → app-of-apps install → wait for applications).
- `internal/shared/executor/mock.go` — a mock `CommandExecutor` used by the test suite; useful as a reference when you want to trace exactly which external commands a code path shells out to without actually running them.

### Testing as a debugging aid

The `Makefile` exposes several test targets useful while developing:

```bash
make test-unit        # unit tests across ., ./cmd/..., ./internal/..., ./tests/testutil/..., ./tests/scripts/...
make test-race        # same, with the race detector (CGO_ENABLED=1)
make test-integration # real k3d clusters — needs docker + k3d installed
make test             # test-unit + test-integration
```

`test-integration` is opt-in via the `integration` build tag specifically so `go test ./...` never triggers it by accident — it provisions real k3d clusters and runs a full bootstrap. Its harness builds its own CLI binary (see `tests/integration/common/cli_runner.go`), so `make build` is not a prerequisite, but Docker and k3d must be installed and running.

There is also a Terraform-template validation target, useful when touching the EKS or GKE provider templates:

```bash
make test-tfvalidate  # needs terraform + network; downloads providers (~1 min)
```

This runs a real `terraform init` and `terraform validate` against the generated root modules in `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`, without requiring cloud credentials.

### Linting and formatting

```bash
make fmt   # gofmt -l -w . — format in place
make vet   # go vet ./...
make lint  # golangci-lint run ./... (govet, staticcheck, errcheck, gosec, ineffassign)
```

`make lint` requires `golangci-lint` to be installed separately; the target checks for it and prints an install link if missing.

### Signed release binaries (not part of local development)

`scripts/sign-binary.sh` signs GoReleaser-built binaries for macOS (codesign + notarytool) and Windows (Azure Trusted Signing via jsign) as part of the release pipeline. It is a no-op unless the `OPENFRAME_SIGN` environment variable is set to `1`, which only happens in the release workflow — local builds and `make build`/`make build-all` never attempt to sign, so you can safely ignore it during day-to-day development.

## Where to Go Next

For the full architecture — command layer, domain services, providers, and how they fit together — see `docs/reference/architecture/overview.md` and the accompanying diagrams in `docs/diagrams/architecture/`. For environment-specific configuration details, see `./environment.md`. For the end-to-end walkthroughs of provisioning and managing a cluster with the built CLI, see `docs/getting-started/how-to/create-a-new-kubernetes-cluster-with-openframe-cli.md`, `docs/getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md`, and `docs/getting-started/how-to/check-the-status-of-your-openframe-cluster-and-apps.md`.
