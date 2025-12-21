# Architecture Overview

This document provides a comprehensive overview of OpenFrame CLI's architecture, including system design, component relationships, data flow, and key design decisions.

## High-Level Architecture

OpenFrame CLI follows a modular, layered architecture designed for maintainability, testability, and extensibility.

```mermaid
graph TB
    %% User Interface Layer
    CLI[CLI Interface] --> Commands[Command Layer]
    
    %% Command Layer
    Commands --> Bootstrap[Bootstrap Module]
    Commands --> Cluster[Cluster Module] 
    Commands --> Chart[Chart Module]
    Commands --> Dev[Dev Module]
    
    %% Service Layer
    Bootstrap --> ClusterSvc[Cluster Services]
    Bootstrap --> ChartSvc[Chart Services]
    
    Cluster --> ClusterSvc
    Chart --> ChartSvc
    Dev --> DevSvc[Dev Services]
    
    %% Infrastructure Layer
    ClusterSvc --> K3dProvider[K3d Provider]
    ClusterSvc --> KubeClient[Kubernetes Client]
    
    ChartSvc --> HelmClient[Helm Client]
    ChartSvc --> ArgoCD[ArgoCD API]
    
    DevSvc --> Telepresence[Telepresence]
    DevSvc --> Skaffold[Skaffold]
    
    %% External Dependencies
    K3dProvider --> Docker[Docker Engine]
    KubeClient --> K8sCluster[Kubernetes Cluster]
    HelmClient --> HelmRepo[Helm Repositories]
    ArgoCD --> GitRepo[Git Repositories]
    
    %% Shared Components
    Commands --> SharedUI[Shared UI Components]
    Commands --> SharedErrors[Shared Error Handling]
    Commands --> SharedConfig[Shared Configuration]
    
    style CLI fill:#e1f5fe
    style Commands fill:#f3e5f5
    style ClusterSvc fill:#e8f5e8
    style ChartSvc fill:#fff3e0
    style DevSvc fill:#fce4ec
```

## Core Components

### Command Layer (`cmd/`)

The command layer provides the CLI interface using the Cobra framework.

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **Bootstrap Command** | `cmd/bootstrap` | Orchestrates complete environment setup |
| **Cluster Commands** | `cmd/cluster` | Manages cluster lifecycle operations |
| **Chart Commands** | `cmd/chart` | Handles ArgoCD and chart management |
| **Dev Commands** | `cmd/dev` | Provides development workflow tools |

#### Command Structure

```text
cmd/
├── bootstrap/
│   └── bootstrap.go        # Complete environment setup
├── cluster/
│   ├── cluster.go         # Root cluster command
│   ├── create.go          # Cluster creation
│   ├── delete.go          # Cluster deletion
│   ├── list.go            # Cluster listing
│   ├── status.go          # Cluster status
│   └── cleanup.go         # Resource cleanup
├── chart/
│   ├── chart.go           # Root chart command
│   └── install.go         # ArgoCD installation
└── dev/
    └── dev.go             # Development tools
```

### Service Layer (`internal/`)

The service layer contains business logic and orchestrates operations between providers.

| Module | Responsibility | Key Components |
|--------|----------------|----------------|
| **Bootstrap Services** | Environment orchestration | Cluster + Chart coordination |
| **Cluster Services** | K3d cluster management | Creation, deletion, status |
| **Chart Services** | ArgoCD and Helm management | Installation, configuration |
| **Dev Services** | Development tools | Traffic interception, workflows |

#### Internal Package Organization

```text
internal/
├── bootstrap/
│   ├── services/          # Bootstrap orchestration logic
│   └── models/            # Bootstrap configuration
├── cluster/
│   ├── services/          # Cluster business logic
│   ├── models/            # Cluster data structures
│   ├── ui/                # Interactive prompts
│   ├── utils/             # Cluster utilities
│   └── prerequisites/     # Dependency validation
├── chart/
│   ├── services/          # Chart management logic
│   ├── models/            # Chart configuration
│   └── prerequisites/     # Chart dependencies
├── dev/
│   ├── services/          # Development tool logic
│   ├── models/            # Dev configuration
│   └── prerequisites/     # Dev tool dependencies
└── shared/
    ├── ui/                # Common UI components
    ├── errors/            # Error handling
    └── config/            # Configuration management
```

## Data Flow Architecture

