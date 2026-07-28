# Testing Guide

OpenFrame CLI uses a layered testing approach: fast unit tests with mock executors, integration tests against real CLI binaries, and shared test utilities that eliminate boilerplate.

---

## Test Structure and Organization

```text
tests/
├── integration/
│   ├── common/
│   │   ├── cli_runner.go       # Build + execute CLI binary, capture output
│   │   ├── cluster_management.go  # Helpers for cluster lifecycle in tests
│   │   └── dependencies.go     # Dependency setup utilities
│   └── ...                     # Integration test files
└── testutil/
    ├── setup.go                # Test mode init, mock executor, flag containers
    ├── patterns.go             # Standard command test patterns (Structure/Flags/CLI/Execution)
    ├── assertions.go           # Custom assertion helpers
    ├── cluster.go              # Cluster-specific test helpers
    ├── command_assertions.go   # CLI output assertion utilities
    ├── flag_contract.go        # Flag contract validation helpers
    └── utilities.go            # General test utilities
```

Unit tests live alongside the source code they test (e.g., `internal/cluster/service_test.go`).

---

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector (recommended)
go test -race ./...

# Run tests for a specific package
go test ./internal/cluster/...
go test ./cmd/bootstrap/...

# Run a specific test by name
go test -run TestBootstrapService ./internal/bootstrap/...
```

### Integration Tests

Integration tests require Docker, k3d, and Helm to be installed and running:

```bash
# Run all integration tests (longer timeout required)
go test ./tests/integration/... -v -timeout 30m

# Run a specific integration test
go test ./tests/integration/... -run TestClusterCreate -v -timeout 10m
```

> **Resource requirement:** Integration tests provision real K3D clusters. Ensure at least 24 GB RAM and 50 GB disk space are available.

### Coverage

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out | tail -1
```

---

## Test Utilities

### Initializing Test Mode

Always call `testutil.InitializeTestMode()` in test setups to enable safe UI rendering (prevents pterm from trying to write to a non-TTY):

```go
func TestMain(m *testing.M) {
    testutil.InitializeTestMode()
    os.Exit(m.Run())
}
```

### Mock Command Executor

The `MockCommandExecutor` replaces real shell-outs with configurable stubs, enabling fully isolated unit tests:

```go
func TestCreateCluster(t *testing.T) {
    testutil.InitializeTestMode()

    mock := testutil.NewTestMockExecutor()
    mock.SetResponse("k3d cluster create", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   `{"name": "test-cluster"}`,
    })

    // Inject mock into the service under test
    svc := cluster.NewClusterService(mock)
    err := svc.CreateCluster(context.Background(), "test-cluster")
    assert.NoError(t, err)

    // Verify the right command was called
    calls := mock.RecordedCalls()
    assert.Contains(t, calls[0].Args, "create")
}
```

### Standard Flag Containers

Use `CreateStandardTestFlags()` for unit tests (mock dependencies) and `CreateIntegrationTestFlags()` for integration tests (real dependencies):

```go
// Unit test — mock executor, no live cluster needed
flags := testutil.CreateStandardTestFlags()

// Integration test — real executor, requires k3d
flags := testutil.CreateIntegrationTestFlags()
```

`CreateStandardTestFlags()` pre-configures common mock responses:
- `k3d cluster list` → empty array `[]`
- `k3d cluster get` → not found

---

## Writing Unit Tests

### Testing a Cobra Command

Use `testutil.TestClusterCommand` to run the four standard sub-tests for any cluster command:

```go
package create_test

import (
    "testing"
    "github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func TestCreateCommand(t *testing.T) {
    testutil.TestClusterCommand(
        t,
        "create",            // command name
        NewCreateCommand,    // func() *cobra.Command
        func() {             // setup
            testutil.InitializeTestMode()
        },
        func() {},           // teardown
    )
}
```

This runs four sub-tests automatically:

| Sub-test | What it checks |
|---|---|
| `Structure` | Name, short/long descriptions, `RunE` presence |
| `Flags` | `--help` succeeds, unknown flags return error |
| `CLI` | Argument count validation via `cmd.Args` |
| `Execution` | `--dry-run` behavior (if registered), `--help` always succeeds |

### Testing Business Logic (Service Layer)

