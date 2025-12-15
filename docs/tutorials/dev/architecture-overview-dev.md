# OpenFrame CLI Architecture Overview

This document provides a comprehensive technical overview of the OpenFrame CLI architecture, design patterns, and internal workings. It's intended for developers who need to understand the system deeply for maintenance, extension, or integration purposes.

## System Architecture

OpenFrame CLI follows a layered architecture with clear separation of concerns, enabling maintainability and testability.

### High-Level Architecture

```mermaid
graph TB
    subgraph "User Interface Layer"
        CLI[CLI Interface]
        TUI[Terminal UI Components]
        Interactive[Interactive Wizards]
    end
    
    subgraph "Command Layer"
        Bootstrap[Bootstrap Command]
        Chart[Chart Command]
        Cluster[Cluster Command]
        Dev[Dev Command]
    end
    
    subgraph "Service Layer"
        BootstrapSvc[Bootstrap Service]
        ChartSvc[Chart Service]
        ClusterSvc[Cluster Service]
        PrereqSvc[Prerequisites Service]
    end
    
    subgraph "Provider Layer"
        ArgoCD[ArgoCD Provider]
        Helm[Helm Provider]
        K3d[K3d Provider]
        Git[Git Provider]
    end
    
    subgraph "Infrastructure Layer"
        Docker[Docker Engine]
        Kubernetes[Kubernetes API]
        FileSystem[File System]
        Network[Network Layer]
    end
    
    CLI --> Bootstrap
    CLI --> Chart
    CLI --> Cluster
    CLI --> Dev
    
    Bootstrap --> BootstrapSvc
    Chart --> ChartSvc
    Cluster --> ClusterSvc
    
    BootstrapSvc --> ChartSvc
    BootstrapSvc --> ClusterSvc
    
    ChartSvc --> PrereqSvc
    ChartSvc --> ArgoCD
    ChartSvc --> Helm
    ChartSvc --> Git
    
    ClusterSvc --> K3d
    
    PrereqSvc --> Helm
    PrereqSvc --> Git
    
    K3d --> Docker
    ArgoCD --> Kubernetes
    Helm --> Kubernetes
    
    TUI --> CLI
    Interactive --> CLI
    
    style CLI fill:#e1f5fe
    style BootstrapSvc fill:#f3e5f5
    style Docker fill:#e8f5e8
    style Kubernetes fill:#e8f5e8
```

## Core Components Deep Dive

### 1. Command Layer (`cmd/`)

The command layer implements the Cobra CLI framework and handles user interaction.

| Component | Responsibility | Key Files |
|-----------|---------------|-----------|
| **Root Command** | CLI entry point, global flags, version info | `cmd/root.go` |
| **Bootstrap Command** | Orchestrates complete environment setup | `cmd/bootstrap/bootstrap.go` |
| **Chart Command** | Manages Helm charts and ArgoCD installation | `cmd/chart/chart.go`, `cmd/chart/install.go` |
| **Cluster Command** | K3d cluster lifecycle management | `cmd/cluster/*.go` |
| **Dev Command** | Development tools and utilities | `cmd/dev/*.go` |

#### Command Pattern Implementation

```go
// Standard command structure
func GetCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "command [args]",
        Short: "Brief description",
        Long:  `Detailed description with examples`,
        RunE: func(cmd *cobra.Command, args []string) error {
            // 1. Parse flags and arguments
            // 2. Validate input
            // 3. Create service instance
            // 4. Execute business logic
            // 5. Handle errors consistently
            return service.Execute(cmd, args)
        },
    }
    
    // Flag definitions
    cmd.Flags().StringP("flag", "f", "default", "Description")
    
    return cmd
}
```

### 2. Service Layer (`internal/*/`)

The service layer contains the core business logic and orchestrates provider interactions.

#### Bootstrap Service Architecture

