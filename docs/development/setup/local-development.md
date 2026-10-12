## Local Development

This guide covers cloning and building the OpenFrame CLI, running it against a local cluster, iterating with fast rebuild/watch loops, and configuring a debugger — all from the Go toolchain and `Makefile` in this repository.

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a Go `service`-role artifact with no recorded upstream dependencies on other repositories in the organization's graph and no recorded downstream consumers — per the ecosystem graph it is a leaf/terminal artifact whose dependencies are exclusively third-party Go modules and external CLI tools (Docker, k3d, Helm, Terraform, gcloud, aws). This document only covers developing the CLI binary itself. It does not cover the OpenFrame platform (ArgoCD, backing services) that the CLI deploys — that lives in the separate **flamingo-stack/openframe-oss-tenant** repository.

> **Note:** This repository has no `package.json`, no Docker Compose files, and no CLI-bootstrap scaffolding of its own — it ships one Go module built with a standard `Makefile`. "Running locally" here means running the compiled `openframe` binary against a Kubernetes cluster, not running a web server.

### Prerequisites

Before cloning, confirm your machine meets the hardware and tooling baseline described in `../../getting-started/prerequisites.md`. At minimum, developing this CLI requires:

- **Go** — the version pinned in `go.mod` (`go 1.26.0`, toolchain `go1.26.5`). The toolchain directive matters: `go.mod` pins `go1.26.5` specifically because Go `1.26.0`/`1.26.1` crash intermittently on `windows-amd64` during GC stack scanning; running `go build`/`go test` with an unpinned toolchain on Windows risks hitting that bug.
- **golangci-lint** — only needed for `make lint`.
- **Docker** and **k3d** — only needed if you intend to actually create a local cluster and install the OpenFrame platform (`openframe bootstrap` / `openframe cluster create`), not to build or unit-test the CLI itself.
- **Terraform** — only needed for the `test-tfvalidate` target or for exercising the EKS/GKE cloud-provider code paths.

### Clone and setup

Clone the repository from its canonical location:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

Verify your module graph is tidy and fetch dependencies (this also happens implicitly on first build/test):

```bash
go mod download
make tidy
```

`make tidy` runs `go mod tidy -diff`, which fails if running `go mod tidy` for real would change `go.mod`/`go.sum` — use it as a pre-commit sanity check rather than a command that mutates files.

Format and vet the source before making changes:

```bash
make fmt   # gofmt -l -w . — formats Go source in place
make vet   # go vet ./...
```

If `golangci-lint` is installed, run the full lint suite (govet, staticcheck, errcheck, gosec, ineffassign):

```bash
make lint
```

### Building

Build a binary for your current platform into `build/`:

```bash
make build
```

This resolves `GOOS`/`GOARCH` via `go env` and runs:

```bash
CGO_ENABLED=0 go build -trimpath -o build/openframe-<os>-<arch> .
```

`CGO_ENABLED=0` and `-trimpath` are not incidental — they are required for every production build in this repository (matching the release pipeline in `.goreleaser.yml`), so local builds already mirror what ships. `CGO_ENABLED=0` produces a fully static binary that runs without a libc on the target, and `-trimpath` strips the build machine's absolute paths for reproducibility.

The only sanctioned exception is the race-detector test target, which explicitly re-enables cgo (see Testing, below).

To cross-compile every release platform at once (useful before verifying platform-specific behavior, e.g. the WSL launcher or Windows signing):

```bash
make build-all
```

This produces `build/openframe-linux-amd64`, `build/openframe-linux-arm64`, `build/openframe-darwin-amd64`, `build/openframe-darwin-arm64`, `build/openframe-windows-amd64.exe`, and `build/openframe-windows-arm64.exe`.

Clean build artifacts with:

```bash
make clean
```

