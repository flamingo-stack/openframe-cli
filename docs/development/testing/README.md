## Testing

OpenFrame CLI is a Go service (module `github.com/flamingo-stack/openframe-cli`), and its test suite is the primary safety net for a tool that provisions real Kubernetes clusters, shells out to Docker/k3d/Helm/Terraform/cloud CLIs, and self-updates signed binaries in place. This document covers how tests are organized, how to run them locally and in CI, how to write new ones, and what is expected before a change merges.

> Per the ecosystem graph, OpenFrame CLI publishes `github.com/flamingo-stack/openframe-cli` to the `go` ecosystem and has no recorded downstream consumers in this organization (coverage is partial, so this is not proof nothing imports it — but its `internal/` packages are treated as private implementation detail regardless). There are no upstream repository dependencies to mock in tests; all external dependencies are third-party Go modules and CLI tools.

### Test structure and organization

Tests live alongside the code they exercise (standard Go convention: `_test.go` files in the same package) plus two dedicated top-level test trees for shared helpers and black-box integration coverage.

| Layer | Location | What it covers |
|---|---|---|
| Unit tests | `*_test.go` beside each package (`cmd/`, `internal/...`) | Package-level logic, using `internal/shared/executor`'s mock instead of real binaries |
| CLI matrix (hermetic, real binary) | `cli_matrix_test.go`, `main_test.go` (root package) | Builds the actual `openframe` binary and runs it with closed stdin, isolated `$HOME`/`$KUBECONFIG`, no network dependency |
| Command-tree help sweep | `cmd/help_matrix_test.go` | Walks the entire Cobra command tree and asserts every `--help` works and every flag has usage text |
| CLI contract tests | `cmd/contract_test.go` | Freezes the root command's persistent flags and top-level subcommand list |
| Shared test utilities | `tests/testutil/` | Reusable assertion helpers, flag-container builders, and Cobra test patterns used across unit tests |
| Integration test harness | `tests/integration/common/` | Builds/runs the real CLI binary and manages real k3d clusters for heavier, opt-in integration scenarios |

#### Key files in the test harness

- `tests/testutil/setup.go` — `InitializeTestMode()`, `NewTestMockExecutor()`, `CreateStandardTestFlags()` (mocked k3d executor/manager) and `CreateIntegrationTestFlags()` (real dependencies).
- `tests/testutil/assertions.go` — fluent `AssertCommandOutput` (`NoError`, `StdoutContains`, `StderrContainsAny`, `ContainsAll`, …) plus `AssertCommandSuccess` / `AssertCommandFailure` shorthands.
- `tests/testutil/command_assertions.go` — security-focused assertions over a `*executor.MockCommandExecutor`'s structured command log: `AssertNoArgContains` (secrets never on argv), `AssertNoBashCConcatenation` and `AssertArgIsDiscrete` (no shell-string injection), `AssertNoDeleteOf` (protected resources like `kube-system` are never targeted), `AssertNoCurlPipeShell` (`curl | bash` patterns are rejected).
- `tests/testutil/flag_contract.go` — `FlagSpec`, `AssertFlag(s)`, `FindSubcommand`, `AssertSubcommands`: freezes a command's flag names, shorthands, types and defaults so a silent breaking change fails loudly.
- `tests/testutil/patterns.go` — `TestClusterCommand` runs four standard sub-tests (`Structure`, `Flags`, `CLI`, `Execution`) against any cluster subcommand factory.
- `tests/testutil/cluster.go` / `tests/testutil/utilities.go` — `TestCommandStructure`, `TestCLICommand`, `TestFlags`, `IntegrationFlags`, `TestClusterConfig`.
- `internal/shared/executor/mock.go` — `MockCommandExecutor`, the thread-safe fake behind nearly every unit test that would otherwise shell out. It records both a flattened command string (legacy) and a structured `RecordedCommand{Name, Args, Env, Stdin}` log; prefer `Commands()` over `GetExecutedCommands()` when asserting argv-level security properties.
- `tests/integration/common/cli_runner.go` — `InitializeCLI()` / `CleanupCLI()` build and remove a real `build/openframe` binary; `RunCLI(args...)` executes it and returns a `CLIResult` with `Stdout`, `Stderr`, `ExitCode`, `Error`, plus `Success()`, `Failed()`, `Output()`, `ErrorMessage()`.
- `tests/integration/common/cluster_management.go` — real k3d cluster lifecycle helpers (`CreateTestCluster`, `DeleteTestCluster`, `ClusterExists`, `CleanupAllTestClusters`) and Docker resource cleanup for integration suites.
- `tests/integration/common/dependencies.go` — `Dependency` (Docker, K3d, Kubectl, Helm) with `IsAvailable()` and `RequireOrSkip(t, msg)`, plus `RequireClusterDependencies`/`RequireK8sDependencies`/`RequireAllDependencies` so a test *skips* (never silently passes) when a tool is missing.

