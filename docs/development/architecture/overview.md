# Architecture Overview

OpenFrame CLI is designed with a clean, modular architecture that separates concerns and makes the codebase maintainable, testable, and extensible. This document provides a comprehensive overview of the system design, component relationships, and key architectural decisions.

## System Architecture

### High-Level Architecture

```mermaid
flowchart TB
    subgraph "User Interface Layer"
        CLI[CLI Commands]
        Interactive[Interactive Prompts]
        Flags[Command Flags]
    end
    
    subgraph "Application Layer"
        Bootstrap[Bootstrap Service]
        Cluster[Cluster Service]
        Chart[Chart Service]
        Dev[Dev Service]
    end
    
    subgraph "Infrastructure Layer"
        K3dProvider[K3d Provider]
        HelmProvider[Helm Provider]
        DockerProvider[Docker Provider]
        KubectlProvider[kubectl Provider]
    end
    
    subgraph "External Systems"
        K3d[K3d Cluster]
        Helm[Helm Charts]
        Docker[Docker Engine]
        ArgoCD[ArgoCD GitOps]
        Telepresence[Telepresence]
        Skaffold[Skaffold]
    end
    
    CLI --> Bootstrap
    CLI --> Cluster
    CLI --> Chart
    CLI --> Dev
    
    Bootstrap --> Cluster
    Bootstrap --> Chart
    
    Cluster --> K3dProvider
    Chart --> HelmProvider
    Dev --> KubectlProvider
    
    K3dProvider --> K3d
    HelmProvider --> Helm
    DockerProvider --> Docker
    
    K3d --> Docker
    Helm --> ArgoCD
```

### Layered Architecture

OpenFrame CLI follows a **3-layer architecture** with clear separation of responsibilities:

| Layer | Purpose | Components | Dependencies |
|-------|---------|------------|--------------|
| **Presentation** | User interaction, command parsing | `cmd/` packages | Application layer only |
| **Application** | Business logic, orchestration | `internal/` services | Infrastructure layer only |
| **Infrastructure** | External tool integration | Provider interfaces | External systems only |

## Core Components

### Command Layer (`cmd/`)

The command layer contains Cobra command definitions that handle user interaction, argument parsing, and flag management.

```mermaid
flowchart LR
    subgraph "Command Structure"
        Root[Root Command]
        Root --> Bootstrap[bootstrap]
        Root --> Cluster[cluster]
        Root --> Chart[chart]
        Root --> Dev[dev]
        
        Cluster --> Create[create]
        Cluster --> Delete[delete]
        Cluster --> List[list]
        Cluster --> Status[status]
        Cluster --> Cleanup[cleanup]
        
        Chart --> Install[install]
        
        Dev --> Intercept[intercept]
        Dev --> SkaffoldCmd[skaffold]
    end
```

**Key Responsibilities:**
- Parse command-line arguments and flags
- Validate user input
- Display help and usage information
- Delegate business logic to service layer
- Handle command aliases and shortcuts

**Example Structure:**
```go
// cmd/cluster/cluster.go
func GetClusterCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:     "cluster",
        Aliases: []string{"k"},
        Short:   "Manage Kubernetes clusters",
        PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
            // Prerequisites check
            return cluster.CheckPrerequisites()
        },
    }
    
    // Add subcommands
    cmd.AddCommand(getCreateCmd())
    cmd.AddCommand(getDeleteCmd())
    cmd.AddCommand(getListCmd())
    cmd.AddCommand(getStatusCmd())
    
    return cmd
}
```

### Service Layer (`internal/`)

The service layer contains business logic and orchestrates complex workflows by coordinating multiple operations.

```mermaid
flowchart TD
    subgraph "Service Layer"
        BootstrapSvc[Bootstrap Service]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Dev Service]
    end
    
    subgraph "Shared Components"
        UI[UI Components]
        Models[Data Models]
        Errors[Error Handling]
        Utils[Utilities]
    end
    
    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc
    BootstrapSvc --> UI
    
    ClusterSvc --> UI
    ClusterSvc --> Models
    ClusterSvc --> Errors
    
    ChartSvc --> UI
    ChartSvc --> Models
    
    DevSvc --> Utils
```