### Bootstrap Command Flow

The bootstrap command orchestrates a complete environment setup:

```mermaid
sequenceDiagram
    participant User
    participant Bootstrap
    participant ClusterSvc as Cluster Service
    participant ChartSvc as Chart Service
    participant K3d
    participant ArgoCD
    participant UI
    
    User->>Bootstrap: bootstrap [cluster-name]
    Bootstrap->>UI: Show deployment mode selection
    UI->>User: Interactive prompts
    User-->>UI: Configuration choices
    UI-->>Bootstrap: Complete configuration
    
    Bootstrap->>ClusterSvc: CreateCluster(config)
    ClusterSvc->>K3d: Create K3d cluster
    K3d-->>ClusterSvc: Cluster ready
    ClusterSvc-->>Bootstrap: Cluster created
    
    Bootstrap->>ChartSvc: InstallCharts(clusterName)
    ChartSvc->>ArgoCD: Install ArgoCD helm chart
    ArgoCD-->>ChartSvc: ArgoCD installed
    ChartSvc->>ArgoCD: Deploy app-of-apps
    ArgoCD-->>ChartSvc: Apps configured
    ChartSvc-->>Bootstrap: Charts installed
    
    Bootstrap-->>User: Environment ready
```

### Cluster Management Flow

Individual cluster operations follow this pattern:

```mermaid
sequenceDiagram
    participant User
    participant ClusterCmd as Cluster Command
    participant UI as Cluster UI
    participant Service as Cluster Service
    participant Provider as K3d Provider
    participant Prereq as Prerequisites
    
    User->>ClusterCmd: cluster create [name]
    ClusterCmd->>Prereq: ValidatePrerequisites()
    Prereq-->>ClusterCmd: Prerequisites OK
    
    ClusterCmd->>UI: GetClusterConfig(name)
    UI->>User: Interactive configuration wizard
    User-->>UI: Configuration choices
    UI-->>ClusterCmd: ClusterConfig
    
    ClusterCmd->>Service: CreateCluster(config)
    Service->>Provider: CreateK3dCluster(config)
    Provider-->>Service: Cluster created
    Service->>Provider: WaitForReady()
    Provider-->>Service: Cluster ready
    Service-->>ClusterCmd: Success
    
    ClusterCmd-->>User: Cluster ready for use
```

## Component Interactions

### Dependency Management

```mermaid
graph LR
    %% Command Dependencies
    BootstrapCmd[Bootstrap Command] --> ClusterCmd[Cluster Command]
    BootstrapCmd --> ChartCmd[Chart Command]
    
    %% Service Dependencies
    BootstrapSvc[Bootstrap Service] --> ClusterSvc[Cluster Service]
    BootstrapSvc --> ChartSvc[Chart Service]
    
    %% UI Dependencies
    ClusterCmd --> ClusterUI[Cluster UI]
    ChartCmd --> ChartUI[Chart UI]
    ClusterUI --> SharedUI[Shared UI]
    ChartUI --> SharedUI
    
    %% Model Dependencies
    ClusterSvc --> ClusterModels[Cluster Models]
    ChartSvc --> ChartModels[Chart Models]
    ClusterUI --> ClusterModels
    
    %% Utility Dependencies
    ClusterSvc --> ClusterUtils[Cluster Utils]
    ClusterUtils --> ClusterModels
    
    %% Shared Dependencies
    ClusterCmd --> SharedErrors[Shared Errors]
    ChartCmd --> SharedErrors
    ClusterSvc --> SharedConfig[Shared Config]
    ChartSvc --> SharedConfig
```

### Interface Boundaries

```mermaid
graph TB
    %% External Interfaces
    UserInterface[User Interface] --> CLICommands[CLI Commands]
    CLICommands --> Services[Service Layer]
    Services --> Providers[Provider Layer]
    Providers --> ExternalTools[External Tools]
    
    %% Internal Interfaces
    Services --> Models[Data Models]
    Services --> UI[UI Components]
    UI --> Models
    
    %% Configuration Flow
    CLICommands --> Config[Configuration]
    Config --> Services
    Config --> Models
    
    %% Error Flow
    Providers --> ErrorHandling[Error Handling]
    Services --> ErrorHandling
    CLICommands --> ErrorHandling
    ErrorHandling --> UserInterface
```

## Key Design Patterns

