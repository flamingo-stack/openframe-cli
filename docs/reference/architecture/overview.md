# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides interactive cluster creation, Helm chart management with ArgoCD integration, and developer tools for traffic interception and service scaffolding. The CLI replaces shell scripts with a modern, wizard-style interactive terminal interface for Kubernetes platform bootstrapping.

## Architecture

The CLI follows a clean architecture pattern with clear separation of concerns across domain-driven layers. It uses the Cobra command framework with dependency injection for testing and modular command organization.

### High-Level System Architecture
```mermaid
graph TB
    CLI[CLI Entry Point] --> Bootstrap[Bootstrap Command]
    CLI --> Cluster[Cluster Commands]
    CLI --> Chart[Chart Commands]
    CLI --> Dev[Dev Commands]
    
    Cluster --> ClusterService[Cluster Service]
    Chart --> ChartService[Chart Service]
    Dev --> DevService[Dev Service]
    Bootstrap --> ClusterService
    Bootstrap --> ChartService
    
    ClusterService --> K3D[K3D Provider]
    ChartService --> Helm[Helm Manager]
    ChartService --> ArgoCD[ArgoCD Manager]
    DevService --> Telepresence[Telepresence Provider]
    DevService --> Skaffold[Skaffold Provider]
    
    K3D --> Docker[Docker/K3D]
    Helm --> K8s[Kubernetes]
    ArgoCD --> K8s
    Telepresence --> K8s
    Skaffold --> K8s
```

## Core Components

| Component | Package | Responsibilities |
|-----------|---------|------------------|
| **CLI Framework** | `cmd/*` | Command definitions, argument parsing, flag handling |
| **Cluster Management** | `internal/cluster` | K3D cluster lifecycle (create, delete, list, status, cleanup) |
| **Chart Management** | `internal/chart` | ArgoCD and Helm chart installation, app-of-apps pattern |
| **Development Tools** | `internal/dev` | Telepresence intercepts, Skaffold scaffolding workflows |
| **Bootstrap Orchestration** | `internal/bootstrap` | Combines cluster + chart operations for complete setup |
| **Shared Infrastructure** | `internal/shared` | Command execution, UI components, error handling, configuration |
| **Prerequisites** | `*/prerequisites` | Tool installation and validation (Docker, k3d, kubectl, helm, etc.) |

## Component Relationships

### Module Dependencies
```mermaid
graph TB
    subgraph "Command Layer"
        CMD[cmd/*]
        Bootstrap[bootstrap]
        Cluster[cluster]
        Chart[chart] 
        Dev[dev]
    end
    
    subgraph "Service Layer"
        ClusterSvc[cluster/service]
        ChartSvc[chart/services]
        DevSvc[dev/services]
    end
    
    subgraph "Provider Layer"
        K3D[cluster/providers/k3d]
        Helm[chart/providers/helm]
        ArgoCD[chart/providers/argocd]
        Telepresence[dev/providers/telepresence]
        Kubectl[dev/providers/kubectl]
    end
    
    subgraph "Shared Infrastructure"
        Executor[shared/executor]
        UI[shared/ui]
        Config[shared/config]
        Errors[shared/errors]
    end
    
    CMD --> Bootstrap
    CMD --> Cluster
    CMD --> Chart
    CMD --> Dev
    
    Bootstrap --> ClusterSvc
    Bootstrap --> ChartSvc
    Cluster --> ClusterSvc
    Chart --> ChartSvc
    Dev --> DevSvc
    
    ClusterSvc --> K3D
    ChartSvc --> Helm
    ChartSvc --> ArgoCD
    DevSvc --> Telepresence
    DevSvc --> Kubectl
    
    K3D --> Executor
    Helm --> Executor
    ArgoCD --> Executor
    Telepresence --> Executor
    
    ClusterSvc --> UI
    ChartSvc --> UI
    DevSvc --> UI
    
    K3D --> Config
    Helm --> Config
    ArgoCD --> Config
```

