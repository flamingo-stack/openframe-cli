# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture pattern with clear separation of concerns:

```
┌─────────────────────────────────────────────┐
│                 CLI Layer                   │
│  ┌─────────────┐ ┌─────────────┐ ┌────────┐ │
│  │   Cobra     │ │   Viper     │ │  UI/UX │ │
│  │ Commands    │ │ Config Mgmt │ │ Output │ │
│  └─────────────┘ └─────────────┘ └────────┘ │
└─────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────┐
│              Service Layer                  │
│  ┌──────────────┐ ┌──────────────┐ ┌──────┐ │
│  │   Cluster    │ │    Chart     │ │ Dev  │ │
│  │  Management  │ │  Management  │ │Tools │ │
│  └──────────────┘ └──────────────┘ └──────┘ │
└─────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────┐
│            Infrastructure Layer             │
│  ┌─────────┐ ┌─────────┐ ┌─────────────────┐ │
│  │   K3d   │ │ Kubectl │ │      Helm       │ │
│  │ Client  │ │ Client  │ │     Client      │ │
│  └─────────┘ └─────────┘ └─────────────────┘ │
└─────────────────────────────────────────────┘
                       │
┌─────────────────────────────────────────────┐
│           External Systems                  │
│  ┌─────────┐ ┌─────────┐ ┌─────────────────┐ │
│  │  K3d    │ │   K8s   │ │    ArgoCD       │ │
│  │Clusters │ │ Cluster │ │   Skaffold      │ │
│  └─────────┘ └─────────┘ └─────────────────┘ │
└─────────────────────────────────────────────┘
```

## Main Components and Their Responsibilities

### CLI Layer
- **Cobra Commands**: Command-line interface definition and argument parsing
- **Viper Config Management**: Configuration file handling and environment variable management
- **UI/UX Output**: User interaction, progress indicators, and formatted output

### Service Layer
- **Cluster Management**: K3d cluster lifecycle operations (create, delete, start, stop, status)
- **Chart Management**: Helm chart installation and ArgoCD bootstrapping
- **Development Tools**: Integration with Skaffold for development workflows and Telepresence for traffic interception

### Infrastructure Layer
- **K3d Client**: Direct interface to K3d for lightweight Kubernetes clusters
- **Kubectl Client**: Kubernetes API interactions for cluster operations
- **Helm Client**: Chart management and deployment operations

### External Systems
- **K3d Clusters**: Local development Kubernetes clusters
- **Kubernetes Cluster**: Target deployment environment
- **ArgoCD/Skaffold**: GitOps and development workflow tools

## Data Flow Between Components

### Cluster Creation Flow
1. **User Input** → CLI Layer parses command and options
2. **CLI Layer** → Cluster Management service validates configuration
3. **Cluster Management** → K3d Client creates cluster with specified parameters
4. **K3d Client** → External K3d daemon provisions cluster resources
5. **Status Updates** → UI/UX layer provides real-time feedback to user

### Chart Installation Flow
1. **User Command** → CLI Layer processes bootstrap/install request
2. **CLI Layer** → Chart Management service orchestrates installation
3. **Chart Management** → Kubectl Client verifies cluster connectivity
4. **Chart Management** → Helm Client installs required charts
5. **Chart Management** → ArgoCD setup and configuration
6. **Progress Updates** → UI/UX layer shows installation progress

### Development Workflow
1. **Dev Command** → CLI Layer routes to appropriate development tool
2. **Development Tools** → Validates cluster and service configuration
3. **Tool Integration** → Spawns Skaffold/Telepresence processes
4. **Live Updates** → Real-time feedback and log streaming

## Key Design Decisions and Patterns

### 1. **Command Pattern with Cobra**
- **Decision**: Use Cobra framework for CLI structure
- **Rationale**: Provides consistent command hierarchy, help generation, and argument validation
- **Implementation**: Each command is a separate module with clear responsibilities

### 2. **Service Layer Abstraction**
- **Decision**: Abstract external tool interactions into service interfaces
- **Rationale**: Enables testing, mocking, and easier maintenance
- **Implementation**: Services encapsulate business logic separate from CLI concerns

### 3. **Progressive Enhancement**
- **Decision**: Graceful degradation when optional tools are unavailable
- **Rationale**: Better user experience across different development environments
- **Implementation**: Tool availability detection and alternative workflows

### 4. **Configuration Management**
- **Decision**: Hierarchical configuration (CLI args → env vars → config files → defaults)
- **Rationale**: Flexibility for different deployment scenarios and user preferences
- **Implementation**: Viper handles the configuration precedence automatically

### 5. **Error Handling Strategy**
- **Decision**: Structured error handling with user-friendly messages
- **Rationale**: Clear feedback for troubleshooting and improved developer experience
- **Implementation**: Custom error types with context and suggested actions

## Directory/Folder Structure

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and global flags
│   ├── cluster/           # Cluster management commands
│   │   ├── create.go      # Cluster creation
│   │   ├── delete.go      # Cluster deletion
│   │   ├── list.go        # List clusters
│   │   └── status.go      # Cluster status
│   ├── chart/             # Chart management commands
│   │   └── install.go     # Helm chart installation
│   ├── dev/               # Development workflow commands
│   │   ├── scaffold.go    # Skaffold integration
│   │   └── intercept.go   # Telepresence integration
│   └── bootstrap.go       # OpenFrame bootstrapping
├── pkg/                   # Core business logic packages
│   ├── cluster/           # Cluster management services
│   │   ├── manager.go     # Main cluster operations
│   │   ├── k3d.go         # K3d-specific implementation
│   │   └── types.go       # Cluster data structures
│   ├── chart/             # Chart management services
│   │   ├── helm.go        # Helm client wrapper
│   │   ├── argocd.go      # ArgoCD installation logic
│   │   └── installer.go   # Chart installation orchestration
│   ├── config/            # Configuration management
│   │   ├── config.go      # Configuration structure and loading
│   │   └── defaults.go    # Default values and validation
│   ├── utils/             # Shared utilities
│   │   ├── output.go      # Output formatting and UI helpers
│   │   ├── spinner.go     # Progress indicators
│   │   └── validation.go  # Input validation functions
│   └── version/           # Version information
│       └── version.go     # Build-time version injection
├── internal/              # Private application code
│   ├── clients/           # External tool client wrappers
│   │   ├── k3d.go         # K3d client interface
│   │   ├── kubectl.go     # Kubectl operations
│   │   └── helm.go        # Helm client operations
│   └── templates/         # Configuration templates
│       ├── cluster.yaml   # Default cluster configuration
│       └── charts.yaml    # Chart installation manifests
├── docs/                  # Documentation
│   └── codewiki/          # Technical documentation
│       └── overview.md    # Project overview
├── scripts/               # Build and development scripts
│   ├── build.sh           # Build automation
│   └── release.sh         # Release preparation
├── .goreleaser.yaml       # Release configuration
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
├── main.go                # Application entry point
└── README.md              # Project documentation
```

### Directory Responsibilities

- **`cmd/`**: Command-line interface definition using Cobra pattern. Each subdirectory represents a command group.
- **`pkg/`**: Public packages that could be imported by other projects. Contains core business logic.
- **`internal/`**: Private application code not intended for external use. Contains implementation details.
- **`docs/`**: Project documentation including architectural decisions and usage guides.
- **`scripts/`**: Automation scripts for building, testing, and releasing the application.

This structure follows Go best practices for CLI applications and provides clear separation between public interfaces, private implementation, and command definitions.