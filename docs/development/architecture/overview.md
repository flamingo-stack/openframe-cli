# Architecture Overview

OpenFrame CLI is built using clean architecture principles with a layered design that promotes maintainability, testability, and extensibility. This document provides a comprehensive overview of the system's architecture, design patterns, and component relationships.

## High-Level Architecture

```mermaid
graph TB
    subgraph "Presentation Layer"
        CLI[CLI Commands]
        UI[Interactive UI]
        Flags[Command Flags]
    end
    
    subgraph "Application Layer"  
        Bootstrap[Bootstrap Service]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Development Service]
    end
    
    subgraph "Domain Layer"
        Models[Domain Models]
        Interfaces[Service Interfaces]
        BusinessLogic[Business Rules]
    end
    
    subgraph "Infrastructure Layer"
        K3DProvider[K3D Provider]
        HelmProvider[Helm Provider]
        ArgoCDProvider[ArgoCD Provider]
        GitProvider[Git Provider]
        TelepresenceProvider[Telepresence Provider]
    end
    
    subgraph "External Systems"
        Docker[Docker Engine]
        K8s[Kubernetes API]
        Git[Git Repositories]
        Registry[Container Registry]
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
    GitProvider --> Git
```

## Core Architectural Principles

### 🏗️ **Clean Architecture**
- **Separation of Concerns**: Each layer has a single, well-defined responsibility
- **Dependency Inversion**: High-level modules don't depend on low-level modules
- **Interface Segregation**: Clients depend only on interfaces they use
- **Single Responsibility**: Each component has one reason to change

### 🔄 **Dependency Injection**
- Constructor-based injection for better testability
- Interface-based abstractions for loose coupling
- Mockable dependencies for comprehensive testing

### 🎯 **Provider Pattern**
- Pluggable implementations for different platforms
- Consistent interfaces across all external integrations
- Easy extension and customization

## Layer Breakdown

### 1. Presentation Layer (`cmd/`)

The presentation layer handles user interaction through the CLI interface.

```mermaid
graph LR
    subgraph "CLI Commands"
        Root[root.go]
        Bootstrap[bootstrap/]
        Cluster[cluster/]
        Chart[chart/]
        Dev[dev/]
    end
    
    Root --> Bootstrap
    Root --> Cluster
    Root --> Chart
    Root --> Dev
```

**Key Responsibilities:**
- Command line argument parsing
- User input validation
- Interactive prompts and wizards
- Output formatting and display
- Error presentation

**Core Components:**

| Component | Purpose |
|-----------|---------|
| **Root Command** | CLI entry point and global configuration |
| **Bootstrap Command** | Orchestrates complete environment setup |
| **Cluster Commands** | K3d cluster lifecycle management |
| **Chart Commands** | Helm and ArgoCD operations |
| **Dev Commands** | Development workflow tools |

### 2. Application Layer (`internal/*/services/`)

The application layer contains business logic and orchestrates use cases.

```mermaid
graph TB
    subgraph "Application Services"
        BootstrapSvc[Bootstrap Service]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        InterceptSvc[Intercept Service]
        ScaffoldSvc[Scaffold Service]
    end
    
    subgraph "Shared Services"
        Config[Configuration Service]
        Validation[Validation Service]
        Prerequisites[Prerequisites Service]
    end
    
    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc
    ClusterSvc --> Prerequisites
    ChartSvc --> Prerequisites
    InterceptSvc --> Config
    ScaffoldSvc --> Validation
```

**Service Responsibilities:**

| Service | Purpose |
|---------|---------|
| **Bootstrap** | Coordinates cluster creation and chart installation |
| **Cluster** | Manages K3d cluster lifecycle operations |
| **Chart** | Handles Helm charts and ArgoCD applications |
| **Intercept** | Manages Telepresence service intercepts |
| **Scaffold** | Generates development scaffolding |

### 3. Domain Layer (`internal/*/models/`)

The domain layer contains business entities, rules, and interfaces.