## Data Flow

### Bootstrap Command Flow
```mermaid
sequenceDiagram
    participant User
    participant Bootstrap
    participant ClusterSvc
    participant ChartSvc
    participant K3D
    participant ArgoCD
    
    User->>Bootstrap: openframe bootstrap
    Bootstrap->>ClusterSvc: CreateCluster()
    ClusterSvc->>K3D: CreateCluster(config)
    K3D-->>ClusterSvc: cluster ready
    ClusterSvc-->>Bootstrap: cluster created
    
    Bootstrap->>ChartSvc: InstallCharts()
    ChartSvc->>ArgoCD: InstallArgoCD()
    ArgoCD-->>ChartSvc: ArgoCD ready
    ChartSvc->>ArgoCD: InstallAppOfApps()
    ArgoCD-->>ChartSvc: apps installed
    ChartSvc->>ArgoCD: WaitForApplications()
    ArgoCD-->>ChartSvc: all apps synced
    ChartSvc-->>Bootstrap: charts ready
    
    Bootstrap-->>User: complete environment
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point and version handling |
| `cmd/root.go` | Root command structure and global configuration |
| `cmd/bootstrap/bootstrap.go` | Combined cluster+chart workflow orchestration |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/chart/services/chart_service.go` | Chart installation orchestration with ArgoCD |
| `internal/dev/services/intercept/service.go` | Telepresence traffic interception workflows |
| `internal/shared/executor/executor.go` | Command execution abstraction with mock support |
| `internal/cluster/providers/k3d/manager.go` | K3D-specific cluster operations |
| `internal/chart/providers/argocd/wait.go` | ArgoCD application synchronization monitoring |

## Dependencies

The project uses several external libraries to provide its functionality:

- **Cobra** (`github.com/spf13/cobra`): CLI framework providing command structure, argument parsing, and help generation
- **pterm** (`github.com/pterm/pterm`): Terminal UI library for progress indicators, tables, interactive prompts, and styled output
- **promptui** (`github.com/manifoldco/promptui`): Interactive CLI prompts for user input and selection menus
- **YAML** (`gopkg.in/yaml.v3`): YAML parsing for Helm values files and Kubernetes manifests
- **testify** (`github.com/stretchr/testify`): Testing assertions and mocking framework
- **golang.org/x/term**: Terminal interface detection and raw input handling for cross-platform compatibility

The CLI integrates with external tools through command execution:
- **k3d**: Lightweight Kubernetes distribution management
- **kubectl**: Kubernetes cluster interaction
- **helm**: Kubernetes package management
- **telepresence**: Service mesh traffic interception
- **skaffold**: Kubernetes development workflows

## CLI Commands

### Cluster Management
```bash
# Create a new K3D cluster
openframe cluster create [NAME]
openframe cluster create --nodes 3 --type k3d --skip-wizard

# List all clusters
openframe cluster list

# Show cluster status
openframe cluster status [NAME]

# Delete a cluster
openframe cluster delete [NAME] --force

# Clean up cluster resources
openframe cluster cleanup [NAME]
```

### Chart Installation
```bash
# Install ArgoCD and app-of-apps
openframe chart install [CLUSTER]
openframe chart install --deployment-mode=oss-tenant --non-interactive

# Bootstrap complete environment (cluster + charts)
openframe bootstrap [NAME]
openframe bootstrap --deployment-mode=saas-shared --verbose
```

### Development Tools
```bash
# Intercept service traffic to local development
openframe dev intercept [SERVICE] --port 8080 --namespace default

# Run Skaffold development workflow
openframe dev skaffold [CLUSTER] --skip-bootstrap
```

### Global Options
```bash
# Common flags available on all commands
--verbose, -v    # Enable detailed output
--dry-run        # Show what would be done without executing
--non-interactive # Skip all prompts for CI/CD usage
```
