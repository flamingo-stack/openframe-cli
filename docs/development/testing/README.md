# Testing

OpenFrame CLI has a layered testing strategy: fast, offline unit tests that mock all external tools, and integration tests that exercise the real, built binary (optionally against real Docker/k3d/kubectl/Helm).

## Test Structure and Organization

| Location | Purpose |
|---|---|
| `tests/testutil/` | Shared unit-test helpers: mock executors, flag containers, command-structure assertions, cluster fixtures |
| `tests/integration/common/` | Integration-test harness: builds/runs the real CLI binary, checks for real external dependencies |
| `*_test.go` files alongside source (e.g., `cmd/cluster/`, `internal/cluster/`) | Package-local unit tests |

Key shared utilities:

- **`testutil.InitializeTestMode()`** — enables test-safe UI rendering (`ui.TestMode = true`), avoiding TTY-dependent behavior in tests.
- **`testutil.CreateStandardTestFlags()`** — builds a `*cluster.FlagContainer` pre-wired with a `MockCommandExecutor` and canned k3d responses (empty cluster list, "not found" on `k3d cluster get`) — the fastest path for unit-testing cluster command logic.
- **`testutil.CreateIntegrationTestFlags()`** — builds a `FlagContainer` using real dependencies (the real k3d manager is resolved at runtime), for use in environments where k3d/Docker are actually installed.
- **`testutil.TestClusterCommand(t, name, newCmdFn, setup, teardown)`** — a standardized pattern that runs four sub-tests automatically (`Structure`, `Flags`, `CLI`, `Execution`) against any cluster subcommand.
- **`common.RequireClusterDependencies(t)` / `RequireK8sDependencies(t)` / `RequireAllDependencies(t)`** — gracefully skip (not fail) integration tests when Docker, k3d, kubectl, or Helm aren't available on the host.
- **`common.InitializeCLI()` / `CleanupCLI()` / `RunCLI(args...)`** — build the real `openframe` binary into `build/openframe` (with mod-time caching) and execute it as a subprocess, capturing stdout/stderr/exit code for black-box assertions.

## Running Tests

Run all unit tests:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./internal/cluster/...
go test ./cmd/cluster/...
```

Run with verbose output:

```bash
go test -v ./...
```

Run integration tests (these build and execute the real binary, and may skip themselves if required tools like Docker/k3d aren't present):

```bash
go test ./tests/integration/...
```

> Integration tests use `common.RequireClusterDependencies(t)` and similar guards to skip gracefully rather than fail when Docker/k3d/kubectl/Helm aren't installed on the CI/dev machine — check the skip message if a test doesn't run as expected.

## Writing New Tests

### Unit tests for command logic

Prefer mocked execution so tests run fast and offline:

```go
func TestClusterCreate(t *testing.T) {
    testutil.InitializeTestMode()
    flags := testutil.CreateStandardTestFlags()

    // flags.Executor is a MockCommandExecutor — no real Docker/k3d required
    result, err := runCreate(flags)
    // assert on result/err
}
```

### Standardized command structure tests

For any new Cobra subcommand under `cmd/cluster/`, add a `TestClusterCommand` invocation to get structure/flags/CLI/execution coverage for free:

```go
func TestCreateCommand(t *testing.T) {
    testutil.TestClusterCommand(
        t,
        "create",
        NewCreateCommand,
        func() { /* setup mocks */ },
        func() { /* teardown */ },
    )
}
```

### Security-sensitive assertions

When a change touches command construction (especially anything derived from user input), assert against the mock's structured argv rather than flattened strings, to catch shell-injection-style bugs:

```go
exec := executor.NewMockCommandExecutor()
// ... exercise code under test ...
for _, cmd := range exec.Commands() {
    for _, arg := range cmd.Args {
        if strings.Contains(arg, "$(") {
            t.Errorf("shell injection in argv: %q", arg)
        }
    }
}
```

### Integration tests against the real binary

```go
func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatalf("setup failed: %v", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}

func TestClusterList(t *testing.T) {
    common.RequireClusterDependencies(t)

    result := common.RunCLI("cluster", "list")
    if result.Failed() {
        t.Fatalf("expected success, got: %s", result.ErrorMessage())
    }
}
```

## Coverage Expectations

There is no single global coverage gate documented in the codebase; instead, the project relies on:

- **Mocked unit tests** covering command structure, flag validation, and orchestration logic across `cmd/` and `internal/` packages (fast, run on every change).
- **Integration tests** that exercise the actual compiled binary end-to-end, gated behind dependency checks so they only run where Docker/k3d/kubectl/Helm are genuinely available.
- **Standardized per-command tests** (`TestClusterCommand`) ensure every cluster subcommand at minimum has consistent Structure/Flags/CLI/Execution coverage — new subcommands should adopt this pattern rather than hand-rolling equivalent checks.

When adding a new command or provider, aim to cover:

1. Flag/argument validation (valid, invalid, and edge-case inputs).
2. The mocked "happy path" orchestration logic.
3. At least one error path (e.g., a failing mocked command) and how the CLI surfaces it.
4. If touching shelled-out commands: an argv-level assertion that user input can't be interpreted as shell syntax.