### 1. Command Pattern

Each CLI command implements a consistent interface:

```go
type Command interface {
    Execute(cmd *cobra.Command, args []string) error
}

// Example implementation
func getCreateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "create [NAME]",
        Short: "Create a new cluster",
        RunE:  runCreateCluster,  // Delegates to service layer
    }
    return cmd
}
```

### 2. Service Layer Pattern

Business logic is isolated in service layers:

```go
type ClusterService interface {
    CreateCluster(config ClusterConfig) error
    DeleteCluster(name string) error
    ListClusters() ([]Cluster, error)
    GetClusterStatus(name string) (*ClusterStatus, error)
}

type clusterService struct {
    provider  ProviderInterface
    validator ValidatorInterface
}
```

### 3. Provider Pattern

External tool interactions are abstracted through providers:

```go
type K3dProvider interface {
    CreateCluster(config K3dConfig) error
    DeleteCluster(name string) error
    ListClusters() ([]K3dCluster, error)
}

type HelmProvider interface {
    InstallChart(config HelmConfig) error
    UninstallChart(name, namespace string) error
}
```

### 4. Configuration Pattern

Configuration is managed through structured models:

```go
type ClusterConfig struct {
    Name         string
    Nodes        int
    K8sVersion   string
    Ports        []PortMapping
    Volumes      []VolumeMount
    Environment  map[string]string
}

func (c *ClusterConfig) Validate() error {
    // Validation logic
}
```

## Data Models

### Core Data Structures

| Model | Package | Purpose | Key Fields |
|-------|---------|---------|------------|
| **ClusterConfig** | `internal/cluster/models` | Cluster configuration | Name, Nodes, K8sVersion, Ports |
| **ChartConfig** | `internal/chart/models` | Chart installation config | ChartName, Namespace, Values |
| **BootstrapConfig** | `internal/bootstrap/models` | Complete setup config | ClusterConfig, ChartConfig, Mode |
| **DevConfig** | `internal/dev/models` | Development tool config | Service, Port, Protocol |

### Configuration Hierarchy

```mermaid
graph TB
    GlobalConfig[Global Configuration] --> BootstrapConfig[Bootstrap Configuration]
    GlobalConfig --> ClusterConfig[Cluster Configuration]
    GlobalConfig --> ChartConfig[Chart Configuration]
    GlobalConfig --> DevConfig[Dev Configuration]
    
    BootstrapConfig --> ClusterConfig
    BootstrapConfig --> ChartConfig
    
    ClusterConfig --> K3dConfig[K3d Configuration]
    ChartConfig --> HelmConfig[Helm Configuration]
    ChartConfig --> ArgoCDConfig[ArgoCD Configuration]
    
    DevConfig --> TelepresenceConfig[Telepresence Config]
    DevConfig --> SkaffoldConfig[Skaffold Config]
```

## Error Handling Strategy

### Error Types

```go
// Domain-specific error types
type ClusterError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType string

const (
    ErrorTypeValidation   ErrorType = "validation"
    ErrorTypePrerequisite ErrorType = "prerequisite"
    ErrorTypeExecution    ErrorType = "execution"
    ErrorTypeTimeout      ErrorType = "timeout"
)
```

### Error Flow

```mermaid
graph TB
    Operation[Operation] --> Validation{Validation}
    Validation -->|Valid| Prerequisites{Prerequisites}
    Validation -->|Invalid| ValidationError[Validation Error]
    
    Prerequisites -->|OK| Execution[Execution]
    Prerequisites -->|Missing| PrerequisiteError[Prerequisite Error]
    
    Execution --> Success[Success]
    Execution --> ExecutionError[Execution Error]
    
    ValidationError --> ErrorHandler[Error Handler]
    PrerequisiteError --> ErrorHandler
    ExecutionError --> ErrorHandler
    
    ErrorHandler --> UserFeedback[User Feedback]
    ErrorHandler --> Logging[Logging]
    ErrorHandler --> Recovery[Recovery Suggestions]
```

## Performance Considerations

### Concurrency Design

