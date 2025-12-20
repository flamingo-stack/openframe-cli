# OpenFrame CLI Architecture Overview

This document provides a comprehensive technical overview of the OpenFrame CLI architecture, designed for engineers working on or extending the codebase.

## High-Level Architecture

OpenFrame CLI follows a layered, modular architecture that separates concerns between command handling, business logic, and external integrations. The design emphasizes testability, extensibility, and clear separation of responsibilities.

```mermaid
graph TB
    subgraph "CLI Interface Layer"
        ROOT[Root Command]
        CLUSTER[Cluster Commands]
        CHART[Chart Commands]
        BOOTSTRAP[Bootstrap Command]
        DEV[Dev Commands]
    end
    
    subgraph "Service Layer"
        CS[Cluster Service]
        CHS[Chart Service] 
        BS[Bootstrap Service]
        IS[Intercept Service]
        SS[Scaffold Service]
    end
    
    subgraph "Provider Abstraction Layer"
        CP[Cluster Providers]
        HP[Helm Provider]
        TP[Telepresence Provider]
        SP[Skaffold Provider]
    end
    
    subgraph "Infrastructure Layer"
        K3D[K3d Implementation]
        KIND[Kind Implementation]
        HELM[Helm Client]
        KUBECTL[Kubectl Client]
        DOCKER[Docker Client]
    end
    
    subgraph "External Systems"
        K8S[Kubernetes Clusters]
        REGISTRY[Container Registries]
        GIT[Git Repositories]
        DOCKER_ENGINE[Docker Engine]
    end
    
    ROOT --> CLUSTER
    ROOT --> CHART
    ROOT --> BOOTSTRAP
    ROOT --> DEV
    
    CLUSTER --> CS
    CHART --> CHS
    BOOTSTRAP --> BS
    DEV --> IS
    DEV --> SS
    
    CS --> CP
    CHS --> HP
    IS --> TP
    SS --> SP
    
    CP --> K3D
    CP --> KIND
    HP --> HELM
    TP --> KUBECTL
    SS --> KUBECTL
    
    K3D --> DOCKER_ENGINE
    KIND --> DOCKER_ENGINE
    HELM --> K8S
    KUBECTL --> K8S
    
    CHS --> GIT
    SS --> REGISTRY
```

## Core Components and Responsibilities

| Component | Package | Primary Responsibilities | Key Interfaces |
|-----------|---------|-------------------------|----------------|
| **Root Command** | `cmd/root.go` | CLI entry point, global flags, version management | `cobra.Command` |
| **Cluster Management** | `cmd/cluster/` | K3d/Kind cluster lifecycle operations | `ClusterProvider` |
| **Chart Management** | `cmd/chart/` | Helm chart installation, ArgoCD setup | `ChartInstaller` |
| **Bootstrap Orchestration** | `cmd/bootstrap/` | End-to-end environment provisioning | `BootstrapService` |
| **Dev Tools** | `cmd/dev/` | Telepresence intercepts, Skaffold workflows | `InterceptService`, `ScaffoldService` |
| **Business Services** | `internal/*/services/` | Core business logic, workflow orchestration | Service interfaces |
| **Data Models** | `internal/*/models/` | Configuration structures, validation | Validation interfaces |
| **UI Components** | `internal/*/ui/` | Interactive prompts, output formatting | `UIHandler` |
| **Shared Utilities** | `internal/shared/` | Cross-cutting concerns (config, errors, execution) | Utility interfaces |

## Data Flow Architecture

### Bootstrap Workflow

The bootstrap command orchestrates the most complex workflow in the system:

```mermaid
sequenceDiagram
    participant User
    participant Bootstrap as Bootstrap Service
    participant Cluster as Cluster Service
    participant Chart as Chart Service
    participant K3d as K3d Provider
    participant Helm as Helm Client
    participant ArgoCD
    
    Note over User,ArgoCD: Complete Environment Bootstrap
    
    User->>Bootstrap: openframe bootstrap [cluster-name]
    Bootstrap->>Bootstrap: Parse deployment mode & flags
    Bootstrap->>Bootstrap: Validate prerequisites
    
    Note over Bootstrap,Chart: Phase 1: Cluster Creation
    Bootstrap->>Cluster: CreateClusterWithPrerequisites()
    Cluster->>Cluster: Validate cluster configuration
    Cluster->>K3d: CreateCluster(config)
    K3d->>K3d: Generate k3d configuration
    K3d-->>Cluster: Cluster created successfully
    Cluster-->>Bootstrap: Cluster ready
    
    Note over Bootstrap,ArgoCD: Phase 2: Chart Installation
    Bootstrap->>Chart: InstallWithDeploymentMode()
    Chart->>Chart: Generate SSL certificates
    Chart->>Helm: Install ArgoCD Helm chart
    Helm->>Helm: Create ArgoCD namespace
    Helm->>Helm: Deploy ArgoCD components
    Helm-->>Chart: ArgoCD ready
    
    Chart->>Chart: Generate app-of-apps configuration
    Chart->>Helm: Install app-of-apps chart
    Helm->>ArgoCD: Deploy OpenFrame applications
    ArgoCD->>ArgoCD: Sync applications from Git
    ArgoCD-->>Helm: Applications deployed
    Helm-->>Chart: Installation complete
    Chart-->>Bootstrap: Charts installed
    
    Bootstrap-->>User: Environment ready
```

### Command Execution Pattern

All commands follow a consistent execution pattern:

```mermaid
sequenceDiagram
    participant CLI as CLI Command
    participant Wrapper as Command Wrapper
    participant Service as Business Service
    participant Provider as Provider Implementation
    participant External as External System
    
    CLI->>Wrapper: RunE with args
    Wrapper->>Wrapper: Display logo & branding
    Wrapper->>Wrapper: Validate prerequisites
    Wrapper->>Wrapper: Parse and validate flags
    Wrapper->>Service: Execute business logic
    
    Service->>Service: Validate configuration
    Service->>Provider: Call provider interface
    Provider->>External: Execute system commands
    External-->>Provider: Return results
    Provider-->>Service: Return processed results
    Service-->>Wrapper: Return execution results
    Wrapper-->>CLI: Return success/error
```

## Module Dependencies

### Dependency Graph

```mermaid
graph LR
    subgraph "Command Layer"
        CMD_ROOT[cmd/root]
        CMD_CLUSTER[cmd/cluster]
        CMD_CHART[cmd/chart] 
        CMD_BOOTSTRAP[cmd/bootstrap]
        CMD_DEV[cmd/dev]
    end
    
    subgraph "Internal Services"
        SVC_CLUSTER[internal/cluster]
        SVC_CHART[internal/chart]
        SVC_BOOTSTRAP[internal/bootstrap]
        SVC_DEV[internal/dev]
        SVC_SHARED[internal/shared]
    end
    
    subgraph "External Dependencies"
        COBRA[github.com/spf13/cobra]
        VIPER[github.com/spf13/viper]
        SURVEY[github.com/AlecAivazis/survey/v2]
        YAML[gopkg.in/yaml.v3]
    end
    
    CMD_ROOT --> CMD_CLUSTER
    CMD_ROOT --> CMD_CHART
    CMD_ROOT --> CMD_BOOTSTRAP
    CMD_ROOT --> CMD_DEV
    
    CMD_CLUSTER --> SVC_CLUSTER
    CMD_CHART --> SVC_CHART
    CMD_BOOTSTRAP --> SVC_BOOTSTRAP
    CMD_DEV --> SVC_DEV
    
    SVC_BOOTSTRAP --> SVC_CLUSTER
    SVC_BOOTSTRAP --> SVC_CHART
    
    SVC_CLUSTER --> SVC_SHARED
    SVC_CHART --> SVC_SHARED
    SVC_DEV --> SVC_SHARED
    
    CMD_ROOT --> COBRA
    SVC_SHARED --> VIPER
    SVC_CLUSTER --> SURVEY
    SVC_CHART --> YAML
```

