# Architecture Overview

OpenFrame CLI follows a clean, layered architecture that separates concerns between user interface, business logic, and external system integration. This design makes the system maintainable, testable, and extensible.

## High-Level System Architecture

```mermaid
graph TB
    subgraph "User Interface Layer"
        CLI[CLI Commands<br/>Cobra Framework]
        UI[Interactive UI<br/>Survey Prompts]
    end
    
    subgraph "Service Layer"
        Bootstrap[Bootstrap Service<br/>Orchestration]
        Cluster[Cluster Service<br/>Lifecycle Management]
        Chart[Chart Service<br/>GitOps Deployment]
        Dev[Dev Service<br/>Local Workflows]
    end
    
    subgraph "Shared Components"
        Models[Data Models]
        Utils[Utilities]
        Errors[Error Handling]
        SharedUI[UI Components]
    end
    
    subgraph "External Integration Layer"
        K3d[K3d Client<br/>Cluster Management]
        Helm[Helm Client<br/>Package Management]
        ArgoCD[ArgoCD API<br/>GitOps Controller]
        Telepresence[Telepresence<br/>Dev Proxy]
        Skaffold[Skaffold<br/>Live Reload]
    end
    
    subgraph "External Systems"
        Docker[Docker Daemon]
        K8s[Kubernetes API]
        Git[Git Repositories]
    end
    
    CLI --> Bootstrap
    CLI --> Cluster
    CLI --> Chart
    CLI --> Dev
    UI --> Bootstrap
    UI --> Cluster
    UI --> Chart
    
    Bootstrap --> Cluster
    Bootstrap --> Chart
    
    Cluster --> Models
    Cluster --> Utils
    Cluster --> Errors
    Cluster --> SharedUI
    
    Chart --> Models
    Chart --> Utils
    Chart --> Errors
    Chart --> SharedUI
    
    Dev --> Models
    Dev --> Utils
    Dev --> Errors
    
    Bootstrap --> SharedUI
    Bootstrap --> Errors
    
    Cluster --> K3d
    Chart --> Helm
    Chart --> ArgoCD
    Dev --> Telepresence
    Dev --> Skaffold
    
    K3d --> Docker
    Helm --> K8s
    ArgoCD --> K8s
    ArgoCD --> Git
    Telepresence --> K8s
    Skaffold --> K8s
    Skaffold --> Docker
```

## Core Components

### CLI Layer (`cmd/`)

The command layer handles user input, flag parsing, and delegates work to service layers.

| Component | Package | Responsibilities |
|-----------|---------|------------------|
| **Bootstrap Command** | `cmd/bootstrap/` | Complete environment setup orchestration |
| **Cluster Commands** | `cmd/cluster/` | Cluster lifecycle management (create, delete, status) |
| **Chart Commands** | `cmd/chart/` | Helm chart installation and ArgoCD setup |
| **Dev Commands** | `cmd/dev/` | Development workflow integration |

**Key Design Patterns:**
- **Cobra Framework**: Structured CLI with subcommands and flags
- **Command Pattern**: Each command encapsulates a specific operation
- **Delegation**: Commands delegate business logic to service layer

### Service Layer (`internal/`)

Business logic layer that implements core functionality without CLI concerns.

| Service | Package | Purpose |
|---------|---------|---------|
| **Bootstrap Service** | `internal/bootstrap/` | Orchestrates cluster creation + chart installation |
| **Cluster Service** | `internal/cluster/services/` | K3d cluster lifecycle operations |
| **Chart Service** | `internal/chart/services/` | ArgoCD installation and chart management |
| **Dev Service** | `internal/dev/` | Telepresence and Skaffold integration |

**Key Design Patterns:**
- **Service Pattern**: Business logic separated from presentation
- **Dependency Injection**: Services receive dependencies via constructors
- **Interface Abstraction**: External dependencies behind interfaces for testing

### Shared Components

| Component | Package | Purpose |
|-----------|---------|---------|
| **Models** | `internal/*/models/` | Data structures and configuration objects |
| **UI Components** | `internal/shared/ui/` | Reusable UI elements (logo, prompts) |
| **Error Handling** | `internal/shared/errors/` | Centralized error management and wrapping |
| **Utilities** | `internal/*/utils/` | Helper functions and common operations |

## Component Relationships

### Internal Dependencies

