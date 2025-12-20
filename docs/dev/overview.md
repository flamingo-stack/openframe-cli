# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing Kubernetes clusters and development workflows, specifically designed for the OpenFrame platform. It provides interactive cluster management, Helm chart installation with ArgoCD, and developer tools like Telepresence intercepts and Skaffold workflows, replacing shell scripts with a wizard-style interface.

## Architecture

The CLI follows a layered architecture with clear separation of concerns between command handling, business logic, and external integrations.

```mermaid
graph TB
    subgraph "CLI Layer"
        CLI[Root Command]
        CC[Cluster Commands]
        ChC[Chart Commands] 
        BC[Bootstrap Command]
        DC[Dev Commands]
    end
    
    subgraph "Service Layer"
        CS[Cluster Service]
        ChS[Chart Service]
        BS[Bootstrap Service]
        IS[Intercept Service]
        SS[Scaffold Service]
    end
    
    subgraph "Provider Layer"
        K3D[K3d Provider]
        KC[Kubectl Provider]
        HC[Helm Client]
        TP[Telepresence]
        SK[Skaffold]
    end
    
    subgraph "External Systems"
        K8S[Kubernetes Cluster]
        GH[GitHub Repositories]
        DR[Docker Registry]
    end
    
    CLI --> CC
    CLI --> ChC
    CLI --> BC
    CLI --> DC
    
    CC --> CS
    ChC --> ChS
    BC --> BS
    DC --> IS
    DC --> SS
    
    CS --> K3D
    ChS --> HC
    IS --> TP
    IS --> KC
    SS --> SK
    
    K3D --> K8S
    HC --> K8S
    TP --> K8S
    KC --> K8S
    ChS --> GH
    SS --> DR
```

## Core Components

| Component | Package | Responsibilities |
|-----------|---------|------------------|
| **Root Command** | `cmd/root.go` | CLI entry point, global flags, version management |
| **Cluster Management** | `cmd/cluster/` | K3d cluster lifecycle (create, delete, list, status, cleanup) |
| **Chart Management** | `cmd/chart/` | Helm/ArgoCD installation and management |
| **Bootstrap** | `cmd/bootstrap/` | Orchestrates cluster creation + chart installation |
| **Dev Tools** | `cmd/dev/` | Telepresence intercepts and Skaffold workflows |
| **Cluster Service** | `internal/cluster/` | Business logic for cluster operations |
| **Chart Service** | `internal/chart/` | Helm chart installation and ArgoCD setup |
| **Prerequisites** | `internal/*/prerequisites/` | Tool validation and installation |
| **UI Components** | `internal/*/ui/` | Interactive prompts and output formatting |
| **Models** | `internal/*/models/` | Data structures and configuration |

## Component Relationships

```mermaid
graph LR
    subgraph "Commands"
        ROOT[root.go]
        CLUSTER[cluster/*]
        CHART[chart/*]
        BOOTSTRAP[bootstrap/*]
        DEV[dev/*]
    end
    
    subgraph "Internal Services"
        CS[cluster/services]
        CHS[chart/services]
        BS[bootstrap/service]
        IS[dev/services/intercept]
        SS[dev/services/scaffold]
    end
    
    subgraph "Prerequisites"
        CP[cluster/prerequisites]
        CHP[chart/prerequisites]
        DP[dev/prerequisites]
    end
    
    subgraph "Shared"
        UI[shared/ui]
        ERR[shared/errors]
        EXEC[shared/executor]
        CFG[shared/config]
    end
    
    ROOT --> CLUSTER
    ROOT --> CHART
    ROOT --> BOOTSTRAP
    ROOT --> DEV
    
    CLUSTER --> CS
    CHART --> CHS
    BOOTSTRAP --> BS
    DEV --> IS
    DEV --> SS
    
    CS --> CP
    CHS --> CHP
    IS --> DP
    SS --> DP
    
    BS --> CS
    BS --> CHS
    
    CS --> UI
    CHS --> UI
    IS --> UI
    SS --> UI
    
    CS --> ERR
    CHS --> ERR
    CS --> EXEC
    CHS --> EXEC
    IS --> EXEC
    SS --> EXEC
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant ClusterSvc
    participant ChartSvc
    participant K3d
    participant Helm
    participant ArgoCD
    
    Note over User,ArgoCD: Bootstrap Workflow
    
    User->>CLI: openframe bootstrap
    CLI->>CLI: Parse flags & args
    CLI->>ClusterSvc: Create cluster
    ClusterSvc->>K3d: k3d cluster create
    K3d-->>ClusterSvc: Cluster ready
    ClusterSvc-->>CLI: Success
    
    CLI->>ChartSvc: Install charts
    ChartSvc->>ChartSvc: Generate certificates
    ChartSvc->>Helm: Install ArgoCD
    Helm-->>ChartSvc: ArgoCD installed
    ChartSvc->>Helm: Install app-of-apps
    Helm->>ArgoCD: Deploy applications
    ArgoCD-->>Helm: Applications synced
    Helm-->>ChartSvc: Success
    ChartSvc-->>CLI: Installation complete
    CLI-->>User: Bootstrap finished
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/root.go` | Main CLI entry point, command registration, version handling |
| `internal/bootstrap/service.go` | Orchestrates cluster + chart installation workflow |
| `internal/cluster/utils/service.go` | Core cluster management business logic |
| `internal/chart/services/install.go` | Helm chart installation and ArgoCD setup |
| `internal/cluster/providers/k3d/k3d.go` | K3d cluster provider implementation |
| `internal/chart/prerequisites/checker.go` | Tool validation (git, helm, memory, certificates) |
| `internal/dev/services/intercept/service.go` | Telepresence traffic interception |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/shared/ui/logo.go` | Consistent UI branding across commands |

## Dependencies

The CLI integrates with several external tools and systems:

- **K3d**: Local Kubernetes cluster creation and management
- **Helm**: Chart installation and package management  
- **kubectl**: Kubernetes cluster interaction
- **Telepresence**: Service traffic interception for development
- **Skaffold**: Live code reloading and development workflows
- **ArgoCD**: GitOps continuous deployment
- **mkcert**: Local certificate generation for HTTPS
- **Docker**: Container runtime for K3d clusters

## CLI Commands

### Cluster Management
```bash
openframe cluster create [name]     # Create new K3d cluster
openframe cluster delete [name]     # Delete cluster and cleanup
openframe cluster list             # Show all managed clusters  
openframe cluster status [name]    # Display cluster details
openframe cluster cleanup [name]   # Remove unused resources
```

### Chart Management
```bash
openframe chart install [cluster]  # Install ArgoCD and app-of-apps
```

### Bootstrap
```bash
openframe bootstrap [cluster]      # Complete setup (cluster + charts)
  --deployment-mode=oss-tenant     # Specify deployment type
  --non-interactive               # Skip prompts
```

### Development Tools
```bash
openframe dev intercept [service]  # Telepresence traffic interception
openframe dev skaffold [cluster]   # Skaffold development workflow
```

### Global Flags
- `--verbose, -v`: Enable detailed logging
- `--dry-run`: Preview actions without execution
- `--help, -h`: Show command help
