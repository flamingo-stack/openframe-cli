# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

## Overview

OpenFrame CLI is a modern, interactive command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides seamless cluster lifecycle management, chart installation with ArgoCD, and developer-friendly tools for service intercepts and scaffolding, replacing shell scripts with a robust Go-based CLI that supports both interactive and automation use cases.

## Architecture

The OpenFrame CLI follows a clean, layered architecture with clear separation of concerns between domain logic, infrastructure providers, and user interfaces.

### High-Level Architecture Diagram
```mermaid
graph TB
    CLI[CLI Layer - Cobra Commands] --> Services[Service Layer]
    Services --> Providers[Provider Layer]
    Providers --> External[External Systems]
    
    CLI --> Bootstrap[Bootstrap Service]
    CLI --> Chart[Chart Service]
    CLI --> Cluster[Cluster Service]
    CLI --> Dev[Dev Service]
    
    Bootstrap --> ClusterProviders[Cluster Providers]
    Bootstrap --> ChartProviders[Chart Providers]
    
    Chart --> Helm[Helm Provider]
    Chart --> ArgoCD[ArgoCD Provider]
    Chart --> Git[Git Provider]
    
    Cluster --> K3D[K3D Provider]
    
    Dev --> Telepresence[Telepresence Provider]
    Dev --> Kubectl[Kubectl Provider]
    Dev --> Scaffold[Scaffold Provider]
    
    Helm --> HelmBinary[Helm Binary]
    ArgoCD --> ArgoCDAPI[ArgoCD APIs]
    Git --> GitCommands[Git Commands]
    K3D --> K3DBinary[K3D Binary]
    Telepresence --> TelepresenceBinary[Telepresence Binary]
    Kubectl --> KubectlBinary[Kubectl Binary]
```

## Core Components

| Component | Package | Responsibility |
|-----------|---------|----------------|
| **CLI Commands** | `cmd/` | Cobra-based command definitions and flag handling |
| **Bootstrap Service** | `internal/bootstrap/` | Orchestrates cluster creation + chart installation |
| **Chart Services** | `internal/chart/` | ArgoCD and Helm chart installation/management |
| **Cluster Services** | `internal/cluster/` | K3D cluster lifecycle operations |
| **Dev Services** | `internal/dev/` | Development tools (intercepts, scaffolding) |
| **Shared Utilities** | `internal/shared/` | Common functionality (UI, errors, execution) |
| **Prerequisites** | `*/prerequisites/` | Tool validation and auto-installation |
| **Providers** | `*/providers/` | External system integrations |
| **UI Components** | `*/ui/` | Interactive terminal interfaces |

## Component Relationships

### Service Dependencies Diagram
```mermaid
graph TD
    Bootstrap[Bootstrap Service] --> ChartService[Chart Service]
    Bootstrap --> ClusterService[Cluster Service]
    
    ChartService --> HelmManager[Helm Manager]
    ChartService --> ArgoCDManager[ArgoCD Manager]
    ChartService --> GitRepo[Git Repository]
    
    ClusterService --> K3DManager[K3D Manager]
    
    DevService[Dev Service] --> InterceptService[Intercept Service]
    DevService --> ScaffoldService[Scaffold Service]
    
    InterceptService --> KubectlProvider[Kubectl Provider]
    InterceptService --> TelepresenceProvider[Telepresence Provider]
    
    ScaffoldService --> ChartProvider[Chart Provider]
    
    HelmManager --> Executor[Command Executor]
    ArgoCDManager --> KubernetesClient[Kubernetes Client]
    K3DManager --> Executor
    KubectlProvider --> Executor
    TelepresenceProvider --> Executor
    
    UI[UI Components] --> Services[All Services]
    Prerequisites[Prerequisites] --> Services
```

## Data Flow

