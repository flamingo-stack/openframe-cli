# Architecture Overview

OpenFrame CLI implements a clean architecture pattern with domain-driven design, providing a maintainable and testable codebase for Kubernetes cluster management and development workflows.

## High-Level Architecture

### System Overview
```mermaid
graph TB
    subgraph "User Interface Layer"
        CLI[CLI Commands]
        Wizard[Interactive Wizards]
        Progress[Progress Indicators]
    end
    
    subgraph "Application Layer"
        Bootstrap[Bootstrap Orchestrator]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Development Service]
    end
    
    subgraph "Infrastructure Layer"
        K3DProvider[K3D Provider]
        HelmProvider[Helm Provider]
        ArgoCDProvider[ArgoCD Provider]
        TelepresenceProvider[Telepresence Provider]
    end
    
    subgraph "External Systems"
        Docker[Docker Engine]
        K8s[Kubernetes API]
        GitRepos[Git Repositories]
    end
    
    CLI --> Bootstrap
    CLI --> ClusterSvc
    CLI --> ChartSvc  
    CLI --> DevSvc
    
    Bootstrap --> ClusterSvc
    Bootstrap --> ChartSvc
    
    ClusterSvc --> K3DProvider
    ChartSvc --> HelmProvider
    ChartSvc --> ArgoCDProvider
    DevSvc --> TelepresenceProvider
    
    K3DProvider --> Docker
    HelmProvider --> K8s
    ArgoCDProvider --> K8s
    TelepresenceProvider --> K8s
    
    HelmProvider --> GitRepos
    ArgoCDProvider --> GitRepos
    
    style CLI fill:#e1f5fe
    style Bootstrap fill:#fff3e0
    style K8s fill:#e8f5e8
```

## Core Design Principles

### 1. Clean Architecture Layers

| Layer | Purpose | Examples |
|-------|---------|----------|
| **Presentation** | User interaction, CLI commands, UI | `cmd/*`, Interactive wizards |
| **Application** | Business logic, orchestration | `internal/*/service.go` |
| **Infrastructure** | External integrations, providers | `internal/*/providers/*` |
| **Shared** | Cross-cutting concerns | `internal/shared/*` |

### 2. Dependency Direction
```mermaid
graph TD
    Presentation --> Application
    Application --> Infrastructure
    Application --> Shared
    Infrastructure --> Shared
    
    style Presentation fill:#e1f5fe
    style Application fill:#fff3e0
    style Infrastructure fill:#e8f5e8
    style Shared fill:#f3e5f5
```

Dependencies flow inward: outer layers depend on inner layers, never the reverse.

### 3. Interface-Driven Design

All external dependencies are abstracted behind interfaces, enabling:
- **Testability**: Mock implementations for unit tests
- **Flexibility**: Swappable implementations (e.g., different cluster providers)
- **Decoupling**: Services don't depend on concrete implementations

## Module Architecture

### Command Structure (cmd/)
```mermaid
graph TB
    Root[root.go] --> Bootstrap[bootstrap/]
    Root --> Cluster[cluster/]
    Root --> Chart[chart/]
    Root --> Dev[dev/]
    
    Cluster --> ClusterCreate[create.go]
    Cluster --> ClusterDelete[delete.go]
    Cluster --> ClusterList[list.go]
    Cluster --> ClusterStatus[status.go]
    Cluster --> ClusterCleanup[cleanup.go]
    
    Chart --> ChartInstall[install.go]
    
    Dev --> DevIntercept[intercept.go]
    Dev --> DevScaffold[scaffold.go]
    
    style Root fill:#e1f5fe
    style Bootstrap fill:#fff3e0
```

Each command module follows this pattern:
1. **Command Definition**: Cobra command with flags and validation
2. **Service Delegation**: Business logic delegated to service layer
3. **Error Handling**: Consistent error formatting and exit codes
4. **Help Integration**: Rich help text and examples

### Service Layer (internal/*/services/)
```mermaid
graph TB
    subgraph "Bootstrap Service"
        BootstrapSvc[Bootstrap Service]
        BootstrapSvc --> ClusterSvc[Cluster Service]
        BootstrapSvc --> ChartSvc[Chart Service]
    end
    
    subgraph "Core Services"
        ClusterSvc --> ClusterProvider[Cluster Provider Interface]
        ChartSvc --> HelmProvider[Helm Provider Interface]
        ChartSvc --> ArgoCDProvider[ArgoCD Provider Interface]
        DevSvc[Dev Service] --> DevProvider[Dev Provider Interface]
    end
    
    subgraph "Shared Services"
        Config[Configuration Service]
        Executor[Command Executor]
        UI[UI Service]
    end
    
    ClusterSvc --> Config
    ChartSvc --> Config
    DevSvc --> Config
    
    ClusterProvider --> Executor
    HelmProvider --> Executor
    ArgoCDProvider --> Executor
    
    ClusterSvc --> UI
    ChartSvc --> UI
    DevSvc --> UI
```

