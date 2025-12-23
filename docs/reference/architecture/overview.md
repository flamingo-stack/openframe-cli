# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern, interactive command-line tool for managing Kubernetes clusters and development workflows specifically designed for the OpenFrame platform. It provides cluster lifecycle management (create, delete, list, status), chart installation with ArgoCD, and development tools for service intercepts and scaffolding, replacing traditional shell scripts with a polished Go-based CLI.

## Architecture

### High-Level System Design

```mermaid
graph TB
    CLI[CLI Layer] --> Bootstrap[Bootstrap Service]
    CLI --> Cluster[Cluster Commands]
    CLI --> Chart[Chart Commands]
    CLI --> Dev[Dev Commands]
    
    Bootstrap --> ClusterSvc[Cluster Service]
    Bootstrap --> ChartSvc[Chart Service]
    
    Cluster --> ClusterSvc
    Chart --> ChartSvc
    Dev --> DevSvc[Dev Service]
    
    ClusterSvc --> K3D[K3D Provider]
    ClusterSvc --> Prerequisites[Prerequisites Checker]
    
    ChartSvc --> Helm[Helm Manager]
    ChartSvc --> ArgoCD[ArgoCD Provider]
    ChartSvc --> Git[Git Repository]
    
    DevSvc --> Telepresence[Telepresence Provider]
    DevSvc --> Skaffold[Skaffold Provider]
    DevSvc --> Kubectl[Kubectl Provider]
    
    K3D --> Docker[Docker/K3D]
    Helm --> K8s[Kubernetes Cluster]
    Telepresence --> K8s
    Kubectl --> K8s
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|---------------|
| **CLI Commands** | `cmd/` | Cobra command definitions and argument parsing |
| **Bootstrap Service** | `internal/bootstrap/` | Orchestrates cluster creation + chart installation |
| **Cluster Management** | `internal/cluster/` | K3D cluster lifecycle operations |
| **Chart Installation** | `internal/chart/` | ArgoCD and Helm chart management |
| **Development Tools** | `internal/dev/` | Telepresence intercepts and Skaffold workflows |
| **Shared Infrastructure** | `internal/shared/` | Common utilities, UI, errors, and execution |
| **Prerequisites** | `*/prerequisites/` | Tool installation and validation |
| **Providers** | `*/providers/` | External tool integrations (K3D, Helm, Git) |

## Component Relationships

```mermaid
graph TB
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
        InterceptSvc[Intercept Service]
        ScaffoldSvc[Scaffold Service]
    end
    
    subgraph "Provider Layer"
        K3DProvider[K3D Manager]
        HelmProvider[Helm Manager]
        ArgoCDProvider[ArgoCD Manager]
        GitProvider[Git Repository]
        TelepresenceProvider[Telepresence Provider]
        KubectlProvider[Kubectl Provider]
    end
    
    subgraph "Shared Infrastructure"
        UI[UI Components]
        Executor[Command Executor]
        ErrorHandler[Error Handling]
        Config[Configuration]
    end
    
    BootstrapCmd --> BootstrapSvc
    ClusterCmd --> ClusterSvc
    ChartCmd --> ChartSvc
    DevCmd --> InterceptSvc
    DevCmd --> ScaffoldSvc
    
    BootstrapSvc --> ClusterSvc
    BootstrapSvc --> ChartSvc
    
    ClusterSvc --> K3DProvider
    ChartSvc --> HelmProvider
    ChartSvc --> ArgoCDProvider
    ChartSvc --> GitProvider
    InterceptSvc --> TelepresenceProvider
    InterceptSvc --> KubectlProvider
    ScaffoldSvc --> KubectlProvider
    
    ClusterSvc --> UI
    ChartSvc --> UI
    InterceptSvc --> UI
    ScaffoldSvc --> UI
    
    K3DProvider --> Executor
    HelmProvider --> Executor
    TelepresenceProvider --> Executor
    KubectlProvider --> Executor