```mermaid
sequenceDiagram
    participant User
    participant BootstrapCmd as Bootstrap Command
    participant BootstrapSvc as Bootstrap Service
    participant ClusterSvc as Cluster Service
    participant ChartSvc as Chart Service
    participant K3d as K3d Provider
    participant ArgoCD as ArgoCD Provider
    
    User->>BootstrapCmd: openframe bootstrap
    BootstrapCmd->>BootstrapSvc: Execute(cmd, args)
    
    BootstrapSvc->>BootstrapSvc: Parse flags (deployment-mode, non-interactive)
    BootstrapSvc->>BootstrapSvc: Validate configuration
    
    BootstrapSvc->>ClusterSvc: CreateClusterWithPrerequisites()
    ClusterSvc->>K3d: Create cluster
    K3d-->>ClusterSvc: Cluster ready
    ClusterSvc-->>BootstrapSvc: Success
    
    BootstrapSvc->>ChartSvc: InstallChartsWithConfig()
    ChartSvc->>ChartSvc: Check prerequisites
    ChartSvc->>ArgoCD: Install ArgoCD
    ChartSvc->>ArgoCD: Deploy applications
    ArgoCD-->>ChartSvc: Installation complete
    ChartSvc-->>BootstrapSvc: Success
    
    BootstrapSvc-->>BootstrapCmd: Bootstrap complete
    BootstrapCmd-->>User: Success message
```

### 3. Provider Layer

Providers encapsulate external tool interactions and implement consistent interfaces.

#### Provider Interface Pattern

```go
// Example provider interface
type Provider interface {
    IsAvailable() bool
    Install() error
    Configure(config Config) error
    Execute(operation Operation) error
}

// Implementation example
type ArgoCDProvider struct {
    kubeConfig string
    namespace  string
}

func (p *ArgoCDProvider) Install() error {
    // Helm-based ArgoCD installation
    return helmProvider.Install("argocd", argoCDChartConfig)
}
```

## Data Flow Architecture

### Request Processing Flow

```mermaid
flowchart TD
    A[CLI Command Input] --> B[Flag Parsing & Validation]
    B --> C[Service Layer Routing]
    
    C --> D{Command Type}
    
    D -->|bootstrap| E[Bootstrap Service]
    D -->|chart| F[Chart Service]  
    D -->|cluster| G[Cluster Service]
    
    E --> H[Prerequisites Check]
    H --> I[Cluster Creation]
    I --> J[Chart Installation]
    J --> K[Application Deployment]
    
    F --> H
    F --> L[ArgoCD Installation]
    L --> M[Repository Setup]
    M --> N[Application Sync]
    
    G --> O[K3d Operations]
    O --> P[Docker Interactions]
    
    K --> Q[Status Monitoring]
    N --> Q
    P --> Q
    
    Q --> R[UI Updates]
    R --> S[Success/Error Response]
    
    style A fill:#e3f2fd
    style S fill:#c8e6c9
    style H fill:#fff3e0
```

### Configuration Flow

```mermaid
graph LR
    A[CLI Flags] --> B[Config Merger]
    C[Config Files] --> B
    D[Environment Variables] --> B
    E[Defaults] --> B
    
    B --> F[Validated Configuration]
    F --> G[Service Execution]
    
    subgraph "Config Sources (Priority Order)"
        A
        C
        D
        E
    end
```

## Key Design Patterns

### 1. Command Pattern
- **Usage**: CLI command structure
- **Implementation**: Each command is a self-contained unit with defined interface
- **Benefits**: Easy to add new commands, testable in isolation

### 2. Service Layer Pattern
- **Usage**: Business logic separation
- **Implementation**: Services orchestrate provider interactions
- **Benefits**: Clear separation of concerns, reusable business logic

### 3. Provider Pattern
- **Usage**: External tool integration
- **Implementation**: Consistent interfaces for different tools
- **Benefits**: Easy to swap implementations, mockable for testing

### 4. Factory Pattern
- **Usage**: Service and provider creation
- **Implementation**: `NewService()` functions with dependency injection
- **Benefits**: Centralized object creation, easy to configure

### 5. Strategy Pattern
- **Usage**: Deployment mode handling
- **Implementation**: Different strategies for oss-tenant, saas-tenant, saas-shared
- **Benefits**: Extensible deployment configurations

## Module Dependencies and Relationships

### Dependency Graph

