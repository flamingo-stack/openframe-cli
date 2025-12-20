# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides a comprehensive set of commands for cluster lifecycle management, chart installation (ArgoCD), and developer tools including traffic interception and scaffolding.

## Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[Main CLI]
        ClusterCmd[Cluster Commands]
        ChartCmd[Chart Commands]
        DevCmd[Dev Commands]
        BootstrapCmd[Bootstrap Command]
    end
    
    subgraph "Service Layer"
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Dev Service]
        BootstrapSvc[Bootstrap Service]
    end
    
    subgraph "Infrastructure Layer"
        K3d[K3d Provider]
        Helm[Helm Provider]
        ArgoCD[ArgoCD Provider]
        Telepresence[Telepresence Provider]
        Skaffold[Skaffold Provider]
    end
    
    subgraph "External Systems"
        K8s[Kubernetes Clusters]
        Docker[Docker]
        GitHub[GitHub Repositories]
    end
    
    CLI --> ClusterCmd
    CLI --> ChartCmd
    CLI --> DevCmd
    CLI --> BootstrapCmd
    
    ClusterCmd --> ClusterSvc
    ChartCmd --> ChartSvc
    DevCmd --> DevSvc
    BootstrapCmd --> BootstrapSvc
    
    ClusterSvc --> K3d
    ChartSvc --> Helm
    ChartSvc --> ArgoCD
    DevSvc --> Telepresence
    DevSvc --> Skaffold
    
    K3d --> K8s
    K3d --> Docker
    Helm --> K8s
    ArgoCD --> K8s
    ArgoCD --> GitHub
```

## Core Components

| Component | Package | Responsibilities |
|-----------|---------|------------------|
| **Cluster Commands** | `cmd/cluster/` | CLI interface for cluster operations (create, delete, list, status, cleanup) |
| **Chart Commands** | `cmd/chart/` | CLI interface for ArgoCD and Helm chart management |
| **Dev Commands** | `cmd/dev/` | CLI interface for development tools (intercept, scaffold) |
| **Bootstrap Commands** | `cmd/bootstrap/` | CLI interface for complete OpenFrame environment setup |
| **Cluster Service** | `internal/cluster/` | Business logic for K3d cluster lifecycle management |
| **Chart Service** | `internal/chart/` | Business logic for ArgoCD and chart installation |
| **Dev Service** | `internal/dev/` | Business logic for development workflows |
| **Bootstrap Service** | `internal/bootstrap/` | Orchestrates cluster creation and chart installation |
| **UI Components** | `internal/*/ui/` | User interface and interaction logic |
| **Prerequisites** | `internal/*/prerequisites/` | System requirements validation and tool installation |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CC[Cluster Commands]
        ChC[Chart Commands]
        DC[Dev Commands]
        BC[Bootstrap Commands]
    end
    
    subgraph "Service Layer"
        CS[Cluster Service]
        ChS[Chart Service]
        DS[Dev Service]
        BS[Bootstrap Service]
    end
    
    subgraph "Shared Components"
        UI[UI Components]
        P[Prerequisites]
        E[Error Handling]
        M[Models/Types]
    end
    
    CC --> CS
    ChC --> ChS
    DC --> DS
    BC --> BS
    
    BS --> CS
    BS --> ChS
    
    CS --> UI
    ChS --> UI
    DS --> UI
    
    CS --> P
    ChS --> P
    DS --> P
    
    CS --> E
    ChS --> E
    DS --> E
    BS --> E
    
    CS --> M
    ChS --> M
    DS --> M
    BS --> M
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Service
    participant UI
    participant Prerequisites
    participant Provider
    participant K8s
    
    User->>CLI: Execute command
    CLI->>UI: Show logo/context
    CLI->>Prerequisites: Check requirements
    Prerequisites-->>CLI: Validation result
    CLI->>Service: Process request
    Service->>UI: Get user input (if interactive)
    UI-->>Service: Configuration
    Service->>Provider: Execute operation
    Provider->>K8s: Apply changes
    K8s-->>Provider: Operation result
    Provider-->>Service: Status
    Service->>UI: Show result
    UI-->>User: Display output
```

## Key Files

| File | Purpose |
|------|---------|
| **cmd/cluster/cluster.go** | Main cluster command definition and subcommand registration |
| **cmd/cluster/create.go** | Cluster creation command with interactive wizard |
| **cmd/bootstrap/bootstrap.go** | Complete OpenFrame environment setup command |
| **cmd/chart/install.go** | ArgoCD and chart installation command |
| **cmd/dev/dev.go** | Development tools command group |
| **internal/cluster/services/** | Core cluster management business logic |
| **internal/chart/services/** | Chart installation and ArgoCD management logic |
| **internal/bootstrap/** | Bootstrap orchestration service |
| **internal/shared/ui/** | Shared UI components and logo display |
| **internal/shared/errors/** | Global error handling utilities |

## Dependencies

The OpenFrame CLI integrates with several external tools and systems:

- **K3d**: For local Kubernetes cluster management
- **Helm**: For chart installation and package management
- **ArgoCD**: For GitOps-based application deployment
- **Telepresence**: For traffic interception during development
- **Skaffold**: For continuous development workflows
- **Docker**: For container management
- **kubectl**: For Kubernetes cluster interaction

## CLI Commands

### Cluster Management
```bash
openframe cluster create [NAME]        # Create new K3d cluster
openframe cluster delete [NAME]        # Delete cluster
openframe cluster list                 # List all clusters
openframe cluster status [NAME]        # Show cluster status
openframe cluster cleanup [NAME]       # Clean up cluster resources
```

### Chart Management
```bash
openframe chart install [CLUSTER]      # Install ArgoCD and charts
```

### Development Tools
```bash
openframe dev intercept [SERVICE]      # Intercept service traffic
openframe dev scaffold [CLUSTER]       # Run Skaffold development
```

### Bootstrap
```bash
openframe bootstrap [CLUSTER]          # Complete OpenFrame setup
```

### Global Options
- `--verbose, -v`: Enable detailed logging
- `--force`: Skip confirmations
- `--dry-run`: Show what would be done without executing
- `--non-interactive`: Skip interactive prompts
- `--deployment-mode`: Specify deployment mode (oss-tenant, saas-tenant, saas-shared)
