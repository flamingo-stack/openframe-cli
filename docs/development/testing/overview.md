# Testing Overview

This guide covers the comprehensive testing strategy for OpenFrame CLI, including test structure, running tests, writing new tests, and maintaining test coverage. Our testing approach ensures reliability, maintainability, and confidence in the codebase.

> **Prerequisites**: Complete [Environment Setup](../setup/environment.md) and [Local Development](../setup/local-development.md) setup.

## 🎯 Testing Philosophy

OpenFrame CLI follows a multi-layered testing approach:

### Testing Pyramid

```mermaid
pyramid LR
    subgraph "Testing Pyramid"
        E2E[End-to-End Tests<br/>Few, High Value]
        INTEGRATION[Integration Tests<br/>Moderate Coverage]
        UNIT[Unit Tests<br/>High Coverage, Fast]
    end
    
    subgraph "Test Characteristics"
        SPEED[Speed: Fast → Slow]
        COST[Cost: Low → High] 
        ISOLATION[Isolation: High → Low]
        CONFIDENCE[Confidence: Medium → High]
    end
    
    UNIT --> SPEED
    INTEGRATION --> COST
    E2E --> ISOLATION
    E2E --> CONFIDENCE
```

### Testing Principles

| Principle | Description | Implementation |
|-----------|-------------|----------------|
| **Fast Feedback** | Tests should run quickly to enable rapid development | Unit tests complete in <1s, integration in <30s |
| **Reliable** | Tests should be deterministic and not flaky | Mock external dependencies, clean state |
| **Maintainable** | Tests should be easy to understand and modify | Clear naming, helper functions, good structure |
| **Comprehensive** | Tests should cover critical paths and edge cases | Aim for >80% coverage on business logic |

## 📁 Test Organization

### Directory Structure

```text
tests/
├── integration/               # Integration tests
│   ├── common/               # Shared integration test utilities
│   │   ├── cli_runner.go    # CLI execution helpers
│   │   ├── cluster_management.go # Cluster setup/teardown
│   │   └── dependencies.go   # External tool management
│   ├── bootstrap_test.go     # Bootstrap command integration tests
│   ├── cluster_test.go       # Cluster command integration tests
│   └── dev_test.go          # Dev command integration tests
│
├── mocks/                    # Test mocks and stubs
│   ├── dev/                 # Development tool mocks
│   │   └── kubernetes.go    # Kubernetes API mocks
│   └── providers/           # Provider mocks
│       ├── k3d.go          # K3d provider mock
│       └── helm.go         # Helm provider mock
│
├── testutil/                 # Test utilities and helpers
│   ├── assertions.go        # Custom test assertions
│   ├── cluster.go          # Test cluster management
│   ├── patterns.go         # Common test patterns
│   ├── setup.go            # Test environment setup
│   └── utilities.go        # General test utilities
│
└── fixtures/                # Test data and configurations
    ├── charts/              # Test Helm charts
    ├── configs/             # Test configuration files
    └── manifests/           # Test Kubernetes manifests
```

### Test Types by Location

| Test Type | Location | Purpose | Scope |
|-----------|----------|---------|--------|
| **Unit Tests** | `*_test.go` (same package) | Test individual functions/methods | Single function or struct |
| **Integration Tests** | `tests/integration/` | Test component interactions | Multiple components |
| **Mock Tests** | `tests/mocks/` | Test with external dependencies mocked | Service-level testing |

## 🧪 Unit Testing

### Unit Test Structure

```go
// Example unit test structure
package cluster

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestClusterService_Create(t *testing.T) {
    tests := []struct {
        name        string
        clusterName string
        config      ClusterConfig
        mockSetup   func(*MockProvider)
        wantErr     bool
        wantErrMsg  string
    }{
        {
            name:        "successful cluster creation",
            clusterName: "test-cluster",
            config:      ClusterConfig{Nodes: 3},
            mockSetup: func(m *MockProvider) {
                m.On("Create", "test-cluster", mock.Anything).Return(nil)
            },
            wantErr: false,
        },
        {
            name:        "invalid cluster name",
            clusterName: "",
            config:      ClusterConfig{},
            mockSetup:   func(m *MockProvider) {},
            wantErr:     true,
            wantErrMsg:  "cluster name cannot be empty",
        },
        {
            name:        "provider failure",
            clusterName: "test-cluster",
            config:      ClusterConfig{Nodes: 3},
            mockSetup: func(m *MockProvider) {
                m.On("Create", "test-cluster", mock.Anything).Return(errors.New("provider error"))
            },
            wantErr:    true,
            wantErrMsg: "failed to create cluster",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockProvider := &MockProvider{}
            tt.mockSetup(mockProvider)
            
            service := &ClusterService{
                provider: mockProvider,
            }

            // Execute
            err := service.Create(tt.clusterName, tt.config)

            // Assert
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.wantErrMsg)
            } else {
                require.NoError(t, err)
            }

            // Verify mock expectations
            mockProvider.AssertExpectations(t)
        })
    }
}
```

