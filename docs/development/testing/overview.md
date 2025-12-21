# Testing Overview

OpenFrame CLI employs a comprehensive testing strategy to ensure reliability, maintainability, and quality. This guide covers test structure, running tests, writing new tests, and coverage requirements.

## Testing Philosophy

Our testing approach follows these principles:

1. **Test Pyramid**: More unit tests, fewer integration tests, minimal E2E tests
2. **Fast Feedback**: Unit tests run quickly for rapid development cycles
3. **Real Environment Testing**: Integration tests use actual Docker and K3d
4. **User Journey Validation**: E2E tests verify complete workflows
5. **Test-Driven Development**: Write tests before implementation when possible

## Test Structure and Organization

### Directory Structure

```
openframe-cli/
├── cmd/                           # Command layer tests
│   ├── bootstrap/
│   │   ├── bootstrap_test.go      # Command integration tests
│   │   └── testdata/              # Test fixtures and data
│   ├── cluster/
│   │   ├── create_test.go         # Individual command tests
│   │   └── cluster_test.go        # Command suite tests
│   └── ...
├── internal/                      # Service layer tests  
│   ├── bootstrap/
│   │   ├── service_test.go        # Unit tests for service logic
│   │   ├── integration_test.go    # Integration tests
│   │   └── mocks/                 # Generated mocks
│   ├── cluster/
│   │   ├── services/
│   │   │   ├── command_service_test.go
│   │   │   └── k3d_provider_test.go
│   │   ├── models/
│   │   │   └── configuration_test.go
│   │   └── ui/
│   │       └── wizard_test.go
│   └── shared/
│       ├── ui/
│       │   └── prompt_test.go
│       └── errors/
│           └── handler_test.go
├── test/                          # E2E and integration tests
│   ├── e2e/                       # End-to-end test suites
│   │   ├── bootstrap_test.go      # Complete workflow tests
│   │   ├── cluster_lifecycle_test.go
│   │   └── chart_installation_test.go
│   ├── integration/               # Cross-service integration tests
│   │   └── cluster_chart_test.go
│   └── testutils/                 # Shared test utilities
│       ├── clusters.go            # Test cluster management
│       ├── docker.go              # Docker test utilities
│       └── fixtures.go            # Test data and fixtures
└── scripts/test/                  # Test automation scripts
    ├── run-unit.sh
    ├── run-integration.sh
    └── run-e2e.sh
```

### Test Categories

| Category | Scope | Tools | Example |
|----------|-------|-------|---------|
| **Unit Tests** | Single function/method | `go test`, `testify`, `gomock` | `TestClusterConfigValidation` |
| **Integration Tests** | Service interactions | `go test`, Docker, K3d | `TestClusterCreateWithK3d` |
| **E2E Tests** | Complete workflows | Full CLI, real clusters | `TestBootstrapWorkflow` |
| **Performance Tests** | Load and timing | `go test -bench`, profiling | `BenchmarkClusterCreation` |

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test ./internal/cluster/...

# Run specific test functions
go test -run TestClusterCreate ./internal/cluster/services

# Run tests with verbose output
go test -v ./...

# Run tests with race detection
go test -race ./...
```

### Integration Tests

Integration tests require Docker and use the `integration` build tag:

```bash
# Run integration tests (requires Docker)
go test -tags=integration ./...

# Run integration tests with cleanup
go test -tags=integration -v ./... -cleanup

# Run specific integration test suite
go test -tags=integration ./internal/cluster -run TestK3dIntegration

# Run integration tests with timeout
go test -tags=integration -timeout=10m ./...
```

### E2E Tests

End-to-end tests create real clusters and test complete workflows:

```bash
# Run E2E tests (requires Docker, K3d, Helm)
go test -tags=e2e ./test/e2e/...

# Run specific E2E test
go test -tags=e2e ./test/e2e -run TestBootstrapWorkflow