#### Design principles reflected in the structure

- **Executor seam, not real processes.** Nearly everything that would otherwise invoke Docker, k3d, Helm, Terraform or a cloud CLI goes through the `CommandExecutor` interface (`internal/shared/executor`), so unit tests inject `MockCommandExecutor` and assert on *how* a command was built rather than running it.
- **Argv-level security assertions.** Because the mock keeps arguments as a discrete `[]string` instead of a flattened string, tests can prove a secret never reached argv, that a value was passed as a standalone argument rather than spliced into a `bash -c` string, and that destructive operations never target protected resources like `kube-system`.
- **Hermetic CLI matrix.** `cli_matrix_test.go` builds the real binary once per test process and runs it with a temp `$HOME`, a non-existent `$KUBECONFIG`, `OPENFRAME_NO_UPDATE_CHECK=1` and `OPENFRAME_NO_WSL_FORWARD=1` — so the exact same matrix behaves identically on native Windows and Linux/macOS, with no live cluster or network call required.
- **Contract freezing.** `cmd/contract_test.go` and `tests/testutil/flag_contract.go` exist specifically to make a renamed flag, dropped shorthand, changed default, or added/removed subcommand fail a test immediately, rather than surfacing as a silent breaking change for CLI users or CI/automation.
- **Exhaustive help coverage with no maintenance burden.** `cmd/help_matrix_test.go` walks the live command tree (`collectCommandPaths`) rather than a hand-maintained list, so a newly added command is covered automatically.

### Running tests

All commands below assume a checkout of this repository (`git clone https://github.com/flamingo-stack/openframe-cli.git`) with the Go toolchain pinned in `go.mod` (`go 1.26.0`, toolchain `go1.26.5` — required on Windows/amd64 due to a GC stack-scanning crash fixed only from 1.26.2 onward).

#### Unit tests

```bash
go test ./...
```

This runs every package's unit tests, including the hermetic `cli_matrix_test.go` and `main_test.go` at the module root, `cmd/help_matrix_test.go`, `cmd/contract_test.go`, and all package-local `_test.go` files under `internal/`. None of these need Docker, k3d, Helm, Terraform, or network access.

To scope to a package:

```bash
go test ./cmd/...
go test ./internal/shared/executor/...
```

#### Unit tests with the race detector

CI runs the race detector on Linux (it needs CGO/gcc):

```bash
go test -race ./...
```

The pterm-based spinner in `internal/shared/ui/spinner` was specifically rewritten with a synchronized wrapper to keep the codebase race-clean under this check — a regression here is a blocking CI failure.

#### Linting and module hygiene

```bash
golangci-lint run ./...
go mod tidy -diff
```

`golangci-lint` is a blocking gate in CI (`.github/workflows/test.yml`); the audit backlog is clear, so any new finding fails the pipeline. `go mod tidy -diff` verifies `go.mod`/`go.sum` are tidy without mutating them.

#### Full release-matrix compile check

The release ships six `GOOS`/`GOARCH` combinations. Locally, you can sanity-check a target the same way CI's `release-build-matrix` job does:

```bash
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build ./...
GOOS=windows GOARCH=arm64 go vet ./...
```

#### Integration tests