```mermaid
graph TD
    subgraph "Bootstrap Service"
        BootstrapSvc[Bootstrap Service] --> ClusterSvc[Cluster Service]
        BootstrapSvc --> ChartSvc[Chart Service]
    end
    
    subgraph "Cluster Management"
        ClusterSvc --> ClusterModels[Cluster Models]
        ClusterSvc --> ClusterUI[Cluster UI]
        ClusterSvc --> ClusterUtils[Cluster Utils]
        ClusterSvc --> ClusterPrereqs[Prerequisites]
    end
    
    subgraph "Chart Management"
        ChartSvc --> ChartModels[Chart Models]
        ChartSvc --> ChartServices[Chart Services]
        ChartSvc --> ChartUtils[Chart Utils]
        ChartSvc --> ChartPrereqs[Chart Prerequisites]
    end
    
    subgraph "Development Tools"
        DevSvc[Dev Service] --> DevModels[Dev Models]
        DevSvc --> DevPrereqs[Dev Prerequisites]
    end
    
    subgraph "Shared Components"
        SharedUI[Shared UI]
        SharedErrors[Shared Errors]
    end
    
    ClusterSvc --> SharedUI
    ChartSvc --> SharedUI
    DevSvc --> SharedUI
    BootstrapSvc --> SharedUI
    
    ClusterSvc --> SharedErrors
    ChartSvc --> SharedErrors
    DevSvc --> SharedErrors
    BootstrapSvc --> SharedErrors
```

### Data Flow Patterns

#### Bootstrap Workflow

```mermaid
sequenceDiagram
    participant User
    participant BootstrapCmd as Bootstrap Command
    participant BootstrapSvc as Bootstrap Service
    participant ClusterSvc as Cluster Service
    participant ChartSvc as Chart Service
    participant K3d
    participant ArgoCD
    
    User->>BootstrapCmd: openframe bootstrap
    BootstrapCmd->>BootstrapSvc: Execute(cmd, args)
    
    BootstrapSvc->>User: Show deployment mode selection
    User->>BootstrapSvc: Select oss-tenant
    
    BootstrapSvc->>ClusterSvc: CreateCluster(config)
    ClusterSvc->>K3d: Create cluster
    K3d-->>ClusterSvc: Cluster created
    ClusterSvc-->>BootstrapSvc: Creation success
    
    BootstrapSvc->>ChartSvc: InstallCharts(clusterName)
    ChartSvc->>ArgoCD: Install ArgoCD
    ChartSvc->>ArgoCD: Deploy app-of-apps
    ArgoCD-->>ChartSvc: Installation complete
    ChartSvc-->>BootstrapSvc: Charts installed
    
    BootstrapSvc-->>BootstrapCmd: Bootstrap complete
    BootstrapCmd-->>User: Success message + next steps
```

#### Cluster Creation Flow

```mermaid
sequenceDiagram
    participant User
    participant ClusterCmd as Cluster Command
    participant ClusterSvc as Cluster Service
    participant UI
    participant K3d
    
    User->>ClusterCmd: openframe cluster create
    ClusterCmd->>UI: Show logo & start wizard
    UI->>User: Cluster configuration prompts
    User->>UI: Provide configuration
    UI->>ClusterSvc: Create cluster with config
    
    ClusterSvc->>ClusterSvc: Validate prerequisites
    ClusterSvc->>ClusterSvc: Generate K3d config
    ClusterSvc->>K3d: Create cluster
    K3d-->>ClusterSvc: Cluster ready
    
    ClusterSvc->>ClusterSvc: Update kubeconfig
    ClusterSvc->>UI: Show success message
    UI-->>User: Cluster ready for use
```

## Key Design Decisions

### Separation of Concerns

**Decision**: Separate CLI handling from business logic  
**Rationale**: Enables testing business logic without CLI dependencies  
**Implementation**: Service layer interfaces with clear boundaries  

### External Tool Integration

**Decision**: Wrap external tools behind interfaces  
**Rationale**: Enables mocking for testing and swapping implementations  
**Implementation**: Client interfaces for K3d, Helm, ArgoCD, etc.  

### Error Handling Strategy

**Decision**: Centralized error handling with context preservation  
**Rationale**: Consistent error messages and debugging information  
**Implementation**: Error wrapping with operation context  

### Configuration Management

**Decision**: Declarative configuration with sensible defaults  
**Rationale**: Reduce cognitive load while maintaining flexibility  
**Implementation**: Structured configuration objects with validation  

## External Dependencies Architecture

### Kubernetes Ecosystem Integration

