# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern Kubernetes platform bootstrapping tool that replaces shell scripts with an interactive terminal UI for managing OpenFrame Kubernetes deployments. It provides cluster management, Helm chart installation with ArgoCD, and developer workflow tools using Telepresence and Skaffold.

## Architecture

The CLI follows a layered architecture pattern with clear separation between commands, business logic, providers, and infrastructure concerns.

### High-Level Architecture Diagram

```mermaid
graph TB
    subgraph "CLI Layer"
        A[Root Command] --> B[Cluster Commands]
        A --> C[Chart Commands] 
        A --> D[Bootstrap Commands]
        A --> E[Dev Commands]
    end
    
    subgraph "Business Logic"
        B --> F[Cluster Service]
        C --> G[Chart Service]
        D --> H[Bootstrap Service]
        E --> I[Dev Services]
    end
    
    subgraph "Providers"
        F --> J[K3d Manager]
        G --> K[Helm Manager]
        G --> L[Git Repository]
        I --> M[Telepresence Provider]
        I --> N[Kubectl Provider]
    end
    
    subgraph "Infrastructure"
        J --> O[Command Executor]
        K --> O
        L --> O
        M --> O
        N --> O
        O --> P[External Tools]
    end
    
    subgraph "External Systems"
        P --> Q[Docker/K3d]
        P --> R[Helm/ArgoCD]
        P --> S[Git Repositories]
        P --> T[Kubernetes API]
    end
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **Command Layer** | `cmd/*` | CLI command definitions, argument parsing, flag management |
| **Cluster Service** | `internal/cluster` | Cluster lifecycle management (create, delete, status, cleanup) |
| **Chart Service** | `internal/chart` | Helm chart installation, ArgoCD management, repository handling |
| **Bootstrap Service** | `internal/bootstrap` | Orchestrates cluster creation + chart installation |
| **Dev Services** | `internal/dev` | Development workflows (Telepresence intercepts, Skaffold) |
| **K3d Manager** | `internal/cluster/providers/k3d` | K3d cluster operations and configuration |
| **Helm Manager** | `internal/chart/providers/helm` | Helm chart lifecycle and ArgoCD installation |
| **Command Executor** | `internal/shared/executor` | Abstraction for external command execution |
| **UI Components** | `internal/shared/ui` | Interactive prompts, tables, progress indicators |
| **Prerequisites** | `internal/*/prerequisites` | Tool validation and auto-installation |

## Component Relationships

### Service Dependencies Diagram

```mermaid
graph TD
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
        K3D[K3d Manager]
        HELM[Helm Manager]
        GIT[Git Repository]
        TP[Telepresence Provider]
        KUBECTL[Kubectl Provider]
    end
    
    subgraph "Infrastructure"
        EXEC[Command Executor]
        UI[UI Components]
        PREREQ[Prerequisites]
    end
    
    CMD --> CS
    CMD --> CHS
    CMD --> BS
    CMD --> DS
    
    CS --> K3D
    CHS --> HELM
    CHS --> GIT
    DS --> TP
    DS --> KUBECTL
    BS --> CS
    BS --> CHS
    
    K3D --> EXEC
    HELM --> EXEC
    GIT --> EXEC
    TP --> EXEC
    KUBECTL --> EXEC
    
    CS --> UI
    CHS --> UI
    DS --> UI
    
    CS --> PREREQ
    CHS --> PREREQ
    DS --> PREREQ
```

## Data Flow

### Bootstrap Workflow Sequence

```mermaid
sequenceDiagram
    participant User
    participant Bootstrap
    participant Cluster
    participant Chart
    participant K3d
    participant Helm
    participant ArgoCD
    
    User->>Bootstrap: bootstrap --deployment-mode=oss-tenant
    Bootstrap->>Cluster: CreateCluster(config)
    Cluster->>K3d: CreateCluster(k3d-config)
    K3d->>K3d: Generate cluster config
    K3d-->>Cluster: Cluster created
    Cluster-->>Bootstrap: Success
    
    Bootstrap->>Chart: InstallCharts(cluster, mode)
    Chart->>Helm: InstallArgoCD()
    Helm-->>Chart: ArgoCD installed
    Chart->>Chart: Generate helm-values.yaml
    Chart->>Helm: InstallAppOfApps(values)
    Helm-->>Chart: App-of-apps installed
    
    Chart->>ArgoCD: WaitForApplications()
    loop Application Sync
        ArgoCD->>ArgoCD: Check app status
        ArgoCD->>ArgoCD: Sync applications
    end
    ArgoCD-->>Chart: All apps healthy
    Chart-->>Bootstrap: Charts installed
    Bootstrap-->>User: Bootstrap complete
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Application entry point, CLI execution |
| `cmd/root.go` | Root command definition, global flags, version info |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/chart/services/chart_service.go` | Chart installation orchestration |
| `internal/bootstrap/service.go` | Bootstrap workflow coordination |
| `internal/cluster/providers/k3d/manager.go` | K3d cluster operations implementation |
| `internal/chart/providers/helm/manager.go` | Helm chart management implementation |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/shared/ui/logo.go` | CLI branding and visual presentation |
| `internal/cluster/models/cluster.go` | Core domain models and types |

## Dependencies

The project leverages several key external libraries:

- **Cobra**: CLI framework for command structure and argument parsing
- **pterm**: Terminal UI components for interactive prompts and styled output  
- **promptui**: User input prompts and selections
- **Viper**: Configuration management (inherited from Cobra)
- **testify**: Testing assertions and mocking framework
- **golang.org/x/term**: Terminal control for interactive features

The architecture minimizes external dependencies by using interfaces and dependency injection, making the codebase testable and maintainable.

## CLI Commands

| Command | Description | Example |
|---------|-------------|---------|
| `cluster create` | Create a new K3d cluster | `openframe cluster create my-cluster` |
| `cluster list` | List all managed clusters | `openframe cluster list` |
| `cluster status` | Show cluster details and health | `openframe cluster status my-cluster` |
| `cluster delete` | Remove a cluster and resources | `openframe cluster delete my-cluster` |
| `cluster cleanup` | Clean up unused resources | `openframe cluster cleanup my-cluster` |
| `chart install` | Install ArgoCD and app-of-apps | `openframe chart install my-cluster` |
| `bootstrap` | Full environment setup | `openframe bootstrap --deployment-mode=oss-tenant` |
| `dev intercept` | Telepresence traffic interception | `openframe dev intercept my-service --port 8080` |
| `dev skaffold` | Development with live reloading | `openframe dev skaffold my-cluster` |

The CLI supports both interactive wizard-style flows for new users and flag-based operation for automation and power users. All commands include comprehensive help text and validation.
