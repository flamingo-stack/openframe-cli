# OpenFrame CLI Architecture Overview

This document provides a comprehensive technical overview of the OpenFrame CLI architecture, designed for engineers working on or integrating with the system.

## High-Level Architecture

OpenFrame CLI follows a layered architecture with clear separation of concerns, built around the Cobra command framework with modular service layers.

```mermaid
graph TB
    subgraph "CLI Layer"
        CMD[Cobra Commands]
        BOOT[Bootstrap Cmd]
        CLUSTER[Cluster Cmds]
        CHART[Chart Cmds] 
        DEV[Dev Cmds]
    end
    
    subgraph "UI Layer"
        LOGO[Logo Display]
        PROMPTS[Interactive Prompts]
        PROGRESS[Progress Indicators]
        CONFIG[Configuration Wizards]
    end
    
    subgraph "Service Layer"
        BS[Bootstrap Service]
        CS[Cluster Service]
        CHS[Chart Service]
        DS[Dev Service]
        PREREQ[Prerequisites Service]
    end
    
    subgraph "Models Layer"
        MODELS[Data Models]
        TYPES[Type Definitions]
        VALID[Validation Logic]
        CONFIG_M[Configuration Structs]
    end
    
    subgraph "Utils Layer"
        ERROR[Error Handling]
        SHARED[Shared Utilities]
        FLAGS[Global Flags]
        WRAP[Command Wrappers]
    end
    
    subgraph "External Tools"
        K3D[K3d CLI]
        HELM[Helm CLI]
        KUBECTL[kubectl]
        ARGO[ArgoCD]
        TELE[Telepresence]
        SKAF[Skaffold]
    end
    
    CMD --> UI
    CMD --> SERVICE[Service Layer]
    CMD --> MODELS
    
    BOOT --> BS
    CLUSTER --> CS
    CHART --> CHS
    DEV --> DS
    
    SERVICE --> UTILS[Utils Layer]
    SERVICE --> External[External Tools]
    
    UI --> MODELS
    MODELS --> UTILS
    
    classDef cli fill:#e3f2fd
    classDef service fill:#f3e5f5
    classDef external fill:#fff3e0
    
    class CMD,BOOT,CLUSTER,CHART,DEV cli
    class BS,CS,CHS,DS,PREREQ service
    class K3D,HELM,KUBECTL,ARGO,TELE,SKAF external
```

## Core Components and Responsibilities

| Component | Package Path | Primary Responsibilities |
|-----------|--------------|-------------------------|
| **Bootstrap Orchestration** | `cmd/bootstrap/`, `internal/bootstrap/` | End-to-end environment setup, coordinates cluster creation and chart installation |
| **Cluster Management** | `cmd/cluster/`, `internal/cluster/` | K3d cluster lifecycle: create, delete, list, status, cleanup operations |
| **Chart Management** | `cmd/chart/`, `internal/chart/` | ArgoCD installation, Helm chart deployment, GitOps application management |
| **Development Tools** | `cmd/dev/`, `internal/dev/` | Telepresence traffic interception, Skaffold live reloading workflows |
| **Prerequisites Validation** | `internal/*/prerequisites/` | Tool availability checks, version validation, installation guidance |
| **Interactive UI** | `internal/*/ui/` | Configuration wizards, progress displays, user prompts and experience |
| **Service Layer** | `internal/*/services/` | Business logic, external tool integration, core functionality implementation |
| **Shared Components** | `internal/shared/` | Common utilities, error handling, UI components, global flag management |