Integration-style tests that provision real k3d clusters or run the compiled binary end-to-end rely on `tests/integration/common` and `tests/testutil`'s `CreateIntegrationTestFlags()` / `IntegrationFlags()`. They require Docker, k3d, kubectl and/or Helm to be installed and running locally — each dependency is probed via `tests/integration/common/dependencies.go`, and a test **skips** (via `t.Skip`, not a silent pass) when its dependency is unavailable, logging the `InstallMsg` guidance. There is no separate `make` target for this tier documented in this repository's material; run the relevant package's tests directly with `go test`, with Docker and k3d available on the `$PATH`.

#### What CI runs (`.github/workflows/test.yml`)

Pull requests against `main` and `develop` (and pushes to them) trigger four jobs:

| Job | What it does |
|---|---|
| `lint` | `golangci-lint` plus `go mod tidy -diff` |
| `release-build-matrix` | Compile-only check across all six GoReleaser `GOOS`/`GOARCH` combinations; `go vet` (which also compiles test files) for the three combinations with no dedicated runner |
| `unit` | `make test-unit` on Linux, Windows and macOS runners, plus `make test-race` on Linux |
| `test-cli` | The full end-to-end lifecycle (real k3d cluster, ArgoCD install, self-update against the real latest release) — gated on `lint` passing, and restricted to `main`-targeted runs or a manual dispatch |

`make test-unit` and `make test-race` are the Makefile targets CI invokes; they wrap the `go test ./...` and `go test -race ./...` invocations shown above for the relevant package scope (root, `cmd`, `internal`, `tests/testutil`, `tests/scripts`). The release workflow (`.github/workflows/release.yml`) also runs `make test-unit` as a release gate before any tag is cut, alongside `golangci-lint`.

The `test-cli` job additionally exercises:
- `--version` / `--help` / shell-completion generation for `bash`, `zsh`, `fish`, `powershell`;
- an exhaustive `--help` walk over the entire real binary's command tree (dynamic discovery, same philosophy as `cmd/help_matrix_test.go` but against the staged binary rather than the in-process command tree);
- `openframe update` against the real latest GitHub release, verifying the cosign-signed trust chain end to end, followed by rollback.

### Writing new tests

#### Prefer the mock executor over real binaries

For any code that shells out via `internal/shared/executor.CommandExecutor`, inject `executor.NewMockCommandExecutor()` (directly, or via `testutil.NewTestMockExecutor()` / `testutil.CreateStandardTestFlags()`), configure expected responses with `SetResponse(pattern, result)`, and assert on the recorded commands:

```go
exec := executor.NewMockCommandExecutor()

exec.SetResponse("helm upgrade", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "Release deployed",
})

result, err := exec.Execute(ctx, "helm", "upgrade", "my-release", "./chart")

// Inspect structured argv rather than the flattened string when the
// assertion is security-sensitive (e.g. no shell injection via $(...)).
for _, cmd := range exec.Commands() {
    for _, arg := range cmd.Args {
        if strings.Contains(arg, "$(") {
            t.Errorf("shell injection in argv: %q", arg)
        }
    }
}
```

Use `exec.SetShouldFail(true, "connection refused")` to simulate failure paths, and `exec.Reset()` between subtests sharing one mock.

#### Use the security-assertion helpers for anything that builds a shell command

When a change touches how the CLI constructs an external command (especially anything that might ever reach `bash -c`/`sh -c`, handle credentials, or delete cluster resources), use `tests/testutil/command_assertions.go` against the mock's `Commands()` log:

```go
cmds := mock.Commands()

testutil.AssertNoArgContains(t, cmds, "ghp_secrettoken123") // secret never in argv
testutil.AssertArgIsDiscrete(t, cmds, "my-release")          // passed verbatim, not concatenated
testutil.AssertNoBashCConcatenation(t, cmds, "my-release")   // no metacharacters reaching a shell
testutil.AssertNoDeleteOf(t, cmds, "kube-system")            // protected namespace never deleted
testutil.AssertNoCurlPipeShell(t, cmds)                      // no curl|bash pattern
```