```mermaid
graph LR
    MainThread[Main Thread] --> CommandProcessing[Command Processing]
    CommandProcessing --> ServiceCalls[Service Calls]
    
    ServiceCalls --> ParallelOps[Parallel Operations]
    ParallelOps --> K3dOps[K3d Operations]
    ParallelOps --> HelmOps[Helm Operations]
    ParallelOps --> ValidationOps[Validation Operations]
    
    K3dOps --> SyncPoint[Synchronization Point]
    HelmOps --> SyncPoint
    ValidationOps --> SyncPoint
    
    SyncPoint --> Results[Aggregated Results]
    Results --> MainThread
```

### Resource Management

| Resource | Management Strategy | Implementation |
|----------|-------------------|----------------|
| **Docker Containers** | Lifecycle management | K3d provider cleanup |
| **Kubernetes Resources** | Namespace isolation | Resource labeling and cleanup |
| **File Descriptors** | Proper closure | Defer statements and context cancellation |
| **Network Connections** | Connection pooling | HTTP client reuse |

## Extension Points

### Adding New Commands

```text
1. Create command file in appropriate cmd/ directory
2. Implement command structure with Cobra
3. Create service layer in internal/ if needed
4. Add models for configuration
5. Implement provider interfaces if needed
6. Add tests and documentation
```

### Adding New Providers

```text
1. Define provider interface
2. Implement provider struct
3. Add to service layer dependency injection
4. Create configuration models
5. Add prerequisite validation
6. Implement error handling
7. Add comprehensive tests
```

### Adding New UI Components

```text
1. Create UI component in internal/shared/ui/
2. Define interface for reusability
3. Implement for different command contexts
4. Add configuration options
5. Test interactive flows
```

## Testing Architecture

### Test Organization

```text
tests/
├── unit/                  # Unit tests alongside source
├── integration/           # Integration tests
│   ├── cluster/          # Cluster integration tests
│   ├── chart/            # Chart integration tests
│   └── bootstrap/        # End-to-end tests
├── fixtures/             # Test data and fixtures
└── utils/                # Test utilities
```

### Testing Strategy

| Test Level | Scope | Tools | Coverage |
|------------|-------|-------|----------|
| **Unit** | Individual functions/methods | Go testing, Testify | >80% |
| **Integration** | Component interactions | Real K3d clusters | Critical paths |
| **End-to-End** | Complete workflows | Full bootstrap process | Happy path + error scenarios |
| **Performance** | Resource usage | Benchmarks | Resource limits |

## Security Considerations

### Security Boundaries

```mermaid
graph TB
    UserInput[User Input] --> Validation[Input Validation]
    Validation --> Sanitization[Input Sanitization]
    
    Sanitization --> ServiceLayer[Service Layer]
    ServiceLayer --> ProviderLayer[Provider Layer]
    
    ProviderLayer --> DockerAPI[Docker API]
    ProviderLayer --> KubernetesAPI[Kubernetes API]
    ProviderLayer --> HelmAPI[Helm API]
    
    DockerAPI --> ContainerRuntime[Container Runtime]
    KubernetesAPI --> ClusterAPI[Cluster API]
    HelmAPI --> ChartRepository[Chart Repository]
```

### Security Measures

| Area | Measures | Implementation |
|------|----------|----------------|
| **Input Validation** | Sanitize all user inputs | Validation functions in models |
| **Privilege Escalation** | Minimal required permissions | Docker group membership only |
| **Network Security** | Local cluster access only | K3d network isolation |
| **Credential Management** | No persistent credentials | Temporary cluster credentials |

## Performance Metrics

### Key Performance Indicators

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Cluster Creation Time** | < 60 seconds | Time from command start to cluster ready |
| **Bootstrap Time** | < 5 minutes | Complete environment setup |
| **Memory Usage** | < 100MB | CLI process memory footprint |
| **Binary Size** | < 50MB | Executable file size |

## Future Architecture Considerations

### Planned Enhancements

1. **Plugin System**: Allow third-party extensions
2. **Remote Cluster Support**: Manage clusters beyond K3d
3. **Configuration Profiles**: Preset configurations for different use cases
4. **Advanced Networking**: Custom network configurations
5. **Multi-Cluster Management**: Coordinate multiple clusters

### Scalability Considerations

- **Horizontal**: Support for multiple concurrent clusters
- **Vertical**: Efficient resource utilization for large clusters
- **Operational**: Simplified maintenance and updates

This architecture provides a solid foundation for OpenFrame CLI's current functionality while allowing for future growth and extension. The modular design ensures maintainability and testability, while the clear separation of concerns makes the codebase accessible to new contributors.