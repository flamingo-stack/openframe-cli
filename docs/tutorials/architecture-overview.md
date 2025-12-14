# Architecture Overview

## High-Level Architecture Diagram Description

The OpenFrame CLI follows a layered architecture pattern that would be visualized as:

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Interface Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Cluster   │  │    Chart    │  │     Dev     │        │
│  │  Commands   │  │  Commands   │  │  Commands   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────┬───────────────────┬───────────────────┘
                      │                   │
┌─────────────────────▼───────────────────▼───────────────────┐
│                  Command Handlers Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Cluster   │  │    Chart    │  │  Bootstrap  │        │
│  │   Service   │  │   Service   │  │   Service   │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────┬───────────────────┬───────────────────┘
                      │                   │
┌─────────────────────▼───────────────────▼───────────────────┐
│                  Integration Layer                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │     K3d     │  │    Helm     │  │   ArgoCD    │        │
│  │   Client    │  │   Client    │  │   Client    │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────┬───────────────────┬───────────────────┘
                      │                   │
┌─────────────────────▼───────────────────▼───────────────────┐
│                External Dependencies                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Kubernetes │  │   Docker    │  │  Skaffold   │        │
│  │   Cluster   │  │   Engine    │  │/Telepresence│        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## Main Components and Responsibilities

### 1. CLI Interface Layer
**Location**: Root command files and command definitions

- **Root Command**: Entry point and global configuration
- **Cluster Commands**: User interface for cluster operations
- **Chart Commands**: User interface for Helm chart management  
- **Dev Commands**: Developer workflow tools interface
- **Bootstrap Command**: Guided setup and installation

**Responsibilities**:
- Parse command-line arguments and flags
- Provide interactive prompts and wizards
- Format and display output to users
- Handle user input validation

### 2. Command Handlers Layer
**Location**: Service packages and business logic

- **Cluster Service**: K3d cluster lifecycle management
- **Chart Service**: Helm chart installation and ArgoCD setup
- **Bootstrap Service**: Orchestrated installation workflow
- **System Detection**: Environment and dependency validation

**Responsibilities**:
- Implement core business logic
- Orchestrate multiple external tool interactions
- Handle error scenarios and recovery
- Maintain state and configuration

### 3. Integration Layer
**Location**: Client wrappers and external tool interfaces

- **K3d Client**: Kubernetes cluster management via K3d
- **Helm Client**: Chart repository and installation management
- **Kubectl Client**: Direct Kubernetes API operations
- **Docker Client**: Container runtime interactions

**Responsibilities**:
- Abstract external tool APIs
- Handle tool-specific configuration and authentication
- Provide consistent error handling across tools
- Manage tool dependencies and version compatibility

### 4. External Dependencies
- **Kubernetes Clusters**: Target deployment environment
- **Docker Engine**: Container runtime for K3d
- **Skaffold**: Development workflow automation
- **Telepresence**: Service traffic interception
- **ArgoCD**: GitOps continuous deployment

## Data Flow Between Components

### 1. Cluster Creation Flow
```
User Input → CLI Parser → Cluster Service → K3d Client → Docker Engine → Kubernetes Cluster
     ↓           ↓            ↓              ↓            ↓              ↓
Status Updates ←─────────────────────────────────────────────────────────────┘
```

### 2. Bootstrap Installation Flow
```
User Input → Bootstrap Service → Chart Service → Helm Client → Kubernetes API
     ↓            ↓                ↓              ↓             ↓
     └─────────→ System Detection → Cluster Service → Status Validation
                     ↓                ↓               ↓
                 ArgoCD Setup ←──────┘               ↓
                     ↓                                ↓
                 Configuration Management ←──────────┘
```

### 3. Development Workflow Flow
```
Dev Command → Service Detection → Tool Selection → External Tool Execution
     ↓             ↓                  ↓                ↓
Real-time Feedback ←─────────────────────────────────┘
```

## Key Design Decisions and Patterns

### 1. **Command Pattern**
- Each command is a separate struct implementing a common interface
- Enables easy addition of new commands without modifying existing code
- Clear separation of concerns between parsing and execution

### 2. **Service Layer Pattern**
- Business logic isolated in service packages
- Commands delegate to services for actual work
- Services are testable independently of CLI interface

### 3. **External Tool Abstraction**
- Wrapper clients for each external dependency
- Consistent error handling and response formatting
- Easy to mock for testing

### 4. **Interactive CLI Design**
- Guided wizards for complex operations
- Smart defaults with override capability
- Progressive disclosure of advanced options

### 5. **Configuration Management**
- File-based configuration with environment variable overrides
- Cluster-specific configuration storage
- Sensible defaults for development workflows

### 6. **Error Handling Strategy**
- Structured error types with context
- User-friendly error messages with actionable suggestions
- Graceful degradation when optional tools are missing

## Directory/Folder Structure Explanation

```
openframe-cli/
├── cmd/                    # CLI command definitions
│   ├── root.go            # Root command and global flags
│   ├── cluster/           # Cluster management commands
│   ├── chart/             # Chart installation commands
│   ├── dev/               # Development workflow commands
│   └── bootstrap/         # Bootstrap installation command
│
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── services/         # Business logic services
│   │   ├── cluster/      # Cluster management service
│   │   ├── chart/        # Chart installation service
│   │   └── system/       # System detection service
│   ├── clients/          # External tool clients
│   │   ├── k3d/          # K3d cluster client
│   │   ├── helm/         # Helm chart client
│   │   └── kubectl/      # Kubernetes API client
│   ├── types/            # Shared data structures
│   └── utils/            # Utility functions
│
├── pkg/                  # Public API packages (if any)
├── docs/                 # Documentation
│   └── codewiki/         # Architecture documentation
├── scripts/              # Build and deployment scripts
├── .github/              # GitHub Actions workflows
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── main.go               # Application entry point
└── README.md             # Project documentation
```

### Key Directories Explained

- **`cmd/`**: Contains all CLI command definitions using the Cobra framework. Each subdirectory represents a command group.

- **`internal/`**: Houses all private application code that shouldn't be imported by other projects.
  - **`services/`**: Core business logic separated from CLI concerns
  - **`clients/`**: Thin wrappers around external tools and APIs
  - **`config/`**: Configuration file handling and environment detection

- **`pkg/`**: Reserved for any public APIs (currently unused but follows Go conventions)

- **`docs/codewiki/`**: Architecture and design documentation for developers

This structure follows Go project conventions and promotes:
- Clear separation of concerns
- Easy testing and mocking
- Maintainable command structure
- Reusable business logic components