Service characteristics:
- **Single Responsibility**: Each service handles one domain
- **Interface Dependencies**: Services depend on interfaces, not implementations
- **Error Handling**: Domain-specific error types and handling
- **Configuration**: Injected configuration dependencies

### Provider Layer (internal/*/providers/)

Providers implement infrastructure concerns:

#### K3D Provider (Cluster Management)
```go
type ClusterProvider interface {
    CreateCluster(config ClusterConfig) (*Cluster, error)
    DeleteCluster(name string) error
    ListClusters() ([]*Cluster, error)
    GetClusterStatus(name string) (*ClusterStatus, error)
}
```

#### Helm Provider (Package Management)
```go
type HelmProvider interface {
    InstallChart(chart ChartConfig) error
    UpgradeChart(chart ChartConfig) error
    UninstallChart(name, namespace string) error
    ListReleases() ([]*Release, error)
}
```

#### ArgoCD Provider (GitOps)
```go
type ArgoCDProvider interface {
    InstallArgoCD(config ArgoCDConfig) error
    CreateApplication(app ApplicationConfig) error
    SyncApplication(name, namespace string) error
    WaitForApplications(timeout time.Duration) error
}
```

## Data Flow Patterns

### Bootstrap Flow (Complete Environment Setup)
```mermaid
sequenceDiagram
    participant User
    participant BootstrapCmd
    participant BootstrapSvc
    participant ClusterSvc
    participant ChartSvc
    participant K3DProvider
    participant HelmProvider
    participant ArgoCDProvider
    
    User->>BootstrapCmd: openframe bootstrap
    BootstrapCmd->>BootstrapSvc: ExecuteBootstrap()
    
    BootstrapSvc->>ClusterSvc: CreateCluster()
    ClusterSvc->>K3DProvider: CreateCluster(config)
    K3DProvider-->>ClusterSvc: cluster created
    ClusterSvc-->>BootstrapSvc: cluster ready
    
    BootstrapSvc->>ChartSvc: InstallCharts()
    ChartSvc->>HelmProvider: InstallChart(argocd)
    HelmProvider-->>ChartSvc: argocd installed
    ChartSvc->>ArgoCDProvider: CreateApplication(app-of-apps)
    ArgoCDProvider-->>ChartSvc: applications created
    ChartSvc->>ArgoCDProvider: WaitForSync()
    ArgoCDProvider-->>ChartSvc: all synced
    ChartSvc-->>BootstrapSvc: charts ready
    
    BootstrapSvc-->>BootstrapCmd: complete
    BootstrapCmd-->>User: success
```

### Error Handling Flow
```mermaid
graph TD
    Operation[Operation] --> Success{Success?}
    Success -->|Yes| Return[Return Result]
    Success -->|No| ErrorType{Error Type?}
    
    ErrorType --> ValidationError[Validation Error]
    ErrorType --> InfrastructureError[Infrastructure Error]
    ErrorType --> TimeoutError[Timeout Error]
    
    ValidationError --> UserFriendlyMsg[User-Friendly Message]
    InfrastructureError --> RetryLogic[Retry Logic]
    TimeoutError --> CleanupLogic[Cleanup Logic]
    
    RetryLogic --> Success
    CleanupLogic --> UserFriendlyMsg
    UserFriendlyMsg --> ExitCode[Exit Code]
    
    style Operation fill:#e1f5fe
    style Return fill:#e8f5e8
    style ExitCode fill:#ffebee
```

## Key Design Patterns

### 1. Command Pattern (CLI Commands)
```go
// Each command implements this pattern
func GetClusterCreateCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "create [cluster-name]",
        Short: "Create a new K3D cluster",
        RunE: func(cmd *cobra.Command, args []string) error {
            // Parse flags and arguments
            config := parseCreateFlags(cmd)
            
            // Delegate to service
            service := cluster.NewService(/* dependencies */)
            return service.CreateCluster(config)
        },
    }
}
```

### 2. Dependency Injection
```go
// Services receive dependencies via constructor
func NewClusterService(
    provider ClusterProvider,
    ui UIService,
    config ConfigService,
) *ClusterService {
    return &ClusterService{
        provider: provider,
        ui: ui,
        config: config,
    }
}
```

### 3. Builder Pattern (Configuration)
```go
// Fluent configuration building
config := cluster.NewConfigBuilder().
    WithName("my-cluster").
    WithMemory("4Gi").
    WithPorts(8080, 8443).
    Build()
```

### 4. Strategy Pattern (Deployment Modes)
```go
type DeploymentStrategy interface {
    GetHelmValues() map[string]interface{}
    GetApplications() []*ApplicationConfig
}

// Different strategies for different deployment modes
type OSSStrategy struct{}
type SaaSSharedStrategy struct{}
type SaaSDedicatedStrategy struct{}
```

## Component Responsibilities

### Core Components

