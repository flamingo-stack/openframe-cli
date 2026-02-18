# openframe-cli Module Documentation

# OpenFrame CLI Architecture

OpenFrame CLI is a modern, interactive command-line tool for managing OpenFrame Kubernetes clusters and development workflows, providing seamless cluster lifecycle management, chart installation with ArgoCD, and developer-friendly tools for service intercepts and scaffolding.

## Architecture Overview

The OpenFrame CLI follows a clean, layered architecture with clear separation of concerns across domain boundaries. The system is built using Go with Cobra for CLI management and integrates with external tools like K3D, Helm, ArgoCD, and Telepresence.

### High-Level Architecture

```mermaid
graph TB
    CLI[CLI Layer - Cobra Commands] --> BS[Bootstrap Service]
    CLI --> CS[Chart Service] 
    CLI --> CLS[Cluster Service]
    CLI --> DS[Dev Service]
    
    BS --> CSvc[Chart Services]
    BS --> CLSvc[Cluster Services]
    
    CSvc --> HP[Helm Provider]
    CSvc --> AP[ArgoCD Provider]
    CSvc --> GP[Git Provider]
    
    CLS --> K3D[K3D Manager]
    CLS --> UI[UI Components]
    
    DS --> TP[Telepresence Provider]
    DS --> KP[Kubectl Provider]
    DS --> SP[Skaffold Provider]
    
    HP --> Helm[Helm Binary]
    AP --> ArgoCD[ArgoCD APIs]
    GP --> Git[Git Commands]
    K3D --> K3DCmd[K3D Binary]
    TP --> TelepresenceCmd[Telepresence Binary]
    KP --> KubectlCmd[Kubectl Binary]
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **CLI Commands** | `cmd/*` | Cobra command definitions and flag management |
| **Bootstrap Service** | `internal/bootstrap` | Orchestrates cluster creation + chart installation |
| **Chart Services** | `internal/chart` | ArgoCD and app-of-apps installation with GitHub integration |
| **Cluster Services** | `internal/cluster` | K3D cluster lifecycle management |
| **Dev Services** | `internal/dev` | Telepresence intercepts and Skaffold workflows |
| **Shared Utilities** | `internal/shared` | Command execution, UI components, error handling |
| **Prerequisites** | `internal/*/prerequisites` | Tool validation and auto-installation |
| **Providers** | `internal/*/providers` | External tool integrations (K3D, Helm, ArgoCD) |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CC[Cluster Commands]
        ChC[Chart Commands]  
        DC[Dev Commands]
        BC[Bootstrap Command]
    end
    
    subgraph "Service Layer"
        CS[Cluster Service]
        ChS[Chart Service]
        DS[Dev Service]
        BS[Bootstrap Service]
    end
    
    subgraph "Provider Layer"
        K3DP[K3D Provider]
        HP[Helm Provider]
        AP[ArgoCD Provider]
        TP[Telepresence Provider]
        KP[Kubectl Provider]
    end
    
    subgraph "External Tools"
        K3D[K3D Binary]
        Helm[Helm Binary]
        ArgoCD[ArgoCD APIs]
        Telepresence[Telepresence Binary]
        Kubectl[Kubectl Binary]
    end
    
    CC --> CS
    ChC --> ChS
    DC --> DS
    BC --> BS
    BC --> CS
    BC --> ChS
    
    CS --> K3DP
    ChS --> HP
    ChS --> AP
    DS --> TP
    DS --> KP
    
    K3DP --> K3D
    HP --> Helm
    AP --> ArgoCD
    TP --> Telepresence
    KP --> Kubectl
```

## Data Flow

### Bootstrap Workflow

```mermaid
sequenceDiagram
    participant U as User
    participant CLI as Bootstrap CMD
    participant CS as Cluster Service
    participant ChS as Chart Service
    participant K3D as K3D Provider
    participant Helm as Helm Provider
    participant ArgoCD as ArgoCD Provider
    
    U->>CLI: openframe bootstrap
    CLI->>CS: CreateCluster()
    CS->>K3D: Create k3d cluster
    K3D-->>CS: *rest.Config
    CS-->>CLI: Cluster ready
    
    CLI->>ChS: InstallCharts()
    ChS->>Helm: Install ArgoCD
    Helm-->>ChS: ArgoCD deployed
    ChS->>ArgoCD: Wait for readiness
    ArgoCD-->>ChS: Ready
    ChS->>Helm: Install app-of-apps
    Helm-->>ChS: Apps deployed
    ChS->>ArgoCD: Wait for sync
    ArgoCD-->>ChS: All apps synced
    ChS-->>CLI: Installation complete
    CLI-->>U: Bootstrap successful
```

### Chart Installation Flow

```mermaid
sequenceDiagram
    participant CLI as Chart CMD
    participant ChS as Chart Service
    participant Git as Git Provider
    participant Helm as Helm Provider
    participant ArgoCD as ArgoCD Provider
    participant UI as Configuration UI
    
    CLI->>UI: Show deployment selection
    UI-->>CLI: User selects mode
    CLI->>UI: Configure helm values
    UI-->>CLI: Generated config
    
    CLI->>ChS: InstallCharts()
    ChS->>Helm: Install ArgoCD
    Helm-->>ChS: ArgoCD ready
    
    ChS->>Git: Clone repository
    Git-->>ChS: Local chart path
    ChS->>Helm: Install app-of-apps
    Helm-->>ChS: Apps installed
    
    ChS->>ArgoCD: Wait for applications
    loop Application Sync
        ArgoCD->>ArgoCD: Sync applications
        ArgoCD-->>ChS: Status update
    end
    ArgoCD-->>ChS: All healthy & synced
    ChS-->>CLI: Success
```

## Key Files

| File Path | Purpose |
|-----------|---------|
| `main.go` | Application entry point |
| `cmd/root.go` | Root Cobra command with version info and global setup |
| `cmd/bootstrap/bootstrap.go` | Bootstrap command orchestrating cluster + chart setup |
| `internal/bootstrap/service.go` | Bootstrap service implementation |
| `internal/cluster/service.go` | Cluster lifecycle management service |
| `internal/cluster/providers/k3d/manager.go` | K3D cluster provider with Windows/WSL support |
| `internal/chart/services/chart_service.go` | Chart installation orchestration |
| `internal/chart/providers/helm/manager.go` | Helm chart management with native Kubernetes clients |
| `internal/chart/providers/argocd/applications.go` | ArgoCD application management |
| `internal/dev/services/intercept/service.go` | Telepresence intercept management |
| `internal/shared/executor/executor.go` | Command execution with WSL support |
| `internal/shared/ui/` | Terminal UI components and wizards |

## Dependencies

The CLI integrates with several external tools and libraries:

### External Tool Dependencies
- **K3D**: Local Kubernetes cluster creation and management
- **Helm**: Package manager for Kubernetes applications
- **ArgoCD**: GitOps continuous delivery tool
- **Telepresence**: Service mesh development tool for traffic interception
- **Kubectl**: Kubernetes command-line interface
- **Docker**: Container runtime required by K3D

### Go Library Dependencies
- **Cobra**: CLI framework for command structure and flag parsing
- **pterm**: Terminal UI library for spinners, progress bars, and interactive prompts
- **client-go**: Official Kubernetes Go client library
- **promptui**: Interactive prompt library for user input
- **yaml.v3**: YAML parsing for configuration files

### Platform-Specific Features
- **Windows/WSL2**: Special handling for Docker networking and path conversion
- **Certificate Management**: Automatic mkcert integration for HTTPS development
- **Prerequisites**: Auto-detection and installation of missing tools

## CLI Commands

### Core Commands

```bash
# Bootstrap complete environment
openframe bootstrap [cluster-name] [--deployment-mode=oss-tenant|saas-tenant|saas-shared]

# Cluster management
openframe cluster create [cluster-name] [--nodes=3] [--skip-wizard]
openframe cluster delete [cluster-name] [--force]  
openframe cluster list [--quiet]
openframe cluster status [cluster-name] [--detailed]
openframe cluster cleanup [cluster-name] [--force]

# Chart management  
openframe chart install [cluster-name] [--deployment-mode=oss-tenant]

# Development tools
openframe dev intercept [service-name] [--port=8080] [--namespace=default]
openframe dev skaffold [cluster-name] [--skip-bootstrap]
```

### Global Flags

```bash
--verbose, -v    Enable detailed logging
--dry-run        Show actions without executing  
--force, -f      Skip confirmation prompts
--silent         Suppress non-error output
```

### Interactive vs Non-Interactive Modes

The CLI provides both interactive wizards for ease of use and flag-based operation for automation:

- **Interactive Mode**: Step-by-step prompts with smart defaults
- **Non-Interactive Mode**: Full flag specification for CI/CD pipelines
- **Hybrid Mode**: Partial flags with prompts for missing required values

Each command includes comprehensive help documentation with examples for both usage patterns.
