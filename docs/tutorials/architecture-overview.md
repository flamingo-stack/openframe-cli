# Architecture Overview

This document provides a comprehensive overview of the OpenFrame CLI architecture, designed to help new developers understand the codebase structure and design decisions.

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture pattern that would be visualized as:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface Layer                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   Cobra     │ │   Viper     │ │     Interactive UI      │ │
│  │  Commands   │ │   Config    │ │    (Prompts/Tables)     │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                 Command Orchestration Layer                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │  Cluster    │ │    Chart    │ │      Development        │ │
│  │  Commands   │ │  Commands   │ │       Commands          │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Core Services Layer                       │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │   K3d API   │ │   Helm      │ │      ArgoCD             │ │
│  │  Wrapper    │ │  Manager    │ │     Manager             │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                External Dependencies Layer                  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────────┐ │
│  │    K3d      │ │  Kubernetes │ │        Docker           │ │
│  │   Binary    │ │     API     │ │        Engine           │ │
│  └─────────────┘ └─────────────┘ └─────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer

**Cobra Command Framework**
- Manages command structure and argument parsing
- Provides help text and command validation
- Handles global flags and configuration

**Viper Configuration Management**
- Manages configuration file loading and parsing
- Handles environment variable overrides
- Provides configuration validation

**Interactive UI Components**
- User prompts for guided workflows
- Progress indicators and status displays
- Formatted table output for cluster information

### 2. Command Orchestration Layer

**Cluster Commands (`cmd/cluster/`)**
- `create`: Orchestrates cluster creation workflow
- `list`: Displays available clusters
- `status`: Shows detailed cluster information
- `delete`: Handles cluster cleanup
- `start/stop`: Manages cluster lifecycle

**Chart Commands (`cmd/chart/`)**
- `install`: Manages Helm chart installations
- Coordinates ArgoCD setup and configuration
- Handles chart dependency resolution

**Development Commands (`cmd/dev/`)**
- `scaffold`: Integrates with Skaffold for development workflows
- `intercept`: Manages Telepresence traffic interception
- Provides developer productivity tools

### 3. Core Services Layer

**K3d API Wrapper**
- Abstracts K3d cluster operations
- Provides consistent error handling
- Manages cluster configuration and networking

**Helm Manager**
- Handles chart installation and upgrades
- Manages repository configuration
- Provides rollback capabilities

**ArgoCD Manager**
- Configures ArgoCD for GitOps workflows
- Manages application deployments
- Handles synchronization and health checks

### 4. External Dependencies

**K3d Binary**
- Lightweight Kubernetes distribution
- Container-based cluster management
- Local development environment

**Kubernetes API**
- Standard Kubernetes client operations
- Resource management and monitoring
- Service discovery and networking

**Docker Engine**
- Container runtime for K3d
- Image management and registry operations
- Network and volume management

## Data Flow Between Components

### Cluster Creation Flow

1. **User Input** → CLI parses command and flags
2. **Configuration Loading** → Viper loads config files and environment variables
3. **Interactive Prompts** → UI layer gathers missing configuration
4. **Validation** → Command layer validates inputs and system requirements
5. **K3d Execution** → Core services layer calls K3d APIs
6. **Status Monitoring** → Real-time feedback to user interface
7. **Resource Setup** → Additional cluster configuration (networking, storage)
8. **Confirmation** → Final status report to user

### Chart Installation Flow

1. **Dependency Check** → Verify cluster availability and Helm installation
2. **Repository Setup** → Configure Helm repositories
3. **Chart Resolution** → Resolve chart dependencies and versions
4. **Value Configuration** → Apply custom values and overrides
5. **Installation** → Deploy charts to Kubernetes cluster
6. **ArgoCD Setup** → Configure GitOps workflows if specified
7. **Health Checks** → Monitor deployment status and readiness

## Key Design Decisions and Patterns

### 1. Command Pattern with Cobra

**Decision**: Use Cobra framework for CLI structure
**Rationale**: 
- Industry standard for Go CLI applications
- Built-in help generation and flag parsing
- Hierarchical command organization
- Consistent user experience

### 2. Dependency Injection Pattern

**Decision**: Inject external tool dependencies rather than direct calls
**Rationale**:
- Improved testability and mocking
- Cleaner separation of concerns
- Easier to swap implementations
- Better error handling and validation

### 3. Interactive Configuration

**Decision**: Provide guided, interactive setup workflows
**Rationale**:
- Reduces cognitive load for new users
- Prevents common configuration errors
- Provides contextual help and validation
- Improves developer experience

### 4. Stateless Design

**Decision**: CLI maintains minimal local state
**Rationale**:
- Simplifies debugging and troubleshooting
- Reduces configuration drift
- Enables consistent behavior across environments
- Easier cluster sharing and collaboration

### 5. Graceful Error Handling

**Decision**: Comprehensive error checking with actionable messages
**Rationale**:
- Improves developer productivity
- Reduces support burden
- Provides clear remediation steps
- Handles common environment issues

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                    # Command implementations
│   ├── cluster/           # Cluster management commands
│   │   ├── create.go      # Cluster creation logic
│   │   ├── list.go        # Cluster listing functionality
│   │   ├── status.go      # Status checking and display
│   │   └── delete.go      # Cluster deletion and cleanup
│   ├── chart/             # Chart and ArgoCD commands
│   │   └── install.go     # Helm chart installation
│   ├── dev/               # Development workflow commands
│   │   ├── scaffold.go    # Skaffold integration
│   │   └── intercept.go   # Telepresence integration
│   ├── bootstrap.go       # Full OpenFrame bootstrap
│   └── root.go            # Root command and global flags
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   │   ├── config.go      # Configuration structures
│   │   └── validation.go  # Input validation logic
│   ├── k3d/               # K3d cluster operations
│   │   ├── client.go      # K3d API wrapper
│   │   └── cluster.go     # Cluster management logic
│   ├── helm/              # Helm chart operations
│   │   ├── client.go      # Helm client wrapper
│   │   └── charts.go      # Chart installation logic
│   ├── ui/                # User interface components
│   │   ├── prompts.go     # Interactive prompts
│   │   ├── tables.go      # Table formatting
│   │   └── progress.go    # Progress indicators
│   └── utils/             # Utility functions
│       ├── system.go      # System detection and validation
│       ├── docker.go      # Docker operations
│       └── network.go     # Network configuration
├── pkg/                   # Public API packages
│   └── version/           # Version information
├── docs/                  # Documentation
│   ├── codewiki/          # Technical documentation
│   └── examples/          # Usage examples
├── scripts/               # Build and deployment scripts
├── .goreleaser.yml        # Release configuration
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── main.go               # Application entry point
└── README.md             # Project overview
```

### Key Directory Purposes

- **`cmd/`**: Contains all CLI command implementations, organized by functional area
- **`internal/`**: Private application code that cannot be imported by other projects
- **`pkg/`**: Public APIs that could be used by other Go projects
- **`docs/`**: All documentation including technical specs and usage examples
- **`scripts/`**: Build automation, testing, and deployment utilities

This architecture provides a clean separation of concerns, making the codebase maintainable and extensible while delivering a consistent user experience across all OpenFrame CLI operations.