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

A modern command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI provides cluster lifecycle management, chart installation with ArgoCD, and development tools for local workflows using Telepresence and Skaffold.

## Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for K3d cluster setup
- ⚡ **Smart K3d Management** - Local Kubernetes development clusters
- 📊 **Real-time Monitoring** - Cluster status and health monitoring  
- 🔧 **System Auto-detection** - Intelligent configuration and validation
- 🛠 **Developer Tools** - Integrated Skaffold and Telepresence workflows
- 📦 **Chart Management** - Helm chart installation and ArgoCD deployment
- 🚀 **Bootstrap Automation** - End-to-end environment setup orchestration
- 💻 **Clear CLI Interface** - Developer-friendly commands with helpful output

## Quick Start

### Installation

**From Release (Recommended)**
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

**From Source**
```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
```

### Basic Usage

```bash
# Create a cluster with interactive wizard
openframe cluster create

# List all clusters
openframe cluster list

# Check cluster status
openframe cluster status

# Bootstrap complete OpenFrame environment
openframe bootstrap --deployment-mode=oss-tenant

# Get help for any command
openframe --help
openframe cluster --help
```

### Essential Commands

| Command | Description |
|---------|-------------|
| `openframe cluster create` | Create new K3d cluster with guided setup |
| `openframe cluster status` | Show detailed cluster information |
| `openframe bootstrap` | Complete OpenFrame environment setup |
| `openframe chart install` | Install Helm charts and ArgoCD |
| `openframe dev scaffold` | Run Skaffold for service development |
| `openframe dev intercept` | Intercept service traffic with Telepresence |

## Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- **Getting Started** - Installation, prerequisites, and first steps
- **Development** - Contributing, architecture, and local development
- **Reference** - Technical architecture and API documentation
- **Diagrams** - Visual system architecture and workflow diagrams

## Architecture

OpenFrame CLI follows a layered architecture with modular components:

- **CLI Layer** - Cobra command interface with intuitive commands
- **Service Layer** - Business logic for cluster, chart, and dev operations  
- **Internal Layer** - Core components for cluster management and bootstrapping
- **External Tools** - Integration with K3d, Helm, ArgoCD, Telepresence, Skaffold

## Contributing

We welcome contributions! Please see our [Contributing Guide](./CONTRIBUTING.md) for:

- Development environment setup
- Code style and standards  
- Testing requirements
- Pull request process

## Support

- 📖 [Documentation](./docs/README.md)
- 🐛 [Issue Tracker](https://github.com/flamingo-stack/openframe-cli/issues)
- 💬 [Discussions](https://github.com/flamingo-stack/openframe-cli/discussions)

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>