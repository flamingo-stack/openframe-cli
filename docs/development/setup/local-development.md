# Local Development

This guide walks through cloning, building, running, and debugging OpenFrame CLI locally.

## Clone the Repository

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
```

## Build

The CLI's entry point is `main.go` at the repository root (`package main`), which calls `cmd.Execute()`.

```bash
# Build a local binary
go build -o build/openframe .

# Or run directly without a persistent binary
go run . --help
```

Version metadata (`version`, `commit`, `date`) is normally injected at release build time via `-ldflags -X`. For local dev builds without ldflags, `cmd.resolveVersionInfo` falls back to Go's embedded VCS build info, so `openframe --version` still reports a real commit/date instead of placeholder values.

## Running Locally

Once built, run subcommands exactly as an end user would:

```bash
./build/openframe --help
./build/openframe prerequisites check
./build/openframe cluster create --skip-wizard
./build/openframe bootstrap --non-interactive
```

Use `--verbose` on any command to see detailed logs, including shelled-out command output (e.g., ArgoCD sync progress, Terraform plan output):

```bash
./build/openframe cluster create --verbose
```

Use `--silent` to suppress non-essential UI (spinners, logo) or `--plain` for non-ANSI output suitable for logs/CI.

## Iterating Quickly

For fast iteration without constantly re-running `go build`, use `go run`:

```bash
go run . cluster status
```

The integration test harness (`tests/integration/common`) builds the binary once into `build/openframe` and skips rebuilds when the binary is newer than `main.go` — the same pattern works well for manual iteration: rebuild only when source changes.

## Debugging

Since this is a standard Go CLI built with Cobra, you can debug it with any Go-compatible debugger:

- **Delve** (`dlv`):

```bash
dlv debug . -- cluster create --skip-wizard --verbose
```

- **VS Code**: create a `launch.json` configuration of type `go`, with `program` set to the repository root and `args` set to your desired subcommand (e.g., `["app", "status", "--verbose"]`).

Because most external interactions (Docker, k3d, Helm, Terraform, cloud CLIs) go through the `internal/shared/executor.CommandExecutor` abstraction, you can also write unit tests against a `MockCommandExecutor` to reproduce a failure without needing a real cluster — see the [Testing](../testing/README.md) guide.

## Working Without a Real Cluster

Many code paths can be exercised without a live cluster or cloud account by using the test utilities in `tests/testutil` (e.g., `CreateStandardTestFlags()`, which wires up a `MockCommandExecutor` with canned k3d responses). This is the fastest way to iterate on flag parsing, validation, and orchestration logic.

For true end-to-end verification, run against a real local cluster:

```bash
./build/openframe bootstrap my-dev-cluster
./build/openframe app status --watch
./build/openframe cluster delete my-dev-cluster
```