**Bootstrap Service** - Orchestrates complete environment setup:
```go
type BootstrapService struct {
    clusterService *cluster.Service
    chartService   *chart.Service
    ui            *ui.Service
}

func (s *BootstrapService) Execute(cmd *cobra.Command, args []string) error {
    // 1. Show logo and check prerequisites
    s.ui.ShowLogo()
    if err := s.checkPrerequisites(); err != nil {
        return err
    }
    
    // 2. Create cluster
    if err := s.clusterService.Create(clusterName); err != nil {
        return err
    }
    
    // 3. Install charts
    if err := s.chartService.Install(deploymentMode); err != nil {
        return err
    }
    
    // 4. Show success message
    s.ui.ShowSuccess("Bootstrap complete!")
    return nil
}
```

**Cluster Service** - Manages K3d cluster lifecycle:
```go
type Service struct {
    k3dProvider   providers.K3dProvider
    ui           *ui.Service
    validator    *validators.ClusterValidator
}

func (s *Service) Create(name string) error {
    // Validate cluster name
    if err := s.validator.ValidateName(name); err != nil {
        return err
    }
    
    // Create cluster with K3d
    return s.k3dProvider.CreateCluster(name, s.getClusterConfig())
}
```

### Infrastructure Layer

The infrastructure layer provides abstractions for external tool integration through provider interfaces.

```mermaid
flowchart LR
    subgraph "Provider Interfaces"
        K3dInterface[K3d Provider Interface]
        HelmInterface[Helm Provider Interface]
        DockerInterface[Docker Provider Interface]
    end
    
    subgraph "Concrete Implementations"
        K3dImpl[K3d CLI Implementation]
        HelmImpl[Helm CLI Implementation]
        DockerImpl[Docker API Implementation]
    end
    
    subgraph "External Tools"
        K3dTool[k3d binary]
        HelmTool[helm binary]
        DockerEngine[Docker Engine]
    end
    
    K3dInterface --> K3dImpl
    HelmInterface --> HelmImpl
    DockerInterface --> DockerImpl
    
    K3dImpl --> K3dTool
    HelmImpl --> HelmTool
    DockerImpl --> DockerEngine
```

**Provider Interface Example:**
```go
type K3dProvider interface {
    CreateCluster(name string, config ClusterConfig) error
    DeleteCluster(name string) error
    ListClusters() ([]Cluster, error)
    GetClusterStatus(name string) (*ClusterStatus, error)
}

type k3dCLIProvider struct {
    execRunner exec.Runner
}

func (p *k3dCLIProvider) CreateCluster(name string, config ClusterConfig) error {
    args := []string{"cluster", "create", name}
    args = append(args, config.ToK3dArgs()...)
    
    return p.execRunner.Run("k3d", args...)
}
```

## Data Flow Architecture

### Bootstrap Command Flow

The most complex operation is the bootstrap command, which demonstrates the full data flow:

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Command
    participant Bootstrap as Bootstrap Service
    participant Cluster as Cluster Service
    participant Chart as Chart Service
    participant K3d as K3d Provider
    participant Helm as Helm Provider
    participant UI as UI Service
    
    User->>CLI: openframe bootstrap my-cluster
    CLI->>Bootstrap: Execute(cmd, args)
    Bootstrap->>UI: ShowLogo()
    Bootstrap->>Bootstrap: checkPrerequisites()
    
    Bootstrap->>Cluster: Create("my-cluster")
    Cluster->>K3d: CreateCluster("my-cluster", config)
    K3d-->>Cluster: Success
    Cluster->>UI: ShowProgress("Cluster created")
    Cluster-->>Bootstrap: Success
    
    Bootstrap->>Chart: Install(deploymentMode)
    Chart->>Helm: InstallArgoCD()
    Helm-->>Chart: Success
    Chart->>Helm: InstallOpenFrameCharts()
    Helm-->>Chart: Success
    Chart->>UI: ShowProgress("Charts installed")
    Chart-->>Bootstrap: Success
    
    Bootstrap->>UI: ShowSuccess("Bootstrap complete")
    Bootstrap-->>CLI: Success
    CLI-->>User: Exit 0
```

### Error Handling Flow

```mermaid
flowchart TD
    Error[Error Occurs] --> Catch[Service Catches Error]
    Catch --> Wrap[Wrap with Context]
    Wrap --> Log[Log Error Details]
    Log --> UI[Display User Message]
    UI --> Cleanup[Cleanup Resources]
    Cleanup --> Exit[Exit with Error Code]
    
    Wrap --> Check{Error Type?}
    Check -->|Prerequisite| PrereqMsg[Prerequisites missing]
    Check -->|Validation| ValidationMsg[Invalid input]
    Check -->|External Tool| ToolMsg[Tool execution failed]
    Check -->|Network| NetworkMsg[Network connectivity]
    
    PrereqMsg --> UI
    ValidationMsg --> UI
    ToolMsg --> UI
    NetworkMsg --> UI