### Running Unit Tests

```bash
# Run all unit tests
go test ./...

# Run unit tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run tests for specific package
go test ./internal/cluster/...

# Run specific test
go test ./internal/cluster -run TestClusterService_Create

# Run with verbose output
go test -v ./internal/cluster

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...
```

### Test Coverage Requirements

| Component | Minimum Coverage | Target Coverage |
|-----------|------------------|-----------------|
| **Service Layer** | 80% | 90% |
| **Domain Models** | 70% | 85% |
| **Utilities** | 75% | 90% |
| **CLI Commands** | 60% | 80% |

## 🔗 Integration Testing

### Integration Test Structure

```go
// Example integration test
package integration

import (
    "testing"
    "github.com/flamingo-stack/openframe-cli/tests/testutil"
)

func TestBootstrapCommand_E2E(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    tests := []struct {
        name           string
        args           []string
        envVars        map[string]string
        setupCluster   bool
        wantExitCode   int
        wantContains   []string
        wantNotContain []string
    }{
        {
            name:         "bootstrap with oss-tenant mode",
            args:         []string{"bootstrap", "test-cluster", "--deployment-mode=oss-tenant", "--non-interactive"},
            envVars:      map[string]string{"OPENFRAME_LOG_LEVEL": "debug"},
            setupCluster: false,
            wantExitCode: 0,
            wantContains: []string{
                "✅ Cluster created successfully",
                "✅ ArgoCD installed",
                "✅ Applications synced",
            },
        },
        {
            name:         "bootstrap with existing cluster",
            args:         []string{"bootstrap", "existing-cluster"},
            setupCluster: true,
            wantExitCode: 1,
            wantContains: []string{"cluster already exists"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup test environment
            testEnv := testutil.NewTestEnvironment(t)
            defer testEnv.Cleanup()

            if tt.setupCluster {
                testEnv.CreateTestCluster("existing-cluster")
            }

            // Set environment variables
            for key, value := range tt.envVars {
                testEnv.SetEnv(key, value)
            }

            // Execute command
            result := testEnv.RunCLI(tt.args...)

            // Assertions
            assert.Equal(t, tt.wantExitCode, result.ExitCode)
            
            for _, want := range tt.wantContains {
                assert.Contains(t, result.Output, want)
            }
            
            for _, notWant := range tt.wantNotContain {
                assert.NotContains(t, result.Output, notWant)
            }
        })
    }
}
```

### Running Integration Tests

```bash
# Run all integration tests
go test -tags=integration ./tests/integration/...

# Run integration tests with setup
make test-integration

# Run specific integration test
go test -tags=integration ./tests/integration -run TestBootstrapCommand

# Run integration tests with timeout
go test -tags=integration -timeout 10m ./tests/integration/...

# Skip integration tests (unit tests only)
go test -short ./...
```

### Integration Test Environment

```bash
# Setup integration test environment
cat > scripts/setup-integration-tests.sh << 'EOF'
#!/bin/bash
set -e

echo "Setting up integration test environment..."

# Create test cluster for integration tests
k3d cluster create openframe-integration-test \
  --agents 2 \
  --port "8080:80@loadbalancer" \
  --wait

# Export kubeconfig for tests
export KUBECONFIG=$(k3d kubeconfig write openframe-integration-test)

# Verify cluster is ready
kubectl cluster-info
kubectl get nodes

echo "Integration test environment ready!"
EOF

chmod +x scripts/setup-integration-tests.sh
```

## 🤖 Mock Testing

### Mock Generation

We use `mockery` for generating mocks from interfaces:

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks for all interfaces
mockery --all --dir internal/ --output tests/mocks/

# Generate mock for specific interface
mockery --name ClusterProvider --dir internal/cluster/ --output tests/mocks/cluster/
```

### Mock Usage Example

```go
// Mock provider interface
type MockClusterProvider struct {
    mock.Mock
}

func (m *MockClusterProvider) Create(name string, config *ClusterConfig) error {
    args := m.Called(name, config)
    return args.Error(0)
}

