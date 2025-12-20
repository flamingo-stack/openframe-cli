# OpenFrame CLI - Architecture Overview

This document provides a comprehensive technical overview of the OpenFrame CLI architecture, focusing on system design, component relationships, data flow, and key design patterns used throughout the codebase.

## High-Level Architecture

OpenFrame CLI follows a layered architecture with clear separation of concerns, making it maintainable, testable, and extensible.

```mermaid
flowchart TB
    subgraph "User Interface Layer"
        CLI[Cobra CLI Commands]
        FLAGS[Flag Management]
        HELP[Help & Documentation]
    end
    
    subgraph "Application Layer" 
        BOOTSTRAP[Bootstrap Orchestrator]
        CLUSTER_SVC[Cluster Service]
        CHART_SVC[Chart Service] 
        DEV_SVC[Dev Service]
    end
    
    subgraph "Domain Layer"
        MODELS[Domain Models]
        INTERFACES[Service Interfaces]
        VALIDATION[Business Rules]
    end
    
    subgraph "Infrastructure Layer"
        K3D[K3d Provider]
        HELM[Helm Client]
        KUBECTL[Kubectl Client] 
        ARGOCD[ArgoCD Integration]
        DOCKER[Docker Integration]
    end
    
    subgraph "Cross-Cutting Concerns"
        UI_COMP[UI Components]
        ERROR[Error Handling]
        CONFIG[Configuration]
        PREREQ[Prerequisites]
    end
    
    CLI --> BOOTSTRAP
    CLI --> CLUSTER_SVC
    CLI --> CHART_SVC
    CLI --> DEV_SVC
    
    BOOTSTRAP --> CLUSTER_SVC
    BOOTSTRAP --> CHART_SVC
    
    CLUSTER_SVC --> K3D
    CHART_SVC --> HELM
    CHART_SVC --> ARGOCD
    DEV_SVC --> KUBECTL
    
    CLUSTER_SVC --> MODELS
    CHART_SVC --> MODELS
    DEV_SVC --> MODELS
    
    CLUSTER_SVC --> UI_COMP
    CHART_SVC --> UI_COMP
    DEV_SVC --> UI_COMP
    
    CLUSTER_SVC --> PREREQ
    CHART_SVC --> PREREQ
    DEV_SVC --> PREREQ
    
    style CLI fill:#e3f2fd
    style BOOTSTRAP fill:#f3e5f5
    style MODELS fill:#e8f5e8
    style K3D fill:#fff3e0
```

## Core Components

### Command Layer (`cmd/`)

The command layer implements the CLI interface using the Cobra framework. Each command group has its own package with clear responsibilities.

| Component | Package | Responsibility | Key Files |
|-----------|---------|---------------|-----------|
| **Bootstrap** | `cmd/bootstrap/` | Complete environment setup orchestration | `bootstrap.go` |
| **Cluster Management** | `cmd/cluster/` | K3d cluster lifecycle operations | `cluster.go`, `create.go`, `delete.go`, `list.go`, `status.go`, `cleanup.go` |
| **Chart Management** | `cmd/chart/` | Helm/ArgoCD installation and management | `chart.go`, `install.go` |
| **Development Tools** | `cmd/dev/` | Local development workflow tools | `dev.go` |

#### Command Structure Pattern

```go
// Standard command structure used across all commands
func GetCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "command [args]",
        Short: "Brief description", 
        Long:  "Detailed description with examples",
        Args:  cobra.MaximumNArgs(1),
        PreRunE: func(cmd *cobra.Command, args []string) error {
            // Validation and prerequisites
            return validatePrerequisites()
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            // Delegate to service layer
            return service.Execute(cmd, args)
        },
    }
    
    // Add command-specific flags
    addFlags(cmd)
    return cmd
}
```

### Service Layer (`internal/*/services/`)

The service layer contains the core business logic, orchestrating interactions between different components while maintaining separation of concerns.

