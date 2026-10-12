
# Testing

OpenFrame CLI is a Go service (module `github.com/flamingo-stack/openframe-cli`) built with Cobra. Its test suite combines fast, hermetic unit/contract tests that run on every push with a heavier, self-hosted end-to-end lifecycle suite that provisions a real k3d cluster and installs the platform via ArgoCD. This document covers how the tests are organized, how to run them locally and in CI, how to write new ones, and what the repository expects before a change is considered covered.

> Per the organization's ecosystem graph, OpenFrame CLI publishes only the `go` artifact `github.com/flamingo-stack/openframe-cli` and has no recorded upstream dependencies or downstream consumers in this organization — it is a terminal artifact. Coverage of that graph is partial, so this should be read as "nothing currently depends on it that the graph can see," not as a guarantee. See `../../reference/architecture/ecosystem.md` for the full ecosystem report.

## Test structure and organization

Tests live alongside the code they exercise (standard Go convention: `_test.go` files in the same package) plus two dedicated test-support trees:

| Location | Purpose |
|---|---|
| `main_test.go`, `cli_matrix_test.go` (repo root) | Top-level entry point tests and a CLI invocation matrix run against the real built binary |
| `cmd/root_test.go`, `cmd/contract_test.go`, `cmd/help_matrix_test.go` | Root command tests, a frozen flag/subcommand contract, and an exhaustive `--help` sweep over the whole command tree |
| `cmd/app/app_test.go` | Tests for the `app` command group (`cmd/app/app.go`) |
| `internal/**/*_test.go` | Unit tests beside each package (e.g. `internal/shared/executor`, `internal/cluster`, `internal/chart`) |
| `tests/testutil/` | Shared, reusable test helpers: assertions, flag contracts, command-structure patterns, mock setup |
| `tests/integration/common/` | Integration test harness: builds/runs the CLI binary, manages real k3d clusters, gates tests on external dependencies |

### Why the split matters

- **Unit tests** (`internal/...`, `cmd/...`) use the mock command executor (`internal/shared/executor/mock.go`) so they never shell out to Docker, k3d, Helm, or Terraform. They run on every OS in CI (Linux, Windows, macOS) with no external tooling required.
- **CLI matrix tests** (`cli_matrix_test.go`) build the real binary and invoke it with a spread of arguments in a hermetic environment (isolated `HOME`, no kubeconfig, update checks and WSL forwarding disabled) to validate exit codes and output without touching a cluster, Docker, or the network.
- **Integration/e2e tests** (`tests/integration/common/`, and the `test-cli` job in `.github/workflows/test.yml`) build the CLI, create a real k3d cluster, and exercise the full `cluster create` → `app install` → `app status` → `app uninstall` lifecycle against ArgoCD on Kubernetes. These need Docker, k3d, kubectl, and Helm, and are gated accordingly.

### Shared test utilities (`tests/testutil/`)

This package is the backbone for writing new unit tests consistently:

| File | What it provides |
|---|---|
| `tests/testutil/setup.go` | `InitializeTestMode()` (enables test-safe UI rendering via `ui.TestMode`), `NewTestMockExecutor()`, `CreateStandardTestFlags()` (a `cluster.FlagContainer` wired to a mock k3d executor/manager), `CreateIntegrationTestFlags()` (real dependencies), `SetVerboseMode()` |
| `tests/testutil/assertions.go` | `AssertCommandOutput` — a fluent, chainable assertion API over stdout/stderr/err (`NoError()`, `StdoutContains()`, `StderrContainsAny()`, `ContainsAll()`, etc.), plus `AssertCommandSuccess`/`AssertCommandFailure` helpers |
| `tests/testutil/command_assertions.go` | Executor-seam security assertions against a `*executor.MockCommandExecutor`'s structured command log: `AssertNoArgContains` (secrets never land in argv), `AssertNoBashCConcatenation` / `AssertNoCurlPipeShell` (no shell-injection patterns), `AssertArgIsDiscrete`, `AssertNoDeleteOf` (protected resources like `kube-system`) |
| `tests/testutil/cluster.go` | `TestCommandStructure`, `TestCLICommand`, `TestFlags()`, `IntegrationFlags()` |
| `tests/testutil/patterns.go` | `TestClusterCommand` — runs four standard sub-tests (`Structure`, `Flags`, `CLI`, `Execution`) for a cluster command in one call |
| `tests/testutil/flag_contract.go` | `FlagSpec`, `AssertFlag`, `AssertFlags`, `FindSubcommand`, `AssertSubcommands` — freezes a command's flag/subcommand surface so a silent rename or dropped default fails a test |
| `tests/testutil/utilities.go` | `TestClusterConfig(name)` — a `*models.ClusterConfig` with sane defaults (k3d, 3 nodes, `v1.28.0`) |

