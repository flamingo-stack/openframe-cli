# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

OpenFrame CLI is a comprehensive command-line interface for managing Kubernetes clusters and deploying OpenFrame environments. It provides streamlined workflows for cluster lifecycle management, ArgoCD chart installation, development tools, and complete environment bootstrapping.

## Architecture

The CLI follows a modular architecture with clear separation between command interfaces, business logic, and external integrations. Each major command group (cluster, chart, dev, bootstrap) has its own module with dedicated services, UI components, and configuration models.

### High-Level Architecture

```mermaid
graph TB
    CLI[OpenFrame CLI] --> Bootstrap[Bootstrap Module]
    CLI --> Cluster[Cluster Module]
    CLI --> Chart[Chart Module]
    CLI --> Dev[Dev Module]
    
    Bootstrap --> ClusterSvc[Cluster Service]
    Bootstrap --> ChartSvc[Chart Service]
    
    Cluster --> K3d[K3d Provider]
    Cluster --> Prerequisites[Prerequisites Check]
    
    Chart --> ArgoCD[ArgoCD Installation]
    Chart --> Helm[Helm Operations]
    
    Dev --> Telepresence[Traffic Intercept]
    Dev --> Skaffold[Live Development]
    
    External[External Tools]
    K3d --> External
    Helm --> External
    Telepresence --> External
    Skaffold --> External
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|---------------|
| **Bootstrap** | `cmd/bootstrap` | Orchestrates complete OpenFrame environment setup |
| **Cluster Management** | `cmd/cluster` | Kubernetes cluster lifecycle (create, delete, status, cleanup) |
| **Chart Management** | `cmd/chart` | ArgoCD and OpenFrame chart installation |
| **Development Tools** | `cmd/dev` | Local development workflows with intercept and skaffold |
| **Cluster Services** | `internal/cluster` | Business logic for cluster operations |
| **Chart Services** | `internal/chart` | Helm chart installation and configuration |
| **Prerequisites** | `internal/*/prerequisites` | Dependency validation and installation |
| **UI Components** | `internal/*/ui` | User interface and interactive prompts |
| **Shared Utilities** | `internal/shared` | Common error handling and UI components |

## Component Relationships

```mermaid
graph TD
    subgraph "Command Layer"
        BootstrapCmd[Bootstrap Command]
        ClusterCmd[Cluster Commands]
        ChartCmd[Chart Commands]
        DevCmd[Dev Commands]
    end
    
    subgraph "Service Layer"
        BootstrapSvc[Bootstrap Service]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
        DevSvc[Dev Service]
    end
    
    subgraph "Infrastructure Layer"
        Prerequisites[Prerequisites]
        UI[UI Components]
        Models[Data Models]
        Utils[Utilities]
    end
    
    subgraph "External Dependencies"
        K3d[K3d CLI]
        Helm[Helm CLI]
        Kubectl[kubectl]
        Telepresence[Telepresence]
        Skaffold[Skaffold CLI]
    end
    
    BootstrapCmd --> BootstrapSvc
    ClusterCmd --> ClusterSvc
    ChartCmd --> ChartSvc
    DevCmd --> DevSvc
    
    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc
    
    ClusterSvc --> Prerequisites
    ChartSvc --> Prerequisites
    DevSvc --> Prerequisites
    
    ClusterSvc --> UI
    ChartSvc --> UI
    DevSvc --> UI
    
    ClusterSvc --> K3d
    ClusterSvc --> Kubectl
    ChartSvc --> Helm
    ChartSvc --> Kubectl
    DevSvc --> Telepresence
    DevSvc --> Skaffold