## Component Interaction Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Command
    participant UI as UI Layer
    participant Svc as Service Layer
    participant Prereq as Prerequisites
    participant Tool as External Tool
    participant K8s as Kubernetes API
    
    Note over User,K8s: Bootstrap Command Flow
    
    User->>CLI: openframe bootstrap
    CLI->>UI: Display logo & collect config
    UI->>User: Show deployment options
    User->>UI: Select oss-tenant
    
    CLI->>Prereq: Check prerequisites
    Prereq->>Tool: Validate k3d, helm, kubectl
    Tool->>Prereq: Tool versions
    Prereq->>CLI: Prerequisites OK
    
    CLI->>Svc: Bootstrap.Execute()
    
    Note over Svc,K8s: Cluster Creation Phase
    
    Svc->>Svc: ClusterService.Create()
    Svc->>Tool: k3d cluster create
    Tool->>K8s: Create cluster
    K8s->>Tool: Cluster ready
    Tool->>Svc: Creation complete
    Svc->>UI: Update progress
    
    Note over Svc,K8s: Chart Installation Phase
    
    Svc->>Svc: ChartService.Install()
    Svc->>Tool: helm install argocd
    Tool->>K8s: Deploy ArgoCD
    K8s->>Tool: ArgoCD running
    
    Svc->>Tool: kubectl apply app-of-apps
    Tool->>K8s: Deploy applications
    K8s->>Tool: Apps deploying
    Tool->>Svc: Installation complete
    
    Svc->>UI: Show final status
    UI->>User: Environment ready
```

## Data Flow Architecture

The CLI follows a unidirectional data flow pattern with clear separation between user input, business logic, and external system integration.

```mermaid
graph LR
    subgraph "Input Layer"
        ARGS[Command Arguments]
        FLAGS[Command Flags]
        PROMPTS[User Prompts]
        CONFIG[Config Files]
    end
    
    subgraph "Processing Layer" 
        VALIDATE[Input Validation]
        TRANSFORM[Data Transformation]
        ORCHESTRATE[Service Orchestration]
    end
    
    subgraph "Output Layer"
        K8S[Kubernetes Resources]
        LOGS[Progress Logs]
        ERRORS[Error Messages]
        STATUS[Status Reports]
    end
    
    ARGS --> VALIDATE
    FLAGS --> VALIDATE
    PROMPTS --> VALIDATE
    CONFIG --> VALIDATE
    
    VALIDATE --> TRANSFORM
    TRANSFORM --> ORCHESTRATE
    
    ORCHESTRATE --> K8S
    ORCHESTRATE --> LOGS
    ORCHESTRATE --> ERRORS
    ORCHESTRATE --> STATUS
    
    classDef input fill:#e8f5e8
    classDef processing fill:#fff3e0
    classDef output fill:#f3e5f5
    
    class ARGS,FLAGS,PROMPTS,CONFIG input
    class VALIDATE,TRANSFORM,ORCHESTRATE processing
    class K8S,LOGS,ERRORS,STATUS output
```

## Design Patterns and Principles

### 1. Command Pattern Implementation

Each CLI command follows a consistent structure:

```go
// Command structure pattern
func GetCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:     "command [args]",
        Short:   "Brief description", 
        Long:    "Detailed description with examples",
        PreRunE: validateInputs,           // Input validation
        RunE:    utils.WrapCommandWithCommonSetup(execute), // Business logic
    }
    
    // Add flags
    addCommandFlags(cmd)
    return cmd
}

// Business logic is delegated to services
func execute(cmd *cobra.Command, args []string) error {
    service := getCommandService()
    return service.Execute(cmd, args)
}
```

### 2. Service Layer Pattern

Services encapsulate business logic and external tool integration:

```go
type ClusterService interface {
    Create(config ClusterConfig) error
    Delete(name string) error
    List() ([]ClusterInfo, error)
    GetStatus(name string) (*ClusterStatus, error)
    Cleanup() error
}

// Implementation handles external tool integration
type clusterService struct {
    k3dClient   K3dClient
    kubectlClient KubectlClient
    ui          UIHandler
}
```

### 3. UI Abstraction Pattern

User interface components are abstracted to enable testing and consistent experience:

```go
type UIHandler interface {
    ShowLogo()
    PromptSelect(message string, options []string) (string, error)
    ShowProgress(message string)
    ShowSuccess(message string)
    ShowError(error)
}

// Implementations can be interactive, quiet, or test-friendly
type InteractiveUI struct{}
type QuietUI struct{}
type TestUI struct{}
```

### 4. Prerequisites Validation Pattern

All commands validate prerequisites before execution:

```go
type PrerequisiteChecker interface {
    CheckAll() error
    CheckTool(name string) error
    InstallIfMissing(tool string) error
}

