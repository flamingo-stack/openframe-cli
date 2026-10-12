
# Local Development

This guide covers setting up the OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) for local development: cloning the repository, building and running it, iterating with a fast rebuild loop, and debugging it in your editor.

> OpenFrame CLI is a standalone Go **service** repository — it publishes the `go` module `github.com/flamingo-stack/openframe-cli` and has no recorded upstream dependencies on, or downstream consumers from, other repositories in this organization's graph, per the repository ecosystem data. The platform it deploys (ArgoCD + app-of-apps) lives in the separate **flamingo-stack/openframe-oss-tenant** repository; this CLI only talks to that platform over the Kubernetes API and `helm`/`git`, it does not vendor or embed its source.

## What this repository is

OpenFrame CLI is an interactive command-line tool for provisioning Kubernetes clusters (local k3d, or cloud GKE/EKS via Terraform) and deploying the OpenFrame platform onto them via ArgoCD's app-of-apps pattern. It is written in Go using Cobra for command routing, and shells out to Docker, k3d, Helm, Terraform, `gcloud`, and `aws` as needed. See `../architecture/README.md` and `../../reference/architecture/overview.md` for the full architectural breakdown.

## Prerequisites

Before working on the CLI itself, make sure you have:

- **Go** — the toolchain pinned in `go.mod` is `go1.26.5` (module declares `go 1.26.0`). A comment in `go.mod` notes that Go 1.26.0/1.26.1 crash at random on windows-amd64 during GC stack scanning, which is exactly why the toolchain line pins `1.26.5`.
- **Git** — for cloning and for the CLI's own `internal/chart/providers/git` package, which clones the app-of-apps chart repository at a given ref.
- **Docker** and **k3d** — only needed if you plan to actually run CLI commands that create a local cluster (`openframe cluster create`, `openframe bootstrap`). Not required to build or unit-test the CLI itself.
- **golangci-lint** — only needed for `make lint`; the Makefile checks for it on `$PATH` and prints an install link if missing.

For the broader hardware/tooling expectations of the OpenFrame platform the CLI deploys, see `../../getting-started/prerequisites.md`.

## Clone and setup

Clone the fork you'll be working in:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

Download dependencies (standard Go module flow — there is no separate `npm install`/`pip install` step, since this is a pure Go CLI with no frontend `package.json` or backend `pom.xml`):

```bash
go mod download
```

Verify `go.mod`/`go.sum` are tidy (this is also checked in CI via the `tidy` Makefile target):

```bash
make tidy
```

Build the binary for your current platform into `build/`:

```bash
make build
```

This resolves `GOOS`/`GOARCH` from `go env` and produces `build/openframe-<os>-<arch>`, built with `CGO_ENABLED=0 go build -trimpath` (the `-trimpath` flag strips absolute build paths for reproducibility, matching the release build).

If you need binaries for every release platform (linux/darwin/windows, amd64/arm64) at once, cross-compile them all with:

```bash
make build-all
```

