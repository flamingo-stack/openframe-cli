## Local Development

This guide covers cloning OpenFrame CLI, building and running it locally, iterating with watch/rebuild workflows, and debugging the Go binary. OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a standalone Go service — a bootstrap and lifecycle-management tool for the OpenFrame platform. It is not a library imported by other repositories: per the organization's dependency graph (coverage: full), it has no recorded upstream dependencies on other managed repositories and no recorded downstream consumers. It depends only on third-party Go modules and shells out to external CLI tools (Docker, k3d, Helm, Terraform, gcloud, aws).

> The platform this CLI deploys lives in a separate repository, [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant). This document covers only the `openframe-cli` codebase itself — cloning its app-of-apps chart at install time is a runtime operation (see `internal/chart/models/repos.go`'s `RepoOSSTenant` constant), not a local development dependency.

### Prerequisites

Before working on this repository, see `../../getting-started/prerequisites.md` for the full hardware and tool requirements. In short, you need:

- **Go 1.26.5** — the `go.mod` pins `go 1.26.0` with `toolchain go1.26.5`. The toolchain pin matters: Go 1.26.0/1.26.1 crash intermittently on windows-amd64 during GC stack scanning, and 1.26.5 is the first release with the fix. Let `go` resolve the toolchain automatically rather than overriding it.
- **Docker** — required to run local k3d clusters exercised by `openframe cluster create` and the integration tests.
- **git** — for cloning and for the CLI's own use of `go-git` when fetching the app-of-apps chart repository.
- Optional, only if you touch specific areas: `golangci-lint` (lint target), `k3d`/`helm`/`terraform` (integration tests and `test-tfvalidate`), `gcloud`/`aws` CLIs (cloud provider code paths).

### Clone and setup

Clone the fork, not an upstream or placeholder URL:

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

Download the module dependencies (this also respects the `toolchain` directive in `go.mod` and fetches Go 1.26.5 if your local toolchain is older):

```bash
go mod download
```

Verify `go.mod`/`go.sum` are tidy before you start (the same check CI performs):

```bash
make tidy
```

`make tidy` runs `go mod tidy -diff` and fails if `go mod tidy` would change the files — use it after adding or removing imports.

### Building

The `Makefile` is the primary entry point for local builds. It defines a `CGO_ENABLED=0 go build -trimpath` base (`GO_BUILD`), matching the reproducible release build in the project's GoReleaser configuration.

Build a binary for your current platform into `build/`:

```bash
make build
```

This produces `build/openframe-$(GOOS)-$(GOARCH)` (with a `.exe` suffix added automatically on Windows), where `GOOS`/`GOARCH` are detected from `go env`.

To cross-compile every release platform at once (matches `.goreleaser.yml`):

```bash
make build-all
```

This builds `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`, and `windows/arm64` binaries into `build/`.

Remove build artifacts when you need a clean slate:

```bash
make clean
```

`make clean` is scoped to only match platform-suffixed binaries (`openframe-linux-*`, `openframe-darwin-*`, `openframe-windows-*`, plus the bare `openframe` and anything already in `build/`), so it does not touch other tracked files such as example Helm values files.

List every available target with descriptions:

```bash
make help
```

### Running locally

Once built, run the binary directly:

```bash
./build/openframe-$(go env GOOS)-$(go env GOARCH) --help
```

Or run straight from source without a separate build step, which is the fastest loop while iterating on command logic:

```bash
go run . --help
go run . bootstrap my-cluster --non-interactive
```

`main.go` is the process entrypoint. It calls `cmd.Execute()` (the root Cobra command defined in `cmd/root.go`) and, on error, decides what to print and what exit code to return:

- If the returned error wraps `sharederrors.AlreadyHandledError`, `main` does **not** re-print it — the message was already shown to the user by the command's own error handler (see `internal/shared/errors/`).
- Otherwise it prints `Error: <err>` to stderr.
- `exitCode(err)` preserves a failed external command's real exit code via `executor.CommandError` (when it is a valid, non-zero code under 256), so automation wrapping the CLI sees the same exit code a raw `helm`/`k3d`/`terraform` invocation would have produced, not a generic `1`.

Representative top-level commands you can exercise locally (see `docs/reference/architecture/overview.md` for the full architecture and command table):

```bash
openframe bootstrap [cluster-name]         # create a local k3d cluster + install the platform
openframe cluster create [NAME]            # create a k3d/GKE/EKS cluster
openframe cluster status [NAME]            # cluster health/resource info
openframe app install [cluster-name]       # install ArgoCD + app-of-apps onto a cluster
openframe app status [--watch|--interactive]
openframe prerequisites check|install
```

For the k3d-based local flow end-to-end (creating a cluster and installing the platform with this CLI), see `../../getting-started/how-to/create-a-kubernetes-cluster-with-openframe-cli.md` and `../../getting-started/how-to/install-the-openframe-platform-onto-a-cluster.md`.

> **Note on scope:** commands like `cluster create` for cloud types (EKS/GKE) shell out to Terraform and cloud CLIs, and will attempt to provision real, billed infrastructure. For local development, prefer the k3d path (`openframe cluster create` with the default/`k3d` type, or `openframe bootstrap`) so you are not charged for cloud resources while iterating.

### Hot reload / watch mode

OpenFrame CLI is a Go CLI binary, not a web server with a dev server — there is no bundler-style hot-module-reload here. The fastest edit-run loop the Makefile and `go` toolchain give you is:

```bash
go run . <subcommand> [flags]
```

`go run .` recompiles on every invocation, so each change to `cmd/` or `internal/` is picked up immediately without a separate `make build` step. This is the recommended loop for iterating on a single command's logic (flag parsing, prompts, output formatting).

For a tighter loop while touching a specific package, run only that package's tests instead of the whole binary — see `make test-unit` below, and scope it with `go test ./internal/chart/services/...` (or any other package path from the source tree) while developing.

There is no file-watcher integration (e.g. `gow`, `air`, `reflex`) configured in this repository's `Makefile` or `go.mod`; if you want automatic re-run on save, wrap `go run .` or `make test-unit` in your own watch loop using whichever tool you already have installed — nothing here requires a specific one.

### Testing during development

The `Makefile` defines several test targets with different scopes and costs — understanding the split matters for fast local iteration:

```bash
make test-unit        # fast: root + cmd/... + internal/... + tests/testutil/... + tests/scripts/...
make test-race        # same unit set, with -race (needs CGO_ENABLED=1)
make test-integration # slow: real k3d clusters via docker + k3d (tests/integration/...)
make test-tfvalidate  # slow: terraform init + validate against generated EKS/GKE templates (needs network)
make test             # test-unit + test-integration
```

`test-unit` runs against the explicit package set `. ./cmd/... ./internal/... ./tests/testutil/... ./tests/scripts/...` — this deliberately includes the root package (whose `main_test.go` is the only place exit-code fidelity is tested) and excludes `./tests/integration/...`, which creates real clusters and must never run by accident from a bare `go test ./...`.

`test-integration` and `test-tfvalidate` are opt-in via Go build tags (`integration`, `tfvalidate` respectively) for exactly this reason — they are not part of the default test run. The integration harness builds its own CLI binary (see `tests/integration/common/cli_runner.go`), so `make build` is not a prerequisite for `make test-integration`, but Docker and k3d must be installed.

Other code-quality targets:

```bash
make lint   # golangci-lint run ./... (govet, staticcheck, errcheck, gosec, ineffassign)
make fmt    # gofmt -l -w . — formats in place
make vet    # go vet ./...
```

`make lint` requires `golangci-lint` to be installed locally; the Makefile fails fast with an install link if it is missing.

### Debug configuration

There is no committed IDE launch configuration (for example, no VS Code `.vscode/launch.json` or GoLand run configuration) in this repository's file tree — debugging is driven directly through the `go` toolchain and standard Go tooling.

**Debugging with `delve`:** since this is a plain Go module (no special build constraints beyond the opt-in test tags), you can attach `dlv` directly to any command:

```bash
dlv debug . -- bootstrap my-cluster --non-interactive
dlv debug . -- app install --context k3d-my-cluster
dlv test ./internal/chart/services/... -- -run TestInstallCharts
```

**Verbose/debug output built into the CLI itself:** most commands accept a root-level `-v`/`--verbose` flag that surfaces detailed logs, including ArgoCD sync progress during install/upgrade/bootstrap flows:

```bash
openframe bootstrap --verbose
openframe app install --non-interactive --verbose
```

Several command groups also apply their own output-mode contract on top of `--verbose`:

- `--silent` suppresses interactive/decorative output.
- `--plain` disables ANSI/color rendering (useful when piping output or debugging in a non-TTY context).
- `--output json|yaml` switches to machine-readable output and disables the logo/prompts entirely — useful when scripting a debug session around the CLI (see `cmd/app/output.go`'s `outputFormat`/`isMachineOutput` helpers, shared across the `app` subcommands).

Note that `app`'s own `PersistentPreRunE` re-applies the global `--silent`/`--verbose` contract explicitly (see `cmd/app/app.go`) because Cobra only runs the closest parent's `PersistentPreRunE` — defining one on a command group shadows the root's. If you add a new command group and need consistent `--verbose` behavior, follow the same pattern rather than assuming the root hook fires.

**Mock-based debugging for business logic:** rather than exercising real external tools while debugging, the executor abstraction (`internal/shared/executor`) lets you substitute a `MockCommandExecutor` for the real one in tests, so you can step through orchestration logic (chart install sequencing, ArgoCD wait loops, cluster provider dispatch) without needing Docker, k3d, Terraform, or cloud credentials present. This is the same seam the test suite uses — look at how a given package's tests construct their executor to reuse the pattern in a debug session.

**Error visibility while debugging:** be aware of the `AlreadyHandledError` sentinel (`internal/shared/errors`) when stepping through error paths — an error that has already been printed to the user by `sharedErrors.HandleGlobalError` is wrapped in this sentinel before returning to `main`, which checks for it via `errors.As` and deliberately does not print it again. If you are debugging "why didn't my error message show up twice," this is why; conversely, any new error-handling path you add should preserve this convention to avoid double-printing.

**Exit codes while debugging automation:** when debugging why a wrapping script received an unexpected exit code, check whether the failing path returns an `*executor.CommandError` — `main.exitCode()` only propagates the original external command's exit code when the error can be unwrapped to that type (and the code is in `1`-`255`); any other error path reports a generic `1`.

### Where to go next

For an overview of the module layout and how commands route into domain services and providers, see `../architecture/README.md` and `../../reference/architecture/overview.md`. For environment-specific configuration details beyond local setup, see `./environment.md`. Contribution conventions (commit style, review expectations) are in `../contributing/guidelines.md`, and the full test suite's structure is covered in `../testing/README.md`.