```mermaid
classDiagram
    class Cluster {
        +Name string
        +Status ClusterStatus
        +Config ClusterConfig
        +Validate() error
        +IsReady() bool
    }
    
    class Application {
        +Name string
        +Namespace string
        +SyncStatus SyncStatus
        +Health HealthStatus
        +Deploy() error
    }
    
    class InterceptConfig {
        +ServiceName string
        +Namespace string
        +Port int
        +LocalPort int
        +Start() error
        +Stop() error
    }
    
    class ClusterProvider {
        <<interface>>
        +Create(config) error
        +Delete(name) error
        +List() []Cluster
        +Status(name) ClusterStatus
    }
    
    Cluster --> ClusterProvider
    Application --> ChartProvider
    InterceptConfig --> InterceptProvider
```

**Domain Models:**
- **Cluster**: Represents Kubernetes cluster state and operations
- **Application**: ArgoCD application with sync and health status
- **Chart**: Helm chart with configuration and dependencies
- **Intercept**: Telepresence intercept configuration

### 4. Infrastructure Layer (`internal/*/providers/`)

The infrastructure layer implements external system integrations.

```mermaid
graph TB
    subgraph "Provider Implementations"
        K3D[K3D Manager]
        Helm[Helm Manager]
        ArgoCD[ArgoCD Manager]
        Git[Git Repository]
        Telepresence[Telepresence Provider]
        Kubectl[Kubectl Provider]
    end
    
    subgraph "External APIs"
        DockerAPI[Docker API]
        K8sAPI[Kubernetes API]
        GitAPI[Git Repositories]
        HelmRepos[Helm Repositories]
    end
    
    K3D --> DockerAPI
    Helm --> K8sAPI
    Helm --> HelmRepos
    ArgoCD --> K8sAPI
    Git --> GitAPI
    Telepresence --> K8sAPI
    Kubectl --> K8sAPI
```

**Provider Interfaces:**

```go
type ClusterProvider interface {
    Create(ctx context.Context, config ClusterConfig) error
    Delete(ctx context.Context, name string) error
    List(ctx context.Context) ([]Cluster, error)
    Status(ctx context.Context, name string) (ClusterStatus, error)
}

type ChartProvider interface {
    Install(ctx context.Context, chart Chart) error
    Upgrade(ctx context.Context, chart Chart) error
    Uninstall(ctx context.Context, name string) error
    Status(ctx context.Context, name string) (ChartStatus, error)
}

type InterceptProvider interface {
    Start(ctx context.Context, config InterceptConfig) error
    Stop(ctx context.Context, name string) error
    List(ctx context.Context) ([]Intercept, error)
}
```

## Data Flow Architecture

### Bootstrap Workflow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Bootstrap
    participant Cluster
    participant Chart
    participant K3D
    participant Helm
    participant ArgoCD
    
    User->>CLI: openframe bootstrap
    CLI->>Bootstrap: Execute()
    
    Bootstrap->>Cluster: ValidatePrerequisites()
    Cluster-->>Bootstrap: Prerequisites OK
    
    Bootstrap->>Cluster: CreateCluster()
    Cluster->>K3D: Create(config)
    K3D->>K3D: Pull images
    K3D->>K3D: Start containers
    K3D-->>Cluster: Cluster ready
    Cluster-->>Bootstrap: Success
    
    Bootstrap->>Chart: InstallCharts()
    Chart->>Helm: Install ArgoCD
    Helm->>Helm: Apply manifests
    Helm-->>Chart: ArgoCD ready
    
    Chart->>ArgoCD: Deploy app-of-apps
    ArgoCD->>ArgoCD: Sync applications
    ArgoCD-->>Chart: Applications synced
    Chart-->>Bootstrap: Charts installed
    
    Bootstrap-->>CLI: Complete
    CLI-->>User: Environment ready
