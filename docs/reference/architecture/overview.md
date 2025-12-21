# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a Kubernetes cluster management tool that provides comprehensive lifecycle management for local development clusters. It offers streamlined workflows for creating K3d clusters, installing ArgoCD charts, and managing development environments with traffic interception capabilities through a unified command-line interface.

## Architecture

### High-Level System Design

```mermaid
graph TB
    CLI[OpenFrame CLI] --> Bootstrap[Bootstrap Module]
    CLI --> Cluster[Cluster Module] 
    CLI --> Chart[Chart Module]
    CLI --> Dev[Dev Module]
    
    Bootstrap --> ClusterCreate[Cluster Creation]
    Bootstrap --> ChartInstall[Chart Installation]
    
    Cluster --> K3d[K3d Provider]
    Cluster --> Prerequisites[Prerequisites Check]
    
    Chart --> ArgoCD[ArgoCD Installation]
    Chart --> AppOfApps[App-of-Apps Setup]
    
    Dev --> Telepresence[Traffic Interception]
    Dev --> Skaffold[Development Workflows]
    
    K3d --> Docker[Docker Engine]
    ArgoCD --> Kubernetes[Kubernetes Cluster]
    Telepresence --> Kubernetes
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **Bootstrap** | `cmd/bootstrap` | Orchestrates complete OpenFrame environment setup |
| **Cluster Management** | `cmd/cluster` | K3d cluster lifecycle operations (create, delete, list, status, cleanup) |
| **Chart Management** | `cmd/chart` | ArgoCD and app-of-apps installation and configuration |
| **Development Tools** | `cmd/dev` | Traffic interception and development workflow management |
| **UI Layer** | `internal/*/ui` | Interactive prompts, configuration wizards, and operation feedback |
| **Service Layer** | `internal/*/services` | Business logic and orchestration between providers |
| **Models** | `internal/*/models` | Data structures, validation, and flag definitions |
| **Prerequisites** | `internal/*/prerequisites` | Dependency validation and installation |

## Component Relationships

### Module Dependencies

```mermaid
graph LR
    Bootstrap --> ClusterServices[Cluster Services]
    Bootstrap --> ChartServices[Chart Services]
    
    ClusterCmd[Cluster Commands] --> ClusterUI[Cluster UI]
    ClusterCmd --> ClusterModels[Cluster Models]
    ClusterCmd --> ClusterUtils[Cluster Utils]
    
    ChartCmd[Chart Commands] --> ChartServices
    ChartCmd --> ChartPrereq[Chart Prerequisites]
    
    DevCmd[Dev Commands] --> DevModels[Dev Models]
    DevCmd --> DevPrereq[Dev Prerequisites]
    
    ClusterUI --> ClusterModels
    ClusterUtils --> ClusterModels
    
    SharedErrors[Shared Errors] --> ClusterCmd
    SharedErrors --> ChartCmd
    SharedUI[Shared UI] --> ClusterCmd
    SharedUI --> ChartCmd
```

## Data Flow

### Bootstrap Command Execution Flow

```mermaid
sequenceDiagram
    participant User
    participant Bootstrap
    participant ClusterSvc as Cluster Service
    participant ChartSvc as Chart Service
    participant K3d
    participant ArgoCD
    
    User->>Bootstrap: bootstrap [cluster-name]
    Bootstrap->>Bootstrap: Parse deployment mode flags
    Bootstrap->>ClusterSvc: Create cluster
    ClusterSvc->>K3d: k3d cluster create
    K3d-->>ClusterSvc: Cluster ready
    ClusterSvc-->>Bootstrap: Cluster created
    
    Bootstrap->>ChartSvc: Install charts
    ChartSvc->>ArgoCD: Install ArgoCD helm chart
    ArgoCD-->>ChartSvc: ArgoCD installed
    ChartSvc->>ArgoCD: Deploy app-of-apps
    ArgoCD-->>ChartSvc: Apps deployed
    ChartSvc-->>Bootstrap: Charts installed
    Bootstrap-->>User: Environment ready
```

### Cluster Operation Flow

```mermaid
sequenceDiagram
    participant User
    participant ClusterCmd as Cluster Command
    participant OperationsUI as Operations UI
    participant ClusterSvc as Cluster Service
    participant Provider as K3d Provider
    
    User->>ClusterCmd: cluster create [name]
    ClusterCmd->>OperationsUI: Get cluster configuration
    OperationsUI->>User: Interactive wizard or flags
    User-->>OperationsUI: Configuration choices
    OperationsUI-->>ClusterCmd: ClusterConfig
    
    ClusterCmd->>ClusterSvc: CreateCluster(config)
    ClusterSvc->>Provider: Create K3d cluster
    Provider-->>ClusterSvc: Cluster created
    ClusterSvc-->>ClusterCmd: Success
    ClusterCmd-->>User: Cluster ready
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bootstrap/bootstrap.go` | Main bootstrap command orchestrating complete environment setup |
| `cmd/cluster/cluster.go` | Cluster command root with subcommand registration and global flags |
| `cmd/cluster/create.go` | Cluster creation with interactive configuration and validation |
| `cmd/chart/chart.go` | Chart management root command with ArgoCD installation capabilities |
| `cmd/chart/install.go` | ArgoCD and app-of-apps installation with configuration management |
| `cmd/dev/dev.go` | Development tools root for traffic interception and workflows |

## Dependencies

The OpenFrame CLI integrates with several external tools and libraries:

### Core Dependencies
- **Cobra CLI Framework**: Command structure, flag parsing, and subcommand organization
- **Kubernetes Client Libraries**: Cluster communication and resource management
- **Helm SDK**: Chart installation and configuration management
- **K3d**: Local Kubernetes cluster provisioning and management

### Development Tools
- **Telepresence**: Traffic interception for local development workflows
- **Skaffold**: Continuous development and deployment automation
- **ArgoCD**: GitOps-based application deployment and synchronization

### Infrastructure
- **Docker**: Container runtime for K3d cluster nodes
- **GitHub Integration**: Repository cloning and branch management for app-of-apps

## CLI Commands

### Bootstrap Commands
```bash
# Complete environment setup
openframe bootstrap                                    # Interactive mode
openframe bootstrap my-cluster                        # Custom cluster name
openframe bootstrap --deployment-mode=oss-tenant     # Skip deployment selection
openframe bootstrap --non-interactive --verbose      # CI/CD mode
```

### Cluster Management Commands
```bash
# Cluster lifecycle
openframe cluster create                    # Interactive cluster creation
openframe cluster create my-cluster        # Create with custom name
openframe cluster delete my-cluster        # Delete specific cluster
openframe cluster list                     # List all clusters
openframe cluster status my-cluster        # Show cluster details
openframe cluster cleanup my-cluster       # Clean unused resources
```

### Chart Management Commands
```bash
# ArgoCD and app-of-apps installation
openframe chart install                              # Interactive installation
openframe chart install my-cluster                  # Install on specific cluster
openframe chart install --deployment-mode=saas-tenant  # Pre-configured mode
openframe chart install --github-branch develop     # Custom branch
```

### Development Commands
```bash
# Development workflows (planned)
openframe dev intercept my-service         # Traffic interception
openframe dev skaffold my-cluster          # Development deployment
```

### Global Flags
- `--verbose, -v`: Enable detailed logging and operation progress
- `--deployment-mode`: Specify deployment type (oss-tenant, saas-tenant, saas-shared)
- `--non-interactive`: Skip all prompts for CI/CD environments
- `--force`: Skip confirmation prompts for destructive operations
- `--dry-run`: Preview operations without execution
