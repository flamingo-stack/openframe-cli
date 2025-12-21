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

A modern CLI tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI simplifies the process of creating K3d clusters, deploying OpenFrame applications, and managing development environments with integrated GitOps workflows.

## Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for K3d cluster setup with smart configuration
- ⚡ **One-Command Bootstrap** - Complete environment setup from cluster creation to application deployment
- 📊 **Real-time Monitoring** - Live cluster status monitoring and health checks
- 🔧 **Smart System Detection** - Automatic prerequisite detection and installation guidance
- 🛠 **Developer Tools** - Integrated Telepresence, Skaffold, and live development workflows
- 📦 **GitOps Ready** - ArgoCD integration with app-of-apps pattern for scalable deployments
- 🚀 **Multi-Mode Support** - OSS tenant, SaaS tenant, and SaaS shared deployment modes

## Quick Start

### Installation

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

# From Source
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
```

### Get Started in 30 Seconds

```bash
# Bootstrap a complete OpenFrame environment
openframe bootstrap --deployment-mode=oss-tenant

# Or step by step:
# 1. Create a cluster
openframe cluster create

# 2. Install OpenFrame
openframe chart install

# 3. Check status
openframe cluster status
```

### Common Commands

```bash
# Cluster Management
openframe cluster list                    # List all clusters
openframe cluster status                  # Show cluster details
openframe cluster delete                  # Clean up cluster

# Development Workflows
openframe dev intercept api              # Intercept service traffic
openframe dev skaffold                   # Live development mode

# Get help
openframe --help
```

## Core Commands

### Cluster Management
- `cluster create` - Interactive cluster creation with guided setup
- `cluster list` - View all managed clusters with status
- `cluster status` - Detailed cluster health and configuration
- `cluster delete` - Clean cluster removal with Docker cleanup
- `cluster start/stop` - Control cluster lifecycle

### Application Deployment
- `bootstrap` - Complete end-to-end environment setup
- `chart install` - Deploy ArgoCD and OpenFrame applications
- GitOps-driven application management with ArgoCD

### Development Tools
- `dev intercept` - Service traffic interception with Telepresence
- `dev scaffold` - Live development workflows with Skaffold
- Hot-reload development environment support

## Prerequisites

The CLI automatically detects and guides installation of required tools:
- Docker Desktop or Docker Engine
- K3d (Kubernetes in Docker)
- kubectl (Kubernetes CLI)
- Helm (Package manager)

Optional for development:
- Telepresence (Service intercepts)
- Skaffold (Live development)

## Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:
- [Getting Started Guide](./docs/getting-started/introduction.md)
- [Architecture Overview](./docs/development/architecture/overview.md)
- [Development Setup](./docs/development/setup/environment.md)
- [Contributing Guidelines](./docs/development/contributing/guidelines.md)

## Architecture

OpenFrame CLI is built with Go and Cobra, providing:
- **Interactive UI** - Rich terminal interfaces with progress indicators
- **Modular Design** - Separate packages for cluster, chart, and dev operations
- **Error Handling** - User-friendly error messages and troubleshooting
- **Extensible** - Plugin-ready architecture for custom workflows

## Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details on:
- Setting up the development environment
- Running tests and quality checks
- Submitting pull requests
- Code standards and review process

## License

This project is licensed under the Flamingo AI Unified License v1.0 - see the [LICENSE.md](LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>