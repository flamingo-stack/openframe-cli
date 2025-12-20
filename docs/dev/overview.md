# openframe-cli Module Documentation

# OpenFrame CLI Architecture

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides comprehensive cluster lifecycle management (create, delete, list, status, cleanup), ArgoCD chart installation, and development tools for service traffic interception and live reloading with Telepresence and Skaffold.

## Architecture

The CLI follows a layered architecture with clear separation of concerns, built around the Cobra command framework with modular service layers.

```mermaid
graph TB
    CLI[CLI Layer - Cobra Commands]
    UI[UI Layer - Interactive Components]
    Service[Service Layer - Business Logic]
    Models[Models Layer - Data Structures]
    Utils[Utils Layer - Common Utilities]
    External[External Tools - K3d, Helm, ArgoCD]
    
    CLI --> UI
    CLI --> Service
    CLI --> Models
    Service --> Utils
    Service --> External
    UI --> Models
    Models --> Utils
    
    subgraph "Command Groups"
        ClusterCmds[Cluster Commands]
        ChartCmds[Chart Commands] 
        DevCmds[Dev Commands]
        BootstrapCmd[Bootstrap Command]
    end
    
    CLI --> ClusterCmds
    CLI --> ChartCmds
    CLI --> DevCmds
    CLI --> BootstrapCmd
```

## Core Components

| Component | Package | Responsibilities |
|-----------|---------|------------------|
| **Cluster Management** | `cmd/cluster/`, `internal/cluster/` | K3d cluster lifecycle (create, delete, list, status, cleanup) |
| **Chart Management** | `cmd/chart/`, `internal/chart/` | ArgoCD installation and Helm chart management |
| **Development Tools** | `cmd/dev/`, `internal/dev/` | Telepresence intercepts and Skaffold workflows |
| **Bootstrap Orchestration** | `cmd/bootstrap/`, `internal/bootstrap/` | End-to-end environment setup combining cluster + charts |
| **UI Components** | `internal/*/ui/` | Interactive prompts, configuration wizards, progress displays |
| **Service Layer** | `internal/*/services/` | Business logic and external tool integration |
| **Models & Types** | `internal/*/models/`, `internal/*/types/` | Data structures, configuration, and validation |
| **Prerequisites** | `internal/*/prerequisites/` | Tool availability checks and installation |
| **Shared Utilities** | `internal/shared/` | Common error handling, UI components, utilities |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CC[Cluster Commands]
        ChC[Chart Commands]
        DC[Dev Commands] 
        BC[Bootstrap Command]
    end
    
    subgraph "Internal Services"
        CS[Cluster Service]
        CHS[Chart Service]
        DS[Dev Service]
        BS[Bootstrap Service]
    end
    
    subgraph "Shared Components"
        UI[Shared UI]
        ERR[Error Handling]
        UTIL[Utilities]
    end
    
    subgraph "External Tools"
        K3D[K3d]
        HELM[Helm]
        ARGO[ArgoCD]
        TELE[Telepresence]
        SKAF[Skaffold]
    end
    
    CC --> CS
    ChC --> CHS
    DC --> DS
    BC --> BS
    
    BS --> CS
    BS --> CHS
    
    CS --> K3D
    CHS --> HELM
    CHS --> ARGO
    DS --> TELE
    DS --> SKAF
    
    CS --> UI
    CHS --> UI
    DS --> UI
    BS --> UI
    
    CS --> ERR
    CHS --> ERR
    DS --> ERR
    BS --> ERR
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant UI
    participant Service
    participant External
    
    User->>CLI: openframe bootstrap
    CLI->>UI: Show logo & collect config
    UI->>User: Interactive prompts
    User->>UI: Provide cluster config
    UI->>Service: Create cluster request
    Service->>External: k3d cluster create
    External->>Service: Cluster created
    Service->>UI: Show progress
    UI->>Service: Install charts request
    Service->>External: helm install argocd
    External->>Service: ArgoCD installed
    Service->>External: kubectl apply app-of-apps
    External->>Service: Apps deployed
    Service->>UI: Installation complete
    UI->>User: Success message
    
    Note over CLI,External: Bootstrap orchestrates cluster creation + chart installation
    
    User->>CLI: openframe dev intercept
    CLI->>UI: Select service
    UI->>Service: Intercept request
    Service->>External: telepresence intercept
    External->>Service: Traffic intercepted
    Service->>UI: Intercept active
    UI->>User: Local development ready
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bootstrap/bootstrap.go` | Main bootstrap command orchestrating full environment setup |
| `cmd/cluster/create.go` | Cluster creation with interactive configuration wizard |
| `cmd/cluster/delete.go` | Cluster deletion with confirmation and cleanup |
| `cmd/chart/install.go` | ArgoCD installation and app-of-apps deployment |
| `cmd/dev/intercept.go` | Telepresence traffic interception for local development |
| `internal/cluster/services/` | Core cluster management business logic |
| `internal/chart/services/` | Chart installation and ArgoCD management |
| `internal/shared/ui/` | Common UI components and interactive elements |
| `internal/shared/errors/` | Centralized error handling and user-friendly messages |

## Dependencies

The CLI integrates with several external Kubernetes tools and libraries:

- **K3d**: Lightweight Kubernetes distribution for local development clusters
- **Helm**: Package manager for installing ArgoCD and other Kubernetes applications  
- **ArgoCD**: GitOps continuous delivery tool for application deployment
- **Telepresence**: Traffic interception for local service development
- **Skaffold**: Local development workflow with live code reloading
- **Cobra**: CLI framework providing command structure and flag parsing
- **Kubernetes Client**: Direct cluster API interaction for status and management

## CLI Commands

### Cluster Management
```bash
openframe cluster create [name]          # Create new K3d cluster with wizard
openframe cluster delete [name]          # Delete cluster with confirmation
openframe cluster list                   # Show all managed clusters
openframe cluster status [name]          # Display detailed cluster info
openframe cluster cleanup [name]         # Clean up unused resources
```

### Chart Management  
```bash
openframe chart install [cluster]        # Install ArgoCD and app-of-apps
```

### Development Tools
```bash
openframe dev intercept [service]        # Intercept service traffic locally
openframe dev skaffold [cluster]         # Deploy with live reloading
```

### Bootstrap (End-to-End)
```bash
openframe bootstrap [cluster]            # Create cluster + install charts
openframe bootstrap --deployment-mode=oss-tenant --non-interactive
```

### Common Flags
- `--verbose, -v`: Detailed logging and progress information
- `--non-interactive`: Skip prompts for CI/CD environments  
- `--deployment-mode`: Pre-select deployment type (oss-tenant, saas-tenant, saas-shared)
- `--force`: Skip confirmations and force operations
- `--dry-run`: Show what would happen without executing
