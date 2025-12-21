# Testing Overview

OpenFrame CLI maintains high code quality through comprehensive testing strategies. This guide covers the testing framework, how to run tests, write new tests, and maintain good test coverage across the codebase.

## Testing Philosophy

Our testing approach follows the testing pyramid with emphasis on fast feedback and reliable results:

```mermaid
pyramid
    title Testing Pyramid
    "E2E Tests<br/>Slow, Expensive<br/>High Confidence" : 10
    "Integration Tests<br/>Medium Speed<br/>Medium Confidence" : 30
    "Unit Tests<br/>Fast, Cheap<br/>Quick Feedback" : 60
```

### Testing Principles

1. **Fast Feedback**: Unit tests run in milliseconds
2. **Reliable**: Tests should not be flaky or environment-dependent  
3. **Maintainable**: Tests should be easy to understand and modify
4. **Comprehensive**: Cover both happy paths and edge cases
5. **Isolated**: Each test should be independent

## Test Structure and Organization

### Directory Structure

```
.
├── cmd/
│   ├── bootstrap/
│   │   ├── bootstrap.go
│   │   └── bootstrap_test.go        # Command layer tests
│   └── cluster/
│       ├── create.go
│       └── create_test.go
├── internal/
│   ├── bootstrap/
│   │   ├── service.go
│   │   ├── service_test.go         # Unit tests
│   │   └── service_integration_test.go  # Integration tests
│   ├── cluster/
│   │   ├── k3d_provider.go
│   │   └── k3d_provider_test.go
│   └── shared/
│       ├── ui/
│       │   ├── ui.go
│       │   └── ui_test.go
├── test/
│   ├── fixtures/                   # Test data and configurations
│   ├── mocks/                      # Generated mocks
│   ├── integration/                # Integration test suites
│   └── e2e/                        # End-to-end test suites
└── scripts/
    ├── test-unit.sh               # Test runner scripts
    ├── test-integration.sh
    └── test-e2e.sh
```

### Test Categories

| Test Type | Location | Purpose | Speed | Dependencies |
|-----------|----------|---------|-------|--------------|
| **Unit** | `*_test.go` | Test individual functions/methods | Fast | None (mocked) |
| **Integration** | `*_integration_test.go` | Test component interactions | Medium | External tools |
| **E2E** | `test/e2e/` | Test complete user workflows | Slow | Full environment |

## Running Tests

### Basic Test Commands

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run tests and generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/cluster/...
go test ./cmd/bootstrap/...
```

### Test Categories with Build Tags

OpenFrame CLI uses build tags to categorize tests:

```bash
# Run only unit tests (default)
go test -short ./...

# Run integration tests  
go test -tags=integration ./...

# Run end-to-end tests
go test -tags=e2e ./...

# Run all tests including slow ones
go test -tags="integration e2e" ./...
```

### Makefile Targets

```bash
# Run unit tests
make test

# Run all tests including integration
make test-all

# Run tests with coverage
make test-coverage

# Run integration tests only
make test-integration

# Run end-to-end tests
make test-e2e

# Generate coverage report
make coverage-html
```

### Continuous Testing

Set up file watching for continuous testing during development:

```bash
# Using entr (install with brew/apt)
find . -name "*.go" | entr -c go test ./...

# Using gotestsum for better output
gotestsum --watch

