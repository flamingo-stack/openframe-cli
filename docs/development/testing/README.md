ic# Testing

OpenFrame CLI (`github.com/flamingo-stack/openframe-cli`) is a Go binary, and its test suite is built entirely on the standard `go test` toolchain plus `testify` assertions. There is no separate frontend or backend stack to test here — the repository publishes a single Go module (confirmed by the organization's dependency graph, which also shows **no recorded downstream consumers**: this CLI is a standalone, user-facing binary, not a library other repositories import). Everything described below lives inside `flamingo-stack/openframe-cli`; the platform it deploys is maintained separately in [`flamingo-stack/openframe-oss-tenant`](https://github.com/flamingo-stack/openframe-oss-tenant) and is out of scope for this document.

> This CLI has no recorded downstream consumers in the organization's code graph (coverage: full), so changing an exported function's signature cannot silently break another repository — but it can still break the CLI's own command surface, which the contract and matrix tests below exist to catch.

## Test structure and organization

Tests in this repository fall into three layers, mirroring the `cmd/` → `internal/` → external-process architecture described in the [architecture overview](../../reference/architecture/overview.md):

| Layer | Where | What it proves |
|---|---|---|
| Unit tests | `*_test.go` files beside the code they test, throughout `cmd/` and `internal/` | Individual functions, command structure, flag contracts — no external processes, no network, no cluster |
| CLI matrix tests | `cli_matrix_test.go`, `main_test.go` (repository root) | The **built binary**, invoked as a subprocess with a spread of argument combinations, asserting exit codes and output — still hermetic (no docker/k3d/helm, no network) |
| Integration tests | `tests/integration/common/` | Helpers for tests that build and run the real CLI binary against a real k3d cluster, Docker, or other external tools |

### Test support packages

Two shared packages carry the weight of keeping tests consistent across the codebase:

- **`tests/testutil/`** — helpers consumed by unit tests across `cmd/` and `internal/`:
  - `tests/testutil/setup.go` — `InitializeTestMode()`, `NewTestMockExecutor()`, `CreateStandardTestFlags()`, `CreateIntegrationTestFlags()`, `SetVerboseMode()`
  - `tests/testutil/assertions.go` — the fluent `AssertCommandOutput` chain (`NoError()`, `StdoutContains()`, `StderrContainsAny()`, etc.) plus `AssertCommandSuccess`/`AssertCommandFailure` shorthands
  - `tests/testutil/command_assertions.go` — security-focused assertions against a `MockCommandExecutor`'s structured command log: `AssertNoArgContains`, `AssertNoBashCConcatenation`, `AssertArgIsDiscrete`, `AssertNoDeleteOf`, `AssertNoCurlPipeShell`, and the exported `HasShellMetachars()` helper
  - `tests/testutil/flag_contract.go` — `FlagSpec`, `AssertFlag`, `AssertFlags`, `FindSubcommand`, `AssertSubcommands` for freezing a command's flag/subcommand surface
  - `tests/testutil/patterns.go` — `TestClusterCommand()`, which runs four standard sub-tests (`Structure`, `Flags`, `CLI`, `Execution`) against any cluster command factory
  - `tests/testutil/cluster.go` — `TestCommandStructure()`, `TestCLICommand()`, `TestFlags()`, `IntegrationFlags()`
  - `tests/testutil/utilities.go` — `TestClusterConfig(name)`, a pre-populated `*models.ClusterConfig` for unit tests

- **`tests/integration/common/`** — helpers specific to tests that exercise real infrastructure:
  - `tests/integration/common/cli_runner.go` — `InitializeCLI()` (builds `build/openframe` with mod-time caching), `RunCLI(args...)`, `CleanupCLI()`, and the `CLIResult` type (`Success()`, `Failed()`, `Output()`, `ErrorMessage()`)
  - `tests/integration/common/cluster_management.go` — k3d cluster lifecycle for tests: `GenerateTestClusterName()`, `CreateTestCluster()`, `DeleteTestCluster()`, `ClusterExists()`, `StopTestCluster()`, `CleanupTestCluster()`, `CleanupAllTestClusters()`
  - `tests/integration/common/dependencies.go` — the `Dependency` type and `RequireOrSkip()`, plus predefined dependencies `Docker`, `K3d`, `Kubectl`, `Helm` and the convenience functions `RequireClusterDependencies()`, `RequireK8sDependencies()`, `RequireAllDependencies()`

### The executor abstraction makes unit testing possible

Every external process the CLI shells out to (`docker`, `k3d`, `helm`, `terraform`, `gcloud`, `aws`, …) goes through the `CommandExecutor` interface in `internal/shared/executor/executor.go`. Its test double, `internal/shared/executor/mock.go`, is what makes most of this repository testable without touching real infrastructure:

- `MockCommandExecutor` is `sync.Mutex`-guarded (safe under `go test -race`) and keeps **two** logs of every execution: a flattened string (`GetExecutedCommands()`, legacy) and a structured `RecordedCommand` log (`Commands()`) that preserves `Name`, `Args []string`, `Env`, and `Stdin` as discrete elements.
- Configure behavior with `SetResponse(pattern, result)` for specific commands, `SetDefaultResult(result)` for the fallback, and `SetShouldFail(fail, message)` to simulate a failing external tool.
- Prefer `Commands()` over `GetExecutedCommands()` whenever a test needs to prove an argument was passed as a discrete argv element and not spliced into a shell string — the flattened log cannot tell `exec("bash", "-c", "helm install "+userInput)` apart from `exec("helm", "install", userInput)`, but the structured log can.

```go
exec := executor.NewMockCommandExecutor()

exec.SetResponse("helm upgrade", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   "Release deployed",
})

result, err := exec.Execute(ctx, "helm", "upgrade", "my-release", "./chart")

for _, cmd := range exec.Commands() {
    for _, arg := range cmd.Args {
        if strings.Contains(arg, "$(") {
            t.Errorf("shell injection in argv: %q", arg)
        }
    }
}
```

## Running tests

### Unit tests

The release pipeline's own quality gate (`.github/workflows/release.yml`) runs:

```bash
make test-unit
```

before any tag is cut — a release is never published if this step fails. Described in that workflow as covering "unit tests (root, cmd, internal, tests/testutil)", this is the fast, hermetic layer: no Docker, no network, no real cluster.

To run the same layer directly with `go test` (for example while iterating on a single package):

```bash
go test ./cmd/... ./internal/... ./tests/testutil/...
```

Narrow to one package while developing:

```bash
go test ./internal/shared/executor/...
go test ./cmd/app/...
```

### CLI matrix and root-level tests

`main_test.go` and `cli_matrix_test.go` live at the repository root and build the real binary as a subprocess:

```bash
go test .
```

`cli_matrix_test.go`'s `TestCLIMatrix` builds the CLI once (via `matrixBin`, guarded by `sync.Once`) and then runs a table of argument combinations through `runMatrix`, which executes the binary in a **hermetic environment**: an isolated `$HOME`/`$USERPROFILE` (a fresh `t.TempDir()`, so no `~/.openframe` state leaks between runs), a `$KUBECONFIG` pointed at a nonexistent file, `OPENFRAME_NO_UPDATE_CHECK=1`, `OPENFRAME_NO_WSL_FORWARD=1`, and closed stdin (so the CLI runs in its non-interactive mode, never blocking on a prompt). Everything in this file is machine-independent by construction; anything that genuinely needs Docker, k3d, or Helm belongs in the e2e workflow (`.github/workflows/test.yml`), not here.

The matrix covers, among other cases:

- the happy read-only surface (`--version`, `--help`, `app --help`, `cluster --help`, `update --help`, shell completions, `update rollback` with nothing saved)
- unknown commands and flags (`bogus`, `--bogus`, a removed flag like `--github-branch`, a legacy flag like `--deployment-mode`)
- flag/argument validation that must fail at parse time, before any cluster is touched (a non-numeric `--nodes`, a non-boolean `--prune`, too many positional args to `bootstrap`, an invalid cluster name)
- command-level guards that must fail fast without contacting a cluster (`app upgrade --ref x --sync` being mutually exclusive, `app uninstall` requiring `--yes` in non-interactive mode, `app install --context no-such-context` failing before any network call)

`main_test.go` additionally covers:

- `TestExitCode` — pure unit coverage of `main.exitCode()`, proving a `*executor.CommandError`'s exit code is propagated (bare and wrapped with `fmt.Errorf`), and that out-of-range or non-failure codes (`0`, `-1`, `4294967295`) fall back to a generic `1`.
- `TestMainIntegration` — builds a temporary test binary and exercises `--help`, `--version`, and an invalid flag end-to-end, with `OPENFRAME_NO_WSL_FORWARD=1` set so the native-Windows WSL-forwarding path is bypassed in environments that have no WSL installed.

### Command-tree sweep (`cmd/help_matrix_test.go`)

`TestEveryCommand_HelpWorks` and `TestEveryCommand_HasShortAndFlagUsages` in `cmd/help_matrix_test.go` walk the **entire** command tree returned by `GetRootCmd` (via `collectCommandPaths`) rather than hard-coding a list of commands, so every current and future command is covered automatically:

- `<command path> --help` must succeed and print a `Usage:` section — help must never be blocked by a prerequisite gate, cluster access, or the network.
- Every command must have a non-empty `Short` description (shown in its parent's command list).
- Every locally declared flag must have a non-empty usage string.

Run just this sweep:

```bash
go test ./cmd/... -run TestEveryCommand
```

### Contract tests

`cmd/contract_test.go` freezes the root command's public surface: `TestRootContract_PersistentFlags` asserts `--verbose` (shorthand `v`, bool, default `false`) and `--silent` (bool, default `false`) exist with exactly these properties, and `TestRootContract_TopLevelSubcommands` asserts that `cluster`, `app`, `bootstrap`, `prerequisites`, and `update` are all present as top-level subcommands (using `testutil.FindSubcommand`). `tests/testutil/flag_contract.go`'s `FlagSpec`/`AssertFlag`/`AssertFlags`/`AssertSubcommands` exist specifically so that a renamed flag, a dropped shorthand, a changed default, or an added/removed subcommand fails loudly instead of silently breaking scripts and CI pipelines that depend on this CLI's argument surface.

### Integration tests

Integration tests build and run the real `openframe` binary against real external tools. `tests/integration/common/dependencies.go` lets a test **skip** (never silently pass) when a required tool is unavailable:

```go
func TestClusterProvisioning(t *testing.T) {
    common.RequireClusterDependencies(t) // requires Docker + k3d
    // ...
}
```

`RequireOrSkip` logs an install hint and calls `t.Skip()`, so a missing dependency is visibly recorded as "not run," distinct from a passing test. The predefined dependencies are `Docker` (`docker info`), `K3d` (`k3d version`), `Kubectl` (`kubectl version --client`), and `Helm` (`helm version`); `RequireClusterDependencies`, `RequireK8sDependencies`, and `RequireAllDependencies` bundle the common combinations.

`tests/integration/common/cli_runner.go`'s `InitializeCLI()` should be called once per test binary (typically from `TestMain`), and `CleanupCLI()` deferred:

```go
func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatalf("setup failed: %v", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}
```

Cluster-level integration tests should generate a unique, disposable cluster name and clean it up unconditionally:

```go
name := common.GenerateTestClusterName()
if err := common.CreateTestCluster(name); err != nil {
    t.Fatalf("failed to create cluster: %v", err)
}
defer common.CleanupTestCluster(name)
```

`CleanupAllTestClusters()` is available for suite-level teardown and removes every k3d cluster whose name matches common test-naming patterns (`test`, `integration`, `stress`, `collision`, `interrupt`, `multi-`, `debug`, and a handful of command-prefixed names), along with any leftover k3d Docker networks, containers, and registries.

## Writing new tests

### Unit tests for a new command

Follow the pattern already used across `cmd/`: build the command with its `GetXxxCmd()`/`getXxxCmd()` factory (per the active rule `OPENFRAM-003`, command groups are always returned by a factory function, never constructed inline), then exercise it with `testutil.TestCommandStructure` and `testutil.TestCLICommand`:

```go
func TestMyCommand(t *testing.T) {
    cmd := NewMyCommand()

    testutil.TestCommandStructure(t, cmd, "my-command", "Short description here")

    testutil.TestCLICommand(t, cmd,
        []string{"--flag", "value"},
        false,             // expectErr
        "expected output",
    )
}
```

For a cluster command specifically, prefer `testutil.TestClusterCommand`, which runs all four standard sub-tests (`Structure`, `Flags`, `CLI`, `Execution`) in one call:

```go
func TestCreateCommand(t *testing.T) {
    testutil.TestClusterCommand(
        t,
        "create",
        NewCreateCommand,
        func() { /* setup: configure mocks/env */ },
        func() { /* teardown */ },
    )
}
```

### Injecting a mock executor

Unit tests should never shell out to a real binary. Use `testutil.CreateStandardTestFlags()`, which wires a `MockCommandExecutor` pre-configured with responses for the common `k3d cluster list` / `k3d cluster get` paths:

```go
func TestMyCommand(t *testing.T) {
    testutil.InitializeTestMode()

    flags := testutil.CreateStandardTestFlags()
    testutil.SetVerboseMode(flags, true)

    result, err := mycommand.Run(flags)
    // assert...
}
```

Call `testutil.InitializeTestMode()` (sets `ui.TestMode = true`) at the start of any test that touches `internal/shared/ui`, so UI components render in a test-safe way. Several existing test files call it from an `init()` function, as seen in `cmd/app/app_test.go` and `cmd/root_test.go`.

### Asserting command output

Chain assertions with `testutil.NewAssertCommandOutput` (or the shorthands):

```go
stdout, stderr, err := runCommand("openframe", "version")

AssertCommandSuccess(t, stdout, stderr, err).
    StdoutContains("openframe").
    ContainsAll("version", "build")

_, stderr, err = runCommand("openframe", "invalid-cmd")
AssertCommandFailure(t, "", stderr, err).
    StderrContainsAny("unknown command", "invalid-cmd")
```

### Writing a security-sensitive test

Several of this CLI's commands build shell commands or pass secrets as arguments to external tools. Any test touching that surface should assert against the **structured** command log, not the flattened one, using `tests/testutil/command_assertions.go`:

```go
func TestHelmInstall_CredentialSafety(t *testing.T) {
    mock := executor.NewMockCommandExecutor()
    // ... run CLI code under test using mock ...

    cmds := mock.Commands()

    testutil.AssertNoArgContains(t, cmds, "ghp_secrettoken123")
    testutil.AssertArgIsDiscrete(t, cmds, "my-release")
    testutil.AssertNoBashCConcatenation(t, cmds, "my-release")
    testutil.AssertNoDeleteOf(t, cmds, "kube-system")
    testutil.AssertNoCurlPipeShell(t, cmds)
}
```

These guard specific, previously-identified failure modes: a credential leaking into an argv element or command name (`AssertNoArgContains`), an argument-derived shell metacharacter reaching a `bash -c`/`sh -c` payload (`AssertNoBashCConcatenation`), an argument that should have been passed as a discrete argv element instead being spliced into a larger string (`AssertArgIsDiscrete`), a destructive command targeting a protected namespace such as `kube-system` (`AssertNoDeleteOf`), and a `curl | bash`-style remote-pipe-to-shell pattern (`AssertNoCurlPipeShell`).

### Freezing a flag or subcommand contract

When a command's flags or subcommands form part of the CLI's public, scriptable surface, lock it down with `tests/testutil/flag_contract.go`:

```go
func TestRootCommandContract(t *testing.T) {
    cmd := NewRootCmd()

    testutil.AssertSubcommands(t, cmd, "install", "upgrade", "version")

    installCmd := testutil.FindSubcommand(t, cmd, "install")
    testutil.AssertFlags(t, installCmd, []testutil.FlagSpec{
        {Name: "namespace", Shorthand: "n", Type: "string", Default: "default"},
        {Name: "dry-run",   Shorthand: "",  Type: "bool",   Default: "false"},
    })
}
```

### Repository conventions a new test must follow

The repository's active code review rules (enforced by the automated reviewer) shape what a passing review expects from new Go test code in this area, in addition to the conventions already shown above:

- **Command groups must be built by factory functions** (`OPENFRAM-003`): a test that needs a command instance should call its `Get<Group>Cmd()`/`get<Sub>Cmd()` factory, never construct a `cobra.Command` literal inline.
- **A command group with its own `PersistentPreRunE` must call `ui.ApplyGlobalOutputFlags(cmd)`** (`OPENFRAM-005-10`): because cobra only runs the closest ancestor's hook, a group like `app` or `cluster` that defines its own hook must apply this call itself, or `--verbose`/`--silent` silently stop working for that whole subtree. `cmd/app/app_test.go`'s `TestAppGroup_PersistentPreRunEHonorsVerbose` is a working example of a regression test for exactly this.
- **Commands targeting a cluster via a flag (e.g. `--context`) must reject stray positional arguments** with a named-flag hint (`OPENFRAM-009-8`), using a validator like `app.noPositionalArgs` — cobra's default `ArbitraryArgs` would otherwise silently ignore them, which is especially dangerous for a destructive command like `uninstall`.
- **Errors already shown to the user must be (or wrap) `sharederrors.AlreadyHandledError`** before being returned up to `main` (`OPENFRAM-004`, `OPENFRAM-003-10`), so `main.go`'s `errors.As` check suppresses a second, redundant print.
- **An external command's exit code must survive as `*executor.CommandError`**, not get collapsed to a generic `1` (`OPENFRAM-005`) — `main.exitCode()` is what unwraps it; `TestExitCode` in `main_test.go` is the regression test for this behavior.
- **User-facing output must go through `pterm` printers or `internal/shared/ui`**, never raw `fmt.Println`/`fmt.Printf` to stdout (`OPENFRAM-007`) — this is what lets tests redirect pterm's writers (as `TestVerboseEnablesDebugOutput` in `cmd/root_test.go` does) and what makes `--silent`/`--plain` work at all.
- **Cross-package imports that would collide on their last path segment** (two packages both named `config`, `ui`, `errors`, etc.) **must use a disambiguating camelCase alias** such as `sharedErrors` or `uiCluster` (`OPENFRAM-001`) — keep this in mind when a new test file imports both a package-under-test and a same-named shared package.

See the [contributing guidelines](../contributing/guidelines.md) for the broader pull-request process these rules feed into.

## Coverage requirements

This repository's release gate (`.github/workflows/release.yml`) enforces two mandatory checks before any tag is cut, both of which a contributor should run locally before opening a pull request:

```bash
# Gate: lint (golangci-lint, same version used in CI)
golangci-lint run

# Gate: unit tests
make test-unit
```

A release is **not published** unless `make test-unit` passes — this is the hard floor the material defines; there is no separately documented numeric coverage percentage threshold for this repository, so the practical bar is: every unit test must pass, every command's `--help` must work (`cmd/help_matrix_test.go`'s tree-wide sweep), and the CLI matrix (`cli_matrix_test.go`) must continue to pass for its full table of argument combinations.

In practice, that means a change is review-ready only when:

- `go test ./cmd/... ./internal/... ./tests/testutil/...` passes (the same scope `make test-unit` covers, per the release workflow's own description — "unit tests (root, cmd, internal, tests/testutil)").
- `go test .` passes at the repository root, covering `main_test.go` and `cli_matrix_test.go`.
- Any new or changed command still satisfies `cmd/help_matrix_test.go`'s whole-tree sweep (`TestEveryCommand_HelpWorks`, `TestEveryCommand_HasShortAndFlagUsages`) — since this walks the live command tree rather than a hard-coded list, a new command is automatically covered, but it must still have a non-empty `Short` description and fully-described flags.
- Any flag or subcommand that forms part of the CLI's scriptable contract has a corresponding `tests/testutil/flag_contract.go` assertion, as `cmd/contract_test.go` does for the root command's `--verbose`/`--silent` flags and its top-level subcommand list.
- Security-sensitive command construction (anything touching credentials, shell concatenation, or destructive operations against cluster resources) is covered by the structured-log assertions in `tests/testutil/command_assertions.go`, not just the legacy flattened-string log.

Beyond this CLI's own binary, the release workflow also runs a cross-cutting smoke test after `goreleaser release --clean` publishes artifacts: it verifies the darwin/arm64 binary in `dist/` actually reports the version that was just tagged (guarding the `-X cmd.version` ldflags injection) and that the released asset URLs (`openframe-cli_linux_amd64.tar.gz`, `checksums.txt`, `checksums.txt.bundle`) resolve — these are post-publish checks, not part of the `make test-unit` gate, but they are the final backstop that a broken release never reaches users silently.