### Integration harness (`tests/integration/common/`)

| File | What it provides |
|---|---|
| `tests/integration/common/cli_runner.go` | `InitializeCLI()` / `CleanupCLI()` build and remove the compiled test binary at `build/openframe` (with mod-time caching to skip rebuilds), `RunCLI(args...)` executes it and returns a `*CLIResult` with `Success()`, `Failed()`, `Output()`, `ErrorMessage()` |
| `tests/integration/common/cluster_management.go` | `GenerateTestClusterName()`, `CreateTestCluster`/`DeleteTestCluster`/`StopTestCluster`/`ClusterExists` (wrap real `k3d` commands), `CleanupTestCluster`/`CleanupAllTestClusters` (also purge leftover Docker networks/containers matching test-name patterns) |
| `tests/integration/common/dependencies.go` | `Dependency` struct + predefined `Docker`, `K3d`, `Kubectl`, `Helm` dependencies; `RequireOrSkip(t, msg)` **skips** (never fails) a test when a tool is unavailable; `RequireClusterDependencies`, `RequireK8sDependencies`, `RequireAllDependencies` convenience gates |

Using `RequireOrSkip` rather than failing outright means a developer without Docker or k3d installed locally still gets a green run for integration-dependent tests — they are clearly logged as skipped, not silently passed.

### The mock command executor

`internal/shared/executor/mock.go` implements `CommandExecutor` for tests. It is thread-safe (`sync.Mutex`-guarded) and keeps two parallel logs:

- a flattened string log (`GetExecutedCommands()`, `WasCommandExecuted(pattern)`, `GetLastCommand()`) — legacy, convenient for simple substring checks;
- a structured log of `RecordedCommand{Name, Args, Env, Stdin}` via `Commands()` — the one to prefer for security-sensitive assertions, because it preserves argv as discrete elements and can distinguish a literal `$(x)` argument from a shell-constructed string.

Typical setup:

```go
exec := executor.NewMockCommandExecutor()
exec.SetResponse("helm upgrade", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "Release deployed",
})
result, err := exec.Execute(ctx, "helm", "upgrade", "my-release", "./chart")
exec.SetShouldFail(true, "connection refused") // simulate failure
exec.Reset()                                    // clean state for the next test
```

## Running tests

All commands below assume a checkout of `https://github.com/flamingo-stack/openframe-cli` with Go installed per `go.mod` (module `github.com/flamingo-stack/openframe-cli`, `go 1.26.0`, pinned toolchain `go1.26.5`).

### Unit tests

The CI workflow (`.github/workflows/test.yml`, `unit` job) runs:

```bash
make test-unit
```

across three OS/arch legs (`linux/amd64`, `windows/amd64`, `darwin/arm64`), covering the root package, `cmd/`, `internal/`, `tests/testutil`, and `tests/scripts`. This target needs no Docker, k3d, Helm, or network access — every dependency is mocked.

### Race detector

Linux-only (the race detector needs CGO/gcc):

```bash
make test-race
```

This is a blocking gate in CI — the pterm-based spinner was specifically replaced with a synchronized wrapper (`internal/shared/ui/spinner`) to keep the codebase race-clean.

### Running a single package or test

Standard `go test` invocations work directly, for example:

```bash
go test ./internal/shared/executor/...
go test ./cmd/... -run TestRootCommand -v
```

### CLI matrix and root-level tests

`main_test.go` and `cli_matrix_test.go` build the binary themselves (`go build -o <tmp> .`) and exercise it as a subprocess, so running them only requires the Go toolchain:

```bash
go test . -run TestCLIMatrix -v
go test . -run TestMainIntegration -v
```

The matrix (`cli_matrix_test.go`) sets `OPENFRAME_NO_UPDATE_CHECK=1` and `OPENFRAME_NO_WSL_FORWARD=1` and points `KUBECONFIG` at a nonexistent path, so it behaves identically on every OS and never reaches the network or a real cluster.

### Lint and go.mod hygiene

The `lint` job in `.github/workflows/test.yml` runs `golangci-lint` (a blocking gate — the backlog is clear, so any new finding fails the pipeline) and checks tidiness:

```bash
go mod tidy -diff
```

### Full release build matrix (compile-only)

`.github/workflows/test.yml`'s `release-build-matrix` job cross-compiles all six GoReleaser platform combos (`linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`, `darwin/amd64`, `darwin/arm64`) and additionally runs `go vet` (which compiles test files too) for the three combos with no dedicated test runner, to catch OS-specific breakage in test code that would otherwise go unnoticed.