| Component | Responsibilities |
|-----------|------------------|
| **Bootstrap Service** | Orchestrates complete environment setup, coordinates cluster and chart services |
| **Cluster Service** | Manages cluster lifecycle, handles cluster-specific configurations |
| **Chart Service** | Manages Helm charts and ArgoCD applications, implements GitOps patterns |
| **Development Service** | Provides developer tools, traffic interception, service scaffolding |

### Infrastructure Providers

| Provider | Responsibilities |
|----------|------------------|
| **K3D Provider** | K3D cluster management, Docker integration, networking configuration |
| **Helm Provider** | Helm chart installation, release management, repository handling |
| **ArgoCD Provider** | ArgoCD installation, application management, sync monitoring |
| **Telepresence Provider** | Traffic interception, local development workflows |

### Shared Components

| Component | Responsibilities |
|-----------|------------------|
| **Command Executor** | External command execution, output capture, error handling |
| **Configuration Service** | Configuration loading, validation, default values |
| **UI Service** | Interactive prompts, progress indicators, formatted output |
| **Error Handler** | Error categorization, user-friendly messages, retry logic |

## Testing Architecture

### Test Structure
```mermaid
graph TB
    subgraph "Unit Tests"
        ServiceTests[Service Layer Tests]
        ProviderTests[Provider Tests with Mocks]
        UtilTests[Utility Function Tests]
    end
    
    subgraph "Integration Tests"
        CLITests[CLI Command Tests]
        E2ETests[End-to-End Workflow Tests]
        ProviderIntegration[Real Provider Integration]
    end
    
    subgraph "Test Utilities"
        Mocks[Mock Implementations]
        TestHelpers[Test Helper Functions]
        Fixtures[Test Data Fixtures]
    end
    
    ServiceTests --> Mocks
    CLITests --> TestHelpers
    E2ETests --> Fixtures
    
    style ServiceTests fill:#e1f5fe
    style E2ETests fill:#e8f5e8
```

### Testing Patterns

#### 1. Service Layer Testing
```go
func TestClusterService_CreateCluster(t *testing.T) {
    // Arrange
    mockProvider := &mocks.ClusterProvider{}
    mockUI := &mocks.UIService{}
    service := cluster.NewService(mockProvider, mockUI, nil)
    
    // Configure mocks
    mockProvider.On("CreateCluster", mock.Anything).Return(cluster, nil)
    
    // Act
    err := service.CreateCluster(config)
    
    // Assert
    assert.NoError(t, err)
    mockProvider.AssertExpectations(t)
}
```

#### 2. CLI Integration Testing
```go
func TestBootstrapCommand(t *testing.T) {
    // Use real CLI with test cluster
    runner := testutil.NewCLIRunner()
    result := runner.Run("bootstrap", "test-cluster", "--non-interactive")
    
    assert.Equal(t, 0, result.ExitCode)
    assert.Contains(t, result.Output, "Bootstrap Complete")
}
```

## Configuration Management

### Configuration Hierarchy
```mermaid
graph TB
    CLI[CLI Flags] --> ENV[Environment Variables]
    ENV --> Config[Config File]
    Config --> Defaults[Default Values]
    
    Defaults --> Final[Final Configuration]
    Config --> Final
    ENV --> Final
    CLI --> Final
    
    style CLI fill:#e1f5fe
    style Final fill:#e8f5e8
```

### Configuration Sources (Priority Order)
1. **CLI Flags**: Highest priority, explicit user input
2. **Environment Variables**: System-level configuration
3. **Configuration File**: Persistent user preferences
4. **Default Values**: Fallback values for all settings

## Performance Considerations

### Optimization Strategies

| Area | Strategy | Implementation |
|------|----------|----------------|
| **Command Startup** | Lazy loading | Load providers only when needed |
| **Network Operations** | Connection pooling | Reuse HTTP clients and connections |
| **File Operations** | Caching | Cache configuration and templates |
| **Docker Operations** | Batch operations | Combine multiple Docker commands |

### Resource Management
- **Memory**: Efficient streaming for large operations
- **CPU**: Parallel operations where possible
- **Disk**: Cleanup temporary files automatically
- **Network**: Timeout and retry configurations

## Security Considerations

### Security Patterns

| Pattern | Implementation | Purpose |
|---------|----------------|---------|
| **Input Validation** | Strict validation of all user inputs | Prevent injection attacks |
| **Credential Management** | Secure storage and rotation | Protect sensitive data |
| **Network Security** | TLS everywhere, certificate validation | Secure communications |
| **Privilege Separation** | Minimal permissions principle | Reduce attack surface |

---

## Next Steps

Understanding the architecture? Continue with:

- **[Testing Overview](../testing/overview.md)** - Learn the testing strategy and patterns
- **[Contributing Guidelines](../contributing/guidelines.md)** - Understand contribution processes
- **[Local Development](../setup/local-development.md)** - Set up your development environment

This architecture provides a solid foundation for maintainable, testable, and extensible CLI development while following industry best practices.