> **Windows users:** prebuilt release archives are also published — see `openframe-cli_windows_amd64.zip` under the [releases page](https://github.com/flamingo-stack/openframe-cli/releases/latest) — but for local development, building from source with `make build` on any platform works identically; the Makefile cross-compiles Windows binaries (`.exe`) from any host OS via `build-all`.

## Running locally

Run the CLI directly with `go run` during development, without building a binary first:

```bash
go run . --help
go run . cluster create
go run . app status
```

Or run the binary you built with `make build`:

```bash
./build/openframe-$(go env GOOS)-$(go env GOARCH) --help
```

The root command and its subcommands (defined across `cmd/root.go`, `cmd/bootstrap/bootstrap.go`, `cmd/cluster/cluster.go`, `cmd/app/app.go`, `cmd/prerequisites/prerequisites.go`, `cmd/update/update.go`) expose the full CLI surface documented in `../../reference/architecture/overview.md`, including:

```bash
openframe bootstrap                 # create a k3d cluster + install the platform in one step
openframe cluster create            # create a k3d, GKE, or EKS cluster
openframe cluster status            # show cluster health/nodes
openframe app install               # install ArgoCD + app-of-apps onto a cluster
openframe app status --watch        # live-refreshing platform readiness view
openframe app access                # print ArgoCD admin credentials
openframe prerequisites check       # verify required local tools are installed
```

For a narrated end-to-end walkthrough of these commands against a real cluster, see `../../getting-started/how-to/create-a-kubernetes-cluster-with-openframe-cli.md`, `../../getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md`, and `../../getting-started/how-to/check-the-status-of-your-openframe-cluster-and-app.md`.

### Global flags

Flags like `--silent`, `--verbose`, and `--plain` are wired in `cmd/root.go` and propagated into every command group. Command groups that define their own `PersistentPreRunE` (such as `app` in `cmd/app/app.go` and `cluster` in `cmd/cluster/cluster.go`) explicitly call `ui.ApplyGlobalOutputFlags(cmd)` as their first action — Cobra only runs the closest ancestor's hook, so this call is what keeps `--silent`/`--verbose` working for the whole subtree rather than being silently shadowed.

```bash
openframe app install --verbose     # detailed logs, including ArgoCD sync progress
openframe app status --output json  # machine-readable output, suppresses the logo/prompts
```

### Machine-readable output and scripting

Several read-only commands (`app access`, `app status`) support `--output json|yaml` via the shared `addOutputFlag`/`outputFormat`/`renderMachine` helpers in `cmd/app/output.go`. When `isMachineOutput(cmd)` is true, the logo and prerequisite checks are skipped so stdout stays clean for scripts — useful when you're iterating on automation or CI integration locally.

## Hot reload / watch mode

This CLI has no `package.json`/`nodemon`/Vite dev server, so "hot reload" here means the fastest practical Go edit-compile-run loop rather than a browser-refresh workflow:

- **`go run .`** is the fastest inner loop for most changes — it compiles and runs in one step without producing a binary in `build/`:

  ```bash
  go run . app status
  ```

  Re-run it after each change; Go's incremental build cache (`$GOCACHE`) keeps recompiles of unchanged packages fast.

- **If you prefer an external watcher** (e.g. a generic file-watch tool that re-runs a shell command on save), point it at `go run .` or `make build` — there is no bundled watch target in the `Makefile`, so this is left to your own editor/toolchain setup rather than a project-provided script.

- **`app status --watch`** is a *runtime* refresh mode of the CLI itself, not a development hot-reload feature: once running, it polls the cluster every 3 seconds and re-renders in place using `pterm.DefaultArea`, until Ctrl+C. It's documented here because it's easy to confuse with a dev-mode watch flag:

  ```bash
  go run . app status --watch
  ```

- **`app status --interactive`** similarly launches a live bubbletea TUI (`internal/app/status/tui`) for navigating ArgoCD applications and triggering syncs — again a runtime feature of the shipped CLI, not a development aid.

## Testing while developing

The `Makefile` separates tests by cost and blast radius — useful to know which to run in your inner loop versus before pushing:

```bash
make test-unit         # fast; . ./cmd/... ./internal/... ./tests/testutil/... ./tests/scripts/...
make test-race         # same packages, with the Go race detector (needs CGO_ENABLED=1)
make test-integration  # opt-in via the "integration" build tag — creates REAL k3d clusters
make test-tfvalidate   # opt-in via the "tfvalidate" build tag — runs real `terraform init`/`validate`
make test              # test-unit + test-integration
```

`test-unit` deliberately excludes `./tests/integration/...`, which spins up real k3d clusters and needs `docker` + `k3d` installed — the harness builds its own CLI binary (see `tests/integration/common/cli_runner.go`), so `make build` is not a prerequisite for integration tests, only Docker and k3d are. `test-tfvalidate` downloads real Terraform providers over the network and needs `terraform` installed; it validates the generated EKS/GKE root modules (`internal/cluster/providers/eks/templates/main.tf`, `internal/cluster/providers/gke/templates/main.tf`) without needing cloud credentials.

Other useful Makefile targets while iterating:

```bash
make fmt   # gofmt -w . — format in place
make vet   # go vet ./...
make lint  # golangci-lint run ./... (govet, staticcheck, errcheck, gosec, ineffassign)
make clean # remove build artifacts from build/ and the repo root
make help  # list all targets with descriptions
```

See `../testing/README.md` for more on the test suite, and `../contributing/guidelines.md` before opening a change.

## Debug configuration

There is no committed IDE launch configuration in this repository (no `.vscode/launch.json` or similar is part of the indexed source tree), so debugging is done with standard Go tooling pointed at the module's entry point, `main.go`.

### Debugging with Delve

If you use [Delve](https://github.com/go-delve/delve) directly:

```bash
dlv debug . -- app status --verbose
```

This builds a debug binary from the module root (`main.go`, which wires up `cmd.Execute()`) and runs it with the given CLI arguments, stopping at any breakpoints you set.

### Debugging from an editor

Most Go-aware editors (VS Code with the Go extension, GoLand/IntelliJ, etc.) can attach to or launch `main.go` directly using their built-in Delve integration — point the run/debug configuration at the module root (`.`) and pass whatever subcommand and flags you want to exercise (e.g. `cluster create --skip-wizard`, `app install --non-interactive`) as program arguments. Since the CLI is a single `main` package with no external services to mock for most commands, breakpoints in `cmd/` and `internal/` packages are hit directly as Cobra dispatches to the relevant `RunE` handler.

### Understanding error flow while debugging

Two conventions in this codebase matter when you're stepping through a failing command:

- **Exit-code fidelity.** `main.go`'s `exitCode(err)` helper unwraps `*executor.CommandError` (from `internal/shared/executor`) to propagate an external command's real exit code to the OS, falling back to `1` otherwise. If you're debugging why a script sees the wrong exit code, check whether the failing path returned a `*executor.CommandError` or a generic error.
- **Already-handled errors.** Errors that have already been printed to the user (via `sharedErrors.HandleGlobalError`) are wrapped in `*sharederrors.AlreadyHandledError` (package `internal/shared/errors`). `main.go` checks for this sentinel with `errors.As` before deciding whether to print `Error: %v` again — so if you see a command fail silently (no stack, no duplicate message) that's this sentinel at work, not a swallowed error.

### Debugging specific command groups

Since each command group is isolated behind a `GetXxxCmd()`/`getXxxCmd()` factory (e.g. `cmd/app/app.go`'s `GetAppCmd()`, `cmd/cluster/cluster.go`'s `GetClusterCmd()`, `cmd/bootstrap/bootstrap.go`'s `GetBootstrapCmd()`), you can set breakpoints directly in the relevant `RunE`/`PersistentPreRunE` closures for the subcommand you're investigating — for example `runInstallCommand` in `cmd/app/install.go`, or `runStatusCommand` in `cmd/app/status.go`. Use `--verbose` when launching the debug target so `pterm.Debug`-level output (gated by `ui.ApplyGlobalOutputFlags`) is visible alongside your breakpoints.

For commands that shell out to external tools (Docker, k3d, Helm, Terraform, `gcloud`, `aws`), `internal/shared/executor` provides both a real and a mock `CommandExecutor` — when debugging logic that depends on command output without wanting to actually invoke the real binaries, look at how the package's tests construct the mock executor rather than running against live tools.

## Next steps

Once your local build runs, continue with `../../getting-started/quick-start.md` and `../../getting-started/first-steps.md` for an end-to-end platform walkthrough, or `../environment.md` for environment-specific configuration details.