### End-to-end / integration tests

The `test-cli` job in `.github/workflows/test.yml` is the full lifecycle test: it builds the CLI, then runs it against a real k3d cluster (`OF_CLUSTER: openframe-test`, kube-context `k3d-openframe-test`), covering version/help/completion, the exhaustive `--help` walk over the whole command tree, and a real self-update/rollback cycle against the latest published release. It is gated to run only on `push` to `main`, on `workflow_dispatch`, or on a pull request targeting `main` from this same repository (not a fork) — and only after the `lint` job passes, to avoid burning an hour of self-hosted runner time on code that fails a three-minute lint.

Locally, the same lifecycle relies on the harness in `tests/integration/common/`. A typical pattern:

```go
func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatalf("setup failed: %v", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}
```

Any test that needs a real cluster should gate itself first:

```go
func TestClusterProvisioning(t *testing.T) {
    common.RequireClusterDependencies(t) // skips if Docker or k3d are unavailable
    name := common.GenerateTestClusterName()
    if err := common.CreateTestCluster(name); err != nil {
        t.Fatalf("failed to create cluster: %v", err)
    }
    defer common.CleanupTestCluster(name)
    // ...
}
```

## Writing new tests

### Command structure and contract tests

Every Cobra command group is wired through a `Get<Group>Cmd()` factory (never constructed inline) per the active code rule `OPENFRAM-003`. Use the shared helpers to validate that contract rather than hand-rolling assertions:

```go
func TestMyCommand(t *testing.T) {
    cmd := NewMyCommand()
    testutil.TestCommandStructure(t, cmd, "my-command", "Short description here")
    testutil.TestCLICommand(t, cmd, []string{"--flag", "value"}, false, "expected output")
}
```

For cluster subcommands specifically, `tests/testutil/patterns.go` runs a full standard battery in one call:

```go
func TestCreateCommand(t *testing.T) {
    testutil.TestClusterCommand(t, "create", NewCreateCommand, setupFunc, teardownFunc)
}
```

This exercises `Structure` (name/short/long/`RunE`), `Flags` (`--help` succeeds, unknown flags fail), `CLI` (argument-count validation via `cmd.Args`), and `Execution` (`--dry-run` where supported, `--help` always succeeds).

### Freezing the flag/subcommand surface

Any command whose flags or subcommands form a public contract (the root command, long-lived command groups) should have a contract test using `tests/testutil/flag_contract.go`, following the pattern in `cmd/contract_test.go`:

```go
func TestRootContract_TopLevelSubcommands(t *testing.T) {
    root := GetRootCmd(VersionInfo{Version: "t", Commit: "t", Date: "t"})
    for _, name := range []string{"cluster", "app", "bootstrap", "prerequisites", "update"} {
        testutil.FindSubcommand(t, root, name)
    }
}
```

`AssertFlags` similarly locks shorthand, type, and default value for a slice of `FlagSpec`s, so a renamed flag or changed default fails loudly instead of silently breaking automation that depends on it.

### Help-matrix style sweeps

`cmd/help_matrix_test.go` walks the **entire** command tree and asserts that every command's `--help` succeeds and prints a `Usage:` section, that every command has a non-empty `Short` description, and that every locally declared flag has non-empty usage text. New commands are covered automatically with no per-command test to add — keep this pattern in mind before writing a bespoke `--help` test for a new command.

### Security-sensitive assertions (executor seam)

Changes that shell out to external tools (Helm, kubectl, Terraform, k3d, Git, cloud CLIs via `internal/shared/executor`) should be tested at the executor seam using the mock's structured command log and `tests/testutil/command_assertions.go`:

```go
func TestHelmInstall_CredentialSafety(t *testing.T) {
    mock := executor.NewMockCommandExecutor()
    // ... run the CLI code under test using mock ...
    cmds := mock.Commands()

    testutil.AssertNoArgContains(t, cmds, "ghp_secrettoken123")   // secret never in argv
    testutil.AssertArgIsDiscrete(t, cmds, "my-release")            // passed verbatim, not concatenated
    testutil.AssertNoBashCConcatenation(t, cmds, "my-release")     // no injected shell metachars
    testutil.AssertNoDeleteOf(t, cmds, "kube-system")               // protected namespace guard
    testutil.AssertNoCurlPipeShell(t, cmds)                         // no curl|bash patterns
}
```

Prefer `Commands()` (structured `[]RecordedCommand`) over `GetExecutedCommands()` (flattened strings) whenever a test needs to prove something about *how* a command was constructed — the flattened log cannot distinguish a literal `$(x)` argument from a shell-built string.