```

## Design Patterns

### 1. Command Pattern

Each CLI command is implemented using the Command pattern with Cobra:

```go
// Command interface (provided by Cobra)
type Command struct {
    Use   string
    RunE  func(cmd *Command, args []string) error
    // ... other fields
}

// Concrete command implementation
func getBootstrapCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "bootstrap [cluster-name]",
        RunE:  func(cmd *cobra.Command, args []string) error {
            return bootstrap.NewService().Execute(cmd, args)
        },
    }
}
```

### 2. Service Layer Pattern

Business logic is encapsulated in service classes:

```go
type Service struct {
    dependencies Dependencies
}

func NewService(deps Dependencies) *Service {
    return &Service{dependencies: deps}
}

func (s *Service) Execute(input Input) (Output, error) {
    // Business logic implementation
}
```

### 3. Provider Pattern

External tool integration uses the Provider pattern for dependency injection:

```go
type Provider interface {
    Operation(params Params) error
}

type Service struct {
    provider Provider
}

// Dependency injection
func NewService(provider Provider) *Service {
    return &Service{provider: provider}
}
```

### 4. Builder Pattern

Complex configurations use the Builder pattern:

```go
type ClusterConfigBuilder struct {
    config ClusterConfig
}

func NewClusterConfigBuilder() *ClusterConfigBuilder {
    return &ClusterConfigBuilder{
        config: ClusterConfig{},
    }
}

func (b *ClusterConfigBuilder) WithName(name string) *ClusterConfigBuilder {
    b.config.Name = name
    return b
}

func (b *ClusterConfigBuilder) WithPorts(ports []int) *ClusterConfigBuilder {
    b.config.Ports = ports
    return b
}

func (b *ClusterConfigBuilder) Build() ClusterConfig {
    return b.config
}
```

## Key Design Decisions

### 1. Why Cobra for CLI Framework?

**Decision**: Use Cobra for command-line interface
**Reasoning**:
- Industry standard for Go CLI applications
- Excellent support for subcommands, flags, and help generation
- Built-in shell completion support
- Used by kubectl, helm, and other Kubernetes tools

### 2. Why Service Layer Architecture?

**Decision**: Separate business logic into service layer
**Reasoning**:
- Testability: Services can be unit tested independently
- Reusability: Services can be used by different commands
- Maintainability: Clear separation of concerns
- Extensibility: Easy to add new features or modify existing ones

### 3. Why Provider Interfaces?

**Decision**: Abstract external tool integration behind interfaces
**Reasoning**:
- Testability: Can mock external dependencies
- Flexibility: Can swap implementations (CLI vs API)
- Reliability: Isolates external tool changes
- Development: Can use fake implementations for testing

### 4. Why Internal Package Structure?

**Decision**: Use `internal/` package for private code
**Reasoning**:
- Encapsulation: Prevents external packages from importing internal code
- API stability: Only public interfaces are exposed
- Refactoring: Internal changes don't break external users
- Go convention: Standard practice in Go projects

## Component Dependencies

### Dependency Graph

```mermaid
flowchart TD
    subgraph "cmd layer"
        CmdBootstrap[cmd/bootstrap]
        CmdCluster[cmd/cluster]
        CmdChart[cmd/chart]
        CmdDev[cmd/dev]
    end
    
    subgraph "service layer"
        SvcBootstrap[internal/bootstrap]
        SvcCluster[internal/cluster]
        SvcChart[internal/chart]
        SvcDev[internal/dev]
    end
    
    subgraph "shared components"
        SharedUI[internal/shared/ui]
        SharedModels[internal/shared/models]
        SharedErrors[internal/shared/errors]
    end
    
    subgraph "external"
        Cobra[github.com/spf13/cobra]
        Docker[Docker API/CLI]
        K3d[k3d CLI]
        Helm[helm CLI]
    end
    
    CmdBootstrap --> SvcBootstrap
    CmdCluster --> SvcCluster
    CmdChart --> SvcChart
    CmdDev --> SvcDev
    
    SvcBootstrap --> SvcCluster
    SvcBootstrap --> SvcChart
    
    SvcCluster --> SharedUI
    SvcChart --> SharedUI
    SvcDev --> SharedUI
    
    SvcCluster --> SharedModels
    SvcChart --> SharedModels
    
    SvcCluster --> K3d
    SvcChart --> Helm
    
    CmdBootstrap --> Cobra
    CmdCluster --> Cobra
    CmdChart --> Cobra
    CmdDev --> Cobra
