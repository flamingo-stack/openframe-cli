# Testing Guide

OpenFrame CLI has a layered testing strategy: unit tests with mock executors for fast, isolated testing, and integration tests that exercise the real CLI binary against actual K3D infrastructure.

---

## Test Structure and Organization

```text
tests/
├── testutil/                    # Shared test utilities (used by all test types)
│   ├── setup.go                 # Flag factories, mock setup, test initialization
│   ├── assertions.go            # Custom assertion helpers
│   ├── cluster.go               # Cluster-specific test utilities
│   ├── patterns.go              # Common test patterns
│   └── utilities.go             # General test utilities
├── integration/                 # End-to-end CLI integration tests
│   └── common/
│       ├── cli_runner.go        # CLI binary builder and runner
│       ├── cluster_management.go # Cluster lifecycle test helpers
│       └── dependencies.go      # Integration test dependency checks
└── mocks/
    └── dev/
        └── kubernetes.go        # Kubernetes mock implementations
```

Unit tests live alongside the source code they test:

```text
internal/
├── cluster/
│   ├── service.go
│   └── service_test.go          # Unit tests for ClusterService
├── chart/
│   └── services/
│       ├── chart_service.go
│       └── chart_service_test.go # Unit tests for ChartService
└── shared/
    └── executor/
        ├── executor.go
        └── executor_test.go      # Unit tests for executor
```

---

## Running Tests

### Run All Tests

```bash
go test ./...
```

### Run with Verbose Output

```bash
go test -v ./...
```

### Run a Specific Package

```bash
go test -v ./internal/cluster/...
go test -v ./internal/chart/...
go test -v ./internal/dev/...
go test -v ./internal/shared/...
```

### Run a Specific Test

```bash
go test -v -run TestClusterCreate ./internal/cluster/...
go test -v -run TestBootstrapService ./internal/bootstrap/...
```

### Run Tests with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # macOS
xdg-open coverage.html  # Linux
```

### Run Integration Tests

> Integration tests require Docker and k3d to be installed and running.

```bash
go test -v -tags integration ./tests/integration/...
```

---

## Test Types

### Unit Tests (Fast, No External Dependencies)

Unit tests use the `MockCommandExecutor` to simulate all external tool calls without actually running Docker, k3d, or Helm. They are fast and can run anywhere.

```go
import (
    "testing"
    "github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func TestClusterCreate(t *testing.T) {
    // Initialize test mode (disables interactive UI)
    testutil.InitializeTestMode()

    // Create mock-based test flags
    flags := testutil.CreateStandardTestFlags()
    testutil.SetVerboseMode(flags, false)

    // Run the operation under test
    // Mock executor returns pre-configured responses
    result, err := flags.ClusterService.ListClusters(context.Background())

    // Assert using testify
    require.NoError(t, err)
    assert.Empty(t, result)
}
```

### Integration Tests (Real K3D, Docker Required)

Integration tests build the CLI binary and run it as a subprocess, capturing stdout/stderr/exit codes:

```go
import (
    "testing"
    "github.com/flamingo-stack/openframe-cli/tests/integration/common"
)

func TestMain(m *testing.M) {
    if err := common.InitializeCLI(); err != nil {
        log.Fatal("Failed to build CLI:", err)
    }
    defer common.CleanupCLI()
    os.Exit(m.Run())
}

