# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing Kubernetes clusters and development workflows. It provides interactive wizards for cluster creation, chart installation with ArgoCD, and development tools like Telepresence intercepts and Skaffold workflows, replacing shell scripts with a robust Go-based CLI.

## Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        A[Root Command] --> B[Cluster Commands]
        A --> C[Chart Commands]
        A --> D[Bootstrap Commands]
        A --> E[Dev Commands]
    end
    
    subgraph "Business Logic"
        F[Cluster Service] --> G[K3d Provider]
        H[Chart Service] --> I[ArgoCD Provider]
        H --> J[Helm Provider]
        K[Bootstrap Service] --> F
        K --> H
        L[Dev Services] --> M[Telepresence Provider]
        L --> N[Skaffold Provider]
    end
    
    subgraph "Infrastructure"
        O[Command Executor] --> P[External Tools]
        Q[UI Components] --> R[Progress Tracking]
        S[Configuration] --> T[File Management]
    end
    
    B --> F
    C --> H
    D --> K
    E --> L
    F --> O
    H --> O
    L --> O
    
    P --> U[k3d/kubectl/helm/telepresence]
    
    style A fill:#e1f5fe
    style F fill:#f3e5f5
    style H fill:#f3e5f5
    style O fill:#fff3e0
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **CLI Commands** | `cmd/*` | Command definitions, flag parsing, user interaction |
| **Cluster Service** | `internal/cluster` | Cluster lifecycle management (create, delete, list, status) |
| **Chart Service** | `internal/chart` | Helm chart and ArgoCD installation management |
| **Bootstrap Service** | `internal/bootstrap` | Orchestrates cluster creation + chart installation |
| **Dev Services** | `internal/dev` | Development workflow tools (intercept, scaffold) |
| **Command Executor** | `internal/shared/executor` | Abstraction for external command execution |
| **UI Components** | `internal/shared/ui` | Interactive prompts, progress indicators, display |
| **K3d Provider** | `internal/cluster/providers/k3d` | K3d-specific cluster operations |
| **ArgoCD Provider** | `internal/chart/providers/argocd` | ArgoCD installation and application sync |
| **Telepresence Provider** | `internal/dev/providers/telepresence` | Traffic interception for development |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CMD[CLI Commands]
    end
    
    subgraph "Service Layer"
        CS[Cluster Service]
        CHS[Chart Service]
        BS[Bootstrap Service]
        DS[Dev Services]
    end
    
    subgraph "Provider Layer"
        K3D[K3d Provider]
        ARGO[ArgoCD Provider]
        HELM[Helm Provider]
        TELE[Telepresence Provider]
        KUBE[Kubectl Provider]
    end
    
    subgraph "Shared Infrastructure"
        EXEC[Command Executor]
        UI[UI Components]
        CONFIG[Configuration]
        ERR[Error Handling]
    end
    
    CMD --> CS
    CMD --> CHS
    CMD --> BS
    CMD --> DS
    
    CS --> K3D
    CHS --> ARGO
    CHS --> HELM
    BS --> CS
    BS --> CHS
    DS --> TELE
    DS --> KUBE
    
    K3D --> EXEC
    ARGO --> EXEC
    HELM --> EXEC
    TELE --> EXEC
    
    CS --> UI
    CHS --> UI
    DS --> UI
    
    CS --> CONFIG
    CHS --> CONFIG
    
    CS --> ERR
    CHS --> ERR
    DS --> ERR
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant ClusterService
    participant K3dProvider
    participant ChartService
    participant ArgoCDProvider
    participant CommandExecutor
    participant ExternalTools
    
    User->>CLI: openframe bootstrap
    CLI->>ClusterService: CreateCluster(config)
    ClusterService->>K3dProvider: CreateCluster(config)
    K3dProvider->>CommandExecutor: Execute("k3d", args...)
    CommandExecutor->>ExternalTools: k3d cluster create
    ExternalTools-->>CommandExecutor: Success
    CommandExecutor-->>K3dProvider: Result
    K3dProvider-->>ClusterService: Success
    ClusterService-->>CLI: Cluster Created
    
    CLI->>ChartService: InstallCharts(config)
    ChartService->>ArgoCDProvider: InstallArgoCD()
    ArgoCDProvider->>CommandExecutor: Execute("helm", args...)
    CommandExecutor->>ExternalTools: helm install argocd
    ExternalTools-->>CommandExecutor: Success
    CommandExecutor-->>ArgoCDProvider: Result
    ArgoCDProvider->>ArgoCDProvider: WaitForApplications()
    ArgoCDProvider-->>ChartService: ArgoCD Ready
    ChartService-->>CLI: Charts Installed
    CLI-->>User: Bootstrap Complete
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Application entry point |
| `cmd/root.go` | Root command definition with version info |
| `cmd/bootstrap/bootstrap.go` | Bootstrap command combining cluster + chart operations |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/chart/services/chart_service.go` | Chart installation orchestration |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/cluster/providers/k3d/manager.go` | K3d cluster provider implementation |
| `internal/chart/providers/argocd/wait.go` | ArgoCD application synchronization logic |
| `internal/shared/ui/logo.go` | CLI branding and visual elements |
| `internal/dev/services/intercept/service.go` | Telepresence traffic interception |

## Dependencies

The project leverages several key external libraries:

- **Cobra** (`spf13/cobra`): CLI framework for command structure and flag parsing
- **PTerm** (`pterm/pterm`): Terminal UI components for progress indicators and interactive elements
- **PromptUI** (`manifoldco/promptui`): Interactive prompts for user input
- **YAML** (`gopkg.in/yaml.v3`): Configuration file parsing for Helm values
- **Testify** (`stretchr/testify`): Testing framework for unit and integration tests

The CLI acts as an orchestrator that shells out to external Kubernetes tools:
- **k3d**: Local Kubernetes cluster management
- **kubectl**: Kubernetes API interaction
- **helm**: Chart installation and management
- **telepresence**: Development traffic interception
- **skaffold**: Development workflow automation

## CLI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `openframe cluster create` | Create a new Kubernetes cluster | `openframe cluster create my-cluster` |
| `openframe cluster list` | List all managed clusters | `openframe cluster list` |
| `openframe cluster status` | Show detailed cluster information | `openframe cluster status my-cluster` |
| `openframe cluster delete` | Remove a cluster | `openframe cluster delete my-cluster` |
| `openframe chart install` | Install ArgoCD and charts | `openframe chart install my-cluster` |
| `openframe bootstrap` | Complete cluster + chart setup | `openframe bootstrap --deployment-mode=oss-tenant` |
| `openframe dev intercept` | Intercept service traffic | `openframe dev intercept my-service --port 8080` |
| `openframe dev skaffold` | Run development workflow | `openframe dev skaffold my-cluster` |

The CLI provides both interactive wizards and flag-based automation, supporting both developer workflows and CI/CD integration.
