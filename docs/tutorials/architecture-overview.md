# Architecture Overview

## High-Level Architecture

The OpenFrame CLI follows a modular, command-based architecture typical of modern CLI applications. If visualized as a diagram, it would show:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Commands  │────│  Business Logic  │────│  External APIs  │
│                 │    │                  │    │                 │
│ • cluster       │    │ • Cluster Mgmt   │    │ • K3d           │
│ • chart         │    │ • Chart Install  │    │ • Kubernetes    │
│ • bootstrap     │    │ • Status Check   │    │ • Helm          │
│ • dev           │    │ • System Config  │    │ • ArgoCD        │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌──────────────────┐
                    │   Utilities &    │
                    │   Shared Libs    │
                    └──────────────────┘
```

## Main Components and Responsibilities

### 1. Command Layer (`cmd/`)
- **Primary Commands**: `cluster`, `chart`, `bootstrap`, `dev`
- **Responsibilities**: 
  - Parse CLI arguments and flags
  - Provide user interface and help text
  - Coordinate calls to business logic
  - Handle command-line validation

### 2. Business Logic Layer (`pkg/` or `internal/`)
- **Cluster Management**: K3d cluster lifecycle operations
- **Chart Installation**: Helm chart deployment and ArgoCD setup
- **Bootstrap Process**: Full OpenFrame platform installation
- **Development Tools**: Skaffold and Telepresence integration
- **System Detection**: Environment and dependency checking

### 3. External Integration Layer
- **K3d API**: Local Kubernetes cluster management
- **Kubernetes API**: Cluster status and resource management
- **Helm**: Chart installation and management
- **ArgoCD**: GitOps deployment management
- **Skaffold**: Development workflow automation
- **Telepresence**: Service traffic interception

### 4. Utilities and Shared Libraries
- **Configuration Management**: CLI settings and defaults
- **Logging**: Structured output and debugging
- **Error Handling**: Consistent error reporting
- **System Utilities**: OS detection, binary checks

## Data Flow Between Components

### 1. Cluster Creation Flow
```
User Input → CLI Parser → Cluster Manager → K3d API → Kubernetes Cluster
                                    ↓
Status Reporter ← System Validator ←┘
```

### 2. Bootstrap Flow
```
User Command → Bootstrap Orchestrator → Chart Installer → Helm
                        ↓                       ↓
              ArgoCD Setup ← Status Monitor ←──┘
```

### 3. Development Workflow
```
Dev Command → Tool Detector → Skaffold/Telepresence → Running Services
                   ↓
         Config Generator → Local Development Environment
```

## Key Design Decisions and Patterns

### 1. **Command Pattern**
- Each major feature is organized as a separate command
- Consistent interface across all operations
- Easy to extend with new functionality

### 2. **Factory Pattern**
- Dynamic creation of cluster configurations
- Flexible chart installation strategies
- Configurable development tool integration

### 3. **Strategy Pattern**
- Multiple deployment modes (OSS, tenant, etc.)
- Different cluster configurations based on environment
- Pluggable development workflows

### 4. **Observer Pattern**
- Real-time status monitoring
- Progress reporting during long-running operations
- Event-driven cluster state changes

### 5. **Configuration-Driven Design**
- YAML-based chart configurations
- Environment-specific settings
- User preference persistence

## Directory Structure

```
openframe-cli/
├── cmd/                          # CLI command definitions
│   ├── root.go                  # Root command and global flags
│   ├── cluster/                 # Cluster management commands
│   │   ├── create.go           # Cluster creation
│   │   ├── list.go             # List clusters
│   │   ├── status.go           # Cluster status
│   │   └── delete.go           # Cluster deletion
│   ├── chart/                   # Chart management commands
│   │   └── install.go          # Chart installation
│   ├── bootstrap.go             # Bootstrap command
│   └── dev/                     # Development commands
│       ├── scaffold.go         # Skaffold integration
│       └── intercept.go        # Telepresence integration
├── pkg/                         # Public packages (if any)
├── internal/                    # Private application logic
│   ├── cluster/                # Cluster management logic
│   │   ├── manager.go          # Main cluster operations
│   │   ├── k3d.go             # K3d-specific implementation
│   │   └── validator.go        # Cluster validation
│   ├── chart/                  # Chart installation logic
│   │   ├── installer.go        # Helm chart installation
│   │   └── argocd.go          # ArgoCD setup
│   ├── bootstrap/              # Bootstrap orchestration
│   │   └── orchestrator.go     # Full platform setup
│   ├── dev/                    # Development tools
│   │   ├── skaffold.go        # Skaffold wrapper
│   │   └── telepresence.go     # Telepresence wrapper
│   ├── config/                 # Configuration management
│   │   ├── loader.go          # Config file handling
│   │   └── defaults.go        # Default settings
│   └── utils/                  # Shared utilities
│       ├── system.go          # System detection
│       ├── logger.go          # Logging utilities
│       └── validation.go      # Input validation
├── configs/                     # Configuration templates
│   ├── charts/                 # Helm chart configurations
│   └── clusters/               # Cluster templates
├── scripts/                     # Build and deployment scripts
├── docs/                        # Documentation
├── go.mod                       # Go module definition
├── go.sum                       # Go module checksums
├── main.go                      # Application entry point
├── Makefile                     # Build automation
└── README.md                    # Project documentation
```

### Key Directory Explanations

- **`cmd/`**: Contains the CLI command structure using Cobra framework
- **`internal/`**: Private application code that implements business logic
- **`pkg/`**: Public packages that could be imported by other projects
- **`configs/`**: Template files and default configurations
- **`scripts/`**: Automation scripts for building, testing, and releasing

This architecture promotes:
- **Separation of Concerns**: Clear boundaries between UI, business logic, and external integrations
- **Testability**: Each layer can be tested independently
- **Extensibility**: New commands and features can be added easily
- **Maintainability**: Related code is grouped together logically