```mermaid
graph TD
    subgraph "External Dependencies"
        Cobra[github.com/spf13/cobra]
        Pterm[github.com/pterm/pterm]
        Promptui[github.com/manifoldco/promptui]
        Testify[github.com/stretchr/testify]
        Yaml[gopkg.in/yaml.v3]
    end
    
    subgraph "Internal Modules"
        CmdRoot[cmd/root]
        CmdBootstrap[cmd/bootstrap]
        CmdChart[cmd/chart]
        CmdCluster[cmd/cluster]
        
        InternalBootstrap[internal/bootstrap]
        InternalChart[internal/chart]
        InternalShared[internal/shared]
    end
    
    subgraph "Internal Sub-modules"
        ChartModels[internal/chart/models]
        ChartPrerequisites[internal/chart/prerequisites]
        ChartProviders[internal/chart/providers]
        SharedUI[internal/shared/ui]
        SharedConfig[internal/shared/config]
        SharedErrors[internal/shared/errors]
    end
    
    CmdRoot --> Cobra
    CmdRoot --> CmdBootstrap
    CmdRoot --> CmdChart
    CmdRoot --> CmdCluster
    
    CmdBootstrap --> InternalBootstrap
    CmdChart --> InternalChart
    
    InternalBootstrap --> InternalChart
    InternalChart --> ChartModels
    InternalChart --> ChartPrerequisites
    InternalChart --> ChartProviders
    
    SharedUI --> Pterm
    SharedUI --> Promptui
    SharedConfig --> Yaml
    
    InternalBootstrap --> SharedErrors
    InternalChart --> SharedUI
    InternalChart --> SharedConfig
```

### Module Responsibility Matrix

| Module | Purpose | Dependencies | Exports |
|--------|---------|--------------|---------|
| `cmd/root` | CLI entry point, global configuration | cobra, internal services | Root command |
| `cmd/bootstrap` | Bootstrap orchestration | internal/bootstrap | Bootstrap command |
| `cmd/chart` | Chart management commands | internal/chart | Chart commands |
| `internal/bootstrap` | Bootstrap business logic | chart, cluster services | Bootstrap service |
| `internal/chart/prerequisites` | System validation | git, helm, certificates | Prerequisite checkers |
| `internal/chart/providers` | External tool providers | argocd, helm, git providers | Provider interfaces |
| `internal/shared/ui` | Terminal UI components | pterm, promptui | UI utilities |

## System Integration Points

### 1. Kubernetes Integration

```go
// Kubernetes client configuration
type KubernetesProvider struct {
    clientset *kubernetes.Clientset
    config    *rest.Config
}

// Integration points:
// - Cluster health checks
// - Resource status monitoring  
// - Application deployment status
// - Log retrieval
```

### 2. Docker Integration

```go
// Docker integration for K3d
type DockerProvider struct {
    client *docker.Client
}

// Integration points:
// - Container lifecycle management
// - Image pulling and management
// - Network configuration
// - Volume management
```

### 3. Git Integration

```go
// Git provider for repository operations
type GitProvider struct {
    repoURL    string
    branch     string
    localPath  string
}

// Integration points:
// - Repository cloning
// - Branch switching
// - Authentication handling
// - Update checking
```

## Error Handling Architecture

### Error Types and Hierarchy

```go
// Custom error types
type OpenFrameError struct {
    Code    string
    Message string
    Cause   error
}

// Specific error types
type ValidationError struct{ OpenFrameError }
type PrerequisiteError struct{ OpenFrameError }
type ProviderError struct{ OpenFrameError }
type NetworkError struct{ OpenFrameError }
```

### Error Flow

```mermaid
flowchart TD
    A[Error Occurs] --> B[Error Wrapping]
    B --> C[Error Classification]
    C --> D{Error Type}
    
    D -->|ValidationError| E[User Input Issue]
    D -->|PrerequisiteError| F[Missing Dependencies]
    D -->|ProviderError| G[External Tool Issue]
    D -->|NetworkError| H[Connectivity Issue]
    
    E --> I[Show Usage Help]
    F --> J[Show Install Instructions]
    G --> K[Show Provider Status]
    H --> L[Show Network Diagnostics]
    
    I --> M[Graceful Exit]
    J --> M
    K --> M
    L --> M
```

## Testing Architecture

### Test Organization

| Test Type | Location | Purpose | Tools |
|-----------|----------|---------|-------|
| **Unit Tests** | `*_test.go` | Individual function testing | testify |
| **Integration Tests** | `*_integration_test.go` | Service interaction testing | testify, docker |
| **E2E Tests** | `tests/e2e/` | Full workflow testing | Custom framework |
| **Mock Tests** | `mocks/` | Provider interface testing | testify/mock |

### Test Patterns

