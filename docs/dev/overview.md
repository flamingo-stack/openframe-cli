# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides comprehensive cluster lifecycle management, Helm chart installation with ArgoCD, and development tools including Telepresence traffic interception and Skaffold workflows.

## Architecture

The CLI follows a modular, layered architecture with clear separation between commands, business logic, and external integrations.

```mermaid
flowchart TD
    CLI[CLI Entry Point] --> Commands{Command Router}
    
    Commands --> Cluster[Cluster Commands]
    Commands --> Chart[Chart Commands] 
    Commands --> Bootstrap[Bootstrap Commands]
    Commands --> Dev[Dev Commands]
    
    Cluster --> ClusterService[Cluster Service Layer]
    Chart --> ChartService[Chart Service Layer]
    Bootstrap --> BootstrapService[Bootstrap Service]
    Dev --> DevService[Dev Service Layer]
    
    ClusterService --> K3d[K3d Provider]
    ClusterService --> ClusterUI[Cluster UI]
    
    ChartService --> Helm[Helm Integration]
    ChartService --> ArgoCD[ArgoCD Management]
    
    DevService --> Telepresence[Telepresence Integration]
    DevService --> Skaffold[Skaffold Integration]
    
    K3d --> Docker[Docker Engine]
    Helm --> K8s[Kubernetes API]
    ArgoCD --> Git[Git Repository]
    
    ClusterUI --> SharedUI[Shared UI Components]
    ChartService --> SharedUI
    DevService --> SharedUI
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **Command Layer** | `cmd/*` | CLI command definitions, argument parsing, and flag handling |
| **Cluster Management** | `internal/cluster` | K3d cluster lifecycle, status monitoring, resource cleanup |
| **Chart Management** | `internal/chart` | Helm chart installation, ArgoCD setup, app-of-apps deployment |
| **Bootstrap Service** | `internal/bootstrap` | Orchestrates complete OpenFrame environment setup |
| **Dev Tools** | `internal/dev` | Telepresence intercepts, Skaffold workflows |
| **Shared Services** | `internal/shared` | Common utilities, error handling, UI components |
| **Prerequisites** | `*/prerequisites` | Tool validation and installation for each command group |

## Component Relationships

```mermaid
flowchart LR
    subgraph "Command Layer"
        CmdCluster[Cluster Commands]
        CmdChart[Chart Commands]
        CmdBootstrap[Bootstrap Commands]
        CmdDev[Dev Commands]
    end
    
    subgraph "Service Layer"
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        BootstrapSvc[Bootstrap Service]
        DevSvc[Dev Service]
    end
    
    subgraph "Infrastructure"
        K3dProvider[K3d Provider]
        HelmProvider[Helm Provider]
        TelepresenceProvider[Telepresence Provider]
        SkaffoldProvider[Skaffold Provider]
    end
    
    subgraph "Shared"
        SharedUI[UI Components]
        SharedErrors[Error Handling]
        Prerequisites[Prerequisites]
    end
    
    CmdCluster --> ClusterSvc
    CmdChart --> ChartSvc
    CmdBootstrap --> BootstrapSvc
    CmdDev --> DevSvc
    
    ClusterSvc --> K3dProvider
    ChartSvc --> HelmProvider
    DevSvc --> TelepresenceProvider
    DevSvc --> SkaffoldProvider
    
    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc
    
    ClusterSvc --> SharedUI
    ChartSvc --> SharedUI
    DevSvc --> SharedUI
    
    ClusterSvc --> SharedErrors
    ChartSvc --> SharedErrors
    
    CmdCluster --> Prerequisites
    CmdChart --> Prerequisites
    CmdDev --> Prerequisites
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Service
    participant Provider
    participant External
    
    User->>CLI: openframe bootstrap
    CLI->>CLI: Parse flags & validate
    CLI->>Service: Bootstrap.Execute()
    
    Service->>Service: Check prerequisites
    Service->>Provider: Create K3d cluster
    Provider->>External: Docker API calls
    External-->>Provider: Cluster created
    Provider-->>Service: Cluster ready
    
    Service->>Provider: Install Helm charts
    Provider->>External: Kubernetes API
    External-->>Provider: Charts installed
    Provider-->>Service: Installation complete
    
    Service->>Provider: Setup ArgoCD
    Provider->>External: Git repository sync
    External-->>Provider: Apps deployed
    Provider-->>Service: Bootstrap complete
    
    Service-->>CLI: Success response
    CLI-->>User: Friendly success message
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/cluster/cluster.go` | Main cluster command entry point with subcommand routing |
| `cmd/bootstrap/bootstrap.go` | Bootstrap command orchestrating complete setup |
| `internal/cluster/services/` | Core cluster management business logic |
| `internal/chart/services/` | Helm chart and ArgoCD installation logic |
| `internal/shared/ui/` | Reusable UI components for consistent user experience |
| `internal/shared/errors/` | Global error handling and user-friendly error messages |
| `internal/cluster/models/` | Data structures and validation for cluster operations |
| `internal/cluster/ui/` | Cluster-specific UI components and interactive wizards |

## Dependencies

The project leverages several key dependencies for its functionality:

| Dependency | Usage | Integration Point |
|------------|-------|-------------------|
| **Cobra** | CLI framework for command structure and flag parsing | `cmd/*` packages |
| **Docker SDK** | Container management for K3d clusters | Cluster providers |
| **Kubernetes Client** | K8s API interactions for status and management | Chart and cluster services |
| **Helm SDK** | Chart installation and repository management | Chart service layer |
| **Survey/Promptui** | Interactive CLI prompts and wizards | UI components |
| **Yaml/JSON** | Configuration file parsing and generation | Model validation |

## CLI Commands

### Cluster Management
```bash
openframe cluster create [name]      # Create new K3d cluster with wizard
openframe cluster delete [name]      # Delete cluster and cleanup resources  
openframe cluster list               # Show all managed clusters
openframe cluster status [name]      # Detailed cluster health information
openframe cluster cleanup [name]     # Remove unused Docker images/resources
```

### Chart Management  
```bash
openframe chart install [cluster]    # Install ArgoCD and app-of-apps
```

### Bootstrap (Complete Setup)
```bash
openframe bootstrap [cluster]        # Create cluster + install charts
openframe bootstrap --deployment-mode=oss-tenant  # Skip deployment selection
openframe bootstrap --non-interactive # CI/CD mode with existing config
```

### Development Tools
```bash
openframe dev intercept [service]    # Telepresence traffic interception
openframe dev skaffold [cluster]     # Live development with Skaffold
```

### Global Flags
- `--verbose, -v` - Detailed logging and operation visibility
- `--force` - Skip confirmations (where applicable)
- `--dry-run` - Preview operations without execution
- `--non-interactive` - Automated mode for CI/CD pipelines