This only removes the platform-suffixed binaries (`openframe-linux-*`, `openframe-darwin-*`, `openframe-windows-*`, and anything under `build/`) — it deliberately does not use a broad `openframe-*` glob, since that previously also matched and deleted tracked files like `openframe-helm-values.example.yaml`.

List all available Make targets with inline descriptions:

```bash
make help
```

### Running locally

Once built, invoke the binary directly from `build/` (or `go run .` for a quick one-off without a persisted binary):

```bash
# Build then run
make build
./build/openframe-<os>-<arch> --help

# Or run without building a persisted binary
go run . --help
```

The root command and its subcommand groups are wired in `cmd/root.go`, with each command group under its own `cmd/` subpackage: `cmd/bootstrap`, `cmd/cluster`, `cmd/app`, `cmd/prerequisites`, `cmd/update`.

Common local workflows while developing:

```bash
# Check which prerequisite tools (Docker, k3d, helm, etc.) are present
go run . prerequisites check

# Create a local k3d cluster and install the platform in one step
go run . bootstrap

# Or do it in two steps
go run . cluster create
go run . app install

# Inspect what you just created
go run . cluster status
go run . app status
```

Every command respects global `--silent`, `--verbose`, and `--plain` flags, plus a `--output json|yaml` flag on read-only subcommands (`app status`, `app access`) for machine-readable output during scripting. `cmd/app/app.go`'s `PersistentPreRunE` shows the CLI logo and current kube-context before most `app` subcommands run, except when machine-readable output is requested — that path is kept clean for scripts.

For full walkthroughs of cluster creation, platform installation, status checks, and teardown, see `../../getting-started/how-to/create-a-kubernetes-cluster-with-openframe-cli.md`, `../../getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md`, `../../getting-started/how-to/check-the-status-of-your-openframe-cluster-and-app.md`, and `../../getting-started/how-to/uninstall-openframe-and-delete-a-cluster.md`.

### Hot reload / watch mode

This repository ships no file-watcher or hot-reload tooling (no `air`, no `nodemon`-equivalent, no watch target in the `Makefile`). The practical "watch mode" loop for Go CLI development here is:

```bash
go run . <command> --verbose
```

`go run .` recompiles the module from source on every invocation, so there is no separate "rebuild" step to remember — edit a file under `cmd/` or `internal/`, then re-run the same `go run . ...` command to pick up the change immediately. This is the fastest iteration loop for exercising a single subcommand's flags or output formatting.

For iterating on logic covered by unit tests, re-run the relevant package's tests after each change rather than exercising the full CLI:

```bash
# Fast inner loop: an iterative unit-test suite is the closest the project
# comes to continuous feedback while editing.
go test -count=1 ./internal/chart/... 
go test -count=1 ./internal/cluster/...
```

`make test-unit` (described under Testing) runs the whole unit-test package set at once and is a good checkpoint before committing, but for the tightest edit/test loop target individual packages directly with `go test`.

> Since the CLI shells out to real tools (Docker, k3d, Helm, Terraform, cloud CLIs) via the `CommandExecutor` abstraction in `internal/shared/executor`, most of the "watch" value comes from unit tests that use the mock executor (`internal/shared/executor/mock.go`) rather than from re-running `go run .` against a live cluster on every keystroke.

### Testing

The `Makefile` defines several test targets, each scoped differently:

```bash
make test-unit         # Unit tests only (fast, no external services)
make test-race         # Unit tests with the race detector (needs CGO_ENABLED=1)
make test-integration  # Real k3d clusters — needs docker + k3d installed
make test-tfvalidate   # Validates generated Terraform templates — needs terraform + network
make test              # test-unit + test-integration together
```

`test-unit` runs against the package set `. ./cmd/... ./internal/... ./tests/testutil/... ./tests/scripts/...` — this deliberately includes the root package (`main_test.go`, which is the only place exit-code fidelity is tested) and `tests/testutil`, both of which a bare `go test ./cmd/... ./internal/...` would silently skip. It deliberately excludes `./tests/integration/...`.