// Each command group has specific prerequisites
var ClusterPrerequisites = []string{"docker", "k3d", "kubectl"}
var ChartPrerequisites = []string{"helm", "kubectl"}
var DevPrerequisites = []string{"telepresence", "skaffold"}
```

## Module Dependencies and Relationships

### Dependency Graph

```mermaid
graph TD
    subgraph "Command Modules"
        BOOTSTRAP[Bootstrap]
        CLUSTER[Cluster] 
        CHART[Chart]
        DEV[Dev]
    end
    
    subgraph "Service Modules"
        BOOTSTRAP_SVC[Bootstrap Service]
        CLUSTER_SVC[Cluster Service]
        CHART_SVC[Chart Service]
        DEV_SVC[Dev Service]
    end
    
    subgraph "Shared Modules"
        UI[Shared UI]
        ERROR[Error Handling]
        UTILS[Utilities]
        PREREQ[Prerequisites]
    end
    
    subgraph "External Dependencies"
        COBRA[Cobra CLI]
        K8S[Kubernetes Client]
        DOCKER[Docker SDK]
        YAML[YAML Parser]
    end
    
    BOOTSTRAP --> BOOTSTRAP_SVC
    CLUSTER --> CLUSTER_SVC
    CHART --> CHART_SVC
    DEV --> DEV_SVC
    
    BOOTSTRAP_SVC --> CLUSTER_SVC
    BOOTSTRAP_SVC --> CHART_SVC
    
    CLUSTER_SVC --> UI
    CHART_SVC --> UI
    DEV_SVC --> UI
    BOOTSTRAP_SVC --> UI
    
    CLUSTER_SVC --> ERROR
    CHART_SVC --> ERROR
    DEV_SVC --> ERROR
    BOOTSTRAP_SVC --> ERROR
    
    CLUSTER_SVC --> PREREQ
    CHART_SVC --> PREREQ
    DEV_SVC --> PREREQ
    
    UI --> UTILS
    ERROR --> UTILS
    PREREQ --> UTILS
    
    BOOTSTRAP --> COBRA
    CLUSTER --> COBRA
    CHART --> COBRA
    DEV --> COBRA
    
    CLUSTER_SVC --> K8S
    CHART_SVC --> K8S
    CLUSTER_SVC --> DOCKER
    
    classDef command fill:#e3f2fd
    classDef service fill:#f3e5f5
    classDef shared fill:#fff3e0
    classDef external fill:#ffebee
    
    class BOOTSTRAP,CLUSTER,CHART,DEV command
    class BOOTSTRAP_SVC,CLUSTER_SVC,CHART_SVC,DEV_SVC service
    class UI,ERROR,UTILS,PREREQ shared
    class COBRA,K8S,DOCKER,YAML external
```

### Import Hierarchy Rules

| Level | Modules | Can Import From |
|-------|---------|-----------------|
| **Commands** | `cmd/*` | Internal services, Cobra, shared utilities |
| **Services** | `internal/*/services/` | Models, shared components, external clients |
| **UI** | `internal/*/ui/` | Models, shared utilities, no external tools |
| **Models** | `internal/*/models/` | Shared utilities, validation libraries |
| **Shared** | `internal/shared/` | Standard library, common utilities only |

### Anti-Patterns to Avoid

❌ **Circular Dependencies**: Services should not import command packages  
❌ **UI in Services**: Services should use UI interfaces, not concrete implementations  
❌ **Tool Logic in Commands**: External tool integration belongs in services  
❌ **Shared State**: Avoid global variables; use dependency injection  

## Configuration Management

### Configuration Sources Priority

1. **Command Line Flags** (highest priority)
2. **Environment Variables**
3. **Configuration Files** (`helm-values.yaml`, cluster configs)
4. **Interactive Prompts**
5. **Default Values** (lowest priority)

### Configuration Flow

```mermaid
graph LR
    FLAGS[CLI Flags] --> MERGE[Configuration Merger]
    ENV[Environment Variables] --> MERGE
    FILES[Config Files] --> MERGE
    PROMPTS[User Prompts] --> MERGE
    DEFAULTS[Default Values] --> MERGE
    
    MERGE --> VALIDATE[Validation]
    VALIDATE --> CONFIG[Final Configuration]
    
    CONFIG --> SERVICES[Service Layer]
    
    classDef input fill:#e8f5e8
    classDef process fill:#fff3e0
    classDef output fill:#f3e5f5
    
    class FLAGS,ENV,FILES,PROMPTS,DEFAULTS input
    class MERGE,VALIDATE process
    class CONFIG,SERVICES output
