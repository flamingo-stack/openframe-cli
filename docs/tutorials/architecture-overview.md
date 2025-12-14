# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture pattern that would be represented in a diagram as follows:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface Layer                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │   cluster   │ │    chart    │ │     dev     │ │  ...   │ │
│  │  commands   │ │  commands   │ │  commands   │ │        │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Command Processing Layer                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │   Cluster   │ │    Chart    │ │     Dev     │           │
│  │   Service   │ │   Service   │ │   Service   │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                 Infrastructure Interface Layer              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │     K3d     │ │    Helm     │ │  ArgoCD     │ │Skaffold│ │
│  │   Client    │ │   Client    │ │   Client    │ │Client  │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    External Systems                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │ Kubernetes  │ │   Docker    │ │   External  │ │  Git   │ │
│  │  Clusters   │ │   Registry  │ │   Services  │ │  Repos │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Main Components and Their Responsibilities

### 1. CLI Interface Layer
- **Root Command Handler**: Main entry point using Cobra CLI framework
- **Command Groups**: Organized command namespaces (cluster, chart, dev)
- **Flag Processing**: Handles command-line arguments and configuration
- **Output Formatting**: Consistent user interface and error handling

### 2. Command Processing Layer
- **Cluster Service**: Manages K3d cluster lifecycle operations
  - Cluster creation, deletion, starting, stopping
  - Status monitoring and health checks
  - Resource cleanup and management
- **Chart Service**: Handles Helm chart and ArgoCD operations
  - Chart installation and management
  - ArgoCD bootstrap and configuration
  - Application deployment coordination
- **Dev Service**: Development workflow tools
  - Skaffold integration for rapid development
  - Telepresence service interception
  - Local development environment setup

### 3. Infrastructure Interface Layer
- **K3d Client**: Wrapper around K3d CLI for cluster management
- **Helm Client**: Interface for Helm chart operations
- **ArgoCD Client**: Manages ArgoCD installations and applications
- **Skaffold Client**: Integrates with Skaffold for development workflows
- **System Detection**: Identifies and validates required tools

### 4. Configuration Management
- **Config Store**: Persistent configuration storage
- **Environment Detection**: System capability and tool availability
- **Default Providers**: Sensible defaults for various environments

## Data Flow Between Components

### 1. Command Execution Flow
```
User Input → CLI Parser → Command Validator → Service Layer → Infrastructure Client → External System
```

### 2. Cluster Creation Flow
```
openframe cluster create
    ↓
CLI Parser (validates flags)
    ↓
Cluster Service (processes request)
    ↓
System Detection (checks prerequisites)
    ↓
K3d Client (creates cluster)
    ↓
Status Reporter (provides feedback)
```

### 3. Bootstrap Flow
```
openframe bootstrap
    ↓
Chart Service (coordinates installation)
    ↓
Helm Client (installs charts) + ArgoCD Client (sets up GitOps)
    ↓
Cluster Service (monitors health)
    ↓
Status Reporter (installation progress)
```

### 4. Development Workflow
```
openframe dev scaffold
    ↓
Dev Service (prepares environment)
    ↓
Skaffold Client (starts development loop)
    ↓
Kubernetes API (deploys/updates services)
```

## Key Design Decisions and Patterns

### 1. **Command Pattern with Cobra Framework**
- **Decision**: Use Cobra for CLI structure
- **Rationale**: Industry standard, excellent flag handling, help generation
- **Implementation**: Hierarchical command structure with consistent interfaces

### 2. **Service Layer Abstraction**
- **Decision**: Separate business logic from CLI commands
- **Rationale**: Testability, reusability, clear separation of concerns
- **Implementation**: Service interfaces with concrete implementations

### 3. **Client Wrapper Pattern**
- **Decision**: Wrap external tools (K3d, Helm, etc.) with internal clients
- **Rationale**: Consistent error handling, easier testing, version compatibility
- **Implementation**: Interface-based clients with mock implementations for testing

### 4. **Configuration-Driven Defaults**
- **Decision**: Extensive use of sensible defaults with override capability
- **Rationale**: Developer experience, reduce cognitive load
- **Implementation**: Layered configuration (defaults → config file → flags → env vars)

### 5. **Progressive Enhancement**
- **Decision**: Core functionality works without optional tools
- **Rationale**: Better user experience, graceful degradation
- **Implementation**: Feature detection and conditional execution

### 6. **Structured Logging and Output**
- **Decision**: Consistent output formatting and progress indication
- **Rationale**: Professional UX, debuggability
- **Implementation**: Structured logging with user-friendly progress bars

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                          # CLI command definitions
│   ├── root.go                   # Root command and global flags
│   ├── cluster/                  # Cluster management commands
│   │   ├── create.go
│   │   ├── list.go
│   │   ├── status.go
│   │   └── delete.go
│   ├── chart/                    # Chart management commands
│   │   └── install.go
│   ├── dev/                      # Development workflow commands
│   │   ├── scaffold.go
│   │   └── intercept.go
│   └── bootstrap.go              # Bootstrap command
│
├── internal/                     # Private application code
│   ├── services/                 # Business logic layer
│   │   ├── cluster.go            # Cluster management service
│   │   ├── chart.go              # Chart management service
│   │   └── dev.go                # Development service
│   │
│   ├── clients/                  # External tool integrations
│   │   ├── k3d.go                # K3d cluster client
│   │   ├── helm.go               # Helm client wrapper
│   │   ├── argocd.go             # ArgoCD client
│   │   └── skaffold.go           # Skaffold integration
│   │
│   ├── config/                   # Configuration management
│   │   ├── config.go             # Configuration structures
│   │   ├── defaults.go           # Default values
│   │   └── validation.go         # Configuration validation
│   │
│   ├── utils/                    # Shared utilities
│   │   ├── system.go             # System detection
│   │   ├── output.go             # Output formatting
│   │   └── errors.go             # Error handling
│   │
│   └── types/                    # Shared data structures
│       ├── cluster.go            # Cluster-related types
│       └── chart.go              # Chart-related types
│
├── pkg/                          # Public API (if any)
├── docs/                         # Documentation
│   └── codewiki/
│       └── overview.md
├── scripts/                      # Build and development scripts
├── .github/                      # GitHub workflows
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
├── main.go                       # Application entry point
├── Makefile                      # Build automation
└── README.md                     # Project documentation
```

### Directory Explanation

- **`cmd/`**: Contains all CLI command definitions using Cobra pattern. Each command group has its own subdirectory.
- **`internal/`**: Private application code that cannot be imported by other projects.
  - **`services/`**: Business logic layer that orchestrates operations
  - **`clients/`**: Wrappers around external tools and APIs
  - **`config/`**: Configuration management and validation
  - **`utils/`**: Shared utility functions and helpers
  - **`types/`**: Common data structures used across the application
- **`pkg/`**: Public packages that could be imported by other projects (currently minimal)
- **`docs/`**: Project documentation including architecture and API docs
- **`scripts/`**: Build, release, and development automation scripts

This structure follows Go best practices and provides clear separation between public interfaces, business logic, and infrastructure concerns, making it easy for new developers to understand and contribute to the codebase.