`test-race` re-enables cgo (`CGO_ENABLED=1 go test -race -count=1 ...`) — the one sanctioned exception to the CGO_ENABLED=0 rule that otherwise governs all builds in this repository.

`test-integration` is opt-in via the `integration` build tag and creates **real k3d clusters**, running a full bootstrap end-to-end. It builds its own CLI binary internally (see `tests/integration/common/cli_runner.go`), so `make build` is not a prerequisite, but Docker and k3d must be installed and will incur real local resource usage.

`test-tfvalidate` is opt-in via the `tfvalidate` build tag and runs a real `terraform init` (downloading the pinned EKS/GKE modules and providers over the network) followed by `terraform validate` against the generated root modules in `internal/cluster/providers/eks/templates/main.tf` and `internal/cluster/providers/gke/templates/main.tf`. No cloud credentials are required, but it does need network access and the `terraform` binary.

Testing conventions, patterns, and shared assertions are covered in `../testing/README.md`.

### Debug configuration

There is no committed IDE launch configuration (no `.vscode/launch.json` or similar) in this repository, so debugging relies on the Go toolchain directly plus the CLI's own verbosity flags.

**Verbose output first.** Before reaching for a debugger, most issues in this CLI are diagnosable with the built-in `--verbose` flag, which most command groups thread through to surface detailed logs — including ArgoCD sync progress during `app install`/`app upgrade`/`bootstrap`:

```bash
go run . bootstrap --verbose
go run . app status --watch --verbose
```

**Debugging with Delve.** Since this is a standard Cobra-based Go binary, attach [Delve](https://github.com/go-delve/delve) directly to a `go run`/`go build` invocation or to the compiled binary in `build/`:

```bash
# Debug a specific subcommand invocation
dlv debug . -- bootstrap --non-interactive

# Debug an already-built binary
dlv exec ./build/openframe-<os>-<arch> -- cluster status
```

Because every external command this CLI shells out to (Docker, k3d, Helm, Terraform, cloud CLIs) goes through the `CommandExecutor` abstraction in `internal/shared/executor`, you can set breakpoints at that boundary (e.g. in the executor's `Execute` method) to inspect exactly which command, arguments, and environment are about to run before they hit a real subprocess — this is often more useful than stepping through the higher-level orchestration code in `internal/cluster` or `internal/chart`.

**Debugging without live infrastructure.** `internal/shared/executor/mock.go` provides a mock `CommandExecutor` used throughout the unit-test suite. When debugging logic that would otherwise require a real cluster, cloud credentials, or Docker, write or run a focused unit test that injects the mock executor instead of attaching a debugger to a live `cluster create`/`app install` run — it is both faster and avoids provisioning real (potentially billed, for GKE/EKS) resources while you iterate.

**Exit codes matter for debugging automation failures.** `main.go`'s `exitCode()` function unwraps `*executor.CommandError` to propagate a failed external command's real exit code (when it is a valid, non-zero code under 256) rather than collapsing every failure to a generic `1`. Errors that have already been printed to the user carry the `*sharederrors.AlreadyHandledError` sentinel (checked via `errors.As` in `main.go`) so they are not printed twice — if you are debugging "why did this error print twice" or "why did this print nothing", check whether the failing code path is using `sharedErrors.HandleGlobalError` correctly before it returns.

### Where to go next

For the broader system this CLI provisions and installs — ArgoCD, the app-of-apps chart, and the services it deploys — see the separate `flamingo-stack/openframe-oss-tenant` repository and its own documentation tree. For this repository's internal architecture (command/service/provider layering, bootstrap and status-aggregation sequences), see `../../reference/architecture/overview.md` and `../../reference/architecture/ecosystem.md`. For contribution conventions and review expectations, see `../contributing/guidelines.md`. For environment-variable and configuration details beyond what this guide covers, see `./environment.md`.