# Using air with test configuration
air -c .air-test.toml
```

## Writing Unit Tests

### Test Structure Pattern

Follow the Arrange-Act-Assert (AAA) pattern:

```go
func TestClusterService_Create_Success(t *testing.T) {
    // Arrange
    mockProvider := &MockK3dProvider{}
    mockUI := &MockUIService{}
    
    mockProvider.On("CreateCluster", "test-cluster", mock.AnythingOfType("ClusterConfig")).
        Return(nil)
    
    service := &ClusterService{
        k3dProvider: mockProvider,
        ui:          mockUI,
    }
    
    // Act
    err := service.Create(nil, []string{"test-cluster"})
    
    // Assert
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

### Mocking Dependencies

#### Using testify/mock

```go
//go:generate mockery --name=K3dProvider --output=../test/mocks
type K3dProvider interface {
    CreateCluster(name string, config ClusterConfig) error
    DeleteCluster(name string) error
    ListClusters() ([]Cluster, error)
}

// In tests
func TestClusterService_Create_ProviderError(t *testing.T) {
    mockProvider := &mocks.K3dProvider{}
    mockProvider.On("CreateCluster", "test", mock.Anything).
        Return(errors.New("k3d error"))
    
    service := NewClusterService(WithK3dProvider(mockProvider))
    
    err := service.Create(nil, []string{"test"})
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "k3d error")
}
```

#### Manual Mocks for Simple Cases

```go
type MockUIService struct {
    LoggedMessages []string
    PromptResponses map[string]string
}

func (m *MockUIService) ShowSuccess(msg string) {
    m.LoggedMessages = append(m.LoggedMessages, "SUCCESS: "+msg)
}

func (m *MockUIService) PromptForInput(prompt string) (string, error) {
    if response, exists := m.PromptResponses[prompt]; exists {
        return response, nil
    }
    return "", errors.New("unexpected prompt: " + prompt)
}
```

### Table-Driven Tests

For testing multiple scenarios:

```go
func TestValidateClusterName(t *testing.T) {
    tests := []struct {
        name        string
        clusterName string
        wantErr     bool
        errContains string
    }{
        {
            name:        "valid cluster name",
            clusterName: "my-cluster",
            wantErr:     false,
        },
        {
            name:        "empty name",
            clusterName: "",
            wantErr:     true,
            errContains: "cannot be empty",
        },
        {
            name:        "invalid characters",
            clusterName: "My_Cluster!",
            wantErr:     true,
            errContains: "invalid characters",
        },
        {
            name:        "too long",
            clusterName: strings.Repeat("a", 64),
            wantErr:     true,
            errContains: "too long",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateClusterName(tt.clusterName)
            
            if tt.wantErr {
                assert.Error(t, err)
                if tt.errContains != "" {
                    assert.Contains(t, err.Error(), tt.errContains)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Testing Error Handling

```go
func TestBootstrapService_Execute_PrerequisiteFailure(t *testing.T) {
    mockPrereqs := &MockPrerequisites{}
    mockPrereqs.On("Check").Return(errors.New("docker not found"))
    
    service := &BootstrapService{
        prerequisites: mockPrereqs,
    }
    
    err := service.Execute(nil, []string{})
    
    // Test specific error type
    var prereqErr *PrerequisiteError
    assert.True(t, errors.As(err, &prereqErr))
    assert.Equal(t, "docker not found", prereqErr.Tool)
}
```

## Integration Tests

Integration tests verify that components work together with real external dependencies.

### Integration Test Structure

```go
//go:build integration
// +build integration

func TestK3dProvider_CreateCluster_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup
    provider := NewK3dProvider()
    clusterName := "integration-test-" + generateRandomString(8)
    
    // Cleanup
    t.Cleanup(func() {
        provider.DeleteCluster(clusterName)
    })
    
    // Test
    config := ClusterConfig{
        Ports: []string{"8080:80"},
        Nodes: 1,
    }
    
    err := provider.CreateCluster(clusterName, config)
    assert.NoError(t, err)
    
    // Verify cluster exists
    clusters, err := provider.ListClusters()
    assert.NoError(t, err)
    
    var found bool
    for _, cluster := range clusters {
        if cluster.Name == clusterName {
            found = true
            break
        }
    }
    assert.True(t, found, "Cluster should exist after creation")
}
```

### Docker Integration Tests

```go
func TestDockerProvider_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    // Ensure Docker is available
    provider := NewDockerProvider()
    err := provider.Ping()
    assert.NoError(t, err, "Docker should be available for integration tests")
    
    // Test container operations
    containerID, err := provider.RunContainer("nginx:alpine", map[string]string{
        "80/tcp": "8080",
    })
    assert.NoError(t, err)
    assert.NotEmpty(t, containerID)
    
    t.Cleanup(func() {
        provider.RemoveContainer(containerID)
    })
    
    // Verify container is running
    containers, err := provider.ListContainers()
    assert.NoError(t, err)
    
    found := false
    for _, container := range containers {
        if container.ID == containerID {
            found = true
            assert.Equal(t, "running", container.Status)
            break
        }
    }
    assert.True(t, found)
}
```

## End-to-End Tests

E2E tests verify complete user workflows with a real environment.

### E2E Test Structure

```go
//go:build e2e
// +build e2e

