# Testing Overview

This guide covers OpenFrame CLI's testing strategy, test organization, running tests, and best practices for writing comprehensive tests.

## Testing Philosophy

OpenFrame CLI follows a comprehensive testing strategy that ensures reliability, maintainability, and confidence in deployments:

- **Test-Driven Development**: Write tests before implementation when possible
- **Comprehensive Coverage**: Unit, integration, and end-to-end testing
- **Real Environment Testing**: Use actual K3d clusters for integration tests
- **Fast Feedback**: Quick unit tests with slower integration tests as needed
- **Reliable Tests**: Tests should be deterministic and not flaky

## Test Structure

### Test Organization

```text
project/
├── cmd/                          # Command layer tests
│   ├── bootstrap/
│   │   ├── bootstrap.go
│   │   └── bootstrap_test.go     # Command-level tests
│   ├── cluster/
│   │   ├── create.go
│   │   ├── create_test.go        # Command tests
│   │   └── ...
│   └── ...
├── internal/                     # Business logic tests
│   ├── cluster/
│   │   ├── services/
│   │   │   ├── cluster.go
│   │   │   └── cluster_test.go   # Service layer tests
│   │   ├── models/
│   │   │   ├── config.go
│   │   │   └── config_test.go    # Model validation tests
│   │   └── ...
│   └── ...
├── tests/                        # Integration and E2E tests
│   ├── integration/
│   │   ├── cluster_test.go       # Cluster integration tests
│   │   ├── bootstrap_test.go     # Bootstrap integration tests
│   │   └── helpers/              # Test helpers
│   ├── e2e/
│   │   ├── full_workflow_test.go # Complete workflow tests
│   │   └── scenarios/            # Different test scenarios
│   ├── fixtures/                 # Test data
│   └── utils/                    # Test utilities
└── scripts/
    └── test/                     # Test execution scripts
```

## Test Categories

### 1. Unit Tests

Unit tests focus on individual functions and methods in isolation.

#### Characteristics:
- **Fast execution** (< 1 second per test)
- **No external dependencies** (mocked)
- **High coverage** of business logic
- **Deterministic** results

#### Example Unit Test:

```go
package models

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestClusterConfig_Validate(t *testing.T) {
    tests := []struct {
        name    string
        config  ClusterConfig
        wantErr bool
        errMsg  string
    }{
        {
            name: "valid config",
            config: ClusterConfig{
                Name:       "test-cluster",
                Nodes:      3,
                K8sVersion: "v1.27.3",
            },
            wantErr: false,
        },
        {
            name: "empty name",
            config: ClusterConfig{
                Name:  "",
                Nodes: 3,
            },
            wantErr: true,
            errMsg:  "cluster name cannot be empty",
        },
        {
            name: "invalid nodes",
            config: ClusterConfig{
                Name:  "test",
                Nodes: 0,
            },
            wantErr: true,
            errMsg:  "number of nodes must be at least 1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### 2. Integration Tests

Integration tests verify component interactions with real external dependencies.

#### Characteristics:
- **Moderate execution time** (10-60 seconds)
- **Real external services** (K3d, Docker)
- **Environment setup/teardown** required
- **Test critical integration points**

#### Example Integration Test:

```go
//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/flamingo-stack/openframe-cli/internal/cluster/services"
    "github.com/flamingo-stack/openframe-cli/internal/cluster/models"
)

func TestClusterService_CreateAndDelete(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    ctx := context.Background()
    clusterName := "test-cluster-" + randomString(6)
    
    // Setup
    service := services.NewClusterService()
    
    // Test cluster creation
    config := models.ClusterConfig{
        Name:       clusterName,
        Nodes:      1,
        K8sVersion: "v1.27.3",
    }
    
    err := service.CreateCluster(ctx, config)
    require.NoError(t, err, "Failed to create cluster")
    
    // Verify cluster exists
    clusters, err := service.ListClusters(ctx)
    require.NoError(t, err, "Failed to list clusters")
    
    found := false
    for _, cluster := range clusters {
        if cluster.Name == clusterName {
            found = true
            assert.Equal(t, "running", cluster.Status)
            break
        }
    }
    assert.True(t, found, "Created cluster not found in list")
    
    // Test cluster deletion
    err = service.DeleteCluster(ctx, clusterName)
    require.NoError(t, err, "Failed to delete cluster")
    
    // Verify cluster is deleted
    clusters, err = service.ListClusters(ctx)
    require.NoError(t, err, "Failed to list clusters after deletion")
    
    for _, cluster := range clusters {
        assert.NotEqual(t, clusterName, cluster.Name, "Cluster still exists after deletion")
    }
}

