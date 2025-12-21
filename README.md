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

A modern CLI tool for managing OpenFrame Kubernetes clusters and development workflows. Built with Go and Cobra, the OpenFrame CLI provides interactive cluster creation, Helm chart management with ArgoCD integration, and developer tools for traffic interception and service scaffolding.

## Features

- 🎯 **Interactive cluster creation** - Guided wizard for K3d cluster setup
- ⚡ **K3d cluster management** - Full lifecycle management for local development
- 📊 **Real-time monitoring** - Cluster status and health monitoring
- 🔧 **Smart configuration** - Automatic system detection and setup
- 🛠 **Developer tools** - Skaffold and Telepresence integration
- 📦 **Chart orchestration** - ArgoCD and Helm chart management
- 🚀 **One-command bootstrap** - Complete environment setup
- 🔄 **CI/CD ready** - Non-interactive mode for automation

## Quick Start

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
```

#### From Source

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
```

### Basic Usage

```bash
# Create and bootstrap a complete OpenFrame environment
openframe bootstrap

# Or step by step:
# 1. Create a cluster
openframe cluster create my-cluster

# 2. Install charts with ArgoCD
openframe chart install my-cluster

# 3. Check cluster status
openframe cluster status

# 4. Start development with traffic interception
openframe dev intercept my-service --port 8080
```

## Core Commands

### Cluster Management
- `openframe cluster create` - Create a new K3d cluster with interactive wizard
- `openframe cluster list` - List all clusters and their status
- `openframe cluster status` - Show detailed cluster information
- `openframe cluster delete` - Remove a cluster and cleanup resources
- `openframe cluster start/stop` - Manage cluster lifecycle

### Chart & Application Management
- `openframe bootstrap` - Complete environment setup (cluster + charts)
- `openframe chart install` - Install ArgoCD and app-of-apps pattern
- Chart installation includes OpenFrame platform components

### Development Workflow
- `openframe dev intercept` - Telepresence traffic interception
- `openframe dev scaffold` - Skaffold development workflows
- Support for local development against remote clusters

### Global Options
- `--verbose, -v` - Enable detailed output and logging
- `--dry-run` - Preview operations without execution
- `--non-interactive` - Skip prompts for CI/CD automation

## Architecture

The CLI follows clean architecture principles with:

- **Command Layer** - Cobra-based CLI commands and argument parsing
- **Service Layer** - Business logic for cluster, chart, and dev operations  
- **Provider Layer** - Tool integrations (K3d, Helm, ArgoCD, Telepresence)
- **Shared Infrastructure** - Command execution, UI, configuration, error handling

Key integrations:
- **K3d** for lightweight Kubernetes clusters
- **ArgoCD** for GitOps-based application deployment
- **Helm** for package management
- **Telepresence** for service mesh traffic interception
- **Skaffold** for development workflow automation

## Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- **Getting Started** - Prerequisites, installation, and first steps
- **Development** - Architecture, contributing, and local development
- **Reference** - Technical documentation and API reference

## Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details on:
- Development environment setup
- Coding standards and conventions
- Testing requirements
- Pull request process

## License

This project is licensed under the Flamingo AI Unified License v1.0. See the [LICENSE.md](./LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>