```mermaid
flowchart LR
    subgraph "Service Layer Architecture"
        CS[Cluster Service]
        CHS[Chart Service] 
        BS[Bootstrap Service]
        DS[Dev Service]
        
        CS --> |uses| MODELS[Domain Models]
        CHS --> |uses| MODELS
        BS --> |orchestrates| CS
        BS --> |orchestrates| CHS
        
        CS --> |calls| PROVIDERS[Infrastructure Providers]
        CHS --> |calls| PROVIDERS
        DS --> |calls| PROVIDERS
        
        CS --> |displays| UI[UI Components]
        CHS --> |displays| UI
        DS --> |displays| UI
    end
    
    style BS fill:#f3e5f5
    style CS fill:#e8f5e8
    style CHS fill:#e8f5e8
    style DS fill:#e8f5e8
```

#### Service Interface Pattern

```go
// Standard service interface pattern
type ClusterService interface {
    CreateCluster(config models.ClusterConfig) error
    DeleteCluster(name string) error
    ListClusters() ([]models.Cluster, error)
    GetClusterStatus(name string) (*models.ClusterStatus, error)
    CleanupResources(name string) error
}

// Service implementation with dependency injection
type clusterService struct {
    k3dProvider   providers.K3dProvider
    ui           ui.ClusterUI
    prerequisites prerequisites.Checker
    logger       log.Logger
}
```

### Domain Layer (`internal/*/models/`)

The domain layer defines the core business entities, validation rules, and interfaces that represent the problem domain.

| Model Category | Purpose | Key Types |
|---------------|---------|-----------|
| **Cluster Models** | Cluster configuration and state | `ClusterConfig`, `ClusterStatus`, `ClusterType` |
| **Chart Models** | Chart installation and deployment | `DeploymentMode`, `ChartConfig`, `ArgoConfig` |
| **Dev Models** | Development workflow configuration | `InterceptConfig`, `SkaffoldConfig` |
| **Common Models** | Shared data structures | `GlobalFlags`, `Prerequisites`, `UIConfig` |

#### Model Validation Pattern

```go
// Domain models with embedded validation
type ClusterConfig struct {
    Name       string      `validate:"required,cluster_name"`
    Type       ClusterType `validate:"required,oneof=k3d kind"`
    K8sVersion string      `validate:"omitempty,k8s_version"`
    NodeCount  int         `validate:"min=1,max=10"`
    Ports      []PortMap   `validate:"dive"`
}

// Validation method pattern
func (c ClusterConfig) Validate() error {
    return validator.New().Struct(c)
}

// Custom validation rules
func ValidateClusterName(fl validator.FieldLevel) bool {
    name := fl.Field().String()
    return regexp.MustCompile(`^[a-z0-9-]+$`).MatchString(name)
}
```

## Data Flow Architecture

The following sequence diagram shows the typical data flow for the `openframe bootstrap` command, which demonstrates the interaction patterns used throughout the system.

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Command
    participant Bootstrap as Bootstrap Service
    participant Cluster as Cluster Service
    participant Chart as Chart Service
    participant UI as UI Components
    participant K3d as K3d Provider
    participant Helm as Helm Provider
    participant Prerequisites as Prerequisites Checker
    
    User->>CLI: openframe bootstrap my-cluster
    CLI->>UI: Display logo and welcome
    CLI->>Prerequisites: Check system requirements
    Prerequisites-->>CLI: Validation results
    
    alt Prerequisites Missing
        CLI->>UI: Show installation instructions
        CLI-->>User: Exit with error
    end
    
    CLI->>Bootstrap: Execute(cmd, args)
    Bootstrap->>UI: Start progress tracking
    
    Note over Bootstrap: Phase 1: Cluster Creation
    Bootstrap->>Cluster: CreateCluster(config)
    Cluster->>UI: Prompt for cluster configuration
    UI->>User: Interactive configuration wizard
    User-->>UI: Configuration choices
    UI-->>Cluster: Validated cluster config
    
    Cluster->>K3d: Create cluster with config
    K3d->>K3d: Pull images, create network
    K3d-->>Cluster: Cluster ready
    Cluster->>UI: Update progress (cluster ready)
    Cluster-->>Bootstrap: Cluster created successfully
    
    Note over Bootstrap: Phase 2: Chart Installation  
    Bootstrap->>Chart: InstallCharts(clusterName, mode)
    Chart->>UI: Prompt for deployment mode
    UI->>User: Deployment mode selection
    User-->>UI: Selected mode (oss-tenant/saas-tenant/etc)
    UI-->>Chart: Deployment configuration
    
    Chart->>Helm: Add ArgoCD repository
    Chart->>Helm: Install ArgoCD with values
    Helm->>Helm: Deploy ArgoCD components
    Helm-->>Chart: ArgoCD installed
    
    Chart->>Chart: Wait for ArgoCD readiness
    Chart->>Helm: Install OpenFrame app-of-apps
    Helm-->>Chart: Apps deployed
    Chart->>UI: Update progress (charts ready)
    Chart-->>Bootstrap: Charts installed successfully
    
    Bootstrap->>UI: Show success summary
    UI->>User: Environment ready message
    Bootstrap-->>CLI: Success
    CLI-->>User: Command completed