```go
// Table-driven tests
func TestBootstrapService_Execute(t *testing.T) {
    tests := []struct {
        name           string
        args           []string
        deploymentMode string
        nonInteractive bool
        wantErr        bool
        expectedCalls  map[string]int
    }{
        {
            name:           "successful bootstrap",
            args:           []string{"test-cluster"},
            deploymentMode: "oss-tenant",
            nonInteractive: true,
            wantErr:        false,
            expectedCalls:  map[string]int{"CreateCluster": 1, "InstallCharts": 1},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Configuration Architecture

### Configuration Hierarchy

```go
// Configuration structure
type Config struct {
    Global    GlobalConfig    `yaml:"global"`
    Bootstrap BootstrapConfig `yaml:"bootstrap"`
    Chart     ChartConfig     `yaml:"chart"`
    Cluster   ClusterConfig   `yaml:"cluster"`
}

// Configuration sources (priority order)
// 1. Command line flags
// 2. Environment variables
// 3. Configuration files
// 4. Default values
```

### Configuration Flow

```mermaid
graph TD
    A[CLI Flags] --> E[Config Merger]
    B[~/.openframe/config.yaml] --> E
    C[Environment Variables] --> E
    D[Default Values] --> E
    
    E --> F[Validation]
    F --> G[Service Configuration]
    G --> H[Provider Configuration]
    
    style A fill:#ffeb3b
    style B fill:#4caf50
    style C fill:#ff9800
    style D fill:#9e9e9e
```

## Performance Considerations

### Optimization Strategies

| Area | Strategy | Implementation |
|------|----------|----------------|
| **Startup Time** | Lazy loading | Load providers only when needed |
| **Memory Usage** | Resource pooling | Reuse Kubernetes clients |
| **Network I/O** | Caching | Cache prerequisite checks |
| **Disk I/O** | Batch operations | Group file operations |

### Monitoring Points

```go
// Performance monitoring
type Metrics struct {
    CommandDuration    time.Duration
    ClusterCreateTime  time.Duration
    ChartInstallTime   time.Duration
    MemoryUsage       int64
}
```

## Security Architecture

### Security Considerations

```mermaid
graph TD
    A[Input Validation] --> B[Authentication]
    B --> C[Authorization]
    C --> D[Secure Communication]
    D --> E[Credential Management]
    E --> F[Audit Logging]
    
    subgraph "Security Layers"
        A
        B
        C
        D
        E
        F
    end
    
    G[External APIs] --> D
    H[File System] --> E
    I[Network] --> D
```

### Security Measures

1. **Input Sanitization**: All user inputs are validated
2. **Credential Storage**: Secure handling of certificates and tokens
3. **Network Security**: TLS for all external communications
4. **File Permissions**: Proper permissions for configuration files
5. **Audit Trail**: Logging of security-relevant operations

## Extension Points

### Adding New Commands

```go
// Extension interface
type CommandProvider interface {
    GetCommand() *cobra.Command
    GetServiceDependencies() []string
}

// Registration in root command
func registerCommand(provider CommandProvider) {
    cmd := provider.GetCommand()
    rootCmd.AddCommand(cmd)
}
```

### Adding New Providers

```go
// Provider interface
type Provider interface {
    Name() string
    IsAvailable() bool
    Install() error
    Configure(map[string]interface{}) error
}

// Provider registration
func RegisterProvider(name string, provider Provider) {
    providers[name] = provider
}
```

## Future Architecture Considerations

### Planned Enhancements

1. **Plugin System**: Support for external plugins
2. **Remote Configuration**: Centralized configuration management
3. **Multi-Cluster Support**: Enhanced cluster orchestration
4. **Event System**: Pub/sub for component communication
5. **GraphQL API**: API-first architecture for UI separation

### Scalability Roadmap

```mermaid
timeline
    title Architecture Evolution
    
    Phase 1 : Current CLI
            : Monolithic architecture
            : Local operations only
    
    Phase 2 : Plugin System
            : Modular architecture
            : External integrations
    
    Phase 3 : API-First
            : Microservices architecture
            : Remote operations
    
    Phase 4 : Cloud Native
            : Distributed architecture
            : Multi-tenant support
```

---

## Conclusion

The OpenFrame CLI architecture is designed for maintainability, extensibility, and reliability. The layered approach with clear separation of concerns enables:

- **Easy Testing**: Each layer can be tested independently
- **Simple Extension**: New commands and providers can be added easily
- **Maintainable Code**: Clear responsibilities and interfaces
- **Robust Error Handling**: Consistent error propagation and handling
- **Performance**: Optimized resource usage and caching

This architecture supports the current requirements while providing a foundation for future enhancements and scaling.

---

> **📚 Next Steps**: Explore the [Developer Getting Started Guide](getting-started-dev.md) to begin contributing to this architecture.