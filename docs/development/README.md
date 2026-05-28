# Development Documentation

Welcome to the OpenFrame CLI development documentation. This section provides comprehensive guides for developers working with, contributing to, or extending OpenFrame CLI.

## Overview

OpenFrame CLI is a Go-based command-line tool that orchestrates Kubernetes cluster management with GitOps automation. The development documentation covers everything from setting up your development environment to understanding the architecture and contributing guidelines.

## Quick Navigation

### Setup and Environment
- **[Environment Setup](setup/environment.md)** - IDE, tools, and development environment configuration
- **[Local Development](setup/local-development.md)** - Clone, build, run, and debug locally

### Architecture and Design
- **[Architecture Overview](architecture/README.md)** - High-level system architecture and component relationships

### Security
- **[Security Guidelines](security/README.md)** - Security best practices, authentication patterns, and vulnerability management

### Testing
- **[Testing Overview](testing/README.md)** - Test structure, running tests, and writing new tests

### Contributing
- **[Contributing Guidelines](contributing/guidelines.md)** - Code style, PR process, and review checklist

## Development Workflow

### Typical Development Process

1. **Setup**: Configure your development environment
2. **Build**: Compile and test the CLI locally
3. **Develop**: Make changes following architecture patterns
4. **Test**: Run comprehensive tests before committing
5. **Security**: Follow security best practices
6. **Contribute**: Submit PRs following contribution guidelines

### Development Commands

```bash
# Build the CLI
go build -o openframe ./main.go

# Run tests
go test ./...

# Run with race detection
go test -race ./...

# Build for multiple platforms
make build-all

# Run linting
golangci-lint run
```

## Technology Stack

OpenFrame CLI is built using:

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Core Language** | Go 1.21+ | CLI implementation and business logic |
| **CLI Framework** | Cobra | Command structure and argument parsing |
| **Container Runtime** | Docker | K3d cluster management |
| **Kubernetes** | K3d, kubectl | Local cluster creation and management |
| **GitOps** | ArgoCD, Helm | Application deployment and management |
| **Development Tools** | Telepresence, Skaffold | Traffic interception and live development |

## Architecture Highlights

### Command Structure
```text
openframe
├── bootstrap/          # Complete environment setup
├── cluster/            # Cluster lifecycle management
│   ├── create         # Interactive cluster creation
│   ├── delete         # Safe cluster removal
│   ├── list           # Display managed clusters
│   ├── status         # Cluster health monitoring
│   └── cleanup        # Resource cleanup
├── chart/             # Helm chart and ArgoCD management
│   └── install        # ArgoCD installation
└── dev/               # Development workflow tools
    ├── intercept      # Traffic interception
    └── skaffold       # Live development
```

### Key Design Principles

- **Modularity**: Each command is self-contained with clear responsibilities
- **User Experience**: Interactive wizards and helpful error messages
- **Safety**: Confirmation prompts and safe defaults
- **Extensibility**: Plugin-friendly architecture for custom workflows
- **GitOps Native**: Built-in ArgoCD integration and app-of-apps pattern

## Development Environment

### Recommended Tools

- **IDE**: VS Code with Go extension, GoLand
- **Go Version**: 1.21 or later
- **Docker**: Latest stable version
- **Git**: Latest version with SSH key setup
- **Make**: For build automation

### Environment Variables

```bash
# Go development
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$PATH

# OpenFrame development
export OPENFRAME_DEV_MODE=true
export OPENFRAME_LOG_LEVEL=debug

# Kubernetes development
export KUBECONFIG=$HOME/.kube/config
```

## Contributing Areas

We welcome contributions in these areas:

### Core Features
- New command implementations
- Enhanced interactive wizards
- Improved error handling and messaging
- Performance optimizations

### Platform Support
- Additional deployment modes
- Cloud provider integrations
- Windows compatibility improvements
- macOS Apple Silicon optimizations

### Development Experience
- Enhanced debugging capabilities
- Better logging and observability
- Developer productivity tools
- Documentation improvements

### Testing and Quality
- Unit test coverage improvements
- Integration test scenarios
- Performance benchmarks
- Security testing automation

## Getting Started with Development

1. **Read the setup guides** to configure your environment
2. **Explore the architecture** to understand the system design
3. **Review security guidelines** before making changes
4. **Check the testing approach** for quality standards
5. **Follow contributing guidelines** for smooth collaboration

## Code Quality Standards

- **Test Coverage**: Minimum 80% for new code
- **Linting**: All code must pass `golangci-lint`
- **Documentation**: Public APIs must have Go doc comments
- **Security**: Follow OWASP guidelines and security best practices
- **Performance**: Benchmark critical paths and avoid regressions

## Resources

### External Documentation
- [Go Documentation](https://golang.org/doc/)
- [Cobra CLI Framework](https://cobra.dev/)
- [K3d Documentation](https://k3d.io/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)

### Community
- **OpenMSP Slack**: [Join here](https://join.slack.com/t/openmsp/shared_invite/zt-36bl7mx0h-3~U2nFH6nqHqoTPXMaHEHA)
- **GitHub Discussions**: Use for feature requests and architectural discussions
- **GitHub Issues**: For bug reports and specific improvements

> 💡 **New to the codebase?** Start with the [Environment Setup](setup/environment.md) guide, then explore the [Architecture Overview](architecture/README.md) to understand how everything fits together.