```

## Data Flow

### Cluster Creation and Chart Installation Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Bootstrap
    participant Cluster
    participant Chart
    participant K3D
    participant Helm
    participant ArgoCD
    
    User->>CLI: openframe bootstrap
    CLI->>Bootstrap: Execute()
    Bootstrap->>Cluster: CreateCluster()
    Cluster->>K3D: Create k3d cluster
    K3D-->>Cluster: Cluster ready
    Cluster-->>Bootstrap: Success
    
    Bootstrap->>Chart: InstallCharts()
    Chart->>Helm: Install ArgoCD
    Helm-->>Chart: ArgoCD ready
    Chart->>ArgoCD: Install app-of-apps
    ArgoCD->>ArgoCD: Sync applications
    ArgoCD-->>Chart: Applications synced
    Chart-->>Bootstrap: Charts installed
    
    Bootstrap-->>CLI: Complete
    CLI-->>User: Success message
```

### Development Intercept Flow

```mermaid
sequenceDiagram
    participant Dev
    participant CLI
    participant Intercept
    participant Kubectl
    participant Telepresence
    participant K8s
    
    Dev->>CLI: openframe dev intercept
    CLI->>Intercept: StartIntercept()
    Intercept->>Kubectl: Find service
    Kubectl->>K8s: Query services
    K8s-->>Kubectl: Service info
    Kubectl-->>Intercept: Service found
    
    Intercept->>Telepresence: Connect to cluster
    Telepresence-->>Intercept: Connected
    Intercept->>Telepresence: Create intercept
    Telepresence->>K8s: Route traffic
    K8s-->>Telepresence: Traffic routed
    Telepresence-->>Intercept: Intercept active
    
    Intercept-->>CLI: Intercept running
    CLI-->>Dev: Traffic intercepted
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Application entry point and version handling |
| `cmd/root.go` | Root command definition and CLI structure |
| `cmd/bootstrap/bootstrap.go` | Bootstrap command for complete setup |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/chart/services/chart_service.go` | Chart installation orchestration |
| `internal/dev/services/intercept/service.go` | Telepresence intercept management |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/shared/ui/logo.go` | CLI branding and visual presentation |
| `internal/cluster/providers/k3d/manager.go` | K3D cluster provider implementation |
| `internal/chart/providers/helm/manager.go` | Helm chart operations |

## Dependencies

The project integrates with several external tools and libraries:

### External Tool Dependencies
- **Docker**: Container runtime for K3D clusters
- **K3D**: Lightweight Kubernetes distribution for local development
- **kubectl**: Kubernetes command-line tool for cluster interaction
- **Helm**: Package manager for Kubernetes applications
- **Telepresence**: Service intercept and local development
- **Skaffold**: Continuous development for Kubernetes applications
- **jq**: JSON processing for parsing command outputs

### Go Library Dependencies
- **Cobra**: CLI framework for command structure and parsing
- **pterm**: Terminal styling and interactive prompts
- **promptui**: Enhanced user input and selection interfaces
- **testify**: Testing utilities and assertions
- **yaml.v3**: YAML parsing for configuration files

### Integration Patterns
The CLI uses a provider pattern to abstract external tool interactions, making it testable and maintainable. Each provider implements specific interfaces for their domain (cluster management, chart operations, development tools).

## CLI Commands

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `openframe bootstrap` | Complete cluster setup (create + charts) | `openframe bootstrap my-cluster` |
| `openframe cluster create` | Create a new K3D cluster | `openframe cluster create dev-cluster` |
| `openframe cluster list` | List all managed clusters | `openframe cluster list` |
| `openframe cluster status` | Show cluster details | `openframe cluster status my-cluster` |
| `openframe cluster delete` | Remove a cluster | `openframe cluster delete my-cluster` |
| `openframe chart install` | Install ArgoCD and applications | `openframe chart install my-cluster` |
| `openframe dev intercept` | Intercept service traffic | `openframe dev intercept my-service --port 8080` |
| `openframe dev skaffold` | Run development workflow | `openframe dev skaffold my-cluster` |

### Bootstrap Workflow

```bash
# Complete setup with interactive configuration
openframe bootstrap

# Non-interactive with specific deployment mode
openframe bootstrap --deployment-mode=oss-tenant --non-interactive

# Verbose mode with custom cluster name
openframe bootstrap my-dev-cluster --verbose
```

### Development Workflow

```bash
# Start traffic intercept for a service
openframe dev intercept api-service --port 8080 --namespace production

# Run Skaffold development environment
openframe dev skaffold --skip-bootstrap

# List available clusters for development
openframe cluster list --quiet
```
