# OpenFrame CLI Architecture Overview

This document provides a comprehensive technical overview of the OpenFrame CLI architecture, including core components, design patterns, and data flow diagrams.

## High-Level Architecture

OpenFrame CLI follows a modular command-driven architecture built on the Cobra CLI framework, with clear separation between presentation logic, business services, and external tool integrations.

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[OpenFrame CLI Entry]
        CMD[Command Router]
        FLAGS[Flag Parser]
    end
    
    subgraph "Command Layer"
        BOOTSTRAP[Bootstrap Cmd]
        CLUSTER[Cluster Cmd]
        CHART[Chart Cmd]
        DEV[Dev Cmd]
    end
    
    subgraph "Service Layer"
        BS[Bootstrap Service]
        CS[Cluster Service]
        CHS[Chart Service]
        DS[Dev Service]
    end
    
    subgraph "Infrastructure Layer"
        K3D[(K3d)]
        HELM[(Helm)]
        ARGOCD[(ArgoCD)]
        TELEPRESENCE[(Telepresence)]
        SKAFFOLD[(Skaffold)]
    end
    
    subgraph "Shared Components"
        UI[UI Components]
        ERR[Error Handler]
        PREREQ[Prerequisites]
    end
    
    CLI --> CMD
    CMD --> FLAGS
    FLAGS --> BOOTSTRAP
    FLAGS --> CLUSTER
    FLAGS --> CHART
    FLAGS --> DEV
    
    BOOTSTRAP --> BS
    CLUSTER --> CS
    CHART --> CHS
    DEV --> DS
    
    BS --> CS
    BS --> CHS
    CS --> K3D
    CHS --> HELM
    CHS --> ARGOCD
    DS --> TELEPRESENCE
    DS --> SKAFFOLD
    
    CS --> UI
    CHS --> UI
    BS --> UI
    DS --> UI
    
    CS --> ERR
    CHS --> ERR
    BS --> ERR
    DS --> ERR
    
    CS --> PREREQ
    CHS --> PREREQ
    DS --> PREREQ
    
    style CLI fill:#e1f5fe
    style CMD fill:#e8f5e8
    style BS fill:#fff3e0
    style UI fill:#f3e5f5
```

## Core Components

### Command Layer Components

| Component | Package | Responsibilities | Dependencies |
|-----------|---------|------------------|--------------|
| **Bootstrap Command** | `cmd/bootstrap/` | Orchestrates complete environment setup | Cluster Service, Chart Service |
| **Cluster Commands** | `cmd/cluster/` | K3d cluster lifecycle management | Cluster Service, K3d |
| **Chart Commands** | `cmd/chart/` | Helm chart and ArgoCD management | Chart Service, Helm, ArgoCD |
| **Dev Commands** | `cmd/dev/` | Development workflow tools | Dev Service, Telepresence, Skaffold |

### Service Layer Components

| Service | Package | Core Functions | External Tools |
|---------|---------|----------------|----------------|
| **Bootstrap Service** | `internal/bootstrap/` | End-to-end environment orchestration | Delegates to Cluster + Chart services |
| **Cluster Service** | `internal/cluster/` | K3d cluster CRUD operations, status monitoring | K3d, Docker |
| **Chart Service** | `internal/chart/` | ArgoCD installation, app-of-apps deployment | Helm, ArgoCD, Git repositories |
| **Dev Service** | `internal/dev/` | Traffic interception, live reload workflows | Telepresence, Skaffold |

### Shared Components

| Component | Package | Purpose | Usage |
|-----------|---------|---------|-------|
| **UI Components** | `internal/shared/ui/` | Interactive menus, progress indicators, logo display | All commands for user interaction |
| **Error Handler** | `internal/shared/errors/` | Centralized error formatting and logging | All services for consistent error handling |
| **Prerequisites** | `*/prerequisites/` | Tool validation and installation | Per-module validation of external dependencies |

## Detailed Component Architecture

### 1. Command Layer Design

The command layer uses the Cobra framework with a hierarchical structure:

```mermaid
graph LR
    ROOT[openframe] --> BOOTSTRAP[bootstrap]
    ROOT --> CLUSTER[cluster]
    ROOT --> CHART[chart]
    ROOT --> DEV[dev]
    
    CLUSTER --> CREATE[create]
    CLUSTER --> DELETE[delete]
    CLUSTER --> LIST[list]
    CLUSTER --> STATUS[status]
    CLUSTER --> CLEANUP[cleanup]
    
    CHART --> INSTALL[install]
    
    DEV --> INTERCEPT[intercept]
    DEV --> SKAFFOLD[skaffold]
    
    style ROOT fill:#e1f5fe
    style CLUSTER fill:#e8f5e8
    style CHART fill:#fff3e0
    style DEV fill:#f3e5f5