func TestClusterListCommand(t *testing.T) {
    result := common.RunCLI("cluster", "list")

    if !result.Success() {
        t.Fatalf("cluster list failed: %s", result.ErrorMessage())
    }

    assert.Contains(t, result.Stdout, "NAME")
}
```

---

## Writing New Tests

### Setting Up a Unit Test

Use `testutil.CreateStandardTestFlags()` to get a fully configured test environment with mock responses:

```go
func TestMyNewFeature(t *testing.T) {
    testutil.InitializeTestMode()
    flags := testutil.CreateStandardTestFlags()

    // Customize mock responses if needed
    flags.MockExecutor.SetResponse("k3d cluster create my-cluster", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   "INFO[0000] Cluster 'my-cluster' created successfully!",
    })

    // Exercise your code
    err := someService.DoOperation(context.Background(), flags)
    assert.NoError(t, err)
}
```

### Testing Error Conditions

```go
func TestClusterCreateFailure(t *testing.T) {
    testutil.InitializeTestMode()
    flags := testutil.CreateStandardTestFlags()

    // Simulate a tool failure
    flags.MockExecutor.SetResponse("k3d cluster create bad-name", &executor.CommandResult{
        ExitCode: 1,
        Stderr:   "Error: cluster name 'bad-name' is invalid",
    })

    err := clusterService.CreateCluster(context.Background(), badConfig)
    assert.Error(t, err)
    assert.True(t, errors.IsCommandError(err))
}
```

### Testing the Configuration Wizard

The wizard uses interactive prompts (`promptui`). In tests, initialize test mode to disable interactive elements:

```go
func TestWizardNonInteractive(t *testing.T) {
    testutil.InitializeTestMode() // Disables all prompts
    // ...
}
```

### Testing Validation

```go
func TestValidateClusterName(t *testing.T) {
    testCases := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid name", "my-cluster", false},
        {"with numbers", "cluster-01", false},
        {"empty name", "", true},
        {"spaces not allowed", "my cluster", true},
        {"too long", strings.Repeat("a", 64), true},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            err := models.ValidateClusterName(tc.input)
            if tc.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## Mock Executor

The `MockCommandExecutor` is the foundation of unit testing in OpenFrame CLI. It implements the `CommandExecutor` interface and returns pre-configured responses:

```go
// Create a new mock executor
executor := testutil.NewTestMockExecutor()

// Set a custom response for a specific command
executor.SetResponse("k3d cluster list --output json", &executor.CommandResult{
    ExitCode: 0,
    Stdout:   `[{"name":"test-cluster","nodes":[]}]`,
})

// The mock records all commands executed — useful for assertions
calledCommands := executor.GetCalledCommands()
assert.Contains(t, calledCommands, "k3d cluster list --output json")
```

---

## Test Utilities Reference

| Function | Package | Purpose |
|---|---|---|
| `InitializeTestMode()` | `testutil` | Disables interactive UI elements (prompts, spinners) |
| `NewTestMockExecutor()` | `testutil` | Creates a new `MockCommandExecutor` |
| `CreateStandardTestFlags()` | `testutil` | Creates unit test environment with mocked k3d responses |
| `CreateIntegrationTestFlags()` | `testutil` | Creates real dependency container for integration tests |
| `SetVerboseMode(flags, bool)` | `testutil` | Configures verbose logging on a test flag container |
| `InitializeCLI()` | `integration/common` | Builds the CLI binary (cached by source timestamp) |
| `CleanupCLI()` | `integration/common` | Removes the test CLI binary |
| `RunCLI(args...)` | `integration/common` | Executes a CLI command and captures output |

---

## Coverage Requirements

The project aims for the following coverage targets:

| Package | Target Coverage |
|---|---|
| `internal/cluster/` | 80%+ |
| `internal/chart/` | 75%+ |
| `internal/dev/` | 70%+ |
| `internal/shared/` | 85%+ |
| `cmd/` | 60%+ (primarily tested via integration) |

Check current coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "total|internal"
```

---

## CI Test Execution

In CI/CD environments, the prerequisite checker automatically detects CI and skips interactive prompts. Unit tests run in all CI environments. Integration tests require Docker and k3d to be available in the CI runner.

```bash
# Unit tests only (no Docker/k3d required)
go test ./internal/... ./cmd/...

# Full test suite including integration
go test ./... -tags integration
```

---

## Troubleshooting Tests

| Issue | Cause | Fix |
|---|---|---|
| Prompts blocking in tests | Test mode not initialized | Call `testutil.InitializeTestMode()` at the start of tests |
| Mock responses not matching | Command string doesn't match exactly | Print `executor.GetCalledCommands()` to see actual strings used |
| Integration test binary not found | CLI not built | Call `common.InitializeCLI()` in `TestMain` |
| Flaky integration tests | Race conditions in cluster operations | Add appropriate timeouts and retries using test utilities |