### Import Rules

1. **Commands** only import their corresponding services
2. **Services** can import shared utilities and other services
3. **Providers** are isolated and only imported by their services
4. **Shared** packages have no internal dependencies
5. **No circular dependencies** between internal packages

## Key Design Patterns

### 1. Provider Pattern

Abstract external tool integrations behind interfaces:

```go
// Provider interface abstraction
type ClusterProvider interface {
    Create(config ClusterConfig) error
    Delete(name string) error
    List() ([]ClusterInfo, error)
    GetStatus(name string) (ClusterStatus, error)
}

// K3d implementation
type K3dProvider struct {
    executor command.Executor
    config   K3dConfig
}

func (p *K3dProvider) Create(config ClusterConfig) error {
    k3dConfig := p.buildK3dConfig(config)
    return p.executor.Execute("k3d", "cluster", "create", k3dConfig...)
}
```

### 2. Service Pattern

Encapsulate business logic in service structs:

```go
type ClusterService struct {
    provider     ClusterProvider
    ui          UIHandler
    prereqChecker PrerequisiteChecker
}

func (s *ClusterService) CreateCluster(config ClusterConfig) error {
    if err := s.prereqChecker.CheckPrerequisites(); err != nil {
        return err
    }
    
    if err := s.ui.ConfirmClusterCreation(config); err != nil {
        return err
    }
    
    return s.provider.Create(config)
}
```

### 3. Command Wrapper Pattern

Standardize command execution with common setup:

```go
func WrapCommandWithCommonSetup(fn CommandFunc) cobra.RunE {
    return func(cmd *cobra.Command, args []string) error {
        ui.ShowLogo()
        
        if err := ValidatePrerequisites(); err != nil {
            return err
        }
        
        return fn(cmd, args)
    }
}
```

### 4. Configuration Builder Pattern

Build complex configurations step by step:

```go
type ClusterConfigBuilder struct {
    config ClusterConfig
}

func NewClusterConfigBuilder() *ClusterConfigBuilder {
    return &ClusterConfigBuilder{
        config: ClusterConfig{
            Type:       ClusterTypeK3d,
            NodeCount:  3,
            K8sVersion: "latest",
        },
    }
}

func (b *ClusterConfigBuilder) WithName(name string) *ClusterConfigBuilder {
    b.config.Name = name
    return b
}

func (b *ClusterConfigBuilder) Build() ClusterConfig {
    return b.config
}
```

## Error Handling Strategy

### Error Types and Hierarchy

```go
// Base error types
type CLIError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType string

const (
    ErrorTypeValidation    ErrorType = "validation"
    ErrorTypePrerequisite  ErrorType = "prerequisite" 
    ErrorTypeExecution     ErrorType = "execution"
    ErrorTypeConfiguration ErrorType = "configuration"
)

// Specialized errors
type PrerequisiteError struct {
    Tool     string
    Required string
    Found    string
}

type ValidationError struct {
    Field   string
    Value   interface{}
    Reason  string
}
```

### Error Handling Flow

```mermaid
graph TD
    A[Command Execution] --> B{Error Occurred?}
    B -->|Yes| C[Determine Error Type]
    B -->|No| D[Success Response]
    
    C --> E{Validation Error?}
    C --> F{Prerequisite Error?}
    C --> G{Execution Error?}
    
    E -->|Yes| H[Show Validation Help]
    F -->|Yes| I[Show Installation Instructions]
    G -->|Yes| J[Show Debug Information]
    
    H --> K[Format User-Friendly Message]
    I --> K
    J --> K
    
    K --> L[Display Error & Exit]
```

## Testing Strategy

### Test Architecture