```

## Key Design Patterns

### 1. Dependency Injection Pattern

Services receive their dependencies through constructor injection, making the code testable and modular.

```go
// Service constructor with dependency injection
func NewClusterService(
    k3dProvider providers.K3dProvider,
    ui ui.ClusterUI,
    prereq prerequisites.Checker,
    logger log.Logger,
) ClusterService {
    return &clusterService{
        k3dProvider:   k3dProvider,
        ui:           ui,
        prerequisites: prereq,
        logger:       logger,
    }
}

// Usage in command layer
func runCreateCluster(cmd *cobra.Command, args []string) error {
    // Dependencies injected during service creation
    service := utils.GetCommandService()
    return service.CreateCluster(config)
}
```

### 2. Provider Pattern for Infrastructure

Infrastructure concerns are abstracted behind provider interfaces, allowing easy testing and future extensibility.

```go
// Provider interface for cluster operations
type K3dProvider interface {
    CreateCluster(config K3dConfig) error
    DeleteCluster(name string) error  
    ListClusters() ([]K3dCluster, error)
    GetClusterStatus(name string) (*K3dStatus, error)
}

// Implementation can be swapped for testing
type k3dProvider struct {
    client k3d.Client
}

// Mock implementation for testing  
type mockK3dProvider struct {
    clusters map[string]*K3dCluster
}
```

### 3. UI Abstraction Pattern

User interface interactions are abstracted into dedicated UI components, separating presentation logic from business logic.

```go
// UI interface for cluster operations
type ClusterUI interface {
    ShowLogo()
    PromptClusterConfig(defaultName string) (ClusterConfig, error)
    ShowProgress(message string)
    ShowConfigurationSummary(config ClusterConfig, dryRun bool)
    ShowSuccess(cluster Cluster)
    ShowError(err error)
}

// Implementation handles all user interaction details
type clusterUI struct {
    interactive bool
    verbose     bool
}

