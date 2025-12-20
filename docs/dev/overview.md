# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides a unified interface for cluster lifecycle management (creation, deletion, monitoring), chart installation with ArgoCD, and development tools like Telepresence and Skaffold integration for local Kubernetes development.

## Architecture

```mermaid
graph TB
    CLI[OpenFrame CLI] --> CMD[Command Layer]
    CMD --> CLUSTER[Cluster Management]
    CMD --> CHART[Chart Management] 
    CMD --> BOOTSTRAP[Bootstrap]
    CMD --> DEV[Development Tools]
    
    CLUSTER --> CLUSTERUI[Cluster UI]
    CLUSTER --> CLUSTERMODELS[Cluster Models]
    CLUSTER --> CLUSTERUTILS[Cluster Utils]
    
    CHART --> CHARTSERVICES[Chart Services]
    CHART --> CHARTTYPES[Chart Types]
    CHART --> CHARTPREREQ[Chart Prerequisites]
    
    BOOTSTRAP --> BOOTSTRAPSERVICE[Bootstrap Service]
    
    DEV --> DEVMODELS[Dev Models]
    DEV --> DEVPREREQ[Dev Prerequisites]
    
    CLUSTERUI --> SHAREDUI[Shared UI]
    CHARTSERVICES --> SHAREDUI
    BOOTSTRAPSERVICE --> SHAREDUI
    
    CLUSTER --> K3D[K3d Integration]
    CHART --> ARGOCD[ArgoCD Integration]
    DEV --> TELEPRESENCE[Telepresence]
    DEV --> SKAFFOLD[Skaffold]
    
    SHAREDUI --> SHAREDERRORS[Shared Errors]
```

## Core Components

| Component | Package Path | Responsibility |
|-----------|-------------|----------------|
| **Command Layer** | `cmd/` | CLI command definitions and flag handling using Cobra |
| **Cluster Management** | `internal/cluster/` | K3d cluster lifecycle, status monitoring, operations |
| **Chart Management** | `internal/chart/` | Helm chart and ArgoCD installation and management |
| **Bootstrap Service** | `internal/bootstrap/` | Orchestrates full OpenFrame environment setup |
| **Development Tools** | `internal/dev/` | Telepresence intercepts and Skaffold integration |
| **Shared UI** | `internal/shared/ui/` | Common UI components, logo display, user interactions |
| **Shared Errors** | `internal/shared/errors/` | Centralized error handling and display |
| **Prerequisites** | `*/prerequisites/` | Tool validation and installation across modules |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CLUSTERCMD[Cluster Commands]
        CHARTCMD[Chart Commands]
        BOOTSTRAPCMD[Bootstrap Command]
        DEVCMD[Dev Commands]
    end
    
    subgraph "Internal Services"
        CLUSTERSERVICE[Cluster Service]
        CHARTSERVICE[Chart Service]  
        BOOTSTRAPSERVICE[Bootstrap Service]
        DEVSERVICE[Dev Service]
    end
    
    subgraph "Shared Components"
        UI[Shared UI]
        ERRORS[Error Handler]
        MODELS[Models/Types]
    end
    
    subgraph "External Tools"
        K3D[K3d]
        HELM[Helm]
        ARGOCD[ArgoCD]
        TELEPRESENCE[Telepresence]
        SKAFFOLD[Skaffold]
    end
    
    CLUSTERCMD --> CLUSTERSERVICE
    CHARTCMD --> CHARTSERVICE
    BOOTSTRAPCMD --> BOOTSTRAPSERVICE
    DEVCMD --> DEVSERVICE
    
    CLUSTERSERVICE --> UI
    CHARTSERVICE --> UI
    BOOTSTRAPSERVICE --> UI
    DEVSERVICE --> UI
    
    CLUSTERSERVICE --> ERRORS
    CHARTSERVICE --> ERRORS
    BOOTSTRAPSERVICE --> ERRORS
    
    BOOTSTRAPSERVICE --> CLUSTERSERVICE
    BOOTSTRAPSERVICE --> CHARTSERVICE
    
    CLUSTERSERVICE --> K3D
    CHARTSERVICE --> HELM
    CHARTSERVICE --> ARGOCD
    DEVSERVICE --> TELEPRESENCE
    DEVSERVICE --> SKAFFOLD
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant ClusterService
    participant ChartService
    participant K3d
    participant ArgoCD
    
    User->>CLI: openframe bootstrap
    CLI->>CLI: Show logo & validate flags
    CLI->>ClusterService: Create cluster
    ClusterService->>K3d: k3d cluster create
    K3d-->>ClusterService: Cluster created
    ClusterService-->>CLI: Success
    
    CLI->>ChartService: Install charts
    ChartService->>ChartService: Generate certificates
    ChartService->>K3d: helm install argocd
    K3d-->>ChartService: ArgoCD installed
    ChartService->>ArgoCD: Deploy app-of-apps
    ArgoCD-->>ChartService: Apps synced
    ChartService-->>CLI: Installation complete
    
    CLI-->>User: Bootstrap successful
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bootstrap/bootstrap.go` | Main bootstrap command orchestrating cluster + chart installation |
| `cmd/cluster/cluster.go` | Cluster management command group with subcommands |
| `cmd/chart/install.go` | ArgoCD and chart installation command with flag handling |
| `internal/cluster/utils/utils.go` | Global flag management and command service initialization |
| `internal/cluster/ui/operations.go` | User interface for cluster operations and selections |
| `internal/chart/services/install.go` | Core chart installation logic and ArgoCD management |
| `internal/shared/ui/logo.go` | Consistent branding and UI presentation |
| `internal/shared/errors/handler.go` | Centralized error handling with verbose logging |

## Dependencies

The OpenFrame CLI integrates with several external tools and libraries:

- **Cobra Framework**: Command-line interface structure and flag management
- **K3d**: Lightweight Kubernetes distribution for local development clusters
- **Helm**: Package manager for Kubernetes applications and chart installation
- **ArgoCD**: GitOps continuous delivery tool for Kubernetes
- **Telepresence**: Local development tool for intercepting cluster traffic  
- **Skaffold**: Continuous development tool for Kubernetes applications
- **Docker**: Container runtime required by K3d for cluster nodes

## CLI Commands

### Cluster Management
```bash
openframe cluster create [name]     # Create new K3d cluster
openframe cluster delete [name]     # Delete cluster and cleanup
openframe cluster list              # Show all clusters
openframe cluster status [name]     # Detailed cluster information
openframe cluster cleanup [name]    # Clean unused resources
```

### Chart Management
```bash
openframe chart install [cluster]   # Install ArgoCD and charts
```

### Bootstrap (Combined Operations)
```bash
openframe bootstrap [cluster]       # Create cluster + install charts
  --deployment-mode=oss-tenant      # Specify deployment type
  --non-interactive                 # Skip prompts
  --verbose                         # Detailed logging
```

### Development Tools
```bash
openframe dev intercept [service]   # Intercept service traffic
openframe dev skaffold [cluster]    # Development deployment
```

Each command group supports interactive mode by default with options for non-interactive automation, making it suitable for both developer workflows and CI/CD pipelines.
