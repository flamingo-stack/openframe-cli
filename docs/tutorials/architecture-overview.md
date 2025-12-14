# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a modular, command-based architecture that would be visualized as follows:

```
┌─────────────────────────────────────────────────────────────────┐
│                        OpenFrame CLI                            │
├─────────────────────────────────────────────────────────────────┤
│  CLI Interface Layer (Cobra Commands)                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐  │
│  │   cluster   │ │    chart    │ │     dev     │ │bootstrap │  │
│  │  commands   │ │  commands   │ │  commands   │ │ commands │  │
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Business Logic Layer                                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐  │
│  │  Cluster    │ │   Chart     │ │Development  │ │Bootstrap │  │
│  │ Management  │ │ Management  │ │   Tools     │ │ Workflow │  │
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  External Integration Layer                                    │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌──────────┐  │
│  │     K3d     │ │    Helm     │ │  Skaffold   │ │  ArgoCD  │  │
│  │   Client    │ │   Client    │ │   Client    │ │  Client  │  │
│  └─────────────┘ └─────────────┘ └─────────────┘ └──────────┘  │
├─────────────────────────────────────────────────────────────────┤
│  Infrastructure Layer                                          │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐              │
│  │ Kubernetes  │ │   Docker    │ │ Local File  │              │
│  │  Clusters   │ │   Engine    │ │   System    │              │
│  └─────────────┘ └─────────────┘ └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer
**Location**: `cmd/` directory
- **Root Command**: Entry point for the CLI application
- **Cluster Commands**: Manage K3d cluster lifecycle (create, delete, start, stop, status)
- **Chart Commands**: Handle Helm chart installations and ArgoCD setup
- **Development Commands**: Integrate with development tools (Skaffold, Telepresence)
- **Bootstrap Commands**: Orchestrate full OpenFrame deployment

### 2. Business Logic Layer
**Location**: `pkg/` directory
- **Cluster Manager**: Abstracts K3d operations, handles cluster configuration
- **Chart Manager**: Manages Helm chart operations and ArgoCD integration
- **Development Tools**: Wraps Skaffold and Telepresence functionality
- **Bootstrap Orchestrator**: Coordinates multi-step deployment processes
- **Configuration Manager**: Handles CLI configuration and user preferences

### 3. External Integration Layer
**Location**: `pkg/clients/` or integrated within business logic
- **K3d Client**: Wrapper for K3d cluster operations
- **Helm Client**: Interface to Helm chart management
- **Kubectl Client**: Kubernetes API interactions
- **Development Tool Clients**: Integration with Skaffold, Telepresence

### 4. Shared Components
**Location**: `pkg/utils/`, `pkg/config/`
- **Logger**: Structured logging with configurable levels
- **Configuration**: Application settings and user preferences
- **Utilities**: Common helper functions (file operations, validation, formatting)
- **Error Handling**: Consistent error types and user-friendly messages

## Data Flow Between Components

### 1. Command Execution Flow
```
User Input → CLI Commands → Business Logic → External Tools → Infrastructure
```

### 2. Cluster Creation Flow
1. **User** executes `openframe cluster create`
2. **CLI Command** validates input and calls cluster manager
3. **Cluster Manager** prepares K3d configuration
4. **K3d Client** creates the cluster via Docker
5. **Configuration Manager** stores cluster details
6. **CLI Command** reports status back to user

### 3. Bootstrap Flow
1. **User** executes `openframe bootstrap`
2. **Bootstrap Command** orchestrates the process
3. **Chart Manager** installs required Helm charts
4. **ArgoCD Client** sets up GitOps workflows
5. **Configuration Manager** updates deployment status
6. **CLI Command** provides progress feedback

### 4. Development Workflow
1. **User** executes `openframe dev scaffold`
2. **Development Manager** validates environment
3. **Skaffold Client** initiates development mode
4. **Kubernetes Client** monitors deployment status
5. **Logger** provides real-time feedback

## Key Design Decisions and Patterns

### 1. Command Pattern
- Uses Cobra framework for clean command structure
- Each command is self-contained with its own validation and execution logic
- Supports nested subcommands for logical grouping

### 2. Dependency Injection
- Business logic components receive their dependencies as interfaces
- Enables easy testing and component swapping
- Clear separation between interface contracts and implementations

### 3. Configuration Management
- Centralized configuration handling
- Support for multiple configuration sources (files, environment variables, flags)
- Graceful defaults for common scenarios

### 4. Error Handling Strategy
- Custom error types for different categories of failures
- User-friendly error messages with actionable suggestions
- Consistent error propagation through the call stack

### 5. External Tool Integration
- Wrapper pattern for external CLI tools (K3d, Helm, Skaffold)
- Consistent interface regardless of underlying tool
- Version compatibility checks and graceful degradation

### 6. Observability
- Structured logging with contextual information
- Progress indicators for long-running operations
- Detailed status reporting for cluster and deployment states

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                    # CLI commands and entry points
│   ├── root.go            # Root command and global flags
│   ├── cluster/           # Cluster management commands
│   │   ├── create.go
│   │   ├── delete.go
│   │   ├── list.go
│   │   └── status.go
│   ├── chart/             # Chart management commands
│   │   └── install.go
│   ├── dev/               # Development workflow commands
│   │   ├── scaffold.go
│   │   └── intercept.go
│   └── bootstrap.go       # Bootstrap command
├── pkg/                   # Core business logic and utilities
│   ├── cluster/           # Cluster management logic
│   │   ├── manager.go
│   │   ├── config.go
│   │   └── status.go
│   ├── chart/             # Chart management logic
│   │   ├── helm.go
│   │   └── argocd.go
│   ├── dev/               # Development tools integration
│   │   ├── skaffold.go
│   │   └── telepresence.go
│   ├── config/            # Configuration management
│   │   ├── config.go
│   │   └── validation.go
│   ├── clients/           # External tool clients
│   │   ├── k3d.go
│   │   ├── kubectl.go
│   │   └── helm.go
│   └── utils/             # Shared utilities
│       ├── logger.go
│       ├── files.go
│       └── validation.go
├── internal/              # Private application code
│   ├── bootstrap/         # Bootstrap orchestration logic
│   └── wizard/            # Interactive CLI wizards
├── configs/               # Configuration files and templates
│   ├── cluster-templates/
│   └── chart-values/
├── docs/                  # Documentation
│   └── codewiki/
├── scripts/               # Build and deployment scripts
├── .github/               # GitHub workflows and templates
├── Makefile              # Build automation
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── main.go               # Application entry point
└── README.md             # Project overview and quick start
```

### Key Directory Explanations

- **`cmd/`**: Contains all CLI commands using the Cobra pattern. Each subdirectory represents a command group with related subcommands.

- **`pkg/`**: Houses the core business logic that can be imported by other packages. Organized by functional domain (cluster, chart, dev).

- **`internal/`**: Private packages that cannot be imported by external projects. Contains application-specific logic like bootstrap orchestration.

- **`configs/`**: Template files and default configurations for clusters and charts. Allows customization without code changes.

- **`scripts/`**: Automation scripts for building, testing, and releasing the CLI tool.

This architecture promotes maintainability, testability, and clear separation of concerns while providing a smooth developer experience for both CLI users and contributors to the codebase.