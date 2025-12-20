# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides cluster lifecycle management (create, delete, status), chart installation with ArgoCD, and developer-focused tools for local Kubernetes development including traffic interception and live reloading.

## Architecture

```mermaid
graph TB
    subgraph "CLI Layer"
        CMD[cmd/]
        COBRA[Cobra Commands]
    end
    
    subgraph "Internal Services"
        CLUSTER[internal/cluster/]
        CHART[internal/chart/]
        DEV[internal/dev/]
        BOOTSTRAP[internal/bootstrap/]
        SHARED[internal/shared/]
    end
    
    subgraph "External Systems"
        K3D[K3d Clusters]
        KUBE[Kubernetes API]
        HELM[Helm/ArgoCD]
        DOCKER[Docker]
        TELEPRESENCE[Telepresence]
        SKAFFOLD[Skaffold]
    end
    
    CMD --> CLUSTER
    CMD --> CHART
    CMD --> DEV
    CMD --> BOOTSTRAP
    
    CLUSTER --> K3D
    CLUSTER --> KUBE
    CLUSTER --> DOCKER
    
    CHART --> HELM
    CHART --> KUBE
    
    DEV --> TELEPRESENCE
    DEV --> SKAFFOLD
    DEV --> KUBE
    
    BOOTSTRAP --> CLUSTER
    BOOTSTRAP --> CHART
    
    CLUSTER --> SHARED
    CHART --> SHARED
    DEV --> SHARED
    BOOTSTRAP --> SHARED
```

## Core Components

| Component | Purpose | Key Responsibilities |
|-----------|---------|---------------------|
| `cmd/` | CLI Command Layer | Command definitions, flag parsing, argument validation |
| `internal/cluster/` | Cluster Management | K3d cluster CRUD, status monitoring, resource cleanup |
| `internal/chart/` | Chart Installation | ArgoCD deployment, Helm chart management, app-of-apps setup |
| `internal/dev/` | Development Tools | Traffic interception, Skaffold workflows, local development |
| `internal/bootstrap/` | Full Setup Orchestration | End-to-end environment provisioning combining cluster + charts |
| `internal/shared/` | Common Utilities | UI components, error handling, configuration management |

## Component Relationships

```mermaid
graph LR
    subgraph "Command Layer"
        CC[cluster commands]
        CHC[chart commands]
        DC[dev commands]
        BC[bootstrap command]
    end
    
    subgraph "Service Layer"
        CS[cluster services]
        CHS[chart services]
        DS[dev services]
        BS[bootstrap service]
    end
    
    subgraph "Shared Infrastructure"
        UI[ui components]
        ERR[error handling]
        PREREQ[prerequisites]
        MODELS[models/types]
    end
    
    CC --> CS
    CHC --> CHS
    DC --> DS
    BC --> BS
    
    BS --> CS
    BS --> CHS
    
    CS --> UI
    CS --> ERR
    CS --> PREREQ
    CHS --> UI
    CHS --> ERR
    CHS --> PREREQ
    DS --> UI
    DS --> PREREQ
    
    CS --> MODELS
    CHS --> MODELS
    DS --> MODELS
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
    participant Kubernetes
    
    Note over User,Kubernetes: Bootstrap Flow (openframe bootstrap)
    
    User->>CLI: openframe bootstrap
    CLI->>CLI: Show logo & check prerequisites
    CLI->>ClusterService: Create cluster
    ClusterService->>K3d: k3d cluster create
    K3d->>Kubernetes: Deploy cluster nodes
    K3d-->>ClusterService: Cluster ready
    ClusterService-->>CLI: Cluster created
    
    CLI->>ChartService: Install charts
    ChartService->>Kubernetes: Install ArgoCD via Helm
    ChartService->>ArgoCD: Deploy app-of-apps
    ArgoCD->>Kubernetes: Sync OpenFrame applications
    ChartService-->>CLI: Charts installed
    
    CLI-->>User: Bootstrap complete
    
    Note over User,Kubernetes: Development Flow
    
    User->>CLI: openframe dev intercept service
    CLI->>Kubernetes: Setup traffic interception
    CLI->>User: Local development proxy active
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/cluster/cluster.go` | Main cluster command with subcommands (create, delete, list, status, cleanup) |
| `cmd/chart/chart.go` | Chart management commands for ArgoCD installation |
| `cmd/bootstrap/bootstrap.go` | End-to-end setup command combining cluster creation and chart installation |
| `cmd/dev/dev.go` | Development tools for traffic interception and live reloading |
| `internal/cluster/models/` | Cluster configuration models and validation |
| `internal/cluster/services/` | Core cluster operations (create, delete, status) |
| `internal/chart/services/` | ArgoCD and Helm chart installation logic |
| `internal/shared/ui/` | Common UI components for consistent user experience |
| `internal/shared/errors/` | Centralized error handling and user-friendly error messages |

## Dependencies

The project uses several key external dependencies:

- **Cobra**: CLI framework for command structure and flag management
- **K3d**: Lightweight Kubernetes distribution for local development clusters
- **Helm**: Package manager for Kubernetes applications
- **Docker**: Container runtime for cluster nodes and application deployment
- **Telepresence**: Traffic interception for local development workflows
- **Skaffold**: Continuous development workflow for Kubernetes applications
- **ArgoCD**: GitOps continuous delivery tool for application management

## CLI Commands

| Command | Purpose | Example |
|---------|---------|---------|
| `cluster create` | Create new K3d cluster with interactive or default configuration | `openframe cluster create my-cluster` |
| `cluster delete` | Remove cluster and clean up all resources | `openframe cluster delete my-cluster --force` |
| `cluster list` | Display all managed clusters in table format | `openframe cluster list` |
| `cluster status` | Show detailed cluster health and application status | `openframe cluster status my-cluster --detailed` |
| `cluster cleanup` | Clean up unused Docker images and resources | `openframe cluster cleanup my-cluster` |
| `chart install` | Install ArgoCD and OpenFrame applications | `openframe chart install --deployment-mode=oss-tenant` |
| `bootstrap` | Complete setup: create cluster + install charts | `openframe bootstrap --deployment-mode=oss-tenant` |
| `dev intercept` | Intercept cluster traffic for local development | `openframe dev intercept my-service` |
| `dev skaffold` | Run Skaffold for continuous development | `openframe dev skaffold my-cluster` |

### Common Flags

- `--verbose, -v`: Show detailed logging and progress information
- `--deployment-mode`: Specify deployment type (oss-tenant, saas-tenant, saas-shared)
- `--non-interactive`: Skip prompts for CI/CD environments
- `--force`: Skip confirmations for destructive operations
- `--dry-run`: Show what would be done without executing