### Output assertions

For stdout/stderr/error checks on command output, use the fluent `AssertCommandOutput` API rather than ad hoc `strings.Contains` calls, so failures report consistently:

```go
func TestMyCommand(t *testing.T) {
    stdout, stderr, err := runCommand("openframe", "version")
    testutil.AssertCommandSuccess(t, stdout, stderr, err).
        StdoutContains("openframe").
        ContainsAll("version", "build")
}
```

### Conventions to follow

- Initialize test-safe UI rendering once per package with `testutil.InitializeTestMode()` (sets `ui.TestMode = true`) — several existing test files do this in an `init()` function (see `cmd/root_test.go`, `cmd/app/app_test.go`).
- Use `CreateStandardTestFlags()` for unit tests (mocked k3d executor/manager, pre-wired responses for `k3d cluster list`/`k3d cluster get`) and `CreateIntegrationTestFlags()` only when the test is explicitly an integration test that should hit the real k3d manager.
- Command groups that define their own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)` (active rule `OPENFRAM-005-10`) — `cmd/app/app_test.go`'s `TestAppGroup_PersistentPreRunEHonorsVerbose` is the regression test pattern to follow when adding a new command group with its own hook.
- Commands that target a cluster via a flag (e.g. `--context`) must reject stray positional arguments with a named-flag hint (`noPositionalArgs` in `cmd/app/app.go`, rule `OPENFRAM-009-8`) — write a test asserting the rejection message when adding a similar command.
- External-command errors must preserve the real exit code via `executor.CommandError` rather than collapsing to a generic failure (rule `OPENFRAM-005`); `main_test.go`'s `TestExitCode` is the pattern for testing this end to end, including the wrapped-error case (`fmt.Errorf("...: %w", err)`).
- Errors already shown to the user must be/wrap `sharederrors.AlreadyHandledError` so `main` does not double-print (rule `OPENFRAM-004`) — cover new error-display paths with a test that checks for this sentinel via `errors.As`.
- A new or changed exported definition should be referenced by at least one test (organization rule `TEST-001`, informational severity) — when adding an exported function, type, or constant, add or extend a test that exercises it, or be ready to explain in the PR why it is not testable.
- Avoid introducing a new exported definition that duplicates an existing one elsewhere in the organization's code graph (rule `DUP-001`); if in doubt, search for an existing helper first rather than writing a near-identical one in `tests/testutil`.

## Coverage requirements

This repository does not enforce a numeric coverage percentage threshold in the material available for this document. Instead, coverage is enforced structurally through CI **gates** that must all pass before code reaches `main` or before a release is cut:

| Gate | Where | What it blocks |
|---|---|---|
| `lint` | `.github/workflows/test.yml` | `golangci-lint` findings and un-tidy `go.mod`/`go.sum` |
| `unit` | `.github/workflows/test.yml` | `make test-unit` failures on any of linux/windows/darwin |
| race detector | `.github/workflows/test.yml` (`unit` job, linux only) | `make test-race` failures |
| `release-build-matrix` | `.github/workflows/test.yml` | compile/`go vet` failures across all six GoReleaser platform combos |
| `test-cli` (e2e) | `.github/workflows/test.yml`, main-bound | full lifecycle failures against a real k3d + ArgoCD cluster |
| `Gate: lint` / `Gate: unit tests` | `.github/workflows/release.yml` | a release is only cut if `golangci-lint` and `make test-unit` both pass first |
| post-release smoke tests | `.github/workflows/release.yml` | the released binary must report the injected version and its published asset URLs must resolve |
| `missing-test` analyzer (`TEST-001`) | organization-wide active code rule | flags new/changed exported definitions with no referencing test (informational, not a hard failure) |
| `dup-scan` analyzer (`DUP-001`) | organization-wide active code rule | flags a new/changed exported definition that duplicates an existing one's normalized structure |

In practice this means: every exported symbol you add or change should be exercised by at least one test (unit, contract, or matrix-style), every command group must pass its structure/flags/CLI/execution battery, and any change touching how an external tool is invoked should carry an executor-seam security assertion (`tests/testutil/command_assertions.go`) proving no secret leaks into argv and no shell-injection pattern is introduced. PRs are also expected to follow Conventional Commit titles (enforced by `.github/workflows/pr-title.yml`) since `semantic-release` derives release versions from them — this is unrelated to test coverage directly, but is part of the same merge gate a change must clear alongside its tests.

For the wider system this CLI operates within, see `../../reference/architecture/overview.md` and `../../reference/architecture/ecosystem.md`; for local environment setup before running these tests, see `../setup/environment.md` and `../setup/local-development.md`.