# Run E2E tests with custom timeout
go test -tags=e2e -timeout=20m ./test/e2e/...

# Run E2E tests with verbose logging
go test -tags=e2e -v ./test/e2e/... -args -openframe-debug
```

### Using Test Scripts

```bash
# Run unit tests only
./scripts/test/run-unit.sh

# Run integration tests with setup
./scripts/test/run-integration.sh

# Run E2E tests with full cleanup
./scripts/test/run-e2e.sh

# Run all tests in CI mode
make test-ci
```

## Writing New Tests

### Unit Test Example

```go
// internal/cluster/models/configuration_test.go
package models

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestClusterConfigValidation(t *testing.T) {
    tests := []struct {
        name        string
        config      ClusterConfig
        expectError bool
        errorMsg    string
    }{
        {
            name: "valid configuration",
            config: ClusterConfig{
                Name:        "test-cluster",
                Nodes:       3,
                KubeAPIPort: 6443,
                HTTPPort:    80,
                HTTPSPort:   443,
            },
            expectError: false,
        },
        {
            name: "invalid cluster name",
            config: ClusterConfig{
                Name:  "", // Empty name should fail
                Nodes: 3,
            },
            expectError: true,
            errorMsg:    "cluster name cannot be empty",
        },
        {
            name: "invalid node count",
            config: ClusterConfig{
                Name:  "test-cluster",
                Nodes: 0, // Zero nodes should fail
            },
            expectError: true,
            errorMsg:    "node count must be at least 1",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.config.Validate()
            
            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}

func TestClusterConfigDefaults(t *testing.T) {
    config := NewClusterConfig("test-cluster")
    
    assert.Equal(t, "test-cluster", config.Name)
    assert.Equal(t, 3, config.Nodes) // Default node count
    assert.Equal(t, 6443, config.KubeAPIPort) // Default API port
    assert.True(t, config.EnableRegistry) // Default registry enabled
}
```

### Integration Test Example

```go
//go:build integration
// +build integration

package services

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/flamingo-stack/openframe-cli/test/testutils"
)

func TestK3dClusterIntegration(t *testing.T) {
    // Skip if Docker is not available
    testutils.RequireDocker(t)
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()
    
    // Create unique cluster name for this test
    clusterName := testutils.GenerateClusterName("integration-test")
    defer testutils.CleanupCluster(t, clusterName)
    
    // Initialize service
    service := NewClusterService()
    require.NotNil(t, service)
    
    // Test cluster creation
    config := &ClusterConfig{
        Name:      clusterName,
        Nodes:     2, // Minimal for testing
        HTTPPort:  8080,
        HTTPSPort: 8443,
    }
    
    err := service.Create(ctx, config)
    require.NoError(t, err, "cluster creation should succeed")
    
    // Verify cluster exists
    clusters, err := service.List(ctx)
    require.NoError(t, err)
    
    found := false
    for _, cluster := range clusters {
        if cluster.Name == clusterName {
            found = true
            assert.Equal(t, "running", cluster.Status)
            assert.Equal(t, 2, len(cluster.Nodes))
            break
        }
    }
    assert.True(t, found, "cluster should be listed after creation")
    
    // Test cluster status
    status, err := service.Status(ctx, clusterName)
    require.NoError(t, err)
    assert.Equal(t, clusterName, status.Name)
    assert.Equal(t, "running", status.Status)
    assert.Greater(t, len(status.Nodes), 0)
    
    // Test cluster deletion
    err = service.Delete(ctx, clusterName)
    require.NoError(t, err, "cluster deletion should succeed")
    
    // Verify cluster is deleted
    clusters, err = service.List(ctx)
    require.NoError(t, err)
    
    for _, cluster := range clusters {
        assert.NotEqual(t, clusterName, cluster.Name, "cluster should not exist after deletion")
    }
}

func TestK3dProviderWithRegistry(t *testing.T) {
    testutils.RequireDocker(t)
    
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
    defer cancel()
    
    clusterName := testutils.GenerateClusterName("registry-test")
    defer testutils.CleanupCluster(t, clusterName)
    
    provider := NewK3dProvider()
    
    config := &ClusterConfig{
        Name:           clusterName,
        Nodes:          1,
        EnableRegistry: true,
        RegistryPort:   5001, // Use non-default port
    }
    
    // Create cluster with registry
    err := provider.Create(ctx, config)
    require.NoError(t, err)
    
    // Verify registry is accessible
    registryURL := fmt.Sprintf("localhost:%d", config.RegistryPort)
    assert.True(t, testutils.IsRegistryAccessible(registryURL))
}
```

### E2E Test Example

```go
//go:build e2e
// +build e2e

package e2e

import (
    "context"
    "os/exec"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/flamingo-stack/openframe-cli/test/testutils"
)

func TestBootstrapWorkflow(t *testing.T) {
    // Require all prerequisites
    testutils.RequireDocker(t)
    testutils.RequireK3d(t)
    testutils.RequireHelm(t)
    testutils.RequireKubectl(t)
    
    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
    defer cancel()
    
    clusterName := testutils.GenerateClusterName("e2e-bootstrap")
    defer testutils.CleanupCluster(t, clusterName)
    
    // Build OpenFrame CLI binary for testing
    binaryPath := testutils.BuildTestBinary(t)
    
    // Execute bootstrap command
    cmd := exec.CommandContext(ctx, binaryPath,
        "bootstrap", clusterName,
        "--deployment-mode=oss-tenant",
        "--non-interactive",
        "--verbose",
    )
    
    output, err := cmd.CombinedOutput()
    require.NoError(t, err, "bootstrap command should succeed: %s", string(output))
    
    // Verify cluster was created
    clusters := testutils.ListClusters(t, binaryPath)
    assert.Contains(t, clusters, clusterName)
    
    // Verify ArgoCD is installed
    testutils.WaitForPods(t, "argocd", "app=argocd-server", 2*time.Minute)
    
    // Verify ArgoCD applications are synced
    apps := testutils.GetArgoApplications(t, clusterName)
    assert.Greater(t, len(apps), 0, "should have ArgoCD applications")
    
    for _, app := range apps {
        assert.Equal(t, "Synced", app.Status, "application %s should be synced", app.Name)
        assert.Equal(t, "Healthy", app.Health, "application %s should be healthy", app.Name)
    }
    
    // Test cluster operations after bootstrap
    t.Run("ClusterOperations", func(t *testing.T) {
        // Test cluster status
        status := testutils.GetClusterStatus(t, binaryPath, clusterName)
        assert.Equal(t, "running", status.Status)
        assert.Greater(t, len(status.Nodes), 0)
        
        // Test cluster list
        clusters := testutils.ListClusters(t, binaryPath)
        assert.Contains(t, clusters, clusterName)
    })
    
    // Test chart operations after bootstrap
    t.Run("ChartOperations", func(t *testing.T) {
        // Verify chart status
        chartStatus := testutils.GetChartStatus(t, binaryPath, clusterName)
        assert.True(t, chartStatus.ArgoCD.Installed)
        assert.True(t, chartStatus.ArgoCD.Ready)
    })
}

func TestClusterLifecycleE2E(t *testing.T) {
    testutils.RequireDocker(t)
    testutils.RequireK3d(t)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    clusterName := testutils.GenerateClusterName("e2e-lifecycle")
    binaryPath := testutils.BuildTestBinary(t)
    
    // Test cluster creation
    t.Run("Create", func(t *testing.T) {
        cmd := exec.CommandContext(ctx, binaryPath,
            "cluster", "create", clusterName,
            "--nodes=2",
            "--non-interactive",
        )
        
        output, err := cmd.CombinedOutput()
        require.NoError(t, err, "cluster create should succeed: %s", string(output))
        
        // Verify cluster exists and is running
        testutils.WaitForCluster(t, clusterName, 2*time.Minute)
    })
    
    // Test cluster status
    t.Run("Status", func(t *testing.T) {
        status := testutils.GetClusterStatus(t, binaryPath, clusterName)
        assert.Equal(t, clusterName, status.Name)
        assert.Equal(t, "running", status.Status)
        assert.Equal(t, 3, len(status.Nodes)) // 1 server + 2 agents
    })
    
    // Test cluster list
    t.Run("List", func(t *testing.T) {
        clusters := testutils.ListClusters(t, binaryPath)
        assert.Contains(t, clusters, clusterName)
    })
    
    // Test cluster deletion
    t.Run("Delete", func(t *testing.T) {
        defer testutils.CleanupCluster(t, clusterName) // Ensure cleanup
        
        cmd := exec.CommandContext(ctx, binaryPath,
            "cluster", "delete", clusterName,
        )
        
        output, err := cmd.CombinedOutput()
        require.NoError(t, err, "cluster delete should succeed: %s", string(output))
        
        // Verify cluster is deleted
        testutils.WaitForClusterDeletion(t, clusterName, 2*time.Minute)
    })
}
```

### Mock Usage Example

```go
// Using gomock for service mocking
package services

import (
    "testing"
    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
)

func TestBootstrapServiceWithMocks(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    // Create mocks
    mockClusterService := NewMockClusterService(ctrl)
    mockChartService := NewMockChartService(ctrl)
    
    // Set up expectations
    clusterConfig := &ClusterConfig{Name: "test-cluster"}
    chartConfig := &ChartConfig{ArgoCD: true}
    
    mockClusterService.EXPECT().
        Create(gomock.Any(), clusterConfig).
        Return(nil).
        Times(1)
        
    mockChartService.EXPECT().
        Install(gomock.Any(), "test-cluster", chartConfig).
        Return(nil).
        Times(1)
    
    // Test the service
    service := &BootstrapService{
        clusterService: mockClusterService,
        chartService:   mockChartService,
    }
    
    err := service.Execute(context.Background(), clusterConfig, chartConfig)
    assert.NoError(t, err)
}
```

## Coverage Requirements

### Coverage Targets

| Component | Unit Test Coverage | Integration Coverage |
|-----------|-------------------|---------------------|
| **Models/Config** | 90%+ | N/A |
| **Service Logic** | 85%+ | 70%+ |
| **UI Components** | 80%+ | 60%+ |
| **Command Handlers** | 75%+ | 80%+ |
| **Overall Project** | 80%+ | 65%+ |

### Measuring Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out -covermode=atomic

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage by package
go tool cover -func=coverage.out | grep -E "(cluster|chart|bootstrap)"

# Generate coverage for specific tags
go test -tags=integration ./... -coverprofile=integration-coverage.out
```

### Coverage Analysis

```bash
# Install coverage analysis tools
go install github.com/boumenot/gocover-cobertura@latest

# Generate Cobertura XML for CI
gocover-cobertura < coverage.out > coverage.xml

# Check coverage threshold
./scripts/check-coverage.sh 80  # Fail if below 80%
```

## Test Utilities and Helpers

### Common Test Utilities

```go
// test/testutils/clusters.go
package testutils

import (
    "fmt"
    "math/rand"
    "testing"
    "time"
)

// GenerateClusterName creates a unique cluster name for testing
func GenerateClusterName(prefix string) string {
    rand.Seed(time.Now().UnixNano())
    return fmt.Sprintf("%s-%d", prefix, rand.Intn(10000))
}

// RequireDocker skips the test if Docker is not available
func RequireDocker(t *testing.T) {
    t.Helper()
    if !IsDockerAvailable() {
        t.Skip("Docker is not available")
    }
}

// CleanupCluster ensures test cluster is deleted
func CleanupCluster(t *testing.T, clusterName string) {
    t.Helper()
    t.Cleanup(func() {
        DeleteCluster(clusterName) // Best effort cleanup
    })
}

// WaitForPods waits for pods to be ready in namespace
func WaitForPods(t *testing.T, namespace, selector string, timeout time.Duration) {
    t.Helper()
    // Implementation...
}
```

### Test Fixtures

```go
// test/testutils/fixtures.go
package testutils

// TestClusterConfigs provides standard test configurations
var TestClusterConfigs = map[string]*ClusterConfig{
    "minimal": {
        Name:  "test-minimal",
        Nodes: 1,
    },
    "standard": {
        Name:        "test-standard",
        Nodes:       3,
        HTTPPort:    80,
        HTTPSPort:   443,
        EnableRegistry: true,
    },
    "development": {
        Name:           "test-dev",
        Nodes:          2,
        EnableRegistry: true,
        RegistryPort:   5001,
        DevMode:        true,
    },
}

// GetTestConfig returns a copy of named test configuration
func GetTestConfig(name string) *ClusterConfig {
    config := TestClusterConfigs[name]
    if config == nil {
        return nil
    }
    // Return deep copy
    return &ClusterConfig{*config}
}
```

## Continuous Integration

### GitHub Actions Example

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run Unit Tests
      run: |
        go test ./... -coverprofile=coverage.out
        go tool cover -func=coverage.out
    
    - name: Upload Coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  integration-tests:
    runs-on: ubuntu-latest
    services:
      docker:
        image: docker:dind
        options: --privileged
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install K3d
      run: |
        curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
    
    - name: Run Integration Tests
      run: |
        go test -tags=integration ./... -v

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Install Dependencies
      run: |
        curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
        curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
        sudo install kubectl /usr/local/bin/
        curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg
        echo "deb [signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
        sudo apt-get update && sudo apt-get install helm
    
    - name: Run E2E Tests
      run: |
        go test -tags=e2e ./test/e2e/... -v -timeout=30m
```

## Best Practices

### Test Organization

1. **Group Related Tests**: Use subtests (`t.Run()`) for related scenarios
2. **Descriptive Names**: Test names should clearly describe what's being tested
3. **Table-Driven Tests**: Use for testing multiple scenarios
4. **Setup/Teardown**: Use `t.Cleanup()` for proper resource cleanup

### Test Data Management

1. **Test Fixtures**: Use consistent test data and configurations
2. **Randomization**: Generate unique names to avoid test conflicts
3. **Isolation**: Each test should be independent and not affect others
4. **Cleanup**: Always clean up resources, even if tests fail

### Performance Testing

```go
func BenchmarkClusterCreation(b *testing.B) {
    if testing.Short() {
        b.Skip("skipping benchmark in short mode")
    }
    
    for i := 0; i < b.N; i++ {
        clusterName := fmt.Sprintf("bench-cluster-%d", i)
        config := &ClusterConfig{
            Name:  clusterName,
            Nodes: 1, // Minimal for benchmarking
        }
        
        b.StartTimer()
        err := service.Create(context.Background(), config)
        b.StopTimer()
        
        if err != nil {
            b.Fatalf("cluster creation failed: %v", err)
        }
        
        // Cleanup
        service.Delete(context.Background(), clusterName)
    }
}
```

---

## Summary

OpenFrame CLI's testing strategy ensures:

1. **Comprehensive Coverage**: Unit, integration, and E2E tests cover all critical paths
2. **Fast Development**: Unit tests provide immediate feedback
3. **Real-World Validation**: Integration tests verify actual tool interactions  
4. **User Journey Testing**: E2E tests validate complete workflows
5. **Quality Gates**: Coverage requirements and CI automation maintain standards

This testing foundation supports confident development and reliable releases. When contributing:

- Write tests for new functionality
- Maintain coverage requirements
- Use appropriate test types for different scenarios
- Follow established patterns and utilities

Next, explore [Contributing Guidelines](../contributing/guidelines.md) to learn about code style, PR processes, and collaboration workflows.