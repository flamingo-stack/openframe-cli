# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a **layered architecture** with clear separation of concerns. A visual diagram would show:

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Interface Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────┐│
│  │   cluster   │  │    chart    │  │     dev     │  │bootstrap ││
│  │  commands   │  │  commands   │  │  commands   │  │ commands ││
│  └─────────────┘  └─────────────┘  └─────────────┘  └──────────┘│
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────┐│
│  │  Cluster    │  │    Helm     │  │  Skaffold   │  │   K3d    ││
│  │  Service    │  │   Service   │  │   Service   │  │ Service  ││
│  └─────────────┘  └─────────────┘  └─────────────┘  └──────────┘│
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                    External Tools Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────┐│
│  │    K3d      │  │    Helm     │  │  Skaffold   │  │Telepresence│
│  │   Binary    │  │   Binary    │  │   Binary    │  │  Binary  ││
│  └─────────────┘  └─────────────┘  └─────────────┘  └──────────┘│
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                         │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │   Docker    │  │ Kubernetes  │  │  File       │              │
│  │   Engine    │  │   Cluster   │  │ System      │              │
│  └─────────────┘  └─────────────┘  └─────────────┘              │
└─────────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer
**Location**: `cmd/` directory
- **Root Command**: Entry point and global configuration
- **Cluster Commands**: K3d cluster lifecycle management
- **Chart Commands**: Helm chart and ArgoCD installation
- **Development Commands**: Skaffold and Telepresence integration
- **Bootstrap Commands**: Full OpenFrame deployment orchestration

### 2. Service Layer
**Location**: `internal/` directory
- **Cluster Service**: Abstracts K3d operations and cluster state management
- **Helm Service**: Manages Helm repositories, charts, and releases
- **Tool Service**: Handles external tool detection, installation, and execution
- **Configuration Service**: Manages CLI settings and cluster configurations

### 3. Utility Layer
**Location**: `pkg/` directory
- **Logging**: Structured logging with different levels
- **System Detection**: OS and architecture detection
- **File Operations**: Template processing and file management
- **Error Handling**: Consistent error reporting and user feedback

## Data Flow Between Components

### 1. Command Execution Flow
```
User Input → CLI Parser → Command Handler → Service Layer → External Tools → Infrastructure
```

### 2. Cluster Creation Flow
```
cluster create → ClusterService → K3d Binary → Docker Engine → Kubernetes Cluster
```

### 3. Bootstrap Flow
```
bootstrap → ChartService → Helm Binary → Kubernetes API → ArgoCD → Application Deployment
```

### 4. Development Workflow Flow
```
dev scaffold → SkaffoldService → Skaffold Binary → Docker Build → K8s Deployment
dev intercept → TelepresenceService → Telepresence Binary → Traffic Routing
```

## Key Design Decisions and Patterns

### 1. **Command Pattern**
- Each CLI command is implemented as a separate struct
- Commands are organized by domain (cluster, chart, dev, bootstrap)
- Enables easy testing and maintenance of individual commands

### 2. **Service Layer Abstraction**
- Business logic separated from CLI interface
- Services can be reused across different commands
- Enables easier testing with dependency injection

### 3. **External Tool Management**
- Tools are detected at runtime rather than bundled
- Automatic installation when tools are missing
- Version compatibility checking

### 4. **Configuration-Driven Design**
- YAML-based configuration for clusters and deployments
- Template-based generation for Kubernetes manifests
- Environment-specific overrides

### 5. **Error Handling Strategy**
- Structured error types with context
- User-friendly error messages
- Automatic cleanup on failures

### 6. **Progressive Enhancement**
- Core functionality works without optional tools
- Advanced features enabled when additional tools are available
- Graceful degradation when tools are unavailable

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                          # CLI command definitions
│   ├── bootstrap.go             # Bootstrap command implementation
│   ├── chart.go                 # Chart management commands
│   ├── cluster.go               # Cluster management commands
│   ├── dev.go                   # Development workflow commands
│   └── root.go                  # Root command and global flags
│
├── internal/                     # Private application code
│   ├── services/                # Business logic services
│   │   ├── cluster.go           # K3d cluster operations
│   │   ├── helm.go              # Helm chart operations
│   │   ├── tools.go             # External tool management
│   │   └── config.go            # Configuration management
│   │
│   ├── models/                  # Data structures
│   │   ├── cluster.go           # Cluster configuration models
│   │   └── config.go            # Application configuration models
│   │
│   └── templates/               # Template files
│       ├── argocd/              # ArgoCD application templates
│       └── k8s/                 # Kubernetes manifest templates
│
├── pkg/                         # Public/reusable packages
│   ├── logger/                  # Logging utilities
│   ├── system/                  # System detection utilities
│   └── utils/                   # General utility functions
│
├── docs/                        # Documentation
│   ├── codewiki/               # Architecture and design docs
│   └── user/                   # User-facing documentation
│
├── scripts/                     # Build and development scripts
│   ├── build.sh               # Build script
│   └── install.sh             # Installation script
│
├── .goreleaser.yml             # Release configuration
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── main.go                     # Application entry point
└── README.md                   # Project documentation
```

### Key Directory Purposes

- **`cmd/`**: Contains all CLI command implementations using the Cobra library
- **`internal/`**: Private application code that cannot be imported by other projects
- **`internal/services/`**: Core business logic separated from CLI concerns
- **`internal/models/`**: Data structures and configuration types
- **`internal/templates/`**: Go templates for generating Kubernetes resources
- **`pkg/`**: Reusable packages that could potentially be imported by other projects
- **`docs/`**: All documentation including architecture decisions and user guides

This structure follows Go best practices and enables:
- Clear separation of concerns
- Easy testing of individual components
- Modular development and maintenance
- Potential code reuse in other projects