func randomString(length int) string {
    // Implementation for generating random string
    return "abc123"
}
```

### 3. End-to-End Tests

E2E tests verify complete workflows from user perspective.

#### Characteristics:
- **Slow execution** (2-10 minutes)
- **Complete user workflows** tested
- **Real CLI execution** via subprocesses
- **Full environment validation**

#### Example E2E Test:

```go
//go:build e2e
// +build e2e

package e2e

import (
    "bytes"
    "os/exec"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestBootstrapWorkflow(t *testing.T) {
    clusterName := "e2e-test-" + randomString(6)
    
    // Build the CLI binary
    buildCmd := exec.Command("go", "build", "-o", "./openframe-test", ".")
    require.NoError(t, buildCmd.Run(), "Failed to build CLI binary")
    defer cleanupBinary(t, "./openframe-test")
    
    // Test bootstrap command
    bootstrapCmd := exec.Command(
        "./openframe-test", 
        "bootstrap", 
        clusterName,
        "--deployment-mode=oss-tenant",
        "--non-interactive",
        "--verbose",
    )
    
    var out bytes.Buffer
    var stderr bytes.Buffer
    bootstrapCmd.Stdout = &out
    bootstrapCmd.Stderr = &stderr
    
    err := bootstrapCmd.Run()
    if err != nil {
        t.Logf("STDOUT: %s", out.String())
        t.Logf("STDERR: %s", stderr.String())
    }
    require.NoError(t, err, "Bootstrap command failed")
    
    // Verify cluster was created
    verifyCluster(t, clusterName)
    
    // Verify ArgoCD was installed
    verifyArgoCD(t, clusterName)
    
    // Cleanup
    cleanup(t, clusterName)
}

func verifyCluster(t *testing.T, clusterName string) {
    // Use kubectl to verify cluster
    cmd := exec.Command("kubectl", "get", "nodes", "--context", "k3d-"+clusterName)
    err := cmd.Run()
    assert.NoError(t, err, "Cluster nodes not accessible")
}

func verifyArgoCD(t *testing.T, clusterName string) {
    // Verify ArgoCD pods are running
    cmd := exec.Command(
        "kubectl", "get", "pods", 
        "-n", "argocd",
        "--context", "k3d-"+clusterName,
        "-o", "jsonpath={.items[*].status.phase}",
    )
    
    output, err := cmd.Output()
    require.NoError(t, err, "Failed to get ArgoCD pods")
    
    phases := string(output)
    assert.Contains(t, phases, "Running", "ArgoCD pods not running")
}
```

## Running Tests

### Test Execution Commands

```bash
# Run all unit tests
make test
go test ./...

# Run unit tests with coverage
make test-coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/cluster/services/...
go test ./cmd/bootstrap/...

# Run integration tests
make test-integration
go test -tags=integration ./tests/integration/...

# Run end-to-end tests
make test-e2e
go test -tags=e2e ./tests/e2e/...

# Run tests in verbose mode
go test -v ./...

# Run tests with race detection
go test -race ./...
```

### Test Execution Targets

Create these Makefile targets for consistent test execution:

```makefile
# Makefile
.PHONY: test test-unit test-integration test-e2e test-coverage test-all

# Run unit tests
test: test-unit

test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./tests/integration/...

# Run end-to-end tests
test-e2e:
	@echo "Running e2e tests..."
	go test -v -tags=e2e -timeout=20m ./tests/e2e/...

# Run with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run all tests
test-all: test-unit test-integration test-e2e

# Quick verification tests
test-quick:
	go test -short -race ./...
```

### Continuous Integration

Example GitHub Actions workflow:

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test-unit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.19'
    
    - name: Run unit tests
      run: make test-unit
    
    - name: Run tests with coverage
      run: make test-coverage
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  test-integration:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.19'
    
    - name: Install dependencies
      run: |
        curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
        kubectl version --client
    
    - name: Run integration tests
      run: make test-integration
```

## Writing Effective Tests

### Test Structure (AAA Pattern)

```go
func TestServiceOperation(t *testing.T) {
    // Arrange - Set up test data and dependencies
    config := models.ClusterConfig{
        Name:  "test-cluster",
        Nodes: 3,
    }
    mockProvider := &mockK3dProvider{}
    service := NewClusterService(mockProvider)
    
    // Act - Execute the operation
    err := service.CreateCluster(config)
    
    // Assert - Verify the results
    assert.NoError(t, err)
    mockProvider.AssertCalled(t, "CreateCluster", config)
}
```

### Table-Driven Tests

```go
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected error
    }{
        {"valid name", "my-cluster", nil},
        {"empty name", "", ErrEmptyName},
        {"invalid chars", "my_cluster!", ErrInvalidChars},
        {"too long", strings.Repeat("a", 64), ErrNameTooLong},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateClusterName(tt.input)
            assert.Equal(t, tt.expected, err)
        })
    }
}
```

### Mocking External Dependencies

```go
// Mock interface
type MockK3dProvider struct {
    mock.Mock
}

func (m *MockK3dProvider) CreateCluster(config K3dConfig) error {
    args := m.Called(config)
    return args.Error(0)
}

// Test using mock
func TestClusterService_CreateCluster(t *testing.T) {
    mockProvider := &MockK3dProvider{}
    service := NewClusterService(mockProvider)
    
    config := models.ClusterConfig{Name: "test"}
    
    // Set up expectation
    mockProvider.On("CreateCluster", mock.AnythingOfType("K3dConfig")).Return(nil)
    
    // Execute
    err := service.CreateCluster(config)
    
    // Verify
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

## Testing Best Practices

### 1. Test Naming Conventions

```go
// Function being tested
func TestFunctionName_Scenario_ExpectedResult(t *testing.T) {}

// Examples:
func TestValidateConfig_EmptyName_ReturnsError(t *testing.T) {}
func TestCreateCluster_ValidConfig_CreatesSuccessfully(t *testing.T) {}
func TestDeleteCluster_NonExistentCluster_ReturnsNotFoundError(t *testing.T) {}
```

### 2. Test Data Management

```go
// Use test helpers for common data
func testClusterConfig(name string) models.ClusterConfig {
    return models.ClusterConfig{
        Name:       name,
        Nodes:      3,
        K8sVersion: "v1.27.3",
        Ports: []models.PortMapping{
            {Host: 8080, Container: 80},
        },
    }
}

// Use table-driven tests for multiple scenarios
func TestConfigValidation(t *testing.T) {
    tests := []struct {
        name      string
        configFn  func() models.ClusterConfig
        wantError bool
    }{
        {
            name:      "valid config",
            configFn:  func() models.ClusterConfig { return testClusterConfig("valid") },
            wantError: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            config := tt.configFn()
            err := config.Validate()
            if tt.wantError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### 3. Test Cleanup

```go
func TestIntegrationScenario(t *testing.T) {
    clusterName := "test-" + randomString(6)
    
    // Ensure cleanup happens even if test fails
    defer func() {
        cleanupCluster(t, clusterName)
    }()
    
    // Test implementation...
}

func cleanupCluster(t *testing.T, name string) {
    cmd := exec.Command("k3d", "cluster", "delete", name)
    if err := cmd.Run(); err != nil {
        t.Logf("Warning: failed to cleanup cluster %s: %v", name, err)
    }
}
```

### 4. Test Timeouts and Context

```go
func TestLongRunningOperation(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    service := NewClusterService()
    
    done := make(chan error, 1)
    go func() {
        done <- service.CreateCluster(ctx, testConfig())
    }()
    
    select {
    case err := <-done:
        assert.NoError(t, err)
    case <-ctx.Done():
        t.Fatal("Test timed out")
    }
}
```

## Test Utilities and Helpers

### Common Test Helpers

```go
// tests/utils/helpers.go
package utils

import (
    "math/rand"
    "time"
    "testing"
    "os/exec"
)

// RandomString generates a random string for test resources
func RandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    rand.Seed(time.Now().UnixNano())
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[rand.Intn(len(charset))]
    }
    return string(b)
}

// EnsureCleanEnvironment cleans up any existing test resources
func EnsureCleanEnvironment(t *testing.T) {
    cmd := exec.Command("k3d", "cluster", "list", "-o", "json")
    output, err := cmd.Output()
    if err != nil {
        return // k3d might not be available
    }
    
    // Parse and cleanup test clusters
    // Implementation...
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(condition func() bool, timeout time.Duration) bool {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        if condition() {
            return true
        }
        time.Sleep(100 * time.Millisecond)
    }
    return false
}
```

### Test Fixtures

```go
// tests/fixtures/cluster_configs.go
package fixtures

import "github.com/flamingo-stack/openframe-cli/internal/cluster/models"

// DefaultClusterConfig returns a valid cluster configuration for testing
func DefaultClusterConfig() models.ClusterConfig {
    return models.ClusterConfig{
        Name:       "test-cluster",
        Nodes:      3,
        K8sVersion: "v1.27.3",
        Ports: []models.PortMapping{
            {Host: 8080, Container: 80, Protocol: "tcp"},
        },
        Environment: map[string]string{
            "TEST_MODE": "true",
        },
    }
}

// MinimalClusterConfig returns a minimal valid configuration
func MinimalClusterConfig() models.ClusterConfig {
    return models.ClusterConfig{
        Name:  "minimal-test",
        Nodes: 1,
    }
}
```

## Coverage Requirements

### Coverage Targets

| Component | Target Coverage | Priority |
|-----------|----------------|----------|
| **Models/Validation** | 95% | High |
| **Service Layer** | 85% | High |
| **Command Layer** | 70% | Medium |
| **UI Components** | 60% | Medium |
| **Utilities** | 80% | Medium |

### Coverage Analysis

```bash
# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage by package
go tool cover -func=coverage.out | sort -k3 -nr

# Check coverage threshold
go test -coverprofile=coverage.out ./...
coverage=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
if (( $(echo "$coverage < 75" | bc -l) )); then
    echo "Coverage $coverage% is below 75% threshold"
    exit 1
fi
```

## Debugging Tests

### Common Test Debugging Techniques

```go
// Add debug logging to tests
func TestComplexOperation(t *testing.T) {
    if testing.Verbose() {
        // Enable debug logging only in verbose mode
        log.SetLevel(log.DebugLevel)
    }
    
    // Test implementation with debug logs...
}

// Use testify's require for early failure
func TestValidation(t *testing.T) {
    config := loadTestConfig()
    require.NotNil(t, config, "Config should not be nil")
    
    err := config.Validate()
    require.NoError(t, err, "Validation should pass")
    
    // Continue with test knowing validation passed...
}
```

### Running Single Tests

```bash
# Run specific test function
go test -run TestSpecificFunction ./internal/cluster/

# Run specific test with verbose output
go test -v -run TestClusterCreation ./tests/integration/

# Debug test with delve
dlv test ./internal/cluster/ -- -test.run TestClusterCreation
```

## Performance Testing

### Benchmark Tests

```go
func BenchmarkClusterCreation(b *testing.B) {
    config := testClusterConfig("bench-test")
    service := NewClusterService()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        err := service.CreateCluster(config)
        if err != nil {
            b.Fatal(err)
        }
        // Cleanup for next iteration
        service.DeleteCluster(config.Name)
    }
}

// Run benchmarks
go test -bench=. ./...
go test -bench=BenchmarkClusterCreation -benchmem ./internal/cluster/
```

## Next Steps

Now that you understand the testing strategy:

1. **[Contributing Guidelines](../contributing/guidelines.md)** - Learn how to contribute tests with your code
2. **[Local Development](../setup/local-development.md)** - Set up your development environment for testing
3. **[Architecture Overview](../architecture/overview.md)** - Understand the system for better test design

Remember: Good tests are as important as good code. They provide confidence, documentation, and regression protection for the entire team.