```

## Data Flow

### Bootstrap Command Flow

```mermaid
sequenceDiagram
    participant User
    participant BootstrapCmd as Bootstrap Command
    participant ClusterSvc as Cluster Service
    participant ChartSvc as Chart Service
    participant K3d
    participant Helm
    
    User->>BootstrapCmd: openframe bootstrap [cluster-name]
    BootstrapCmd->>BootstrapCmd: Parse flags (deployment-mode, non-interactive, verbose)
    BootstrapCmd->>ClusterSvc: Create cluster
    ClusterSvc->>K3d: k3d cluster create
    K3d-->>ClusterSvc: Cluster created
    ClusterSvc-->>BootstrapCmd: Cluster ready
    BootstrapCmd->>ChartSvc: Install charts
    ChartSvc->>Helm: Install ArgoCD
    ChartSvc->>Helm: Install app-of-apps
    Helm-->>ChartSvc: Charts installed
    ChartSvc-->>BootstrapCmd: Installation complete
    BootstrapCmd-->>User: Environment ready
```

### Cluster Management Flow

```mermaid
sequenceDiagram
    participant User
    participant ClusterCmd as Cluster Command
    participant UI as Operations UI
    participant ClusterSvc as Cluster Service
    participant K3d
    
    User->>ClusterCmd: openframe cluster create [name]
    ClusterCmd->>UI: Show configuration wizard
    UI-->>User: Interactive prompts
    User-->>UI: Configuration choices
    UI-->>ClusterCmd: Cluster config
    ClusterCmd->>ClusterSvc: Create cluster with config
    ClusterSvc->>K3d: k3d cluster create with options
    K3d-->>ClusterSvc: Cluster status
    ClusterSvc-->>ClusterCmd: Creation result
    ClusterCmd-->>User: Success/Error message
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bootstrap/bootstrap.go` | Entry point for complete environment setup |
| `cmd/cluster/cluster.go` | Main cluster management command structure |
| `cmd/cluster/create.go` | Cluster creation with interactive configuration |
| `cmd/chart/install.go` | ArgoCD and OpenFrame chart installation |
| `cmd/dev/dev.go` | Development tools command structure |
| `internal/cluster/services/` | Core cluster management business logic |
| `internal/chart/services/` | Chart installation and configuration logic |
| `internal/bootstrap/` | Bootstrap orchestration service |
| `internal/shared/ui/` | Reusable UI components and logo display |
| `internal/shared/errors/` | Centralized error handling utilities |

## Dependencies

Based on the modular structure, this project likely depends on:

- **Cobra CLI Framework**: Command structure and flag management
- **Kubernetes Client Libraries**: Cluster interaction and resource management  
- **Helm Libraries**: Chart installation and repository management
- **External CLI Tools**: K3d, kubectl, Telepresence, Skaffold integration
- **Terminal UI Libraries**: Interactive prompts and progress indicators
- **YAML Processing**: Configuration file management
- **Git Libraries**: Repository cloning and branch management for app-of-apps

The architecture emphasizes loose coupling between command interfaces and implementation details, allowing for easy testing and maintenance of individual components.

## CLI Commands

| Command | Description | Examples |
|---------|-------------|----------|
| `openframe bootstrap [cluster-name]` | Complete environment setup (cluster + charts) | Interactive mode, CI/CD mode with flags |
| `openframe cluster create [name]` | Create new Kubernetes cluster | Interactive wizard or direct creation |
| `openframe cluster delete [name]` | Delete cluster and cleanup resources | With confirmation or `--force` |
| `openframe cluster list` | Show all managed clusters | Table format with status |
| `openframe cluster status [name]` | Detailed cluster information | Health, nodes, applications |
| `openframe cluster cleanup [name]` | Remove unused cluster resources | Free disk space, cleanup images |
| `openframe chart install [cluster-name]` | Install ArgoCD and app-of-apps | Deployment mode selection |
| `openframe dev intercept [service-name]` | Traffic interception for development | Telepresence integration |
| `openframe dev skaffold [cluster-name]` | Live development with hot reload | Skaffold workflow automation |

Each command supports both interactive and non-interactive modes, with comprehensive flag options for CI/CD automation and detailed logging with the `--verbose` flag.