### Bootstrap Command Sequence
```mermaid
sequenceDiagram
    participant User
    participant CLI as CLI Command
    participant Bootstrap as Bootstrap Service
    participant Cluster as Cluster Service
    participant Chart as Chart Service
    participant K3D as K3D Provider
    participant Helm as Helm Provider
    participant ArgoCD as ArgoCD Manager
    
    User->>CLI: openframe bootstrap
    CLI->>Bootstrap: Execute()
    
    Bootstrap->>Cluster: CreateCluster()
    Cluster->>K3D: Create k3d cluster
    K3D->>K3D: Configure networking & certificates
    K3D-->>Cluster: Return REST config
    Cluster-->>Bootstrap: Cluster ready + REST config
    
    Bootstrap->>Chart: InstallCharts(REST config)
    Chart->>Helm: Install ArgoCD
    Helm->>Helm: Apply Helm values
    Helm-->>Chart: ArgoCD installed
    
    Chart->>ArgoCD: Install app-of-apps
    ArgoCD->>ArgoCD: Clone GitHub repository
    ArgoCD->>ArgoCD: Apply manifests
    ArgoCD-->>Chart: App-of-apps deployed
    
    Chart->>ArgoCD: WaitForApplications()
    ArgoCD->>ArgoCD: Monitor sync status
    ArgoCD-->>Chart: Applications synced
    
    Chart-->>Bootstrap: Installation complete
    Bootstrap-->>CLI: Success
    CLI-->>User: Ready for use
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point and version information |
| `cmd/root.go` | Root command definition and global configuration |
| `cmd/bootstrap/bootstrap.go` | One-command cluster + chart setup |
| `internal/cluster/service.go` | Core cluster management business logic |
| `internal/chart/services/chart_service.go` | Chart installation orchestration |
| `internal/bootstrap/service.go` | Bootstrap workflow coordination |
| `internal/cluster/providers/k3d/manager.go` | K3D cluster operations |
| `internal/chart/providers/helm/manager.go` | Helm chart operations |
| `internal/chart/providers/argocd/applications.go` | ArgoCD application management |
| `internal/shared/executor/executor.go` | Command execution abstraction |
| `internal/shared/ui/prompts.go` | Interactive terminal UI |

## Dependencies

The OpenFrame CLI integrates with several external systems and tools:

### External Tool Dependencies
- **K3D**: Local Kubernetes cluster creation and management
- **Helm**: Kubernetes package management for ArgoCD installation
- **kubectl**: Kubernetes cluster interaction
- **ArgoCD**: GitOps continuous delivery
- **Telepresence**: Service mesh intercepts for development
- **Docker**: Container runtime for K3D clusters
- **Git**: Repository operations for chart sources

### Go Library Dependencies
- **Cobra**: CLI framework for commands and flags
- **pterm**: Terminal UI components and styling
- **promptui**: Interactive prompts and selection
- **client-go**: Kubernetes Go client library
- **yaml.v3**: YAML configuration parsing
- **testify**: Testing framework and assertions

### Certificate Management
The CLI includes automatic mkcert integration for local HTTPS development, generating trusted certificates for localhost access to ArgoCD and other services.

## CLI Commands

### Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `bootstrap` | Complete environment setup (cluster + charts) | `openframe bootstrap` |
| `cluster create` | Create new K3D cluster | `openframe cluster create my-cluster` |
| `cluster delete` | Remove existing cluster | `openframe cluster delete my-cluster` |
| `cluster list` | Show all clusters | `openframe cluster list` |
| `cluster status` | Detailed cluster information | `openframe cluster status my-cluster` |
| `chart install` | Install ArgoCD and app-of-apps | `openframe chart install` |
| `dev intercept` | Telepresence service intercept | `openframe dev intercept my-service` |
| `dev skaffold` | Development workflow with hot reload | `openframe dev skaffold` |

### Bootstrap Options

```bash
# Interactive mode (default)
openframe bootstrap

# Pre-configured deployment mode
openframe bootstrap --deployment-mode=oss-tenant

# Non-interactive with existing helm-values.yaml
openframe bootstrap --deployment-mode=saas-shared --non-interactive

# Verbose output with detailed logs
openframe bootstrap --verbose
```

### Cluster Management

```bash
# Interactive cluster creation with wizard
openframe cluster create

# Direct creation with defaults
openframe cluster create my-cluster --skip-wizard

# Custom configuration
openframe cluster create --nodes 5 --type k3d --version v1.31.5-k3s1

# Cluster operations
openframe cluster status my-cluster --detailed
openframe cluster cleanup my-cluster --force
```

### Development Tools

```bash
# Interactive service intercept
openframe dev intercept

# Direct service intercept
openframe dev intercept my-service --port 8080 --namespace production

# Skaffold development workflow
openframe dev skaffold my-cluster --port 3000
```

The CLI provides both interactive wizards for new users and comprehensive flag-based operation for automation and power users, making it suitable for both local development and CI/CD pipelines.
