<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-dark-bg.png">
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-light-bg.png">
    <img alt="OpenFrame Logo" src="https://raw.githubusercontent.com/flamingo-stack/openframe-oss-tenant/main/docs/assets/logo-openframe-full-light-bg.png" width="400">
  </picture>
</div>

<p align="center">
  <a href="LICENSE.md"><img alt="License" src="https://img.shields.io/badge/LICENSE-FLAMINGO%20AI%20Unified%20v1.0-%23FFC109?style=for-the-badge&labelColor=white"></a>
</p>

# OpenFrame CLI

A modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI simplifies the creation, management, and development experience for cloud-native applications on Kubernetes, providing everything from cluster lifecycle management to live development workflows.

## Features

- 🎯 **Interactive Cluster Management** - Create and manage K3d clusters with guided wizards and smart configuration detection
- ⚡ **One-Command Bootstrap** - Complete OpenFrame environment setup with a single command
- 📊 **Real-time Monitoring** - Cluster status monitoring and health checks with detailed diagnostics
- 🔧 **GitOps Ready** - Automated ArgoCD installation and app-of-apps deployment patterns
- 🛠 **Developer Tools** - Integrated Telepresence traffic interception and Skaffold live reload workflows
- 📦 **Chart Management** - Helm chart installation with ArgoCD for declarative application management
- 🚀 **CI/CD Integration** - Non-interactive modes perfect for automated testing and deployment pipelines
- 🔄 **Multi-Cluster Support** - Manage multiple environments (dev, staging, test) with easy context switching

## Quick Start

### Prerequisites

Ensure you have these tools installed:
- [Docker](https://docs.docker.com/get-docker/) (20.10+)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) (1.24+)
- [Helm](https://helm.sh/docs/intro/install/) (3.8+)
- [Git](https://git-scm.com/downloads) (2.30+)

### Installation

#### From Release (Recommended)

```bash
# macOS (ARM64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_arm64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_darwin_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/flamingo-stack/openframe-cli/releases/latest/download/openframe-cli_linux_amd64.tar.gz | tar xz
sudo mv openframe /usr/local/bin/

# Windows (AMD64)
# Download from: https://github.com/flamingo-stack/openframe-cli/releases/latest
```

#### From Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
sudo mv openframe /usr/local/bin/
```

### Get Started in 30 Seconds

```bash
# Create a complete OpenFrame environment
openframe bootstrap

# Or with a custom cluster name
openframe bootstrap my-dev-cluster

# Check cluster status
openframe cluster status

# Start developing with live reload
openframe dev scaffold
```

This creates a local Kubernetes cluster, installs ArgoCD, and sets up the complete OpenFrame development environment.

## Core Commands

### Cluster Management

```bash
# Create a new K3d cluster with interactive setup
openframe cluster create

# List all managed clusters
openframe cluster list

# Check cluster health and status
openframe cluster status

# Delete a cluster and cleanup resources
openframe cluster delete my-cluster

# Start/stop cluster management
openframe cluster start my-cluster
openframe cluster cleanup
```

### Bootstrap (Complete Setup)

```bash
# Interactive bootstrap with guided setup
openframe bootstrap

# Quick setup with default configuration
openframe bootstrap --deployment-mode=oss-tenant

# CI/CD friendly non-interactive mode
openframe bootstrap --non-interactive --deployment-mode=oss-tenant
```

### Chart and Application Management

```bash
# Install ArgoCD and Helm charts
openframe chart install

# Bootstrap includes chart installation automatically
openframe bootstrap  # Includes chart setup
```

### Development Workflows

```bash
# Live development with Skaffold
openframe dev scaffold

# Traffic interception for local development
openframe dev intercept my-service

# Access service locally while connected to cluster
```

## Development Workflows

### Local Development with Live Reload

```bash
# Start your development environment
openframe bootstrap my-dev

# Begin live development (auto-rebuilds on code changes)
openframe dev scaffold

# Your code changes are automatically deployed to the cluster
```

### Traffic Interception

```bash
# Intercept traffic from a service in the cluster
openframe dev intercept user-service

# Now requests to user-service route to your local machine
# Develop locally while using real cluster data and dependencies
```

### Multi-Environment Management

```bash
# Create different environments
openframe cluster create frontend-dev
openframe cluster create backend-test
openframe cluster create integration

# Switch between environments
kubectl config use-context k3d-frontend-dev
openframe cluster status
```

## Configuration Options

### Deployment Modes

| Mode | Description | Use Case |
|------|-------------|----------|
| `oss-tenant` | Open source single-tenant setup | Individual development |
| `saas-tenant` | SaaS multi-tenant configuration | Team development |
| `saas-shared` | Shared SaaS environment | Testing and staging |

### Advanced Options

```bash
# Custom cluster with specific node count
openframe cluster create production-test --nodes 5

# Verbose logging for troubleshooting
openframe bootstrap --verbose

# Force operations without confirmation
openframe cluster delete my-cluster --force

# Dry run to preview operations
openframe bootstrap --dry-run
```

## Architecture

OpenFrame CLI follows a modular architecture with clear separation between commands, services, and providers:

- **Command Layer** - CLI interface with Cobra framework
- **Service Layer** - Business logic for cluster, chart, and development operations  
- **Provider Layer** - Integration with K3d, Helm, ArgoCD, and development tools
- **Shared Components** - UI components, error handling, and utilities

For detailed architecture documentation, see [docs/dev/overview.md](docs/dev/overview.md).

## Documentation

- **[Getting Started Guide](docs/tutorials/user/getting-started.md)** - Complete setup walkthrough
- **[Common Use Cases](docs/tutorials/user/common-use-cases.md)** - Practical examples and workflows
- **[Developer Guide](docs/tutorials/dev/getting-started-dev.md)** - Contributing to OpenFrame CLI
- **[Architecture Overview](docs/dev/overview.md)** - Technical architecture and design decisions

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Docker not running | Start Docker Desktop or `sudo systemctl start docker` |
| Port conflicts | Stop services using ports 80, 443, 6443 |
| kubectl context issues | Run `openframe cluster status` to verify active cluster |
| Permission denied | Ensure user is in docker group: `sudo usermod -aG docker $USER` |

### Debug Commands

```bash
# Check cluster health
openframe cluster status

# View detailed operation logs  
openframe bootstrap --verbose

# Clean up stuck resources
openframe cluster cleanup

# Reset cluster completely
openframe cluster delete my-cluster
openframe cluster create my-cluster
```

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:

- Setting up the development environment
- Code style and testing requirements
- Submitting pull requests
- Reporting issues

## License

This project is licensed under the Flamingo AI Unified License v1.0. See the [LICENSE](LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>