func (ui *clusterUI) PromptClusterConfig(defaultName string) (ClusterConfig, error) {
    // Interactive prompts, validation, etc.
    return config, nil
}
```

### 4. Command Wrapper Pattern

Common command setup and error handling are abstracted into wrapper functions.

```go
// Command wrapper providing common functionality
func WrapCommandWithCommonSetup(runFunc func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
    return func(cmd *cobra.Command, args []string) error {
        // Common setup: logo, prerequisites, context
        ui.ShowLogo()
        
        if err := checkPrerequisites(); err != nil {
            return err
        }
        
        // Execute the actual command
        if err := runFunc(cmd, args); err != nil {
            ui.ShowError(err)
            return err
        }
        
        return nil
    }
}
```

## Module Dependencies and Relationships

### Internal Module Structure

```mermaid
graph TB
    subgraph "cmd/ - Command Layer"
        CMD_BOOTSTRAP[bootstrap/]
        CMD_CLUSTER[cluster/]  
        CMD_CHART[chart/]
        CMD_DEV[dev/]
    end
    
    subgraph "internal/ - Internal Packages"
        INT_BOOTSTRAP[bootstrap/]
        INT_CLUSTER[cluster/]
        INT_CHART[chart/] 
        INT_DEV[dev/]
        INT_SHARED[shared/]
    end
    
    subgraph "External Dependencies"
        COBRA[github.com/spf13/cobra]
        K8S[k8s.io/client-go]
        HELM[helm.sh/helm/v3]
    end
    
    CMD_BOOTSTRAP --> INT_BOOTSTRAP
    CMD_CLUSTER --> INT_CLUSTER
    CMD_CHART --> INT_CHART
    CMD_DEV --> INT_DEV
    
    INT_BOOTSTRAP --> INT_CLUSTER
    INT_BOOTSTRAP --> INT_CHART
    INT_CLUSTER --> INT_SHARED
    INT_CHART --> INT_SHARED
    INT_DEV --> INT_SHARED
    
    CMD_BOOTSTRAP --> COBRA
    CMD_CLUSTER --> COBRA
    CMD_CHART --> COBRA
    CMD_DEV --> COBRA
    
    INT_CLUSTER --> K8S
    INT_CHART --> HELM
    INT_DEV --> K8S
    
    style INT_SHARED fill:#f9f9f9
    style COBRA fill:#e3f2fd
    style K8S fill:#e3f2fd
    style HELM fill:#e3f2fd