```

### Development Intercept Workflow

```mermaid
sequenceDiagram
    participant Dev
    participant CLI
    participant Intercept
    participant Kubectl
    participant Telepresence
    participant K8s
    
    Dev->>CLI: openframe dev intercept
    CLI->>Intercept: StartIntercept()
    
    Intercept->>Kubectl: DiscoverServices()
    Kubectl->>K8s: List services
    K8s-->>Kubectl: Service list
    Kubectl-->>Intercept: Available services
    
    Intercept->>Intercept: PromptForService()
    Intercept->>Telepresence: Connect()
    Telepresence->>K8s: Establish tunnel
    K8s-->>Telepresence: Tunnel ready
    
    Telepresence->>Telepresence: CreateIntercept()
    Telepresence->>K8s: Route traffic
    K8s-->>Telepresence: Traffic routed
    Telepresence-->>Intercept: Intercept active
    
    Intercept-->>CLI: Success
    CLI-->>Dev: Local development ready
```

## Component Interactions

### Shared Infrastructure Components

```mermaid
graph TB
    subgraph "Shared Infrastructure"
        Executor[Command Executor]
        UI[UI Components]
        Config[Configuration]
        Errors[Error Handling]
        Logger[Logging]
        Progress[Progress Tracking]
    end
    
    subgraph "Services"
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Dev Service]
    end
    
    ClusterSvc --> Executor
    ClusterSvc --> UI
    ClusterSvc --> Config
    ClusterSvc --> Progress
    
    ChartSvc --> Executor
    ChartSvc --> UI  
    ChartSvc --> Errors
    
    DevSvc --> Executor
    DevSvc --> Logger
    DevSvc --> Progress
```

**Shared Components:**

| Component | Purpose |
|-----------|---------|
| **Command Executor** | Abstracts external command execution with logging and error handling |
| **UI Components** | Consistent terminal UI with progress bars, prompts, and formatting |
| **Configuration** | Manages CLI settings, credentials, and environment variables |
| **Error Handling** | Centralized error processing with user-friendly messages |
| **Progress Tracking** | Visual progress indicators for long-running operations |

### Configuration Management

```mermaid
graph LR
    subgraph "Configuration Sources"
        Flags[Command Flags]
        Env[Environment Variables]
        Files[Config Files]
        Defaults[Default Values]
    end
    
    subgraph "Configuration Hierarchy"
        Merged[Merged Configuration]
    end
    
    Flags --> Merged
    Env --> Merged
    Files --> Merged
    Defaults --> Merged
    
    Merged --> Services[Application Services]
```

Configuration precedence (highest to lowest):
1. Command line flags
2. Environment variables
3. Configuration files
4. Default values

## Error Handling Strategy

### Error Types and Handling

```mermaid
graph TB
    subgraph "Error Categories"
        UserError[User Input Errors]
        SystemError[System Errors]
        NetworkError[Network Errors]
        InfraError[Infrastructure Errors]
    end
    
    subgraph "Error Handling"
        Validation[Input Validation]
        Retry[Retry Logic]
        Fallback[Fallback Strategies]
        Recovery[Error Recovery]
    end
    
    UserError --> Validation
    SystemError --> Retry
    NetworkError --> Retry
    InfraError --> Fallback
    
    Validation --> Recovery
    Retry --> Recovery
    Fallback --> Recovery
```

**Error Handling Patterns:**
- **Validation Errors**: Immediate feedback with suggestions
- **Transient Errors**: Automatic retry with exponential backoff
- **Infrastructure Errors**: Graceful degradation and cleanup
- **Fatal Errors**: Clean shutdown with helpful error messages

## Testing Architecture

### Test Layer Structure

```mermaid
graph TB
    subgraph "Test Types"
        Unit[Unit Tests]
        Integration[Integration Tests]
        E2E[End-to-End Tests]
        Performance[Performance Tests]
    end
    
    subgraph "Test Infrastructure"
        Mocks[Mock Objects]
        Fixtures[Test Fixtures]
        Utilities[Test Utilities]
        Helpers[Test Helpers]
    end
    
    Unit --> Mocks
    Integration --> Fixtures
    E2E --> Utilities
    Performance --> Helpers
