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

A modern CLI tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI replaces complex shell scripts with an interactive terminal UI, making it easy to bootstrap OpenFrame Kubernetes deployments with guided workflows.

## ✨ Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for K3d cluster setup
- ⚡ **Smart K3d Management** - Full lifecycle management of local development clusters
- 📊 **Real-time Monitoring** - Live cluster status and health monitoring
- 🔧 **Auto-Configuration** - Smart system detection and automated tool installation
- 🛠 **Developer Tools** - Integrated Skaffold and Telepresence workflows
- 📦 **ArgoCD Integration** - Automated Helm chart deployment with GitOps
- 🚀 **One-Command Bootstrap** - Complete OpenFrame environment in minutes

## 🚀 Quick Start

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

# Windows (AMD64)
# Download from: https://github.com/flamingo-stack/openframe-cli/releases/latest
```

### Get Started in 60 Seconds

```bash
# Bootstrap a complete OpenFrame environment
openframe bootstrap --deployment-mode=oss-tenant

# Or step-by-step:
# 1. Create a cluster
openframe cluster create

# 2. Install OpenFrame charts
openframe chart install

# 3. Check status
openframe cluster status
```

## 🎯 Core Commands

### Cluster Management
- `openframe cluster create` - Create a new K3d cluster with interactive setup
- `openframe cluster list` - List all managed clusters
- `openframe cluster status` - Show detailed cluster health and resources
- `openframe cluster delete` - Clean removal of clusters and resources
- `openframe cluster cleanup` - Remove unused Docker resources

### Chart & GitOps
- `openframe chart install` - Install Helm charts and configure ArgoCD
- `openframe bootstrap` - Complete environment setup (cluster + charts)

### Development Workflows
- `openframe dev intercept` - Intercept service traffic with Telepresence
- `openframe dev scaffold` - Live development with Skaffold hot-reload

## 📋 Prerequisites

The CLI automatically detects and installs missing tools:

- **Docker** - Container runtime (auto-installed on macOS/Windows)
- **K3d** - Lightweight Kubernetes (auto-installed)
- **kubectl** - Kubernetes CLI (auto-installed)
- **Helm** - Package manager (auto-installed)

**Hardware Requirements:**
- **Minimum**: 8GB RAM, 4 CPU cores, 20GB disk
- **Recommended**: 16GB RAM, 8 CPU cores, 50GB disk

## 🏗 Architecture

OpenFrame CLI follows a layered architecture with clear separation of concerns:

- **Command Layer**: Cobra-based CLI interface with interactive prompts
- **Service Layer**: Business logic for cluster, chart, and development operations  
- **Provider Layer**: Abstractions for K3d, Helm, Git, and Telepresence
- **Infrastructure**: Command execution, UI components, and prerequisites management

The CLI supports both interactive wizard workflows for beginners and flag-based automation for CI/CD pipelines.

## 📚 Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- 🚀 [Getting Started Guide](./docs/getting-started/introduction.md)
- ⚙️ [Development Setup](./docs/development/setup/environment.md)
- 🏛 [Architecture Overview](./docs/reference/architecture/overview.md)
- 🧪 [Testing Guide](./docs/development/testing/overview.md)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for details on:

- Setting up the development environment
- Running tests and linting
- Submitting pull requests
- Code style and conventions

## 📄 License

This project is licensed under the Flamingo AI Unified License v1.0. See the [LICENSE.md](LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>