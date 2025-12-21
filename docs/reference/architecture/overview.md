# openframe-cli Module Documentation

# OpenFrame CLI Architecture Documentation

OpenFrame CLI is a modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. It provides cluster lifecycle management, chart installation with ArgoCD, and development tools for local workflows using Telepresence and Skaffold.

## Architecture

The OpenFrame CLI follows a layered architecture with clear separation between command handling, business logic, and infrastructure concerns. The system is built with modular components that can be independently tested and maintained.

### System Architecture
```mermaid
graph TB
    CLI[CLI Layer<br/>Cobra Commands] --> Services[Service Layer<br/>Business Logic]
    Services --> Internal[Internal Layer<br/>Core Components]
    Internal --> External[External Tools<br/>K3d, Helm, ArgoCD]
    
    CLI --> Models[Models<br/>Data Structures]
    CLI --> UI[UI Components<br/>User Interface]
    
    Services --> Prerequisites[Prerequisites<br/>Tool Validation]
    Services --> Errors[Error Handling<br/>Shared Utilities]
    
    Internal --> Cluster[Cluster Management]
    Internal --> Chart[Chart Management]
    Internal --> Dev[Dev Tools]
    Internal --> Bootstrap[Bootstrap Service]
    
    External --> K3d[K3d Clusters]
    External --> Helm[Helm Charts]
    External --> ArgoCD[ArgoCD Apps]
    External --> Telepresence[Telepresence]
    External --> Skaffold[Skaffold]
```

## Core Components

| Component | Package | Responsibilities |
|-----------|---------|-----------------|
| **Cluster Management** | `internal/cluster` | K3d cluster lifecycle, validation, operations |
| **Chart Management** | `internal/chart` | Helm chart installation, ArgoCD deployment |
| **Development Tools** | `internal/dev` | Telepresence intercepts, Skaffold workflows |
| **Bootstrap Service** | `internal/bootstrap` | End-to-end environment setup orchestration |
| **Shared UI** | `internal/shared/ui` | Common UI components, logo display |
| **Error Handling** | `internal/shared/errors` | Centralized error management |
| **Command Layer** | `cmd/` | Cobra command definitions and flag handling |

## Component Relationships

### Internal Dependencies
```mermaid
graph TD
    Bootstrap[Bootstrap Service] --> Cluster[Cluster Service]
    Bootstrap --> Chart[Chart Service]
    
    Cluster --> Models[Cluster Models]
    Cluster --> UI[Cluster UI]
    Cluster --> Utils[Cluster Utils]
    Cluster --> Prerequisites[Cluster Prerequisites]
    
    Chart --> ChartModels[Chart Models]
    Chart --> ChartServices[Chart Services]
    Chart --> ChartUtils[Chart Utils]
    Chart --> ChartPrereqs[Chart Prerequisites]
    
    Dev --> DevModels[Dev Models]
    Dev --> DevPrereqs[Dev Prerequisites]
    
    Chart --> SharedUI[Shared UI]
    Cluster --> SharedUI
    Dev --> SharedUI
    Bootstrap --> SharedUI
    
    Chart --> SharedErrors[Shared Errors]
    Cluster --> SharedErrors
    Dev --> SharedErrors
    Bootstrap --> SharedErrors
```

## Data Flow

### Bootstrap Workflow
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Bootstrap
    participant Cluster
    participant Chart
    participant External
    
    User->>CLI: openframe bootstrap
    CLI->>Bootstrap: Execute bootstrap
    Bootstrap->>Cluster: Create cluster
    Cluster->>External: K3d cluster create
    External-->>Cluster: Cluster ready
    Cluster-->>Bootstrap: Creation success
    Bootstrap->>Chart: Install charts
    Chart->>External: Helm install ArgoCD
    Chart->>External: Apply app-of-apps
    External-->>Chart: Installation complete
    Chart-->>Bootstrap: Charts installed
    Bootstrap-->>CLI: Bootstrap complete
    CLI-->>User: Success message
```

### Cluster Operations Flow
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Service
    participant UI
    participant K3d
    
    User->>CLI: openframe cluster create
    CLI->>UI: Show logo & wizard
    UI->>User: Configuration prompts
    User->>UI: Provide config
    UI->>Service: Cluster config
    Service->>K3d: Create cluster
    K3d-->>Service: Cluster created
    Service->>UI: Show success
    UI-->>User: Cluster ready
```

## Key Files

| File | Purpose |
|------|---------|
| `cmd/bootstrap/bootstrap.go` | Bootstrap command entry point with orchestration |
| `cmd/cluster/cluster.go` | Main cluster command with subcommands |
| `cmd/chart/chart.go` | Chart management command structure |
| `cmd/dev/dev.go` | Development tools command group |
| `internal/bootstrap/` | End-to-end environment setup service |
| `internal/cluster/services/` | Core cluster management business logic |
| `internal/chart/services/` | Chart installation and ArgoCD management |
| `internal/shared/ui/` | Reusable UI components and logo display |
| `internal/shared/errors/` | Centralized error handling utilities |

## Dependencies

The OpenFrame CLI integrates with several external tools and libraries:

### External Tool Dependencies
- **K3d**: Lightweight Kubernetes distribution for local development
- **Helm**: Kubernetes package manager for chart installation
- **ArgoCD**: GitOps continuous delivery tool for application management
- **Telepresence**: Local development tool for service intercepts
- **Skaffold**: Development workflow automation for Kubernetes

### Go Library Dependencies
- **Cobra**: Command-line interface framework for CLI structure
- **Survey**: Interactive prompts and user input handling
- **Logrus**: Structured logging throughout the application
- **YAML**: Configuration file parsing and generation

## CLI Commands

### Cluster Management Commands
```bash
openframe cluster create [name]     # Create new K3d cluster
openframe cluster list              # List all clusters  
openframe cluster status [name]     # Show cluster status
openframe cluster delete [name]     # Delete cluster
openframe cluster cleanup [name]    # Clean up resources
```

### Chart Management Commands
```bash
openframe chart install [cluster]   # Install ArgoCD and charts
```

### Development Commands
```bash
openframe dev intercept [service]   # Intercept service traffic
openframe dev scaffold [cluster]    # Run Skaffold workflows
```

### Bootstrap Commands
```bash
openframe bootstrap [cluster]       # Complete environment setup
openframe bootstrap --deployment-mode=oss-tenant  # Skip mode selection
openframe bootstrap --non-interactive --verbose   # CI/CD mode
```

### Global Flags
- `--verbose, -v`: Enable detailed logging
- `--force`: Skip confirmation prompts
- `--dry-run`: Show what would be done without executing
- `--non-interactive`: Skip all interactive prompts