```

### Dependency Rules

| Layer | Can Import From | Cannot Import From | Rationale |
|-------|-----------------|-------------------|-----------|
| **cmd/** | `internal/*`, external packages | Other `cmd/*` packages | Commands are entry points, not shared |
| **internal/services/** | `internal/models`, `internal/providers`, external | `cmd/*`, other service packages | Services are independent business logic |
| **internal/models/** | Standard library, validation packages | `internal/services`, `cmd/*` | Models are pure domain objects |
| **internal/providers/** | `internal/models`, external clients | `internal/services`, `cmd/*` | Providers are infrastructure adapters |
| **internal/shared/** | Standard library, common external packages | Other `internal/*` packages except models | Shared utilities only |

## Performance and Scalability Considerations

### Resource Management

```go
// Resource cleanup pattern used throughout
type ResourceManager interface {
    Acquire() error
    Release() error
}

// Example: Docker resource management
type dockerResourceManager struct {
    containers []string
    networks   []string
}

func (r *dockerResourceManager) Acquire() error {
    // Create resources
    return nil
}

func (r *dockerResourceManager) Release() error {
    // Cleanup containers, networks, volumes
    for _, container := range r.containers {
        docker.RemoveContainer(container)
    }
    return nil
}
```

### Concurrent Operations

```go
// Safe concurrent operations with proper synchronization
type ClusterCache struct {
    mu       sync.RWMutex
    clusters map[string]*ClusterInfo
}

func (c *ClusterCache) GetCluster(name string) (*ClusterInfo, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    cluster, exists := c.clusters[name]
    return cluster, exists
}

func (c *ClusterCache) UpdateCluster(name string, info *ClusterInfo) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.clusters[name] = info
}
```

### Memory Optimization

- **Lazy Loading**: Cluster information loaded on-demand
- **Resource Pooling**: Docker client connections reused
- **Stream Processing**: Large outputs (logs, events) processed as streams
- **Context Cancellation**: Operations can be cancelled gracefully

## Error Handling Strategy

### Error Hierarchy

```go
// Custom error types with context
type OpenFrameError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Context map[string]interface{}
}

type ErrorCode string

const (
    ErrClusterExists     ErrorCode = "CLUSTER_EXISTS"
    ErrClusterNotFound   ErrorCode = "CLUSTER_NOT_FOUND"
    ErrInvalidConfig     ErrorCode = "INVALID_CONFIG"
    ErrPrerequisites     ErrorCode = "PREREQUISITES_FAILED"
    ErrInfrastructure    ErrorCode = "INFRASTRUCTURE_ERROR"
)

// Error wrapping with context
func WrapError(err error, code ErrorCode, message string) error {
    return &OpenFrameError{
        Code:    code,
        Message: message,
        Cause:   err,
        Context: map[string]interface{}{
            "timestamp": time.Now(),
            "version":   version.GetVersion(),
        },
    }
}
```

### Error Recovery Patterns

```go
// Retry with exponential backoff
func (s *clusterService) CreateClusterWithRetry(config ClusterConfig) error {
    return retry.Do(
        func() error {
            return s.k3dProvider.CreateCluster(config)
        },
        retry.Attempts(3),
        retry.Delay(time.Second),
        retry.DelayType(retry.BackOffDelay),
        retry.OnRetry(func(n uint, err error) {
            s.logger.Warnf("Cluster creation attempt %d failed: %v", n+1, err)
        }),
    )
}

// Graceful degradation
func (s *chartService) InstallWithFallback(config ChartConfig) error {
    // Try ArgoCD installation
    if err := s.installArgoCD(config); err != nil {
        s.logger.Warnf("ArgoCD installation failed: %v", err)
        
        // Fallback to direct Helm installation
        s.ui.ShowWarning("Falling back to direct Helm installation")
        return s.installHelmCharts(config)
    }
    return nil
}
```

## Testing Architecture

### Testing Pyramid

```mermaid
pyramid
    title OpenFrame CLI Testing Strategy
    
    level1: Unit Tests
        - Service layer logic
        - Model validation
        - UI components
        - Provider implementations
        
    level2: Integration Tests
        - Command execution
        - Service interactions  
        - External tool integration
        - Configuration handling
        
    level3: End-to-End Tests
        - Full workflow scenarios
        - CLI command combinations
        - Real cluster operations
```

### Test Patterns

```go
// Service testing with mocked dependencies
func TestClusterService_CreateCluster(t *testing.T) {
    // Setup
    mockProvider := &mockK3dProvider{}
    mockUI := &mockClusterUI{}
    service := NewClusterService(mockProvider, mockUI, nil, nil)
    
    config := ClusterConfig{
        Name: "test-cluster",
        Type: ClusterTypeK3d,
        NodeCount: 1,
    }
    
    // Execute
    err := service.CreateCluster(config)
    
    // Assert
    assert.NoError(t, err)
    assert.True(t, mockProvider.CreateClusterCalled)
    assert.Equal(t, "test-cluster", mockProvider.LastConfig.Name)
}

// Integration testing with real dependencies
func TestBootstrapIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }
    
    // Setup test environment
    testCluster := "test-bootstrap-" + uuid.New().String()[:8]
    
    // Cleanup
    defer func() {
        exec.Command("openframe", "cluster", "delete", testCluster).Run()
    }()
    
    // Execute bootstrap
    cmd := exec.Command("openframe", "bootstrap", testCluster, "--non-interactive")
    output, err := cmd.CombinedOutput()
    
    // Assert
    assert.NoError(t, err)
    assert.Contains(t, string(output), "Bootstrap completed successfully")
}
```

## Future Architecture Considerations

### Extensibility Points

1. **Plugin System**: Design interfaces for future plugin support
2. **Multiple Providers**: Abstract cluster providers for kind, minikube, etc.
3. **Custom Charts**: Support for user-defined chart repositories
4. **Remote Clusters**: Extend beyond local development clusters

### Scalability Improvements

1. **Parallel Operations**: Concurrent cluster operations
2. **Caching Layer**: Persistent cache for cluster state
3. **Event System**: Pub/sub for cross-component communication
4. **Configuration Management**: Advanced config file support

---

This architecture overview provides the foundation for understanding and extending the OpenFrame CLI codebase. For specific implementation details, refer to the inline code documentation and the [Developer Getting Started Guide](getting-started-dev.md).