# Architecture Overview

## High-Level Architecture

The OpenFrame CLI follows a modular, layered architecture designed for extensibility and maintainability. The system can be visualized as follows:

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Interface Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Cobra     │  │  Interactive │  │    Formatting &     │ │
│  │  Commands   │  │   Prompts    │  │      Output         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  Business Logic Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Cluster   │  │    Chart    │  │    Development      │ │
│  │ Management  │  │ Management  │  │      Tools          │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │     K3d     │  │    Helm     │  │   Kubernetes API    │ │
│  │    Client   │  │   Client    │  │      Client         │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                  External Systems                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │   Docker    │  │ Kubernetes  │  │    Development      │ │
│  │   Engine    │  │  Clusters   │  │      Tools          │ │
│  │             │  │   (K3d)     │  │ (Skaffold/Telepres) │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer

**Commands (`cmd/`)**
- **Responsibility**: Define command structure, parse arguments, and orchestrate business logic
- **Key Components**:
  - `root.go`: Main CLI entry point and global configuration
  - `cluster.go`: Cluster lifecycle management commands
  - `chart.go`: Chart installation and management
  - `bootstrap.go`: Full OpenFrame deployment orchestration
  - `dev.go`: Development workflow commands

**Interactive UI (`internal/ui/`)**
- **Responsibility**: Provide user-friendly prompts and formatted output
- **Key Components**:
  - Interactive prompts for cluster creation
  - Status displays and progress indicators
  - Error formatting and help text

### 2. Business Logic Layer

**Cluster Management (`internal/cluster/`)**
- **Responsibility**: K3d cluster lifecycle operations
- **Operations**: Create, start, stop, delete, list, and monitor clusters
- **Features**: Automatic port management, registry configuration, resource validation

**Chart Management (`internal/chart/`)**
- **Responsibility**: Helm chart installation and ArgoCD setup
- **Operations**: Chart templating, dependency resolution, ArgoCD application management

**Development Tools (`internal/dev/`)**
- **Responsibility**: Developer workflow automation
- **Operations**: Skaffold integration, Telepresence traffic interception

### 3. Infrastructure Layer

**Kubernetes Clients**
- **Responsibility**: Interface with Kubernetes APIs
- **Components**: Native Kubernetes client, custom resource handling

**External Tool Integration**
- **Responsibility**: Manage external tool dependencies
- **Tools**: K3d, Helm, Docker, Skaffold, Telepresence

## Data Flow Between Components

### 1. Command Execution Flow

```
User Input → Cobra Command → Business Logic → Infrastructure Clients → External Systems
     ↓           ↓              ↓                    ↓                      ↓
CLI Args → Validation → Processing → API Calls → System Changes
     ↑           ↑              ↑                    ↑                      ↑
User Output ← Formatting ← Results ← Responses ← Status Updates
```

### 2. Cluster Creation Flow

1. **User Input**: Interactive prompts collect cluster configuration
2. **Validation**: System checks prerequisites and validates input
3. **K3d Creation**: Business logic calls K3d client to create cluster
4. **Post-Setup**: Configure registries, networking, and validation
5. **Status Report**: Display cluster information and next steps

### 3. Bootstrap Flow

1. **Cluster Detection**: Verify target cluster accessibility
2. **Chart Installation**: Deploy base charts (ingress, cert-manager, etc.)
3. **ArgoCD Setup**: Install and configure ArgoCD for GitOps
4. **Application Deployment**: Deploy OpenFrame applications via ArgoCD
5. **Validation**: Health checks and status verification

## Key Design Decisions and Patterns

### 1. **Command Pattern with Cobra**
- **Decision**: Use Cobra framework for CLI structure
- **Rationale**: Industry standard, excellent UX, automatic help generation
- **Implementation**: Hierarchical command structure with shared flags and configuration

### 2. **Dependency Injection**
- **Decision**: Inject clients and configuration into business logic
- **Rationale**: Improves testability and modularity
- **Implementation**: Factory functions and interface-based design

### 3. **Graceful Error Handling**
- **Decision**: Structured error handling with user-friendly messages
- **Rationale**: Better developer experience, easier debugging
- **Implementation**: Custom error types with context and suggestions

### 4. **Interactive User Experience**
- **Decision**: Provide interactive prompts with sensible defaults
- **Rationale**: Reduces learning curve, prevents configuration errors
- **Implementation**: Conditional prompts based on available options

### 5. **External Tool Abstraction**
- **Decision**: Abstract external tool interactions behind interfaces
- **Rationale**: Easier testing, potential for alternative implementations
- **Implementation**: Client interfaces with concrete implementations

## Directory Structure

```
openframe-cli/
├── cmd/                          # CLI commands and entry points
│   ├── root.go                   # Main CLI configuration and global flags
│   ├── cluster.go                # Cluster management commands
│   ├── chart.go                  # Chart installation commands
│   ├── bootstrap.go              # Bootstrap orchestration
│   └── dev.go                    # Development workflow commands
│
├── internal/                     # Private application code
│   ├── cluster/                  # Cluster management logic
│   │   ├── manager.go            # Main cluster operations
│   │   ├── k3d.go               # K3d client wrapper
│   │   └── validation.go        # Cluster validation
│   │
│   ├── chart/                    # Chart management logic
│   │   ├── installer.go         # Helm chart installation
│   │   └── argocd.go           # ArgoCD management
│   │
│   ├── dev/                     # Development tools integration
│   │   ├── skaffold.go          # Skaffold integration
│   │   └── telepresence.go      # Telepresence management
│   │
│   ├── ui/                      # User interface components
│   │   ├── prompts.go           # Interactive prompts
│   │   └── output.go            # Formatted output
│   │
│   └── config/                  # Configuration management
│       ├── config.go            # Application configuration
│       └── defaults.go          # Default values
│
├── pkg/                         # Public API packages
│   └── types/                   # Shared types and interfaces
│
├── docs/                        # Documentation
│   └── codewiki/               # Technical documentation
│
├── scripts/                     # Build and development scripts
├── .github/                     # GitHub workflows and templates
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── main.go                      # Application entry point
└── README.md                    # Project overview
```

### Directory Guidelines

- **`cmd/`**: Contains only CLI command definitions and basic orchestration
- **`internal/`**: All private business logic, not importable by external packages
- **`pkg/`**: Public APIs that could be imported by other projects
- **`docs/`**: All documentation, including this architecture overview
- **`scripts/`**: Automation scripts for building, testing, and releasing

### Package Organization Principles

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Dependency Direction**: Dependencies flow from CLI → Business Logic → Infrastructure
3. **Interface Boundaries**: Clear interfaces between layers to enable testing and modularity
4. **Encapsulation**: Internal packages hide implementation details from external consumers

This architecture supports the OpenFrame CLI's goals of providing a developer-friendly tool for managing Kubernetes clusters while maintaining clean, testable, and extensible code.