func TestBootstrapWorkflow_E2E(t *testing.T) {
    // Setup test environment
    testDir := setupTestDirectory(t)
    clusterName := "e2e-test-" + generateRandomString(8)
    
    t.Cleanup(func() {
        cleanupE2ETest(t, clusterName, testDir)
    })
    
    // Test complete bootstrap workflow
    cmd := exec.Command("./openframe", "bootstrap", clusterName, 
        "--deployment-mode=oss-tenant", 
        "--non-interactive")
    cmd.Dir = testDir
    
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, "Bootstrap should succeed. Output: %s", string(output))
    
    // Verify cluster exists
    verifyClusterExists(t, clusterName)
    
    // Verify ArgoCD is installed
    verifyArgoCDInstalled(t, clusterName)
    
    // Verify applications are synced
    verifyApplicationsSynced(t, clusterName)
}

func verifyClusterExists(t *testing.T, clusterName string) {
    cmd := exec.Command("k3d", "cluster", "list", clusterName)
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, "Failed to check cluster: %s", string(output))
    assert.Contains(t, string(output), clusterName)
}

func verifyArgoCDInstalled(t *testing.T, clusterName string) {
    cmd := exec.Command("kubectl", "--context", "k3d-"+clusterName, 
        "get", "pods", "-n", "argocd")
    output, err := cmd.CombinedOutput()
    assert.NoError(t, err, "Failed to check ArgoCD: %s", string(output))
    assert.Contains(t, string(output), "argocd-server")
}
```

### CLI Testing Utilities

Create utilities for testing CLI behavior:

```go
// test/utils/cli.go
type CLITestRunner struct {
    Binary string
    Env    []string
}

func NewCLITestRunner(binary string) *CLITestRunner {
    return &CLITestRunner{
        Binary: binary,
        Env:    os.Environ(),
    }
}

func (r *CLITestRunner) Run(args ...string) (*CLIResult, error) {
    cmd := exec.Command(r.Binary, args...)
    cmd.Env = r.Env
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    err := cmd.Run()
    
    return &CLIResult{
        ExitCode: cmd.ProcessState.ExitCode(),
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
    }, err
}

type CLIResult struct {
    ExitCode int
    Stdout   string
    Stderr   string
}

func (r *CLIResult) AssertSuccess(t *testing.T) {
    assert.Equal(t, 0, r.ExitCode, "Command should succeed. Stderr: %s", r.Stderr)
}

func (r *CLIResult) AssertContains(t *testing.T, text string) {
    assert.Contains(t, r.Stdout, text)
}
```

## Test Data Management

### Fixtures

Store test data in dedicated fixture files:

```go
// test/fixtures/cluster_configs.go
var TestClusterConfigs = map[string]ClusterConfig{
    "minimal": {
        Nodes: 1,
        Ports: []string{},
    },
    "development": {
        Nodes: 1,
        Ports: []string{"8080:80", "8443:443"},
        Registry: RegistryConfig{
            Create: true,
            Port:   "5000",
        },
    },
    "multi-node": {
        Nodes: 3,
        Ports: []string{"8080:80"},
    },
}