```mermaid
graph TB
    subgraph "OpenFrame CLI"
        CLI[CLI Commands]
        Services[Service Layer]
    end
    
    subgraph "Kubernetes Tools"
        K3d[K3d<br/>Lightweight K8s]
        Helm[Helm<br/>Package Manager]
        ArgoCD[ArgoCD<br/>GitOps Controller]
        kubectl[kubectl<br/>K8s CLI]
    end
    
    subgraph "Development Tools"
        Telepresence[Telepresence<br/>Local Proxy]
        Skaffold[Skaffold<br/>Live Reload]
        Docker[Docker<br/>Container Runtime]
    end
    
    subgraph "Data Sources"
        Git[Git Repositories<br/>Charts & Config]
        Registry[Container Registry<br/>Images]
        Charts[Helm Chart Registry]
    end
    
    CLI --> Services
    Services --> K3d
    Services --> Helm
    Services --> ArgoCD
    Services --> Telepresence
    Services --> Skaffold
    
    K3d --> Docker
    Helm --> kubectl
    ArgoCD --> kubectl
    ArgoCD --> Git
    Helm --> Charts
    Skaffold --> Docker
    
    Docker --> Registry
    kubectl -.-> K3d
```

### Integration Patterns

| Tool | Integration Pattern | Purpose |
|------|-------------------|---------|
| **K3d** | Direct CLI execution | Lightweight Kubernetes cluster creation |
| **Helm** | Go SDK + CLI | Chart installation and management |
| **ArgoCD** | REST API + kubectl | GitOps application deployment |
| **Telepresence** | CLI execution | Local development proxy |
| **Skaffold** | CLI execution + config files | Live reload development workflows |

## Performance Characteristics

### Startup Time

```mermaid
graph LR
    A[CLI Parsing<br/>~10ms] --> B[Prerequisite Check<br/>~100ms]
    B --> C[Service Initialization<br/>~50ms]
    C --> D[External Tool Call<br/>~500-5000ms]
    D --> E[Result Processing<br/>~100ms]
```

**Optimization Strategies:**
- **Lazy Loading**: Initialize services only when needed
- **Prerequisite Caching**: Cache tool availability checks
- **Parallel Execution**: Run independent operations concurrently

### Memory Usage

| Component | Memory Pattern | Optimization |
|-----------|---------------|---------------|
| **CLI Layer** | Low, short-lived | Minimal object allocation |
| **Service Layer** | Medium, request-scoped | Object pooling for heavy operations |
| **External Tools** | High, tool-dependent | Process isolation, cleanup |

### Scalability Patterns

- **Single-User Focus**: Optimized for individual developer workflows
- **Stateless Operations**: No persistent state between commands  
- **Resource Cleanup**: Automatic cleanup of temporary resources
- **Concurrent Safety**: Thread-safe operations for parallel execution

## Testing Architecture

### Test Strategy by Layer

| Layer | Test Strategy | Tools |
|-------|--------------|--------|
| **CLI Layer** | Integration tests with mock services | Cobra testing, mock interfaces |
| **Service Layer** | Unit tests with mocked external dependencies | Go testing, testify mocks |
| **External Integration** | Integration tests with real tools | Docker containers, test clusters |

### Test Structure

```mermaid
graph TB
    subgraph "Unit Tests"
        ServiceTests[Service Layer Tests<br/>Mock External Tools]
        ModelTests[Model Tests<br/>Data Validation]
        UtilTests[Utility Tests<br/>Pure Functions]
    end
    
    subgraph "Integration Tests" 
        CLITests[CLI Integration Tests<br/>End-to-End Commands]
        ToolTests[External Tool Tests<br/>Real K3d/Helm/ArgoCD]
    end
    
    subgraph "E2E Tests"
        WorkflowTests[Complete Workflow Tests<br/>Bootstrap → Deploy → Cleanup]
    end
    
    ServiceTests --> CLITests
    ToolTests --> CLITests
    CLITests --> WorkflowTests
```

## Extension Points

### Adding New Commands

1. **Create command package** in `cmd/new-command/`
2. **Implement service layer** in `internal/new-command/`
3. **Add models and utilities** as needed
4. **Integrate with shared components** (UI, errors)
5. **Add to root command** registration

### Adding External Tool Support

1. **Define client interface** for the external tool
2. **Implement client** with error handling
3. **Add prerequisite checks** for tool availability
4. **Create service wrapper** for business logic
5. **Add CLI commands** that use the service

### Extending Configuration

1. **Add fields to model structs** with validation tags
2. **Update UI prompts** for new configuration options
3. **Add CLI flags** for non-interactive mode
4. **Update documentation** and help text

---

This architecture provides a solid foundation for extending OpenFrame CLI while maintaining clean separation of concerns and testability. The modular design allows for incremental feature development and easy maintenance. 🏗️