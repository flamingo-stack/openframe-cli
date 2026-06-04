# Testing Guide

This guide describes the test structure, how to run tests, how to write new tests, and coverage requirements for the OpenFrame CLI project.

---

## Test Structure and Organization

The OpenFrame CLI test suite is organized into three layers:

```text
tests/
├── integration/
│   └── common/
│       ├── cli_runner.go        # Build + execute CLI binary in tests
│       ├── cluster_management.go # K3D cluster lifecycle helpers
│       └── dependencies.go      # Shared test dependency setup
├── mocks/
│   └── dev/
│       └── kubernetes.go        # Kubernetes mock implementations
└── testutil/
    ├── setup.go                 # Test environment initialization
    ├── assertions.go            # Custom assertion helpers
    ├── cluster.go               # Cluster test utilities
    ├── patterns.go              # Reusable test patterns
    └── utilities.go             # General test utilities
```

### Unit Tests

Unit tests live alongside source files in their respective packages under `internal/`. They use `MockCommandExecutor` to simulate external tool responses without requiring real K3D, Docker, or Kubernetes.

Example: `internal/cluster/service_test.go`, `internal/chart/services/installer_test.go`

### Integration Tests

Integration tests live in `tests/integration/`. They use a real K3D installation and Docker daemon. The `CLIRunner` utility builds the CLI binary (with timestamp-based caching) and executes it as a subprocess, capturing stdout/stderr and exit codes.

---

## Running Tests

### All Unit Tests

```bash
go test ./...
```

### Verbose Output

```bash
go test -v ./...
```

### With Race Detection

```bash
go test -race ./...
```

### Specific Package

```bash
# Test cluster package only
go test -v ./internal/cluster/...

# Test chart package only
go test -v ./internal/chart/...

# Test dev package only
go test -v ./internal/dev/...

# Test shared utilities
go test -v ./internal/shared/...
```

### Single Test Case

```bash
go test -v -run TestClusterCreate ./internal/cluster/...
go test -v -run TestBootstrapService ./internal/bootstrap/...
```

### Integration Tests

Integration tests require Docker and K3D to be installed and running:

```bash
# Run integration tests (creates/deletes real clusters)
go test -v -tags integration ./tests/integration/...
```

> **Warning:** Integration tests are slow (several minutes) and resource-intensive. Run them before submitting a PR, not on every code change.

### Test with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Open `coverage.html` in a browser to explore line-by-line coverage.

---

## Test Utilities

### `testutil.InitializeTestMode()`

Call this at the start of any test that involves UI components to disable interactive prompts:

```go
func TestMyFeature(t *testing.T) {
    testutil.InitializeTestMode() // Disables pterm/promptui interactive elements
    // ... rest of test
}
```

### `testutil.NewTestMockExecutor()`

Creates a `MockCommandExecutor` with pre-configured responses for standard K3D commands:

```go
func TestClusterList(t *testing.T) {
    testutil.InitializeTestMode()
    executor := testutil.NewTestMockExecutor()
    
    // Inject a custom response for a specific command
    executor.SetResponse("k3d cluster list", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   `[{"name":"test-cluster","nodes":[]}]`,
    })
    
    manager := k3d.NewK3dManager(executor, false)
    clusters, err := manager.ListClusters(context.Background())
    
    assert.NoError(t, err)
    assert.Len(t, clusters, 1)
}
```

### `testutil.CreateStandardTestFlags()`

Sets up a fully mocked dependency container for testing command handlers:

```go
flags := testutil.CreateStandardTestFlags()
testutil.SetVerboseMode(flags, true)

// Use flags to test service behavior with mocked K3D
```

### `testutil.CreateIntegrationTestFlags()`

Creates a real dependency container for integration tests that exercise actual K3D operations:

```go
flags := testutil.CreateIntegrationTestFlags()
// This will use real k3d commands — requires Docker + K3D
```

---

## Writing New Tests

### Unit Test Pattern

```go
package mypackage_test

import (
    "testing"
    "context"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func TestMyNewFeature(t *testing.T) {
    // 1. Initialize test mode (disables UI)
    testutil.InitializeTestMode()
    
    // 2. Create mock executor
    executor := testutil.NewTestMockExecutor()
    
    // 3. Set up mock responses
    executor.SetResponse("some-tool --arg value", &executor.CommandResult{
        ExitCode: 0,
        Stdout:   "expected output",
    })
    
    // 4. Create the system under test with injected mock
    svc := NewMyService(executor, false)
    
    // 5. Execute and assert
    result, err := svc.DoSomething(context.Background(), "input")
    require.NoError(t, err)
    assert.Equal(t, "expected result", result)
}
```

### Error Path Testing

Always test failure scenarios alongside happy paths:

```go
func TestMyFeature_CommandFails(t *testing.T) {
    testutil.InitializeTestMode()
    executor := testutil.NewTestMockExecutor()
    
    executor.SetResponse("some-tool --arg value", &executor.CommandResult{
        ExitCode: 1,
        Stderr:   "command failed: permission denied",
    })
    
    svc := NewMyService(executor, false)
    _, err := svc.DoSomething(context.Background(), "input")
    
    require.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

### Interface Compliance Tests

When implementing a new provider, add a compile-time interface check:

```go
// At the top of your implementation file
var _ types.MyProviderInterface = (*MyProvider)(nil)
```

This causes a compile error if `MyProvider` doesn't satisfy the interface, catching issues before tests run.

---

## Mocking External Commands

The `MockCommandExecutor` in `internal/shared/executor/mock.go` uses pattern-based response injection. Patterns are matched against the full command string `"tool arg1 arg2 ..."`:

```go
executor.SetResponse("k3d cluster create my-cluster", &CommandResult{
    ExitCode: 0,
    Stdout:   "INFO[0000] Cluster 'my-cluster' created successfully!",
})

// Wildcard patterns (if supported)
executor.SetResponse("kubectl get pods", &CommandResult{
    ExitCode: 0,
    Stdout:   "No resources found.",
})
```

---

## Coverage Requirements

| Package | Minimum Coverage Target |
|---|---|
| `internal/cluster/` | 70% |
| `internal/chart/` | 70% |
| `internal/dev/` | 60% |
| `internal/shared/` | 80% |
| `cmd/` | 50% |

> Coverage is checked during CI. PRs that significantly reduce coverage for a package will be flagged for review.

---

## CI Integration

Tests run automatically on every pull request. The CI pipeline runs:

```bash
# Unit tests with race detection
go test -race ./...

# Linting
golangci-lint run ./...

# Vulnerability scan
govulncheck ./...
```

Integration tests are run on a separate schedule (not on every PR) due to their resource requirements.