```

### Module Dependencies

```go
// go.mod dependencies
require (
    github.com/spf13/cobra v1.7.0        // CLI framework
    github.com/spf13/viper v1.16.0       // Configuration management
    gopkg.in/yaml.v3 v3.0.1              // YAML processing
    github.com/stretchr/testify v1.8.4   // Testing framework
)
```

## Error Handling Architecture

### Error Types

```go
type ErrorType string

const (
    ErrorTypePrerequisite ErrorType = "prerequisite"
    ErrorTypeValidation  ErrorType = "validation"
    ErrorTypeExecution   ErrorType = "execution"
    ErrorTypeNetwork     ErrorType = "network"
)

type OpenFrameError struct {
    Type    ErrorType
    Message string
    Cause   error
    Context map[string]interface{}
}
```

### Error Propagation

```mermaid
flowchart TD
    ExternalError[External Tool Error] --> WrapError[Wrap with Context]
    WrapError --> ServiceError[Service Error]
    ServiceError --> CommandError[Command Error]
    CommandError --> UserError[User-Friendly Message]
    
    WrapError --> LogError[Log Technical Details]
    UserError --> ExitCode[Exit with Code]
```

## Testing Architecture

### Test Organization

```
tests/
├── unit/              # Unit tests for individual components
│   ├── cluster/       # Cluster service tests
│   ├── chart/         # Chart service tests
│   └── bootstrap/     # Bootstrap service tests
├── integration/       # Integration tests with external tools
│   ├── k3d/           # K3d integration tests
│   ├── helm/          # Helm integration tests
│   └── e2e/           # End-to-end workflow tests
├── mocks/             # Generated mocks for testing
└── fixtures/          # Test data and fixtures
```

### Test Strategy

```mermaid
flowchart LR
    subgraph "Test Pyramid"
        Unit[Unit Tests<br/>Fast, Isolated]
        Integration[Integration Tests<br/>External Tools]
        E2E[End-to-End Tests<br/>Full Workflows]
    end
    
    Unit --> Integration
    Integration --> E2E
    
    Unit -.-> UnitMocks[Mocked Dependencies]
    Integration -.-> IntegrationReal[Real Tools]
    E2E -.-> E2EReal[Real Environment]
```

## Performance Considerations

### Concurrent Operations

```go
// Example: Parallel prerequisite checks
func (s *Service) checkPrerequisites() error {
    checks := []func() error{
        s.checkDocker,
        s.checkKubectl,
        s.checkHelm,
        s.checkK3d,
    }
    
    errCh := make(chan error, len(checks))
    
    for _, check := range checks {
        go func(fn func() error) {
            errCh <- fn()
        }(check)
    }
    
    for range checks {
        if err := <-errCh; err != nil {
            return err
        }
    }
    
    return nil
}
```

### Resource Management

```go
type ResourceManager struct {
    resources []Resource
    mu        sync.Mutex
}

func (rm *ResourceManager) Add(resource Resource) {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    rm.resources = append(rm.resources, resource)
}

func (rm *ResourceManager) Cleanup() error {
    rm.mu.Lock()
    defer rm.mu.Unlock()
    
    var errs []error
    for i := len(rm.resources) - 1; i >= 0; i-- {
        if err := rm.resources[i].Cleanup(); err != nil {
            errs = append(errs, err)
        }
    }
    
    return errors.Join(errs...)
}
```

## Future Architecture Considerations

### Planned Improvements

1. **Plugin Architecture**: Support for external plugins
2. **Configuration Management**: Better support for multiple environments
3. **Async Operations**: Background operations for long-running tasks
4. **Caching**: Cache external tool results for performance
5. **Observability**: Structured logging and metrics

### Extension Points

```go
// Future plugin interface
type Plugin interface {
    Name() string
    Execute(ctx context.Context, args []string) error
}

// Future configuration management
type ConfigManager interface {
    LoadConfig(path string) (*Config, error)
    SaveConfig(config *Config, path string) error
    MergeConfigs(configs ...*Config) *Config
}
```

---

This architecture provides a solid foundation for OpenFrame CLI that is maintainable, testable, and extensible while following Go best practices and proven architectural patterns.