func GetTestConfig(name string) ClusterConfig {
    config, exists := TestClusterConfigs[name]
    if !exists {
        panic("Unknown test config: " + name)
    }
    return config
}
```

### Golden Files

For testing complex outputs:

```go
func TestFormatClusterStatus(t *testing.T) {
    cluster := &Cluster{
        Name:   "test-cluster",
        Status: "running", 
        Nodes:  []Node{{Name: "node1", Status: "ready"}},
    }
    
    output := FormatClusterStatus(cluster)
    
    // Compare with golden file
    goldenFile := "testdata/cluster_status.golden"
    if *update {
        ioutil.WriteFile(goldenFile, []byte(output), 0644)
    }
    
    expected, err := ioutil.ReadFile(goldenFile)
    assert.NoError(t, err)
    assert.Equal(t, string(expected), output)
}
```

## Coverage Requirements

### Coverage Targets

| Component | Target Coverage | Reason |
|-----------|-----------------|---------|
| **Service Layer** | 90%+ | Core business logic |
| **Command Layer** | 70%+ | CLI interface, harder to test |
| **Shared Utilities** | 95%+ | Reused across components |
| **Providers** | 80%+ | External integrations |

### Measuring Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage by package
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage threshold
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//'
```

### Coverage Configuration

Add to Makefile:

```makefile
COVERAGE_THRESHOLD := 80

.PHONY: test-coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$(echo "$$coverage < $(COVERAGE_THRESHOLD)" | bc)" -eq 1 ]; then \
		echo "❌ Coverage $$coverage% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✅ Coverage $$coverage% meets threshold"; \
	fi
```

## Best Practices

### Test Naming

```go
// Good: TestFunctionName_Scenario_ExpectedBehavior
func TestClusterService_Create_WithInvalidName_ReturnsValidationError(t *testing.T)

// Good: TestFunctionName_ExpectedBehavior  
func TestClusterService_Create_Success(t *testing.T)

// Bad: TestCreate(t *testing.T)
// Bad: TestClusterCreation(t *testing.T)
```

### Test Independence

```go
// ✅ Good: Each test is independent
func TestClusterService_Create(t *testing.T) {
    service := NewClusterService()  // Fresh instance
    // Test logic
}

// ❌ Bad: Tests depend on shared state
var globalService *ClusterService

func TestClusterService_Create(t *testing.T) {
    globalService.Create(...)  // Shared state can cause flaky tests
}
```

### Error Testing

```go
// ✅ Good: Test specific error types and messages
func TestValidateClusterName_EmptyName(t *testing.T) {
    err := ValidateClusterName("")
    
    assert.Error(t, err)
    assert.IsType(t, &ValidationError{}, err)
    assert.Contains(t, err.Error(), "empty")
}

// ❌ Bad: Only test that error occurred
func TestValidateClusterName_EmptyName(t *testing.T) {
    err := ValidateClusterName("")
    assert.Error(t, err)  // Too generic
}
```

## Continuous Integration

### GitHub Actions Integration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v4
      with:
        go-version: '1.23'
    
    - name: Run unit tests
      run: make test
    
    - name: Run integration tests
      run: make test-integration
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

## Debugging Test Failures

### Verbose Test Output

```bash
# Run with verbose output
go test -v ./internal/cluster/...

# Run specific test
go test -v -run TestClusterService_Create ./internal/cluster/...

# Run tests with debugging
OPENFRAME_DEBUG=true go test -v ./...
```

### Test Debugging Tools

```go
func TestClusterService_Create_Debug(t *testing.T) {
    if testing.Verbose() {
        // Additional logging for debugging
        log.SetLevel(log.DebugLevel)
    }
    
    // Use testify's require for early failure
    require.NotNil(t, service)
    require.NoError(t, err)
}
```

## Next Steps

Now that you understand the testing strategy:

1. **Run the Test Suite** - Execute `make test` to see current coverage
2. **Write Your First Test** - Add a test for a new feature or fix
3. **Review Test Coverage** - Use `make test-coverage` to identify gaps
4. **Read Contributing Guidelines** - Understand code quality expectations in [Contributing Guidelines](../contributing/guidelines.md)

Quality testing is essential for maintaining OpenFrame CLI's reliability and enabling confident refactoring. Happy testing! 🧪