```go
func TestChartServiceInstall(t *testing.T) {
    testutil.InitializeTestMode()

    mock := testutil.NewTestMockExecutor()
    // Configure mock responses for helm commands
    mock.SetResponse("helm upgrade --install argo-cd", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   "Release \"argo-cd\" has been upgraded.",
    })

    svc := chart.NewChartService(mock, fakeK8sClient)
    err := svc.InstallArgoCD(context.Background(), cfg)
    assert.NoError(t, err)
}
```

### Testing Error Paths

```go
func TestClusterCreateFailure(t *testing.T) {
    testutil.InitializeTestMode()
    mock := testutil.NewTestMockExecutor()

    // Simulate k3d failure
    mock.SetResponse("k3d cluster create", &executor.CommandResult{
        ExitCode: 1,
        Stderr:   "cluster already exists",
    })

    svc := cluster.NewClusterService(mock)
    err := svc.CreateCluster(context.Background(), "existing-cluster")

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "already exists")
}
```

---

## Writing Integration Tests

Integration tests use the `common.CLIRunner` to build and execute the real binary:

```go
package integration_test

import (
    "log"
    "os"
    "strings"
    "testing"
    "github.com/flamingo-stack/openframe-cli/tests/integration/common"
)

func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatalf("CLI build failed: %v", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}

func TestClusterList(t *testing.T) {
    result := common.RunCLI("cluster", "list", "--output", "json")

    if result.Failed() {
        t.Fatalf("cluster list failed: %s", result.ErrorMessage())
    }

    // Verify JSON output
    if !strings.HasPrefix(strings.TrimSpace(result.Stdout), "[") {
        t.Errorf("expected JSON array output, got: %s", result.Stdout)
    }
}
```

### CLIResult Methods

| Method | Description |
|---|---|
| `result.Success()` | `true` when exit code is 0 and no error |
| `result.Failed()` | Inverse of `Success()` |
| `result.Output()` | Concatenates stdout + stderr |
| `result.ErrorMessage()` | Extracts first `Error: ...` line from stderr |
| `result.Stdout` | Raw stdout string |
| `result.Stderr` | Raw stderr string |
| `result.ExitCode` | Integer exit code |

### CLI Binary Caching

`InitializeCLI()` builds the binary to `build/openframe` and caches it by comparing mod times against `main.go`. Subsequent test runs skip the rebuild if the binary is newer than the source — significantly speeding up iterative testing.

---

## Test Patterns and Conventions

### Table-Driven Tests

Prefer table-driven tests for commands with multiple argument/flag combinations:

```go
func TestClusterNameValidation(t *testing.T) {
    tests := []struct {
        name        string
        clusterName string
        wantError   bool
    }{
        {"valid name", "openframe-dev", false},
        {"too short", "ab", true},
        {"uppercase", "MyCluster", true},
        {"injection attempt", "test;rm -rf /", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := models.ValidateClusterName(tt.clusterName)
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Redaction Cleanup in Tests

When testing code that uses `redact.RegisterSecret()`, always clean up in teardown:

```go
func TestWithSecret(t *testing.T) {
    defer redact.ClearSecrets()

    redact.RegisterSecret("test-token")
    // ... test code
}
```

### Non-Interactive Mode in Tests

Set non-interactive mode to prevent tests from blocking on prompts:

```go
func TestNonInteractiveBehavior(t *testing.T) {
    // Option 1: Use the --non-interactive flag in integration tests
    result := common.RunCLI("bootstrap", "--non-interactive")

    // Option 2: Set UI test mode (unit tests)
    testutil.InitializeTestMode() // sets ui.TestMode = true
}
```

---

## Coverage Requirements

| Package Type | Target Coverage |
|---|---|
| Core services (`internal/`) | ≥ 80% |
| Command layer (`cmd/`) | ≥ 70% |
| Provider implementations | ≥ 75% |
| Shared utilities | ≥ 85% |

Run coverage and check targets:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -v "100.0%"
```

---

## CI Test Environment

The GitHub Actions CI pipeline runs:

1. **Unit tests** with race detector on every push
2. **Integration tests** on PRs targeting `main`
3. **Coverage reporting** on every PR

Integration tests in CI use a matrix of Go versions and operating systems. Ensure your tests pass on both Linux and macOS.