```

**Testing Strategy:**
- **Unit Tests**: Fast, isolated tests with mocked dependencies
- **Integration Tests**: Service interaction tests with test containers
- **E2E Tests**: Complete workflow tests with real clusters
- **Performance Tests**: Benchmarks and load testing

## Extension Points

### Adding New Providers

```mermaid
graph LR
    subgraph "Provider Interface"
        IProvider[Provider Interface]
    end
    
    subgraph "Implementations"
        K3D[K3D Provider]
        Kind[Kind Provider]
        EKS[EKS Provider]
        New[New Provider]
    end
    
    IProvider --> K3D
    IProvider --> Kind
    IProvider --> EKS
    IProvider --> New
```

**Extension Process:**
1. Implement provider interface
2. Add configuration options
3. Register with provider factory
4. Add integration tests
5. Update documentation

### Plugin Architecture (Future)

```mermaid
graph TB
    subgraph "Plugin System"
        Registry[Plugin Registry]
        Loader[Plugin Loader]
        Manager[Plugin Manager]
    end
    
    subgraph "Plugin Types"
        ProviderPlugin[Provider Plugins]
        CommandPlugin[Command Plugins]
        UIPlugin[UI Plugins]
    end
    
    Registry --> Loader
    Loader --> Manager
    Manager --> ProviderPlugin
    Manager --> CommandPlugin
    Manager --> UIPlugin
```

## Design Decisions

### Key Architectural Decisions

| Decision | Rationale | Trade-offs |
|----------|-----------|------------|
| **Clean Architecture** | Maintainability and testability | More complexity for simple operations |
| **Provider Pattern** | Extensibility and platform support | Additional abstraction layer |
| **Cobra CLI Framework** | Rich CLI features and community support | Framework dependency |
| **Go Language** | Performance, concurrency, and deployment simplicity | Learning curve for some developers |
| **K3d for Local Development** | Lightweight and fast cluster creation | Limited to local development |

### Future Architecture Evolution

```mermaid
graph TB
    subgraph "Current Architecture"
        CLI[Monolithic CLI]
        Local[Local Only]
        K3D[K3D Only]
    end
    
    subgraph "Future Architecture"
        Microservices[Service Architecture]
        Cloud[Cloud Integration]
        MultiProvider[Multi-Provider Support]
        WebUI[Web Interface]
    end
    
    CLI --> Microservices
    Local --> Cloud
    K3D --> MultiProvider
    CLI --> WebUI
```

**Planned Enhancements:**
- Plugin system for custom providers
- Web-based management interface
- Cloud provider integrations
- Distributed service architecture
- Advanced GitOps workflows

## Performance Considerations

### Optimization Strategies

| Area | Strategy | Implementation |
|------|----------|----------------|
| **Command Startup** | Lazy loading of providers | On-demand initialization |
| **Concurrent Operations** | Go routines for parallel tasks | Cluster creation and chart installation |
| **Resource Usage** | Efficient resource cleanup | Defer statements and context cancellation |
| **Network Operations** | Connection pooling and caching | HTTP client reuse |

## Security Architecture

### Security Layers

```mermaid
graph TB
    subgraph "Security Layers"
        Input[Input Validation]
        Auth[Authentication]
        Authz[Authorization] 
        Audit[Audit Logging]
        Encryption[Data Encryption]
    end
    
    subgraph "Security Controls"
        RBAC[RBAC Integration]
        Secrets[Secret Management]
        Network[Network Security]
        Container[Container Security]
    end
    
    Input --> Auth
    Auth --> Authz
    Authz --> Audit
    Audit --> Encryption
    
    RBAC --> Authz
    Secrets --> Encryption
    Network --> Container
```

**Security Principles:**
- Least privilege access
- Secure by default configuration
- Credential management best practices
- Audit logging for compliance

## Next Steps

To dive deeper into the OpenFrame CLI architecture:

1. **[Code Structure Guide](code-structure.md)** - Detailed package organization
2. **[Design Patterns](design-patterns.md)** - Common patterns used throughout the codebase
3. **[API Reference](../reference/api.md)** - Internal API documentation
4. **[Testing Architecture](../testing/overview.md)** - How testing is structured

> **💡 Understanding the Flow**: Start by tracing a command from the CLI layer through the services to the providers. This will give you a concrete understanding of how the architecture works in practice.