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

A modern, interactive command-line tool for managing OpenFrame Kubernetes clusters and development workflows. OpenFrame CLI provides seamless cluster lifecycle management, chart installation with ArgoCD, and developer-friendly tools for service intercepts and scaffolding.

## 🚀 Watch OpenFrame in Action

[![OpenFrame Preview Webinar](https://img.youtube.com/vi/bINdW0CQbvY/maxresdefault.jpg)](https://www.youtube.com/watch?v=bINdW0CQbvY)

## ✨ Features

- 🎯 **Interactive Cluster Creation** - Guided wizard for creating K3D clusters with smart configuration
- ⚡ **Kubernetes Management** - Complete cluster lifecycle operations (create, list, status, delete)
- 📦 **Chart Installation** - Automated ArgoCD setup and OpenFrame application deployment
- 🔧 **Development Tools** - Service intercepts with Telepresence and Skaffold workflows
- 📊 **Real-time Monitoring** - Cluster status and health monitoring with rich terminal UI
- 🛠 **Smart Detection** - Automatic system prerequisite checking and installation guidance
- 🚀 **Bootstrap Workflow** - One-command setup for complete OpenFrame development environment

## 🏁 Quick Start

### Prerequisites

- Docker Desktop or Docker Engine
- 24GB RAM minimum (32GB recommended)
- 6 CPU cores minimum (12 recommended)

### Installation

**From Release (Recommended):**

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

**From Source:**

```bash
git clone https://github.com/flamingo-stack/openframe-cli.git
cd openframe-cli
go build -o openframe .
```

### Get Started in 5 Minutes

```bash
# Complete OpenFrame setup (creates cluster + installs charts)
openframe bootstrap

# Or step by step:
# 1. Create a cluster
openframe cluster create my-dev-cluster

# 2. Install OpenFrame charts
openframe chart install my-dev-cluster

# 3. Check status
openframe cluster status my-dev-cluster

# Start developing with service intercepts
openframe dev intercept my-service --port 8080
```

## 🎯 Core Commands

### Cluster Management
```bash
openframe cluster create      # Interactive cluster creation
openframe cluster list       # List all managed clusters
openframe cluster status     # Show detailed cluster information
openframe cluster delete     # Remove cluster and cleanup
openframe cluster start      # Start a stopped cluster
openframe cluster cleanup    # Clean up cluster resources
```

### Bootstrap & Charts
```bash
openframe bootstrap          # Complete setup (cluster + charts)
openframe chart install      # Install ArgoCD and OpenFrame apps
```

### Development Workflow
```bash
openframe dev intercept      # Telepresence service intercepts
openframe dev scaffold       # Run Skaffold development workflow
```

## 🏗 Architecture

OpenFrame CLI follows a modular, service-oriented architecture:

- **Command Layer**: Cobra-based CLI with rich terminal UI
- **Service Layer**: Business logic for cluster, chart, and dev operations  
- **Provider Layer**: Integrations with K3D, Helm, ArgoCD, Telepresence
- **Shared Infrastructure**: Common utilities, UI components, and error handling

The CLI orchestrates external tools like K3D for clusters, Helm for package management, ArgoCD for GitOps, and Telepresence for development workflows.

## 📚 Documentation

📚 See the [Documentation](./docs/README.md) for comprehensive guides including:

- **Getting Started**: Prerequisites, quick start, and first steps
- **Development**: Environment setup, architecture, testing, and contributing
- **Reference**: Technical architecture and API documentation

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](./CONTRIBUTING.md) for details on:

- Development environment setup
- Code standards and testing
- Pull request process
- Issue reporting

## 📄 License

This project is licensed under the Flamingo AI Unified License v1.0 - see the [LICENSE](LICENSE.md) file for details.

---
<div align="center">
  Built with 💛 by the <a href="https://www.flamingo.run/about"><b>Flamingo</b></a> team
</div>