#### Use the Cobra test helpers for new commands

New commands under `cmd/` should be exercised with the existing contract and structure helpers rather than hand-rolled Cobra plumbing:

```go
func TestMyCommand(t *testing.T) {
    cmd := NewMyCommand()

    testutil.TestCommandStructure(t, cmd, "my-command", "Short description here")

    testutil.TestCLICommand(t, cmd,
        []string{"--flag", "value"},
        false,              // expectErr
        "expected output",  // output substrings
    )
}
```

For a full cluster-command surface (structure, flags, CLI argument validation, execution with `--dry-run`/`--help`), use the four-in-one runner:

```go
func TestCreateCommand(t *testing.T) {
    testutil.TestClusterCommand(
        t,
        "create",
        NewCreateCommand,
        func() { /* setup: mocks/env */ },
        func() { /* teardown */ },
    )
}
```

When a command (or a flag's name, shorthand, type, or default) is part of the CLI's public surface, lock it with `tests/testutil/flag_contract.go`:

```go
func TestRootCommandContract(t *testing.T) {
    cmd := NewRootCmd()

    testutil.AssertSubcommands(t, cmd, "install", "upgrade", "version")

    installCmd := testutil.FindSubcommand(t, cmd, "install")
    testutil.AssertFlags(t, installCmd, []testutil.FlagSpec{
        {Name: "namespace", Shorthand: "n", Type: "string", Default: "default"},
        {Name: "dry-run", Shorthand: "", Type: "bool", Default: "false"},
    })
}
```

`cmd/contract_test.go` already does this for the root command's persistent flags (`--verbose`, `--silent`) and top-level subcommands (`cluster`, `app`, `bootstrap`, `prerequisites`, `update`) — extend it, don't duplicate it, when the root surface changes.

#### Flag containers for cluster-related tests

Use `testutil.CreateStandardTestFlags()` (mocked k3d executor and manager, pre-configured for an empty cluster list) for unit tests, and `testutil.CreateIntegrationTestFlags()` for tests meant to exercise the real k3d manager:

```go
func TestMyFeature(t *testing.T) {
    testutil.InitializeTestMode() // ui.TestMode = true, suppresses logo rendering

    flags := testutil.CreateStandardTestFlags()
    testutil.SetVerboseMode(flags, true)

    result, err := mycommand.Run(flags)
    // assert...
}
```

`testutil.TestClusterConfig(name)` returns a populated `*models.ClusterConfig` (k3d type, 3 nodes, Kubernetes `v1.28.0`) for tests that need a baseline cluster configuration object.

#### Adding to the help/command-tree sweep

New commands require no additional test registration for `--help` coverage: `cmd/help_matrix_test.go` discovers the full tree via `collectCommandPaths`. What you must still provide yourself is a non-empty `Short` description and usage text on every locally declared flag — both are asserted automatically (`TestEveryCommand_HasShortAndFlagUsages`) and will fail the suite if missing.

#### Command-group conventions that affect testability

Per this repository's active review rules:
- Each Cobra command group must be returned by a `GetXxxCmd()` factory (exported) or `get<Sub>Cmd()` (unexported, leaf commands) — never constructed as inline `cobra.Command` literals in the root wiring. This is what lets tests call the factory directly (as `cmd/app/app_test.go` does with `GetAppCmd()`).
- A command group that defines its own `PersistentPreRunE` shadows the root's hook (Cobra only runs the closest ancestor's) and must call `ui.ApplyGlobalOutputFlags(cmd)` itself, or `--verbose`/`--silent` silently stop working for that subtree. `cmd/app/app_test.go`'s `TestAppGroup_PersistentPreRunEHonorsVerbose` is the regression test pattern to follow for any new command group.
- Commands that target a cluster via a flag (e.g. `--context`) must reject stray positional arguments with a named-flag hint, via a validator like `noPositionalArgs` in `cmd/app/app.go` — write a test asserting the rejection message names the correct flag.

#### Hermetic CLI-level tests

If a new behavior is best verified against the actual compiled binary (exit codes, flag parsing edge cases, machine-readable output), add a case to the table in `cli_matrix_test.go` rather than writing a new harness:

