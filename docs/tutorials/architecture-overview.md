# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture pattern that would be represented in a diagram as follows:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface Layer                      │
├─────────────────────────────────────────────────────────────┤
│  Command Layer (cobra commands: cluster, chart, dev, etc.)  │
├─────────────────────────────────────────────────────────────┤
│              Service/Business Logic Layer                   │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │   Cluster    │ │    Chart     │ │    Development       │ │
│  │   Service    │ │   Service    │ │     Service          │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                Integration/Client Layer                     │
│  ┌──────────────┐ ┌──────────────┐ ┌──────────────────────┐ │
│  │     K3d      │ │     Helm     │ │   Skaffold/Telepres  │ │
│  │    Client    │ │    Client    │ │        Client        │ │
│  └──────────────┘ └──────────────┘ └──────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│              Infrastructure Layer                           │
│    Docker ←→ Kubernetes ←→ ArgoCD ←→ Local System          │
└─────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer
- **Entry Point**: Main CLI application with global flags and configuration
- **Responsibility**: Command parsing, global middleware, error handling, and user interaction

### 2. Command Layer
- **Cluster Commands**: `cluster create`, `cluster list`, `cluster status`, `cluster delete`, etc.
- **Chart Commands**: `chart install` for Helm charts and ArgoCD management
- **Development Commands**: `dev scaffold`, `dev intercept` for development workflows
- **Bootstrap Commands**: `bootstrap` for full OpenFrame installation
- **Responsibility**: Command-specific logic, input validation, and orchestrating service calls

### 3. Service/Business Logic Layer

#### Cluster Service
- Manages K3d cluster lifecycle (create, start, stop, delete)
- Handles cluster configuration and validation
- Provides cluster status monitoring and health checks

#### Chart Service
- Manages Helm chart installations
- Handles ArgoCD setup and configuration
- Manages chart dependencies and upgrades

#### Development Service
- Integrates with Skaffold for development workflows
- Manages Telepresence for service interception
- Handles development environment setup

### 4. Integration/Client Layer

#### K3d Client
- Abstracts K3d API interactions
- Handles cluster configuration generation
- Manages Docker registry integration

#### Helm Client
- Manages Helm chart operations
- Handles repository management
- Provides chart templating and installation

#### Development Tool Clients
- **Skaffold Client**: Manages build and deploy pipelines
- **Telepresence Client**: Handles traffic interception setup

### 5. Infrastructure Layer
- **Docker**: Container runtime for K3d clusters
- **Kubernetes**: Orchestration platform
- **ArgoCD**: GitOps deployment management
- **Local System**: File system, network, and OS integration

## Data Flow Between Components

### Cluster Creation Flow
```
User Input → CLI Parser → Cluster Command → Cluster Service → K3d Client → Docker/Kubernetes
```

### Chart Installation Flow
```
User Input → CLI Parser → Chart Command → Chart Service → Helm Client → Kubernetes → ArgoCD
```

### Development Workflow Flow
```
User Input → CLI Parser → Dev Command → Development Service → Skaffold/Telepresence Client → Kubernetes
```

### Status Monitoring Flow
```
CLI Command → Cluster Service → K3d Client → Kubernetes API → Status Response → CLI Output
```

## Key Design Decisions and Patterns

### 1. Command Pattern
- Uses Cobra library for command structure
- Each command encapsulates specific functionality
- Supports nested subcommands for logical grouping

### 2. Service Layer Pattern
- Business logic separated from CLI commands
- Services are reusable across different commands
- Clear separation of concerns

### 3. Client Abstraction Pattern
- External tool interactions abstracted behind client interfaces
- Enables easier testing and mocking
- Provides consistent error handling

### 4. Configuration-Driven Approach
- Uses structured configuration for cluster and chart management
- Supports multiple deployment modes (OSS tenant, etc.)
- Environment-specific configuration overrides

### 5. Interactive UX Pattern
- Guided wizards for complex operations
- Real-time status updates and progress indicators
- Clear error messages with actionable suggestions

## Directory/Folder Structure Explanation

Based on the project context and typical Go CLI patterns, the expected structure would be:

```
openframe-cli/
├── cmd/                          # Command definitions (Cobra commands)
│   ├── root.go                   # Root command and global flags
│   ├── cluster/                  # Cluster management commands
│   │   ├── create.go            # cluster create command
│   │   ├── list.go              # cluster list command
│   │   ├── status.go            # cluster status command
│   │   └── delete.go            # cluster delete command
│   ├── chart/                    # Chart management commands
│   │   └── install.go           # chart install command
│   ├── dev/                      # Development commands
│   │   ├── scaffold.go          # dev scaffold command
│   │   └── intercept.go         # dev intercept command
│   └── bootstrap.go              # Bootstrap command
├── internal/                     # Private application code
│   ├── services/                 # Business logic layer
│   │   ├── cluster.go           # Cluster service implementation
│   │   ├── chart.go             # Chart service implementation
│   │   └── dev.go               # Development service implementation
│   ├── clients/                  # External tool clients
│   │   ├── k3d.go               # K3d client wrapper
│   │   ├── helm.go              # Helm client wrapper
│   │   ├── skaffold.go          # Skaffold client wrapper
│   │   └── telepresence.go      # Telepresence client wrapper
│   ├── config/                   # Configuration management
│   │   ├── config.go            # Configuration structures
│   │   └── defaults.go          # Default configurations
│   └── utils/                    # Utility functions
│       ├── validation.go        # Input validation
│       ├── output.go            # Output formatting
│       └── system.go            # System detection utilities
├── docs/                         # Documentation
│   └── codewiki/                # Code documentation
├── pkg/                          # Public library code (if any)
├── scripts/                      # Build and deployment scripts
├── main.go                       # Application entry point
├── go.mod                        # Go module definition
├── go.sum                        # Go module checksums
└── README.md                     # Project documentation
```

### Key Directory Purposes

- **`cmd/`**: Contains all CLI command definitions using the Cobra pattern
- **`internal/`**: Private application code that cannot be imported by other projects
- **`internal/services/`**: Core business logic implementation
- **`internal/clients/`**: Abstraction layer for external tool integration
- **`internal/config/`**: Configuration management and validation
- **`internal/utils/`**: Shared utility functions
- **`docs/`**: Project documentation and guides
- **`pkg/`**: Public library code (if the CLI exposes reusable packages)

This architecture supports maintainability, testability, and extensibility while following Go best practices for CLI applications.