```

#### Flag Management Pattern
```go
// Global flags are managed through utils package
type GlobalFlags struct {
    Create *CreateFlags
    Verbose bool
    DryRun bool
}

// Each command adds specific flags
func addCreateFlags(cmd *cobra.Command, flags *CreateFlags) {
    cmd.Flags().IntVar(&flags.Nodes, "nodes", 1, "Number of nodes")
    cmd.Flags().StringVar(&flags.Type, "type", "k3d", "Cluster type")
}
```

### 2. Service Layer Architecture

Services implement business logic and coordinate external tool interactions:

```mermaid
sequenceDiagram
    participant CMD as Command Layer
    participant SVC as Service Layer  
    participant UI as UI Components
    participant EXT as External Tools
    
    CMD->>SVC: Execute(config)
    SVC->>UI: ShowProgress("Starting...")
    SVC->>EXT: ValidatePrerequisites()
    EXT-->>SVC: Prerequisites OK
    SVC->>EXT: ExecuteOperation()
    EXT-->>SVC: Operation Result
    SVC->>UI: ShowSuccess("Complete")
    SVC-->>CMD: Return Result
```

#### Service Interface Pattern
```go
// Common interface pattern for services
type ClusterService interface {
    Create(ctx context.Context, config ClusterConfig) error
    Delete(ctx context.Context, name string) error
    List(ctx context.Context) ([]Cluster, error)
    Status(ctx context.Context, name string) (*ClusterStatus, error)
}

// Implementation with dependency injection
type service struct {
    ui     ui.Handler
    k3d    k3d.Client
    errors errors.Handler
}
```

### 3. Data Flow Patterns

#### Bootstrap Flow (Complete Environment Setup)

```mermaid
sequenceDiagram
    participant User
    participant Bootstrap
    participant ClusterSvc
    participant ChartSvc
    participant K3d
    participant ArgoCD
    
    User->>Bootstrap: openframe bootstrap my-cluster
    Bootstrap->>Bootstrap: Parse flags & validate
    Bootstrap->>ClusterSvc: CreateCluster(config)
    
    ClusterSvc->>K3d: Create cluster
    K3d-->>ClusterSvc: Cluster ready
    ClusterSvc-->>Bootstrap: Cluster created
    
    Bootstrap->>ChartSvc: InstallCharts(cluster)
    ChartSvc->>ChartSvc: Generate certificates
    ChartSvc->>K3d: Install ArgoCD via Helm
    K3d-->>ChartSvc: ArgoCD installed
    
    ChartSvc->>ArgoCD: Deploy app-of-apps
    ArgoCD->>ArgoCD: Sync applications
    ArgoCD-->>ChartSvc: Apps synced
    
    ChartSvc-->>Bootstrap: Charts installed
    Bootstrap-->>User: Environment ready
```

#### Development Workflow Integration

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant DevSvc as Dev Service
    participant Telepresence
    participant Cluster
    participant LocalApp as Local App
    
    Dev->>DevSvc: openframe dev intercept my-service
    DevSvc->>Telepresence: Setup intercept
    Telepresence->>Cluster: Modify service routing
    Cluster-->>Telepresence: Routing configured
    Telepresence-->>DevSvc: Intercept active
    
    DevSvc->>LocalApp: Start port forwarding
    LocalApp-->>DevSvc: App listening locally
    DevSvc-->>Dev: Ready for development
    
    Note over Dev,LocalApp: Traffic to my-service in cluster<br/>now routes to local development app
```

## Design Patterns and Principles

### 1. Command Pattern Implementation
Each CLI command follows the command pattern with clear separation:
- **Command**: Defines interface and flags (`cmd/` packages)
- **Invoker**: Cobra framework handles execution
- **Receiver**: Service layer implements business logic (`internal/` packages)

### 2. Dependency Injection
Services use constructor injection for testability:
```go
func NewClusterService(ui ui.Handler, k3d k3d.Client) ClusterService {
    return &service{
        ui:  ui,
        k3d: k3d,
    }
}
```

### 3. Interface Segregation
Interfaces are focused and role-specific:
```go
// UI interfaces are action-specific
type ProgressReporter interface {
    StartProgress(message string)
    UpdateProgress(percentage int)
    StopProgress(success bool)
}

type UserPrompter interface {
    PromptSelect(options []string) (int, error)
    PromptConfirm(message string) bool
}
```

### 4. Error Handling Strategy
Centralized error handling with context preservation:
```go
// Errors include context and suggested actions
type OpenFrameError struct {
    Operation string
    Cause     error
    Suggestion string
}

func (e *OpenFrameError) Error() string {
    return fmt.Sprintf("%s failed: %v\nSuggestion: %s", 
        e.Operation, e.Cause, e.Suggestion)
}
```

## Module Dependencies

### Internal Module Relationships

