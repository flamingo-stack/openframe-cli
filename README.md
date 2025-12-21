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

A modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI provides an interactive interface for creating K3d clusters, installing ArgoCD and OpenFrame charts, and managing development tools like Telepresence and Skaffold.

## Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for setting up local K3d clusters
- ⚡ **K3d Cluster Management** - Complete lifecycle management for development clusters
- 📊 **Real-time Monitoring** - Cluster status monitoring and health checks
- 🔧 **Smart Configuration** - Automatic system detection and tool validation
- 🛠 **Developer Tools Integration** - Built-in support for Skaffold and Telepresence
- 📦 **Chart Management** - Seamless Helm chart and ArgoCD installation
- 🚀 **One-Command Bootstrap** - End-to-end OpenFrame setup with single command
- 💻 **Cross-Platform Support** - Works on macOS, Linux, and Windows

## Quick Start

### Installation

Download the latest release for your platform:

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
```

### Basic Usage

```bash
# Complete OpenFrame setup (recommended for new users)
openframe bootstrap --deployment-mode=oss-tenant

# Or step by step:
# 1. Create a cluster
openframe cluster create my-dev-cluster

# 2. Check cluster status
openframe cluster status my-dev-cluster

# 3. Install OpenFrame charts
openframe chart install --deployment-mode=oss-tenant

# Get help for any command
openframe --help
```

## Commands

### Cluster Management
- `openframe cluster create` - Create a new K3d cluster with interactive wizard
- `openframe cluster list` - List all available clusters
- `openframe cluster status` - Show detailed cluster information and health
- `openframe cluster delete` - Remove a cluster and clean up resources
- `openframe cluster start` - Start a stopped cluster
- `openframe cluster cleanup` - Clean up cluster resources and Docker containers

### Chart Management
- `openframe chart install` - Install Helm charts and ArgoCD with app-of-apps pattern

### Development Tools
- `openframe dev scaffold` - Run Skaffold for continuous development workflows
- `openframe dev intercept` - Intercept service traffic with Telepresence for local debugging

### Bootstrap
- `openframe bootstrap` - End-to-end cluster setup with charts and development tools

## Prerequisites

OpenFrame CLI automatically checks for and guides you through installing these dependencies:

- **Docker** - Container runtime for K3d clusters
- **K3d** - Lightweight Kubernetes distribution
- **kubectl** - Kubernetes command-line tool
- **Helm** - Package manager for Kubernetes

Optional development tools:
- **Telepresence** - For local service development and debugging
- **Skaffold** - For continuous development workflows

## Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- **Getting Started** - Installation, prerequisites, and first steps
- **Development** - Contributing, architecture, and local development setup
- **Reference** - Technical architecture and API documentation

## Architecture

OpenFrame CLI follows a layered architecture with clear separation between command handling, business logic, and external system interactions. Built on the Cobra CLI framework with modular command groups for cluster management, chart operations, and development tools.

## Contributing

We welcome contributions! Please see our [Contributing Guidelines](./CONTRIBUTING.md) for details on:

- Development setup and workflows
- Code standards and testing requirements
- Pull request process
- Issue reporting and feature requests

## License

This project is licensed under the Flamingo AI Unified License v1.0. See the [LICENSE](./LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>