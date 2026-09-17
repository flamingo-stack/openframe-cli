# Testing

OpenFrame CLI's test suite spans fast, hermetic unit tests (built around a mock command executor) and slower integration tests that build and drive the real `openframe` binary.

## Test Structure

```text
tests/
  testutil/                  Shared unit-test helpers
    setup.go                 InitializeTestMode, NewTestMockExecutor, CreateStandardTestFlags
    flag_contract.go         FlagSpec, AssertFlag(s), AssertSubcommands — CLI surface contract testing
    command_assertions.go    Assertions over recorded/mocked commands
    assertions.go            General-purpose test assertions
    cluster.go / patterns.go / utilities.go   Additional cluster/test helpers
  integration/
    common/
      cli_runner.go           Builds the real binary (build/openframe) and runs it as a subprocess
      cluster_management.go   Higher-level integration helpers for cluster lifecycle tests
      dependencies.go         Dependency/tooling checks for integration test environments
```

Package-level unit tests live alongside the source they test (standard Go convention), e.g. `internal/cluster/models/flags_test.go` would sit next to `internal/cluster/models/flags.go`.

## Running Tests

Run the full unit test suite:

```bash
go test ./...
```

Run tests for a specific package:

```bash
go test ./internal/cluster/...
go test ./internal/shared/executor/...
```

Run with verbose output and race detection:

```bash
go test -v -race ./...
```

### Integration Tests

Integration tests build the real CLI binary and exercise it as a subprocess (via `tests/integration/common.InitializeCLI()` / `RunCLI()`). They require Docker (and other real tooling, depending on the test) to be available:

```bash
go test ./tests/integration/...
```

`InitializeCLI()` caches the built binary at `build/openframe` and skips rebuilding when it's newer than `main.go`, so repeated integration test runs are fast during local iteration.

## Writing New Tests

### Unit Tests with the Mock Executor

Most business logic depends on `internal/shared/executor.CommandExecutor`. Use `testutil.NewTestMockExecutor()` or `testutil.CreateStandardTestFlags()` to avoid invoking real `docker`/`k3d`/`helm`/`terraform`:

```go
func TestMyClusterLogic(t *testing.T) {
    testutil.InitializeTestMode()

    flags := testutil.CreateStandardTestFlags() // mock k3d executor pre-wired
    testutil.SetVerboseMode(flags, true)

    // flags.Executor is a *executor.MockCommandExecutor
    mock := flags.Executor.(*executor.MockCommandExecutor)
    mock.SetResponse("k3d cluster list", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   `[{"name":"my-cluster"}]`,
    })

    // ... call the code under test, then assert on mock.Commands()
}
```

`CreateStandardTestFlags()` pre-configures common responses (`k3d cluster list` → empty array, `k3d cluster get` → not found), covering the most frequent test paths without a live cluster. Use `CreateIntegrationTestFlags()` when you specifically want the real k3d manager (e.g., in an environment with k3d installed).

### Asserting on Executed Commands Safely

Prefer the structured `Commands()` log over the flattened `GetExecutedCommands()` string log when your assertion cares about exact arguments (especially for security-sensitive checks like shell-injection prevention):

```go
for _, cmd := range mock.Commands() {
    for _, arg := range cmd.Args {
        if strings.Contains(arg, "$(") {
            t.Errorf("shell injection in argv: %q", arg)
        }
    }
}
```

### CLI Flag/Subcommand Contract Tests

Because OpenFrame CLI's flags and subcommands are a public contract for scripts and CI pipelines, use `tests/testutil/flag_contract.go` to freeze that surface and catch accidental breaking changes:

```go
func TestClusterCreateContract(t *testing.T) {
    cmd := getCreateCmd()

    testutil.AssertFlags(t, cmd, []testutil.FlagSpec{
        {Name: "type", Shorthand: "t", Type: "string", Default: "k3d"},
        {Name: "nodes", Shorthand: "", Type: "int", Default: "1"},
    })
}

func TestClusterCommandSubcommands(t *testing.T) {
    cmd := GetClusterCmd()
    testutil.AssertSubcommands(t, cmd, "create", "delete", "list", "status", "use", "cleanup")
}
```

Any renamed flag, dropped shorthand, changed default, or added/removed subcommand will fail these tests loudly — treat a contract-test failure as a signal that you're making a breaking CLI change, and confirm that's intentional (and documented) before merging.

### Integration Tests with the Real Binary

```go
func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatalf("setup failed: %v", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}

func TestClusterCreateIntegration(t *testing.T) {
    result := common.RunCLI("cluster", "create", "dev-test", "--skip-wizard", "--nodes", "1")

    if result.Failed() {
        t.Fatalf("expected success, got: %s", result.ErrorMessage())
    }
}
```

`RunCLI` inherits the current process's environment, so any environment variables set in your test process (e.g., `OPENFRAME_NO_WSL_FORWARD`) are forwarded to the binary under test.

## Coverage Expectations

There is no single enforced coverage percentage documented in the codebase; instead, the project relies on:

- **Mock-executor-based unit tests** covering business logic in `internal/` without requiring real infrastructure.
- **Flag/subcommand contract tests** guarding the CLI's public surface against silent breaking changes.
- **Integration tests** validating real end-to-end behavior of the compiled binary against real (or realistically mocked) external tools.

When contributing new functionality, add unit tests for the business logic (using the mock executor) and, where the change affects the public CLI surface (new/changed flags or subcommands), extend the relevant contract test in `tests/testutil/flag_contract.go` usage.