```mermaid
graph TD
    subgraph "Bootstrap Module"
        BS[Bootstrap Service]
    end
    
    subgraph "Cluster Module"
        CS[Cluster Service]
        CUI[Cluster UI]
        CUTILS[Cluster Utils]
        CMODELS[Cluster Models]
    end
    
    subgraph "Chart Module"
        CHS[Chart Service]
        CHTYPES[Chart Types]
        CHPREREQ[Chart Prerequisites]
    end
    
    subgraph "Dev Module"
        DS[Dev Service]
        DMODELS[Dev Models]
        DPREREQ[Dev Prerequisites]
    end
    
    subgraph "Shared Modules"
        UI[Shared UI]
        ERR[Shared Errors]
    end
    
    BS --> CS
    BS --> CHS
    
    CS --> CUI
    CS --> CUTILS
    CS --> CMODELS
    CS --> UI
    CS --> ERR
    
    CHS --> CHTYPES
    CHS --> CHPREREQ
    CHS --> UI
    CHS --> ERR
    
    DS --> DMODELS
    DS --> DPREREQ
    DS --> UI
    DS --> ERR
    
    CUI --> UI
    
    style BS fill:#fff3e0
    style UI fill:#f3e5f5
    style ERR fill:#ffebee
```

### External Dependencies

| Category | Tools | Integration Method | Purpose |
|----------|-------|-------------------|----------|
| **Container Runtime** | Docker | Direct CLI calls | K3d cluster node management |
| **Kubernetes Distribution** | K3d | Go client library | Local cluster creation and management |
| **Package Management** | Helm | CLI integration | Chart installation and management |
| **GitOps** | ArgoCD | Kubernetes API + Web UI | Application deployment and synchronization |
| **Development Tools** | Telepresence, Skaffold | CLI integration | Local development workflows |

## Configuration Management

### Configuration Hierarchy
1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration files** (`~/.openframe/config.yaml`)
4. **Default values** (lowest priority)

### Configuration Structure
```yaml
# ~/.openframe/config.yaml
clusters:
  default_type: k3d
  default_nodes: 1
  registry_mirror: docker.io

charts:
  github_repo: https://github.com/flamingo-stack/openframe-apps
  github_branch: main
  argocd_version: 8.2.7

development:
  telepresence_timeout: 300s
  skaffold_profile: dev
```

## Performance Considerations

### Resource Management
- **Cluster Resources**: K3d clusters consume ~1GB RAM per cluster
- **Concurrent Operations**: Limited to prevent Docker daemon overload
- **Caching**: Helm charts and Docker images cached locally

### Optimization Strategies
- **Lazy Loading**: Services initialized only when needed
- **Connection Pooling**: Kubernetes clients reused across operations
- **Progress Feedback**: Long operations show progress to improve UX

## Security Architecture

### Credential Management
- **Kubernetes Contexts**: Managed through standard kubeconfig
- **Certificate Generation**: Automatic TLS cert creation for local development
- **Secret Handling**: No hardcoded secrets, environment-based configuration

### Network Security
- **Local-only**: All clusters created for local development
- **Ingress Control**: Configurable ingress controllers for service exposure
- **Network Policies**: Support for Kubernetes network policy testing

## Testing Strategy

### Test Pyramid Structure
```mermaid
graph TD
    UNIT[Unit Tests<br/>70% coverage] 
    INTEGRATION[Integration Tests<br/>Service interactions]
    E2E[End-to-End Tests<br/>Full CLI workflows]
    
    UNIT --> INTEGRATION
    INTEGRATION --> E2E
    
    style UNIT fill:#c8e6c9
    style INTEGRATION fill:#fff3e0
    style E2E fill:#ffebee
```

#### Testing Approach
- **Unit Tests**: Mock external dependencies, test business logic
- **Integration Tests**: Test service interactions with real tools
- **End-to-End Tests**: Full CLI command execution against real clusters

## Extension Points

### Adding New Commands
1. Create command package in `cmd/newcommand/`
2. Implement service in `internal/newcommand/`
3. Add to main command router
4. Include prerequisites validation
5. Add comprehensive tests

### External Tool Integration
1. Create client interface in service package
2. Implement prerequisite checking
3. Add configuration options
4. Include error handling and user feedback

---

## Summary

The OpenFrame CLI architecture provides:

✅ **Modularity**: Clear separation between commands, services, and tools  
✅ **Extensibility**: Easy to add new commands and external tool integrations  
✅ **Testability**: Interfaces and dependency injection enable comprehensive testing  
✅ **User Experience**: Consistent UI patterns and error handling across all operations  
✅ **Maintainability**: Well-defined patterns and clear module boundaries  

This architecture supports the goal of providing a unified, user-friendly interface for complex Kubernetes development workflows while maintaining code quality and developer productivity.