```go
{"non-numeric --nodes", []string{"cluster", "create", "x", "--nodes", "abc"}, 1,
    []string{`invalid argument "abc"`}, nil},
```

Keep such cases hermetic: no Docker/k3d/Helm, no network, no interactive prompts (stdin is closed by `runMatrix`). Anything that genuinely needs real infrastructure belongs in the `test-cli` job of `.github/workflows/test.yml`, not in this file.

#### Error-handling and exit-code tests

`main_test.go`'s `TestExitCode` is the template for testing anything that produces or consumes exit codes: it asserts that a bare error falls back to `1`, while a wrapped `*executor.CommandError` (within range `1`–`255`) has its exit code preserved end to end — mirroring the `OPENFRAM-005` convention that external-command exit codes must never be collapsed to a generic failure. Similarly, any new error path that has already printed output to the user must return (or wrap) `*sharederrors.AlreadyHandledError`, matching `main.go`'s check in `cmd.Execute()`'s error handling, so `main` doesn't double-print it.

### Coverage requirements

This repository's testing strategy relies on CI gates rather than a single numeric coverage threshold enforced by a config file. In practice, coverage is enforced through several concrete, blocking mechanisms:

- **`golangci-lint run ./...` must report zero issues.** This is a hard gate in both `.github/workflows/test.yml` and the release workflow — any new lint finding blocks the merge or the release.
- **`go test ./...` (`make test-unit`) must pass on Linux, Windows, and macOS.** The `unit` job in `.github/workflows/test.yml` runs this matrix; a failure on any OS blocks the PR.
- **`go test -race ./...` (`make test-race`) must pass on Linux.** The codebase is kept race-clean deliberately (see the spinner rewrite noted above); a new race blocks CI.
- **Every current and future CLI command must answer `--help` correctly and carry a non-empty `Short` description and flag usage text.** `cmd/help_matrix_test.go` enforces this structurally for the whole tree, so this requirement is automatically satisfied for new commands with no extra test-writing — but it will fail if you skip `Short`/usage text.
- **The CLI's public surface (persistent flags, top-level subcommands, individual command flag specs) is frozen by contract tests.** `cmd/contract_test.go` and the `tests/testutil/flag_contract.go` helpers must be updated deliberately whenever the surface intentionally changes; an unintentional change fails the test.
- **The active code-review rule `TEST-001`** (info severity) states that a new or changed exported definition should be referenced by a test; a new exported symbol with no test reference is flagged during review, and the reviewer expects either a test or a stated reason it is not testable.
- **The active code-review rule `DUP-001`** (warn severity) applies to new or changed exported definitions, including test helpers: before adding a new assertion or setup helper, check whether `tests/testutil/` already provides an equivalent (e.g. `assertions.go`, `command_assertions.go`, `flag_contract.go`) rather than duplicating it.
- **Release gating.** `.github/workflows/release.yml` re-runs `golangci-lint` and `make test-unit` as its own gate before any tag is cut or binaries are signed — a failure here blocks the release regardless of what merged to `main`.
- **Smoke tests on the released artifact.** The release workflow additionally verifies the signed binary reports the injected version and that the published GitHub Release asset URLs resolve — these are not unit tests, but they are part of what "fully tested" means for a change that affects versioning, build flags, or release asset naming.

When adding security-sensitive logic (anything touching command construction, credential handling, or destructive cluster/namespace operations), treat the `tests/testutil/command_assertions.go` helpers as a required addition, not an optional one — they exist specifically to make the class of bugs described in their doc comments (shell injection via `bash -c` concatenation, secrets leaking into argv, `curl | bash` patterns, deletion of protected resources) a hard test failure rather than something caught only in review.

For broader context on this CLI's architecture and how these packages relate to each other, see the architecture overview (`../architecture/README.md`) and the reference architecture document (`../../reference/architecture/overview.md`). For local environment setup needed before running integration-style tests (Docker, k3d, kubectl, Helm), see `../setup/local-development.md` and `../setup/environment.md`.
