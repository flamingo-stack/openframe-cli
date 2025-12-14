# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line interface tool designed for managing Kubernetes clusters and OpenFrame development workflows. It provides cluster lifecycle management (create, delete, start, status), prerequisite installation (Docker, kubectl, k3d), and chart installation capabilities for bootstrapping OpenFrame environments on local K3D clusters.

## Architecture

```mermaid
graph TB
    CLI[CLI Commands] --> Bootstrap[Bootstrap Service]
    CLI --> Cluster[Cluster Commands]
    CLI --> Chart[Chart Commands]
    
    Bootstrap --> ClusterSvc[Cluster Service]
    Bootstrap --> ChartSvc[Chart Service]
    
    Cluster --> ClusterSvc
    Chart --> ChartSvc
    
    ClusterSvc --> K3D[K3D Provider]
    ClusterSvc --> Prerequisites[Prerequisites]
    
    ChartSvc --> Helm[Helm Operations]
    ChartSvc --> ArgoCD[ArgoCD Installation]
    
    K3D --> Executor[Command Executor]
    Prerequisites --> Docker[Docker Installer]
    Prerequisites --> Kubectl[Kubectl Installer]
    Prerequisites --> K3DInstaller[K3D Installer]
    
    UI[UI Components] --> ClusterSvc
    UI --> ChartSvc
    
    Shared[Shared Components] --> Executor
    Shared --> Errors[Error Handling]
    Shared --> Flags[Flag Management]
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| Bootstrap Service | `internal/bootstrap` | Orchestrates cluster creation + chart installation |
| Cluster Service | `internal/cluster` | Manages K3D cluster lifecycle operations |
| Chart Service | `internal/chart` | Handles Helm chart and ArgoCD installations |
| K3D Provider | `internal/cluster/providers/k3d` | K3D-specific cluster operations |
| Prerequisites | `internal/cluster/prerequisites` | System dependency installation |
| Command Executor | `internal/shared/executor` | Abstracted command execution interface |
| UI Components | `internal/cluster/ui`, `internal/shared/ui` | User interface and interaction logic |
| Flag Management | `internal/shared/flags` | Common CLI flag handling |
| Error Handling | `internal/shared/errors` | Centralized error processing |

## Component Relationships

```mermaid
graph LR
    subgraph "CLI Layer"
        CMD[Commands]
    end
    
    subgraph "Service Layer"
        Bootstrap[Bootstrap Service]
        ClusterSvc[Cluster Service]
        ChartSvc[Chart Service]
    end
    
    subgraph "Provider Layer"
        K3D[K3D Manager]
        Prerequisites[Prerequisites]
        Helm[Helm Operations]
    end
    
    subgraph "Infrastructure Layer"
        Executor[Command Executor]
        UI[UI Components]
        Shared[Shared Utils]
    end
    
    CMD --> Bootstrap
    CMD --> ClusterSvc
    CMD --> ChartSvc
    
    Bootstrap --> ClusterSvc
    Bootstrap --> ChartSvc
    
    ClusterSvc --> K3D
    ClusterSvc --> Prerequisites
    ChartSvc --> Helm
    
    K3D --> Executor
    Prerequisites --> Executor
    Helm --> Executor
    
    ClusterSvc --> UI
    ChartSvc --> UI
    
    K3D --> Shared
    Prerequisites --> Shared
```

## Data Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Bootstrap
    participant ClusterSvc
    participant K3D
    participant ChartSvc
    participant Executor
    
    User->>CLI: openframe bootstrap
    CLI->>Bootstrap: Execute with flags
    Bootstrap->>ClusterSvc: CreateCluster(config)
    ClusterSvc->>K3D: CreateCluster(ctx, config)
    K3D->>Executor: Execute k3d commands
    Executor->>K3D: Command result
    K3D->>ClusterSvc: Success/Error
    ClusterSvc->>Bootstrap: Cluster ready
    
    Bootstrap->>ChartSvc: InstallCharts(config)
    ChartSvc->>Executor: Execute helm commands
    Executor->>ChartSvc: Installation result
    ChartSvc->>Bootstrap: Charts installed
    
    Bootstrap->>CLI: Complete workflow
    CLI->>User: Success message + next steps
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/bootstrap/service.go` | Main bootstrap orchestration logic |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/cluster/providers/k3d/manager.go` | K3D cluster provider implementation |
| `internal/cluster/prerequisites/checker.go` | System prerequisite validation |
| `internal/chart/services/install.go` | Chart installation workflow |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/cluster/models/cluster.go` | Domain models for cluster operations |
| `internal/cluster/ui/wizard.go` | Interactive cluster configuration |

## Dependencies

The project uses several key external dependencies:

- **Cobra**: CLI framework for command structure and flag management
- **pterm**: Terminal UI library for progress indicators, tables, and colored output
- **promptui**: Interactive prompts for user input during wizards
- **testify**: Testing framework with assertions and mocking capabilities

The architecture abstracts external tool dependencies (Docker, kubectl, k3d, helm) through the Command Executor interface, enabling testability and consistent error handling across different operating systems.

## CLI Commands

| Command | Description | Usage |
|---------|-------------|-------|
| `openframe bootstrap` | Complete cluster + chart setup | `openframe bootstrap [cluster-name] --deployment-mode=oss-tenant` |
| `openframe cluster create` | Create new K3D cluster | `openframe cluster create [name] --nodes=3 --version=latest` |
| `openframe cluster list` | List all clusters | `openframe cluster list [--quiet]` |
| `openframe cluster status` | Show cluster details | `openframe cluster status [name] [--detailed]` |
| `openframe cluster delete` | Remove cluster | `openframe cluster delete [name] [--force]` |
| `openframe cluster start` | Start stopped cluster | `openframe cluster start [name]` |
| `openframe cluster cleanup` | Clean up resources | `openframe cluster cleanup [name] [--force]` |
| `openframe chart install` | Install Helm charts | `openframe chart install [cluster] --github-repo=URL --github-branch=main` |

### Global Flags

- `--verbose, -v`: Enable detailed output
- `--dry-run`: Show what would be executed without running commands
- `--force`: Skip confirmation prompts
- `--non-interactive`: Disable interactive prompts for automation