func (m *MockClusterProvider) Delete(name string) error {
    args := m.Called(name)
    return args.Error(0)
}

// Test using mock
func TestClusterService_CreateAndDelete(t *testing.T) {
    mockProvider := &MockClusterProvider{}
    
    // Setup expectations
    mockProvider.On("Create", "test-cluster", mock.AnythingOfType("*ClusterConfig")).Return(nil)
    mockProvider.On("Delete", "test-cluster").Return(nil)
    
    service := &ClusterService{provider: mockProvider}
    
    // Test create
    err := service.Create("test-cluster", &ClusterConfig{})
    assert.NoError(t, err)
    
    // Test delete
    err = service.Delete("test-cluster")
    assert.NoError(t, err)
    
    // Verify all expectations were met
    mockProvider.AssertExpectations(t)
}
```

## 🛠️ Test Utilities

### Test Helper Functions

```go
// testutil/setup.go
package testutil

import (
    "testing"
    "os"
    "path/filepath"
)

// TestEnvironment provides isolated test environment
type TestEnvironment struct {
    t           *testing.T
    tempDir     string
    originalEnv map[string]string
}

func NewTestEnvironment(t *testing.T) *TestEnvironment {
    tempDir, err := os.MkdirTemp("", "openframe-test-*")
    require.NoError(t, err)
    
    return &TestEnvironment{
        t:           t,
        tempDir:     tempDir,
        originalEnv: make(map[string]string),
    }
}

func (te *TestEnvironment) SetEnv(key, value string) {
    te.originalEnv[key] = os.Getenv(key)
    os.Setenv(key, value)
}

func (te *TestEnvironment) CreateTempFile(name, content string) string {
    filePath := filepath.Join(te.tempDir, name)
    err := os.WriteFile(filePath, []byte(content), 0644)
    require.NoError(te.t, err)
    return filePath
}

func (te *TestEnvironment) Cleanup() {
    // Restore environment variables
    for key, originalValue := range te.originalEnv {
        if originalValue == "" {
            os.Unsetenv(key)
        } else {
            os.Setenv(key, originalValue)
        }
    }
    
    // Remove temp directory
    os.RemoveAll(te.tempDir)
}
```

### CLI Test Runner

```go
// testutil/cli_runner.go
package testutil

import (
    "bytes"
    "os/exec"
    "strings"
)

type CLIResult struct {
    ExitCode int
    Output   string
    Error    string
}

func (te *TestEnvironment) RunCLI(args ...string) *CLIResult {
    // Build OpenFrame CLI binary for testing
    binaryPath := filepath.Join(te.tempDir, "openframe")
    buildCmd := exec.Command("go", "build", "-o", binaryPath, "../../main.go")
    err := buildCmd.Run()
    require.NoError(te.t, err)
    
    // Run the CLI command
    cmd := exec.Command(binaryPath, args...)
    cmd.Dir = te.tempDir
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err = cmd.Run()
    
    exitCode := 0
    if err != nil {
        if exitError, ok := err.(*exec.ExitError); ok {
            exitCode = exitError.ExitCode()
        }
    }
    
    return &CLIResult{
        ExitCode: exitCode,
        Output:   stdout.String(),
        Error:    stderr.String(),
    }
}
```

## 📊 Test Coverage and Quality

### Coverage Analysis

```bash
# Generate detailed coverage report
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold of 80%"
    exit 1
fi
```

### Quality Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Test Coverage** | >80% | `go tool cover -func=coverage.out` |
| **Test Speed** | <30s for full suite | `go test -timeout 30s ./...` |
| **Flaky Tests** | 0% failure rate | CI/CD monitoring |
| **Test Maintainability** | Clear, readable tests | Code reviews |

### Automated Quality Checks

```bash
# Create test quality script
cat > scripts/test-quality.sh << 'EOF'
#!/bin/bash
set -e

echo "Running test quality checks..."

# Run tests with coverage
echo "Running tests with coverage..."
go test -race -coverprofile=coverage.out -timeout 30s ./...

# Check coverage threshold
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
echo "Test coverage: $COVERAGE%"

if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "❌ Coverage $COVERAGE% is below threshold of 80%"
    exit 1
fi

# Check for race conditions
echo "Running race detection tests..."
go test -race ./...

# Run integration tests if not in CI short mode
if [ "$CI_SHORT" != "true" ]; then
    echo "Running integration tests..."
    go test -tags=integration ./tests/integration/...
fi

