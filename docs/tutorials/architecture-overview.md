# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture that would be visualized as follows:

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Interface Layer                      │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   cluster   │ │    chart    │ │     dev     │ │    misc     ││
│  │  commands   │ │  commands   │ │  commands   │ │  commands   ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                      Service Layer                              │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │   Cluster   │ │    Chart    │ │  Bootstrap  │ │     Dev     ││
│  │   Service   │ │   Service   │ │   Service   │ │   Service   ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │     K3d     │ │    Helm     │ │   ArgoCD    │ │ Telepresence││
│  │   Client    │ │   Client    │ │   Client    │ │   Client    ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────────┐
│                      External Systems                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐│
│  │ Kubernetes  │ │    Docker   │ │    Local    │ │   Remote    ││
│  │   Cluster   │ │   Runtime   │ │  File Sys   │ │   Repos     ││
│  └─────────────┘ └─────────────┘ └─────────────┘ └─────────────┘│
└─────────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer
- **Command Handlers**: Process user input and route commands to appropriate services
- **Interactive Wizards**: Provide guided workflows for complex operations
- **Output Formatting**: Present results in user-friendly formats with colors and progress indicators

### 2. Service Layer
- **Cluster Service**: Manages K3d cluster lifecycle (create, start, stop, delete)
- **Chart Service**: Handles Helm chart installations and ArgoCD configurations
- **Bootstrap Service**: Orchestrates complete OpenFrame installation workflows
- **Dev Service**: Manages development tools like Skaffold and Telepresence

### 3. Infrastructure Layer
- **K3d Client**: Interfaces with K3d for local Kubernetes cluster management
- **Helm Client**: Manages Helm chart installations and releases
- **ArgoCD Client**: Configures and manages ArgoCD applications
- **Telepresence Client**: Handles service interception for development

### 4. External Systems
- **Kubernetes Cluster**: The target cluster where applications are deployed
- **Docker Runtime**: Container runtime for cluster nodes
- **Local File System**: Configuration files, charts, and temporary data
- **Remote Repositories**: Git repositories and chart registries

## Data Flow Between Components

### Cluster Creation Flow
```
User Input → CLI Commands → Cluster Service → K3d Client → Docker → Kubernetes Cluster
```

### Chart Installation Flow
```
User Input → CLI Commands → Chart Service → Helm Client → Kubernetes API → Cluster Resources
```

### Bootstrap Flow
```
User Input → CLI Commands → Bootstrap Service → 
  ├── Cluster Service (create cluster)
  ├── Chart Service (install base charts)
  └── ArgoCD Service (configure GitOps)
```

### Development Workflow
```
User Input → CLI Commands → Dev Service → 
  ├── Skaffold (build/deploy cycle)
  └── Telepresence (traffic interception)
```

## Key Design Decisions and Patterns

### 1. **Command Pattern**
- Each CLI command is implemented as a separate handler
- Commands are organized into logical groups (cluster, chart, dev)
- Enables easy extension and testing of individual commands

### 2. **Service Layer Abstraction**
- Business logic is separated from CLI presentation
- Services can be reused across different command implementations
- Facilitates unit testing and mocking of external dependencies

### 3. **Dependency Injection**
- External tool clients are injected into services
- Enables easy mocking for testing
- Allows for different implementations (e.g., different Kubernetes distributions)

### 4. **Configuration-Driven Architecture**
- Chart installations and cluster configurations are data-driven
- YAML/JSON configuration files define deployment parameters
- Reduces hardcoded values and improves flexibility

### 5. **Progressive Enhancement**
- Core functionality works with minimal dependencies
- Advanced features are enabled when additional tools are available
- Graceful degradation when optional tools are missing

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   ├── dev/               # Development workflow commands
│   └── root.go            # Root command and global flags
├── internal/              # Private application code
│   ├── services/          # Business logic services
│   │   ├── cluster/       # Cluster management service
│   │   ├── chart/         # Chart installation service
│   │   ├── bootstrap/     # Bootstrap orchestration
│   │   └── dev/           # Development tools service
│   ├── clients/           # External tool clients
│   │   ├── k3d/          # K3d cluster client
│   │   ├── helm/         # Helm chart client
│   │   ├── argocd/       # ArgoCD management client
│   │   └── telepresence/ # Telepresence client
│   ├── config/           # Configuration management
│   ├── utils/            # Shared utilities
│   └── types/            # Shared type definitions
├── pkg/                  # Public API (if any)
├── configs/              # Default configuration files
│   ├── charts/           # Chart value files
│   └── cluster/          # Cluster configuration templates
├── scripts/              # Build and deployment scripts
├── docs/                 # Documentation
└── main.go              # Application entry point
```

### Key Directory Explanations

- **`cmd/`**: Contains all CLI command definitions using Cobra framework
- **`internal/services/`**: Core business logic, isolated from CLI concerns
- **`internal/clients/`**: Wrapper clients for external tools (K3d, Helm, etc.)
- **`internal/config/`**: Configuration loading and validation logic
- **`configs/`**: Default configuration files and templates
- **`scripts/`**: Automation scripts for building, testing, and releasing

This architecture promotes:
- **Separation of Concerns**: Clear boundaries between CLI, business logic, and external integrations
- **Testability**: Services and clients can be easily mocked and tested
- **Extensibility**: New commands and services can be added with minimal impact
- **Maintainability**: Well-organized code structure with clear responsibilities