```

## Error Handling Strategy

### Error Categories and Handling

| Error Type | Example | Handling Strategy |
|------------|---------|-------------------|
| **User Input Errors** | Invalid cluster name | User-friendly message, retry option |
| **Prerequisite Errors** | Missing Docker | Installation guidance, clear instructions |
| **External Tool Errors** | K3d creation fails | Wrapped error with context, troubleshooting tips |
| **Network Errors** | Helm repo unreachable | Retry logic, fallback options |
| **System Errors** | Out of disk space | Clear diagnosis, cleanup suggestions |

### Error Wrapping Pattern

```go
// Service layer - add context
func (s *ClusterService) CreateCluster(name string) error {
    if err := s.validatePrerequisites(); err != nil {
        return fmt.Errorf("prerequisites check failed for cluster %s: %w", name, err)
    }
    
    if err := s.k3d.CreateCluster(name); err != nil {
        return fmt.Errorf("k3d cluster creation failed: %w", err)
    }
    
    return nil
}

// Command layer - user-friendly messages
func runCreateCommand(cmd *cobra.Command, args []string) error {
    if err := service.CreateCluster(name); err != nil {
        return fmt.Errorf("❌ Failed to create cluster '%s'.\n\n%v\n\n💡 Try: openframe cluster cleanup", name, err)
    }
    return nil
}
```

## Performance Considerations

### Optimization Strategies

| Component | Optimization | Impact |
|-----------|--------------|--------|
| **Prerequisites** | Parallel checking | 50% faster startup |
| **UI Updates** | Buffered progress | Smoother experience |
| **K8s Operations** | Client reuse | Reduced connection overhead |
| **Build Process** | Go modules caching | Faster CI/CD |

### Resource Management

- **Memory**: Minimal allocation in CLI operations, cleanup after external tool execution
- **CPU**: Parallel operations where safe, avoid blocking on I/O
- **Network**: Connection pooling for Kubernetes API calls, retry with backoff
- **Disk**: Temporary file cleanup, configurable cache directories

## Testing Architecture

### Test Strategy Overview

```mermaid
graph TB
    subgraph "Test Pyramid"
        E2E[End-to-End Tests<br/>Full workflows]
        INTEGRATION[Integration Tests<br/>Service + External tools]
        UNIT[Unit Tests<br/>Individual functions]
    end
    
    subgraph "Test Types"
        MOCK[Mock Tests<br/>External dependencies]
        CONTRACT[Contract Tests<br/>External tool interfaces] 
        PERFORMANCE[Performance Tests<br/>Large clusters]
    end
    
    UNIT --> INTEGRATION
    INTEGRATION --> E2E
    
    UNIT -.-> MOCK
    INTEGRATION -.-> CONTRACT
    E2E -.-> PERFORMANCE
    
    classDef test fill:#e8f5e8
    class E2E,INTEGRATION,UNIT,MOCK,CONTRACT,PERFORMANCE test
```

### Test Organization

| Test Level | Location | Purpose | Dependencies |
|------------|----------|---------|--------------|
| **Unit** | `*_test.go` | Function-level testing | None |
| **Integration** | `tests/integration/` | Service + tool interaction | Docker, K3d |
| **End-to-End** | `tests/e2e/` | Full command workflows | Full tool stack |
| **Performance** | `tests/perf/` | Large-scale scenarios | Dedicated test clusters |

---

## Next Steps for Contributors

1. **Study the service layer** in `internal/*/services/` to understand business logic
2. **Review UI patterns** in `internal/*/ui/` for consistent user experience
3. **Understand error handling** in `internal/shared/errors/` for proper error management
4. **Explore external integrations** to understand tool orchestration patterns

**Questions or need clarification?** Check the [Developer Getting Started Guide](getting-started-dev.md) or reach out to the development team.