echo "✅ All test quality checks passed!"
EOF

chmod +x scripts/test-quality.sh
```

## 🚀 Writing New Tests

### Test Writing Guidelines

1. **Follow the AAA Pattern**:
   ```go
   func TestExample(t *testing.T) {
       // Arrange - Set up test conditions
       input := "test-input"
       expected := "expected-output"
       
       // Act - Execute the function under test
       result := functionUnderTest(input)
       
       // Assert - Verify the results
       assert.Equal(t, expected, result)
   }
   ```

2. **Use Descriptive Test Names**:
   ```go
   // ✅ Good
   func TestClusterService_Create_WithValidName_ReturnsSuccess(t *testing.T) {}
   
   // ❌ Bad
   func TestCreate(t *testing.T) {}
   ```

3. **Test Edge Cases**:
   ```go
   func TestValidateClusterName(t *testing.T) {
       tests := []struct {
           name     string
           input    string
           wantErr  bool
       }{
           {"valid name", "my-cluster", false},
           {"empty name", "", true},
           {"too long", strings.Repeat("a", 64), true},
           {"invalid chars", "my_cluster", true},
           {"starts with number", "1cluster", true},
       }
       // ... test implementation
   }
   ```

### Test Templates

#### Unit Test Template

```go
func TestServiceName_MethodName(t *testing.T) {
    tests := []struct {
        name      string
        input     InputType
        mockSetup func(*MockDependency)
        want      OutputType
        wantErr   bool
    }{
        {
            name:  "successful case",
            input: InputType{/* valid input */},
            mockSetup: func(m *MockDependency) {
                m.On("Method", mock.Anything).Return(nil)
            },
            want:    OutputType{/* expected output */},
            wantErr: false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockDep := &MockDependency{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockDep)
            }
            
            service := &ServiceName{
                dependency: mockDep,
            }

            // Execute
            got, err := service.MethodName(tt.input)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, got)
            }
            
            mockDep.AssertExpectations(t)
        })
    }
}
```

#### Integration Test Template

```go
func TestCommandName_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    testEnv := testutil.NewTestEnvironment(t)
    defer testEnv.Cleanup()

    // Setup test conditions
    testEnv.SetEnv("OPENFRAME_LOG_LEVEL", "debug")

    // Execute command
    result := testEnv.RunCLI("command", "subcommand", "--flag=value")

    // Assertions
    assert.Equal(t, 0, result.ExitCode)
    assert.Contains(t, result.Output, "expected output")
}
```

## 🔍 Debugging Tests

### Test Debugging Strategies

```bash
# Run single test with verbose output
go test -v ./internal/cluster -run TestClusterService_Create

# Run test with debugger
dlv test ./internal/cluster -- -test.run TestClusterService_Create

# Add debug prints in tests
t.Logf("Debug: variable value = %v", variable)

# Use testify require for early exit on failures
require.NoError(t, err, "Setup failed, cannot continue test")
```

### Common Test Issues

| Issue | Symptom | Solution |
|-------|---------|----------|
| **Flaky tests** | Intermittent failures | Remove time dependencies, improve isolation |
| **Slow tests** | Tests timeout | Mock external calls, optimize setup/teardown |
| **Resource leaks** | Memory/disk usage grows | Ensure proper cleanup in test teardown |
| **Test pollution** | Tests affect each other | Improve test isolation, reset state |

## 📈 Continuous Testing

### CI/CD Integration

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          
      - name: Run tests
        run: |
          go test -race -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
          
      - name: Run integration tests
        run: |
          make setup-integration-env
          go test -tags=integration ./tests/integration/...
          
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

### Pre-commit Hooks

```bash
# .git/hooks/pre-commit
#!/bin/bash
set -e

echo "Running pre-commit tests..."

# Run fast tests
go test -short ./...

# Run linter
golangci-lint run

echo "Pre-commit checks passed!"
```

## 📚 Testing Resources

### Testing Tools
- **[testify](https://github.com/stretchr/testify)** - Testing toolkit
- **[mockery](https://github.com/vektra/mockery)** - Mock generation
- **[ginkgo](https://github.com/onsi/ginkgo)** - BDD testing framework
- **[gomega](https://github.com/onsi/gomega)** - Matcher library

### Best Practices
- **[Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)**
- **[Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)**
- **[Advanced Testing](https://golang.org/doc/tutorial/fuzz)**

---

*Ready to contribute to OpenFrame CLI? Check out our [contributing guidelines](../contributing/guidelines.md) to get started with the development workflow.*