| Test Type | Location | Purpose | Tools |
|-----------|----------|---------|-------|
| **Unit Tests** | `*_test.go` files | Test individual functions and methods | Go testing, testify |
| **Integration Tests** | `tests/integration/` | Test command workflows end-to-end | Go testing, Docker |
| **Provider Tests** | `internal/*/providers/*_test.go` | Test external tool integrations | Go testing, mocks |
| **UI Tests** | `internal/*/ui/*_test.go` | Test interactive prompts | Go testing, survey mocks |

### Mock Strategy

```go
// Provider mocks for testing
type MockClusterProvider struct {
    mock.Mock
}

func (m *MockClusterProvider) Create(config ClusterConfig) error {
    args := m.Called(config)
    return args.Error(0)
}

func TestClusterService_CreateCluster(t *testing.T) {
    mockProvider := new(MockClusterProvider)
    mockProvider.On("Create", mock.Anything).Return(nil)
    
    service := NewClusterService(mockProvider)
    err := service.CreateCluster(validConfig)
    
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

## Performance Considerations

### Command Execution Optimization

1. **Lazy Loading**: Providers are initialized only when needed
2. **Parallel Execution**: Independent operations run concurrently
3. **Caching**: Command outputs cached to avoid repeated executions
4. **Resource Cleanup**: Automatic cleanup of temporary resources

### Memory Management

- **Streaming**: Large outputs processed in streams
- **Buffer Limits**: Configurable limits on command output buffers
- **Garbage Collection**: Explicit cleanup of large objects

## Security Considerations

### Input Validation

```go
func ValidateClusterName(name string) error {
    if matched, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]$`, name); !matched {
        return errors.New("cluster name must contain only alphanumeric characters and hyphens")
    }
    
    if len(name) > 63 {
        return errors.New("cluster name cannot exceed 63 characters")
    }
    
    return nil
}
```

### Command Injection Prevention

```go
func (e *Executor) Execute(command string, args ...string) error {
    // Use exec.Command to prevent shell injection
    cmd := exec.Command(command, args...)
    cmd.Env = e.buildSecureEnv()
    
    return cmd.Run()
}
```

## Extension Points

### Adding New Commands

1. **Create command package**: `cmd/newcommand/`
2. **Implement service**: `internal/newcommand/service.go`
3. **Register command**: Add to `cmd/root.go`
4. **Add tests**: Unit and integration tests

### Adding New Providers

1. **Define interface**: Extend or create provider interface
2. **Implement provider**: `internal/*/providers/newprovider/`
3. **Register provider**: Add to provider factory
4. **Add configuration**: Extend models for provider config

### Adding New Deployment Modes

1. **Extend models**: Add to deployment mode enum
2. **Update validation**: Add validation rules
3. **Extend chart service**: Handle new mode in installation
4. **Update documentation**: Add mode-specific documentation

## Monitoring and Observability

### Logging Strategy

```go
// Structured logging with levels
log.WithFields(logrus.Fields{
    "cluster": clusterName,
    "action":  "create",
    "duration": duration,
}).Info("Cluster creation completed")

// Error logging with context
log.WithError(err).WithFields(logrus.Fields{
    "cluster": clusterName,
    "step":    "k3d-create",
}).Error("Failed to create cluster")
```

### Metrics Collection

- **Command execution time**: Track performance of operations
- **Success/failure rates**: Monitor command reliability  
- **Resource usage**: Track memory and CPU usage during operations
- **Error categorization**: Classify and count error types

---

## Future Architecture Considerations

### Planned Enhancements

1. **Plugin System**: Support for external command plugins
2. **Remote Providers**: Cloud provider integrations (EKS, GKE, AKS)
3. **Configuration Management**: Centralized config with profiles
4. **API Mode**: REST API for programmatic access

### Scalability Improvements

1. **Async Operations**: Background cluster operations
2. **Batch Processing**: Multiple cluster operations
3. **Resource Pooling**: Shared resources across operations
4. **Caching Layer**: Persistent caching for expensive operations

This architecture provides a solid foundation for the OpenFrame CLI while maintaining flexibility for future enhancements and integrations.